package rabbitmq

import (
	"errors"
	"fmt"
	"github.com/go-logr/logr/testr"
	guuid "github.com/google/uuid"
	. "github.com/onsi/gomega"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/seldonio/seldon-core/executor/api"
	"github.com/seldonio/seldon-core/executor/api/payload"
	"github.com/seldonio/seldon-core/executor/api/rest"
	"github.com/seldonio/seldon-core/executor/api/test"
	v1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
)

var (
	model           = v1.MODEL
	deploymentName  = "testDep"
	namespace       = "testNs"
	protocol        = api.ProtocolSeldon
	transport       = api.TransportRest
	brokerUrl       = "amqp://something.com"
	inputQueue      = "inputQueue"
	outputQueue     = "outputQueue"
	fullHealthCheck = false
)

func TestRabbitMqServer(t *testing.T) {
	log := testr.New(t)
	// this bit down to where `p` is defined is taken from rest/server_test.go
	g := NewGomegaWithT(t)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bodyBytes, err := ioutil.ReadAll(r.Body)
		g.Expect(err).To(BeNil())
		g.Expect(r.Header.Get(payload.SeldonPUIDHeader)).To(Equal("1"))
		//called = true
		_, _ = w.Write(bodyBytes)
	})
	server := httptest.NewServer(handler)
	defer server.Close()
	serverUrl, err := url.Parse(server.URL)
	g.Expect(err).Should(BeNil())
	urlParts := strings.Split(serverUrl.Host, ":")
	port, err := strconv.Atoi(urlParts[1])
	g.Expect(err).Should(BeNil())

	p := v1.PredictorSpec{
		Name: "p",
		Graph: v1.PredictiveUnit{
			Type: &model,
			Endpoint: &v1.Endpoint{
				ServiceHost: urlParts[0],
				ServicePort: int32(port),
				Type:        v1.REST,
				HttpPort:    int32(port),
			},
		},
	}
	serverUrl, _ = url.Parse("http://localhost")

	testServer := SeldonRabbitMQServer{
		Client:          test.SeldonMessageTestClient{},
		DeploymentName:  deploymentName,
		Namespace:       namespace,
		Protocol:        protocol,
		Transport:       transport,
		ServerUrl:       *serverUrl,
		Predictor:       p,
		BrokerUrl:       brokerUrl,
		InputQueueName:  inputQueue,
		OutputQueueName: outputQueue,
		Log:             log,
		FullHealthCheck: fullHealthCheck,
	}

	testPuid := guuid.New().String()
	testHeaders := map[string]interface{}{payload.SeldonPUIDHeader: testPuid}

	invalidErrorJsonResponse := fmt.Sprintf(
		`{"status":{"info":"Prediction Failed","reason":"unknown payload type 'bogus'","status":"FAILURE"},"meta":{"puid":"%v"}}`,
		testPuid,
	)
	invalidErrorPublishing := amqp.Publishing{
		ContentType: "application/json",
		Body:        []byte(invalidErrorJsonResponse),
		Headers:     testHeaders,
	}

	t.Run("create server", func(t *testing.T) {
		server, err := CreateRabbitMQServer(RabbitMQServerOptions{
			DeploymentName:  deploymentName,
			Namespace:       namespace,
			Protocol:        protocol,
			Transport:       transport,
			Annotations:     map[string]string{},
			ServerUrl:       *serverUrl,
			Predictor:       p,
			BrokerUrl:       brokerUrl,
			InputQueueName:  inputQueue,
			OutputQueueName: outputQueue,
			Log:             log,
			FullHealthCheck: fullHealthCheck,
		})

		assert.NoError(t, err)
		assert.NotNil(t, server)
		assert.NotNil(t, server.Client)
		assert.Equal(t, deploymentName, server.DeploymentName)
		assert.Equal(t, namespace, server.Namespace)
		assert.Equal(t, transport, server.Transport)
		assert.Equal(t, p, server.Predictor)
		assert.Equal(t, *serverUrl, server.ServerUrl)
		assert.Equal(t, brokerUrl, server.BrokerUrl)
		assert.Equal(t, inputQueue, server.InputQueueName)
		assert.Equal(t, outputQueue, server.OutputQueueName)
		assert.NotNil(t, server.Log)
		assert.Equal(t, protocol, server.Protocol)
		assert.Equal(t, fullHealthCheck, server.FullHealthCheck)
	})

	/*
	 * This makes sure the Serve() and predictAndPublishResponse() code runs and makes the proper calls
	 * by hacking a bunch of mocks.
	 * It is not doing anything to validate the messages are properly processed.  That's challenging in a
	 * unit test since the code connects to RabbitMQ.
	 */
	t.Run("serve", func(t *testing.T) {
		mockRmqConn := &mockConnection{}
		mockRmqChan := &mockChannel{}
		mockConn := &connection{
			conn:    mockRmqConn,
			channel: mockRmqChan,
		}

		testDelivery := amqp.Delivery{
			Acknowledger: mockRmqChan,
			ContentType:  rest.ContentTypeJSON,
			Body:         []byte(`{ "data": { "ndarray": [[1,2,3,4]] } }`),
			Headers:      testHeaders,
		}

		mockRmqChan.On("QueueDeclare", outputQueue, queueDurable, queueAutoDelete, queueExclusive,
			queueNoWait, queueArgs).Return(amqp.Queue{}, nil)
		mockRmqChan.On("QueueDeclare", inputQueue, queueDurable, queueAutoDelete, queueExclusive,
			queueNoWait, queueArgs).Return(amqp.Queue{}, nil)

		mockDeliveries := make(chan amqp.Delivery, 1)
		mockDeliveries <- testDelivery
		close(mockDeliveries)

		mockRmqChan.On("Consume", inputQueue, mock.Anything, false, false, false, false,
			amqp.Table{}).Return(mockDeliveries, nil)
		mockRmqChan.On("Publish", "", outputQueue, publishMandatory, publishImmediate,
			mock.MatchedBy(func(p amqp.Publishing) bool { return true })).Return(nil)

		err := testServer.serve(mockConn)

		assert.NoError(t, err)

		mockRmqChan.AssertExpectations(t)
	})

	/*
	 * This makes sure the Serve(), predictAndPublishResponse(), and createAndPublishErrorResponse() code runs and
	 * makes the proper calls and returns an appropriate error message by hacking a bunch of mocks.
	 */
	t.Run("serveError", func(t *testing.T) {
		mockRmqConn := &mockConnection{}
		mockRmqChan := &mockChannel{}
		mockConn := &connection{
			conn:    mockRmqConn,
			channel: mockRmqChan,
		}

		invalidDelivery := amqp.Delivery{
			Acknowledger: mockRmqChan,
			ContentType:  "bogus",
			Body:         []byte(`bogus`),
			Headers:      testHeaders,
			Redelivered:  true,
		}

		mockRmqChan.On("QueueDeclare", outputQueue, queueDurable, queueAutoDelete, queueExclusive,
			queueNoWait, queueArgs).Return(amqp.Queue{}, nil)
		mockRmqChan.On("QueueDeclare", inputQueue, queueDurable, queueAutoDelete, queueExclusive,
			queueNoWait, queueArgs).Return(amqp.Queue{}, nil)

		mockDeliveries := make(chan amqp.Delivery, 1)
		mockDeliveries <- invalidDelivery
		close(mockDeliveries)

		mockRmqChan.On("Consume", inputQueue, mock.Anything, false, false, false, false,
			amqp.Table{}).Return(mockDeliveries, nil)
		mockRmqChan.On("Publish", "", outputQueue, publishMandatory, publishImmediate,
			invalidErrorPublishing).Return(nil)

		err := testServer.serve(mockConn)

		assert.NoError(t, err)

		mockRmqChan.AssertExpectations(t)
	})

	t.Run("createAndPublishErrorResponse", func(t *testing.T) {
		mockRmqConn := &mockConnection{}
		mockRmqChan := &mockChannel{}
		mockConn := &connection{
			conn:    mockRmqConn,
			channel: mockRmqChan,
		}

		testDelivery := amqp.Delivery{
			Acknowledger: mockRmqChan,
			ContentType:  rest.ContentTypeJSON,
			Body:         []byte(`{ "data": { "ndarray": [[1,2,3,4]] } }`),
			Headers:      testHeaders,
		}

		mockRmqChan.On("QueueDeclare", outputQueue, queueDurable, queueAutoDelete, queueExclusive,
			queueNoWait, queueArgs).Return(amqp.Queue{}, nil)

		publisher := &publisher{*mockConn, outputQueue}

		// valid payload
		error1Text := "error 1"
		error1 := errors.New(error1Text)
		generatedErrorPublishing1 := amqp.Publishing{
			ContentType: "application/json",
			Body: []byte(fmt.Sprintf(
				`{"status":{"info":"Prediction Failed","reason":"%v","status":"FAILURE"},"meta":{"puid":"%v"}}`,
				error1Text,
				testPuid,
			)),
			Headers: testHeaders,
		}
		mockRmqChan.On("Publish", "", outputQueue, true, false, generatedErrorPublishing1).Return(nil)
		pl1, _ := DeliveryToPayload(testDelivery)
		consumerError1 := ConsumerError{
			err:      error1,
			delivery: testDelivery,
			pl:       pl1,
		}
		err1 := testServer.createAndPublishErrorResponse(consumerError1, publisher)
		assert.NoError(t, err1)

		mockRmqChan.AssertExpectations(t)

		// no payload
		error2Text := "error 2"
		error2 := errors.New(error2Text)
		generatedErrorPublishing2 := amqp.Publishing{
			ContentType: "application/json",
			Body: []byte(fmt.Sprintf(
				`{"status":{"info":"Prediction Failed","reason":"%v","status":"FAILURE"},"meta":{"puid":"%v"}}`,
				error2Text,
				testPuid,
			)),
			Headers: testHeaders,
		}
		mockRmqChan.On("Publish", "", outputQueue, true, false, generatedErrorPublishing2).Return(nil)
		consumerError2 := ConsumerError{
			err:      error2,
			delivery: testDelivery,
		}
		err2 := testServer.createAndPublishErrorResponse(consumerError2, publisher)
		assert.NoError(t, err2)

		mockRmqChan.AssertExpectations(t)
	})

}
