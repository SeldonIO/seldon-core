/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package gateway

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry"
	"github.com/signalfx/splunk-otel-go/instrumentation/github.com/confluentinc/confluent-kafka-go/v2/kafka/splunkkafka"
	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	kafkaconfig "github.com/seldonio/seldon-core/components/kafka/v2/pkg/config"

	kafka2 "github.com/seldonio/seldon-core/scheduler/v2/pkg/kafka"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/kafka/pipeline"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/kafka/schema"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/util"
)

const (
	pollTimeoutMilliseconds     = 1000
	DefaultNumWorkers           = 8
	EnvVarNumWorkers            = "MODELGATEWAY_NUM_WORKERS"
	DefaultWorkerTimeoutMs      = 2 * 60 * 1000
	EnvVarWorkerTimeoutMs       = "MODELGATEWAY_WORKER_TIMEOUT_MS"
	replicationFactorKey        = "replicationFactor"
	numPartitionsKey            = "numPartitions"
	envDefaultReplicationFactor = "KAFKA_DEFAULT_REPLICATION_FACTOR"
	envDefaultNumPartitions     = "KAFKA_DEFAULT_NUM_PARTITIONS"
	defaultReplicationFactor    = 1
	defaultNumPartitions        = 1
)

type InferKafkaHandler struct {
	logger               log.FieldLogger
	mu                   sync.RWMutex
	loadedModels         map[string]bool
	subscribedTopics     map[string]bool
	workers              []*InferWorker
	consumer             *kafka.Consumer
	producer             *kafka.Producer
	done                 chan bool
	shutdownComplete     chan struct{}
	tracer               trace.Tracer
	topicNamer           *kafka2.TopicNamer
	consumerConfig       *ManagerConfig
	adminClient          *kafka.AdminClient
	consumerName         string
	replicationFactor    int
	numPartitions        int
	tlsClientOptions     *util.TLSOptions
	producerMu           sync.RWMutex
	producerActive       atomic.Bool
	schemaRegistryClient schemaregistry.Client
}

func GetIntConfigMapValue(configMap kafka.ConfigMap, key string, defaultValue int) (int, error) {
	configMapValue, ok := configMap[key]
	if !ok {
		return defaultValue, nil
	}

	if configMapValueInt, ok := configMapValue.(int); ok {
		if configMapValueInt < 0 {
			return -1, fmt.Errorf("%s: %d must not be negative", key, configMapValueInt)
		}
		return configMapValueInt, nil
	}

	configMapValueStr, ok := configMapValue.(string)
	if !ok {
		return defaultValue, fmt.Errorf("%s key has wrong type: %T", key, configMapValue)
	}

	value, err := strconv.Atoi(configMapValueStr)
	if err != nil {
		return -1, fmt.Errorf("invalid value %s in %s with error: %v", configMapValueStr, key, err)
	}

	if value < 0 {
		return -1, fmt.Errorf("%s: %d must be bigger than 0", key, value)
	}

	return value, nil
}

func NewInferKafkaHandler(
	logger log.FieldLogger,
	consumerConfig *ManagerConfig,
	consumerConfigMap kafka.ConfigMap,
	producerConfigMap kafka.ConfigMap,
	topicsConfigMap kafka.ConfigMap,
	consumerName string,
	schemaRegistryClient schemaregistry.Client,
) (*InferKafkaHandler, error) {
	replicationFactor, err := util.GetIntEnvar(envDefaultReplicationFactor, defaultReplicationFactor)
	if err != nil {
		return nil, fmt.Errorf("error getting default replication factor: %v", err)
	}
	numPartitions, err := util.GetIntEnvar(envDefaultNumPartitions, defaultNumPartitions)
	if err != nil {
		return nil, fmt.Errorf("invalid Kafka topic configuration: %w", err)
	}
	tlsClientOptions, err := util.CreateTLSClientOptions()
	if err != nil {
		return nil, err
	}
	topicNamer, err := kafka2.NewTopicNamer(consumerConfig.Namespace, consumerConfig.SeldonKafkaConfig.TopicPrefix)
	if err != nil {
		return nil, err
	}

	ic := &InferKafkaHandler{
		logger:               logger.WithField("source", "InferConsumer"),
		done:                 make(chan bool),
		tracer:               consumerConfig.TraceProvider.GetTraceProvider().Tracer("Worker"),
		topicNamer:           topicNamer,
		loadedModels:         make(map[string]bool),
		subscribedTopics:     make(map[string]bool),
		shutdownComplete:     make(chan struct{}),
		consumerConfig:       consumerConfig,
		consumerName:         consumerName,
		replicationFactor:    replicationFactor,
		numPartitions:        numPartitions,
		tlsClientOptions:     tlsClientOptions,
		schemaRegistryClient: schemaRegistryClient,
	}

	return ic, ic.setup(consumerConfigMap, producerConfigMap)
}

