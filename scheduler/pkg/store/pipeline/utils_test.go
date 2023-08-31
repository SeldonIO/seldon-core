/*
Copyright 2022 Seldon Technologies Ltd.

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

package pipeline

import (
	"testing"

	. "github.com/onsi/gomega"

	"github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"
)

func TestUpdateInputsSteps(t *testing.T) {
	g := NewGomegaWithT(t)
	type test struct {
		name         string
		pipelineName string
		inputs       []string
		expected     []string
	}

	tests := []test{
		{
			name:         "test update inputs",
			pipelineName: "pipeline",
			inputs:       []string{"a", "a.outputs", "a.inputs", "a.inputs.t1", "pipeline", "pipeline.inputs.t1"},
			expected:     []string{"a.outputs", "a.outputs", "a.inputs", "a.inputs.t1", "pipeline.inputs", "pipeline.inputs.t1"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			updated := updateInternalInputSteps(test.pipelineName, test.inputs)
			g.Expect(updated).To(Equal(test.expected))
		})
	}
}

func TestCreatePipelineFromProto(t *testing.T) {
	g := NewGomegaWithT(t)
	type test struct {
		name     string
		proto    *scheduler.Pipeline
		pipeline *PipelineVersion
		err      error
	}

	getUintPtr := func(val uint32) *uint32 { return &val }
	tests := []test{
		{
			name: "simple",
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
			name: "with pipeline input",
			proto: &scheduler.Pipeline{
				Name: "pipeline",
				Steps: []*scheduler.PipelineStep{
					{
						Name:   "a",
						Inputs: []string{"pipeline"},
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
				Name: "pipeline",
				Steps: map[string]*PipelineStep{
					"a": {
						Name:   "a",
						Inputs: []string{"pipeline.inputs"},
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
			name: "simple with tensor map",
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
					Steps:     []string{"b"},
					TensorMap: map[string]string{"output1": "output"},
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
					Steps:     []string{"b"},
					TensorMap: map[string]string{"output1": "output"},
				},
				State: &PipelineState{},
			},
		},
		{
			name: "simple with join and batch",
			proto: &scheduler.Pipeline{
				Name: "pipeline",
				Steps: []*scheduler.PipelineStep{
					{
						Name:   "a",
						Inputs: []string{},
					},
					{
						Name:       "b",
						Inputs:     []string{"a.outputs"},
						TensorMap:  map[string]string{"output1": "input1"},
						InputsJoin: scheduler.PipelineStep_OUTER,
						Batch: &scheduler.Batch{
							Size:     getUintPtr(100),
							WindowMs: getUintPtr(1000),
						},
					},
				},
				Output: &scheduler.PipelineOutput{
					Steps:     []string{"b"},
					StepsJoin: scheduler.PipelineOutput_INNER,
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
						Name:           "b",
						Inputs:         []string{"a.outputs"},
						TensorMap:      map[string]string{"output1": "input1"},
						InputsJoinType: JoinOuter,
						Batch: &Batch{
							Size:     getUintPtr(100),
							WindowMs: getUintPtr(1000),
						},
					},
				},
				Output: &PipelineOutput{
					Steps:         []string{"b"},
					StepsJoinType: JoinInner,
				},
				State: &PipelineState{},
			},
		},
		{
			name: "simple with trigger",
			proto: &scheduler.Pipeline{
				Name: "pipeline",
				Steps: []*scheduler.PipelineStep{
					{
						Name:   "a",
						Inputs: []string{},
					},
					{
						Name:         "b",
						Inputs:       []string{"a.outputs"},
						TensorMap:    map[string]string{"output1": "input1"},
						InputsJoin:   scheduler.PipelineStep_OUTER,
						Triggers:     []string{"a"},
						TriggersJoin: scheduler.PipelineStep_INNER,
						Batch: &scheduler.Batch{
							Size:     getUintPtr(100),
							WindowMs: getUintPtr(1000),
						},
					},
				},
				Output: &scheduler.PipelineOutput{
					Steps:     []string{"b"},
					StepsJoin: scheduler.PipelineOutput_OUTER,
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
						Name:             "b",
						Inputs:           []string{"a.outputs"},
						TensorMap:        map[string]string{"output1": "input1"},
						InputsJoinType:   JoinOuter,
						Triggers:         []string{"a.outputs"},
						TriggersJoinType: JoinInner,
						Batch: &Batch{
							Size:     getUintPtr(100),
							WindowMs: getUintPtr(1000),
						},
					},
				},
				Output: &PipelineOutput{
					Steps:         []string{"b"},
					StepsJoinType: JoinOuter,
				},
				State: &PipelineState{},
			},
		},
		{
			name: "simple with k8s meta",
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
			name: "multi input",
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
			name: "pipeline step repeated",
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
			pv, err := CreatePipelineVersionFromProto(test.proto)
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

func TestUpdateExternalInputSteps(t *testing.T) {
	g := NewGomegaWithT(t)
	type test struct {
		name     string
		inputs   []string
		expected []string
	}

	tests := []test{
		{
			name:     "test update external inputs",
			inputs:   []string{"p1", "p1.outputs", "p1.inputs", "p1.inputs.t1", "p1.step.m1", "p1.step.m1.outputs.t1"},
			expected: []string{"p1.outputs", "p1.outputs", "p1.inputs", "p1.inputs.t1", "p1.step.m1.outputs", "p1.step.m1.outputs.t1"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			updated := updateExternalInputSteps(test.inputs)
			g.Expect(updated).To(Equal(test.expected))
		})
	}
}
