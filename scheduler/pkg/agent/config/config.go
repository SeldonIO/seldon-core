/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package config

import (
	"time"

	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/config"
)

const (
	AgentConfigYamlFilename      = "agent.yaml"
	ConfigMapName                = "seldon-agent"
	ServiceReadyRetryMaxInterval = 30 * time.Second
)

var (
	DefaultAgentConfiguration = AgentConfiguration{
		Rclone: &RcloneConfiguration{
			ConfigSecrets: []string{},
			Config:        []string{},
		},
		Kafka: &KafkaConfiguration{
			Active: false,
		},
	}
)

type AgentConfiguration struct {
	Rclone *RcloneConfiguration `json:"rclone,omitempty" yaml:"rclone,omitempty"`
	Kafka  *KafkaConfiguration  `json:"kafka,omitempty" yaml:"kafka,omitempty"`
}

type RcloneConfiguration struct {
	ConfigSecrets []string `json:"config_secrets,omitempty" yaml:"config_secrets,omitempty"`
	Config        []string `json:"config,omitempty" yaml:"config,omitempty"`
}

type KafkaConfiguration struct {
	Active bool   `json:"active,omitempty" yaml:"active,omitempty"`
	Broker string `json:"broker,omitempty" yaml:"broker,omitempty"`
}

type AgentConfigHandler = config.ConfigWatcher[AgentConfiguration, *AgentConfiguration]

func (ac *AgentConfiguration) DeepCopy() AgentConfiguration {
	var rcloneCopy *RcloneConfiguration
	var kafkaCopy *KafkaConfiguration

	if ac.Rclone != nil {
		// Maintain nil slices if settings are not present.
		// This is important because json.Marshal treats nil and empty slices differently.
		var cs []string
		if len(ac.Rclone.ConfigSecrets) > 0 {
			cs = make([]string, len(ac.Rclone.ConfigSecrets))
			copy(cs, ac.Rclone.ConfigSecrets)
		}
		var cfg []string
		if len(ac.Rclone.Config) > 0 {
			cfg = make([]string, len(ac.Rclone.Config))
			copy(cfg, ac.Rclone.Config)
		}
		rcloneDeepCopy := RcloneConfiguration{
			ConfigSecrets: cs,
			Config:        cfg,
		}
		rcloneCopy = &rcloneDeepCopy
	} else {
		rcloneCopy = nil
	}

	if ac.Kafka != nil {
		kafkaDeepCopy := *ac.Kafka
		kafkaCopy = &kafkaDeepCopy
	} else {
		kafkaCopy = nil
	}

	return AgentConfiguration{
		Rclone: rcloneCopy,
		Kafka:  kafkaCopy,
	}
}

func (ac *AgentConfiguration) Default() AgentConfiguration {
	return DefaultAgentConfiguration.DeepCopy()
}

func NewAgentConfigHandler(configPath string, namespace string, logger log.FieldLogger, clientset kubernetes.Interface) (*AgentConfigHandler, error) {
	return config.NewConfigWatcher(
		configPath,
		AgentConfigYamlFilename,
		namespace,
		false, // watch mounted config file rather than using k8s informer on the config map
		ConfigMapName,
		clientset,
		onConfigUpdate,
		logger.WithField("source", "AgentConfigHandler"),
	)
}

func onConfigUpdate(config *AgentConfiguration, logger log.FieldLogger) error {
	if config.Rclone != nil {
		logger.Infof("Rclone Config loaded %v", config.Rclone)
	}
	return nil
}
