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
	"time"

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
	pollTimeoutMillisecs     = 10000
	timeoutWaitForPartitions = time.Second * 10
)

type PipelineInferer interface {
	LoadOrStorePipeline(resourceName string, isModel bool, loadOnly bool) (*Pipeline, error)
	DeletePipeline(resourceName string, isModel bool) error
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
	partition  int32
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

func (km *KafkaManager) DeletePipeline(resourceName string, isModel bool) error {
	logger := km.logger.WithField("func", "DeletePipeline")
	key := getPipelineKey(resourceName, isModel)

	km.mu.Lock()
	defer km.mu.Unlock()

	if val, ok := km.pipelines.Load(key); ok {
		pipeline := val.(*Pipeline)
		err := pipeline.consumer.RemoveTopic(
			km.topicNamer.GetPipelineTopicOutputs(resourceName),
			createRebalanceCb(km, pipeline.consumer),
		)
		if err != nil {
			logger.WithError(err).Errorf("Failed to remove topic for resource %s", resourceName)
			return err
		}

		// If the consumer has no topics left, we can remove it from the map
		// to avoid reusing a closed consumer.
		if len(pipeline.consumer.topics) == 0 {
			if pipeline.isModel {
				delete(km.consumerManager.modelsConsumers, pipeline.consumer.id)
			} else {
				delete(km.consumerManager.pipelinesConsumers, pipeline.consumer.id)
			}
		}

		km.pipelines.Delete(key)
		logger.Infof("Deleted pipeline %s", resourceName)
	} else {
		logger.Warnf("No pipeline found for resource %s", resourceName)
	}
	return nil
}

func (km *KafkaManager) LoadOrStorePipeline(resourceName string, isModel bool, loadOnly bool) (*Pipeline, error) {
	logger := km.logger.WithField("func", "loadOrStorePipeline")
	key := getPipelineKey(resourceName, isModel)

	// try to load the pipeline from the map
	km.mu.RLock()
	if val, ok := km.pipelines.Load(key); ok {
		km.mu.RUnlock()
		val.(*Pipeline).wg.Wait()
		return val.(*Pipeline), nil
	}
	km.mu.RUnlock()

	// don't create a new pipeline if loadOnly is true. In case of invalid envoy
	// routes, we don't want to create a new pipeline on the wrong replica.
	if !isModel && loadOnly {
		return nil, fmt.Errorf("pipeline %s not found", resourceName)
	}

	// acquire write lock to potentially create and store
	km.mu.Lock()
	defer km.mu.Unlock()

	// check again in case another goroutine stored it
	if val, ok := km.pipelines.Load(key); ok {
		val.(*Pipeline).wg.Wait()
		return val.(*Pipeline), nil
	}

	// create new pipeline
	pipeline, err := km.createPipeline(resourceName, isModel)
	if err != nil {
		return nil, err
	}
	pipeline.wg.Add(1) // wait set to allow consumer to say when started
	km.pipelines.Store(key, pipeline)

	go func() {
		err := km.consume(pipeline)
		if err != nil {
			km.logger.WithError(err).Errorf("Failed running consumer for resource %s", resourceName)
		}
	}()

	logger.Debugf("Waiting for consumer to be ready for %s", resourceName)
	pipeline.wg.Wait() // wait (maybe) for consumer start
	return pipeline, nil
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

	pipeline, err := km.LoadOrStorePipeline(resourceName, isModel, true)
	if err != nil {
		return nil, err
	}

	// We lock here since the partition assignment can change in rebalance
	pipeline.consumer.rebalanceMu.RLock()
	partitions := pipeline.consumer.partitions
	if len(partitions) == 0 {
		listener := pipeline.consumer.partitionsReady.Subscribe()
		// we must unlock to allow the rebalance callback to notify us when partitions are available
		pipeline.consumer.rebalanceMu.RUnlock()

		logger.WithField("resource_name", resourceName).Info("Waiting for partition to be available")
		select {
		case <-listener:
			pipeline.consumer.rebalanceMu.RLock()
		case <-time.After(timeoutWaitForPartitions):
			return nil, fmt.Errorf("timed out waiting for partitions to be assigned to consumer for pipeline %s", resourceName)
		}

		logger.WithField("resource_name", resourceName).Info("Received signal, partition ready")
	}

	// Randomly select a partition to produce the message to
	partition := partitions[rand.Intn(len(partitions))]
	logger.Debugf("Using partition %d for resource %s", partition, resourceName)

	// Use composite key to differentiate multiple pipelines (i.e. mirror) using the same message
	// Note that we add the partition to the key to ensure that the message will be sent to
	// a partition for which the consumer is subscribed. For modelgw, it is enought to send the
	// message to the same partition as the one we read from. For dataflow engine on the other hand,
	// we need to read the partition from the request id.
	compositeKey := getCompositeKey(strconv.Itoa(int(partition)), resourceName, requestId, ".")
	request := &Request{
		active:    true,
		wg:        new(sync.WaitGroup),
		key:       compositeKey,
		partition: partition,
	}
	pipeline.consumer.requests.Set(compositeKey, request)
	defer pipeline.consumer.requests.Remove(compositeKey)
	request.wg.Add(1)

	// We release the lock here in case a rebalance happens while we are producing the message.
	// The rebalance callback function will invalidate the request if the partition is revoked.
	// Note that we cannot hold the lock until the end of the function because the poll function
	// may call the rebalance callback which holds the same lock and this would lead to a deadlock.
	pipeline.consumer.rebalanceMu.RUnlock()

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
	logger.Debugf("Waiting for response for request id %s for resource %s on parititon %d", requestId, resourceName, partition)
	request.wg.Wait()
	logger.Debugf("Got response for request id %s for resource %s on parition %d", requestId, resourceName, partition)
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

func createRebalanceCb(km *KafkaManager, mtConsumer *MultiTopicsKafkaConsumer) kafka.RebalanceCb {
	logger := km.logger.WithField("func", "createRebalanceCb")
	return func(consumer *kafka.Consumer, ev kafka.Event) error {
		mtConsumer.rebalanceMu.Lock()
		defer mtConsumer.rebalanceMu.Unlock()

		switch e := ev.(type) {
		case kafka.AssignedPartitions:
			logger.Info("Rebalance: Assigned partitions:", e.Partitions)
			err := consumer.Assign(e.Partitions)
			if err != nil {
				// Don't modify mtConsumer.partitions on assign failure
				// as the consumer state hasn't changed
				return fmt.Errorf("assign error: %w", err)
			}

			// Only update partitions after successful assignment
			mtConsumer.partitions = make([]int32, len(e.Partitions))
			for i, partition := range e.Partitions {
				mtConsumer.partitions[i] = partition.Partition
			}

			if len(e.Partitions) > 0 && mtConsumer.partitionsReady.HasListeners() {
				// signal to unblock waiting goroutines to proceed sending inference reqs
				logger.Info("Broadcasting to waiting goroutines - partitions are ready")
				mtConsumer.partitionsReady.Broadcast()
				logger.Info("Broadcast complete")
			}

		case kafka.RevokedPartitions:
			logger.Info("Rebalance: Revoked partitions:", e.Partitions)
			err := consumer.Unassign()
			if err != nil {
				return fmt.Errorf("unassign error: %w", err)
			}

			revokedPartitionSet := make(map[int32]bool)
			for _, partition := range e.Partitions {
				revokedPartitionSet[partition.Partition] = true
			}

			// We have to cancel all requests for revoked partitions. Due to repartitioning,
			// our consumer may now consume from a different partition and thus the infer
			// method will block waiting for a response that will never come.
			canceledRequests := [](*Request){}
			for _, request := range mtConsumer.requests.Items() {
				req := request.(*Request)
				req.mu.Lock()
				if revokedPartitionSet[req.partition] {
					logger.Debugf("Revoking request %s for partition %d", req.key, req.partition)
					canceledRequests = append(canceledRequests, req)
					req.response = []byte("Request revoked due to partition reassignment")
					req.isError = true
					req.wg.Done()
					req.active = false
				}
				req.mu.Unlock()
			}

			// Remove canceled requests from the map
			for _, req := range canceledRequests {
				mtConsumer.requests.Remove(req.key)
			}

			// Only clear partitions after successful unassign
			mtConsumer.partitions = nil
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
	err := pipeline.consumer.AddTopic(topicName, createRebalanceCb(km, pipeline.consumer))
	pipeline.wg.Done()
	logger.Infof("Topic %s added in consumer id %s", topicName, pipeline.consumer.id)
	if err != nil {
		return err
	}
	return nil
}
