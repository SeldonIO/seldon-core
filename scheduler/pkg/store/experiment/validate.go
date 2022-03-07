package experiment

func (es *ExperimentStore) validateNoExistingBaseline(experiment *Experiment) error {
	if experiment.Baseline != nil {
		if _, ok := es.baselines[experiment.Baseline.ModelName]; ok {
			return &ExperimentBaselineExists{modelName: experiment.Baseline.ModelName, experimentName: experiment.Name}
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

func (es *ExperimentStore) validate(experiment *Experiment) error {
	if err := es.validateNoExistingBaseline(experiment); err != nil {
		return err
	}
	if err := validateHasCandidates(experiment); err != nil {
		return err
	}
	return nil
}
