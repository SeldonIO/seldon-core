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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ExplainerSpec defines the desired state of Explainer
type ExplainerSpec struct {
	ModelSpec `json:",inline"`
	BlackBox  *BlackBox `json:"blackbox,omitempty"`
	WhiteBox  *WhiteBox `json:"whitebox,omitempty"`
}

// Either ModelRef or PipelineRef is required
type BlackBox struct {
	// Reference to Model for Black Box models
	// +optional
	ModelRef *string `json:"modelRef,omitempty"`
	// Reference to Pipeline for Black Box models
	PipelineRef *string `json:"pipelineRef,omitempty"`
	// Reference to Pipeline step - must include pipeline
	PipelineStep *string `json:"pipelineStep,omitempty"`
}

type WhiteBox struct {
	// Storage URI for the model repository
	// +optional
	InferenceArtifactSpec `json:",inline"`
}

// ExplainerStatus defines the observed state of Explainer
type ExplainerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:shortName=mle

// Explainer is the Schema for the explainers API
type Explainer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ExplainerSpec   `json:"spec,omitempty"`
	Status ExplainerStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ExplainerList contains a list of Explainer
type ExplainerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Explainer `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Explainer{}, &ExplainerList{})
}
