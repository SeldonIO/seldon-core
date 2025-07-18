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

	"github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type ResourceType string

const (
	ModelResourceType    ResourceType = "model"
	PipelineResourceType ResourceType = "pipeline"
)

// ExperimentSpec defines the desired state of Experiment
type ExperimentSpec struct {
	Default      *string               `json:"default,omitempty"`
	Candidates   []ExperimentCandidate `json:"candidates"`
	Mirror       *ExperimentMirror     `json:"mirror,omitempty"`
	ResourceType ResourceType          `json:"resourceType,omitempty"`
}

type ExperimentCandidate struct {
	Name   string `json:"name"`
	Weight uint32 `json:"weight"`
}

type ExperimentMirror struct {
	Name    string `json:"name"`
	Percent uint32 `json:"percent"`
}

// ExperimentStatus defines the observed state of Experiment
type ExperimentStatus struct {
	duckv1.Status `json:",inline"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:shortName=mlx
//+kubebuilder:printcolumn:name="Experiment ready",type=string,JSONPath=`.status.conditions[?(@.type=='ExperimentReady')].status`,description="Experiment ready status"
//+kubebuilder:printcolumn:name="Candidates ready",type=string,JSONPath=`.status.conditions[?(@.type=='CandidatesReady')].status`,description="Candidates ready status",priority=1
//+kubebuilder:printcolumn:name="Mirror ready",type=string,JSONPath=`.status.conditions[?(@.type=='MirrorReady')].status`,description="Mirror ready status",priority=1
//+kubebuilder:printcolumn:name="Message",type=string,JSONPath=`.status.conditions[?(@.type=='Ready')].message`,description="Status message"
//+kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

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
		Name:   candidate.Name,
		Weight: candidate.Weight,
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
			Name:    e.Spec.Mirror.Name,
			Percent: e.Spec.Mirror.Percent,
		}
	}
	var resourceType scheduler.ResourceType
	switch e.Spec.ResourceType {
	case PipelineResourceType:
		resourceType = scheduler.ResourceType_PIPELINE
	default:
		resourceType = scheduler.ResourceType_MODEL
	}
	return &scheduler.Experiment{
		Name:       e.Name,
		Default:    e.Spec.Default,
		Candidates: candidates,
		Mirror:     mirror,
		KubernetesMeta: &scheduler.KubernetesMeta{
			Namespace:  e.Namespace,
			Generation: e.Generation,
		},
		ResourceType: resourceType,
	}
}

const (
	ExperimentReady apis.ConditionType = "ExperimentReady"
	CandidatesReady apis.ConditionType = "CandidatesReady"
	MirrorReady     apis.ConditionType = "MirrorReady"
)

var experimentConditionSet = apis.NewLivingConditionSet(
	ExperimentReady,
	CandidatesReady,
	MirrorReady,
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
		return
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
