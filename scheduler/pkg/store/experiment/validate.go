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

package experiment

import "fmt"

func (es *ExperimentStore) validateNoExistingDefault(experiment *Experiment) error {
	if experiment.Default != nil {
		switch experiment.ResourceType {
		case PipelineResourceType:
			if baselineExperiment, ok := es.pipelineBaselines[*experiment.Default]; ok {
				if baselineExperiment.Name != experiment.Name {
					return &ExperimentBaselineExists{name: *experiment.Default, experimentName: experiment.Name}
				}
			}
		case ModelResourceType:
			if baselineExperiment, ok := es.modelBaselines[*experiment.Default]; ok {
				if baselineExperiment.Name != experiment.Name {
					return &ExperimentBaselineExists{name: *experiment.Default, experimentName: experiment.Name}
				}
			}
		default:
			return fmt.Errorf("Unknown resource type %v", experiment.ResourceType)
		}
	}
	return nil
}

func validateHasCandidateOrMirror(experiment *Experiment) error {
	if len(experiment.Candidates) == 0 && experiment.Mirror == nil {
		return &ExperimentNoCandidatesOrMirrors{experimentName: experiment.Name}
	}
	return nil
}

func validateDefaultModelIsCandidate(experiment *Experiment) error {
	if experiment.Default != nil {
		for _, candidate := range experiment.Candidates {
			if candidate.Name == *experiment.Default {
				return nil
			}
		}
		return &ExperimentDefaultNotFound{experimentName: experiment.Name, defaultResource: *experiment.Default}
	}
	return nil
}

func validateNoDuplicateNames(experiment *Experiment) error {
	names := map[string]bool{}
	for _, candidate := range experiment.Candidates {
		if _, ok := names[candidate.Name]; ok {
			return &ExperimentNoDuplicates{experimentName: experiment.Name, resource: candidate.Name}
		}
		names[candidate.Name] = true
	}
	if experiment.Mirror != nil {
		if _, ok := names[experiment.Mirror.Name]; ok {
			return &ExperimentNoDuplicates{experimentName: experiment.Name, resource: experiment.Mirror.Name}
		}
		names[experiment.Mirror.Name] = true
	}
	return nil
}

func (es *ExperimentStore) validate(experiment *Experiment) error {
	if err := es.validateNoExistingDefault(experiment); err != nil {
		return err
	}
	if err := validateDefaultModelIsCandidate(experiment); err != nil {
		return err
	}
	if err := validateHasCandidateOrMirror(experiment); err != nil {
		return err
	}
	if err := validateNoDuplicateNames(experiment); err != nil {
		return err
	}
	return nil
}
