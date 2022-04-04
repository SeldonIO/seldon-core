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

// PipelineSpec defines the desired state of Pipeline
type PipelineSpec struct {
	// The steps of this inference graph pipeline
	Steps  []PipelineStep  `json:"steps"`
	Output *PipelineOutput `json:"output,omitempty"`
}

type PipelineStep struct {
	// Name of the step
	Name string `json:"name"`
	// Previous step to receive data from
	Inputs []string `json:"inputs,omitempty"`
	// Whether empty tensors output should be passed to onwards
	// Default: false
	PassEmptyResponses bool `json:"passEmptyResponses,omitempty"`
	// msecs to wait for messages from multiple inputs to arrive before joining the inputs
	JoinWindowMs *uint32 `json:"joinWindowMs,omitempty"`
	// Map of tensor name conversions to use e.g. output1 -> input1
	TensorMap map[string]string `json:"tensorMap,omitempty"`
}

type PipelineOutput struct {
	// Previous step to receive data from
	Inputs []string `json:"inputs,omitempty"`
	// msecs to wait for messages from multiple inputs to arrive before joining the inputs
	JoinWindowMs uint32 `json:"joinWindowMs,omitempty"`
}

// PipelineStatus defines the observed state of Pipeline
type PipelineStatus struct {
	duckv1.Status `json:",inline"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Pipeline is the Schema for the pipelines API
type Pipeline struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PipelineSpec   `json:"spec,omitempty"`
	Status PipelineStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// PipelineList contains a list of Pipeline
type PipelineList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Pipeline `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Pipeline{}, &PipelineList{})
}

func (p Pipeline) AsSchedulerPipeline() *scheduler.Pipeline {
	var steps []*scheduler.PipelineStep
	var output *scheduler.PipelineOutput
	for _, step := range p.Spec.Steps {
		steps = append(steps, &scheduler.PipelineStep{
			Name:               step.Name,
			Inputs:             step.Inputs,
			JoinWindowMs:       step.JoinWindowMs,
			PassEmptyResponses: step.PassEmptyResponses,
			TensorMap:          step.TensorMap,
		})
	}
	if p.Spec.Output != nil {
		output = &scheduler.PipelineOutput{
			Inputs:       p.Spec.Output.Inputs,
			JoinWindowMs: p.Spec.Output.JoinWindowMs,
		}
	}
	return &scheduler.Pipeline{
		Name:           p.GetName(),
		Steps:          steps,
		Output:         output,
		KubernetesMeta: &scheduler.KubernetesMeta{Namespace: p.Namespace, Generation: p.Generation},
	}
}

const (
	PipelineReady apis.ConditionType = "PipelineReady"
)

var pipelineConditionSet = apis.NewLivingConditionSet(
	PipelineReady,
)

var _ apis.ConditionsAccessor = (*PipelineStatus)(nil)

func (ps *PipelineStatus) InitializeConditions() {
	pipelineConditionSet.Manage(ps).InitializeConditions()
}

func (ps *PipelineStatus) IsReady() bool {
	return pipelineConditionSet.Manage(ps).IsHappy()
}

func (ps *PipelineStatus) GetCondition(t apis.ConditionType) *apis.Condition {
	return pipelineConditionSet.Manage(ps).GetCondition(t)
}

func (ps *PipelineStatus) IsConditionReady(t apis.ConditionType) bool {
	return pipelineConditionSet.Manage(ps).GetCondition(t) != nil && pipelineConditionSet.Manage(ps).GetCondition(t).Status == v1.ConditionTrue
}

func (ps *PipelineStatus) SetCondition(conditionType apis.ConditionType, condition *apis.Condition) {
	switch {
	case condition == nil:
		pipelineConditionSet.Manage(ps).MarkUnknown(conditionType, "", "")
	case condition.Status == v1.ConditionUnknown:
		pipelineConditionSet.Manage(ps).MarkUnknown(conditionType, condition.Reason, condition.Message)
	case condition.Status == v1.ConditionTrue:
		pipelineConditionSet.Manage(ps).MarkTrueWithReason(conditionType, condition.Reason, condition.Message)
	case condition.Status == v1.ConditionFalse:
		pipelineConditionSet.Manage(ps).MarkFalse(conditionType, condition.Reason, condition.Message)
	}
}

func (ps *PipelineStatus) CreateAndSetCondition(conditionType apis.ConditionType, isTrue bool, reason string) {
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
	ps.SetCondition(conditionType, &condition)
}
