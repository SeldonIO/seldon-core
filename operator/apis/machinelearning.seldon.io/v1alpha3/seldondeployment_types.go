/*
Copyright 2019 The Seldon Authors.

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

package v1alpha3

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
	seldondeploymentlog.Info("Defaulting Seldon Deployment called", "name", r.Name)

	if r.ObjectMeta.Namespace == "" {
		r.ObjectMeta.Namespace = "default"
	}
	r.Spec.DefaultSeldonDeployment(r.Name, r.ObjectMeta.Namespace)
}
