package pipeline

import (
	"context"
	"sync"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	cmap "github.com/orcaman/concurrent-map"
	"github.com/seldonio/seldon-core/scheduler/pkg/kafka/config"
	"github.com/signalfx/splunk-otel-go/instrumentation/github.com/confluentinc/confluent-kafka-go/kafka/splunkkafka"
	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type MultiTopicsKafkaConsumer struct {
	config   *config.KafkaConfig
	logger   log.FieldLogger
	mu       sync.RWMutex
	topics   map[string]struct{}
	id       string
	consumer *kafka.Consumer
	isActive bool
	// map of kafka id to request
	requests cmap.ConcurrentMap
	tracer   trace.Tracer
}

func NewMultiTopicsKafkaConsumer(logger log.FieldLogger, consumerConfig *config.KafkaConfig, id string, tracer trace.Tracer) (*MultiTopicsKafkaConsumer, error) {
	consumer := &MultiTopicsKafkaConsumer{
		logger:   logger.WithField("source", "MultiTopicsKafkaConsumer"),
		config:   consumerConfig,
		mu:       sync.RWMutex{},
		topics:   make(map[string]struct{}),
		id:       id,
		isActive: false,
		requests: cmap.New(),
		tracer:   tracer,
	}
	err := consumer.createConsumer()
	if err == nil {
		consumer.isActive = true
	}
	return consumer, err
}

func (c *MultiTopicsKafkaConsumer) createConsumer() error {

	consumerConfig := config.CloneKafkaConfigMap(c.config.Consumer)
	consumerConfig["group.id"] = c.id
	c.logger.Infof("Creating consumer with config %v", consumerConfig)
	consumer, err := kafka.NewConsumer(&consumerConfig)
	if err != nil {
		return err
	}
	c.logger.Infof("Created consumer %s", c.id)
	c.consumer = consumer
	go func() {
		err := c.pollAndMatch()
		c.logger.WithError(err).Infof("Consumer %s failed", c.id)
	}()
	return nil
}

func (c *MultiTopicsKafkaConsumer) AddTopic(topic string, cb kafka.RebalanceCb) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.topics[topic]; ok {
		return nil
	}

	c.topics[topic] = struct{}{}
	return c.subscribeTopics(cb)
}

func (c *MultiTopicsKafkaConsumer) RemoveTopic(topic string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.topics[topic]; !ok {
		return nil
	}

	delete(c.topics, topic)
	if len(c.topics) == 0 {
		c.isActive = false
		return c.consumer.Close()
	} else {
		// TODO: we want to make sure that this does not affect the already existing subscription
		// specifically after we mark a given consumer to be ready initially (with a cb)
		return c.subscribeTopics(nil)
	}
}

func (c *MultiTopicsKafkaConsumer) GetNumTopics() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.topics)
}

func (c *MultiTopicsKafkaConsumer) GetConsumer() *kafka.Consumer {
	if c.isActive {
		return c.consumer
	} else {
		return nil
	}
}

func (c *MultiTopicsKafkaConsumer) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.isActive = false
	return c.consumer.Close()
}

func (c *MultiTopicsKafkaConsumer) subscribeTopics(cb kafka.RebalanceCb) error {
	topics := make([]string, len(c.topics))
	idx := 0
	for k := range c.topics {
		topics[idx] = k
		idx++
	}
	return c.consumer.SubscribeTopics(topics, cb)
}

func (c *MultiTopicsKafkaConsumer) pollAndMatch() error {
	for c.isActive {

		ev := c.consumer.Poll(pollTimeoutMillisecs)
		if ev == nil {
			continue
		}

		switch e := ev.(type) {
		case *kafka.Message:
			c.logger.Debugf("Received message from %s with key %s", *e.TopicPartition.Topic, string(e.Key))
			if val, ok := c.requests.Get(string(e.Key)); ok {

				// Add tracing span
				ctx := context.Background()
				carrierIn := splunkkafka.NewMessageCarrier(e)
				ctx = otel.GetTextMapPropagator().Extract(ctx, carrierIn)
				_, span := c.tracer.Start(ctx, "Consume")
				span.SetAttributes(attribute.String(RequestIdHeader, string(e.Key)))

				request := val.(*Request)
				request.mu.Lock()
				if request.active {
					c.logger.Debugf("Process response for key %s", string(e.Key))
					request.isError = hasErrorHeader(e.Headers)
					request.response = e.Value
					request.headers = e.Headers
					request.wg.Done()
					request.active = false
				} else {
					c.logger.Warnf("Got duplicate request with key %s", string(e.Key))
				}
				request.mu.Unlock()
				span.End()
			}

		case kafka.Error:
			c.logger.Error(e, "Received stream error")
			if e.Code() == kafka.ErrAllBrokersDown {
				break
			}
		default:
			c.logger.Debug("Ignored %s", e.String())
		}
	}
	c.isActive = false
	return c.consumer.Close()
}
