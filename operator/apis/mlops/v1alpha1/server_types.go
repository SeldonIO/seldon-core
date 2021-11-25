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

// ServerSpec defines the desired state of Server
// TODO do we need autoscaling spec?
type ServerSpec struct {
	// Server definition
	Server ServerDefn `json:"server,omitempty"`
	// Either Models or Mesh needs to be provides
	// Preloaded models for non mesh server
	Models []PreLoadedModelSpec `json:"models,omitempty"`
	// Seldon mesh specifications for mesh servers that can load models dynamically
	Mesh *MeshDefn `json:"mesh"`
	// PodSpec overrides
	PodOverride PodSpec `json:"podSpec,omitempty"`
	// Number of replicas - defaults to 1
	Replicas *int32 `json:"replicas,omitempty"`
}

type PreLoadedModelSpec struct {
	// Name override
	// +optional
	Name                  *string `json:"name,omitempty"`
	InferenceArtifactSpec `json:",inline"`
}

type MeshDefn struct {
	// The capabilities this server will advertise in the mesh
	Capabilities []string `json:"capabilities,omitempty"`
	// How much memory to push to disk to allow overcommited models
	SwapMemoryBytes bool `json:"swapMemoryBytes,omitempty"`
	// The Init container overrides to download preset models
	Init *v1.Container `json:"init,omitempty"`
	// The Agent overrides
	Agent *v1.Container `json:"agent,omitempty"`
	// The RClone server overrides
	RClone *v1.Container `json:"rclone,omitempty"`
}

type ServerDefn struct {
	// Server type - mlserver, triton or left out if custom container
	Type *string `json:"type,omitempty"`
	// +optional
	RuntimeVersion *string `json:"runtimeVersion,omitempty"`
	// Container overrides for server
	Container *v1.Container `json:"container,omitempty"`
}

// ServerStatus defines the observed state of Server
type ServerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Server is the Schema for the servers API
type Server struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ServerSpec   `json:"spec,omitempty"`
	Status ServerStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ServerList contains a list of Server
type ServerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Server `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Server{}, &ServerList{})
}
