/*
Copyright 2021 The Seldon Authors.

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// InferenceServerInstanceSpec defines the desired state of InferenceServerInstance
type InferenceServerInstanceSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Replicas is an indication of load on the instance and will be used to rebalance models onto other servers
	Replicas int `json:"replicas,omitempty"`
}

// InferenceServerInstanceStatus defines the observed state of InferenceServerInstance
type InferenceServerInstanceStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Replicas int    `json:"replicas,omitempty"`
	Selector string `json:"selector,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
// +kubebuilder:subresource:scale:specpath=.spec.replicas,statuspath=.status.replicas,selectorpath=.status.selector
// +kubebuilder:resource:shortName=sin

// InferenceServerInstance is the Schema for the inferenceserverinstances API
type InferenceServerInstance struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   InferenceServerInstanceSpec   `json:"spec,omitempty"`
	Status InferenceServerInstanceStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// InferenceServerInstanceList contains a list of InferenceServerInstance
type InferenceServerInstanceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []InferenceServerInstance `json:"items"`
}

func init() {
	SchemeBuilder.Register(&InferenceServerInstance{}, &InferenceServerInstanceList{})
}
