/*
Copyright 2021.

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

const (
	SchedulerName       = "seldon-scheduler"
	EnvoyName           = "seldon-envoy"
	DataflowEngineName  = "dataflow-engine"
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
	Ratio                string `json:"ratio,omitempty"`
}

type ComponentDefn struct {
	// +kubebuilder:validation:Required
	Name                 string                  `json:"name"`
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
	err := client.Get(context.TODO(), types.NamespacedName{Name: seldonConfigName, Namespace: constants.SeldonNamespace}, &sc)
	return &sc, err
}

func (s *SeldonConfiguration) AddDefaults(defaults SeldonConfiguration) {
	s.TracingConfig.addDefaults(defaults.TracingConfig)
	s.KafkaConfig.addDefaults(defaults.KafkaConfig)
	s.AgentConfig.addDefaults(defaults.AgentConfig)
	s.ServiceConfig.addDefaults(defaults.ServiceConfig)
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
}

func (sc *ServiceConfig) addDefaults(defaults ServiceConfig) {
	if sc.GrpcServicePrefix == "" {
		sc.GrpcServicePrefix = defaults.GrpcServicePrefix
	}
	if sc.ServiceType == "" {
		sc.ServiceType = defaults.ServiceType
	}
}
