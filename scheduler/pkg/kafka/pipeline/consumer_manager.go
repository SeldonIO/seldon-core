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
	"fmt"
	"sync"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/trace"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/kafka/config"
)

const (
	EnvMaxNumConsumers             = "PIPELINEGATEWAY_MAX_NUM_CONSUMERS"
	EnvMaxNumTopicPerConsumer      = "PIPELINEGATEWAY_MAX_NUM_TOPICS_PER_CONSUMER"
	DefaultMaxNumTopicsPerConsumer = 100
	DefaultMaxNumConsumers         = 200
	kafkaConsumerNamePrefix        = "seldon-pipelinegateway"
)

type ConsumerManager struct {
	logger log.FieldLogger
	mu     sync.Mutex
	// all consumers we have
	consumers               []*MultiTopicsKafkaConsumer
	consumerConfig          *config.KafkaConfig
	maxNumConsumers         int
	maxNumTopicsPerConsumer int
	tracer                  trace.Tracer
}

func NewConsumerManager(
	logger log.FieldLogger,
	consumerConfig *config.KafkaConfig,
	maxNumTopicsPerConsumer,
	maxNumConsumers int,
	tracer trace.Tracer,
) *ConsumerManager {
	logger.Infof(
		"Setting consumer manager with max num consumers: %d, max topics per consumers: %d",
		maxNumConsumers,
		maxNumTopicsPerConsumer,
	)

	return &ConsumerManager{
		logger:                  logger.WithField("source", "ConsumerManager"),
		consumerConfig:          consumerConfig,
		maxNumTopicsPerConsumer: maxNumTopicsPerConsumer,
		maxNumConsumers:         maxNumConsumers,
		tracer:                  tracer,
	}
}

func (cm *ConsumerManager) createConsumer() error {
	if len(cm.consumers) == cm.maxNumTopicsPerConsumer {
		return fmt.Errorf("Max number of consumers reached")
	}

	c, err := NewMultiTopicsKafkaConsumer(
		cm.logger,
		cm.consumerConfig,
		getKafkaConsumerName(kafkaConsumerNamePrefix, uuid.New().String()),
		cm.tracer,
	)
	if err != nil {
		return err
	}

	cm.consumers = append(cm.consumers, c)
	return nil
}

func (cm *ConsumerManager) getKafkaConsumer() (*MultiTopicsKafkaConsumer, error) {
	// TODO: callers can get the same consumer and can AddTopics that can get this consumer beyond maxNumTopicsPerConsumer
	// this is fine for now
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if len(cm.consumers) == 0 {
		if err := cm.createConsumer(); err != nil {
			return nil, err
		}
	}
	c := cm.consumers[len(cm.consumers)-1]

	if c.GetNumTopics() < cm.maxNumTopicsPerConsumer {
		return c, nil
	} else {
		err := cm.createConsumer()
		if err != nil {
			return nil, err
		} else {
			return cm.consumers[len(cm.consumers)-1], nil
		}
	}
}

func (cm *ConsumerManager) GetNumModels() int {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	tot := 0
	for _, c := range cm.consumers {
		tot += c.GetNumTopics()
	}
	return tot
}

func (cm *ConsumerManager) Stop() {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	for _, c := range cm.consumers {
		err := c.Close()
		if err != nil {
			cm.logger.Warnf("Consumer %s failed to close", c.id)
		}
	}
}

func getKafkaConsumerName(prefix, id string) string {
	return fmt.Sprintf("%s-%s", prefix, id)
}
