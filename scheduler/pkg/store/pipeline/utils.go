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
			OuterJoin:    step.OuterJoin,
		}
		if step.Batch != nil {
			protoStep.Batch = &scheduler.Batch{
				Size:     step.Batch.Size,
				WindowMs: step.Batch.WindowMs,
				Rolling:  step.Batch.Rolling,
			}
		}
		protoSteps = append(protoSteps, protoStep)
	}
	if pv.Output != nil {
		protoOutput = &scheduler.PipelineOutput{
			Steps:        pv.Output.Steps,
			JoinWindowMs: pv.Output.JoinWindowMs,
			OuterJoin:    pv.Output.OuterJoin,
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
			Inputs:       updateInputSteps(stepProto.Inputs),
			TensorMap:    stepProto.TensorMap,
			JoinWindowMs: stepProto.JoinWindowMs,
			OuterJoin:    stepProto.OuterJoin,
		}
		if stepProto.Batch != nil {
			step.Batch = &Batch{
				Size:     stepProto.Batch.Size,
				WindowMs: stepProto.Batch.WindowMs,
				Rolling:  stepProto.Batch.Rolling,
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
			Steps:        updateInputSteps(pipelineProto.Output.Steps),
			JoinWindowMs: pipelineProto.Output.JoinWindowMs,
			OuterJoin:    pipelineProto.Output.OuterJoin,
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

func updateInputSteps(inputs []string) []string {
	if len(inputs) == 0 {
		return inputs
	}
	var updatedInputs []string
	for _, inp := range inputs {
		parts := strings.Split(inp, StepNameSeperator)
		switch len(parts) {
		case 1:
			updatedInputs = append(updatedInputs, fmt.Sprintf("%s.%s", inp, StepOutputSpecifier))
		default:
			updatedInputs = append(updatedInputs, inp)
		}
	}
	return updatedInputs
}
