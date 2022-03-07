package experiment

import "github.com/seldonio/seldon-core/scheduler/pkg/coordinator"

const (
	experimentStateEventSource = "experiment.state"
)

// Internal methods so they assume locks are in place

func (es *ExperimentStore) addModelReference(modelName string, experiment *Experiment) {
	experiments := es.modelReferences[modelName]
	if experiments == nil {
		experiments = make(map[string]*Experiment)
	}
	experiments[experiment.Name] = experiment
	es.modelReferences[modelName] = experiments
}

func (es *ExperimentStore) removeModelReference(modelName string, experiment *Experiment) {
	experiments := es.modelReferences[modelName]
	if experiments != nil {
		delete(experiments, experiment.Name)
	}
}

func (es *ExperimentStore) addModelReferences(experiment *Experiment) {
	if experiment.Baseline != nil {
		es.addModelReference(experiment.Baseline.ModelName, experiment)
	}
	for _, candidate := range experiment.Candidates {
		es.addModelReference(candidate.ModelName, experiment)
	}
}

func (es *ExperimentStore) removeModelReferences(experiment *Experiment) {
	if experiment.Baseline != nil {
		es.removeModelReference(experiment.Baseline.ModelName, experiment)
	}
	for _, candidate := range experiment.Candidates {
		es.removeModelReference(candidate.ModelName, experiment)
	}
}

func (es *ExperimentStore) getTotalModelReferences() int {
	tot := 0
	for _, refs := range es.modelReferences {
		tot = tot + len(refs)
	}
	return tot
}

func (es *ExperimentStore) cleanExperimentState(experiment *Experiment) {
	existingExperiment := es.experiments[experiment.Name]
	if existingExperiment == nil {
		return
	}
	// if Baseline changed update
	if existingExperiment.Baseline != nil {
		delete(es.baselines, existingExperiment.Baseline.ModelName)
		if (experiment.Baseline != nil && existingExperiment.Baseline.ModelName != experiment.Baseline.ModelName) ||
			experiment.Baseline == nil {
			// Model connected has been changed or removed so need to update it
			if es.eventHub != nil {
				es.eventHub.PublishModelEvent(experimentStateEventSource, coordinator.ModelEventMsg{
					ModelName: existingExperiment.Baseline.ModelName,
					//Empty model version
				})
			}
		}
	}
	es.removeModelReferences(existingExperiment)
}

func (es *ExperimentStore) updateExperimentState(experiment *Experiment) {
	if experiment.Baseline != nil {
		es.baselines[experiment.Baseline.ModelName] = experiment
	}
	es.addModelReferences(experiment)
}
