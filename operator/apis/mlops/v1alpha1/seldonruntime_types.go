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
}

type SchedulerSpec struct {
	Replicas  *int32                  `json:"replicas,omitempty"`
	Resources v1.ResourceRequirements `json:"resources,omitempty"`
}

type EnvoySpec struct {
	Replicas  *int32                  `json:"replicas,omitempty"`
	Resources v1.ResourceRequirements `json:"resources,omitempty"`
}

type DataFlowSpec struct {
	Replicas  *int32                  `json:"replicas,omitempty"`
	Resources v1.ResourceRequirements `json:"resources,omitempty"`
}

type ModelGatewaySpec struct {
	Replicas  *int32                  `json:"replicas,omitempty"`
	Resources v1.ResourceRequirements `json:"resources,omitempty"`
}

type PipelineGatewaySpec struct {
	Replicas  *int32                  `json:"replicas,omitempty"`
	Resources v1.ResourceRequirements `json:"resources,omitempty"`
}

type HodometerSpec struct {
	Active    bool                    `json:"active,omitempty"`
	Replicas  *int32                  `json:"replicas,omitempty"`
	Resources v1.ResourceRequirements `json:"resources,omitempty"`
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
