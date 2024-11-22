/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package gateway

import (
	"encoding/json"
	"sync"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	log "github.com/sirupsen/logrus"

	config_tls "github.com/seldonio/seldon-core/components/tls/v2/pkg/config"
	kafka_config "github.com/seldonio/seldon-core/components/kafka/v2/pkg/config"
	seldontracer "github.com/seldonio/seldon-core/scheduler/v2/pkg/tracing"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/util"
)

const (
	modelGatewayConsumerNamePrefix = "seldon-modelgateway"
	EnvMaxNumConsumers             = "MODELGATEWAY_MAX_NUM_CONSUMERS"
	DefaultMaxNumConsumers         = 100
)

type ConsumerManager struct {
	logger log.FieldLogger
	mu     sync.Mutex
	// all consumers we have
	consumers         map[string]*InferKafkaHandler
	managerConfig     *ManagerConfig
	maxNumConsumers   int
	consumerConfigMap kafka.ConfigMap
	producerConfigMap kafka.ConfigMap
}

type ManagerConfig struct {
	SeldonKafkaConfig     *kafka_config.KafkaConfig
	Namespace             string
	InferenceServerConfig *InferenceServerConfig
	TraceProvider         *seldontracer.TracerProvider
	NumWorkers            int // infer workers
}

func cloneKafkaConfigMap(m kafka.ConfigMap) kafka.ConfigMap {
	m2 := make(kafka.ConfigMap)
	for k, v := range m {
		m2[k] = v
	}
	return m2
}

func NewConsumerManager(
	logger log.FieldLogger,
	managerConfig *ManagerConfig,
	maxNumConsumers int,
) (*ConsumerManager, error) {
	cm := &ConsumerManager{
		logger:          logger.WithField("source", "ConsumerManager"),
		managerConfig:   managerConfig,
		consumers:       make(map[string]*InferKafkaHandler),
		maxNumConsumers: maxNumConsumers,
	}
	err := cm.createKafkaConfigs(managerConfig)
	if err != nil {
		return nil, err
	}
	return cm, nil
}

func (cm *ConsumerManager) createKafkaConfigs(kafkaConfig *ManagerConfig) error {
	logger := cm.logger.WithField("func", "createKafkaConfigs")
	var err error

	producerConfig := kafka_config.CloneKafkaConfigMap(kafkaConfig.SeldonKafkaConfig.Producer)
	producerConfig["go.delivery.reports"] = true
	err = config_tls.AddKafkaSSLOptions(producerConfig)
	if err != nil {
		return err
	}

	producerConfigMasked := kafka_config.WithoutSecrets(producerConfig)
	producerConfigMaskedJSON, err := json.Marshal(&producerConfigMasked)
	if err != nil {
		logger.WithField("config", &producerConfigMasked).Info("Creating producer config for use later")
	} else {
		logger.WithField("config", string(producerConfigMaskedJSON)).Info("Creating producer config for use later")
	}

	consumerConfig := kafka_config.CloneKafkaConfigMap(kafkaConfig.SeldonKafkaConfig.Consumer)
	err = config_tls.AddKafkaSSLOptions(consumerConfig)
	if err != nil {
		return err
	}

	consumerConfigMasked := kafka_config.WithoutSecrets(consumerConfig)
	consumerConfigMaskedJson, err := json.Marshal(&consumerConfigMasked)
	if err != nil {
		logger.WithField("config", &consumerConfigMasked).Info("Creating consumer config for use later")
	} else {
		logger.WithField("config", string(consumerConfigMaskedJson)).Info("Creating consumer config for use later")
	}

	cm.consumerConfigMap = consumerConfig
	cm.producerConfigMap = producerConfig
	return nil
}

func (cm *ConsumerManager) getInferKafkaConsumer(modelName string, create bool) (*InferKafkaHandler, error) {
	logger := cm.logger.WithField("func", "getInferKafkaConsumer")

	consumerBucketId := util.GetKafkaConsumerName(
		cm.managerConfig.Namespace,
		cm.managerConfig.SeldonKafkaConfig.ConsumerGroupIdPrefix,
		modelName,
		modelGatewayConsumerNamePrefix,
		cm.maxNumConsumers)
	ic, ok := cm.consumers[consumerBucketId]
	logger.Debugf("Getting consumer for model %s with bucket id %s", modelName, consumerBucketId)
	if !ok && !create {
		return nil, nil
	}

	if !ok {
		var err error
		ic, err = NewInferKafkaHandler(cm.logger,
			cm.managerConfig,
			cloneKafkaConfigMap(cm.consumerConfigMap),
			cloneKafkaConfigMap(cm.producerConfigMap),
			consumerBucketId)
		if err != nil {
			return nil, err
		}

		go ic.Serve()
		logger.Debugf("Created consumer for model %s with bucket id %s", modelName, consumerBucketId)
		cm.consumers[consumerBucketId] = ic
	}
	return ic, nil
}

func (cm *ConsumerManager) AddModel(modelName string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	ic, err := cm.getInferKafkaConsumer(modelName, true)
	if err != nil {
		return err
	}
	err = ic.AddModel(modelName)
	if err != nil {
		return err
	}

	return nil
}

func (cm *ConsumerManager) stopEmptyConsumer(ic *InferKafkaHandler) {
	logger := cm.logger.WithField("func", "stopEmptyConsumer")
	numModelsInConsumer := ic.GetNumModels()
	if numModelsInConsumer == 0 {
		logger.Debugf("Deleting consumer with no models with bucket id %s", ic.consumerName)
		ic.Stop()
		delete(cm.consumers, ic.consumerName)
	}
}

func (cm *ConsumerManager) RemoveModel(modelName string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	ic, err := cm.getInferKafkaConsumer(modelName, false)
	if err != nil {
		return err
	}
	if ic == nil {
		return nil
	}
	err = ic.RemoveModel(modelName)
	if err != nil {
		return err
	}
	cm.stopEmptyConsumer(ic)
	return nil
}

func (cm *ConsumerManager) GetNumModels() int {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	tot := 0
	for _, ic := range cm.consumers {
		tot += ic.GetNumModels()
	}
	return tot
}

func (cm *ConsumerManager) Exists(modelName string) bool {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	ic, err := cm.getInferKafkaConsumer(modelName, false)
	if err != nil {
		return false
	}

	return ic != nil && ic.Exists(modelName)
}

func (cm *ConsumerManager) Stop() {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.logger.Infof("Stopping")
	// stop all consumers
	cm.logger.Infof("Number of consumers to stop %d", len(cm.consumers))
	for icKey, ic := range cm.consumers {
		cm.logger.Info("Stopping consumer")
		delete(cm.consumers, icKey)
		ic.Stop()
	}
}
