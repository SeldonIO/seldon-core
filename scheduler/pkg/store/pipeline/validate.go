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
	"strings"
)

// Step inputs can be reference a previous step name and tensor output/input
// e.g.
//
//	step1 <output from step1>
//	step1.outputs.out1 <out1 named tensor from step1>
//	step1.inputs.in1 <in1 names tensor from step1>
const (
	StepInputSpecifier    = "inputs"
	StepOutputSpecifier   = "outputs"
	PipelineStepSpecifier = "step"
	StepNameSeperator     = "."
)

func validate(pv *PipelineVersion) error {
	if err := checkStepsExist(pv); err != nil {
		return err
	}
	if err := checkStepNameNotPipelineName(pv); err != nil {
		return err
	}
	if err := checkStepInputs(pv); err != nil {
		return err
	}
	if err := checkStepTriggers(pv); err != nil {
		return err
	}
	if err := checkStepReferencesExist(pv); err != nil {
		return err
	}
	if err := checkPipelineOutputs(pv); err != nil {
		return err
	}
	if err := checkForCycles(pv); err != nil {
		return err
	}
	if err := checkInputsAndTriggersDiffer(pv); err != nil {
		return err
	}
	if err := checkPipelineInput(pv); err != nil {
		return err
	}
	if err := checkStepFilterPercent(pv); err != nil {
		return err
	}
	return nil
}

func checkStepFilterPercent(pv *PipelineVersion) error {
	for _, v := range pv.Steps {
		if v.FilterPercent < 0 || v.FilterPercent > 100 {
			return &PipelineStepFilterErr{pipeline: pv.Name, step: v.Name, filterPercent: v.FilterPercent}
		}
	}
	return nil
}

func checkStepsExist(pv *PipelineVersion) error {
	if len(pv.Steps) == 0 {
		return &PipelineStepsEmptyErr{pipeline: pv.Name}
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
			if strings.TrimSpace(inp) == "" || strings.Index(inp, StepNameSeperator) == 0 {
				return &PipelineStepInputEmptyErr{pv.Name, v.Name, false}
			}
			parts := strings.Split(inp, StepNameSeperator)
			switch len(parts) {
			case 2, 3:
				if !(parts[1] == StepInputSpecifier || parts[1] == StepOutputSpecifier) {
					return &PipelineStepInputSpecifierErr{
						pipeline:   pv.Name,
						step:       v.Name,
						outputStep: inp,
						isTrigger:  false,
					}
				}
			default:
				return &PipelineStepInputSpecifierErr{
					pipeline:   pv.Name,
					step:       v.Name,
					outputStep: inp,
					isTrigger:  false,
				}
			}
		}
	}
	return nil
}

func checkStepTriggers(pv *PipelineVersion) error {
	for _, v := range pv.Steps {
		for _, inp := range v.Triggers {
			if strings.TrimSpace(inp) == "" || strings.Index(inp, StepNameSeperator) == 0 {
				return &PipelineStepInputEmptyErr{pv.Name, v.Name, true}
			}
			parts := strings.Split(inp, StepNameSeperator)
			switch len(parts) {
			case 2, 3:
				if !(parts[1] == StepInputSpecifier || parts[1] == StepOutputSpecifier) {
					return &PipelineStepInputSpecifierErr{
						pipeline:   pv.Name,
						step:       v.Name,
						outputStep: inp,
						isTrigger:  true,
					}
				}
			default:
				return &PipelineStepInputSpecifierErr{
					pipeline:   pv.Name,
					step:       v.Name,
					outputStep: inp,
					isTrigger:  true,
				}
			}
		}
	}
	return nil
}

const (
	pipelineInputEmptyErr               = "At least one pipeline input must be specified"
	pipelineInputEmptyErrReason         = "Input name must not be empty"
	pipelineInputOnlyPipelineNameReason = "A Pipeline name must also specify one of inputs, outputs or a step name"
	pipelineInputInvalidPrefixReason    = "A Pipeline inputs referencing another pipeline must be <pipeineName>.(inputs|outputs|step.<stepName>)"
	pipelineInputStepBadSuffix          = "A pipeline step must be <pipelineName>.step.<stepName>.(inputs|outputs)"
	pipelineInputTooLongReason          = "The input is too long. It must be <pipelineName>.(inputs|outputs).(tensorName)? or <pipelineName>.step.<stepName>.<inputs|outputs>.<tensorName>"
)

func checkPipelineInput(pv *PipelineVersion) error {
	if pv.Input != nil {
		if len(pv.Input.ExternalInputs) == 0 {
			return &PipelineInputErr{pv.Name, "", pipelineInputEmptyErr}
		}
		for _, v := range pv.Input.ExternalInputs {
			if strings.TrimSpace(v) == "" {
				return &PipelineInputErr{pv.Name, v, pipelineInputEmptyErrReason}
			}
			parts := strings.Split(v, StepNameSeperator)
			switch len(parts) {
			case 1:
				return &PipelineInputErr{pv.Name, v, pipelineInputOnlyPipelineNameReason}
			case 2:
				if !(parts[1] == StepInputSpecifier || parts[1] == StepOutputSpecifier) {
					return &PipelineInputErr{pv.Name, v, pipelineInputInvalidPrefixReason}
				}
			case 3:
				if !(parts[1] == StepInputSpecifier || parts[1] == StepOutputSpecifier || parts[1] == PipelineStepSpecifier) {
					return &PipelineInputErr{pv.Name, v, pipelineInputInvalidPrefixReason}
				}
				if parts[1] == PipelineStepSpecifier {
					return &PipelineInputErr{pv.Name, v, pipelineInputStepBadSuffix}
				}
			default:
				if !(parts[1] == StepInputSpecifier || parts[1] == StepOutputSpecifier || parts[1] == PipelineStepSpecifier) {
					return &PipelineInputErr{pv.Name, v, pipelineInputInvalidPrefixReason}
				}
				if !(parts[3] == StepInputSpecifier || parts[3] == StepOutputSpecifier) {
					return &PipelineInputErr{pv.Name, v, pipelineInputStepBadSuffix}
				}
				if parts[1] == StepInputSpecifier || parts[1] == StepOutputSpecifier {
					return &PipelineInputErr{pv.Name, v, pipelineInputInvalidPrefixReason}
				}
				if len(parts) > 5 {
					return &PipelineInputErr{pv.Name, v, pipelineInputTooLongReason}
				}
			}
		}
	}
	return nil
}

func checkPipelineOutputs(pv *PipelineVersion) error {
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
