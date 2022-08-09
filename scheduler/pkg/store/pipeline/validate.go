package pipeline

import (
	"strings"
)

// Step inputs can be reference a previous step name and tensor output/input
// e.g.
//
//	step1 <output from step1>
//	step1.outputs.out1 <out1 named tensor from step1>
//	step1.inputs.in1 <in1 names tensor from step1>
const (
	StepInputSpecifier  = "inputs"
	StepOutputSpecifier = "outputs"
	StepNameSeperator   = "."
)

func validate(pv *PipelineVersion) error {
	if err := checkStepNameNotPipelineName(pv); err != nil {
		return err
	}
	if err := checkStepReferencesExist(pv); err != nil {
		return err
	}
	if err := checkStepInputs(pv); err != nil {
		return err
	}
	if err := checkStepOutputs(pv); err != nil {
		return err
	}
	if err := checkForCycles(pv); err != nil {
		return err
	}
	if err := checkInputsAndTriggersDiffer(pv); err != nil {
		return err
	}
	return nil
}

func checkStepNameNotPipelineName(pv *PipelineVersion) error {
	for _, v := range pv.Steps {
		if v.Name == pv.Name {
			return &PipelineStepNameEqualsPipelineNameErr{pipeline: pv.Name}
		}
	}
	return nil
}

func checkInputsAndTriggersDiffer(pv *PipelineVersion) error {
	for _, v := range pv.Steps {
		inputMap := make(map[string]bool)
		for _, inp := range v.Inputs {
			inputMap[getStepNameFromInput(inp)] = true
		}
		for _, trg := range v.Triggers {
			if _, ok := inputMap[trg]; ok {
				return &PipelineInputAndTriggerErr{pipeline: pv.Name, input: trg}
			}
		}
	}
	return nil
}

func getStepNameFromInput(stepName string) string {
	return strings.Split(stepName, StepNameSeperator)[0]
}

func checkForCyclesFromStep(step *PipelineStep, pv *PipelineVersion, visited map[string]bool) error {
	visited[step.Name] = true
	stepNames := make(map[string]bool)
	for _, inp := range step.Inputs {
		stepName := getStepNameFromInput(inp)
		if stepName != pv.Name {
			stepNames[stepName] = true
		}
	}
	for _, inp := range step.Triggers {
		stepName := getStepNameFromInput(inp)
		if stepName != pv.Name {
			stepNames[stepName] = true
		}
	}
	for stepName := range stepNames {
		if _, ok := visited[stepName]; ok {
			return &PipelineCycleErr{pipeline: pv.Name}
		}
		err := checkForCyclesFromStep(pv.Steps[stepName], pv, visited)
		if err != nil {
			return err
		}
	}
	delete(visited, step.Name)
	return nil
}

func checkForCycles(pv *PipelineVersion) error {
	checked := make(map[string]bool)
	for k, v := range pv.Steps {
		if _, ok := checked[k]; ok {
			continue
		}
		visited := make(map[string]bool)
		err := checkForCyclesFromStep(v, pv, visited)
		if err != nil {
			return err
		}
		for k := range visited {
			checked[k] = true
		}
	}
	return nil
}

func checkStepReferencesExist(pv *PipelineVersion) error {
	for k, v := range pv.Steps {
		for _, inp := range v.Inputs {
			stepName := getStepNameFromInput(inp)
			if _, ok := pv.Steps[stepName]; !ok && stepName != pv.Name {
				return &PipelineStepNotFoundErr{pipeline: pv.Name, step: k, badRef: stepName}
			}
		}
	}
	if pv.Output != nil {
		for _, step := range pv.Output.Steps {
			stepName := getStepNameFromInput(step)
			if _, ok := pv.Steps[stepName]; !ok {
				return &PipelineOutputStepNotFoundErr{pipeline: pv.Name, step: stepName}
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

func checkStepOutputs(pv *PipelineVersion) error {
	if pv.Output != nil {
		for _, v := range pv.Output.Steps {
			parts := strings.Split(v, StepNameSeperator)
			switch len(parts) {
			case 2, 3:
				if !(parts[1] == StepInputSpecifier || parts[1] == StepOutputSpecifier) {
					return &PipelineOutputSpecifierErr{
						pipeline:  pv.Name,
						specifier: v,
					}
				}
			default:
				return &PipelineOutputSpecifierErr{
					pipeline:  pv.Name,
					specifier: v,
				}
			}
		}
	}
	return nil
}
