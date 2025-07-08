/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package pipeline

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"sync"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/signalfx/splunk-otel-go/instrumentation/github.com/confluentinc/confluent-kafka-go/v2/kafka/splunkkafka"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	kafka_config "github.com/seldonio/seldon-core/components/kafka/v2/pkg/config"
	config_tls "github.com/seldonio/seldon-core/components/tls/v2/pkg/config"

	kafka2 "github.com/seldonio/seldon-core/scheduler/v2/pkg/kafka"
	seldontracer "github.com/seldonio/seldon-core/scheduler/v2/pkg/tracing"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/util"
)

const (
	pollTimeoutMillisecs = 10000
)

type PipelineInferer interface {
	StorePipeline(resourceName string, isModel bool) error
	LoadPipeline(resourceName string, isModel bool) (*Pipeline, error)
	Infer(
		ctx context.Context,
		resourceName string,
		isModel bool,
		data []byte,
		headers []kafka.Header,
		requestId string,
	) (*Request, error)
}

type KafkaManager struct {
	kafkaConfig     *kafka_config.KafkaConfig
	producer        *kafka.Producer
	pipelines       sync.Map
	logger          logrus.FieldLogger
	mu              sync.RWMutex
	topicNamer      *kafka2.TopicNamer
	tracer          trace.Tracer
	consumerManager *ConsumerManager
}

type Pipeline struct {
	resourceName string
	consumer     *MultiTopicsKafkaConsumer
	isModel      bool
	wg           *sync.WaitGroup
}

type Request struct {
	mu         sync.Mutex
	active     bool
	wg         *sync.WaitGroup
	key        string
	response   []byte
	headers    []kafka.Header
	isError    bool
	errorModel string
}

func NewKafkaManager(
	logger logrus.FieldLogger,
	namespace string,
	kafkaConfig *kafka_config.KafkaConfig,
	traceProvider *seldontracer.TracerProvider,
	maxNumConsumers int,
) (*KafkaManager, error) {
	topicNamer, err := kafka2.NewTopicNamer(namespace, kafkaConfig.TopicPrefix)
	if err != nil {
		return nil, err
	}

	tracer := traceProvider.GetTraceProvider().Tracer("KafkaManager")
	km := &KafkaManager{
		kafkaConfig:     kafkaConfig,
		logger:          logger.WithField("source", "KafkaManager"),
		topicNamer:      topicNamer,
		tracer:          tracer,
		consumerManager: NewConsumerManager(namespace, logger, kafkaConfig, maxNumConsumers, tracer),
		mu:              sync.RWMutex{},
	}

	err = km.createProducer()
	if err != nil {
		return nil, err
	}

	return km, nil
}

func (km *KafkaManager) Stop() {
	logger := km.logger.WithField("func", "Stop")
	logger.Info("Stopping pipelines")

	km.mu.Lock()
	defer km.mu.Unlock()

	km.producer.Close()
	km.consumerManager.Stop()
	logger.Info("Stopped all pipelines")
}

func (km *KafkaManager) createProducer() error {
	if km.producer != nil {
		km.producer.Close()
	}
	var err error

	producerConfigMap := kafka_config.CloneKafkaConfigMap(km.kafkaConfig.Producer)
	producerConfigMap["go.delivery.reports"] = true
	err = config_tls.AddKafkaSSLOptions(producerConfigMap)
	if err != nil {
		return err
	}

	configWithoutSecrets := kafka_config.WithoutSecrets(producerConfigMap)
	km.logger.Infof("Creating producer with config %v", configWithoutSecrets)

	km.producer, err = kafka.NewProducer(&producerConfigMap)
	return err
}

func (km *KafkaManager) createPipeline(resource string, isModel bool) (*Pipeline, error) {
	consumer, err := km.consumerManager.getKafkaConsumer(resource, isModel)
	if err != nil {
		return nil, err
	}
	return &Pipeline{
		resourceName: resource,
		consumer:     consumer,
		isModel:      isModel,
		wg:           new(sync.WaitGroup),
	}, nil
}

