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

package gateway

import (
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/kafka/config"
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
	consumers       map[string]*InferKafkaHandler
	consumerConfig  *ConsumerConfig
	maxNumConsumers int
}

type ConsumerConfig struct {
	KafkaConfig           *config.KafkaConfig
	Namespace             string
	InferenceServerConfig *InferenceServerConfig
	TraceProvider         *seldontracer.TracerProvider
	NumWorkers            int // infer workers
}

func NewConsumerManager(logger log.FieldLogger, consumerConfig *ConsumerConfig, maxNumConsumers int) *ConsumerManager {
	return &ConsumerManager{
		logger:          logger.WithField("source", "ConsumerManager"),
		consumerConfig:  consumerConfig,
		consumers:       make(map[string]*InferKafkaHandler),
		maxNumConsumers: maxNumConsumers,
	}
}

func (cm *ConsumerManager) getInferKafkaConsumer(modelName string, create bool) (*InferKafkaHandler, error) {
	logger := cm.logger.WithField("func", "getInferKafkaConsumer")
	consumerBucketId := util.GetKafkaConsumerName(modelName, modelGatewayConsumerNamePrefix, cm.maxNumConsumers)
	ic, ok := cm.consumers[consumerBucketId]
	logger.Debugf("Getting consumer for model %s with bucket id %s", modelName, consumerBucketId)
	if !ok && !create {
		return nil, nil
	}

	if !ok {
		var err error
		ic, err = NewInferKafkaHandler(cm.logger, cm.consumerConfig, consumerBucketId)
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
