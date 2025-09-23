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
	"fmt"
	"sync"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry"
	log "github.com/sirupsen/logrus"

	kafkaconfig "github.com/seldonio/seldon-core/components/kafka/v2/pkg/config"
	configtls "github.com/seldonio/seldon-core/components/tls/v2/pkg/config"

	seldontracer "github.com/seldonio/seldon-core/scheduler/v2/pkg/tracing"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/util"
)

const (
	modelGatewayConsumerNamePrefix = "seldon-modelgateway"
	EnvMaxNumConsumers             = "MODELGATEWAY_MAX_NUM_CONSUMERS"
	DefaultMaxNumConsumers         = 100
)

type ConsumerManager interface {
	Healthy() error
	AddModel(modelName string) error
	RemoveModel(modelName string, cleanTopicsOnDeletion bool, keepTopics bool) error
	GetNumModels() int
	Exists(modelName string) bool
	Stop()
}

type KafkaConsumerManager struct {
	logger log.FieldLogger
	mu     sync.RWMutex
	// all consumers we have
	consumers            map[string]*InferKafkaHandler
	managerConfig        *ManagerConfig
	maxNumConsumers      int
	consumerConfigMap    kafka.ConfigMap
	producerConfigMap    kafka.ConfigMap
	topicsConfigMap      kafka.ConfigMap
	schemaRegistryClient schemaregistry.Client
}

type ManagerConfig struct {
	SeldonKafkaConfig     *kafkaconfig.KafkaConfig
	Namespace             string
	InferenceServerConfig *InferenceServerConfig
	TraceProvider         *seldontracer.TracerProvider
	NumWorkers            int // infer workers
	WorkerTimeout         int // timeout for workers in ms
}

func cloneKafkaConfigMap(m kafka.ConfigMap) kafka.ConfigMap {
	m2 := make(kafka.ConfigMap)
	for k, v := range m {
		m2[k] = v
	}
	return m2
}

func NewKafkaConsumerManager(
	logger log.FieldLogger,
	managerConfig *ManagerConfig,
	maxNumConsumers int,
	schemaRegistryClient schemaregistry.Client,
) (*ConsumerManager, error) {
	cm := &ConsumerManager{
		logger:               logger.WithField("source", "ConsumerManager"),
		managerConfig:        managerConfig,
		consumers:            make(map[string]*InferKafkaHandler),
		maxNumConsumers:      maxNumConsumers,
		schemaRegistryClient: schemaRegistryClient,
	}
	err := cm.createKafkaConfigs(managerConfig)
	if err != nil {
		return nil, err
	}
	return cm, nil
}

func (cm *KafkaConsumerManager) createKafkaConfigs(kafkaConfig *ManagerConfig) error {
	logger := cm.logger.WithField("func", "createKafkaConfigs")
	var err error

	producerConfig := kafkaconfig.CloneKafkaConfigMap(kafkaConfig.SeldonKafkaConfig.Producer)
	producerConfig["go.delivery.reports"] = true
	err = configtls.AddKafkaSSLOptions(producerConfig)
	if err != nil {
		return err
	}

	producerConfigMasked := kafkaconfig.WithoutSecrets(producerConfig)
	producerConfigMaskedJSON, err := json.Marshal(&producerConfigMasked)
	if err != nil {
		logger.WithField("config", &producerConfigMasked).Info("Creating producer config for use later")
	} else {
		logger.WithField("config", string(producerConfigMaskedJSON)).Info("Creating producer config for use later")
	}

	consumerConfig := kafkaconfig.CloneKafkaConfigMap(kafkaConfig.SeldonKafkaConfig.Consumer)
	err = configtls.AddKafkaSSLOptions(consumerConfig)
	if err != nil {
		return err
	}

	consumerConfigMasked := kafkaconfig.WithoutSecrets(consumerConfig)
	consumerConfigMaskedJson, err := json.Marshal(&consumerConfigMasked)
	if err != nil {
		logger.WithField("config", &consumerConfigMasked).Info("Creating consumer config for use later")
	} else {
		logger.WithField("config", string(consumerConfigMaskedJson)).Info("Creating consumer config for use later")
	}

	topicsConfig := kafkaconfig.CloneKafkaConfigMap(kafkaConfig.SeldonKafkaConfig.Topics)
	topicsConfigJSON, err := json.Marshal(&topicsConfig)
	if err != nil {
		logger.WithField("config", &topicsConfig).Info("Creating topics config for use later")
	} else {
		logger.WithField("config", string(topicsConfigJSON)).Info("Creating topics config for use later")
	}

	cm.consumerConfigMap = consumerConfig
	cm.producerConfigMap = producerConfig
	cm.topicsConfigMap = topicsConfig
	return nil
}

