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

	"github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"
)

// PipelineSpec defines the desired state of Pipeline
type PipelineSpec struct {
	// External inputs to this pipeline, optional
	Input *PipelineInput `json:"input,omitempty"`
	// The steps of this inference graph pipeline
	Steps []PipelineStep `json:"steps"`
	// Synchronous output from this pipeline, optional
	Output *PipelineOutput `json:"output,omitempty"`
}

type JoinType string

const (
	JoinTypeInner JoinType = "inner"
	JoinTypeOuter JoinType = "outer"
	JoinTypeAny   JoinType = "any"
)

type PipelineStep struct {
	// Name of the step
	Name string `json:"name"`
	// Previous step to receive data from
	Inputs []string `json:"inputs,omitempty"`
	// msecs to wait for messages from multiple inputs to arrive before joining the inputs
	JoinWindowMs *uint32 `json:"joinWindowMs,omitempty"`
	// Map of tensor name conversions to use e.g. output1 -> input1
	TensorMap map[string]string `json:"tensorMap,omitempty"`
	// Triggers required to activate step
	Triggers []string `json:"triggers,omitempty"`
	// One of inner (default), outer, or any
	// inner - do an inner join: data must be available from all inputs
	// outer - do an outer join: data will include any data from any inputs at end of window
	// any - first data input that arrives will be forwarded
	InputsJoinType *JoinType `json:"inputsJoinType,omitempty"`
	// One of inner (default), outer, or any (see above for details)
	TriggersJoinType *JoinType `json:"triggersJoinType,omitempty"`
	// Batch size of request required before data will be sent to this step
	Batch *PipelineBatch `json:"batch,omitempty"`
}

type PipelineBatch struct {
	Size     *uint32 `json:"size,omitempty"`
	WindowMs *uint32 `json:"windowMs,omitempty"`
	Rolling  bool    `json:"rolling,omitempty"`
}

type PipelineInput struct {
	// Previous external pipeline steps to receive data from
	ExternalInputs []string `json:"externalInputs,omitempty"`
	// Triggers required to activate inputs
	ExternalTriggers []string `json:"externalTriggers,omitempty"`
	// msecs to wait for messages from multiple inputs to arrive before joining the inputs
	JoinWindowMs *uint32 `json:"joinWindowMs,omitempty"`
	// One of inner (default), outer, or any (see above for details)
	JoinType *JoinType `json:"joinType,omitempty"`
	// One of inner (default), outer, or any (see above for details)
	TriggersJoinType *JoinType `json:"triggersJoinType,omitempty"`
	// Map of tensor name conversions to use e.g. output1 -> input1
	TensorMap map[string]string `json:"tensorMap,omitempty"`
}

type PipelineOutput struct {
	// Previous step to receive data from
	Steps []string `json:"steps,omitempty"`
	// msecs to wait for messages from multiple inputs to arrive before joining the inputs
	JoinWindowMs uint32 `json:"joinWindowMs,omitempty"`
	// One of inner (default), outer, or any (see above for details)
	StepsJoin *JoinType `json:"stepsJoin,omitempty"`
	// Map of tensor name conversions to use e.g. output1 -> input1
	TensorMap map[string]string `json:"tensorMap,omitempty"`
}

