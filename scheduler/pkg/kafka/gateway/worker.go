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
	key     []byte
	value   []byte
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
		httpClient: http.DefaultClient,
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
		grpc.WithInsecure(),
		grpc.WithStreamInterceptor(grpc_retry.StreamClientInterceptor(retryOpts...)),
		grpc.WithUnaryInterceptor(grpc_retry.UnaryClientInterceptor(retryOpts...)),
	}
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", host, port), opts...)
	if err != nil {
		return nil, err
	}

	return v2.NewGRPCInferenceServiceClient(conn), nil
}

func getProtoInferRequest(job *InferWork) (*v2.ModelInferRequest, error) {
	ireq := v2.ModelInferRequest{}
	err := proto.Unmarshal(job.value, &ireq)
	if err != nil {
		iresp := v2.ModelInferResponse{}
		err := proto.Unmarshal(job.value, &iresp)
		if err != nil {
			return nil, err
		}
		return chainProtoResponseToRequest(&iresp), nil
	}
	return &ireq, nil
}

func (iw *InferWorker) Start(jobChan <-chan *InferWork, cancelChan <-chan struct{}) {
	for {
		select {
		case <-cancelChan:
			return

		case job := <-jobChan:
			err := iw.processRequest(job)
			if err != nil {
				iw.logger.WithError(err).Errorf("Failed to process request for model %s", iw.consumer.modelConfig.ModelName)
			}
		}
	}
}

func (iw *InferWorker) processRequest(job *InferWork) error {
	// Has Type Header
	if typeValue, ok := job.headers[HeaderKeyType]; ok {
		switch typeValue {
		case HeaderValueJsonReq:
			return iw.restRequest(job, false)
		case HeaderValueJsonRes:
			return iw.restRequest(job, true)
		case HeaderValueProtoReq:
			protoRequest, err := getProtoInferRequest(job)
			if err != nil {
				return err
			}
			return iw.grpcRequest(job, protoRequest)
		case HeaderValueProtoRes:
			protoRequest, err := getProtoRequestAssumingResponse(job.value)
			if err != nil {
				return err
			}
			return iw.grpcRequest(job, protoRequest)
		default:
			return fmt.Errorf("Header %s with unknown type %s", HeaderKeyType, typeValue)
		}
	} else { // Does not have type header - this is the general case to allow easy use
		protoRequest, err := getProtoInferRequest(job)
		if err != nil {
			return iw.restRequest(job, true)
		} else {
			return iw.grpcRequest(job, protoRequest)
		}
	}
}

func (iw *InferWorker) produce(job *InferWork, b []byte, headerType string) error {
	logger := iw.logger.WithField("func", "produce")
	kafkaHeaders := make([]kafka.Header, 1)
	kafkaHeaders[0] = kafka.Header{Key: HeaderKeyType, Value: []byte(headerType)}
	logger.Infof("Produce response to topic %s with header %s", iw.consumer.modelConfig.OutputTopic, headerType)
	err := iw.consumer.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &iw.consumer.modelConfig.OutputTopic, Partition: kafka.PartitionAny},
		Key:            job.key,
		Value:          b,
		Headers:        kafkaHeaders,
	}, nil)
	if err != nil {
		iw.logger.WithError(err).Errorf("Failed to produce response for model %s", iw.consumer.modelConfig.ModelName)
		return err
	}
	return nil
}

func (iw *InferWorker) restRequest(job *InferWork, maybeConvert bool) error {
	if maybeConvert {
		job.value = maybeChainRest(job.value)
	}
	req, err := http.NewRequest("POST", iw.restUrl.String(), bytes.NewBuffer(job.value))
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
		return fmt.Errorf("Failed infer call code:%d", response.StatusCode)
	}
	return iw.produce(job, b, HeaderValueJsonRes)
}

func (iw *InferWorker) grpcRequest(job *InferWork, req *v2.ModelInferRequest) error {
	//Update req with correct modelName
	req.ModelName = iw.consumer.modelConfig.ModelName
	req.ModelVersion = fmt.Sprintf("%d", util.GetPinnedModelVersion())
	ctx := context.TODO()
	ctx = metadata.AppendToOutgoingContext(ctx, resources.SeldonModelHeader, iw.consumer.modelConfig.ModelName)
	resp, err := iw.grpcClient.ModelInfer(ctx, req)
	if err != nil {
		return err
	}
	b, err := proto.Marshal(resp)
	if err != nil {
		return err
	}
	return iw.produce(job, b, HeaderValueProtoRes)
}
