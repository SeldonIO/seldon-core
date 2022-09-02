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
