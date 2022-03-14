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
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
)

const (
	AgentContainerName  = "agent"
	RcloneContainerName = "rclone"
	ServerContainerName = "server"
)

// ServerSpec defines the desired state of Server
// TODO do we need autoscaling spec?
type ServerSpec struct {
	// Server definition
	Server ServerDefn `json:"server,omitempty"`
	// Seldon mesh specifications for multi model servers that can load models dynamically
	MMS *MultiModelSpec `json:"mms,omitempty"`
	// PodSpec overrides
	// Slices such as containers would be appended not overridden
	PodSpec *PodSpec `json:"podSpec,omitempty"`
	// Scaling spec
	ScalingSpec `json:",inline"`
}

type MultiModelSpec struct {
	// The capabilities this server will advertise
	Capabilities []string `json:"capabilities,omitempty"`
	// The Init container overrides to download preset models
	Init *v1.Container `json:"init,omitempty"`
	// The Agent overrides
	Agent *v1.Container `json:"agent,omitempty"`
	// The RClone server overrides
	RClone *v1.Container `json:"rclone,omitempty"`
}

type ServerDefn struct {
	// Server config name to match
	// Required
	Config string `json:"config"`
}

// ServerStatus defines the observed state of Server
type ServerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	duckv1.Status `json:",inline"`
	// Number of loade models
	LoadedModelReplicas int32 `json:"loadedModels"`
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

func (s *Server) Default() {
	s.Spec.Default()
}

func (s *ServerSpec) Default() {
	s.ScalingSpec.Default()
}

const (
	StatefulSetReady apis.ConditionType = "StatefulSetReady"
)

var serverConditionSet = apis.NewLivingConditionSet(
	StatefulSetReady,
)

var _ apis.ConditionsAccessor = (*ServerStatus)(nil)

func (ss *ServerStatus) InitializeConditions() {
	serverConditionSet.Manage(ss).InitializeConditions()
}

func (ss *ServerStatus) IsReady() bool {
	return serverConditionSet.Manage(ss).IsHappy()
}

func (ss *ServerStatus) GetCondition(t apis.ConditionType) *apis.Condition {
	return serverConditionSet.Manage(ss).GetCondition(t)
}

func (ss *ServerStatus) IsConditionReady(t apis.ConditionType) bool {
	c := serverConditionSet.Manage(ss).GetCondition(t)
	return c != nil && c.Status == v1.ConditionTrue
}

func (ss *ServerStatus) SetCondition(condition *apis.Condition) {
	switch {
	case condition == nil:
		serverConditionSet.Manage(ss).MarkUnknown(condition.Type, "", "")
	case condition.Status == v1.ConditionUnknown:
		serverConditionSet.Manage(ss).MarkUnknown(condition.Type, condition.Reason, condition.Message)
	case condition.Status == v1.ConditionTrue:
		serverConditionSet.Manage(ss).MarkTrueWithReason(condition.Type, condition.Reason, condition.Message)
	case condition.Status == v1.ConditionFalse:
		serverConditionSet.Manage(ss).MarkFalse(condition.Type, condition.Reason, condition.Message)
	}
}
