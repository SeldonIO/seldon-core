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
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// InferenceArtifactSpec defines the desired state of InferenceArtifact
type InferenceArtifactSpec struct {
	Name string `json:"name,omitempty" protobuf:"bytes,1,opt,name=name"`
	// The artifact type sklearn, pytorch, alibi-explain, alibi-detect
	// +optional
	Type string `json:"type,omitempty" protobuf:"bytes,2,opt,name=type"`
	// List of extra requirements for this model to be loaded on a compatible server, e.g. sklearn19
	Requirements []string `json:"requirements,omitempty" protobuf:"bytes,3,opt,name=requirements"`
	// Storage URI for the model repository
	StorageURI string `json:"storageUri" protobuf:"bytes,4,name=storageUri"`
	// Resources needed for artifact, cpu, memory etc
	// +optional
	Resources v1.ResourceList `json:"resources,omitempty" protobuf:"bytes,5,rep,name=resources,casttype=ResourceList,castkey=ResourceName"`
	//
	Memory resource.Quantity `json:"memory" protobuf:"bytes,6,name=memory"`
	// Name of the InferenceServer to deploy this artifact
	Server *string `json:"server,omitempty" protobuf:"bytes,7,opt,name=server"`
	// Whether the artifact can be run on a shared server
	// +optional
	Shared bool `json:"shared,omitempty" protobuf:"bytes,8,opt,name=shared"`
	//TODO move to metadata sub struct
	//
	// swagger URI endpoint
	// +optional
	SwaggerURI *string `json:"swaggerURI,omitempty" protobuf:"bytes,9,opt,name=swaggerURI"`
	// model server setting URI endpoint
	// +optional
	modelSettingsURI *string `json:"modelSettingsURI,omitempy" protobuf:"bytes,10,opt,name=modelSettingsURI"`
	// Prediction schema URI endpoint
	// +optional
	predictionSchemaURI *string `json:"predictionSchemaURI,omitempy" protobuf:"bytes,11,opt,name=predictionSchemaURI"`
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
