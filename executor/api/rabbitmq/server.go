package rabbitmq

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang/protobuf/jsonpb"
	guuid "github.com/google/uuid"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/seldonio/seldon-core/executor/api/grpc/seldon/proto"
	"net/url"
	"os"
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
	Protocol        string
	Transport       string
	ServerUrl       url.URL
	Predictor       v1.PredictorSpec
	BrokerUrl       string
	InputQueueName  string
	OutputQueueName string
	Log             logr.Logger
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
		err = errors.New("full graph not currently supported")
		log.Error(err, "tried to use full graph mode but not currently supported for RabbitMQ server")
		return nil, err
	}

	switch args.Transport {
	case api.TransportRest:
		log.Info("Start http rabbitmq graph")
		apiClient, err = rest.NewJSONRestClient(protocol, deploymentName, &predictor, annotations)
		if err != nil {
			log.Error(err, "error creating json rest client")
			return nil, fmt.Errorf("error '%w' creating json rest client", err)
		}
	case api.TransportGrpc:
		log.Info("Start grpc rabbitmq graph")
		if protocol == "seldon" {
			apiClient = seldon.NewSeldonGrpcClient(&predictor, deploymentName, annotations)
		} else {
			apiClient = tensorflow.NewTensorflowGrpcClient(&predictor, deploymentName, annotations)
		}
	default:
		return nil, fmt.Errorf("unknown transport '%s'", transport)
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
		rs.Log.Error(err, "error connecting to rabbitmq")
		return fmt.Errorf("error '%w' connecting to rabbitmq", err)
	}

	return rs.serve(conn)
}

func (rs *SeldonRabbitMQServer) serve(conn *connection) error {
	//TODO not sure if this is the best pattern or better to pass in pod name explicitly somehow
	consumerTag, err := os.Hostname()
	if err != nil {
		return fmt.Errorf("error '%w' retrieving hostname", err)
	}

	consumer := &consumer{*conn, rs.InputQueueName, consumerTag}
	rs.Log.Info("Created", "consumer", consumer, "input queue", rs.InputQueueName)

	// wait for graph to be ready
	ready := false
	for ready == false {
		err := pred.Ready(rs.Protocol, &rs.Predictor.Graph, rs.FullHealthCheck)
		ready = err == nil
		if !ready {
			rs.Log.Info("Waiting for graph to be ready")
			time.Sleep(2 * time.Second)
		}
	}

	// create output queue immediately instead of waiting until the first time we publish a response
	_, err = conn.DeclareQueue(rs.OutputQueueName)
	if err != nil {
		rs.Log.Error(err, "error declaring rabbitmq queue", "uri", conn.uri)
		return fmt.Errorf("error '%w' declaring rabbitmq queue", err)
	}

	producer := &publisher{*conn, rs.OutputQueueName}

	// consumer creates input queue if it doesn't exist
	err = consumer.Consume(
		func(reqPl *SeldonPayloadWithHeaders) error { return rs.predictAndPublishResponse(reqPl, producer) },
		func(args ConsumerError) error { return rs.createAndPublishErrorResponse(args, producer) },
	)
	if err != nil {
		rs.Log.Error(err, "error in consumer")
		return fmt.Errorf("error '%w' in consumer", err)
	}

	rs.Log.Info("Consumer exited without error")
	return nil
}

