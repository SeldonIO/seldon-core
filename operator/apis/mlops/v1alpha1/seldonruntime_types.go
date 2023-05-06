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
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ProtocolType string

const (
	SecurityProtocolPlaintxt ProtocolType = "PLAINTXT"
	SecurityProtocolSSL      ProtocolType = "SSL"
	SecurityProtocolSASLSSL  ProtocolType = "SASL_SSL"
)

// SeldonRuntimeSpec defines the desired state of SeldonRuntime
type SeldonRuntimeSpec struct {
	Scheduler       *SchedulerSpec       `json:"scheduler,omitempty"`
	Envoy           *EnvoySpec           `json:"envoy,omitempty"`
	DataflowEngine  *DataFlowSpec        `json:"dataflowEngine,omitempty"`
	ModelGateway    *ModelGatewaySpec    `json:"modelGateway,omitempty"`
	PipelineGateway *PipelineGatewaySpec `json:"pipelineGateway,omitempty"`
	Hodometer       *HodometerSpec       `json:"hodometer,omitempty"`
	Security        SecuritySettings     `json:"security,omitempty"`
}

type DeploymentSpec struct {
	Replicas         *int32                    `json:"replicas,omitempty"`
	ImagePullSecrets []v1.LocalObjectReference `json:"imagePullSecrets,omitempty"`
}

type SchedulerSpec struct {
	Container      *v1.Container   `json:"container,omitempty"`
	DeploymentSpec *DeploymentSpec `json:"deploymentSpec,omitempty"`
}

type EnvoySpec struct {
	Container      *v1.Container   `json:"container,omitempty"`
	DeploymentSpec *DeploymentSpec `json:"deploymentSpec,omitempty"`
}

type DataFlowSpec struct {
	Container      *v1.Container   `json:"container,omitempty"`
	DeploymentSpec *DeploymentSpec `json:"deploymentSpec,omitempty"`
}

type ModelGatewaySpec struct {
	Container      *v1.Container   `json:"container,omitempty"`
	DeploymentSpec *DeploymentSpec `json:"deploymentSpec,omitempty"`
}

type PipelineGatewaySpec struct {
	Container      *v1.Container   `json:"container,omitempty"`
	DeploymentSpec *DeploymentSpec `json:"deploymentSpec,omitempty"`
}

type HodometerSpec struct {
	Active         bool            `json:"active,omitempty"`
	Container      *v1.Container   `json:"container,omitempty"`
	DeploymentSpec *DeploymentSpec `json:"deploymentSpec,omitempty"`
}

type SecuritySettings struct {
	Controlplane ControlAuth `json:"controlplane,omitempty"`
	Envoy        EnvoyAuth   `json:"envoy,omitempty"`
	Kafka        KafkaAuth   `json:"kafka,omitempty"`
}

type ControlAuth struct {
	// +kubebuilder:default=PLAINTEXT
	Protocol ProtocolType `json:"protocol,omitempty"`
	SSL      SSLAuth      `json:"ssl,omitempty"`
}

type KafkaAuth struct {
	// +kubebuilder:default=PLAINTEXT
	Protocol ProtocolType `json:"protocol,omitempty"`
	SASL     SASLAuth     `json:"sasl,omitempty"`
	SSL      KafkaSSLAuth `json:"ssl,omitempty"`
}

type EnvoyAuth struct {
	// +kubebuilder:default=PLAINTEXT
	Protocol ProtocolType `json:"protocol,omitempty"`
	SSL      EnvoySSLAuth `json:"ssl,omitempty"`
}

type EnvoySSLAuth struct {
	Upstream   EnvoyUpstreamSSLAuth   `json:"upstream,omitempty"`
	Downstream EnvoyDownstreamSSLAuth `json:"downstream,omitempty"`
}

type EnvoyUpstreamSSLAuth struct {
	Server ServerSSLAuth `json:"server,omitempty"`
	Client ClientSSLAuth `json:"client,omitempty"`
}

type EnvoyDownstreamSSLAuth struct {
	Server ServerSSLAuth `json:"server,omitempty"`
	Client ClientSSLAuth `json:"client,omitempty"`
}

type KafkaSSLAuth struct {
	Secret                          string `json:"secret,omitempty"`
	BrokerValidationSecret          string `json:"brokerValidationSecret,omitempty"`
	KeyPath                         string `json:"keyPath,omitempty"`
	CrtPath                         string `json:"crtPath,omitempty"`
	CaPath                          string `json:"caPath,omitempty"`
	BrokerCaPath                    string `json:"brokerCaPath,omitempty"`
	EndpointIdentificationAlgorithm string `json:"endpointIdentificationAlgorithm,omitempty"`
}

type SASLAuth struct {
	Client *SASLClientAuth `json:"client,omitempty"`
}

type SASLClientAuth struct {
	Username     string `json:"username,omitempty"`
	Secret       string `json:"secret,omitempty"`
	PasswordPath string `json:"passwordPath,omitempty"`
}

type SSLAuth struct {
	Server ServerSSLAuth `json:"server,omitempty"`
	Client ClientSSLAuth `json:"client,omitempty"`
}

type ClientSSLAuth struct {
	Secret                 string `json:"secret,omitempty"`
	ServerValidationSecret string `json:"serverValidationSecret,omitempty"`
	KeyPath                string `json:"keyPath,omitempty"`
	CrtPath                string `json:"crtPath,omitempty"`
	CaPath                 string `json:"caPath,omitempty"`
	ServerCaPath           string `json:"serverCaPath,omitempty"`
	MTLS                   bool   `json:"MTLS,omitempty"`
}

type ServerSSLAuth struct {
	Secret                 string `json:"secret,omitempty"`
	ClientValidationSecret string `json:"clientValidationSecret,omitempty"`
	KeyPath                string `json:"keyPath,omitempty"`
	CrtPath                string `json:"crtPath,omitempty"`
	CaPath                 string `json:"caPath,omitempty"`
	ClientCaPath           string `json:"clientCaPath,omitempty"`
}

// SeldonRuntimeStatus defines the observed state of SeldonRuntime
type SeldonRuntimeStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// SeldonRuntime is the Schema for the seldonruntimes API
type SeldonRuntime struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SeldonRuntimeSpec   `json:"spec,omitempty"`
	Status SeldonRuntimeStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// SeldonRuntimeList contains a list of SeldonRuntime
type SeldonRuntimeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SeldonRuntime `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SeldonRuntime{}, &SeldonRuntimeList{})
}
