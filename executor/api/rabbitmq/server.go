package rabbitmq

import (
	"context"
	"errors"
	"fmt"
	guuid "github.com/google/uuid"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"net/url"
	"os"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"time"

	"github.com/go-logr/logr"
	"github.com/seldonio/seldon-core/executor/api"
	"github.com/seldonio/seldon-core/executor/api/client"
	"github.com/seldonio/seldon-core/executor/api/grpc/seldon"
	"github.com/seldonio/seldon-core/executor/api/grpc/tensorflow"
	"github.com/seldonio/seldon-core/executor/api/payload"
	"github.com/seldonio/seldon-core/executor/api/rest"
	pred "github.com/seldonio/seldon-core/executor/predictor"
	v1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
)

/*
 * based on `kafka/server.go`
 */

const (
	ENV_RABBITMQ_BROKER_URL   = "RABBITMQ_BROKER_URL"
	ENV_RABBITMQ_INPUT_QUEUE  = "RABBITMQ_INPUT_QUEUE"
	ENV_RABBITMQ_OUTPUT_QUEUE = "RABBITMQ_OUTPUT_QUEUE"
	ENV_RABBITMQ_FULL_GRAPH   = "RABBITMQ_FULL_GRAPH"
)

type SeldonRabbitMQServer struct {
	Client          client.SeldonApiClient
	DeploymentName  string
	Namespace       string
	Transport       string
	Predictor       v1.PredictorSpec
	ServerUrl       url.URL
	BrokerUrl       string
	InputQueueName  string
	OutputQueueName string
	Log             logr.Logger
	Protocol        string
	FullHealthCheck bool
}

type RabbitMQServerOptions struct {
	FullGraph       bool
	DeploymentName  string
	Namespace       string
	Protocol        string
	Transport       string
	Annotations     map[string]string
	ServerUrl       url.URL
	Predictor       v1.PredictorSpec
	BrokerUrl       string
	InputQueueName  string
	OutputQueueName string
	Log             logr.Logger
	FullHealthCheck bool
}

func CreateRabbitMQServer(args RabbitMQServerOptions) (*SeldonRabbitMQServer, error) {
	deploymentName, protocol, transport, annotations, predictor, log :=
		args.DeploymentName, args.Protocol, args.Transport, args.Annotations, args.Predictor, args.Log

	var apiClient client.SeldonApiClient
	var err error
	if args.FullGraph {
		log.Info("Starting full graph rabbitmq server")
		//apiClient = NewRabbitMqClient(brokerUrl.Hostname(), deploymentName, namespace, protocol, transport, predictor, broker, log)
		return nil, errors.New("full graph not currently supported")
	}

	switch args.Transport {
	case api.TransportRest:
		log.Info("Start http rabbitmq graph")
		apiClient, err = rest.NewJSONRestClient(protocol, deploymentName, &predictor, annotations)
		if err != nil {
			return nil, fmt.Errorf("error %w creating json rest client", err)
		}
	case api.TransportGrpc:
		log.Info("Start grpc rabbitmq graph")
		if protocol == "seldon" {
			apiClient = seldon.NewSeldonGrpcClient(&predictor, deploymentName, annotations)
		} else {
			apiClient = tensorflow.NewTensorflowGrpcClient(&predictor, deploymentName, annotations)
		}
	default:
		return nil, fmt.Errorf("unknown Transport %s", transport)
	}

	return &SeldonRabbitMQServer{
		Client:          apiClient,
		DeploymentName:  deploymentName,
		Namespace:       args.Namespace,
		Transport:       transport,
		Predictor:       predictor,
		ServerUrl:       args.ServerUrl,
		BrokerUrl:       args.BrokerUrl,
		InputQueueName:  args.InputQueueName,
		OutputQueueName: args.OutputQueueName,
		Log:             log.WithName("RabbitMqServer"),
		Protocol:        protocol,
		FullHealthCheck: args.FullHealthCheck,
	}, nil
}

func (rs *SeldonRabbitMQServer) Serve() error {
	conn, err := NewConnection(rs.BrokerUrl, rs.Log)
	if err != nil {
		return fmt.Errorf("error %w connecting to rabbitmq", err)
	}

	return rs.serve(conn)
}

func (rs *SeldonRabbitMQServer) serve(conn *connection) error {
	//TODO not sure if this is the best pattern or better to pass in pod name explicitly somehow
	consumerTag, err := os.Hostname()
	if err != nil {
		return fmt.Errorf("error %w retrieving hostname", err)
	}

	// blank consumer tag means auto-generate
	c := &consumer{*conn, rs.InputQueueName, consumerTag}
	rs.Log.Info("Created", "consumer", c, "input queue", rs.InputQueueName)

	//wait for graph to be ready
	ready := false
	for ready == false {
		err := pred.Ready(rs.Protocol, &rs.Predictor.Graph, rs.FullHealthCheck)
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

	// create output queue immediately instead of waiting until the first time we publish a response
	_, err = conn.DeclareQueue(rs.OutputQueueName)
	if err != nil {
		return fmt.Errorf("error %w declaring rabbitmq queue", err)
	}

	// consumer creates input queue if it doesn't exist
	err = c.Consume(
		func(reqPl SeldonPayloadWithHeaders) error { return rs.predictAndPublishResponse(reqPl, conn) },
		errorHandler)
	if err != nil {
		return fmt.Errorf("error %w in consumer", err)
	}

	rs.Log.Info("Consumer exited without error")
	return nil
}

func (rs *SeldonRabbitMQServer) predictAndPublishResponse(reqPayload SeldonPayloadWithHeaders, conn *connection) error {
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
	seldonPredictorProcess := pred.NewPredictorProcess(
		ctx, rs.Client, logf.Log.WithName("RabbitMqClient"), &rs.ServerUrl, rs.Namespace, reqPayload.Headers, "")

	resPayload, err := seldonPredictorProcess.Predict(&rs.Predictor.Graph, reqPayload)
	if err != nil && resPayload == nil {
		// normal errors from the predict process contain a status failed payload
		// this is handling an unexpected case, so failing entirely, at least for now
		return fmt.Errorf("unhandled error %w from predictor process", err)
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
