/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
)

// SeldonRuntimeSpec defines the desired state of SeldonRuntime
type SeldonRuntimeSpec struct {
	SeldonConfig string              `json:"seldonConfig"`
	Overrides    []*OverrideSpec     `json:"overrides,omitempty"`
	Config       SeldonConfiguration `json:"config,omitempty"`
	// +Optional
	// If set then when the referenced SeldonConfig changes we will NOT update the SeldonRuntime immediately.
	// Explicit changes to the SeldonRuntime itself will force a reconcile though
	DisableAutoUpdate bool `json:"disableAutoUpdate,omitempty"`
}

type OverrideSpec struct {
	Name        string         `json:"name"`
	Disable     bool           `json:"disable,omitempty"`
	Replicas    *int32         `json:"replicas,omitempty"`
	ServiceType v1.ServiceType `json:"serviceType,omitempty"`
	PodSpec     *PodSpec       `json:"podSpec,omitempty"`
}

// SeldonRuntimeStatus defines the observed state of SeldonRuntime
type SeldonRuntimeStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	duckv1.Status `json:",inline"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// SeldonRuntime is the Schema for the seldonruntimes API
type SeldonRuntime struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SeldonRuntimeSpec   `json:"spec,omitempty"`
	Status SeldonRuntimeStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// SeldonRuntimeList contains a list of SeldonRuntime
type SeldonRuntimeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SeldonRuntime `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SeldonRuntime{}, &SeldonRuntimeList{})
}

var ConditionNameMap = map[string]apis.ConditionType{
	SchedulerName:       SchedulerReady,
	EnvoyName:           EnvoyReady,
	DataflowEngineName:  DataflowEngineReady,
	ModelGatewayName:    ModelGatewayReady,
	PipelineGatewayName: PipelineGatewayReady,
	HodometerName:       HodometerReady,
}

const (
	SchedulerReady       apis.ConditionType = "SchedulerReady"
	DataflowEngineReady  apis.ConditionType = "DataflowEngineReady"
	ModelGatewayReady    apis.ConditionType = "ModelGatewayReady"
	PipelineGatewayReady apis.ConditionType = "PipelineGatewayReady"
	EnvoyReady           apis.ConditionType = "EnvoyReady"
	HodometerReady       apis.ConditionType = "HodometerReady"
)

var seldonRuntimeConditionSet = apis.NewLivingConditionSet(
	SchedulerReady,
	DataflowEngineReady,
	ModelGatewayReady,
	PipelineGatewayReady,
	EnvoyReady,
	HodometerReady,
)

var _ apis.ConditionsAccessor = (*SeldonRuntimeStatus)(nil)

func (ss *SeldonRuntimeStatus) InitializeConditions() {
	seldonRuntimeConditionSet.Manage(ss).InitializeConditions()
}

func (ss *SeldonRuntimeStatus) IsReady() bool {
	return seldonRuntimeConditionSet.Manage(ss).IsHappy()
}

func (ss *SeldonRuntimeStatus) GetCondition(t apis.ConditionType) *apis.Condition {
	return seldonRuntimeConditionSet.Manage(ss).GetCondition(t)
}

func (ss *SeldonRuntimeStatus) IsConditionReady(t apis.ConditionType) bool {
	c := seldonRuntimeConditionSet.Manage(ss).GetCondition(t)
	return c != nil && c.Status == v1.ConditionTrue
}

func (ss *SeldonRuntimeStatus) SetCondition(condition *apis.Condition) {
	switch {
	case condition == nil:
		return
	case condition.Status == v1.ConditionUnknown:
		seldonRuntimeConditionSet.Manage(ss).MarkUnknown(condition.Type, condition.Reason, condition.Message)
	case condition.Status == v1.ConditionTrue:
		seldonRuntimeConditionSet.Manage(ss).MarkTrueWithReason(condition.Type, condition.Reason, condition.Message)
	case condition.Status == v1.ConditionFalse:
		seldonRuntimeConditionSet.Manage(ss).MarkFalse(condition.Type, condition.Reason, condition.Message)
	}
}
