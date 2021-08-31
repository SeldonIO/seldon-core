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
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// InferenceExplainerSpec defines the desired state of InferenceExplainer
type InferenceExplainerSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	StorageURI string `json:"storageUri,omitempty" protobuf:"bytes,1,opt,name=storageUri"`
	// Resources needed for artifact, cpu, memory etc
	// +optional
	Resources v1.ResourceList `json:"resources,omitempty" protobuf:"bytes,2,rep,name=resources,casttype=ResourceList,castkey=ResourceName"`
	// Name of the InferenceServer to deploy this artifact
	Server string `json:"server,omitempty" protobuf:"bytes,3,opt,name=server"`
	// Whether the artifact can be run on a shared server
	// +optional
	Shared bool `json:"shared,omitempty" protobuf:"bytes,4,opt,name=shared"`
	// Reference to model InferenceArtifact for Black Box models
	// +optional
	ModelRef string `json:"modelRef,omitempty" protobuf:"bytes,5,opt,name=modelRef"`
}

// InferenceExplainerStatus defines the observed state of InferenceExplainer
type InferenceExplainerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// InferenceExplainer is the Schema for the inferenceexplainers API
type InferenceExplainer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   InferenceExplainerSpec   `json:"spec,omitempty"`
	Status InferenceExplainerStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// InferenceExplainerList contains a list of InferenceExplainer
type InferenceExplainerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []InferenceExplainer `json:"items"`
}

func init() {
	SchemeBuilder.Register(&InferenceExplainer{}, &InferenceExplainerList{})
}
