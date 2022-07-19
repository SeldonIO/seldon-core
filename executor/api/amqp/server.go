package amqp

import (
	"context"
	"errors"
	"fmt"
	guuid "github.com/google/uuid"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"net/url"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"time"

	"github.com/go-logr/logr"
	"github.com/seldonio/seldon-core/executor/api"
	"github.com/seldonio/seldon-core/executor/api/client"
	"github.com/seldonio/seldon-core/executor/api/grpc/seldon"
	"github.com/seldonio/seldon-core/executor/api/grpc/tensorflow"
	"github.com/seldonio/seldon-core/executor/api/payload"
	"github.com/seldonio/seldon-core/executor/api/rest"
	"github.com/seldonio/seldon-core/executor/predictor"
	v1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
)

const (
	ENV_AMQP_BROKER_URL   = "AMQP_BROKER_URL"
	ENV_AMQP_INPUT_QUEUE  = "AMQP_INPUT_QUEUE"
	ENV_AMQP_OUTPUT_QUEUE = "AMQP_OUTPUT_QUEUE"
	ENV_AMQP_FULL_GRAPH   = "AMQP_FULL_GRAPH"
)

type SeldonAmqpServer struct {
	Client          client.SeldonApiClient
	DeploymentName  string
	Namespace       string
	Transport       string
	Predictor       *v1.PredictorSpec
	ServerUrl       *url.URL
	BrokerUrl       string
	InputQueueName  string
	OutputQueueName string
	Log             logr.Logger
	Protocol        string
	FullHealthCheck bool
}

func NewAmqpServer(
	fullGraph bool,
	deploymentName,
	namespace,
	protocol,
	transport string,
	annotations map[string]string,
	serverUrl *url.URL,
	predictor *v1.PredictorSpec,
	brokerUrl string,
	inputQueueName,
	outputQueueName string,
	log logr.Logger,
	fullHealthCheck bool,
) (*SeldonAmqpServer, error) {
	var apiClient client.SeldonApiClient
	var err error
	if fullGraph {
		log.Info("Starting full graph amqp server")
		//apiClient = NewAmqpClient(brokerUrl.Hostname(), deploymentName, namespace, protocol, transport, predictor, broker, log)
		return nil, errors.New("full graph not currently supported")
	} else {
		switch transport {
		case api.TransportRest:
			log.Info("Start http amqp graph")
			apiClient, err = rest.NewJSONRestClient(protocol, deploymentName, predictor, annotations)
			if err != nil {
				return nil, err
			}
		case api.TransportGrpc:
			log.Info("Start grpc amqp graph")
			if protocol == "seldon" {
				apiClient = seldon.NewSeldonGrpcClient(predictor, deploymentName, annotations)
			} else {
				apiClient = tensorflow.NewTensorflowGrpcClient(predictor, deploymentName, annotations)
			}
		default:
			return nil, fmt.Errorf("Unknown transport %s", transport)
		}
	}

	return &SeldonAmqpServer{
		Client:          apiClient,
		DeploymentName:  deploymentName,
		Namespace:       namespace,
		Transport:       transport,
		Predictor:       predictor,
		ServerUrl:       serverUrl,
		BrokerUrl:       brokerUrl,
		InputQueueName:  inputQueueName,
		OutputQueueName: outputQueueName,
		Log:             log.WithName("AmqpServer"),
		Protocol:        protocol,
		FullHealthCheck: fullHealthCheck,
	}, nil
}

func (as *SeldonAmqpServer) Serve() error {

	// TODO consumerTag likely needs to have a pod/container-level identifier appended onto it
	c, err := NewConsumer(as.BrokerUrl, as.InputQueueName, as.DeploymentName, as.Log)
	if err != nil {
		return err
	}
	as.Log.Info("Created", "consumer", c, "input queue", as.InputQueueName)

	//wait for graph to be ready
	ready := false
	for ready == false {
		err := predictor.Ready(as.Protocol, &as.Predictor.Graph, as.FullHealthCheck)
		ready = err == nil
		if !ready {
			as.Log.Info("Waiting for graph to be ready")
			time.Sleep(2 * time.Second)
		}
	}

	errorHandler := func(errToHandle error) {
		// TODO probably need to do something better than this
		as.Log.Error(errToHandle, "error processing message")
	}

	err = c.Consume(as.PredictAndPublishResponse, errorHandler)
	if err != nil {
		return err
	}

	as.Log.Info("Consumer exited")
	return nil
}

func (as *SeldonAmqpServer) PredictAndPublishResponse(reqPayload SeldonPayloadWithHeaders) error {

	producer, err := NewPublisher(as.BrokerUrl, as.OutputQueueName, as.Log)
	if err != nil {
		return err
	}

	ctx := context.Background()
	// Add Seldon Puid to Context
	if reqPayload.Headers[payload.SeldonPUIDHeader] == nil {
		reqPayload.Headers[payload.SeldonPUIDHeader] = []string{guuid.New().String()}
	}

	ctx = context.WithValue(ctx, payload.SeldonPUIDHeader, reqPayload.Headers[payload.SeldonPUIDHeader][0])

	// Apply tracing if active
	if opentracing.IsGlobalTracerRegistered() {
		tracer := opentracing.GlobalTracer()
		serverSpan := tracer.StartSpan("amqpServer", ext.RPCServerOption(nil))
		ctx = opentracing.ContextWithSpan(ctx, serverSpan)
		defer serverSpan.Finish()
	}

	as.Log.Info("amqp server values", "server url", as.ServerUrl)
	seldonPredictorProcess := predictor.NewPredictorProcess(
		ctx, as.Client, logf.Log.WithName("AmqpClient"), as.ServerUrl, as.Namespace, reqPayload.Headers, "")

	resPayload, err := seldonPredictorProcess.Predict(&as.Predictor.Graph, reqPayload)
	if err != nil {
		//as.Log.Error(err, "Failed prediction")
		// TODO should this just place an error into the channel or bomb out?
		return err
	}

	resHeaders := make(map[string][]string)
	resHeaders[payload.SeldonPUIDHeader] = reqPayload.Headers[payload.SeldonPUIDHeader]
	//TODO might need more headers

	resPayloadWithHeaders := SeldonPayloadWithHeaders{
		resPayload,
		resHeaders,
	}

	return producer.Publish(resPayloadWithHeaders)
}
