/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package v1alpha1

import (
	appsv1 "k8s.io/api/apps/v1"
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
type ServerSpec struct {
	// Server definition
	ServerConfig string `json:"serverConfig"`
	// The extra capabilities this server will advertise
	// These are added to the capabilities exposed by the referenced ServerConfig
	ExtraCapabilities []string `json:"extraCapabilities,omitempty"`
	// The capabilities this server will advertise
	// This will override any from the referenced ServerConfig
	Capabilities []string `json:"capabilities,omitempty"`
	// Image overrides
	ImageOverrides *ContainerOverrideSpec `json:"imageOverrides,omitempty"`
	// PodSpec overrides
	// Slices such as containers would be appended not overridden
	PodSpec *PodSpec `json:"podSpec,omitempty"`
	// StatefulSetPersistentVolumeClaimRetentionPolicy policy for stateful set pvc
	StatefulSetPersistentVolumeClaimRetentionPolicy *appsv1.StatefulSetPersistentVolumeClaimRetentionPolicy `json:"statefulSetPersistentVolumeClaimRetentionPolicy,omitempty"`
	// Scaling spec
	ScalingSpec `json:",inline"`
	// +Optional
	// If set then when the referenced ServerConfig changes we will NOT update the Server immediately.
	// Explicit changes to the Server itself will force a reconcile though
	DisableAutoUpdate bool `json:"disableAutoUpdate,omitempty"`
}

type ContainerOverrideSpec struct {
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
	// Number of loaded models
	LoadedModelReplicas int32  `json:"loadedModels"`
	Replicas            int32  `json:"replicas"`
	Selector            string `json:"selector"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:subresource:scale:specpath=.spec.replicas,statuspath=.status.replicas,selectorpath=.status.selector
//+kubebuilder:resource:shortName=mls
//+kubebuilder:printcolumn:name="Ready",type=string,JSONPath=`.status.conditions[?(@.type=="Ready")].status`,description="Server ready status"
//+kubebuilder:printcolumn:name="Replicas",type=integer,JSONPath=`.status.replicas`,description="Number of replicas"
//+kubebuilder:printcolumn:name="Loaded Models",type=integer,JSONPath=`.status.loadedModels`,description="Number of loaded models"
//+kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

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
		return
	case condition.Status == v1.ConditionUnknown:
		serverConditionSet.Manage(ss).MarkUnknown(condition.Type, condition.Reason, condition.Message)
	case condition.Status == v1.ConditionTrue:
		serverConditionSet.Manage(ss).MarkTrueWithReason(condition.Type, condition.Reason, condition.Message)
	case condition.Status == v1.ConditionFalse:
		serverConditionSet.Manage(ss).MarkFalse(condition.Type, condition.Reason, condition.Message)
	}
}

func (ss *ServerStatus) CreateAndSetCondition(conditionType apis.ConditionType, isTrue bool, reason string) {
	condition := apis.Condition{}
	if isTrue {
		condition.Status = v1.ConditionTrue
	} else {
		condition.Status = v1.ConditionFalse
	}
	condition.Type = conditionType
	condition.Reason = reason
	condition.LastTransitionTime = apis.VolatileTime{
		Inner: metav1.Now(),
	}
	ss.SetCondition(&condition)
}
