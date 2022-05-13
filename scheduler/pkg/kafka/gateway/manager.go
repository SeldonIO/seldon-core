package gateway

import (
	"sync"

	"github.com/seldonio/seldon-core/scheduler/pkg/kafka/config"

	seldontracer "github.com/seldonio/seldon-core/scheduler/pkg/tracing"

	"github.com/seldonio/seldon-core/scheduler/pkg/kafka"

	"github.com/seldonio/seldon-core/scheduler/apis/mlops/scheduler"

	log "github.com/sirupsen/logrus"
)

const DEFAULT_NWORKERS = 4

type KafkaManager struct {
	logger         log.FieldLogger
	mu             sync.Mutex
	gateways       map[string]*InferKafkaGateway //internal model name to infer consumer
	broker         string
	serverConfig   *KafkaServerConfig
	kafkaConfig    *config.KafkaConfig
	topicNamer     *kafka.TopicNamer
	tracerProvider *seldontracer.TracerProvider
}

func (km *KafkaManager) Name() string {
	panic("implement me")
}

func NewKafkaManager(logger log.FieldLogger, serverConfig *KafkaServerConfig, namespace string, kafkaConfig *config.KafkaConfig, traceProvider *seldontracer.TracerProvider) *KafkaManager {
	return &KafkaManager{
		logger:         logger.WithField("source", "KafkaManager"),
		gateways:       make(map[string]*InferKafkaGateway),
		serverConfig:   serverConfig,
		kafkaConfig:    kafkaConfig,
		topicNamer:     kafka.NewTopicNamer(namespace),
		tracerProvider: traceProvider,
	}
}

func (km *KafkaManager) Stop() error {
	km.mu.Lock()
	defer km.mu.Unlock()
	for _, ic := range km.gateways {
		ic.Stop()
	}
	return nil
}

func (km *KafkaManager) AddModel(modelName string, streamSpec *scheduler.StreamSpec) error {
	logger := km.logger.WithField("func", "AddModel")
	km.mu.Lock()
	defer km.mu.Unlock()
	if _, ok := km.gateways[modelName]; ok {
		logger.Infof("kafka gateway already exists for model %s", modelName)
	} else {
		modelConfig := &KafkaModelConfig{
			ModelName:   modelName,
			InputTopic:  km.topicNamer.GetModelTopicInputs(modelName),
			OutputTopic: km.topicNamer.GetModelTopicOutputs(modelName),
			ErrorTopic:  km.topicNamer.GetModelErrorTopic(),
		}
		if streamSpec != nil {
			if streamSpec.InputTopic != "" {
				modelConfig.InputTopic = streamSpec.InputTopic
			}
			if streamSpec.OutputTopic != "" {
				modelConfig.OutputTopic = streamSpec.OutputTopic
			}
		}
		km.logger.Infof("Adding consumer to broker %s for model %s, input topic %s output topic %s", km.broker, modelName, modelConfig.InputTopic, modelConfig.OutputTopic)
		inferGateway, err := NewInferKafkaGateway(km.logger, DEFAULT_NWORKERS, km.kafkaConfig, modelConfig, km.serverConfig, km.tracerProvider)
		km.gateways[modelName] = inferGateway
		if err != nil {
			return err
		}
		go inferGateway.Serve()
	}
	return nil
}

func (km *KafkaManager) RemoveModel(modelName string) {
	km.mu.Lock()
	defer km.mu.Unlock()
	inferConsumer, ok := km.gateways[modelName]
	if ok {
		inferConsumer.Stop()
		delete(km.gateways, modelName)
	}
}
