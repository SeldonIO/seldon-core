/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package pipeline

import "fmt"

type PipelineStepInputEmptyErr struct {
	pipeline  string
	stepName  string
	isTrigger bool
}

func (psi *PipelineStepInputEmptyErr) Error() string {
	if psi.isTrigger {
		return fmt.Sprintf("pipeline %s step %s has an empty trigger", psi.pipeline, psi.stepName)
	} else {
		return fmt.Sprintf("pipeline %s step %s has an empty input", psi.pipeline, psi.stepName)
	}
}

type PipelineStepInputSpecifierErr struct {
	pipeline   string
	step       string
	outputStep string
	isTrigger  bool
}

func (pse *PipelineStepInputSpecifierErr) Error() string {
	if pse.isTrigger {
		return fmt.Sprintf("pipeline step trigger invalid pipeline %s step %s input step %s.", pse.pipeline, pse.step, pse.outputStep)
	} else {
		return fmt.Sprintf("pipeline step input invalid pipeline %s step %s input step %s.", pse.pipeline, pse.step, pse.outputStep)
	}
}

type PipelineOutputSpecifierErr struct {
	pipeline  string
	specifier string
}

func (pos *PipelineOutputSpecifierErr) Error() string {
	return fmt.Sprintf("pipeline %s output specifier %s invalid", pos.pipeline, pos.specifier)
}

type PipelineNotFoundErr struct {
	pipeline string
}

func (pnf *PipelineNotFoundErr) Error() string {
	return fmt.Sprintf("pipeline %s not found", pnf.pipeline)
}

type PipelineTerminatingErr struct {
	pipeline string
}

func (pnf *PipelineTerminatingErr) Error() string {
	return fmt.Sprintf("pipeline %s still terminating", pnf.pipeline)
}

type PipelineAlreadyTerminatedErr struct {
	pipeline string
}

func (pnf *PipelineAlreadyTerminatedErr) Error() string {
	return fmt.Sprintf("pipeline %s not found", pnf.pipeline)
}

type PipelineVersionNotFoundErr struct {
	pipeline string
	version  uint32
}

func (pve *PipelineVersionNotFoundErr) Error() string {
	return fmt.Sprintf("pipeline version %s:%d not found", pve.pipeline, pve.version)
}

type PipelineVersionUidMismatchErr struct {
	pipeline    string
	version     uint32
	uidActual   string
	uidExpected string
}

func (pvm *PipelineVersionUidMismatchErr) Error() string {
	return fmt.Sprintf("pipeline version uid mismatch %s:%d expected %s found %s", pvm.pipeline, pvm.version, pvm.uidExpected, pvm.uidActual)
}

type PipelineStepsEmptyErr struct {
	pipeline string
}

func (psee *PipelineStepsEmptyErr) Error() string {
	return fmt.Sprintf("pipeline %s has no steps defined", psee.pipeline)
}

type PipelineStepNotFoundErr struct {
	pipeline string
	step     string
	badRef   string
}

func (psnf *PipelineStepNotFoundErr) Error() string {
	return fmt.Sprintf("pipeline %s step %s has input %s for step that does not exist", psnf.pipeline, psnf.step, psnf.badRef)
}

type PipelineMultipleInputsErr struct {
	pipeline string
}

func (pms *PipelineMultipleInputsErr) Error() string {
	return fmt.Sprintf("pipeline %s must have a single input", pms.pipeline)
}

type PipelineOutputRequiredErr struct {
	pipeline string
}

func (por *PipelineOutputRequiredErr) Error() string {
	return fmt.Sprintf("pipeline %s must have an output", por.pipeline)
}

type PipelineOutputEmptyErr struct {
	pipeline string
}

func (por *PipelineOutputEmptyErr) Error() string {
	return fmt.Sprintf("pipeline %s must have a non empty output", por.pipeline)
}

type PipelineOutputStepNotFoundErr struct {
	pipeline string
	step     string
}

func (por *PipelineOutputStepNotFoundErr) Error() string {
	return fmt.Sprintf("pipeline %s output step %s not found", por.pipeline, por.step)
}

type PipelineMultiStepNoOutput struct {
	pipeline string
}

func (por *PipelineMultiStepNoOutput) Error() string {
	return fmt.Sprintf("multi step pipeline %s must specify output", por.pipeline)
}

type PipelineStepRepeatedErr struct {
	pipeline string
	step     string
}

func (psr *PipelineStepRepeatedErr) Error() string {
	return fmt.Sprintf("pipeline %s has repeated step %s", psr.pipeline, psr.step)
}

type PipelineCycleErr struct {
	pipeline string
}

func (psr *PipelineCycleErr) Error() string {
	return fmt.Sprintf("pipeline %s has a cycle", psr.pipeline)
}

type PipelineInputAndTriggerErr struct {
	pipeline string
	input    string
}

func (psr *PipelineInputAndTriggerErr) Error() string {
	return fmt.Sprintf("pipeline %s : inputs and triggers must differ, but found %s in both", psr.pipeline, psr.input)
}

type PipelineStepNameEqualsPipelineNameErr struct {
	pipeline string
}

func (psr *PipelineStepNameEqualsPipelineNameErr) Error() string {
	return fmt.Sprintf("pipeline %s must not have a step name with the same name as pipeline name", psr.pipeline)
}

type PipelineInputErr struct {
	pipeline string
	input    string
	reason   string
}

func (pie *PipelineInputErr) Error() string {
	return fmt.Sprintf("pipeline %s input %s is invalid. %s", pie.pipeline, pie.input, pie.reason)
}

type PipelineNameValidationErr struct {
	pipeline string
}

func (pnve *PipelineNameValidationErr) Error() string {
	return fmt.Sprintf("pipeline %s does not have a valid name - it must be alphanmumeric and can contain underscores and hyphens", pnve.pipeline)
}
