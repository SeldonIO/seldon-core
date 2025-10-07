/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package config

import (
	log "github.com/sirupsen/logrus"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/config"
)

const (
	ScalingConfigYamlFilename = "scaling.yaml"
	ConfigMapName             = "seldon-scaling"
)

var (
	DefaultPipelineScalingConfig = PipelineScalingConfig{
		MaxShardCountMultiplier: 1,
		isDefault:               true,
	}
	DefaultModelScalingConfig = ModelScalingConfig{
		Enable:    false,
		isDefault: true,
	}
	DefaultServerScalingConfig = ServerScalingConfig{
		Enable:                     true,
		ScaleDownPackingEnabled:    false,
		ScaleDownPackingPercentage: 0,
		isDefault:                  true,
	}
	DefaultScalingConfig = ScalingConfig{
		Models:    &DefaultModelScalingConfig,
		Servers:   &DefaultServerScalingConfig,
		Pipelines: &DefaultPipelineScalingConfig,
		isDefault: true,
	}
)

type ScalingConfig struct {
	Models  *ModelScalingConfig  `json:"models,omitempty" yaml:"models,omitempty"`
	Servers *ServerScalingConfig `json:"servers,omitempty" yaml:"servers,omitempty"`
	// Scaling config impacting pipeline-gateway, dataflow-engine and model-gateway
	Pipelines *PipelineScalingConfig `json:"pipelines,omitempty" yaml:"pipelines,omitempty"`
	isDefault bool
}

type ModelScalingConfig struct {
	Enable    bool `json:"enabled,omitempty" yaml:"enabled,omitempty"`
	isDefault bool
}

type ServerScalingConfig struct {
	Enable                     bool `json:"enabled,omitempty" yaml:"enabled,omitempty"`
	ScaleDownPackingEnabled    bool `json:"scaleDownPackingEnabled,omitempty" yaml:"scaleDownPackingEnabled,omitempty"`
	ScaleDownPackingPercentage int  `json:"scaleDownPackingPercentage,omitempty" yaml:"scaleDownPackingPercentage,omitempty"`
	isDefault                  bool
}

type PipelineScalingConfig struct {
	// MaxShardCountMultiplier influences the way the inferencing workload is sharded over the
	// replicas of pipeline components.
	//
	// - For each of pipeline-gateway and dataflow-engine, the max number of replicas is
	//   `maxShardCountMultiplier * number of pipelines`
	// - For model-gateway, the max number of replicas is
	//   `maxShardCountMultiplier * number of consumers`
	//
	// It doesn't make sense to set this to a value larger than the number of partitions for kafka
	// topics used in the Core 2 install.
	MaxShardCountMultiplier int `json:"maxShardCountMultiplier,omitempty" yaml:"maxShardCountMultiplier,omitempty"`
	isDefault               bool
}

func (sc *ScalingConfig) DeepCopy() ScalingConfig {
	var modelsCopy *ModelScalingConfig
	var serversCopy *ServerScalingConfig
	var pipelinesCopy *PipelineScalingConfig

	var modelsDeepCopy ModelScalingConfig
	if sc.Models != nil {
		modelsDeepCopy = *sc.Models
	} else {
		modelsDeepCopy = ModelScalingConfig(DefaultModelScalingConfig)
	}
	modelsCopy = &modelsDeepCopy

	var serversDeepCopy ServerScalingConfig
	if sc.Servers != nil {
		serversDeepCopy = *sc.Servers
	} else {
		serversDeepCopy = ServerScalingConfig(DefaultServerScalingConfig)
	}
	serversCopy = &serversDeepCopy

	var pipelinesDeepCopy PipelineScalingConfig
	if sc.Pipelines != nil {
		pipelinesDeepCopy = *sc.Pipelines
	} else {
		pipelinesDeepCopy = PipelineScalingConfig(DefaultPipelineScalingConfig)
	}
	pipelinesCopy = &pipelinesDeepCopy

	res := ScalingConfig{
		Models:    modelsCopy,
		Servers:   serversCopy,
		Pipelines: pipelinesCopy,
		isDefault: sc.isDefault,
	}
	return res
}

func (sc *ScalingConfig) Default() ScalingConfig {
	return DefaultScalingConfig.DeepCopy()
}

func (sc *ScalingConfig) IsDefault() bool {
	return sc.isDefault ||
		(sc.Models.IsDefault() && sc.Servers.IsDefault() && sc.Pipelines.IsDefault())
}

func (msc *ModelScalingConfig) IsDefault() bool {
	return msc.isDefault
}

func (ssc *ServerScalingConfig) IsDefault() bool {
	return ssc.isDefault
}

func (psc *PipelineScalingConfig) IsDefault() bool {
	return psc.isDefault
}

func LogWhenUsingDefaultScalingConfig(scalingConfig *ScalingConfig, logger log.FieldLogger) {
	if scalingConfig.IsDefault() {
		logger.Infof("Using default scaling config")
	} else {
		if scalingConfig.Models != nil && scalingConfig.Models.IsDefault() {
			logger.Infof("Using default model scaling config")
		}
		if scalingConfig.Servers != nil && scalingConfig.Servers.IsDefault() {
			logger.Infof("Using default server scaling config")
		}
		if scalingConfig.Pipelines != nil && scalingConfig.Pipelines.IsDefault() {
			logger.Infof("Using default pipeline scaling config")
		}
	}
}

type ScalingConfigHandler = config.ConfigWatcher[ScalingConfig, *ScalingConfig]

func NewScalingConfigHandler(configPath string, namespace string, logger log.FieldLogger) (*ScalingConfigHandler, error) {
	return config.NewConfigWatcher(
		configPath,
		ScalingConfigYamlFilename,
		namespace,
		false, // watch mounted config file rather than using k8s informer on the config map
		ConfigMapName,
		nil,
		onConfigUpdate,
		logger.WithField("source", "ScalingConfigHandler"),
	)
}

func onConfigUpdate(config *ScalingConfig, logger log.FieldLogger) error {
	// Any missing top-level config sections (Models, Server, Pipelines) are set to their defaults.
	// However, setting an empty section is treated differently, with all the fields being
	// considered explicitly set to the zero value of their datatype.
	//
	// We also ensure minimal validation of values, so that (for example) when the zero value of
	// the type is not valid, we set it to the default value.
	if config.Pipelines == nil {
		config.Pipelines = &DefaultPipelineScalingConfig
	} else {
		if config.Pipelines.MaxShardCountMultiplier <= 0 {
			config.Pipelines.MaxShardCountMultiplier = DefaultPipelineScalingConfig.MaxShardCountMultiplier
		}
	}
	if config.Models == nil {
		config.Models = &DefaultModelScalingConfig
	}
	if config.Servers == nil {
		config.Servers = &DefaultServerScalingConfig
	}
	return nil
}
