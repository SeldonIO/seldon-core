package pipeline

import (
	"sort"

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
			Name:               step.Name,
			Inputs:             step.Inputs,
			PassEmptyResponses: step.PassEmptyOutputs,
			JoinWindowMs:       step.JoinWindowMs,
		}
		protoSteps = append(protoSteps, protoStep)
	}
	if pv.Output != nil {
		protoOutput = &scheduler.PipelineOutput{
			Inputs:       pv.Output.Inputs,
			JoinWindowMs: pv.Output.JoinWindowMs,
		}
	}
	return &scheduler.Pipeline{
		Name:   pv.Name,
		Steps:  protoSteps,
		Output: protoOutput,
	}
}

func CreatePipelineFromProto(pipelineProto *scheduler.Pipeline, version uint32) (*PipelineVersion, error) {
	steps := make(map[string]*PipelineStep)
	for _, stepProto := range pipelineProto.Steps {
		step := &PipelineStep{
			Name:             stepProto.GetName(),
			Inputs:           stepProto.Inputs,
			PassEmptyOutputs: stepProto.PassEmptyResponses,
			JoinWindowMs:     stepProto.JoinWindowMs,
		}
		steps[stepProto.Name] = step
	}
	var output *PipelineOutput
	if pipelineProto.Output != nil {
		output = &PipelineOutput{
			Inputs:       pipelineProto.Output.Inputs,
			JoinWindowMs: pipelineProto.Output.JoinWindowMs,
		}
	}

	return &PipelineVersion{
		Name:    pipelineProto.Name,
		Version: version,
		UID:     xid.New().String(),
		Steps:   steps,
		State:   &PipelineState{},
		Output:  output,
	}, nil
}