// PipelineStatus defines the observed state of Pipeline
type PipelineStatus struct {
	duckv1.Status `json:",inline"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:shortName=mlp

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
	var input *scheduler.PipelineInput
	if p.Spec.Input != nil {
		input = &scheduler.PipelineInput{
			ExternalInputs:   p.Spec.Input.ExternalInputs,
			ExternalTriggers: p.Spec.Input.ExternalTriggers,
			JoinWindowMs:     p.Spec.Input.JoinWindowMs,
			TensorMap:        p.Spec.Input.TensorMap,
		}
		if p.Spec.Input.JoinType != nil {
			switch *p.Spec.Input.JoinType {
			case JoinTypeInner:
				input.JoinType = scheduler.PipelineInput_INNER
			case JoinTypeOuter:
				input.JoinType = scheduler.PipelineInput_OUTER
			case JoinTypeAny:
				input.JoinType = scheduler.PipelineInput_ANY
			default:
				input.JoinType = scheduler.PipelineInput_INNER
			}
		}
		if p.Spec.Input.TriggersJoinType != nil {
			switch *p.Spec.Input.TriggersJoinType {
			case JoinTypeInner:
				input.TriggersJoin = scheduler.PipelineInput_INNER
			case JoinTypeOuter:
				input.TriggersJoin = scheduler.PipelineInput_OUTER
			case JoinTypeAny:
				input.TriggersJoin = scheduler.PipelineInput_ANY
			default:
				input.TriggersJoin = scheduler.PipelineInput_INNER
			}
		}
	}
	for _, step := range p.Spec.Steps {
		pipelineStep := &scheduler.PipelineStep{
			Name:         step.Name,
			Inputs:       step.Inputs,
			JoinWindowMs: step.JoinWindowMs,
			TensorMap:    step.TensorMap,
			Triggers:     step.Triggers,
		}
		if step.InputsJoinType != nil {
			switch *step.InputsJoinType {
			case JoinTypeInner:
				pipelineStep.InputsJoin = scheduler.PipelineStep_INNER
			case JoinTypeOuter:
				pipelineStep.InputsJoin = scheduler.PipelineStep_OUTER
			case JoinTypeAny:
				pipelineStep.InputsJoin = scheduler.PipelineStep_ANY
			default:
				pipelineStep.InputsJoin = scheduler.PipelineStep_INNER
			}
		}
		if step.TriggersJoinType != nil {
			switch *step.TriggersJoinType {
			case JoinTypeInner:
				pipelineStep.TriggersJoin = scheduler.PipelineStep_INNER
			case JoinTypeOuter:
				pipelineStep.TriggersJoin = scheduler.PipelineStep_OUTER
			case JoinTypeAny:
				pipelineStep.TriggersJoin = scheduler.PipelineStep_ANY
			default:
				pipelineStep.TriggersJoin = scheduler.PipelineStep_INNER
			}
		}
		if step.Batch != nil {
			pipelineStep.Batch = &scheduler.Batch{
				Size:     step.Batch.Size,
				WindowMs: step.Batch.WindowMs,
			}
		}
		steps = append(steps, pipelineStep)
	}
	if p.Spec.Output != nil {
		output = &scheduler.PipelineOutput{
			Steps:        p.Spec.Output.Steps,
			JoinWindowMs: p.Spec.Output.JoinWindowMs,
			TensorMap:    p.Spec.Output.TensorMap,
		}
		if p.Spec.Output.StepsJoin != nil {
			switch *p.Spec.Output.StepsJoin {
			case JoinTypeInner:
				output.StepsJoin = scheduler.PipelineOutput_INNER
			case JoinTypeOuter:
				output.StepsJoin = scheduler.PipelineOutput_OUTER
			case JoinTypeAny:
				output.StepsJoin = scheduler.PipelineOutput_ANY
			default:
				output.StepsJoin = scheduler.PipelineOutput_INNER
			}
		}
	}
	return &scheduler.Pipeline{
		Name:           p.GetName(),
		Uid:            "", // ID Will be set on scheduler side. IDs don't change on k8s when updates are made so can't use it for each version
		Input:          input,
		Steps:          steps,
		Output:         output,
		KubernetesMeta: &scheduler.KubernetesMeta{Namespace: p.Namespace, Generation: p.Generation},
	}
}

const (
	PipelineReady apis.ConditionType = "PipelineReady"
	ModelsReady   apis.ConditionType = "ModelsReady"
)

var pipelineConditionSet = apis.NewLivingConditionSet(
	PipelineReady,
	ModelsReady,
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

func (ps *PipelineStatus) CreateAndSetCondition(
	conditionType apis.ConditionType,
	isTrue bool,
	message string,
	reason string,
) {
	condition := apis.Condition{}
	if isTrue {
		condition.Status = v1.ConditionTrue
	} else {
		condition.Status = v1.ConditionFalse
	}
	condition.Type = conditionType
	condition.Message = message
	condition.Reason = reason
	condition.LastTransitionTime = apis.VolatileTime{
		Inner: metav1.Now(),
	}
	ps.SetCondition(conditionType, &condition)
}
