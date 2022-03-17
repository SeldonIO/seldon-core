package experiment

import (
	"sync"

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
}

func NewExperimentServer(logger logrus.FieldLogger, eventHub *coordinator.EventHub) *ExperimentStore {

	es := &ExperimentStore{
		logger:          logger.WithField("source", "experimentServer"),
		experiments:     make(map[string]*Experiment),
		baselines:       make(map[string]*Experiment),
		modelReferences: make(map[string]map[string]*Experiment),
		eventHub:        eventHub,
	}

	eventHub.RegisterModelEventHandler(
		modelEventHandlerName,
		pendingSyncsQueueSize,
		es.logger,
		es.handleModelEvents,
	)

	return es
}

// This function will publish experiment events for any model that has changed which is part of an existing experiment
func (es *ExperimentStore) handleModelEvents(event coordinator.ModelEventMsg) {
	logger := es.logger.WithField("func", "handleModelEvents")
	logger.Infof("Received event %s", event.String())
	es.mu.Lock()
	defer es.mu.Unlock()
	refs := es.modelReferences[event.ModelName]
	if len(refs) == 0 {
		logger.Debugf("no experiment set for %s", event.ModelName)
		return
	} else {
		for _, experiment := range refs {
			es.eventHub.PublishExperimentEvent(experimentStartEventSource, coordinator.ExperimentEventMsg{
				ExperimentName: experiment.Name,
			})
		}
	}
}

func (es *ExperimentStore) GetExperimentForBaselineModel(modelName string) *Experiment {
	es.mu.RLock()
	defer es.mu.RUnlock()
	return es.baselines[modelName]
}

func (es *ExperimentStore) SetStatus(experimentName string, active bool, reason string) error {
	es.mu.Lock()
	defer es.mu.Unlock()
	if experiment, ok := es.experiments[experimentName]; !ok {
		return &ExperimentNotFound{experimentName: experimentName}
	} else {
		currentActive := experiment.Active
		experiment.Active = active
		experiment.StatusDescription = reason
		if currentActive != experiment.Active {
			var k8sMeta *coordinator.KubernetesMeta
			if experiment.KubernetesMeta != nil {
				k8sMeta = &coordinator.KubernetesMeta{
					Namespace:  experiment.KubernetesMeta.Namespace,
					Generation: experiment.KubernetesMeta.Generation,
				}
			}
			es.eventHub.PublishExperimentEvent(experimentStartEventSource, coordinator.ExperimentEventMsg{
				ExperimentName: experiment.Name,
				Status: &coordinator.ExperimentEventStatus{
					Active:            experiment.Active,
					StatusDescription: experiment.StatusDescription,
				},
				KubernetesMeta: k8sMeta,
			})
		}
	}
	return nil
}

func (es *ExperimentStore) StartExperiment(experiment *Experiment) error {
	es.mu.Lock()
	defer es.mu.Unlock()
	logger := es.logger.WithField("func", "StartExperiment")
	logger.Infof("Start %s", experiment.Name)
	err := es.validate(experiment)
	if err != nil {
		return err
	}
	es.cleanExperimentState(experiment)
	es.updateExperimentState(experiment)
	es.experiments[experiment.Name] = experiment
	if es.eventHub != nil {
		es.eventHub.PublishExperimentEvent(experimentStartEventSource, coordinator.ExperimentEventMsg{
			ExperimentName: experiment.Name,
		})
	}
	return nil
}

func (es *ExperimentStore) StopExperiment(experimentName string) error {
	logger := es.logger.WithField("func", "StopExperiment")
	logger.Infof("Stop %s", experimentName)
	es.mu.Lock()
	defer es.mu.Unlock()
	if experiment, ok := es.experiments[experimentName]; ok {
		experiment.Deleted = true
		experiment.Active = false
		es.cleanExperimentState(experiment)
		if es.eventHub != nil {
			if experiment.DefaultModel != nil {
				es.eventHub.PublishModelEvent(experimentStateEventSource, coordinator.ModelEventMsg{
					ModelName: *experiment.DefaultModel,
					//Empty model version
				})
			}
			//Update state of model
			var k8sMeta *coordinator.KubernetesMeta
			if experiment.KubernetesMeta != nil {
				k8sMeta = &coordinator.KubernetesMeta{
					Namespace:  experiment.KubernetesMeta.Namespace,
					Generation: experiment.KubernetesMeta.Generation,
				}
			}
			es.eventHub.PublishExperimentEvent(experimentStopEventSource, coordinator.ExperimentEventMsg{
				ExperimentName: experimentName,
				Status: &coordinator.ExperimentEventStatus{
					Active:            experiment.Active,
					StatusDescription: experiment.StatusDescription,
				},
				KubernetesMeta: k8sMeta,
			})
		}
		return nil
	} else {
		return &ExperimentNotFound{
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