func (cm *KafkaConsumerManager) getInferKafkaConsumer(modelName string, create bool) (*InferKafkaHandler, error) {
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
			cloneKafkaConfigMap(cm.topicsConfigMap),
			consumerBucketId, cm.schemaRegistryClient)
		if err != nil {
			return nil, err
		}

		go ic.Serve()
		logger.Debugf("Created consumer for model %s with bucket id %s", modelName, consumerBucketId)
		cm.consumers[consumerBucketId] = ic
	}
	return ic, nil
}

func (cm *KafkaConsumerManager) Healthy() error {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	for _, consumer := range cm.consumers {
		if !consumer.producerIsActive() {
			return fmt.Errorf("producer %s not active", consumer.producer.String())
		}
		if consumer.consumer.IsClosed() {
			return fmt.Errorf("consumer %s is closed", consumer.consumer.String())
		}
	}

	return nil
}

func (cm *KafkaConsumerManager) AddModel(modelName string) error {
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

func (cm *KafkaConsumerManager) stopEmptyConsumer(ic *InferKafkaHandler) {
	logger := cm.logger.WithField("func", "stopEmptyConsumer")
	numModelsInConsumer := ic.GetNumModels()
	if numModelsInConsumer == 0 {
		logger.Debugf("Deleting consumer with no models with bucket id %s", ic.consumerName)
		ic.Stop(false)
		delete(cm.consumers, ic.consumerName)
	}
}

func (cm *KafkaConsumerManager) RemoveModel(modelName string, cleanTopicsOnDeletion bool, keepTopics bool) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	ic, err := cm.getInferKafkaConsumer(modelName, false)
	if err != nil {
		return err
	}
	if ic == nil {
		return nil
	}

	cm.logger.WithField("model", modelName).Info("Removing model from consumer")

	err = ic.RemoveModel(modelName, cleanTopicsOnDeletion, keepTopics)
	if err != nil {
		return err
	}
	cm.stopEmptyConsumer(ic)
	return nil
}

func (cm *KafkaConsumerManager) GetNumModels() int {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	tot := 0
	for _, ic := range cm.consumers {
		tot += ic.GetNumModels()
	}
	return tot
}

func (cm *KafkaConsumerManager) Exists(modelName string) bool {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	ic, err := cm.getInferKafkaConsumer(modelName, false)
	if err != nil {
		return false
	}

	return ic != nil && ic.Exists(modelName)
}

func (cm *KafkaConsumerManager) Stop() {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.logger.Infof("Stopping")
	// stop all consumers
	cm.logger.Infof("Number of consumers to stop: %d", len(cm.consumers))

	wg := &sync.WaitGroup{}
	wg.Add(len(cm.consumers))
	// consumers can take a while to shut down, we shut them down concurrently to mitigate risk of exceeded the k8s
	// terminate graceful shutdown period (terminationGracePeriodSeconds)
	for bucketID, ic := range cm.consumers {
		go func() {
			defer wg.Done()
			cm.logger.WithField("bucket_id", bucketID).Info("Stopping consumer")
			delete(cm.consumers, bucketID)
			// we block until the consumers have completed their shutdown sequence with kafka and communicated their
			// latest offsets
			ic.Stop(true)
		}()
	}

	cm.logger.Infof("Waiting for all consumers to stop")
	wg.Wait()
	cm.logger.Infof("All consumers stopped")
}
