package pipeline

import (
	"fmt"
	"sort"
	"strings"

	"github.com/rs/xid"
	"github.com/seldonio/seldon-core/scheduler/apis/mlops/scheduler"
)

func CreateProtoFromPipeline(pv *PipelineVersion) *scheduler.Pipeline {
	var protoSteps []*scheduler.PipelineStep
	var protoOutput *scheduler.PipelineOutput
	keys := make([]string, 0)
	for k := range pv.Steps {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, stepName := range keys {
		step := pv.Steps[stepName]
		protoStep := &scheduler.PipelineStep{
			Name:         step.Name,
			Inputs:       step.Inputs,
			TensorMap:    step.TensorMap,
			JoinWindowMs: step.JoinWindowMs,
			Triggers:     step.Triggers,
		}
		switch step.InputsJoinType {
		case JoinInner:
			protoStep.InputsJoin = scheduler.PipelineStep_INNER
		case JoinOuter:
			protoStep.InputsJoin = scheduler.PipelineStep_OUTER
		case JoinAny:
			protoStep.InputsJoin = scheduler.PipelineStep_ANY
		}
		switch step.TriggersJoinType {
		case JoinInner:
			protoStep.TriggersJoin = scheduler.PipelineStep_INNER
		case JoinOuter:
			protoStep.TriggersJoin = scheduler.PipelineStep_OUTER
		case JoinAny:
			protoStep.TriggersJoin = scheduler.PipelineStep_ANY
		}
		if step.Batch != nil {
			protoStep.Batch = &scheduler.Batch{
				Size:     step.Batch.Size,
				WindowMs: step.Batch.WindowMs,
			}
		}
		protoSteps = append(protoSteps, protoStep)
	}
	if pv.Output != nil {
		protoOutput = &scheduler.PipelineOutput{
			Steps:        pv.Output.Steps,
			JoinWindowMs: pv.Output.JoinWindowMs,
			TensorMap:    pv.Output.TensorMap,
		}
		switch pv.Output.StepsJoinType {
		case JoinInner:
			protoOutput.StepsJoin = scheduler.PipelineOutput_INNER
		case JoinOuter:
			protoOutput.StepsJoin = scheduler.PipelineOutput_OUTER
		case JoinAny:
			protoOutput.StepsJoin = scheduler.PipelineOutput_ANY
		}
	}
	var kubernetesMeta *scheduler.KubernetesMeta
	if pv.KubernetesMeta != nil {
		kubernetesMeta = &scheduler.KubernetesMeta{
			Namespace:  pv.KubernetesMeta.Namespace,
			Generation: pv.KubernetesMeta.Generation,
		}
	}
	return &scheduler.Pipeline{
		Name:           pv.Name,
		Steps:          protoSteps,
		Output:         protoOutput,
		KubernetesMeta: kubernetesMeta,
	}
}

func CreatePipelineFromProto(pipelineProto *scheduler.Pipeline, version uint32) (*PipelineVersion, error) {
	steps := make(map[string]*PipelineStep)
	for _, stepProto := range pipelineProto.Steps {
		step := &PipelineStep{
			Name:         stepProto.GetName(),
			Inputs:       updateInputSteps(pipelineProto.Name, stepProto.Inputs),
			TensorMap:    stepProto.TensorMap,
			JoinWindowMs: stepProto.JoinWindowMs,
			Triggers:     updateInputSteps(pipelineProto.Name, stepProto.Triggers),
		}
		switch stepProto.InputsJoin {
		case scheduler.PipelineStep_INNER:
			step.InputsJoinType = JoinInner
		case scheduler.PipelineStep_OUTER:
			step.InputsJoinType = JoinOuter
		case scheduler.PipelineStep_ANY:
			step.InputsJoinType = JoinAny
		}
		switch stepProto.TriggersJoin {
		case scheduler.PipelineStep_INNER:
			step.TriggersJoinType = JoinInner
		case scheduler.PipelineStep_OUTER:
			step.TriggersJoinType = JoinOuter
		case scheduler.PipelineStep_ANY:
			step.TriggersJoinType = JoinAny
		}
		if stepProto.Batch != nil {
			step.Batch = &Batch{
				Size:     stepProto.Batch.Size,
				WindowMs: stepProto.Batch.WindowMs,
			}
		}
		if _, ok := steps[stepProto.Name]; ok {
			return nil, &PipelineStepRepeatedErr{pipeline: pipelineProto.GetName(), step: stepProto.GetName()}
		}
		steps[stepProto.Name] = step
	}

	var output *PipelineOutput
	if pipelineProto.Output != nil {
		output = &PipelineOutput{
			Steps:        updateInputSteps(pipelineProto.Name, pipelineProto.Output.Steps),
			JoinWindowMs: pipelineProto.Output.JoinWindowMs,
			TensorMap:    pipelineProto.Output.TensorMap,
		}
		switch pipelineProto.Output.StepsJoin {
		case scheduler.PipelineOutput_INNER:
			output.StepsJoinType = JoinInner
		case scheduler.PipelineOutput_OUTER:
			output.StepsJoinType = JoinOuter
		case scheduler.PipelineOutput_ANY:
			output.StepsJoinType = JoinAny
		}
	}
	var kubernetesMeta *KubernetesMeta
	if pipelineProto.KubernetesMeta != nil {
		kubernetesMeta = &KubernetesMeta{
			Namespace:  pipelineProto.KubernetesMeta.Namespace,
			Generation: pipelineProto.KubernetesMeta.Generation,
		}
	}

	return &PipelineVersion{
		Name:           pipelineProto.Name,
		Version:        version,
		UID:            xid.New().String(),
		Steps:          steps,
		State:          &PipelineState{},
		Output:         output,
		KubernetesMeta: kubernetesMeta,
	}, nil
}

func updateInputSteps(pipelineName string, inputs []string) []string {
	if len(inputs) == 0 {
		return inputs
	}
	var updatedInputs []string
	for _, inp := range inputs {
		parts := strings.Split(inp, StepNameSeperator)
		switch len(parts) {
		case 1:
			// For pipeline name we default to inputs otherwise as its a previous step being referred to we default to outputs
			if inp == pipelineName {
				updatedInputs = append(updatedInputs, fmt.Sprintf("%s.%s", inp, StepInputSpecifier))
			} else {
				updatedInputs = append(updatedInputs, fmt.Sprintf("%s.%s", inp, StepOutputSpecifier))
			}
		default:
			updatedInputs = append(updatedInputs, inp)
		}
	}
	return updatedInputs
}
