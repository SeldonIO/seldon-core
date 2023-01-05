/*
Copyright 2022 Seldon Technologies Ltd.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package gateway

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"google.golang.org/grpc/credentials"

	seldontracer "github.com/seldonio/seldon-core/scheduler/v2/pkg/tracing"
	"go.opentelemetry.io/otel/attribute"

	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/credentials/insecure"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	kafka2 "github.com/seldonio/seldon-core/scheduler/v2/pkg/kafka"
	"github.com/signalfx/splunk-otel-go/instrumentation/github.com/confluentinc/confluent-kafka-go/kafka/splunkkafka"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/envoy/resources"
	"google.golang.org/grpc/metadata"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	v2 "github.com/seldonio/seldon-core/apis/go/v2/mlops/v2_dataplane"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/util"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

type InferWorker struct {
	logger      log.FieldLogger
	grpcClient  v2.GRPCInferenceServiceClient
	httpClient  *http.Client
	consumer    *InferKafkaHandler
	tracer      trace.Tracer
	callOptions []grpc.CallOption
	topicNamer  *kafka2.TopicNamer
}

type InferWork struct {
	modelName string
	headers   map[string]string
	msg       *kafka.Message
}

type V2Error struct {
	Error string `json:"error"`
}

func NewInferWorker(consumer *InferKafkaHandler, logger log.FieldLogger, traceProvider *seldontracer.TracerProvider, topicNamer *kafka2.TopicNamer) (*InferWorker, error) {
	opts := []grpc.CallOption{
		grpc.MaxCallSendMsgSize(math.MaxInt32),
		grpc.MaxCallRecvMsgSize(math.MaxInt32),
	}
	iw := &InferWorker{
		logger:      logger.WithField("source", "KafkaInferWorker"),
		httpClient:  util.GetHttpClientFromTLSOptions(consumer.tlsClientOptions),
		consumer:    consumer,
		tracer:      traceProvider.GetTraceProvider().Tracer("Worker"),
		callOptions: opts,
		topicNamer:  topicNamer,
	}
	// Create gRPC clients
	grpcClient, err := iw.getGrpcClient(consumer.consumerConfig.InferenceServerConfig.Host, consumer.consumerConfig.InferenceServerConfig.GrpcPort)
	if err != nil {
		return nil, err
	}
	iw.grpcClient = grpcClient

	return iw, nil
}

func getRestUrl(tls bool, host string, port int, modelName string) *url.URL {
	scheme := "http"
	if tls {
		scheme = "https"
	}
	return &url.URL{
		Scheme: scheme,
		Host:   net.JoinHostPort(host, strconv.Itoa(port)),
		Path:   fmt.Sprintf("/v2/models/%s/infer", modelName),
	}
}

func (iw *InferWorker) getGrpcClient(host string, port int) (v2.GRPCInferenceServiceClient, error) {
	logger := iw.logger.WithField("func", "getGrpcClient")
	retryOpts := []grpc_retry.CallOption{
		grpc_retry.WithBackoff(grpc_retry.BackoffExponential(util.GrpcRetryBackoffMillisecs * time.Millisecond)),
		grpc_retry.WithMax(util.GrpcRetryMaxCount),  // retry envoy connection
	}
	var creds credentials.TransportCredentials
	if iw.consumer.tlsClientOptions.TLS {
		logger.Info("Creating TLS credentials")
		creds = iw.consumer.tlsClientOptions.Cert.CreateClientTransportCredentials()
	} else {
		logger.Info("Creating insecure credentials")
		creds = insecure.NewCredentials()
	}
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(creds),
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
				iw.logger.WithError(err).Errorf("Failed to process request for model %s", job.modelName)
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

func existsKafkaHeader(headers []kafka.Header, key string, val string) bool {
	for _, header := range headers {
		if header.Key == key && string(header.Value) == val {
			return true
		}
	}
	return false
}

func (iw *InferWorker) produce(ctx context.Context, job *InferWork, topic string, b []byte, errorTopic bool, headers map[string][]string) error {
	logger := iw.logger.WithField("func", "produce")

	kafkaHeaders := job.msg.Headers
	if errorTopic {
		kafkaHeaders = append(kafkaHeaders, kafka.Header{Key: kafka2.TopicErrorHeader, Value: []byte(job.modelName)})
	}
	for k, vs := range headers {
		for _, v := range vs {
			if !existsKafkaHeader(kafkaHeaders, k, v) {
				logger.Infof("Adding header to kafka response %s:%s", k, v)
				kafkaHeaders = append(kafkaHeaders, kafka.Header{Key: k, Value: []byte(v)})
			}
		}
	}
	logger.Infof("Produce response to topic %s", topic)

	msg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Key:            job.msg.Key,
		Value:          b,
		Headers:        kafkaHeaders,
	}

	ctx, span := iw.tracer.Start(ctx, "Produce")
	span.SetAttributes(attribute.String(util.RequestIdHeader, string(job.msg.Key)))
	carrierOut := splunkkafka.NewMessageCarrier(msg)
	otel.GetTextMapPropagator().Inject(ctx, carrierOut)

	deliveryChan := make(chan kafka.Event)
	err := iw.consumer.Produce(msg, deliveryChan)
	if err != nil {
		iw.logger.WithError(err).Errorf("Failed to produce response for model %s", topic)
		return err
	}
	go func() {
		<-deliveryChan
		span.End()
	}()

	return nil
}

func (iw *InferWorker) restRequest(ctx context.Context, job *InferWork, maybeConvert bool) error {
	logger := iw.logger.WithField("func", "restRequest")
	restUrl := getRestUrl(iw.consumer.tlsClientOptions.TLS, iw.consumer.consumerConfig.InferenceServerConfig.Host, iw.consumer.consumerConfig.InferenceServerConfig.HttpPort, job.modelName)
	logger.Debugf("REST request to %s for %s", restUrl.String(), job.modelName)
	data := job.msg.Value
	if maybeConvert {
		data = maybeChainRest(job.msg.Value)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, restUrl.String(), bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(resources.SeldonModelHeader, job.modelName)
	if reqId, ok := job.headers[util.RequestIdHeader]; ok {
		req.Header[util.RequestIdHeader] = []string{reqId}
	}
	response, err := iw.httpClient.Do(req)
	if err != nil {
		return err
	}
	b, err := io.ReadAll(response.Body)
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
		return iw.produce(ctx, job, iw.topicNamer.GetModelErrorTopic(), b, true, nil)
	}
	return iw.produce(ctx, job, iw.topicNamer.GetModelTopicOutputs(job.modelName), b, false, extractHeadersHttp(response.Header))
}

func (iw *InferWorker) grpcRequest(ctx context.Context, job *InferWork, req *v2.ModelInferRequest) error {
	logger := iw.logger.WithField("func", "grpcRequest")
	logger.Debugf("gRPC request for %s", job.modelName)
	//Update req with correct modelName
	req.ModelName = job.modelName
	req.ModelVersion = fmt.Sprintf("%d", util.GetPinnedModelVersion())

	ctx = metadata.AppendToOutgoingContext(ctx, resources.SeldonModelHeader, job.modelName)
	if reqId, ok := job.headers[util.RequestIdHeader]; ok {
		ctx = metadata.AppendToOutgoingContext(ctx, util.RequestIdHeader, reqId)
	}
	var header, trailer metadata.MD
	opts := append(iw.callOptions, grpc.Header(&header))
	opts = append(opts, grpc.Trailer(&trailer))
	resp, err := iw.grpcClient.ModelInfer(ctx, req, opts...)
	if err != nil {
		logger.WithError(err).Warnf("Failed infer request")
		return iw.produce(ctx, job, iw.topicNamer.GetModelErrorTopic(), []byte(err.Error()), true, nil)
	}
	b, err := proto.Marshal(resp)
	if err != nil {
		return err
	}
	return iw.produce(ctx, job, iw.topicNamer.GetModelTopicOutputs(job.modelName), b, false, extractHeadersGrpc(header, trailer))
}
