/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package v1alpha2

import (
	seldonv1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:noStatus
// +kubebuilder:object:root=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SeldonDeployment is the Schema for the seldondeployments API
// +k8s:openapi-gen=true
// +kubebuilder:resource:shortName=sdep
// +kubebuilder:subresource:status
// +kubebuilder:subresource:scale:specpath=.spec.replicas,statuspath=.status.replicas
type SeldonDeployment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   seldonv1.SeldonDeploymentSpec   `json:"spec,omitempty"`
	Status seldonv1.SeldonDeploymentStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SeldonDeploymentList contains a list of SeldonDeployment
type SeldonDeploymentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SeldonDeployment `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SeldonDeployment{}, &SeldonDeploymentList{})
}

// defaulting codes

func (r *SeldonDeployment) Default() {
	seldondeploymentLog.Info("Defaulting Seldon Deployment called", "name", r.Name)

	if r.ObjectMeta.Namespace == "" {
		r.ObjectMeta.Namespace = "default"
	}
	r.Spec.DefaultSeldonDeployment(r.Name, r.ObjectMeta.Namespace)
}
