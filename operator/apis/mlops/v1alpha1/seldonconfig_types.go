/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package v1alpha1

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/seldonio/seldon-core/operator/v2/pkg/constants"
)

// +genclient

const (
	SchedulerName       = "seldon-scheduler"
	EnvoyName           = "seldon-envoy"
	DataflowEngineName  = "seldon-dataflow-engine"
	HodometerName       = "hodometer"
	ModelGatewayName    = "seldon-modelgateway"
	PipelineGatewayName = "seldon-pipelinegateway"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// SeldonConfigSpec defines the desired state of SeldonConfig
type SeldonConfigSpec struct {
	Components []*ComponentDefn    `json:"components,omitempty"`
	Config     SeldonConfiguration `json:"config,omitempty"`
}

type SeldonConfiguration struct {
	// Control scaling parameters for various components
	ScalingConfig ScalingConfig      `json:"scalingConfig,omitempty"`
	TracingConfig TracingConfig      `json:"tracingConfig,omitempty"`
	KafkaConfig   KafkaConfig        `json:"kafkaConfig,omitempty"`
	AgentConfig   AgentConfiguration `json:"agentConfig,omitempty"`
	ServiceConfig ServiceConfig      `json:"serviceConfig,omitempty"`
}

type ServiceConfig struct {
	GrpcServicePrefix string         `json:"grpcServicePrefix,omitempty"`
	ServiceType       v1.ServiceType `json:"serviceType,omitempty"`
}

type KafkaConfig struct {
	BootstrapServers      string                        `json:"bootstrap.servers,omitempty"`
	ConsumerGroupIdPrefix string                        `json:"consumerGroupIdPrefix,omitempty"`
	Debug                 string                        `json:"debug,omitempty"`
	Consumer              map[string]intstr.IntOrString `json:"consumer,omitempty"`
	Producer              map[string]intstr.IntOrString `json:"producer,omitempty"`
	Streams               map[string]intstr.IntOrString `json:"streams,omitempty"`
	Topics                map[string]intstr.IntOrString `json:"topics,omitempty"`
	TopicPrefix           string                        `json:"topicPrefix,omitempty"`
}

type AgentConfiguration struct {
	Rclone RcloneConfiguration `json:"rclone,omitempty" yaml:"rclone,omitempty"`
}

type RcloneConfiguration struct {
	ConfigSecrets []string `json:"config_secrets,omitempty" yaml:"config_secrets,omitempty"`
	Config        []string `json:"config,omitempty" yaml:"config,omitempty"`
}

type TracingConfig struct {
	Disable              bool   `json:"disable,omitempty"`
	OtelExporterEndpoint string `json:"otelExporterEndpoint,omitempty"`
	OtelExporterProtocol string `json:"otelExporterProtocol,omitempty"`
	Ratio                string `json:"ratio,omitempty"`
}

type ScalingConfig struct {
	Models  *ModelScalingConfig  `json:"models,omitempty"`
	Servers *ServerScalingConfig `json:"servers,omitempty"`
	// Scaling config impacting pipeline-gateway, dataflow-engine and model-gateway
	Pipelines *PipelineScalingConfig `json:"pipelines,omitempty"`
}

type ModelScalingConfig struct {
	Enable bool `json:"enabled,omitempty"`
}

type ServerScalingConfig struct {
	Enable                     bool  `json:"enabled,omitempty"`
	ScaleDownPackingEnabled    bool  `json:"scaleDownPackingEnabled,omitempty"`
	ScaleDownPackingPercentage int32 `json:"scaleDownPackingPercentage,omitempty"`
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
	MaxShardCountMultiplier int32 `json:"maxShardCountMultiplier,omitempty"`
}

type ComponentDefn struct {
	// +kubebuilder:validation:Required

	Name                 string                  `json:"name"`
	Labels               map[string]string       `json:"labels,omitempty"`
	Annotations          map[string]string       `json:"annotations,omitempty"`
	Replicas             *int32                  `json:"replicas,omitempty"`
	PodSpec              *v1.PodSpec             `json:"podSpec,omitempty"`
	VolumeClaimTemplates []PersistentVolumeClaim `json:"volumeClaimTemplates,omitempty"`
}

// SeldonConfigStatus defines the observed state of SeldonConfig
type SeldonConfigStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// SeldonConfig is the Schema for the seldonconfigs API
type SeldonConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SeldonConfigSpec   `json:"spec,omitempty"`
	Status SeldonConfigStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// SeldonConfigList contains a list of SeldonConfig
type SeldonConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SeldonConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SeldonConfig{}, &SeldonConfigList{})
}

