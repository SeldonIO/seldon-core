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

// SeldonRuntimeSpec defines the desired state of SeldonRuntime
type SeldonRuntimeSpec struct {
	SeldonConfig string          `json:"seldonConfig"`
	Overrides    []*OverrideSpec `json:"overrides,omitempty"`
}

type OverrideSpec struct {
	Name        string         `json:"name"`
	Disable     bool           `json:"disable,omitempty"`
	Replicas    *int32         `json:"replicas,omitempty"`
	ServiceType v1.ServiceType `json:"serviceType,omitempty"`
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
)

var _ apis.ConditionsAccessor = (*ServerStatus)(nil)

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
		seldonRuntimeConditionSet.Manage(ss).MarkUnknown(condition.Type, "", "")
	case condition.Status == v1.ConditionUnknown:
		seldonRuntimeConditionSet.Manage(ss).MarkUnknown(condition.Type, condition.Reason, condition.Message)
	case condition.Status == v1.ConditionTrue:
		seldonRuntimeConditionSet.Manage(ss).MarkTrueWithReason(condition.Type, condition.Reason, condition.Message)
	case condition.Status == v1.ConditionFalse:
		seldonRuntimeConditionSet.Manage(ss).MarkFalse(condition.Type, condition.Reason, condition.Message)
	}
}
