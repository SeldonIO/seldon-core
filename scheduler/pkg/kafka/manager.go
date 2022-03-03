package kafka

import (
	"fmt"
	"sync"

	"github.com/seldonio/seldon-core/scheduler/apis/mlops/scheduler"
	"github.com/seldonio/seldon-core/scheduler/pkg/agent/config"

	log "github.com/sirupsen/logrus"
)

const DEFAULT_NWORKERS = 4

type KafkaManager struct {
	active       bool
	logger       log.FieldLogger
	mu           sync.Mutex
	gateways     map[string]*InferKafkaGateway //internal model name to infer consumer
	broker       string
	serverConfig *KafkaServerConfig
	configChan   chan config.AgentConfiguration
}

func (km *KafkaManager) Name() string {
	panic("implement me")
}

func NewKafkaManager(logger log.FieldLogger, serverConfig *KafkaServerConfig) *KafkaManager {
	return &KafkaManager{
		active:       false,
		logger:       logger.WithField("source", "KafkaManager"),
		gateways:     make(map[string]*InferKafkaGateway),
		serverConfig: serverConfig,
		configChan:   make(chan config.AgentConfiguration),
	}
}

func (km *KafkaManager) updateConfig(config *config.AgentConfiguration) {
	logger := km.logger.WithField("func", "updateConfig")
	km.mu.Lock()
	defer km.mu.Unlock()
	if config != nil && config.Kafka != nil {
		km.active = config.Kafka.Active
		km.broker = config.Kafka.Broker
		logger.Infof("Updating config active %v broker %s", km.active, km.broker)
	}
}

func (km *KafkaManager) StartConfigListener(configHandler *config.AgentConfigHandler) {
	logger := km.logger.WithField("func", "StartConfigListener")
	// Start config listener
	go km.listenForConfigUpdates()
	// Add ourself as listener on channel and handle initial config
	logger.Info("Loading initial stream configuration")
	km.updateConfig(configHandler.AddListener(km.configChan))
}

func (km *KafkaManager) listenForConfigUpdates() {
	logger := km.logger.WithField("func", "listenForConfigUpdates")
	for config := range km.configChan {
		logger.Info("Received config update")
		config := config
		km.updateConfig(&config)
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
	if km.active {
		if _, ok := km.gateways[modelName]; ok {
			logger.Infof("kafka gateway already exists for model %s", modelName)
		} else {
			modelConfig := &KafkaModelConfig{
				ModelName:   modelName,
				InputTopic:  fmt.Sprintf("%s-in", modelName),
				OutputTopic: fmt.Sprintf("%s-out", modelName),
				ErrorTopic:  fmt.Sprintf("%s-err", modelName),
			}
			if streamSpec != nil {
				if streamSpec.InputTopic != "" {
					modelConfig.InputTopic = streamSpec.InputTopic
				}
				if streamSpec.OutputTopic != "" {
					modelConfig.OutputTopic = streamSpec.OutputTopic
				}
			}
			km.logger.Infof("Adding consumer to broker %s for model %s", km.broker, modelName)
			inferGateway, err := NewInferKafkaGateway(km.logger, DEFAULT_NWORKERS, km.broker, modelConfig, km.serverConfig)
			km.gateways[modelName] = inferGateway
			if err != nil {
				return err
			}
			go inferGateway.Serve()
		}
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