func GetSeldonConfigForSeldonRuntime(seldonConfigName string, client client.Client) (*SeldonConfig, error) {
	if seldonConfigName == "" {
		return nil, fmt.Errorf("SeldonConfig not specified and is required")
	}
	sc := SeldonConfig{}
	ctx, cancel := context.WithTimeout(context.Background(), constants.K8sAPISingleCallTimeout)
	defer cancel()
	err := client.Get(ctx, types.NamespacedName{Name: seldonConfigName, Namespace: constants.SeldonNamespace}, &sc)
	return &sc, err
}

func (s *SeldonConfiguration) AddDefaults(defaults SeldonConfiguration) {
	s.TracingConfig.addDefaults(defaults.TracingConfig)
	s.KafkaConfig.addDefaults(defaults.KafkaConfig)
	s.AgentConfig.addDefaults(defaults.AgentConfig)
	s.ServiceConfig.addDefaults(defaults.ServiceConfig)
	s.ScalingConfig.addDefaults(defaults.ScalingConfig)
}

func (k *KafkaConfig) addDefaults(defaults KafkaConfig) {
	if k.BootstrapServers == "" {
		k.BootstrapServers = defaults.BootstrapServers
	}
	if k.ConsumerGroupIdPrefix == "" {
		k.ConsumerGroupIdPrefix = defaults.ConsumerGroupIdPrefix
	}
	if k.Consumer == nil && defaults.Consumer != nil {
		k.Consumer = make(map[string]intstr.IntOrString)
	}
	for key, val := range defaults.Consumer {
		if _, ok := k.Consumer[key]; !ok {
			k.Consumer[key] = val
		}
	}
	if k.Producer == nil && defaults.Producer != nil {
		k.Producer = make(map[string]intstr.IntOrString)
	}
	for key, val := range defaults.Producer {
		if _, ok := k.Producer[key]; !ok {
			k.Producer[key] = val
		}
	}
	if k.Streams == nil && defaults.Streams != nil {
		k.Streams = make(map[string]intstr.IntOrString)
	}
	if k.Topics == nil && defaults.Topics != nil {
		k.Topics = make(map[string]intstr.IntOrString)
		for key, val := range defaults.Topics {
			if _, ok := k.Topics[key]; !ok {
				k.Topics[key] = val
			}
		}
	}
	for key, val := range defaults.Streams {
		if _, ok := k.Streams[key]; !ok {
			k.Streams[key] = val
		}
	}
	if k.Debug == "" {
		k.Debug = defaults.Debug
	}
	if k.TopicPrefix == "" {
		k.TopicPrefix = defaults.TopicPrefix
	}
}

func (sc *ScalingConfig) addDefaults(defaults ScalingConfig) {
	if sc.Models == nil && defaults.Models != nil {
		sc.Models = defaults.Models
	}
	if sc.Servers == nil && defaults.Servers != nil {
		sc.Servers = defaults.Servers
	}
	if sc.Pipelines == nil && defaults.Pipelines != nil {
		sc.Pipelines = defaults.Pipelines
	}
}

func (a *AgentConfiguration) addDefaults(defaults AgentConfiguration) {
	a.Rclone.addDefaults(defaults.Rclone)
}

// Not presently checking for duplicates
func (r *RcloneConfiguration) addDefaults(defaults RcloneConfiguration) {
	r.Config = append(r.Config, defaults.Config...)
	r.ConfigSecrets = append(r.ConfigSecrets, defaults.ConfigSecrets...)
}

func (t *TracingConfig) addDefaults(defaults TracingConfig) {
	if t.Ratio == "" {
		t.Ratio = defaults.Ratio
	}
	if t.OtelExporterEndpoint == "" {
		t.OtelExporterEndpoint = defaults.OtelExporterEndpoint
	}
	if t.OtelExporterProtocol == "" {
		t.OtelExporterProtocol = defaults.OtelExporterProtocol
	}
}

func (sc *ServiceConfig) addDefaults(defaults ServiceConfig) {
	if sc.GrpcServicePrefix == "" {
		sc.GrpcServicePrefix = defaults.GrpcServicePrefix
	}
	if sc.ServiceType == "" {
		sc.ServiceType = defaults.ServiceType
	}
}
