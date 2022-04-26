package gateway

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"google.golang.org/grpc/credentials/insecure"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	kafka2 "github.com/seldonio/seldon-core/scheduler/pkg/kafka"
	"github.com/signalfx/splunk-otel-go/instrumentation/github.com/confluentinc/confluent-kafka-go/kafka/splunkkafka"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"

	"github.com/seldonio/seldon-core/scheduler/pkg/envoy/resources"
	"google.golang.org/grpc/metadata"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	v2 "github.com/seldonio/seldon-core/scheduler/apis/mlops/v2_dataplane"
	"github.com/seldonio/seldon-core/scheduler/pkg/util"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

type InferWorker struct {
	logger     log.FieldLogger
	grpcClient v2.GRPCInferenceServiceClient
	httpClient *http.Client
	restUrl    *url.URL
	consumer   *InferKafkaGateway
}

type InferWork struct {
	headers map[string]string
	msg     *kafka.Message
}

type V2Error struct {
	Error string `json:"error"`
}

func NewInferWorker(consumer *InferKafkaGateway, logger log.FieldLogger) (*InferWorker, error) {
	grpcClient, err := getGrpcClient(consumer.serverConfig.Host, consumer.serverConfig.GrpcPort)
	if err != nil {
		return nil, err
	}
	restUrl := getRestUrl(consumer.serverConfig.Host, consumer.serverConfig.HttpPort, consumer.modelConfig.ModelName)
	return &InferWorker{
		logger:     logger.WithField("source", "KafkaInferWorker"),
		grpcClient: grpcClient,
		restUrl:    restUrl,
		httpClient: &http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)},
		consumer:   consumer,
	}, nil
}

func getRestUrl(host string, port int, modelName string) *url.URL {
	return &url.URL{
		Scheme: "http",
		Host:   net.JoinHostPort(host, strconv.Itoa(port)),
		Path:   fmt.Sprintf("/v2/models/%s/infer", modelName),
	}
}

func getGrpcClient(host string, port int) (v2.GRPCInferenceServiceClient, error) {
	retryOpts := []grpc_retry.CallOption{
		grpc_retry.WithBackoff(grpc_retry.BackoffExponential(100 * time.Millisecond)),
	}
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(grpc_middleware.ChainUnaryClient(grpc_retry.UnaryClientInterceptor(retryOpts...), otelgrpc.UnaryClientInterceptor())),
	}
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", host, port), opts...)
	if err != nil {
		return nil, err
	}

	return v2.NewGRPCInferenceServiceClient(conn), nil
}

func getProtoInferRequest(job *InferWork) (*v2.ModelInferRequest, error) {
	ireq := v2.ModelInferRequest{}
	err := proto.Unmarshal(job.msg.Value, &ireq)
	if err != nil {
		iresp := v2.ModelInferResponse{}
		err := proto.Unmarshal(job.msg.Value, &iresp)
		if err != nil {
			return nil, err
		}
		return chainProtoResponseToRequest(&iresp), nil
	}
	return &ireq, nil
}

// Extract tracing context from Kafka message
func createContextFromKafkaMsg(job *InferWork) context.Context {
	ctx := context.Background()
	carrierIn := splunkkafka.NewMessageCarrier(job.msg)
	ctx = otel.GetTextMapPropagator().Extract(ctx, carrierIn)
	return ctx
}

func (iw *InferWorker) Start(jobChan <-chan *InferWork, cancelChan <-chan struct{}) {
	for {
		select {
		case <-cancelChan:
			return

		case job := <-jobChan:
			ctx := createContextFromKafkaMsg(job)
			err := iw.processRequest(ctx, job)
			if err != nil {
				iw.logger.WithError(err).Errorf("Failed to process request for model %s", iw.consumer.modelConfig.ModelName)
			}
		}
	}
}

