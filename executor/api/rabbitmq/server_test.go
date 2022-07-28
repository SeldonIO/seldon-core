package rabbitmq

import (
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

func TestRabbitMqServer(t *testing.T) {
	// this bit down to where `p` is defined is taken from rest/server_test.go
	g := NewGomegaWithT(t)
	//called := false

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bodyBytes, err := ioutil.ReadAll(r.Body)
		g.Expect(err).To(BeNil())
		g.Expect(r.Header.Get(payload.SeldonPUIDHeader)).To(Equal("1"))
		//called = true
		w.Write([]byte(bodyBytes))
	})
	server := httptest.NewServer(handler)
	defer server.Close()
	url, err := url.Parse(server.URL)
	g.Expect(err).Should(BeNil())
	urlParts := strings.Split(url.Host, ":")
	port, err := strconv.Atoi(urlParts[1])
	g.Expect(err).Should(BeNil())

	model := v1.MODEL
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

	serverUrl, _ := url.Parse("http://localhost")
	deploymentName := "testDep"
	namespace := "testNs"
	protocol := api.ProtocolSeldon
	transport := api.TransportRest
	brokerUrl := "amqp://something.com"
	inputQueue := "inputQueue"
	outputQueue := "outputQueue"
	fullHealthCheck := false

	testServer := SeldonRabbitMQServer{
		Client:          test.SeldonMessageTestClient{},
		DeploymentName:  deploymentName,
		Namespace:       namespace,
		Transport:       transport,
		Predictor:       &p,
		ServerUrl:       serverUrl,
		BrokerUrl:       brokerUrl,
		InputQueueName:  inputQueue,
		OutputQueueName: outputQueue,
		Log:             logger,
		Protocol:        protocol,
		FullHealthCheck: fullHealthCheck,
	}

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
	}

	t.Run("create server", func(t *testing.T) {
		server, err := CreateRabbitMQServer(RabbitMQServerOptions{
			DeploymentName:  deploymentName,
			Namespace:       namespace,
			Protocol:        protocol,
			Transport:       transport,
			Annotations:     map[string]string{},
			ServerUrl:       serverUrl,
			Predictor:       &p,
			BrokerUrl:       brokerUrl,
			InputQueueName:  inputQueue,
			OutputQueueName: outputQueue,
			Log:             logger,
			FullHealthCheck: fullHealthCheck,
		})

		assert.NoError(t, err)
		assert.NotNil(t, server)
		assert.NotNil(t, server.Client)
		assert.Equal(t, deploymentName, server.DeploymentName)
		assert.Equal(t, namespace, server.Namespace)
		assert.Equal(t, transport, server.Transport)
		assert.Equal(t, &p, server.Predictor)
		assert.Equal(t, serverUrl, server.ServerUrl)
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

}
