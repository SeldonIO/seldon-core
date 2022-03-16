package pipeline

import (
	"testing"

	. "github.com/onsi/gomega"
)

type validateTest struct {
	name            string
	pipelineVersion *PipelineVersion
	err             error
}

func TestCheckStepReferencesExist(t *testing.T) {
	g := NewGomegaWithT(t)
	tests := []validateTest{
		{
			name: "valid references",
			pipelineVersion: &PipelineVersion{
				Name: "test",
				Steps: map[string]*PipelineStep{
					"a": {
						Name: "a",
					},
					"b": {
						Name:   "b",
						Inputs: []string{"a.outputs.t1", "a.inputs", "a.outputs", "a"},
					},
					"c": {
						Name:   "c",
						Inputs: []string{"b"},
					},
					"": {
						Name:   "",
						Inputs: []string{"c"},
					},
				},
			},
		},
		{
			name: "step does not exist",
			pipelineVersion: &PipelineVersion{
				Name: "test",
				Steps: map[string]*PipelineStep{
					"a": {
						Name: "a",
					},
					"b": {
						Name:   "b",
						Inputs: []string{"a.outputs.out1"},
					},
					"c": {
						Name:   "c",
						Inputs: []string{"f"},
					},
				},
			},
			err: &PipelineStepNotFoundErr{pipeline: "test", step: "c", badRef: "f"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := checkStepReferencesExist(test.pipelineVersion)
			if test.err == nil {
				g.Expect(err).To(BeNil())
			} else {
				g.Expect(err.Error()).To(Equal(test.err.Error()))
			}
			err = validate(test.pipelineVersion)
			if test.err == nil {
				g.Expect(err).To(BeNil())
			} else {
				g.Expect(err.Error()).To(Equal(test.err.Error()))
			}
		})
	}
}

func TestCheckStepInputsSpecifier(t *testing.T) {
	g := NewGomegaWithT(t)
	tests := []validateTest{
		{
			name: "valid inputs",
			pipelineVersion: &PipelineVersion{
				Name: "test",
				Steps: map[string]*PipelineStep{
					"a": {
						Name: "a",
					},
					"b": {
						Name:   "b",
						Inputs: []string{"a.outputs.t1", "a.inputs", "a.outputs", "a"},
					},
					"c": {
						Name:   "c",
						Inputs: []string{"a.outputs.t1", "b"},
					},
				},
			},
		},
		{
			name: "bad specifier not inputs or ouputs",
			pipelineVersion: &PipelineVersion{
				Name: "test",
				Steps: map[string]*PipelineStep{
					"a": {
						Name: "a",
					},
					"b": {
						Name:   "b",
						Inputs: []string{"a.ffo.t1"},
					},
				},
			},
			err: &PipelineStepInputSpecifierErr{pipeline: "test", step: "b", outputStep: "a.ffo.t1"},
		},
		{
			name: "bad specifier has too many parts",
			pipelineVersion: &PipelineVersion{
				Name: "test",
				Steps: map[string]*PipelineStep{
					"a": {
						Name: "a",
					},
					"b": {
						Name:   "b",
						Inputs: []string{"a.inputs.t1.foo"},
					},
				},
			},
			err: &PipelineStepInputSpecifierErr{pipeline: "test", step: "b", outputStep: "a.inputs.t1.foo"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := checkStepInputs(test.pipelineVersion)
			if test.err == nil {
				g.Expect(err).To(BeNil())
			} else {
				g.Expect(err.Error()).To(Equal(test.err.Error()))
			}
			err = validate(test.pipelineVersion)
			if test.err == nil {
				g.Expect(err).To(BeNil())
			} else {
				g.Expect(err.Error()).To(Equal(test.err.Error()))
			}
		})
	}
}

func TestCheckOnlyOneInput(t *testing.T) {
	g := NewGomegaWithT(t)
	tests := []validateTest{
		{
			name: "single input ok",
			pipelineVersion: &PipelineVersion{
				Name: "test",
				Steps: map[string]*PipelineStep{
					"a": {
						Name: "a",
					},
					"b": {
						Name:   "b",
						Inputs: []string{"a"},
					},
					"c": {
						Name:   "c",
						Inputs: []string{"b"},
					},
					"": {
						Name:   "",
						Inputs: []string{"c"},
					},
				},
			},
		},
		{
			name: "multiple inputs",
			pipelineVersion: &PipelineVersion{
				Name: "test",
				Steps: map[string]*PipelineStep{
					"a": {
						Name: "a",
					},
					"b": {
						Name: "b",
					},
				},
			},
			err: &PipelineMultipleInputsErr{pipeline: "test"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := checkOnlyOneInput(test.pipelineVersion)
			if test.err == nil {
				g.Expect(err).To(BeNil())
			} else {
				g.Expect(err.Error()).To(Equal(test.err.Error()))
			}
			err = validate(test.pipelineVersion)
			if test.err == nil {
				g.Expect(err).To(BeNil())
			} else {
				g.Expect(err.Error()).To(Equal(test.err.Error()))
			}
		})
	}
}
