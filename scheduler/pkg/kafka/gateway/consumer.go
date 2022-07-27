package gateway

import (
	"context"
	"sync"
	"time"

	"github.com/seldonio/seldon-core/scheduler/pkg/kafka/pipeline"

	kafka2 "github.com/seldonio/seldon-core/scheduler/pkg/kafka"
	"github.com/seldonio/seldon-core/scheduler/pkg/kafka/config"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/signalfx/splunk-otel-go/instrumentation/github.com/confluentinc/confluent-kafka-go/kafka/splunkkafka"
	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const (
	pollTimeoutMillisecs = 10000
	DefaultNumWorkers    = 8
	EnvVarNumWorkers     = "MODELGATEWAY_NUM_WORKERS"
)

type InferKafkaConsumer struct {
	logger           log.FieldLogger
	mu               sync.Mutex
	loadedModels     map[string]bool
	subscribedTopics map[string]bool
	workers          []*InferWorker
	consumer         *kafka.Consumer
	producer         *kafka.Producer
	done             chan bool
	tracer           trace.Tracer
	topicNamer       *kafka2.TopicNamer
	consumerConfig   *ConsumerConfig
	adminClient      *kafka.AdminClient
	consumerName     string
}

func NewInferKafkaConsumer(logger log.FieldLogger, consumerConfig *ConsumerConfig, consumerName string) (*InferKafkaConsumer, error) {
	ic := &InferKafkaConsumer{
		logger:           logger.WithField("source", "InferConsumer"),
		done:             make(chan bool),
		tracer:           consumerConfig.TraceProvider.GetTraceProvider().Tracer("Worker"),
		topicNamer:       kafka2.NewTopicNamer(consumerConfig.Namespace),
		loadedModels:     make(map[string]bool),
		subscribedTopics: make(map[string]bool),
		consumerConfig:   consumerConfig,
		consumerName:     consumerName,
	}
	return ic, ic.setup()
}

func (kc *InferKafkaConsumer) setup() error {
	logger := kc.logger.WithField("func", "setup")
	var err error

	producerConfigMap := config.CloneKafkaConfigMap(kc.consumerConfig.KafkaConfig.Producer)
	producerConfigMap["go.delivery.reports"] = true
	kc.logger.Infof("Creating producer with config %v", producerConfigMap)
	kc.producer, err = kafka.NewProducer(&producerConfigMap)
	if err != nil {
		return err
	}
	logger.Infof("Created producer %s", kc.producer.String())

	consumerConfig := config.CloneKafkaConfigMap(kc.consumerConfig.KafkaConfig.Consumer)
	// we map topics consistently to consumers and we choose the consumer group.id based on this mapping
	// for eg. hash(topic1) -> modelgateway-0
	// this is done by the caller i.e. ConsumerManager (manager.go)
	consumerConfig["group.id"] = kc.consumerName
	kc.logger.Infof("Creating consumer with config %v", consumerConfig)
	kc.consumer, err = kafka.NewConsumer(&consumerConfig)
	if err != nil {
		return err
	}
	logger.Infof("Created consumer %s", kc.consumer.String())

	if kc.consumerConfig.KafkaConfig.HasKafkaBootstrapServer() {
		kc.adminClient, err = kafka.NewAdminClient(&consumerConfig)
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

func (kc *InferKafkaConsumer) Stop() {
	close(kc.done)
}

func (kc *InferKafkaConsumer) subscribeTopics() error {
	topics := make([]string, len(kc.subscribedTopics))
	idx := 0
	for k := range kc.subscribedTopics {
		topics[idx] = k
		idx++
	}
	err := kc.consumer.SubscribeTopics(topics, nil)
	return err
}

func (kc *InferKafkaConsumer) GetNumModels() int {
	return len(kc.loadedModels)
}

func (kc *InferKafkaConsumer) createTopics(topicNames []string) error {
	logger := kc.logger.WithField("func", "createTopic")
	if kc.adminClient == nil {
		logger.Warnf("Can't create topics %v as no admin client", topicNames)
		return nil
	}
	t1 := time.Now()

	var topicSpecs []kafka.TopicSpecification
	for _, topicName := range topicNames {
		topicSpecs = append(topicSpecs, kafka.TopicSpecification{
			Topic:             topicName,
			NumPartitions:     1,
			ReplicationFactor: 1,
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
	logger.Infof("Topic create in %d millis", t2.Sub(t1).Milliseconds())
	return nil
}

func (kc *InferKafkaConsumer) AddModel(modelName string) error {
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

func (kc *InferKafkaConsumer) RemoveModel(modelName string) error {
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

func (kc *InferKafkaConsumer) Serve() {
	logger := kc.logger.WithField("func", "Serve")
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
			kc.logger.Infof("Stopping")
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
				span.SetAttributes(attribute.String(pipeline.RequestIdHeader, string(e.Key)))

				headers := collectHeaders(e.Headers)

				job := InferWork{
					modelName: modelName,
					msg:       e,
					headers:   headers,
				}
				// enqueue a job
				jobChan <- &job
				span.End()

			case kafka.Error:
				kc.logger.Error(e, "Received stream error")
				if e.Code() == kafka.ErrAllBrokersDown {
					run = false
				}
			default:
				kc.logger.Info("Ignored %s", e.String())
			}
		}
	}

	kc.logger.Info("Closing consumer")
	close(cancelChan)
	kc.producer.Close()
	_ = kc.consumer.Close()
}
