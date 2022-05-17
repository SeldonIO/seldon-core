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
	"github.com/seldonio/seldon-core/operatorv2/scheduler/apis/mlops/scheduler"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ExperimentSpec defines the desired state of Experiment
type ExperimentSpec struct {
	DefaultModel *string               `json:"defaultModel,omitempty"`
	Candidates   []ExperimentCandidate `json:"candidates"`
	Mirror       *ExperimentMirror     `json:"mirror,omitempty"`
}

type ExperimentCandidate struct {
	ModelName string `json:"modelName"`
	Weight    uint32 `json:"weight"`
}

type ExperimentMirror struct {
	ModelName string `json:"model_name"`
	Percent   uint32 `json:"percent"`
}

// ExperimentStatus defines the observed state of Experiment
type ExperimentStatus struct {
	duckv1.Status `json:",inline"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:shortName=mlx

// Experiment is the Schema for the experiments API
type Experiment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ExperimentSpec   `json:"spec,omitempty"`
	Status ExperimentStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ExperimentList contains a list of Experiment
type ExperimentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Experiment `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Experiment{}, &ExperimentList{})
}

func asSchedulerCandidate(candidate *ExperimentCandidate) *scheduler.ExperimentCandidate {
	return &scheduler.ExperimentCandidate{
		ModelName: candidate.ModelName,
		Weight:    candidate.Weight,
	}
}

func (e *Experiment) AsSchedulerExperimentRequest() *scheduler.Experiment {
	var candidates []*scheduler.ExperimentCandidate
	var mirror *scheduler.ExperimentMirror
	for _, candidate := range e.Spec.Candidates {
		candidates = append(candidates, asSchedulerCandidate(&candidate))
	}
	if e.Spec.Mirror != nil {
		mirror = &scheduler.ExperimentMirror{
			ModelName: e.Spec.Mirror.ModelName,
			Percent:   e.Spec.Mirror.Percent,
		}
	}
	return &scheduler.Experiment{
		Name:         e.Name,
		DefaultModel: e.Spec.DefaultModel,
		Candidates:   candidates,
		Mirror:       mirror,
		KubernetesMeta: &scheduler.KubernetesMeta{
			Namespace:  e.Namespace,
			Generation: e.Generation,
		},
	}
}

const (
	ExperimentReady apis.ConditionType = "ExperimentReady"
)

var experimentConditionSet = apis.NewLivingConditionSet(
	ExperimentReady,
)

var _ apis.ConditionsAccessor = (*ExperimentStatus)(nil)

func (es *ExperimentStatus) InitializeConditions() {
	experimentConditionSet.Manage(es).InitializeConditions()
}

func (es *ExperimentStatus) IsReady() bool {
	return experimentConditionSet.Manage(es).IsHappy()
}

func (es *ExperimentStatus) GetCondition(t apis.ConditionType) *apis.Condition {
	return experimentConditionSet.Manage(es).GetCondition(t)
}

func (es *ExperimentStatus) IsConditionReady(t apis.ConditionType) bool {
	return experimentConditionSet.Manage(es).GetCondition(t) != nil && experimentConditionSet.Manage(es).GetCondition(t).Status == v1.ConditionTrue
}

func (es *ExperimentStatus) SetCondition(conditionType apis.ConditionType, condition *apis.Condition) {
	switch {
	case condition == nil:
		experimentConditionSet.Manage(es).MarkUnknown(conditionType, "", "")
	case condition.Status == v1.ConditionUnknown:
		experimentConditionSet.Manage(es).MarkUnknown(conditionType, condition.Reason, condition.Message)
	case condition.Status == v1.ConditionTrue:
		experimentConditionSet.Manage(es).MarkTrueWithReason(conditionType, condition.Reason, condition.Message)
	case condition.Status == v1.ConditionFalse:
		experimentConditionSet.Manage(es).MarkFalse(conditionType, condition.Reason, condition.Message)
	}
}

func (es *ExperimentStatus) CreateAndSetCondition(conditionType apis.ConditionType, isTrue bool, reason string) {
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
	es.SetCondition(conditionType, &condition)
}
