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
	"k8s.io/apimachinery/pkg/api/resource"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// InferenceArtifactSpec defines the desired state of InferenceArtifact
type InferenceArtifactSpec struct {
	// Name of inference artifact
	Name string `json:"name,omitempty" protobuf:"bytes,1,name=name"`
	// List of extra requirements for this model to be loaded on a compatible server,
	// e.g. sklearn, sklearn18, pytorch, tensorflow, mlflow
	Requirements []string `json:"requirements,omitempty" protobuf:"bytes,2,opt,name=requirements"`
	// Storage URI for the model repository
	StorageURI string `json:"storageUri" protobuf:"bytes,3,name=storageUri"`
	// Memory needed for artifact
	// +optional
	Memory resource.Quantity `json:"memory,omitempty" protobuf:"bytes,4,name=memory"`
	// Name of the InferenceServer to deploy this artifact
	// +optional
	Server *string `json:"server,omitempty" protobuf:"bytes,5,opt,name=server"`
	// Whether the artifact can be run on a shared server
	// +optional
	Shared bool `json:"shared,omitempty" protobuf:"bytes,6,opt,name=shared"`
	// Metadata
	// +optional
	Metadata *ArtifactMetadataSpec `json:"metadata,omitempty" protobuf:"bytes,7,opt,name=metadata"`
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
