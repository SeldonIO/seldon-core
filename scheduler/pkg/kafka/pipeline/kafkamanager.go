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
	"sync"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/signalfx/splunk-otel-go/instrumentation/github.com/confluentinc/confluent-kafka-go/v2/kafka/splunkkafka"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	kafka_config "github.com/seldonio/seldon-core/components/kafka/v2/pkg/config"
	config_tls "github.com/seldonio/seldon-core/components/tls/v2/pkg/config"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/envoy/resources"
	kafka2 "github.com/seldonio/seldon-core/scheduler/v2/pkg/kafka"
	seldontracer "github.com/seldonio/seldon-core/scheduler/v2/pkg/tracing"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/util"
)

const (
	pollTimeoutMillisecs = 10000
)

type PipelineInferer interface {
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
	maxNumConsumers,
	maxNumTopicsPerConsumer int,
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
		consumerManager: NewConsumerManager(namespace, logger, kafkaConfig, maxNumTopicsPerConsumer, maxNumConsumers, tracer),
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
	consumer, err := km.consumerManager.getKafkaConsumer()
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

func (km *KafkaManager) loadOrStorePipeline(resourceName string, isModel bool) (*Pipeline, error) {
	logger := km.logger.WithField("func", "loadOrStorePipeline")
	key := getPipelineKey(resourceName, isModel)
	if val, ok := km.pipelines.Load(key); ok {
		val.(*Pipeline).wg.Wait()
		return val.(*Pipeline), nil
	} else {
		pipeline, err := km.createPipeline(resourceName, isModel)
		if err != nil {
			return nil, err
		}
		pipeline.wg.Add(1) // wait set to allow consumer to say when started

		val, loaded := km.pipelines.LoadOrStore(key, pipeline)
		if loaded { // we can still have a race condition where multiple "create" are happening, we have to store the first one
			pipeline = val.(*Pipeline)
		} else {
			go func() {
				err := km.consume(pipeline)
				if err != nil {
					km.logger.WithError(err).Errorf("Failed running consumer for resource %s", resourceName)
				}
			}()
		}

		logger.Debugf("Waiting for consumer to be ready for %s", resourceName)
		pipeline.wg.Wait() // wait (maybe) for consumer start
		return pipeline, nil
	}
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
	km.mu.RLock()
	pipeline, err := km.loadOrStorePipeline(resourceName, isModel)
	if err != nil {
		km.mu.RUnlock()
		return nil, err
	}
	// Use composite key to differentiate multiple pipelines (i.e. mirror) using the same message
	compositeKey := getCompositeKey(resourceName, requestId, ".")
	request := &Request{
		active: true,
		wg:     new(sync.WaitGroup),
		key:    compositeKey,
	}
	pipeline.consumer.requests.Set(compositeKey, request)
	defer pipeline.consumer.requests.Remove(compositeKey)
	request.wg.Add(1)

	outputTopic := km.topicNamer.GetPipelineTopicInputs(resourceName)
	if isModel {
		outputTopic = km.topicNamer.GetModelTopicInputs(resourceName)
	}
	logger.Debugf("Produce on topic %s with key %s", outputTopic, compositeKey)
	kafkaHeaders := append(headers, kafka.Header{Key: resources.SeldonPipelineHeader, Value: []byte(resourceName)})
	kafkaHeaders = addRequestIdToKafkaHeadersIfMissing(kafkaHeaders, requestId)

	msg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &outputTopic, Partition: kafka.PartitionAny},
		Key:            []byte(compositeKey),
		Value:          data,
		Headers:        kafkaHeaders,
	}

	ctx, span := km.tracer.Start(ctx, "Produce")
	span.SetAttributes(attribute.String(util.RequestIdHeader, requestId))
	// Add trace headers
	carrier := splunkkafka.NewMessageCarrier(msg)
	otel.GetTextMapPropagator().Inject(ctx, carrier)

	deliveryChan := make(chan kafka.Event)
	err = km.producer.Produce(msg, deliveryChan)
	if err != nil {
		km.mu.RUnlock()
		span.End()
		return nil, err
	}
	go func() {
		evt := <-deliveryChan
		logger.Infof("Received delivery event %s", evt.String())
		span.End()
	}()
	km.mu.RUnlock()
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

func (km *KafkaManager) consume(pipeline *Pipeline) error {
	logger := km.logger.WithField("func", "consume")
	topicName := km.topicNamer.GetPipelineTopicOutputs(pipeline.resourceName)
	if pipeline.isModel {
		topicName = km.topicNamer.GetModelTopicOutputs(pipeline.resourceName)
	}
	err := pipeline.consumer.AddTopic(topicName, nil)
	pipeline.wg.Done()
	logger.Infof("Topic %s added in consumer id %s", topicName, pipeline.consumer.id)
	if err != nil {
		return err
	}
	return nil
}