func (rs *SeldonRabbitMQServer) predictAndPublishResponse(
	reqPayload *SeldonPayloadWithHeaders,
	publisher *publisher,
) error {
	if reqPayload == nil {
		err := errors.New("missing request payload")
		rs.Log.Error(err, "passed in request payload was blank")
		return err
	}

	seldonPuid := assignAndReturnPUID(reqPayload)
	ctx := context.WithValue(context.Background(), payload.SeldonPUIDHeader, seldonPuid)

	// Apply tracing if active
	if opentracing.IsGlobalTracerRegistered() {
		tracer := opentracing.GlobalTracer()
		serverSpan := tracer.StartSpan("rabbitMqServer", ext.RPCServerOption(nil))
		ctx = opentracing.ContextWithSpan(ctx, serverSpan)
		defer serverSpan.Finish()
	}

	rs.Log.Info("rabbitmq server values", "server url", rs.ServerUrl)
	seldonPredictorProcess := pred.NewPredictorProcess(
		ctx, rs.Client, rs.Log.WithName("RabbitMqClient"), &rs.ServerUrl, rs.Namespace, reqPayload.Headers, "")

	resPayload, err := seldonPredictorProcess.Predict(&rs.Predictor.Graph, reqPayload)
	if err != nil && resPayload == nil {
		// normal errors from the predict process contain a status failed payload
		// this is handling an unexpected case, so failing entirely, at least for now
		rs.Log.Error(err, "unhandled error from predictor process")
		return fmt.Errorf("unhandled error %w from predictor process", err)
	}

	return publishPayload(publisher, resPayload, seldonPuid)
}

func (rs *SeldonRabbitMQServer) createAndPublishErrorResponse(errorArgs ConsumerError, publisher *publisher) error {
	reqPayload := errorArgs.pl

	seldonPuid := assignAndReturnPUID(reqPayload)

	message := &proto.SeldonMessage{
		Status: &proto.Status{
			Code:   0,
			Info:   "Prediction Failed",
			Reason: errorArgs.err.Error(),
			Status: proto.Status_FAILURE,
		},
		Meta: &proto.Meta{
			Puid: seldonPuid,
		},
	}

	var resPayload payload.SeldonPayload
	switch errorArgs.delivery.ContentType {
	case payload.APPLICATION_TYPE_PROTOBUF:
		resPayload = &payload.ProtoPayload{Msg: message}
		break
	case rest.ContentTypeJSON:
		var err error
		resPayload, err = rs.createJsonResponsePayload(message)
		if err != nil {
			rs.Log.Error(err, "error creating json response")
			return fmt.Errorf("error '%w' creating json response", err)
		}
		break
	default: // default to json response if unknown payload
		err := fmt.Errorf("unknown content type %v", errorArgs.delivery.ContentType)
		rs.Log.Error(err, "unexpected content type", "default response content-type", rest.ContentTypeJSON)
		resPayload, err = rs.createJsonResponsePayload(message)
		if err != nil {
			rs.Log.Error(err, "error creating json response")
			return fmt.Errorf("error '%w' creating json response", err)
		}
		break
	}

	return publishPayload(publisher, resPayload, seldonPuid)
}

func (rs *SeldonRabbitMQServer) createJsonResponsePayload(message *proto.SeldonMessage) (payload.SeldonPayload, error) {

	jsonStr, err := new(jsonpb.Marshaler).MarshalToString(message)
	if err != nil {
		rs.Log.Error(err, "error marshaling seldon message")
		return nil, fmt.Errorf("error '%w' marshaling seldon message", err)
	}
	return &payload.BytesPayload{
		Msg:         []byte(jsonStr),
		ContentType: rest.ContentTypeJSON,
	}, nil
}

func assignAndReturnPUID(pl *SeldonPayloadWithHeaders) string {
	if pl == nil {
		return guuid.New().String()
	}
	if pl.Headers == nil {
		pl.Headers = make(map[string][]string)
	}
	if pl.Headers[payload.SeldonPUIDHeader] == nil {
		pl.Headers[payload.SeldonPUIDHeader] = []string{guuid.New().String()}
	}
	return pl.Headers[payload.SeldonPUIDHeader][0]
}

func publishPayload(publisher *publisher, pl payload.SeldonPayload, seldonPuid string) error {
	resHeaders := map[string][]string{payload.SeldonPUIDHeader: {seldonPuid}}
	//TODO might need more headers

	resPayloadWithHeaders := SeldonPayloadWithHeaders{
		pl,
		resHeaders,
	}

	return publisher.Publish(resPayloadWithHeaders)
}