func (kc *InferKafkaHandler) setup(consumerConfig kafka.ConfigMap, producerConfig kafka.ConfigMap) error {
	logger := kc.logger.WithField("func", "setup")
	var err error

	producerConfigWithoutSecrets := kafkaconfig.WithoutSecrets(producerConfig)
	kc.logger.Infof("Creating producer with config %v", producerConfigWithoutSecrets)
	kc.producer, err = kafka.NewProducer(&producerConfig)
	if err != nil {
		return err
	}
	kc.producerActive.Store(true)
	logger.Infof("Created producer %s", kc.producer.String())

	// we map topics consistently to consumers and we choose the consumer group.id based on this mapping
	// for eg. hash(topic1) -> modelgateway-0
	// this is done by the caller i.e. ConsumerManager (store.go)
	consumerConfig["group.id"] = kc.consumerName
	consumerConfigWithoutSecrets := kafkaconfig.WithoutSecrets(consumerConfig)
	kc.logger.Infof("Creating consumer with config %v", consumerConfigWithoutSecrets)
	kc.consumer, err = kafka.NewConsumer(&consumerConfig)
	if err != nil {
		return err
	}
	logger.Infof("Created consumer %s", kc.consumer.String())

	if kc.consumerConfig.SeldonKafkaConfig.HasKafkaBootstrapServer() {
		kc.adminClient, err = kafka.NewAdminClientFromProducer(kc.producer)
		if err != nil {
			return err
		}
	}

	for i := 0; i < kc.consumerConfig.NumWorkers; i++ {
		worker, err := NewInferWorker(kc, kc.logger, kc.consumerConfig.TraceProvider, kc.topicNamer, kc.schemaRegistryClient)
		if err != nil {
			return err
		}
		kc.workers = append(kc.workers, worker)
	}
	return nil
}

// Will overwrite duplicate stream headers
func collectHeaders(kheaders []kafka.Header) map[string]string {
	headers := make(map[string]string)
	for _, kheader := range kheaders {
		headers[kheader.Key] = string(kheader.Value)
	}
	return headers
}

func (kc *InferKafkaHandler) Produce(msg *kafka.Message, deliveryChan chan kafka.Event) error {
	logger := kc.logger.WithField("func", "Produce")
	kc.producerMu.RLock()
	defer kc.producerMu.RUnlock()
	if kc.producerActive.Load() {
		return kc.producer.Produce(msg, deliveryChan)
	} else {
		err := fmt.Errorf("The infer producer %s is no longer running", kc.producer.String())
		logger.WithError(err).Error("Failed to produce kafka message")
		return err
	}
}

func (kc *InferKafkaHandler) producerIsActive() bool {
	return kc.producerActive.Load()
}

func (kc *InferKafkaHandler) closeProducer() {
	kc.producerMu.Lock()
	defer kc.producerMu.Unlock()
	kc.producer.Close()
}

func (kc *InferKafkaHandler) Stop(waitForShutdown bool) {
	close(kc.done)
	if waitForShutdown {
		<-kc.shutdownComplete
	}
}

func (kc *InferKafkaHandler) subscribeTopics() error {
	topics := make([]string, len(kc.subscribedTopics))
	idx := 0
	for k := range kc.subscribedTopics {
		topics[idx] = k
		idx++
	}
	err := kc.consumer.SubscribeTopics(topics, nil)
	return err
}

func (kc *InferKafkaHandler) GetNumModels() int {
	kc.mu.RLock()
	defer kc.mu.RUnlock()
	return len(kc.loadedModels)
}

