package experiment

import (
	"sync"

	"github.com/seldonio/seldon-core/scheduler/pkg/store"

	"github.com/mitchellh/copystructure"
	"github.com/seldonio/seldon-core/scheduler/pkg/coordinator"
	"github.com/sirupsen/logrus"
)

const (
	pendingSyncsQueueSize      int = 100
	experimentStartEventSource     = "experiment.store.start"
	experimentStopEventSource      = "experiment.store.stop"
	modelEventHandlerName          = "experiment.store.models"
)

type ExperimentServer interface {
	StartExperiment(experiment *Experiment) error
	StopExperiment(experimentName string) error
	GetExperiment(experimentName string) (*Experiment, error)
	GetExperiments() ([]*Experiment, error)
	GetExperimentForBaselineModel(modelName string) *Experiment
	SetStatus(experimentName string, active bool, reason string) error
}

type ExperimentStore struct {
	logger          logrus.FieldLogger
	mu              sync.RWMutex
	experiments     map[string]*Experiment
	baselines       map[string]*Experiment            // modelName to the single baseline experiment it appears in
	modelReferences map[string]map[string]*Experiment // modelName to experiments it appears in
	eventHub        *coordinator.EventHub
	store           store.ModelStore
}

func NewExperimentServer(logger logrus.FieldLogger, eventHub *coordinator.EventHub, store store.ModelStore) *ExperimentStore {

	es := &ExperimentStore{
		logger:          logger.WithField("source", "experimentServer"),
		experiments:     make(map[string]*Experiment),
		baselines:       make(map[string]*Experiment),
		modelReferences: make(map[string]map[string]*Experiment),
		eventHub:        eventHub,
		store:           store,
	}

	if eventHub != nil {
		eventHub.RegisterModelEventHandler(
			modelEventHandlerName,
			pendingSyncsQueueSize,
			es.logger,
			es.handleModelEvents,
		)
	}

	return es
}

func (es *ExperimentStore) publishModelEvent(experiment *Experiment) {
	es.eventHub.PublishModelEvent(experimentStateEventSource, coordinator.ModelEventMsg{
		ModelName: *experiment.DefaultModel,
		//Empty model version
	})
}

func (es *ExperimentStore) publishExperimentEvent(experiment *Experiment, source string, updatedExperiment bool) {
	var k8sMeta *coordinator.KubernetesMeta
	if experiment.KubernetesMeta != nil {
		k8sMeta = &coordinator.KubernetesMeta{
			Namespace:  experiment.KubernetesMeta.Namespace,
			Generation: experiment.KubernetesMeta.Generation,
		}
	}
	es.eventHub.PublishExperimentEvent(source, coordinator.ExperimentEventMsg{
		ExperimentName:    experiment.Name,
		UpdatedExperiment: updatedExperiment,
		Status: &coordinator.ExperimentEventStatus{
			Active:            experiment.Active,
			CandidatesReady:   experiment.AreCandidatesReady(),
			MirrorReady:       experiment.IsMirrorReady(),
			StatusDescription: experiment.StatusDescription,
		},
		KubernetesMeta: k8sMeta,
	})
}

func (es *ExperimentStore) createExperimentEventMsg(experiment *Experiment, updatedExperiment bool) *coordinator.ExperimentEventMsg {
	var k8sMeta *coordinator.KubernetesMeta
	if experiment.KubernetesMeta != nil {
		k8sMeta = &coordinator.KubernetesMeta{
			Namespace:  experiment.KubernetesMeta.Namespace,
			Generation: experiment.KubernetesMeta.Generation,
		}
	}
	return &coordinator.ExperimentEventMsg{
		ExperimentName:    experiment.Name,
		UpdatedExperiment: updatedExperiment,
		Status: &coordinator.ExperimentEventStatus{
			Active:            experiment.Active,
			CandidatesReady:   experiment.AreCandidatesReady(),
			MirrorReady:       experiment.IsMirrorReady(),
			StatusDescription: experiment.StatusDescription,
		},
		KubernetesMeta: k8sMeta,
	}
}

// This function will publish experiment events for any model that has changed which is part of an existing experiment
func (es *ExperimentStore) handleModelEvents(event coordinator.ModelEventMsg) {
	logger := es.logger.WithField("func", "handleModelEvents")
	logger.Infof("Received event %s", event.String())

	go func() {
		var updatedExperiments []*Experiment
		es.mu.Lock()
		refs := es.modelReferences[event.ModelName]
		for _, experiment := range refs {
			for _, candidate := range experiment.Candidates {
				if candidate.ModelName == event.ModelName {
					model, err := es.store.GetModel(event.ModelName)
					if err != nil {
						logger.WithError(err).Warnf("Failed to get model %s for candidate check for experiment %s", event.ModelName, experiment.Name)
					} else {
						if model.GetLatest() != nil && model.GetLatest().ModelState().State == store.ModelAvailable {
							candidate.Ready = true
						} else {
							candidate.Ready = false
						}
					}
				}
			}
			if experiment.Mirror != nil && experiment.Mirror.ModelName == event.ModelName {
				model, err := es.store.GetModel(event.ModelName)
				if err != nil {
					logger.WithError(err).Warnf("Failed to get model %s for mirror check for experiment %s", event.ModelName, experiment.Name)
				} else {
					if model.GetLatest() != nil && model.GetLatest().ModelState().State == store.ModelAvailable {
						experiment.Mirror.Ready = true
					} else {
						experiment.Mirror.Ready = false
					}
				}
			}
			copied, err := copystructure.Copy(experiment)
			if err != nil {
				logger.WithError(err).Errorf("Failed to copy experiment %s", experiment.Name)
			} else {
				updatedExperiments = append(updatedExperiments, copied.(*Experiment))
			}
		}
		es.mu.Unlock()
		for _, experiment := range updatedExperiments {
			es.publishExperimentEvent(experiment, experimentStartEventSource, true)
		}
	}()
}

