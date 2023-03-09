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
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/util"

	kafka2 "github.com/seldonio/seldon-core/scheduler/v2/pkg/kafka"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/kafka/config"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/signalfx/splunk-otel-go/instrumentation/github.com/confluentinc/confluent-kafka-go/kafka/splunkkafka"
	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const (
	pollTimeoutMillisecs        = 10000
	DefaultNumWorkers           = 8
	EnvVarNumWorkers            = "MODELGATEWAY_NUM_WORKERS"
	envDefaultReplicationFactor = "KAFKA_DEFAULT_REPLICATION_FACTOR"
	envDefaultNumPartitions     = "KAFKA_DEFAULT_NUM_PARTITIONS"
	defaultReplicationFactor    = 1
	defaultNumPartitions        = 1
)

type InferKafkaHandler struct {
	logger            log.FieldLogger
	mu                sync.RWMutex
	loadedModels      map[string]bool
	subscribedTopics  map[string]bool
	workers           []*InferWorker
	consumer          *kafka.Consumer
	producer          *kafka.Producer
	done              chan bool
	tracer            trace.Tracer
	topicNamer        *kafka2.TopicNamer
	consumerConfig    *ConsumerConfig
	adminClient       *kafka.AdminClient
	consumerName      string
	replicationFactor int
	numPartitions     int
	tlsClientOptions  *util.TLSOptions
	producerMu        sync.RWMutex
	producerActive    atomic.Bool
}

func NewInferKafkaHandler(logger log.FieldLogger, consumerConfig *ConsumerConfig, consumerName string) (*InferKafkaHandler, error) {
	replicationFactor, err := util.GetIntEnvar(envDefaultReplicationFactor, defaultReplicationFactor)
	if err != nil {
		return nil, err
	}
	numPartitions, err := util.GetIntEnvar(envDefaultNumPartitions, defaultNumPartitions)
	if err != nil {
		return nil, err
	}
	tlsClientOptions, err := util.CreateTLSClientOptions()
	if err != nil {
		return nil, err
	}
	topicNamer, err := kafka2.NewTopicNamer(consumerConfig.Namespace, consumerConfig.KafkaConfig.TopicPrefix)
	if err != nil {
		return nil, err
	}
	ic := &InferKafkaHandler{
		logger:            logger.WithField("source", "InferConsumer"),
		done:              make(chan bool),
		tracer:            consumerConfig.TraceProvider.GetTraceProvider().Tracer("Worker"),
		topicNamer:        topicNamer,
		loadedModels:      make(map[string]bool),
		subscribedTopics:  make(map[string]bool),
		consumerConfig:    consumerConfig,
		consumerName:      consumerName,
		replicationFactor: replicationFactor,
		numPartitions:     numPartitions,
		tlsClientOptions:  tlsClientOptions,
	}
	return ic, ic.setup()
}

