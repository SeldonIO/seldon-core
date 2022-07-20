package rabbitmq

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
	ENV_RABBITMQ_BROKER_URL   = "RABBITMQ_BROKER_URL"
	ENV_RABBITMQ_INPUT_QUEUE  = "RABBITMQ_INPUT_QUEUE"
	ENV_RABBITMQ_OUTPUT_QUEUE = "RABBITMQ_OUTPUT_QUEUE"
	ENV_RABBITMQ_FULL_GRAPH   = "RABBITMQ_FULL_GRAPH"
)

type SeldonRabbitMqServer struct {
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

func NewRabbitMqServer(
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
) (*SeldonRabbitMqServer, error) {
	var apiClient client.SeldonApiClient
	var err error
	if fullGraph {
		log.Info("Starting full graph rabbitmq server")
		//apiClient = NewRabbitMqClient(brokerUrl.Hostname(), deploymentName, namespace, protocol, transport, predictor, broker, log)
		return nil, errors.New("full graph not currently supported")
	} else {
		switch transport {
		case api.TransportRest:
			log.Info("Start http rabbitmq graph")
			apiClient, err = rest.NewJSONRestClient(protocol, deploymentName, predictor, annotations)
			if err != nil {
				return nil, err
			}
		case api.TransportGrpc:
			log.Info("Start grpc rabbitmq graph")
			if protocol == "seldon" {
				apiClient = seldon.NewSeldonGrpcClient(predictor, deploymentName, annotations)
			} else {
				apiClient = tensorflow.NewTensorflowGrpcClient(predictor, deploymentName, annotations)
			}
		default:
			return nil, fmt.Errorf("Unknown transport %s", transport)
		}
	}

	return &SeldonRabbitMqServer{
		Client:          apiClient,
		DeploymentName:  deploymentName,
		Namespace:       namespace,
		Transport:       transport,
		Predictor:       predictor,
		ServerUrl:       serverUrl,
		BrokerUrl:       brokerUrl,
		InputQueueName:  inputQueueName,
		OutputQueueName: outputQueueName,
		Log:             log.WithName("RabbitMqServer"),
		Protocol:        protocol,
		FullHealthCheck: fullHealthCheck,
	}, nil
}

func (rs *SeldonRabbitMqServer) Serve() error {

	conn, err := NewConnection(rs.BrokerUrl, rs.Log)

	// TODO consumerTag likely needs to have a pod/container-level identifier appended onto it
	c := &consumer{*conn, rs.InputQueueName, rs.DeploymentName}
	rs.Log.Info("Created", "consumer", c, "input queue", rs.InputQueueName)

	//wait for graph to be ready
	ready := false
	for ready == false {
		err := predictor.Ready(rs.Protocol, &rs.Predictor.Graph, rs.FullHealthCheck)
		ready = err == nil
		if !ready {
			rs.Log.Info("Waiting for graph to be ready")
			time.Sleep(2 * time.Second)
		}
	}

	errorHandler := func(errToHandle error) {
		// TODO probably need to do something better than this
		rs.Log.Error(errToHandle, "error processing message")
	}

	err = c.Consume(
		func(reqPl SeldonPayloadWithHeaders) error { return rs.PredictAndPublishResponse(reqPl, conn) },
		errorHandler)
	if err != nil {
		return err
	}

	rs.Log.Info("Consumer exited")
	return nil
}

func (rs *SeldonRabbitMqServer) PredictAndPublishResponse(reqPayload SeldonPayloadWithHeaders, conn *connection) error {

	producer := &publisher{*conn, rs.OutputQueueName}

	ctx := context.Background()
	// Add Seldon Puid to Context
	if reqPayload.Headers[payload.SeldonPUIDHeader] == nil {
		reqPayload.Headers[payload.SeldonPUIDHeader] = []string{guuid.New().String()}
	}

	ctx = context.WithValue(ctx, payload.SeldonPUIDHeader, reqPayload.Headers[payload.SeldonPUIDHeader][0])

	// Apply tracing if active
	if opentracing.IsGlobalTracerRegistered() {
		tracer := opentracing.GlobalTracer()
		serverSpan := tracer.StartSpan("rabbitMqServer", ext.RPCServerOption(nil))
		ctx = opentracing.ContextWithSpan(ctx, serverSpan)
		defer serverSpan.Finish()
	}

	rs.Log.Info("rabbitmq server values", "server url", rs.ServerUrl)
	seldonPredictorProcess := predictor.NewPredictorProcess(
		ctx, rs.Client, logf.Log.WithName("RabbitMqClient"), rs.ServerUrl, rs.Namespace, reqPayload.Headers, "")

	resPayload, err := seldonPredictorProcess.Predict(&rs.Predictor.Graph, reqPayload)
	if err != nil {
		//rs.Log.Error(err, "Failed prediction")
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
