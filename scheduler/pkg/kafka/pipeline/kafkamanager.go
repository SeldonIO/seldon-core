package pipeline

import (
	"context"
	"fmt"
	"sync"

	"github.com/signalfx/splunk-otel-go/instrumentation/github.com/confluentinc/confluent-kafka-go/kafka/splunkkafka"
	"go.opentelemetry.io/otel"

	"github.com/seldonio/seldon-core/scheduler/pkg/envoy/resources"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	cmap "github.com/orcaman/concurrent-map"
	"github.com/rs/xid"
	"github.com/seldonio/seldon-core/scheduler/pkg/agent/config"
	kafka2 "github.com/seldonio/seldon-core/scheduler/pkg/kafka"
	"github.com/sirupsen/logrus"
)

const (
	pollTimeoutMillisecs = 100
)

type PipelineInferer interface {
	Infer(ctx context.Context, resourceName string, isModel bool, data []byte) ([]byte, error)
}

type KafkaManager struct {
	active     bool
	broker     string
	producer   *kafka.Producer
	pipelines  sync.Map
	logger     logrus.FieldLogger
	mu         sync.RWMutex
	configChan chan config.AgentConfiguration
	topicNamer *kafka2.TopicNamer
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
	isError  bool
}

func NewKafkaManager(logger logrus.FieldLogger, namespace string) *KafkaManager {
	return &KafkaManager{
		logger:     logger.WithField("source", "KafkaManager"),
		configChan: make(chan config.AgentConfiguration),
		topicNamer: kafka2.NewTopicNamer(namespace),
	}
}

func (km *KafkaManager) updateConfig(config *config.AgentConfiguration) {
	km.mu.Lock()
	defer km.mu.Unlock()
	logger := km.logger.WithField("func", "updateConfig")
	if config != nil && config.Kafka != nil {
		km.active = config.Kafka.Active
		km.broker = config.Kafka.Broker
		logger.Infof("Updated config to active %v broker %s", km.active, km.broker)
		if km.active {
			err := km.recreateProducer()
			if err != nil {
				logger.WithError(err).Error("Failed to update kafka producer")
			}
		}
	}
}

func (km *KafkaManager) StartConfigListener(configHandler *config.AgentConfigHandler) {
	logger := km.logger.WithField("func", "StartConfigListener")
	// Start config listener
	go km.listenForConfigUpdates()
	// Add ourself as listener on channel and handle initial config
	logger.Info("Loading initial stream configuration")
	km.updateConfig(configHandler.AddListener(km.configChan))
}

func (km *KafkaManager) listenForConfigUpdates() {
	logger := km.logger.WithField("func", "listenForConfigUpdates")
	for conf := range km.configChan {
		logger.Info("Received config update")
		conf := conf
		km.updateConfig(&conf)
	}
}

func (km *KafkaManager) recreateProducer() error {
	if km.producer != nil {
		km.producer.Close()
	}
	var err error
	var producerConfigMap = kafka.ConfigMap{
		"bootstrap.servers":   km.broker,
		"go.delivery.reports": false, // Need this othewise will get memory leak
		"linger.ms":           0,     // to ensure low latency - will need configuration in future
	}
	km.logger.Infof("Creating producer with broker %s", km.broker)
	km.producer, err = kafka.NewProducer(&producerConfigMap)
	return err
}

func (km *KafkaManager) createPipeline(resource string, isModel bool) (*Pipeline, error) {

	// Create consumer
	consumerConfig := kafka.ConfigMap{
		"broker.address.family": "v4",
		"group.id":              resource,
		"session.timeout.ms":    6000,
		"auto.offset.reset":     "earliest",
		"bootstrap.servers":     km.broker,
	}

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

func (km *KafkaManager) Infer(ctx context.Context, resourceName string, isModel bool, data []byte) ([]byte, error) {
	logger := km.logger.WithField("func", "Infer")
	km.mu.RLock()
	defer km.mu.RUnlock()
	pipeline, err := km.loadOrStorePipeline(resourceName, isModel)
	if err != nil {
		return nil, err
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
	kafkaHeaders := make([]kafka.Header, 1)
	kafkaHeaders[0] = kafka.Header{Key: resources.SeldonPipelineHeader, Value: []byte(resourceName)}

	msg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &outputTopic, Partition: kafka.PartitionAny},
		Key:            []byte(key),
		Value:          data,
		Headers:        kafkaHeaders,
	}

	// Add trace headers
	carrier := splunkkafka.NewMessageCarrier(msg)
	otel.GetTextMapPropagator().Inject(ctx, carrier)

	err = km.producer.Produce(msg, nil)
	if err != nil {
		return nil, err
	}
	request.wg.Wait()
	if request.isError {
		return nil, fmt.Errorf("%s", string(request.response))
	}
	return request.response, nil
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
					request := val.(*Request)
					request.mu.Lock()
					if request.active {
						request.isError = hasErrorHeader(e.Headers)
						request.response = e.Value
						request.wg.Done()
						request.active = false
					} else {
						logger.Warnf("Got duolicate request with key %s", string(e.Key))
					}
					request.mu.Unlock()
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
