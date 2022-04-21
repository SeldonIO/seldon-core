package pipeline

import "fmt"

type PipelineStepInputInvalidErr struct {
	pipeline       string
	stepOutputName string
}

func (e *PipelineStepInputInvalidErr) Is(tgt error) bool {
	_, ok := tgt.(*PipelineStepInputInvalidErr)
	return ok
}

func (psi *PipelineStepInputInvalidErr) Error() string {
	return fmt.Sprintf("pipeline step invalid pipeline %s step input [%s]", psi.pipeline, psi.stepOutputName)
}

type PipelineStepInputSpecifierErr struct {
	pipeline   string
	step       string
	outputStep string
}

func (pse *PipelineStepInputSpecifierErr) Error() string {
	return fmt.Sprintf("pipeline step input invalid pipeline %s step %s input step %s.", pse.pipeline, pse.step, pse.outputStep)
}

type PipelineOutputSpecifierdErr struct {
	pipeline  string
	specifier string
}

func (pos *PipelineOutputSpecifierdErr) Error() string {
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