func getPipelineKey(resourceName string, isModel bool) string {
	if isModel {
		return fmt.Sprintf("%s.model", resourceName)
	} else {
		return fmt.Sprintf("%s.pipeline", resourceName)
	}
}

func loadPipeline(resourceName string, isModel bool, pipelines *sync.Map) (*Pipeline, error) {
	key := getPipelineKey(resourceName, isModel)
	if val, ok := pipelines.Load(key); ok {
		return val.(*Pipeline), nil
	}
	return nil, fmt.Errorf("pipeline for resource %s not found", resourceName)
}

func (km *KafkaManager) LoadPipeline(resourceName string, isModel bool) (*Pipeline, error) {
	km.mu.RLock()
	defer km.mu.RUnlock()
	return loadPipeline(resourceName, isModel, &km.pipelines)
}

func (km *KafkaManager) StorePipeline(resourceName string, isModel bool) error {
	km.mu.Lock()
	defer km.mu.Unlock()

	_, err := loadPipeline(resourceName, isModel, &km.pipelines)
	if err == nil {
		return nil
	}

	logger := km.logger.WithField("func", "StorePipeline")
	pipeline, err := km.createPipeline(resourceName, isModel)
	if err != nil {
		return err
	}

	pipeline.wg.Add(1) // wait set to allow consumer to say when started

	key := getPipelineKey(resourceName, isModel)
	km.pipelines.Store(key, pipeline)

	go func() {
		err := km.consume(pipeline)
		if err != nil {
			km.logger.WithError(err).Errorf("Failed running consumer for resource %s", resourceName)
		}
	}()

	logger.Debugf("Waiting for consumer to be ready for %s", resourceName)
	pipeline.wg.Wait() // wait (maybe) for consumer start
	return nil
}

func (km *KafkaManager) Infer(
	ctx context.Context,
	resourceName string,
	isModel bool,
	data []byte,
	headers []kafka.Header,
	requestId string,
) (*Request, error) {
	logger := km.logger.WithField("func", "Infer")

	var (
		pipeline *Pipeline
		err      error
	)
	pipeline, err = km.LoadPipeline(resourceName, isModel)
	if err != nil {
		if isModel {
			// We only allow lazy loading of model. This path should not be reached
			// through envoy but only sending requests directly to the pipeline gateway.
			// Note that due to lazy loading, having multiple replicas of the pipelinegw
			// may result in unexpected behaviour. This is because due to the load balancing
			// of the requests, consumers with the same group id may end up consuming from
			// different topics.
			logger.Warn("Lazy loading of the models is only suppoerted for a single pipeline gateway instance.")
			err := km.StorePipeline(resourceName, true)
			if err != nil {
				return nil, fmt.Errorf("failed to store model pipeline %s: %w", resourceName, err)
			}

			// In this case, the existence of the pipeline is guaranteed
			pipeline, _ = km.LoadPipeline(resourceName, true)
		} else {
			return nil, err
		}
	}

	// Randomly select a partition to produce the message to
	km.mu.RLock()
	partitions := pipeline.consumer.partitions
	km.mu.RUnlock()
	if len(partitions) == 0 {
		return nil, fmt.Errorf("no partitions assigned for topic %s", resourceName)
	}

	partition := partitions[rand.Intn(len(partitions))]
	logger.Debugf("Using partition %d for resource %s", partition, resourceName)

	// Use composite key to differentiate multiple pipelines (i.e. mirror) using the same message
	// Note that we add the partition to the key to ensure that the message will be sent to
	// a partition for which the consumer is subscribed. For modelgw, it is enought to send the
	// message to the same partition as the one we read from. For dataflow engine on the other hand,
	// we need to read the partition from the request id.
	compositeKey := getCompositeKey(strconv.Itoa(int(partition)), resourceName, requestId, ".")
	request := &Request{
		active: true,
		wg:     new(sync.WaitGroup),
		key:    compositeKey,
	}
	pipeline.consumer.requests.Set(compositeKey, request)
	defer pipeline.consumer.requests.Remove(compositeKey)
	request.wg.Add(1)

	inputTopic := km.topicNamer.GetPipelineTopicInputs(resourceName)
	if isModel {
		inputTopic = km.topicNamer.GetModelTopicInputs(resourceName)
	}
	logger.Debugf("Produce on topic %s with key %s", inputTopic, compositeKey)
	kafkaHeaders := append(headers, kafka.Header{Key: util.SeldonPipelineHeader, Value: []byte(resourceName)})
	kafkaHeaders = addRequestIdToKafkaHeadersIfMissing(kafkaHeaders, requestId)

	msg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &inputTopic,
			Partition: partition,
		},
		Key:     []byte(compositeKey),
		Value:   data,
		Headers: kafkaHeaders,
	}

	ctx, span := km.tracer.Start(ctx, "Produce")
	span.SetAttributes(attribute.String(util.RequestIdHeader, requestId))
	// Add trace headers
	carrier := splunkkafka.NewMessageCarrier(msg)
	otel.GetTextMapPropagator().Inject(ctx, carrier)

	deliveryChan := make(chan kafka.Event)
	err = km.producer.Produce(msg, deliveryChan)
	if err != nil {
		span.End()
		return nil, err
	}
	go func() {
		evt := <-deliveryChan
		logger.Infof("Received delivery event %s", evt.String())
		span.End()
	}()
	logger.Debugf("Waiting for response for request id %s for resource %s", requestId, resourceName)
	request.wg.Wait()
	logger.Debugf("Got response for request id %s for resource %s", requestId, resourceName)
	return request, nil
}