func (es *ExperimentStore) GetExperimentForBaselineModel(modelName string) *Experiment {
	es.mu.RLock()
	defer es.mu.RUnlock()
	return es.baselines[modelName]
}

func (es *ExperimentStore) SetStatus(experimentName string, active bool, reason string) error {
	evt, err := es.setStatusImpl(experimentName, active, reason)
	if err != nil {
		return err
	}
	if es.eventHub != nil && evt != nil {
		es.eventHub.PublishExperimentEvent(experimentStateEventSource, *evt)
	}
	return nil
}

func (es *ExperimentStore) setStatusImpl(experimentName string, active bool, reason string) (*coordinator.ExperimentEventMsg, error) {
	es.mu.Lock()
	defer es.mu.Unlock()
	if experiment, ok := es.experiments[experimentName]; !ok {
		return nil, &ExperimentNotFound{experimentName: experimentName}
	} else {
		if !experiment.Deleted || !active { //can't reactivate a deleted experiment
			currentActive := experiment.Active
			experiment.Active = active
			experiment.StatusDescription = reason
			if currentActive != experiment.Active {
				return es.createExperimentEventMsg(experiment, false), nil
			}
		}
	}
	return nil, nil
}

func (es *ExperimentStore) StartExperiment(experiment *Experiment) error {
	expEvt, modelEvt, err := es.startExperimentImpl(experiment)
	if err != nil {
		return err
	}
	if es.eventHub != nil {
		if modelEvt != nil {
			es.eventHub.PublishModelEvent(experimentStateEventSource, *modelEvt)
		}
		if expEvt != nil {
			es.eventHub.PublishExperimentEvent(experimentStartEventSource, *expEvt)
		}
	}
	return nil
}

func (es *ExperimentStore) startExperimentImpl(experiment *Experiment) (*coordinator.ExperimentEventMsg, *coordinator.ModelEventMsg, error) {
	es.mu.Lock()
	defer es.mu.Unlock()
	var modelEvt *coordinator.ModelEventMsg
	logger := es.logger.WithField("func", "StartExperiment")
	logger.Infof("Start %s", experiment.Name)
	err := es.validate(experiment)
	if err != nil {
		return nil, nil, err
	}
	if es.cleanExperimentState(experiment) {
		modelEvt = &coordinator.ModelEventMsg{
			ModelName: *experiment.DefaultModel,
		}
	}
	es.updateExperimentState(experiment)
	es.experiments[experiment.Name] = experiment
	return es.createExperimentEventMsg(experiment, true), modelEvt, nil
}

func (es *ExperimentStore) StopExperiment(experimentName string) error {
	expEvt, modelEvt, err := es.stopExperimentImpl(experimentName)
	if err != nil {
		return err
	}
	if es.eventHub != nil {
		if modelEvt != nil {
			es.eventHub.PublishModelEvent(experimentStopEventSource, *modelEvt)
		}
		if expEvt != nil {
			es.eventHub.PublishExperimentEvent(experimentStopEventSource, *expEvt)
		}
	}
	return nil
}

func (es *ExperimentStore) stopExperimentImpl(experimentName string) (*coordinator.ExperimentEventMsg, *coordinator.ModelEventMsg, error) {
	logger := es.logger.WithField("func", "StopExperiment")
	logger.Infof("Stop %s", experimentName)
	es.mu.Lock()
	defer es.mu.Unlock()
	if experiment, ok := es.experiments[experimentName]; ok {
		var modelEvt *coordinator.ModelEventMsg
		experiment.Deleted = true
		experiment.Active = false
		es.cleanExperimentState(experiment)
		if experiment.DefaultModel != nil {
			modelEvt = &coordinator.ModelEventMsg{
				ModelName: *experiment.DefaultModel,
			}
		}
		return es.createExperimentEventMsg(experiment, true), modelEvt, nil
	} else {
		return nil, nil, &ExperimentNotFound{
			experimentName: experimentName,
		}
	}
}

func (es *ExperimentStore) GetExperiment(experimentName string) (*Experiment, error) {
	es.mu.RLock()
	defer es.mu.RUnlock()
	if experiment, ok := es.experiments[experimentName]; ok {
		copiedExperiment, err := copystructure.Copy(experiment)
		if err != nil {
			return nil, err
		}
		return copiedExperiment.(*Experiment), nil
	} else {
		return nil, &ExperimentNotFound{
			experimentName: experimentName,
		}
	}
}

func (es *ExperimentStore) GetExperiments() ([]*Experiment, error) {
	es.mu.RLock()
	defer es.mu.RUnlock()

	foundExperiments := []*Experiment{}
	for _, e := range es.experiments {
		if !e.Deleted {
			copied, err := copystructure.Copy(e)
			if err != nil {
				return nil, err
			}

			foundExperiments = append(foundExperiments, copied.(*Experiment))
		}
	}
	return foundExperiments, nil
}
