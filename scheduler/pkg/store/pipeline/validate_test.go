/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

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
						Inputs: []string{"f.outputs"},
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

func TestCheckStepInputs(t *testing.T) {
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
			err: &PipelineStepInputSpecifierErr{pipeline: "test", step: "b", outputStep: "a.ffo.t1", isTrigger: false},
		},
		{
			name: "empty input name in step",
			pipelineVersion: &PipelineVersion{
				Name: "test",
				Steps: map[string]*PipelineStep{
					"a": {
						Name: "a",
					},
					"b": {
						Name:   "b",
						Inputs: []string{""},
					},
				},
			},
			err: &PipelineStepInputEmptyErr{pipeline: "test", stepName: "b", isTrigger: false},
		},
		{
			name: "empty input name in step with spaces",
			pipelineVersion: &PipelineVersion{
				Name: "test",
				Steps: map[string]*PipelineStep{
					"a": {
						Name: "a",
					},
					"b": {
						Name:   "b",
						Inputs: []string{"  "},
					},
				},
			},
			err: &PipelineStepInputEmptyErr{pipeline: "test", stepName: "b", isTrigger: false},
		},
		{
			name: "step separator at start",
			pipelineVersion: &PipelineVersion{
				Name: "test",
				Steps: map[string]*PipelineStep{
					"a": {
						Name: "a",
					},
					"b": {
						Name:   "b",
						Inputs: []string{".outputs"},
					},
				},
			},
			err: &PipelineStepInputEmptyErr{pipeline: "test", stepName: "b", isTrigger: false},
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
			err: &PipelineStepInputSpecifierErr{pipeline: "test", step: "b", outputStep: "a.inputs.t1.foo", isTrigger: false},
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

func TestCheckStepTriggers(t *testing.T) {
	g := NewGomegaWithT(t)
	tests := []validateTest{
		{
			name: "valid triggers",
			pipelineVersion: &PipelineVersion{
				Name: "test",
				Steps: map[string]*PipelineStep{
					"a": {
						Name: "a",
					},
					"b": {
						Name:     "b",
						Triggers: []string{"a.outputs.t1", "a.inputs", "a.outputs"},
					},
					"c": {
						Name:     "c",
						Triggers: []string{"a.outputs.t1"},
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
						Name:     "b",
						Triggers: []string{"a.ffo.t1"},
					},
				},
			},
			err: &PipelineStepInputSpecifierErr{pipeline: "test", step: "b", outputStep: "a.ffo.t1", isTrigger: true},
		},
		{
			name: "empty input name in step",
			pipelineVersion: &PipelineVersion{
				Name: "test",
				Steps: map[string]*PipelineStep{
					"a": {
						Name: "a",
					},
					"b": {
						Name:     "b",
						Triggers: []string{""},
					},
				},
			},
			err: &PipelineStepInputEmptyErr{pipeline: "test", stepName: "b", isTrigger: true},
		},
		{
			name: "empty input name in step with spaces",
			pipelineVersion: &PipelineVersion{
				Name: "test",
				Steps: map[string]*PipelineStep{
					"a": {
						Name: "a",
					},
					"b": {
						Name:     "b",
						Triggers: []string{"  "},
					},
				},
			},
			err: &PipelineStepInputEmptyErr{pipeline: "test", stepName: "b", isTrigger: true},
		},
		{
			name: "step separator at start",
			pipelineVersion: &PipelineVersion{
				Name: "test",
				Steps: map[string]*PipelineStep{
					"a": {
						Name: "a",
					},
					"b": {
						Name:     "b",
						Triggers: []string{".outputs"},
					},
				},
			},
			err: &PipelineStepInputEmptyErr{pipeline: "test", stepName: "b", isTrigger: true},
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
						Name:     "b",
						Triggers: []string{"a.inputs.t1.foo"},
					},
				},
			},
			err: &PipelineStepInputSpecifierErr{pipeline: "test", step: "b", outputStep: "a.inputs.t1.foo", isTrigger: true},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := checkStepTriggers(test.pipelineVersion)
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

func TestPipelineInput(t *testing.T) {
	g := NewGomegaWithT(t)
	tests := []validateTest{
		{
			name: "No input",
			pipelineVersion: &PipelineVersion{
				Name: "test",
				Steps: map[string]*PipelineStep{
					"a": {
						Name: "a",
					},
				},
			},
		},
		{
			name: "No external inputs",
			pipelineVersion: &PipelineVersion{
				Name:  "test",
				Input: &PipelineInput{},
				Steps: map[string]*PipelineStep{
					"a": {
						Name: "a",
					},
				},
			},
			err: &PipelineInputErr{"test", "", pipelineInputEmptyErr},
		},
		{
			name: "Valid pipeline input",
			pipelineVersion: &PipelineVersion{
				Name: "test",
				Input: &PipelineInput{
					ExternalInputs: []string{
						"foo.inputs",
					},
				},
				Steps: map[string]*PipelineStep{
					"a": {
						Name: "a",
					},
				},
			},
		},
		{
			name: "Valid pipeline outputs",
			pipelineVersion: &PipelineVersion{
				Name: "test",
				Input: &PipelineInput{
					ExternalInputs: []string{
						"foo.outputs",
					},
				},
				Steps: map[string]*PipelineStep{
					"a": {
						Name: "a",
					},
				},
			},
		},
		{
			name: "Bad input specifier",
			pipelineVersion: &PipelineVersion{
				Name: "test",
				Input: &PipelineInput{
					ExternalInputs: []string{
						"foo.foo",
					},
				},
				Steps: map[string]*PipelineStep{
					"a": {
						Name: "a",
					},
				},
			},
			err: &PipelineInputErr{"test", "foo.foo", pipelineInputInvalidPrefixReason},
		},
		{
			name: "Bad input specifier",
			pipelineVersion: &PipelineVersion{
				Name: "test",
				Input: &PipelineInput{
					ExternalInputs: []string{
						"foo.step",
					},
				},
				Steps: map[string]*PipelineStep{
					"a": {
						Name: "a",
					},
				},
			},
			err: &PipelineInputErr{"test", "foo.step", pipelineInputInvalidPrefixReason},
		},
		{
			name: "Bad input step no suffix",
			pipelineVersion: &PipelineVersion{
				Name: "test",
				Input: &PipelineInput{
					ExternalInputs: []string{
						"foo.step.bar",
					},
				},
				Steps: map[string]*PipelineStep{
					"a": {
						Name: "a",
					},
				},
			},
			err: &PipelineInputErr{"test", "foo.step.bar", pipelineInputStepBadSuffix},
		},
		{
			name: "Bad input step no suffix",
			pipelineVersion: &PipelineVersion{
				Name: "test",
				Input: &PipelineInput{
					ExternalInputs: []string{
						"foo.step.bar.zee",
					},
				},
				Steps: map[string]*PipelineStep{
					"a": {
						Name: "a",
					},
				},
			},
			err: &PipelineInputErr{"test", "foo.step.bar.zee", pipelineInputStepBadSuffix},
		},
		{
			name: "Bad input step inputs ok",
			pipelineVersion: &PipelineVersion{
				Name: "test",
				Input: &PipelineInput{
					ExternalInputs: []string{
						"foo.step.bar.inputs",
					},
				},
				Steps: map[string]*PipelineStep{
					"a": {
						Name: "a",
					},
				},
			},
		},
		{
			name: "Bad input step outputs ok",
			pipelineVersion: &PipelineVersion{
				Name: "test",
				Input: &PipelineInput{
					ExternalInputs: []string{
						"foo.step.bar.outputs",
					},
				},
				Steps: map[string]*PipelineStep{
					"a": {
						Name: "a",
					},
				},
			},
		},
		{
			name: "input step inputs tensor ok",
			pipelineVersion: &PipelineVersion{
				Name: "test",
				Input: &PipelineInput{
					ExternalInputs: []string{
						"foo.step.bar.inputs.tensor",
					},
				},
				Steps: map[string]*PipelineStep{
					"a": {
						Name: "a",
					},
				},
			},
		},
		{
			name: "Bad input step inputs too long",
			pipelineVersion: &PipelineVersion{
				Name: "test",
				Input: &PipelineInput{
					ExternalInputs: []string{
						"foo.step.bar.inputs.tensor.xyz",
					},
				},
				Steps: map[string]*PipelineStep{
					"a": {
						Name: "a",
					},
				},
			},
			err: &PipelineInputErr{"test", "foo.step.bar.inputs.tensor.xyz", pipelineInputTooLongReason},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := checkPipelineInput(test.pipelineVersion)
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

func TestCheckStepOutputs(t *testing.T) {
	g := NewGomegaWithT(t)
	tests := []validateTest{
		{
			name: "valid outputs",
			pipelineVersion: &PipelineVersion{
				Name: "test",
				Steps: map[string]*PipelineStep{
					"a": {
						Name: "a",
					},
				},
				Output: &PipelineOutput{
					Steps: []string{"a.outputs", "a.inputs", "a.outputs.t1"},
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
				},
				Output: &PipelineOutput{
					Steps: []string{"a.sdfs.s"},
				},
			},
			err: &PipelineOutputSpecifierErr{pipeline: "test", specifier: "a.sdfs.s"},
		},
		{
			name: "bad specifier has too many parts",
			pipelineVersion: &PipelineVersion{
				Name: "test",
				Steps: map[string]*PipelineStep{
					"a": {
						Name: "a",
					},
				},
				Output: &PipelineOutput{
					Steps: []string{"a.inputs.t1.x"},
				},
			},
			err: &PipelineOutputSpecifierErr{pipeline: "test", specifier: "a.inputs.t1.x"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := checkPipelineOutputs(test.pipelineVersion)
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

func TestCheckName(t *testing.T) {
	g := NewGomegaWithT(t)
	tests := []validateTest{
		{
			name: "a valid name",
			pipelineVersion: &PipelineVersion{
				Name: "1-name_that-isVa1id0",
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
			name: "a invalid name with dots",
			pipelineVersion: &PipelineVersion{
				Name: "a-name_that-is-not-Valid.10.1",
			},
			err: &PipelineNameValidationErr{pipeline: "a-name_that-is-not-Valid.10.1"},
		},
		{
			name: "a invalid name with a special character",
			pipelineVersion: &PipelineVersion{
				Name: "aNameThatIs%notValid",
			},
			err: &PipelineNameValidationErr{pipeline: "aNameThatIs%notValid"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := checkName(test.pipelineVersion)
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
