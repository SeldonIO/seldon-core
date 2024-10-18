/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
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
	"strings"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"github.com/signalfx/splunk-otel-go/instrumentation/github.com/confluentinc/confluent-kafka-go/v2/kafka/splunkkafka"
	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"

	v2 "github.com/seldonio/seldon-core/apis/go/v2/mlops/v2_dataplane"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/envoy/resources"
	kafka2 "github.com/seldonio/seldon-core/scheduler/v2/pkg/kafka"
	pipeline "github.com/seldonio/seldon-core/scheduler/v2/pkg/kafka/pipeline"
	seldontracer "github.com/seldonio/seldon-core/scheduler/v2/pkg/tracing"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/util"
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

func NewInferWorker(
	consumer *InferKafkaHandler,
	logger log.FieldLogger,
	traceProvider *seldontracer.TracerProvider,
	topicNamer *kafka2.TopicNamer,
) (*InferWorker, error) {
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
	grpcClient, err := iw.getGrpcClient(
		consumer.consumerConfig.InferenceServerConfig.Host,
		consumer.consumerConfig.InferenceServerConfig.GrpcPort,
	)
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
		grpc_retry.WithBackoff(grpc_retry.BackoffExponential(util.GRPCRetryBackoff)),
		grpc_retry.WithMax(util.GRPCRetryMaxCount), // retry envoy connection
	}

	var creds credentials.TransportCredentials
	if iw.consumer.tlsClientOptions.TLS {
		logger.Info("Creating TLS credentials")
		creds = iw.consumer.tlsClientOptions.Cert.CreateClientTransportCredentials()
	} else {
		logger.Info("Creating insecure credentials")
		creds = insecure.NewCredentials()
	}

	kacp := keepalive.ClientParameters{
		Time:                util.ClientKeapAliveTime,
		Timeout:             util.ClientKeapAliveTimeout,
		PermitWithoutStream: util.ClientKeapAlivePermit,
	}

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(creds),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(util.GRPCMaxMsgSizeBytes),
			grpc.MaxCallSendMsgSize(util.GRPCMaxMsgSizeBytes),
		),
		grpc.WithStatsHandler(
			otelgrpc.NewClientHandler(),
		),
		grpc.WithUnaryInterceptor(
			grpc_retry.UnaryClientInterceptor(retryOpts...),
		),
		grpc.WithKeepaliveParams(kacp),
	}

	conn, err := grpc.NewClient(fmt.Sprintf("%s:%d", host, port), opts...)
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

func (iw *InferWorker) produce(
	ctx context.Context,
	job *InferWork,
	topic string,
	b []byte,
	errorTopic bool,
	headers map[string][]string,
) error {
	logger := iw.logger.WithField("func", "produce")

	kafkaHeaders := job.msg.Headers
	if errorTopic {
		kafkaHeaders = append(kafkaHeaders, kafka.Header{Key: kafka2.TopicErrorHeader, Value: []byte(job.modelName)})
	}

	for k, vs := range headers {
		for _, v := range vs {
			if !existsKafkaHeader(kafkaHeaders, k, v) {
				logger.Debugf("Adding header to kafka response %s:%s", k, v)
				kafkaHeaders = append(kafkaHeaders, kafka.Header{Key: k, Value: []byte(v)})
			}
		}
	}

	if logger.Logger.IsLevelEnabled(log.DebugLevel) {
		for _, h := range kafkaHeaders {
			logger.Debugf("Adding kafka header for topic %s %s:%s", topic, h.Key, string(h.Value))
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
	requestId := pipeline.GetRequestIdFromKafkaHeaders(kafkaHeaders)
	if requestId == "" {
		logger.Warnf("Missing request id in Kafka headers for key %s", string(job.msg.Key))
	}
	span.SetAttributes(attribute.String(util.RequestIdHeader, requestId))
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

	restUrl := getRestUrl(
		iw.consumer.tlsClientOptions.TLS,
		iw.consumer.consumerConfig.InferenceServerConfig.Host,
		iw.consumer.consumerConfig.InferenceServerConfig.HttpPort,
		job.modelName,
	)

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

	return iw.produce(
		ctx,
		job,
		iw.topicNamer.GetModelTopicOutputs(job.modelName),
		b,
		false,
		extractHeadersHttp(response.Header),
	)
}

// Add all external headers to request metadata
func addMetadataToOutgoingContext(ctx context.Context, job *InferWork, logger log.FieldLogger) context.Context {
	for k, v := range job.headers {
		if strings.HasPrefix(k, resources.ExternalHeaderPrefix) &&
			k != resources.SeldonRouteHeader { // We don;t want to send x-seldon-route as this will confuse envoy
			logger.Debugf("Adding outgoing ctx metadata %s:%s", k, v)
			ctx = metadata.AppendToOutgoingContext(ctx, k, v)
		}
	}
	ctx = metadata.AppendToOutgoingContext(ctx, resources.SeldonModelHeader, job.modelName)
	return ctx
}

func (iw *InferWorker) grpcRequest(ctx context.Context, job *InferWork, req *v2.ModelInferRequest) error {
	logger := iw.logger.WithField("func", "grpcRequest")
	logger.Debugf("gRPC request for %s", job.modelName)
	//Update req with correct modelName
	req.ModelName = job.modelName
	req.ModelVersion = fmt.Sprintf("%d", util.GetPinnedModelVersion())

	ctx = addMetadataToOutgoingContext(ctx, job, logger)

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
	return iw.produce(
		ctx,
		job,
		iw.topicNamer.GetModelTopicOutputs(job.modelName),
		b,
		false,
		extractHeadersGrpc(header, trailer),
	)
}
