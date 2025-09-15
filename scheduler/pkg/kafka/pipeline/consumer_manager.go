/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package pipeline

import (
	"sync"

	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry"
	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/trace"

	kafkaconfig "github.com/seldonio/seldon-core/components/kafka/v2/pkg/config"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/util"
)

const (
	EnvMaxNumConsumers                = "PIPELINEGATEWAY_MAX_NUM_CONSUMERS"
	DefaultMaxNumConsumers            = 100
	pipelineGatewayConsumerNamePrefix = "seldon-pipelinegateway"
)

type ConsumerManager struct {
	logger log.FieldLogger
	mu     sync.Mutex
	// all consumers we have
	pipelinesConsumers   map[string]*MultiTopicsKafkaConsumer
	modelsConsumers      map[string]*MultiTopicsKafkaConsumer
	consumerConfig       *kafkaconfig.KafkaConfig
	maxNumConsumers      int
	tracer               trace.Tracer
	namespace            string
	schemaRegistryClient schemaregistry.Client
}

func NewConsumerManager(
	namespace string,
	logger log.FieldLogger,
	consumerConfig *kafkaconfig.KafkaConfig,
	maxNumConsumers int,
	tracer trace.Tracer,
	schemaRegistryClient schemaregistry.Client,
) *ConsumerManager {
	logger.
		WithField("max consumers", maxNumConsumers).
		Info("creating consumer manager")

	return &ConsumerManager{
		namespace:            namespace,
		logger:               logger.WithField("source", "ConsumerManager"),
		pipelinesConsumers:   make(map[string]*MultiTopicsKafkaConsumer),
		modelsConsumers:      make(map[string]*MultiTopicsKafkaConsumer),
		consumerConfig:       consumerConfig,
		maxNumConsumers:      maxNumConsumers,
		tracer:               tracer,
		schemaRegistryClient: schemaRegistryClient,
	}
}

func (cm *ConsumerManager) createConsumer(consumerName string, consumers map[string]*MultiTopicsKafkaConsumer) (*MultiTopicsKafkaConsumer, error) {
	c, err := NewMultiTopicsKafkaConsumer(
		cm.logger,
		cm.consumerConfig,
		consumerName,
		cm.tracer,
		cm.schemaRegistryClient,
	)
	if err != nil {
		return nil, err
	}
	consumers[consumerName] = c
	return c, nil
}

func (cm *ConsumerManager) getKafkaConsumer(pipelineOrModelName string, isModel bool) (*MultiTopicsKafkaConsumer, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	consumerName := util.GetKafkaConsumerName(
		cm.namespace,
		cm.consumerConfig.ConsumerGroupIdPrefix,
		pipelineOrModelName,
		pipelineGatewayConsumerNamePrefix,
		cm.maxNumConsumers,
	)
	consumers := cm.pipelinesConsumers
	cm.logger.Debugf("Getting consumer for %s with name %s", pipelineOrModelName, consumerName)
	if isModel {
		consumers = cm.modelsConsumers
	}
	if consumer, ok := consumers[consumerName]; ok {
		return consumer, nil
	}
	return cm.createConsumer(consumerName, consumers)
}

func stop(consumers map[string]*MultiTopicsKafkaConsumer) {
	for _, c := range consumers {
		err := c.Close()
		if err != nil {
			log.Warnf("Consumer %s failed to close", c.id)
		}
	}
}

func (cm *ConsumerManager) Stop() {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	stop(cm.modelsConsumers)
	stop(cm.pipelinesConsumers)
}