func (kc *InferKafkaHandler) setup() error {
	logger := kc.logger.WithField("func", "setup")
	var err error

	producerConfig := config.CloneKafkaConfigMap(kc.consumerConfig.KafkaConfig.Producer)
	producerConfig["go.delivery.reports"] = true
	err = config.AddKafkaSSLOptions(producerConfig)
	if err != nil {
		return err
	}

	producerConfigAsJSON, err := json.Marshal(&producerConfig)
	if err != nil {
		logger.WithField("config", &producerConfig).Info("Creating producer")
	} else {
		logger.WithField("config", string(producerConfigAsJSON)).Info("Creating producer")
	}
	kc.logger.Infof("Creating producer with config %v", producerConfig)
	kc.producer, err = kafka.NewProducer(&producerConfig)
	if err != nil {
		return err
	}
	kc.producerActive.Store(true)
	logger.Infof("Created producer %s", kc.producer.String())

	consumerConfig := config.CloneKafkaConfigMap(kc.consumerConfig.KafkaConfig.Consumer)
	// we map topics consistently to consumers and we choose the consumer group.id based on this mapping
	// for eg. hash(topic1) -> modelgateway-0
	// this is done by the caller i.e. ConsumerManager (store.go)
	consumerConfig["group.id"] = kc.consumerName
	err = config.AddKafkaSSLOptions(consumerConfig)
	if err != nil {
		return err
	}

	consumerConfigAsJson, err := json.Marshal(&consumerConfig)
	if err != nil {
		logger.WithField("config", &consumerConfig).Info("Creating consumer")
	} else {
		logger.WithField("config", string(consumerConfigAsJson)).Info("Creating consumer")
	}

	kc.consumer, err = kafka.NewConsumer(&consumerConfig)
	if err != nil {
		return err
	}
	logger.Infof("Created consumer %s", kc.consumer.String())

	if kc.consumerConfig.KafkaConfig.HasKafkaBootstrapServer() {
		adminConfig := kafka.ConfigMap{
			config.KafkaBootstrapServers: kc.consumerConfig.KafkaConfig.BootstrapServers,
		}
		err = config.AddKafkaSSLOptions(adminConfig)
		if err != nil {
			return err
		}
		kc.adminClient, err = kafka.NewAdminClient(&adminConfig)
		if err != nil {
			return err
		}
	}

	for i := 0; i < kc.consumerConfig.NumWorkers; i++ {
		worker, err := NewInferWorker(kc, kc.logger, kc.consumerConfig.TraceProvider, kc.topicNamer)
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

func (kc *InferKafkaHandler) closeProducer() {
	kc.producerMu.Lock()
	defer kc.producerMu.Unlock()
	kc.producer.Close()
}

func (kc *InferKafkaHandler) Stop() {
	close(kc.done)
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
		logger.Warnf("Can't create topics %v as no admin client", topicNames)
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
	results, err := kc.adminClient.CreateTopics(context.Background(), topicSpecs, kafka.SetAdminOperationTimeout(time.Minute))
	if err != nil {
		return err
	}
	for _, result := range results {
		logger.Debugf("Topic result for %s", result.String())
	}
	t2 := time.Now()
	logger.Infof("Topic created in %d millis", t2.Sub(t1).Milliseconds())
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
		return err
	}

	kc.subscribedTopics[inputTopic] = true
	err := kc.subscribeTopics()
	if err != nil {
		kc.logger.WithError(err).Warn("Failed to subscribe to topics")
		return nil
	}
	return nil
}

func (kc *InferKafkaHandler) RemoveModel(modelName string) error {
	kc.mu.Lock()
	defer kc.mu.Unlock()
	delete(kc.loadedModels, modelName)
	delete(kc.subscribedTopics, kc.topicNamer.GetModelTopicInputs(modelName))
	if len(kc.subscribedTopics) > 0 {
		err := kc.subscribeTopics()
		if err != nil {
			kc.logger.WithError(err).Errorf("Failed to subscribe to topics")
			return nil
		}
	}
	return nil
}

func (kc *InferKafkaHandler) Serve() {
	logger := kc.logger.WithField("func", "Serve").WithField("consumerName", kc.consumerName)
	run := true
	// create a cancel and job channel
	cancelChan := make(chan struct{})
	jobChan := make(chan *InferWork, kc.consumerConfig.NumWorkers)
	// Start workers
	for i := 0; i < kc.consumerConfig.NumWorkers; i++ {
		go kc.workers[i].Start(jobChan, cancelChan)
	}

	for run {
		select {
		case <-kc.done:
			logger.Infof("Stopping consumer %s", kc.consumer.String())
			kc.producerActive.Store(false)
			run = false
		default:
			ev := kc.consumer.Poll(pollTimeoutMillisecs)
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

				kc.mu.Lock()
				if _, ok := kc.loadedModels[modelName]; !ok {
					kc.mu.Unlock()
					logger.Infof("Failed to find model %s in loaded models", modelName)
					continue
				}
				kc.mu.Unlock()

				// Add tracing span
				ctx := context.Background()
				carrierIn := splunkkafka.NewMessageCarrier(e)
				ctx = otel.GetTextMapPropagator().Extract(ctx, carrierIn)
				_, span := kc.tracer.Start(ctx, "Consume")
				span.SetAttributes(attribute.String(util.RequestIdHeader, string(e.Key)))

				headers := collectHeaders(e.Headers)
				logger.Debugf("Headers received from kafka for model %s %v", modelName, e.Headers)

				job := InferWork{
					modelName: modelName,
					msg:       e,
					headers:   headers,
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
	_ = kc.consumer.Close()
}
