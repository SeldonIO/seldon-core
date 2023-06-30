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

package pipeline

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	cmap "github.com/orcaman/concurrent-map"
	"github.com/signalfx/splunk-otel-go/instrumentation/github.com/confluentinc/confluent-kafka-go/kafka/splunkkafka"
	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/kafka/config"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/util"
)

type MultiTopicsKafkaConsumer struct {
	config   *config.KafkaConfig
	logger   log.FieldLogger
	mu       sync.RWMutex
	topics   map[string]struct{}
	id       string
	consumer *kafka.Consumer
	isActive atomic.Bool
	// map of kafka id to request
	requests cmap.ConcurrentMap
	tracer   trace.Tracer
}

func NewMultiTopicsKafkaConsumer(
	logger log.FieldLogger,
	consumerConfig *config.KafkaConfig,
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
	err := consumer.createConsumer()
	return consumer, err
}

func (c *MultiTopicsKafkaConsumer) createConsumer() error {

	consumerConfig := config.CloneKafkaConfigMap(c.config.Consumer)
	consumerConfig["group.id"] = c.id
	err := config.AddKafkaSSLOptions(consumerConfig)
	if err != nil {
		return err
	}
	c.logger.Infof("Creating consumer with config %v", consumerConfig)
	consumer, err := kafka.NewConsumer(&consumerConfig)
	if err != nil {
		return err
	}
	c.logger.Infof("Created consumer %s", c.id)
	c.consumer = consumer
	c.isActive.Store(true)
	go func() {
		err := c.pollAndMatch()
		c.logger.WithError(err).Infof("Consumer %s failed and is ending", c.id)
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
		return c.Close()
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

func (c *MultiTopicsKafkaConsumer) Close() error {
	if c.isActive.Load() {
		c.isActive.Store(false)
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
		if ev == nil {
			continue
		}

		switch e := ev.(type) {
		case *kafka.Message:
			logger.Debugf("Received message from %s with key %s", *e.TopicPartition.Topic, string(e.Key))
			if val, ok := c.requests.Get(string(e.Key)); ok {

				// Add tracing span
				ctx := context.Background()
				carrierIn := splunkkafka.NewMessageCarrier(e)
				ctx = otel.GetTextMapPropagator().Extract(ctx, carrierIn)
				_, span := c.tracer.Start(ctx, "Consume")
				span.SetAttributes(attribute.String(util.RequestIdHeader, string(e.Key)))

				request := val.(*Request)
				request.mu.Lock()
				if request.active {
					logger.Debugf("Process response for key %s", string(e.Key))
					request.errorModel, request.isError = extractErrorHeader(e.Headers)
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
			logger.Errorf("Kafka error, code: [%s] msg: [%s]", e.Code().String(), e.Error())
		default:
			logger.Debugf("Ignored %s", e.String())
		}
	}
	logger.Warning("Ending kafka consumer poll")
	return nil // assumption here is that the connection has already terminated
}
