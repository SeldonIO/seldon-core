package gateway

import (
	"fmt"
	"sync"

	"github.com/seldonio/seldon-core/scheduler/pkg/kafka/config"
	seldontracer "github.com/seldonio/seldon-core/scheduler/pkg/tracing"
	log "github.com/sirupsen/logrus"
)

const (
	modelGatewayConsumerNamePrefix = "modelgateway"
	EnvMaxNumConsumers             = "MODELGATEWAY_MAX_NUM_CONSUMERS"
	DefaultMaxNumConsumers         = 100
)

type ConsumerManager struct {
	logger log.FieldLogger
	mu     sync.Mutex
	// all consumers we have
	consumers       map[string]*InferKafkaConsumer
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
		consumers:       make(map[string]*InferKafkaConsumer),
		maxNumConsumers: maxNumConsumers,
	}
}

func (cm *ConsumerManager) getKafkaConsumerName(modelName string) string {
	idx := modelIdToConsumerBucket(modelName, cm.maxNumConsumers)
	return fmt.Sprintf("%s-%d", modelGatewayConsumerNamePrefix, idx)
}

func (cm *ConsumerManager) getInferKafkaConsumer(modelName string, create bool) (*InferKafkaConsumer, error) {
	consumerBucketId := cm.getKafkaConsumerName(modelName)
	ic, ok := cm.consumers[consumerBucketId]

	if !ok && !create {
		return nil, nil
	}

	if !ok {
		var err error
		ic, err = NewInferKafkaConsumer(cm.logger, cm.consumerConfig, consumerBucketId)
		if err != nil {
			return nil, err
		}
		go ic.Serve()
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

func (cm *ConsumerManager) stopEmptyConsumer(ic *InferKafkaConsumer) {
	numModelsInConsumer := ic.GetNumModels()
	if numModelsInConsumer == 0 {
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
