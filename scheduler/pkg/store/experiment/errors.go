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
	name           string
}

func (ebe *ExperimentBaselineExists) Error() string {
	return fmt.Sprintf("Resource %s already in experiment %s as a baseline. A model or pipeline can only appear in one experiment as a baseline", ebe.name, ebe.experimentName)
}

type ExperimentNoCandidates struct {
	experimentName string
}

func (enc *ExperimentNoCandidates) Error() string {
	return fmt.Sprintf("experiment %s has no candidates", enc.experimentName)
}

type ExperimentDefaultNotFound struct {
	experimentName  string
	defaultResource string
}

func (enc *ExperimentDefaultNotFound) Is(tgt error) bool {
	_, ok := tgt.(*ExperimentDefaultNotFound)
	return ok
}

func (enc *ExperimentDefaultNotFound) Error() string {
	return fmt.Sprintf("default model/pipeline %s not found in experiment %s candidates", enc.defaultResource, enc.experimentName)
}
