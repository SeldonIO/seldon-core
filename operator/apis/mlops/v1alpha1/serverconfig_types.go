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

	"github.com/seldonio/seldon-core/operatorv2/pkg/constants"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ServerConfigSpec defines the desired state of ServerConfig
type ServerConfigSpec struct {
	// PodSpec
	PodSpec              v1.PodSpec              `json:"podSpec"`
	VolumeClaimTemplates []PersistentVolumeClaim `json:"volumeClaimTemplates"`
}

// We use our own type rather than v1.PersistentVolumeClaim as metadata inlined is not handled correctly by CRDs
type PersistentVolumeClaim struct {
	Name string                       `json:"name"`
	Spec v1.PersistentVolumeClaimSpec `json:"spec"`
}

// ServerConfigStatus defines the observed state of ServerConfig
type ServerConfigStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ServerConfig is the Schema for the serverconfigs API
type ServerConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ServerConfigSpec   `json:"spec,omitempty"`
	Status ServerConfigStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ServerConfigList contains a list of ServerConfig
type ServerConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ServerConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ServerConfig{}, &ServerConfigList{})
}

func GetServerConfigForServer(serverType ServerType, client client.Client) (*ServerConfig, error) {
	if serverType == "" {
		return nil, fmt.Errorf("ServerType not specified and is required")
	}
	sc := ServerConfig{}
	err := client.Get(context.TODO(), types.NamespacedName{Name: string(serverType), Namespace: constants.SeldonNamespace}, &sc)
	return &sc, err
}
