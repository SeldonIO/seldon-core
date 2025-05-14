/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package v1alpha1

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/seldonio/seldon-core/operator/v2/pkg/constants"
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
//+kubebuilder:resource:shortName=mlc

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

func GetServerConfigForServer(serverConfig string, client client.Client) (*ServerConfig, error) {
	if serverConfig == "" {
		return nil, fmt.Errorf("ServerType not specified and is required")
	}
	sc := ServerConfig{}
	err := client.Get(context.TODO(), types.NamespacedName{Name: serverConfig, Namespace: constants.SeldonNamespace}, &sc)
	return &sc, err
}
