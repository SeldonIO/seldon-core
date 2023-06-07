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

package pipeline

import (
	"context"
	"fmt"
	"sync"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/signalfx/splunk-otel-go/instrumentation/github.com/confluentinc/confluent-kafka-go/kafka/splunkkafka"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/envoy/resources"
	kafka2 "github.com/seldonio/seldon-core/scheduler/v2/pkg/kafka"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/kafka/config"
	seldontracer "github.com/seldonio/seldon-core/scheduler/v2/pkg/tracing"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/util"
)

const (
	pollTimeoutMillisecs = 10000
)

type PipelineInferer interface {
	Infer(ctx context.Context, resourceName string, isModel bool, data []byte, headers []kafka.Header, requestId string) (*Request, error)
}

type KafkaManager struct {
	kafkaConfig     *config.KafkaConfig
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

func NewKafkaManager(logger logrus.FieldLogger, namespace string, kafkaConfig *config.KafkaConfig,
	traceProvider *seldontracer.TracerProvider, maxNumConsumers, maxNumTopicsPerConsumer int) (*KafkaManager, error) {
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
		consumerManager: NewConsumerManager(logger, kafkaConfig, maxNumTopicsPerConsumer, maxNumConsumers, tracer),
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

	producerConfigMap := config.CloneKafkaConfigMap(km.kafkaConfig.Producer)
	producerConfigMap["go.delivery.reports"] = true
	err = config.AddKafkaSSLOptions(producerConfigMap)
	if err != nil {
		return err
	}
	km.logger.Infof("Creating producer with config %v", producerConfigMap)
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

func (km *KafkaManager) Infer(ctx context.Context, resourceName string, isModel bool, data []byte, headers []kafka.Header, requestId string) (*Request, error) {
	logger := km.logger.WithField("func", "Infer")
	km.mu.RLock()
	pipeline, err := km.loadOrStorePipeline(resourceName, isModel)
	if err != nil {
		km.mu.RUnlock()
		return nil, err
	}
	key := requestId
	request := &Request{
		active: true,
		wg:     new(sync.WaitGroup),
		key:    key,
	}
	pipeline.consumer.requests.Set(key, request)
	defer pipeline.consumer.requests.Remove(key)
	request.wg.Add(1)

	outputTopic := km.topicNamer.GetPipelineTopicInputs(resourceName)
	if isModel {
		outputTopic = km.topicNamer.GetModelTopicInputs(resourceName)
	}
	logger.Debugf("Produce on topic %s with key %s", outputTopic, key)
	kafkaHeaders := append(headers, kafka.Header{Key: resources.SeldonPipelineHeader, Value: []byte(resourceName)})
	kafkaHeaders = addRequestIdToKafkaHeadersIfMissing(kafkaHeaders, requestId)

	msg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &outputTopic, Partition: kafka.PartitionAny},
		Key:            []byte(key),
		Value:          data,
		Headers:        kafkaHeaders,
	}

	ctx, span := km.tracer.Start(ctx, "Produce")
	span.SetAttributes(attribute.String(util.RequestIdHeader, key))
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
	logger.Debugf("Waiting for response for key %s", key)
	request.wg.Wait()
	logger.Debugf("Got response for key %s", key)
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
