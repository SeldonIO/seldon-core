package experiment

import (
	"github.com/seldonio/seldon-core/scheduler/pkg/coordinator"
	"github.com/seldonio/seldon-core/scheduler/pkg/store"
)

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
	for _, candidate := range experiment.Candidates {
		es.addModelReference(candidate.ModelName, experiment)
	}
}

func (es *ExperimentStore) removeModelReferences(experiment *Experiment) {
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
	if existingExperiment.DefaultModel != nil {
		delete(es.baselines, *existingExperiment.DefaultModel)
		if (experiment.DefaultModel != nil && *existingExperiment.DefaultModel != *experiment.DefaultModel) ||
			experiment.DefaultModel == nil {
			// Model connected has been changed or removed so need to update it
			if es.eventHub != nil {
				es.eventHub.PublishModelEvent(experimentStateEventSource, coordinator.ModelEventMsg{
					ModelName: *existingExperiment.DefaultModel,
					//Empty model version
				})
			}
		}
	}
	es.removeModelReferences(existingExperiment)
}

func (es *ExperimentStore) updateExperimentState(experiment *Experiment) {
	if experiment.DefaultModel != nil {
		es.baselines[*experiment.DefaultModel] = experiment
	}
	es.addModelReferences(experiment)
	es.setCandidateAndMirrorReadiness(experiment)
}

func (es *ExperimentStore) setCandidateAndMirrorReadiness(experiment *Experiment) {
	logger := es.logger.WithField("func", "setCandidateAndMirrorReadiness")
	if es.store != nil {
		for _, candidate := range experiment.Candidates {
			model, err := es.store.GetModel(candidate.ModelName)
			if err != nil {
				logger.WithError(err).Infof("Failed to get model %s for candidate check for experiment %s", candidate.ModelName, experiment.Name)
			} else {
				if model.GetLatest() != nil && model.GetLatest().ModelState().State == store.ModelAvailable {
					candidate.Ready = true
				} else {
					candidate.Ready = false
				}
			}
		}
		if experiment.Mirror != nil {
			model, err := es.store.GetModel(experiment.Mirror.ModelName)
			if err != nil {
				logger.WithError(err).Warnf("Failed to get model %s for mirror check for experiment %s", experiment.Mirror.ModelName, experiment.Name)
			} else {
				if model.GetLatest() != nil && model.GetLatest().ModelState().State == store.ModelAvailable {
					experiment.Mirror.Ready = true
				} else {
					experiment.Mirror.Ready = false
				}
			}
		}
	}
}