func (kc *InferKafkaHandler) createTopics(topicNames []string) error {
	logger := kc.logger.WithField("func", "createTopics")
	if kc.adminClient == nil {
		logger.Warnf("no kafka admin client, can't create any of the following topics: %v", topicNames)
		// An error would typically be returned here, but a missing adminClient typically
		// indicates we're running tests. Instead of failing tests, we return nil here.
		// TODO: we should create interfaces for adminClient and a wrapper around kafka so we can provide
		//  interfaces which can be easily mocked with mockgen
		return nil
	}
	t1 := time.Now()

	var topicSpecs []kafka.TopicSpecification
	for _, topicName := range topicNames {
		topicSpecs = append(topicSpecs, kafka.TopicSpecification{
			Topic:             topicName,
			NumPartitions:     kc.numPartitions,
			ReplicationFactor: kc.replicationFactor,
		})
	}
	results, err := kc.adminClient.CreateTopics(
		context.Background(),
		topicSpecs,
		kafka.SetAdminOperationTimeout(TopicCreateTimeout),
	)
	if err != nil {
		return err
	}

	// Wait for topic creation
	logFailure := func(err error, delay time.Duration) {
		logger.WithError(err).Errorf("still waiting for all topics to be created...")
	}

	logger.Infof("waiting for kafka topic creation")
	retryPolicy := backoff.WithMaxRetries(
		backoff.NewConstantBackOff(TopicDescribeRetryDelay),
		TopicDescribeMaxRetries,
	)
	err = backoff.RetryNotify(
		func() error {
			return kc.ensureTopicsExist(topicNames)
		},
		retryPolicy,
		logFailure)

	if err != nil {
		logger.WithError(err).Errorf("some topics not created, giving up")
		return err
	} else {
		logger.Infof("all topics created")
	}

	for _, result := range results {
		logger.Debugf("topic result for %s", result.String())
	}

	t2 := time.Now()
	logger.Debugf("kafka topics created in %d millis", t2.Sub(t1).Milliseconds())

	return nil
}

func (kc *InferKafkaHandler) deleteTopics(topicNames []string) error {
	logger := kc.logger.WithField("func", "deleteTopics")
	if kc.adminClient == nil {
		logger.Warnf("no kafka admin client, can't delete any of the following topics: %v", topicNames)
		return nil
	}
	t1 := time.Now()

	results, err := kc.adminClient.DeleteTopics(
		context.Background(),
		topicNames,
		kafka.SetAdminOperationTimeout(TopicDeleteTimeout),
	)
	if err != nil {
		return err
	}

	var failedTopics []string
	for _, result := range results {
		if result.Error.Code() != kafka.ErrNoError {
			failedTopics = append(failedTopics, result.Topic)
			logger.Errorf("failed to delete topic %s: %s", result.Topic, result.Error.Error())
		} else {
			logger.Infof("topic %s deleted", result.Topic)
		}
	}

	if len(failedTopics) > 0 {
		return fmt.Errorf("failed to delete topics: %v", failedTopics)
	}

	t2 := time.Now()
	logger.Debugf("kafka topics deleted in %d millis", t2.Sub(t1).Milliseconds())
	return nil
}

func (kc *InferKafkaHandler) ensureTopicsExist(topicNames []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), TopicDescribeTimeout)
	defer cancel()
	topicsDescResult, err := kc.adminClient.DescribeTopics(
		ctx,
		kafka.NewTopicCollectionOfTopicNames(topicNames),
		kafka.SetAdminOptionIncludeAuthorizedOperations(false))
	if err != nil {
		return err
	}

	for _, topicDescription := range topicsDescResult.TopicDescriptions {
		if topicDescription.Error.Code() != kafka.ErrNoError {
			return fmt.Errorf("topic description failure: %s", topicDescription.Error.Error())
		}
	}

	return nil
}

func (kc *InferKafkaHandler) AddModel(modelName string) error {
	kc.mu.Lock()
	defer kc.mu.Unlock()
	kc.loadedModels[modelName] = true

	// create topics
	inputTopic := kc.topicNamer.GetModelTopicInputs(modelName)
	outputTopic := kc.topicNamer.GetModelTopicOutputs(modelName)
	if err := kc.createTopics([]string{inputTopic, outputTopic}); err != nil {
		// we must delete the model as we've failed to create the topics, scheduler will re-issue a create cmd, if model
		// exists in loadedModels model-gw will respond with success and it'll be no-op and topics won't be retried to
		// be created
		delete(kc.loadedModels, modelName)
		return err
	}

	kc.subscribedTopics[inputTopic] = true
	err := kc.subscribeTopics()
	if err != nil {
		kc.logger.WithError(err).Errorf("failed to subscribe to topics")
		return nil
	}
	return nil
}

