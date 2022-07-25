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

func validateHasCandidates(experiment *Experiment) error {
	if len(experiment.Candidates) == 0 {
		return &ExperimentNoCandidates{experimentName: experiment.Name}
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

func (es *ExperimentStore) validate(experiment *Experiment) error {
	if err := es.validateNoExistingDefault(experiment); err != nil {
		return err
	}
	if err := validateDefaultModelIsCandidate(experiment); err != nil {
		return err
	}
	if err := validateHasCandidates(experiment); err != nil {
		return err
	}
	return nil
}
