package gateway

import (
	"context"
	"sync"

	kafka2 "github.com/seldonio/seldon-core/scheduler/pkg/kafka"

	"github.com/seldonio/seldon-core/scheduler/pkg/kafka/config"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	seldontracer "github.com/seldonio/seldon-core/scheduler/pkg/tracing"
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
	logger                log.FieldLogger
	mu                    sync.Mutex
	loadedModels          map[string]bool
	nworkers              int
	workers               []*InferWorker
	kafkaConfig           *config.KafkaConfig
	inferenceServerConfig *InferenceServerConfig
	consumer              *kafka.Consumer
	producer              *kafka.Producer
	done                  chan bool
	tracerProvider        *seldontracer.TracerProvider
	tracer                trace.Tracer
	topicNamer            *kafka2.TopicNamer
}

func NewInferKafkaConsumer(logger log.FieldLogger, nworkers int, kafkaConfig *config.KafkaConfig, namespace string, inferenceServerConfig *InferenceServerConfig, traceProvider *seldontracer.TracerProvider) (*InferKafkaConsumer, error) {
	ic := &InferKafkaConsumer{
		logger:                logger.WithField("source", "InferConsumer"),
		nworkers:              nworkers,
		kafkaConfig:           kafkaConfig,
		inferenceServerConfig: inferenceServerConfig,
		done:                  make(chan bool),
		tracerProvider:        traceProvider,
		tracer:                traceProvider.GetTraceProvider().Tracer("Worker"),
		topicNamer:            kafka2.NewTopicNamer(namespace),
		loadedModels:          make(map[string]bool),
	}
	return ic, ic.setup()
}

func (kc *InferKafkaConsumer) setup() error {
	logger := kc.logger.WithField("func", "setup")
	var err error

	producerConfigMap := config.CloneKafkaConfigMap(kc.kafkaConfig.Producer)
	producerConfigMap["go.delivery.reports"] = true
	kc.logger.Infof("Creating producer with config %v", producerConfigMap)
	kc.producer, err = kafka.NewProducer(&producerConfigMap)
	if err != nil {
		return err
	}
	logger.Infof("Created producer %s", kc.producer.String())

	consumerConfig := config.CloneKafkaConfigMap(kc.kafkaConfig.Consumer)
	consumerConfig["group.id"] = "seldon-modelgateway"
	kc.logger.Infof("Creating consumer with config %v", consumerConfig)
	kc.consumer, err = kafka.NewConsumer(&consumerConfig)
	if err != nil {
		return err
	}
	logger.Infof("Created consumer %s", kc.consumer.String())

	err = kc.consumer.SubscribeTopics([]string{kc.topicNamer.GetKafkaModelTopicRegex()}, nil)
	if err != nil {
		return err
	}

	for i := 0; i < kc.nworkers; i++ {
		worker, err := NewInferWorker(kc, kc.logger, kc.tracerProvider, kc.topicNamer)
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

func (kc *InferKafkaConsumer) AddModel(modelName string) {
	kc.mu.Lock()
	defer kc.mu.Unlock()
	kc.loadedModels[modelName] = true
}

func (kc *InferKafkaConsumer) RemoveModel(modelName string) {
	kc.mu.Lock()
	defer kc.mu.Unlock()
	delete(kc.loadedModels, modelName)
}

func (kc *InferKafkaConsumer) Serve() {
	logger := kc.logger.WithField("func", "Serve")
	run := true
	// create a cancel and job channel
	cancelChan := make(chan struct{})
	jobChan := make(chan *InferWork, kc.nworkers)
	// Start workers
	for i := 0; i < kc.nworkers; i++ {
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
				span.SetAttributes(attribute.String(seldontracer.SELDON_REQUEST_ID, string(e.Key)))

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