func (iw *InferWorker) processRequest(ctx context.Context, job *InferWork) error {
	// Has Type Header
	if typeValue, ok := job.headers[HeaderKeyType]; ok {
		switch typeValue {
		case HeaderValueJsonReq:
			return iw.restRequest(ctx, job, false)
		case HeaderValueJsonRes:
			return iw.restRequest(ctx, job, true)
		case HeaderValueProtoReq:
			protoRequest, err := getProtoInferRequest(job)
			if err != nil {
				return err
			}
			return iw.grpcRequest(ctx, job, protoRequest)
		case HeaderValueProtoRes:
			protoRequest, err := getProtoRequestAssumingResponse(job.msg.Value)
			if err != nil {
				return err
			}
			return iw.grpcRequest(ctx, job, protoRequest)
		default:
			return fmt.Errorf("Header %s with unknown type %s", HeaderKeyType, typeValue)
		}
	} else { // Does not have type header - this is the general case to allow easy use
		protoRequest, err := getProtoInferRequest(job)
		if err != nil {
			return iw.restRequest(ctx, job, true)
		} else {
			return iw.grpcRequest(ctx, job, protoRequest)
		}
	}
}

func (iw *InferWorker) produce(ctx context.Context, job *InferWork, topic string, b []byte) error {
	logger := iw.logger.WithField("func", "produce")

	var kafkaHeaders []kafka.Header
	//kafkaHeaders = append(kafkaHeaders, kafka.Header{Key: HeaderKeyType, Value: []byte(headerType)})
	if pipelineName, ok := job.headers[resources.SeldonPipelineHeader]; ok {
		logger.Debugf("Adding pipeline header %s:%s", resources.SeldonPipelineHeader, pipelineName)
		kafkaHeaders = append(kafkaHeaders, kafka.Header{Key: resources.SeldonPipelineHeaderSuffix, Value: []byte(pipelineName)})
	}
	if topic == iw.consumer.modelConfig.ErrorTopic {
		kafkaHeaders = append(kafkaHeaders, kafka.Header{Key: kafka2.TopicErrorHeader, Value: []byte("")})
	}
	logger.Infof("Produce response to topic %s", topic)

	msg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Key:            job.msg.Key,
		Value:          b,
		Headers:        kafkaHeaders,
	}

	carrierOut := splunkkafka.NewMessageCarrier(msg)
	otel.GetTextMapPropagator().Inject(ctx, carrierOut)

	err := iw.consumer.producer.Produce(msg, nil)
	if err != nil {
		iw.logger.WithError(err).Errorf("Failed to produce response for model %s", topic)
		return err
	}
	return nil
}

func (iw *InferWorker) restRequest(ctx context.Context, job *InferWork, maybeConvert bool) error {
	logger := iw.logger.WithField("func", "restRequest")

	data := job.msg.Value
	if maybeConvert {
		data = maybeChainRest(job.msg.Value)
	}
	req, err := http.NewRequestWithContext(ctx, "POST", iw.restUrl.String(), bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(resources.SeldonModelHeader, iw.consumer.modelConfig.ModelName)
	response, err := iw.httpClient.Do(req)
	if err != nil {
		return err
	}
	b, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}
	err = response.Body.Close()
	if err != nil {
		return err
	}
	iw.logger.Infof("v2 server response: %s", b)
	if response.StatusCode != http.StatusOK {
		logger.Warnf("Failed infer request with status code %d and payload %s", response.StatusCode, string(b))
		return iw.produce(ctx, job, iw.consumer.modelConfig.ErrorTopic, b)
	}
	return iw.produce(ctx, job, iw.consumer.modelConfig.OutputTopic, b)
}

func (iw *InferWorker) grpcRequest(ctx context.Context, job *InferWork, req *v2.ModelInferRequest) error {
	logger := iw.logger.WithField("func", "grpcRequest")
	//Update req with correct modelName
	req.ModelName = iw.consumer.modelConfig.ModelName
	req.ModelVersion = fmt.Sprintf("%d", util.GetPinnedModelVersion())

	ctx = metadata.AppendToOutgoingContext(ctx, resources.SeldonModelHeader, iw.consumer.modelConfig.ModelName)
	resp, err := iw.grpcClient.ModelInfer(ctx, req)
	if err != nil {
		logger.WithError(err).Warnf("Failed infer request")
		return iw.produce(ctx, job, iw.consumer.modelConfig.ErrorTopic, []byte(err.Error()))
	}
	b, err := proto.Marshal(resp)
	if err != nil {
		return err
	}
	return iw.produce(ctx, job, iw.consumer.modelConfig.OutputTopic, b)
}
