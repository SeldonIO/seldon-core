package experiment

func (es *ExperimentStore) validateNoExistingDefaultModel(experiment *Experiment) error {
	if experiment.DefaultModel != nil {
		if baselineExperiment, ok := es.baselines[*experiment.DefaultModel]; ok {
			if baselineExperiment.Name != experiment.Name {
				return &ExperimentBaselineExists{modelName: *experiment.DefaultModel, experimentName: experiment.Name}
			}
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
	if experiment.DefaultModel != nil {
		for _, candidate := range experiment.Candidates {
			if candidate.ModelName == *experiment.DefaultModel {
				return nil
			}
		}
		return &ExperimentDefaultModelNotFound{experimentName: experiment.Name, defaultModel: *experiment.DefaultModel}
	}
	return nil
}

func (es *ExperimentStore) validate(experiment *Experiment) error {
	if err := es.validateNoExistingDefaultModel(experiment); err != nil {
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
