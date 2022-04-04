package pipeline

import (
	"strings"
)

// Step inputs can be reference a previous step name and tensor output/input
// e.g.
//  step1 <output from step1>
//  step1.outputs.out1 <out1 named tensor from step1>
//  step1.inputs.in1 <in1 names tensor from step1>
const (
	StepInputSpecifier  = "inputs"
	StepOutputSpecifier = "outputs"
	StepNameSeperator   = "."
)

func validate(pv *PipelineVersion) error {
	if err := checkStepReferencesExist(pv); err != nil {
		return err
	}
	if err := checkStepInputs(pv); err != nil {
		return err
	}
	//if err := checkOnlyOneInput(pv); err != nil {
	//	return err
	//}
	return nil
}

func getStepNameFromInput(stepName string) string {
	return strings.Split(stepName, StepNameSeperator)[0]
}

func checkStepReferencesExist(pv *PipelineVersion) error {
	for k, v := range pv.Steps {
		for _, inp := range v.Inputs {
			stepName := getStepNameFromInput(inp)
			if _, ok := pv.Steps[stepName]; !ok {
				return &PipelineStepNotFoundErr{pipeline: pv.Name, step: k, badRef: stepName}
			}
		}
	}
	return nil
}

func checkStepInputs(pv *PipelineVersion) error {
	for _, v := range pv.Steps {
		for _, inp := range v.Inputs {
			parts := strings.Split(inp, StepNameSeperator)
			switch len(parts) {
			case 2, 3:
				if !(parts[1] == StepInputSpecifier || parts[1] == StepOutputSpecifier) {
					return &PipelineStepInputSpecifierErr{
						pipeline:   pv.Name,
						step:       v.Name,
						outputStep: inp,
					}
				}
			default:
				return &PipelineStepInputSpecifierErr{
					pipeline:   pv.Name,
					step:       v.Name,
					outputStep: inp,
				}
			}
		}
	}
	return nil
}

func checkOnlyOneInput(pv *PipelineVersion) error {
	foundInputStep := false
	for _, v := range pv.Steps {
		if len(v.Inputs) == 0 {
			if foundInputStep {
				return &PipelineMultipleInputsErr{pipeline: pv.Name}
			} else {
				foundInputStep = true
			}
		}
	}
	return nil
}
