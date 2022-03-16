package experiment

import "fmt"

type ExperimentNotFound struct {
	experimentName string
}

func (enf *ExperimentNotFound) Is(tgt error) bool {
	_, ok := tgt.(*ExperimentNotFound)
	return ok
}

func (enf *ExperimentNotFound) Error() string {
	return fmt.Sprintf("Experiment not found %s", enf.experimentName)
}

type ExperimentBaselineExists struct {
	experimentName string
	modelName      string
}

func (ebe *ExperimentBaselineExists) Error() string {
	return fmt.Sprintf("Model %s already in experiment %s as a baseline. A model can only appear in one experiment as a baseline model", ebe.modelName, ebe.experimentName)
}

type ExperimentNoCandidates struct {
	experimentName string
}

func (enc *ExperimentNoCandidates) Error() string {
	return fmt.Sprintf("experiment %s has no candidates", enc.experimentName)
}
