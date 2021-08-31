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

// InferenceArtifactSpec defines the desired state of InferenceArtifact
type InferenceArtifactSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// The artifact type sklearn, pytorch, alibi-explain, alibi-detect
	// +optional
	Type string `json:"type,omitempty" protobuf:"bytes,1,opt,name=type"`
	// Storage URI for the model repository
	StorageURI string `json:"storageUri" protobuf:"bytes,2,name=storageUri"`
	// Resources needed for artifact, cpu, memory etc
	// +optional
	Resources v1.ResourceList `json:"resources,omitempty" protobuf:"bytes,3,rep,name=resources,casttype=ResourceList,castkey=ResourceName"`
	// Name of the InferenceServer to deploy this artifact
	Server string `json:"server,omitempty" protobuf:"bytes,4,opt,name=server"`
	// Whether the artifact can be run on a shared server
	// +optional
	Shared bool `json:"shared,omitempty" protobuf:"bytes,5,opt,name=shared"`
}

// InferenceArtifactStatus defines the observed state of InferenceArtifact
type InferenceArtifactStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// InferenceArtifact is the Schema for the inferenceartifacts API
type InferenceArtifact struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   InferenceArtifactSpec   `json:"spec,omitempty"`
	Status InferenceArtifactStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// InferenceArtifactList contains a list of InferenceArtifact
type InferenceArtifactList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []InferenceArtifact `json:"items"`
}

func init() {
	SchemeBuilder.Register(&InferenceArtifact{}, &InferenceArtifactList{})
}
