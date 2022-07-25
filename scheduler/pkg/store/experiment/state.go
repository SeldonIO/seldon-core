package experiment

import (
	"github.com/seldonio/seldon-core/scheduler/pkg/store"
	pipeline2 "github.com/seldonio/seldon-core/scheduler/pkg/store/pipeline"
)

const (
	experimentStateEventSource = "experiment.state"
)

// Internal methods so they assume locks are in place

func (es *ExperimentStore) addReference(name string, experiment *Experiment) {
	switch experiment.ResourceType {
	case PipelineResourceType:
		experiments := es.pipelineReferences[name]
		if experiments == nil {
			experiments = make(map[string]*Experiment)
		}
		experiments[experiment.Name] = experiment
		es.pipelineReferences[name] = experiments
	case ModelResourceType:
		experiments := es.modelReferences[name]
		if experiments == nil {
			experiments = make(map[string]*Experiment)
		}
		experiments[experiment.Name] = experiment
		es.modelReferences[name] = experiments
	}
}

func (es *ExperimentStore) removeReference(name string, experiment *Experiment) {
	switch experiment.ResourceType {
	case PipelineResourceType:
		experiments := es.pipelineReferences[name]
		if experiments != nil {
			delete(experiments, experiment.Name)
		}
	case ModelResourceType:
		experiments := es.modelReferences[name]
		if experiments != nil {
			delete(experiments, experiment.Name)
		}
	}
}

func (es *ExperimentStore) addReferences(experiment *Experiment) {
	for _, candidate := range experiment.Candidates {
		es.addReference(candidate.Name, experiment)
	}
}

func (es *ExperimentStore) removeReferences(experiment *Experiment) {
	for _, candidate := range experiment.Candidates {
		es.removeReference(candidate.Name, experiment)
	}
}

func (es *ExperimentStore) getTotalModelReferences() int {
	tot := 0
	for _, refs := range es.modelReferences {
		tot = tot + len(refs)
	}
	return tot
}

func (es *ExperimentStore) getTotalPipelineReferences() int {
	tot := 0
	for _, refs := range es.pipelineReferences {
		tot = tot + len(refs)
	}
	return tot
}

func (es *ExperimentStore) cleanExperimentState(experiment *Experiment) bool {
	publishEvent := false
	existingExperiment := es.experiments[experiment.Name]
	if existingExperiment == nil {
		return false
	}
	// if Baseline changed update
	if existingExperiment.Default != nil {
		switch existingExperiment.ResourceType {
		case PipelineResourceType:
			delete(es.pipelineBaselines, *existingExperiment.Default)
		case ModelResourceType:
			delete(es.modelBaselines, *existingExperiment.Default)
		}

		if (experiment.Default != nil && *existingExperiment.Default != *experiment.Default) ||
			experiment.Default == nil {
			// Model connected has been changed or removed so need to update it
			publishEvent = true
		}
	}
	es.removeReferences(existingExperiment)
	return publishEvent
}

func (es *ExperimentStore) updateExperimentState(experiment *Experiment) {
	if experiment.Default != nil {
		switch experiment.ResourceType {
		case PipelineResourceType:
			es.pipelineBaselines[*experiment.Default] = experiment
		case ModelResourceType:
			es.modelBaselines[*experiment.Default] = experiment
		}
	}
	es.addReferences(experiment)
	es.setCandidateAndMirrorReadiness(experiment)
}

func (es *ExperimentStore) setCandidateAndMirrorReadiness(experiment *Experiment) {
	logger := es.logger.WithField("func", "setCandidateAndMirrorReadiness")
	switch experiment.ResourceType {
	case PipelineResourceType:
		if es.pipelineStore != nil {
			for _, candidate := range experiment.Candidates {
				pipeline, err := es.pipelineStore.GetPipeline(candidate.Name)
				if err != nil {
					logger.WithError(err).Infof("Failed to get pipeline %s for candidate check for experiment %s", candidate.Name, experiment.Name)
				} else {
					if pipeline.GetLatestPipelineVersion() != nil && pipeline.GetLatestPipelineVersion().State.Status == pipeline2.PipelineReady {
						candidate.Ready = true
					} else {
						candidate.Ready = false
					}
				}
			}
			if experiment.Mirror != nil {
				pipeline, err := es.pipelineStore.GetPipeline(experiment.Mirror.Name)
				if err != nil {
					logger.WithError(err).Warnf("Failed to get pipeline %s for mirror check for experiment %s", experiment.Mirror.Name, experiment.Name)
				} else {
					if pipeline.GetLatestPipelineVersion() != nil && pipeline.GetLatestPipelineVersion().State.Status == pipeline2.PipelineReady {
						experiment.Mirror.Ready = true
					} else {
						experiment.Mirror.Ready = false
					}
				}
			}
		}
	case ModelResourceType:
		if es.store != nil {
			for _, candidate := range experiment.Candidates {
				model, err := es.store.GetModel(candidate.Name)
				if err != nil {
					logger.WithError(err).Infof("Failed to get model %s for candidate check for experiment %s", candidate.Name, experiment.Name)
				} else {
					if model.GetLatest() != nil && model.GetLatest().ModelState().State == store.ModelAvailable {
						candidate.Ready = true
					} else {
						candidate.Ready = false
					}
				}
			}
			if experiment.Mirror != nil {
				model, err := es.store.GetModel(experiment.Mirror.Name)
				if err != nil {
					logger.WithError(err).Warnf("Failed to get model %s for mirror check for experiment %s", experiment.Mirror.Name, experiment.Name)
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
}
