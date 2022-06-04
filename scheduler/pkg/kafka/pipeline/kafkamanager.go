package pipeline

import (
	"context"
	"fmt"
	"sync"

	"github.com/seldonio/seldon-core/scheduler/pkg/kafka/config"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/signalfx/splunk-otel-go/instrumentation/github.com/confluentinc/confluent-kafka-go/kafka/splunkkafka"
	"go.opentelemetry.io/otel"

	"github.com/seldonio/seldon-core/scheduler/pkg/envoy/resources"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	cmap "github.com/orcaman/concurrent-map"
	"github.com/rs/xid"
	kafka2 "github.com/seldonio/seldon-core/scheduler/pkg/kafka"
	seldontracer "github.com/seldonio/seldon-core/scheduler/pkg/tracing"
	"github.com/sirupsen/logrus"
)

const (
	pollTimeoutMillisecs = 10000
)

type PipelineInferer interface {
	Infer(ctx context.Context, resourceName string, isModel bool, data []byte, headers []kafka.Header) ([]byte, []kafka.Header, error)
}

type KafkaManager struct {
	kafkaConfig *config.KafkaConfig
	producer    *kafka.Producer
	pipelines   sync.Map
	logger      logrus.FieldLogger
	mu          sync.RWMutex
	topicNamer  *kafka2.TopicNamer
	tracer      trace.Tracer
}

type Pipeline struct {
	resourceName string
	consumer     *kafka.Consumer
	// map of kafka id to request
	requests cmap.ConcurrentMap
	done     chan bool
	isModel  bool
}

type Request struct {
	mu       sync.Mutex
	active   bool
	wg       *sync.WaitGroup
	response []byte
	headers  []kafka.Header
	isError  bool
}

func NewKafkaManager(logger logrus.FieldLogger, namespace string, kafkaConfig *config.KafkaConfig, traceProvider *seldontracer.TracerProvider) (*KafkaManager, error) {
	km := &KafkaManager{
		kafkaConfig: kafkaConfig,
		logger:      logger.WithField("source", "KafkaManager"),
		topicNamer:  kafka2.NewTopicNamer(namespace),
		tracer:      traceProvider.GetTraceProvider().Tracer("KafkaManager"),
	}
	err := km.createProducer()
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
	km.pipelines.Range(func(key interface{}, value interface{}) bool {
		pipeline := value.(*Pipeline)
		close(pipeline.done)
		return true
	})
	logger.Info("Stopped all pipelines")
}

func (km *KafkaManager) createProducer() error {
	if km.producer != nil {
		km.producer.Close()
	}
	var err error

	producerConfigMap := config.CloneKafkaConfigMap(km.kafkaConfig.Producer)
	producerConfigMap["go.delivery.reports"] = true
	km.logger.Infof("Creating producer with config %v", producerConfigMap)
	km.producer, err = kafka.NewProducer(&producerConfigMap)
	return err
}

