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
						Inputs: []string{"a.outputs.t1", "a.inputs", "a.outputs"},
					},
					"c": {
						Name:   "c",
						Inputs: []string{"b.inputs"},
					},
				},
				Output: &PipelineOutput{
					Steps: []string{"c.outputs"},
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
		{
			name: "pipeline input reference",
			pipelineVersion: &PipelineVersion{
				Name: "test",
				Steps: map[string]*PipelineStep{
					"a": {
						Name:   "a",
						Inputs: []string{"test.inputs"},
					},
				},
			},
		},
		{
			name: "output step does not exist",
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
				},
				Output: &PipelineOutput{
					Steps: []string{"a", "b", "foo"},
				},
			},
			err: &PipelineOutputStepNotFoundErr{pipeline: "test", step: "foo"},
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
						Inputs: []string{"a.outputs.t1", "a.inputs", "a.outputs"},
					},
					"c": {
						Name:   "c",
						Inputs: []string{"a.outputs.t1"},
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

func TestCheckForCycles(t *testing.T) {
	g := NewGomegaWithT(t)
	tests := []validateTest{
		{
			name: "no loops",
			pipelineVersion: &PipelineVersion{
				Name: "test",
				Steps: map[string]*PipelineStep{
					"a": {
						Name: "a",
					},
					"b": {
						Name:   "b",
						Inputs: []string{"a.outputs.t1", "a.inputs", "a.outputs"},
					},
					"c": {
						Name:   "c",
						Inputs: []string{"a.outputs.t1"},
					},
				},
			},
		},
		{
			name: "loop",
			pipelineVersion: &PipelineVersion{
				Name: "test",
				Steps: map[string]*PipelineStep{
					"a": {
						Name: "a",
					},
					"b": {
						Name:   "b",
						Inputs: []string{"a.outputs", "c.outputs"},
					},
					"c": {
						Name:   "c",
						Inputs: []string{"b.outputs"},
					},
				},
			},
			err: &PipelineCycleErr{pipeline: "test"},
		},
		{
			name: "loop via trigger",
			pipelineVersion: &PipelineVersion{
				Name: "test",
				Steps: map[string]*PipelineStep{
					"a": {
						Name: "a",
					},
					"b": {
						Name:     "b",
						Inputs:   []string{"a.outputs"},
						Triggers: []string{"c.outputs"},
					},
					"c": {
						Name:   "c",
						Inputs: []string{"b.outputs"},
					},
				},
			},
			err: &PipelineCycleErr{pipeline: "test"},
		},
		{
			name: "separate loop",
			pipelineVersion: &PipelineVersion{
				Name: "test",
				Steps: map[string]*PipelineStep{
					"a": {
						Name: "a",
					},
					"b": {
						Name:   "b",
						Inputs: []string{"a.outputs"},
					},
					"c": {
						Name:   "c",
						Inputs: []string{"a.outputs", "d.outputs"},
					},
					"d": {
						Name:   "c",
						Inputs: []string{"c.outputs"},
					},
				},
			},
			err: &PipelineCycleErr{pipeline: "test"},
		},
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
						Inputs: []string{"a.outputs.t1", "a.inputs", "a.outputs"},
					},
					"c": {
						Name:     "c",
						Inputs:   []string{"b.outputs", "a.outputs"},
						Triggers: []string{},
					},
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := checkForCycles(test.pipelineVersion)
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

func TestCheckInputsAndTriggersDiffer(t *testing.T) {
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
						Inputs: []string{"a.outputs.t1", "a.inputs", "a.outputs"},
					},
					"c": {
						Name:     "c",
						Inputs:   []string{"b.outputs", "a.outputs"},
						Triggers: []string{},
					},
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := checkInputsAndTriggersDiffer(test.pipelineVersion)
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

func TestCheckStepNameNotPipelineName(t *testing.T) {
	g := NewGomegaWithT(t)
	tests := []validateTest{
		{
			name: "step has same name as pipeline",
			pipelineVersion: &PipelineVersion{
				Name: "test",
				Steps: map[string]*PipelineStep{
					"test": {
						Name: "test",
					},
					"b": {
						Name:   "b",
						Inputs: []string{"a.outputs.t1", "a.inputs", "a.outputs"},
					},
					"c": {
						Name:     "c",
						Inputs:   []string{"b.outputs", "a.outputs"},
						Triggers: []string{},
					},
				},
			},
			err: &PipelineStepNameEqualsPipelineNameErr{pipeline: "test"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := checkStepNameNotPipelineName(test.pipelineVersion)
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