func extractErrorHeader(headers []kafka.Header) (string, bool) {
	for _, header := range headers {
		if header.Key == kafka2.TopicErrorHeader {
			return string(header.Value), true
		}
	}
	return "", false
}

func createResponseErrorPayload(modelName string, response []byte) []byte {
	return append([]byte(modelName+" : "), response...)
}

func createRebalanceCb(km *KafkaManager, pipeline *Pipeline) kafka.RebalanceCb {
	logger := km.logger.WithField("func", "createRebalanceCb")
	return func(consumer *kafka.Consumer, ev kafka.Event) error {
		switch e := ev.(type) {
		case kafka.AssignedPartitions:
			km.mu.Lock()
			km.mu.Unlock()

			logger.Debug("Rebalance: Assigned partitions:", e.Partitions)
			err := consumer.Assign(e.Partitions)
			if err != nil {
				pipeline.consumer.partitions = nil
				return fmt.Errorf("assign error: %w", err)
			}

			// Update the pipeline consumer partitions
			pipeline.consumer.partitions = make([]int32, len(e.Partitions))
			for i, partition := range e.Partitions {
				pipeline.consumer.partitions[i] = partition.Partition
			}
		case kafka.RevokedPartitions:
			km.mu.Lock()
			km.mu.Unlock()

			logger.Debug("Rebalance: Revoked partitions:", e.Partitions)
			err := consumer.Unassign()
			pipeline.consumer.partitions = nil
			if err != nil {
				return fmt.Errorf("unassign error: %w", err)
			}
		}
		return nil
	}
}

func (km *KafkaManager) consume(pipeline *Pipeline) error {
	logger := km.logger.WithField("func", "consume")
	topicName := km.topicNamer.GetPipelineTopicOutputs(pipeline.resourceName)
	if pipeline.isModel {
		topicName = km.topicNamer.GetModelTopicOutputs(pipeline.resourceName)
	}
	err := pipeline.consumer.AddTopic(topicName, createRebalanceCb(km, pipeline))
	pipeline.wg.Done()
	logger.Infof("Topic %s added in consumer id %s", topicName, pipeline.consumer.id)
	if err != nil {
		return err
	}
	return nil
}
