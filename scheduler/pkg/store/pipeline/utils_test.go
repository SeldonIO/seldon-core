package pipeline

import (
	"testing"

	. "github.com/onsi/gomega"
	"github.com/seldonio/seldon-core/scheduler/apis/mlops/scheduler"
)

func TestUpdateInputsSteps(t *testing.T) {
	g := NewGomegaWithT(t)
	type test struct {
		name     string
		inputs   []string
		expected []string
	}

	tests := []test{
		{
			name:     "test update inputs",
			inputs:   []string{"a", "a.outputs", "a.inputs", "a.inputs.t1"},
			expected: []string{"a.outputs", "a.outputs", "a.inputs", "a.inputs.t1"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			updated := updateInputSteps(test.inputs)
			g.Expect(updated).To(Equal(test.expected))
		})
	}
}

func TestCreatePipelineFromProto(t *testing.T) {
	g := NewGomegaWithT(t)
	type test struct {
		name     string
		proto    *scheduler.Pipeline
		version  uint32
		pipeline *PipelineVersion
		err      error
	}

	getUintPtr := func(val uint32) *uint32 { return &val }
	tests := []test{
		{
			name:    "simple",
			version: 1,
			proto: &scheduler.Pipeline{
				Name: "pipeline",
				Steps: []*scheduler.PipelineStep{
					{
						Name:   "a",
						Inputs: []string{},
					},
					{
						Name:   "b",
						Inputs: []string{"a"},
					},
				},
				Output: &scheduler.PipelineOutput{
					Steps: []string{"b"},
				},
			},
			pipeline: &PipelineVersion{
				Name:    "pipeline",
				Version: 1,
				Steps: map[string]*PipelineStep{
					"a": {
						Name:   "a",
						Inputs: []string{},
					},
					"b": {
						Name:   "b",
						Inputs: []string{"a.outputs"},
					},
				},
				Output: &PipelineOutput{
					Steps: []string{"b"},
				},
				State: &PipelineState{},
			},
		},
		{
			name:    "simple with tensor map",
			version: 1,
			proto: &scheduler.Pipeline{
				Name: "pipeline",
				Steps: []*scheduler.PipelineStep{
					{
						Name:   "a",
						Inputs: []string{},
					},
					{
						Name:      "b",
						Inputs:    []string{"a.outputs"},
						TensorMap: map[string]string{"output1": "input1"},
					},
				},
				Output: &scheduler.PipelineOutput{
					Steps: []string{"b"},
				},
			},
			pipeline: &PipelineVersion{
				Name:    "pipeline",
				Version: 1,
				Steps: map[string]*PipelineStep{
					"a": {
						Name:   "a",
						Inputs: []string{},
					},
					"b": {
						Name:      "b",
						Inputs:    []string{"a.outputs"},
						TensorMap: map[string]string{"output1": "input1"},
					},
				},
				Output: &PipelineOutput{
					Steps: []string{"b"},
				},
				State: &PipelineState{},
			},
		},
		{
			name:    "simple with join and batch",
			version: 1,
			proto: &scheduler.Pipeline{
				Name: "pipeline",
				Steps: []*scheduler.PipelineStep{
					{
						Name:   "a",
						Inputs: []string{},
					},
					{
						Name:      "b",
						Inputs:    []string{"a.outputs"},
						TensorMap: map[string]string{"output1": "input1"},
						OuterJoin: true,
						Batch: &scheduler.Batch{
							Size:     getUintPtr(100),
							WindowMs: getUintPtr(1000),
							Rolling:  true,
						},
					},
				},
				Output: &scheduler.PipelineOutput{
					Steps:     []string{"b"},
					OuterJoin: true,
				},
			},
			pipeline: &PipelineVersion{
				Name:    "pipeline",
				Version: 1,
				Steps: map[string]*PipelineStep{
					"a": {
						Name:   "a",
						Inputs: []string{},
					},
					"b": {
						Name:      "b",
						Inputs:    []string{"a.outputs"},
						TensorMap: map[string]string{"output1": "input1"},
						OuterJoin: true,
						Batch: &Batch{
							Size:     getUintPtr(100),
							WindowMs: getUintPtr(1000),
							Rolling:  true,
						},
					},
				},
				Output: &PipelineOutput{
					Steps:     []string{"b"},
					OuterJoin: true,
				},
				State: &PipelineState{},
			},
		},
		{
			name:    "simple with k8s meta",
			version: 1,
			proto: &scheduler.Pipeline{
				Name: "pipeline",
				Steps: []*scheduler.PipelineStep{
					{
						Name:   "step1",
						Inputs: []string{},
					},
				},
				KubernetesMeta: &scheduler.KubernetesMeta{
					Namespace:  "default",
					Generation: 22,
				},
			},
			pipeline: &PipelineVersion{
				Name:    "pipeline",
				Version: 1,
				Steps: map[string]*PipelineStep{
					"step1": {
						Name:   "step1",
						Inputs: []string{},
					},
				},
				State: &PipelineState{},
				KubernetesMeta: &KubernetesMeta{
					Namespace:  "default",
					Generation: 22,
				},
			},
		},
		{
			name:    "multi input",
			version: 1,
			proto: &scheduler.Pipeline{
				Name: "pipeline",
				Steps: []*scheduler.PipelineStep{
					{
						Name:   "step1",
						Inputs: []string{"a.inputs", "c.inputs.inp1", "d.outputs", "e.outputs.out1", "f"},
					},
				},
				Output: &scheduler.PipelineOutput{
					Steps: []string{"step1"},
				},
			},
			pipeline: &PipelineVersion{
				Name:    "pipeline",
				Version: 1,
				Steps: map[string]*PipelineStep{
					"step1": {
						Name:   "step1",
						Inputs: []string{"a.inputs", "c.inputs.inp1", "d.outputs", "e.outputs.out1", "f.outputs"},
					},
				},
				Output: &PipelineOutput{
					Steps: []string{"b"},
				},
				State: &PipelineState{},
			},
		},
		{
			name:    "pipeline step repeated",
			version: 1,
			proto: &scheduler.Pipeline{
				Name: "pipeline",
				Steps: []*scheduler.PipelineStep{
					{
						Name:   "foo",
						Inputs: []string{},
					},
					{
						Name:   "foo",
						Inputs: []string{},
					},
				},
				Output: &scheduler.PipelineOutput{
					Steps: []string{"b"},
				},
			},
			err: &PipelineStepRepeatedErr{pipeline: "pipeline", step: "foo"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			pv, err := CreatePipelineFromProto(test.proto, test.version)
			if test.err == nil {
				pv.UID = ""
				g.Expect(pv.Name).To(Equal(test.pipeline.Name))
				for k, v := range pv.Steps {
					g.Expect(test.pipeline.Steps[k]).To(Equal(v))
				}
			} else {
				g.Expect(err.Error()).To(Equal(test.err.Error()))
			}
		})
	}

}
