package gateway

import (
	"context"

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
)

type InferKafkaGateway struct {
	logger         log.FieldLogger
	nworkers       int
	workers        []*InferWorker
	kafkaConfig    *config.KafkaConfig
	modelConfig    *KafkaModelConfig
	serverConfig   *KafkaServerConfig
	consumer       *kafka.Consumer
	producer       *kafka.Producer
	done           chan bool
	tracerProvider *seldontracer.TracerProvider
	tracer         trace.Tracer
}

func NewInferKafkaGateway(logger log.FieldLogger, nworkers int, kafkaConfig *config.KafkaConfig, modelConfig *KafkaModelConfig, serverConfig *KafkaServerConfig, traceProvider *seldontracer.TracerProvider) (*InferKafkaGateway, error) {
	ic := &InferKafkaGateway{
		logger:         logger.WithField("source", "InferConsumer"),
		nworkers:       nworkers,
		kafkaConfig:    kafkaConfig,
		modelConfig:    modelConfig,
		serverConfig:   serverConfig,
		done:           make(chan bool),
		tracerProvider: traceProvider,
		tracer:         traceProvider.GetTraceProvider().Tracer("Worker"),
	}
	return ic, ic.setup()
}

func (ig *InferKafkaGateway) setup() error {
	logger := ig.logger.WithField("func", "setup")
	var err error

	producerConfigMap := config.CloneKafkaConfigMap(ig.kafkaConfig.Producer)
	producerConfigMap["go.delivery.reports"] = true
	ig.logger.Infof("Creating producer with config %v", producerConfigMap)
	ig.producer, err = kafka.NewProducer(&producerConfigMap)
	if err != nil {
		return err
	}
	logger.Infof("Created producer %s", ig.producer.String())

	consumerConfig := config.CloneKafkaConfigMap(ig.kafkaConfig.Consumer)
	consumerConfig["group.id"] = ig.modelConfig.ModelName
	ig.logger.Infof("Creating consumer with config %v", consumerConfig)
	ig.consumer, err = kafka.NewConsumer(&consumerConfig)
	if err != nil {
		return err
	}
	logger.Infof("Created consumer %s", ig.consumer.String())

	err = ig.consumer.SubscribeTopics([]string{ig.modelConfig.InputTopic}, nil)
	if err != nil {
		return err
	}

	for i := 0; i < ig.nworkers; i++ {
		worker, err := NewInferWorker(ig, ig.logger, ig.tracerProvider)
		if err != nil {
			return err
		}
		ig.workers = append(ig.workers, worker)
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

func (ig *InferKafkaGateway) Stop() {
	close(ig.done)
}

func (ig *InferKafkaGateway) Serve() {
	run := true
	// create a cancel and job channel
	cancelChan := make(chan struct{})
	jobChan := make(chan *InferWork, ig.nworkers)
	// Start workers
	for i := 0; i < ig.nworkers; i++ {
		go ig.workers[i].Start(jobChan, cancelChan)
	}

	for run {
		select {
		case <-ig.done:
			ig.logger.Infof("Stopping")
			run = false
		default:
			ev := ig.consumer.Poll(pollTimeoutMillisecs)
			if ev == nil {
				continue
			}

			switch e := ev.(type) {
			case *kafka.Message:

				// Add tracing span
				ctx := context.Background()
				carrierIn := splunkkafka.NewMessageCarrier(e)
				ctx = otel.GetTextMapPropagator().Extract(ctx, carrierIn)
				_, span := ig.tracer.Start(ctx, "Consume")
				span.SetAttributes(attribute.String(seldontracer.SELDON_REQUEST_ID, string(e.Key)))

				headers := collectHeaders(e.Headers)

				job := InferWork{
					msg:     e,
					headers: headers,
				}
				// enqueue a job
				jobChan <- &job
				span.End()

			case kafka.Error:
				ig.logger.Error(e, "Received stream error")
				if e.Code() == kafka.ErrAllBrokersDown {
					run = false
				}
			default:
				ig.logger.Info("Ignored %s", e.String())
			}
		}
	}

	ig.logger.Info("Closing consumer")
	close(cancelChan)
	ig.producer.Close()
	_ = ig.consumer.Close()
}
