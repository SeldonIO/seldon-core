package gateway

import (
	"sync"

	"github.com/seldonio/seldon-core/scheduler/pkg/kafka/config"
	seldontracer "github.com/seldonio/seldon-core/scheduler/pkg/tracing"
	log "github.com/sirupsen/logrus"
)

const (
	EnvMaxModelsPerConsumer     = "MODELGATEWAY_MAX_MODELS_PER_CONSUMER"
	DefaultMaxModelsPerConsumer = 1
)

type ConsumerManager struct {
	logger log.FieldLogger
	mu     sync.Mutex
	// all consumers we have
	consumers map[*InferKafkaConsumer]bool
	// consumers that can take more models
	availableConsumers map[*InferKafkaConsumer]bool
	// which consumer a model is connected to
	modelToConsumer      map[string]*InferKafkaConsumer
	maxModelsPerConsumer int
	consumerConfig       *ConsumerConfig
}

type ConsumerConfig struct {
	KafkaConfig           *config.KafkaConfig
	Namespace             string
	InferenceServerConfig *InferenceServerConfig
	TraceProvider         *seldontracer.TracerProvider
	NumWorkers            int
}

func NewConsumerManager(logger log.FieldLogger, consumerConfig *ConsumerConfig, maxModelsPerConsumer int) *ConsumerManager {
	return &ConsumerManager{
		logger:               logger.WithField("source", "ConsumerManager"),
		maxModelsPerConsumer: maxModelsPerConsumer,
		consumerConfig:       consumerConfig,
		consumers:            make(map[*InferKafkaConsumer]bool),
		modelToConsumer:      make(map[string]*InferKafkaConsumer),
		availableConsumers:   make(map[*InferKafkaConsumer]bool),
	}
}

func (cm *ConsumerManager) createConsumer() (*InferKafkaConsumer, error) {
	return NewInferKafkaConsumer(cm.logger, cm.consumerConfig)
}

// At present this used to get any consumer. We don't need it to be truly random.
func (cm *ConsumerManager) getRandomConsumer() *InferKafkaConsumer {
	for k := range cm.availableConsumers {
		return k
	}
	return nil
}

func (cm *ConsumerManager) getInferKafkaConsumer() (*InferKafkaConsumer, error) {
	var ic *InferKafkaConsumer
	var err error
	// Find or create a consumer
	if len(cm.availableConsumers) == 0 {
		ic, err = NewInferKafkaConsumer(cm.logger, cm.consumerConfig)
		if err != nil {
			return nil, err
		}
		go ic.Serve()
		cm.consumers[ic] = true
	} else {
		ic = cm.getRandomConsumer()
	}
	return ic, nil
}

func (cm *ConsumerManager) AddModel(modelName string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	ic := cm.modelToConsumer[modelName]
	if ic == nil {
		ic, err := cm.getInferKafkaConsumer()
		if err != nil {
			return err
		}
		err = ic.AddModel(modelName)
		if err != nil {
			return err
		}
		cm.modelToConsumer[modelName] = ic
		// Push back onto stack if still has space for more models
		if ic.GetNumModels() >= cm.maxModelsPerConsumer {
			delete(cm.availableConsumers, ic)
		} else {
			cm.availableConsumers[ic] = true
		}
	} else {
		ic.logger.Infof("Ignore add model for %s as consumer already exists", modelName)
	}
	return nil
}

func (cm *ConsumerManager) updateAvailableConsumer(ic *InferKafkaConsumer) {
	numModelsInConsumer := ic.GetNumModels()
	if numModelsInConsumer == 0 {
		ic.Stop()
		delete(cm.consumers, ic)
		delete(cm.availableConsumers, ic)
	} else if numModelsInConsumer == cm.maxModelsPerConsumer-1 {
		cm.availableConsumers[ic] = true
	}
}

func (cm *ConsumerManager) RemoveModel(modelName string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	ic := cm.modelToConsumer[modelName]
	if ic == nil {
		return nil
	}
	err := ic.RemoveModel(modelName)
	if err != nil {
		return err
	}
	delete(cm.modelToConsumer, modelName)
	cm.updateAvailableConsumer(ic)
	return nil
}

func (cm *ConsumerManager) GetNumModels() int {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	tot := 0
	for ic := range cm.consumers {
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
	for ic := range cm.consumers {
		cm.logger.Infof("Stopping consumer")
		delete(cm.consumers, ic)
		ic.Stop()
	}
}