func (kc *InferKafkaHandler) RemoveModel(modelName string, cleanTopicsOnDeletion bool, keepTopics bool) error {
	kc.mu.Lock()
	defer kc.mu.Unlock()

	if _, ok := kc.loadedModels[modelName]; !ok {
		kc.logger.WithField("model", modelName).Info("Model does not exist, no topics to remove")
		return nil
	}

	delete(kc.loadedModels, modelName)
	delete(kc.subscribedTopics, kc.topicNamer.GetModelTopicInputs(modelName))
	if len(kc.subscribedTopics) > 0 {
		kc.logger.WithField("topics", kc.subscribedTopics).Debug("Re-subscribing to remaining topics after model deletion")
		err := kc.subscribeTopics()
		if err != nil {
			kc.logger.WithError(err).Errorf("failed to subscribe to topics")
			return nil
		}
	}

	inputTopic := kc.topicNamer.GetModelTopicInputs(modelName)
	outputTopic := kc.topicNamer.GetModelTopicOutputs(modelName)

	if !keepTopics && cleanTopicsOnDeletion {
		// delete input and output topics from kafka
		err := kc.deleteTopics([]string{inputTopic, outputTopic})
		if err != nil {
			return err
		}
	}

	if kc.schemaRegistryClient == nil {
		return nil
	}

	_, err := kc.schemaRegistryClient.DeleteSubject(inputTopic, cleanTopicsOnDeletion)
	if err != nil {
		kc.logger.WithError(err).Errorf("failed to delete schema for input topic %s and model name %s", inputTopic, modelName)
	}

	_, err = kc.schemaRegistryClient.DeleteSubject(outputTopic, cleanTopicsOnDeletion)
	if err != nil {
		kc.logger.WithError(err).Errorf("failed to delete schema for input topic %s and model name %s", outputTopic, modelName)
	}

	return nil
}

func (kc *InferKafkaHandler) Exists(modelName string) bool {
	kc.mu.RLock()
	defer kc.mu.RUnlock()
	_, ok := kc.loadedModels[modelName]
	return ok
}

func (kc *InferKafkaHandler) Serve() {
	logger := kc.logger.WithField("func", "Serve").WithField("consumerName", kc.consumerName)
	run := true
	// create a cancel and job channel
	cancelChan := make(chan struct{})
	jobChan := make(chan *InferWork, kc.consumerConfig.NumWorkers)
	// Start workers
	for i := 0; i < kc.consumerConfig.NumWorkers; i++ {
		go kc.workers[i].Start(jobChan, cancelChan, kc.consumerConfig.WorkerTimeout)
	}

	for run {
		select {
		case <-kc.done:
			logger.Infof("stopping consumer %s", kc.consumer.String())
			kc.producerActive.Store(false)
			run = false
		default:
			ev := kc.consumer.Poll(pollTimeoutMilliseconds)
			if ev == nil {
				continue
			}

			switch e := ev.(type) {
			case *kafka.Message:

				var modelName string
				var err error
				if e.TopicPartition.Topic != nil {
					modelName, err = kc.topicNamer.GetModelNameFromModelInputTopic(*e.TopicPartition.Topic)
					if err != nil {
						logger.WithError(err).Errorf("Failed to extract modelName from topic %s", *e.TopicPartition.Topic)
						continue
					}
				} else {
					logger.Errorf("Received message with no topic name")
					continue
				}

				if !kc.Exists(modelName) {
					logger.Infof("Failed to find model %s in loaded models", modelName)
					continue
				}

				if kc.schemaRegistryClient != nil {
					e.Value = schema.TrimSchemaID(e.Value)
				}

				// Add tracing span
				ctx := context.Background()
				carrierIn := splunkkafka.NewMessageCarrier(e)
				ctx = otel.GetTextMapPropagator().Extract(ctx, carrierIn)

				_, span := kc.tracer.Start(ctx, "Consume")
				requestId := pipeline.GetRequestIdFromKafkaHeaders(e.Headers)
				if requestId == "" {
					logger.Warnf("Missing request id in Kafka headers for key %s", string(e.Key))
				}
				span.SetAttributes(attribute.String(util.RequestIdHeader, requestId))

				headers := collectHeaders(e.Headers)
				logger.Debugf("Headers received from kafka for model %s %v", modelName, e.Headers)

				_, spanJob := kc.tracer.Start(ctx, "WaitForWorker")
				spanJob.SetAttributes(attribute.String(util.RequestIdHeader, requestId))

				job := InferWork{
					modelName: modelName,
					msg:       e,
					headers:   headers,
					span:      spanJob,
				}
				// enqueue a job
				jobChan <- &job
				span.End()

			case kafka.Error:
				logger.Errorf("Kafka error, code: [%s] msg: [%s]", e.Code().String(), e.Error())
			default:
				logger.Debugf("Ignored %s", e.String())
			}
		}
	}

	logger.Info("Closing consumer")
	close(cancelChan)
	kc.closeProducer()
	if _, err := kc.consumer.Commit(); err != nil {
		logger.WithError(err).Error("Failed to commit offsets")
	}
	if err := kc.consumer.Close(); err != nil {
		logger.WithError(err).Error("Failure closing consumer")
	}
	close(kc.shutdownComplete)
}
