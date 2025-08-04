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
	"sync"
	"sync/atomic"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	cmap "github.com/orcaman/concurrent-map"
	"github.com/signalfx/splunk-otel-go/instrumentation/github.com/confluentinc/confluent-kafka-go/v2/kafka/splunkkafka"
	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	kafka_config "github.com/seldonio/seldon-core/components/kafka/v2/pkg/config"
	config_tls "github.com/seldonio/seldon-core/components/tls/v2/pkg/config"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/util"
)

type MultiTopicsKafkaConsumer struct {
	config     *kafka_config.KafkaConfig
	logger     log.FieldLogger
	mu         sync.RWMutex
	topics     map[string]struct{}
	partitions []int32
	id         string
	consumer   *kafka.Consumer
	isActive   atomic.Bool
	// map of kafka id to request
	requests cmap.ConcurrentMap
	tracer   trace.Tracer
	wg       sync.WaitGroup
}

func NewMultiTopicsKafkaConsumer(
	logger log.FieldLogger,
	consumerConfig *kafka_config.KafkaConfig,
	id string,
	tracer trace.Tracer,
) (*MultiTopicsKafkaConsumer, error) {
	consumer := &MultiTopicsKafkaConsumer{
		logger:   logger.WithField("source", "MultiTopicsKafkaConsumer"),
		config:   consumerConfig,
		mu:       sync.RWMutex{},
		topics:   make(map[string]struct{}),
		id:       id,
		requests: cmap.New(),
		tracer:   tracer,
	}
	err := consumer.createConsumer(logger)
	return consumer, err
}

func (c *MultiTopicsKafkaConsumer) createConsumer(logger log.FieldLogger) error {
	consumerConfig := kafka_config.CloneKafkaConfigMap(c.config.Consumer)
	consumerConfig["group.id"] = c.id
	consumerConfig["go.application.rebalance.enable"] = true
	consumerConfig["partition.assignment.strategy"] = "roundrobin"

	err := config_tls.AddKafkaSSLOptions(consumerConfig)
	if err != nil {
		return err
	}

	configWithoutSecrets := kafka_config.WithoutSecrets(consumerConfig)
	c.logger.Infof("Creating consumer with config %v", configWithoutSecrets)
	consumer, err := kafka.NewConsumer(&consumerConfig)
	if err != nil {
		return err
	}

	c.logger.Infof("Created consumer %s", c.id)
	c.consumer = consumer
	c.isActive.Store(true)

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		err := c.pollAndMatch()
		c.logger.WithError(err).Infof("Consumer %s is ending", c.id)
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

func (c *MultiTopicsKafkaConsumer) RemoveTopic(topic string, cb kafka.RebalanceCb) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.topics[topic]; !ok {
		return nil
	}

	delete(c.topics, topic)
	if len(c.topics) == 0 {
		return c.Close()
	} else {
		// TODO: we want to make sure that this does not affect the already existing subscription
		// specifically after we mark a given consumer to be ready initially (with a cb)
		return c.subscribeTopics(cb)
	}
}

func (c *MultiTopicsKafkaConsumer) GetNumTopics() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.topics)
}

func (c *MultiTopicsKafkaConsumer) Close() error {
	if c.isActive.Load() {
		// Stop the consumer from polling
		c.isActive.Store(false)
		c.wg.Wait()

		// Explicitly release partitions
		err := c.consumer.Unassign()
		if err != nil {
			c.logger.Errorf("Error unassigning partitions: %v", err)
		}

		return c.consumer.Close()
	}
	return nil
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
	logger := c.logger.WithField("func", "pollAndMatch")
	for c.isActive.Load() {
		ev := c.consumer.Poll(pollTimeoutMillisecs)

		// Closing the consumer cleanly. First stop polling and then close the connection.
		if !c.isActive.Load() {
			logger.Info("Consumer is not active, stopping poll")
			return nil
		}

		if ev == nil {
			continue
		}

		switch e := ev.(type) {
		case *kafka.Message:
			logger.
				WithField("topic", *e.TopicPartition.Topic).
				WithField("key", string(e.Key)).
				Debugf("received message")

			key := string(e.Key)
			if val, ok := c.requests.Get(key); ok {
				ctx := createBaseContextFromKafkaMsg(e)

				// Add tracing span
				_, span := c.tracer.Start(ctx, "Consume")
				// Use the original request id from kafka headers, as key here is a composite key with the resource name
				requestId := GetRequestIdFromKafkaHeaders(e.Headers)
				if requestId == "" {
					logger.Warnf("Missing request id in Kafka headers for key %s", key)
				}
				span.SetAttributes(attribute.String(util.RequestIdHeader, requestId))

				request := val.(*Request)
				request.mu.Lock()
				if request.active {
					logger.Debugf("Process response for key %s", key)
					request.errorModel, request.isError = extractErrorHeader(e.Headers)
					request.response = e.Value
					request.headers = e.Headers
					request.wg.Done()
					request.active = false
				} else {
					logger.Warnf("Got duplicate request with key %s", key)
				}
				request.mu.Unlock()
				span.End()
			}

		case kafka.Error:
			logger.Errorf("Kafka error, code: [%s] msg: [%s]", e.Code().String(), e.Error())
		default:
			logger.Debugf("Ignored %s", e.String())
		}
	}
	logger.Warning("Ending kafka consumer poll")
	return nil // assumption here is that the connection has already terminated
}

func createBaseContextFromKafkaMsg(msg *kafka.Message) context.Context {
	// these are just a base context for a new span
	// callers should add timeout, etc for this context as they see fit.
	ctx := context.Background()
	carrierIn := splunkkafka.NewMessageCarrier(msg)
	return otel.GetTextMapPropagator().Extract(ctx, carrierIn)
}