func (km *KafkaManager) createPipeline(resource string, isModel bool) (*Pipeline, error) {
	consumerConfig := config.CloneKafkaConfigMap(km.kafkaConfig.Consumer)
	consumerConfig["group.id"] = resource
	km.logger.Infof("Creating consumer with config %v", consumerConfig)
	consumer, err := kafka.NewConsumer(&consumerConfig)
	if err != nil {
		return nil, err
	}
	km.logger.Infof("Created consumer %s", consumer.String())
	return &Pipeline{
		resourceName: resource,
		consumer:     consumer,
		requests:     cmap.New(),
		done:         make(chan bool),
		isModel:      isModel,
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
	key := getPipelineKey(resourceName, isModel)
	if val, ok := km.pipelines.Load(key); ok {
		return val.(*Pipeline), nil
	} else {
		pipeline, err := km.createPipeline(resourceName, isModel)
		if err != nil {
			return nil, err
		}
		val, loaded := km.pipelines.LoadOrStore(key, pipeline)
		if loaded {
			return val.(*Pipeline), nil
		} else {
			go func() {
				err := km.consume(pipeline)
				if err != nil {
					km.logger.WithError(err).Errorf("Failed running consumer for resource %s", resourceName)
				}
			}()
			return pipeline, nil
		}
	}
}

func (km *KafkaManager) Infer(ctx context.Context, resourceName string, isModel bool, data []byte, headers []kafka.Header) ([]byte, []kafka.Header, error) {
	logger := km.logger.WithField("func", "Infer")
	km.mu.RLock()
	defer km.mu.RUnlock()
	pipeline, err := km.loadOrStorePipeline(resourceName, isModel)
	if err != nil {
		return nil, nil, err
	}
	key := xid.New().String()
	request := &Request{
		active: true,
		wg:     new(sync.WaitGroup),
	}
	pipeline.requests.Set(key, request)
	defer pipeline.requests.Remove(key)
	request.wg.Add(1)

	outputTopic := km.topicNamer.GetPipelineTopicInputs(resourceName)
	if isModel {
		outputTopic = km.topicNamer.GetModelTopicInputs(resourceName)
	}
	logger.Infof("Produce on topic %s with key %s", outputTopic, key)
	kafkaHeaders := append(headers, kafka.Header{Key: resources.SeldonPipelineHeader, Value: []byte(resourceName)})

	msg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &outputTopic, Partition: kafka.PartitionAny},
		Key:            []byte(key),
		Value:          data,
		Headers:        kafkaHeaders,
	}

	ctx, span := km.tracer.Start(ctx, "Produce")
	span.SetAttributes(attribute.String(seldontracer.SELDON_REQUEST_ID, key))
	// Add trace headers
	carrier := splunkkafka.NewMessageCarrier(msg)
	otel.GetTextMapPropagator().Inject(ctx, carrier)

	deliveryChan := make(chan kafka.Event)
	err = km.producer.Produce(msg, deliveryChan)
	if err != nil {
		span.End()
		return nil, nil, err
	}
	go func() {
		<-deliveryChan
		span.End()
	}()
	request.wg.Wait()
	if request.isError {
		return nil, nil, fmt.Errorf("%s", string(request.response))
	}
	return request.response, request.headers, nil
}

func hasErrorHeader(headers []kafka.Header) bool {
	for _, header := range headers {
		if header.Key == kafka2.TopicErrorHeader {
			return true
		}
	}
	return false
}

func (km *KafkaManager) consume(pipeline *Pipeline) error {
	logger := km.logger.WithField("func", "consume")
	topicName := km.topicNamer.GetPipelineTopicOutputs(pipeline.resourceName)
	if pipeline.isModel {
		topicName = km.topicNamer.GetModelTopicOutputs(pipeline.resourceName)
	}

	err := pipeline.consumer.SubscribeTopics([]string{topicName}, nil)
	if err != nil {
		return err
	}
	logger.Infof("Started consumer for topic (pipeline:%v) %s", !pipeline.isModel, topicName)
	run := true
	for run {
		select {
		case <-pipeline.done:
			run = false
		default:
			ev := pipeline.consumer.Poll(pollTimeoutMillisecs)
			if ev == nil {
				continue
			}

			switch e := ev.(type) {
			case *kafka.Message:
				logger.Infof("Received message from %s with key %s", topicName, string(e.Key))
				if val, ok := pipeline.requests.Get(string(e.Key)); ok {

					// Add tracing span
					ctx := context.Background()
					carrierIn := splunkkafka.NewMessageCarrier(e)
					ctx = otel.GetTextMapPropagator().Extract(ctx, carrierIn)
					_, span := km.tracer.Start(ctx, "Consume")
					span.SetAttributes(attribute.String(seldontracer.SELDON_REQUEST_ID, string(e.Key)))

					request := val.(*Request)
					request.mu.Lock()
					if request.active {
						request.isError = hasErrorHeader(e.Headers)
						request.response = e.Value
						request.headers = e.Headers
						request.wg.Done()
						request.active = false
					} else {
						logger.Warnf("Got duplicate request with key %s", string(e.Key))
					}
					request.mu.Unlock()
					span.End()
				}

			case kafka.Error:
				km.logger.Error(e, "Received stream error")
				if e.Code() == kafka.ErrAllBrokersDown {
					run = false
				}
			default:
				km.logger.Debug("Ignored %s", e.String())
			}
		}
	}
	return pipeline.consumer.Close()
}
