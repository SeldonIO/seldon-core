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
	"github.com/seldonio/seldon-core/operator/v2/pkg/constants"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
	Components    []*ComponentDefn   `json:"components,omitempty"`
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
	BootstrapServers string                        `json:"bootstrap.servers,omitempty"`
	Debug            string                        `json:"debug,omitempty"`
	Consumer         map[string]intstr.IntOrString `json:"consumer,omitempty"`
	Producer         map[string]intstr.IntOrString `json:"producer,omitempty"`
	Streams          map[string]intstr.IntOrString `json:"streams,omitempty"`
	TopicPrefix      string                        `json:"topicPrefix,omitempty"`
}

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

type TracingConfig struct {
	Enable               bool   `json:"enable,omitempty"`
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
