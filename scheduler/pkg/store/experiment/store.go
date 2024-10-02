/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package experiment

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/mitchellh/copystructure"
	"github.com/sirupsen/logrus"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/coordinator"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store/pipeline"
)

const (
	deletedExperimentTTL       time.Duration = time.Duration(time.Hour * 24)
	pendingSyncsQueueSize      int           = 1000
	experimentStartEventSource               = "experiment.store.start"
	experimentStopEventSource                = "experiment.store.stop"
	modelEventHandlerName                    = "experiment.store.models"
	pipelineEventHandlerName                 = "experiment.store.pipelines"
	experimentDbFolder                       = "experimentdb"
)

type ExperimentServer interface {
	StartExperiment(experiment *Experiment) error
	StopExperiment(experimentName string) error
	GetExperiment(experimentName string) (*Experiment, error)
	GetExperiments() ([]*Experiment, error)
	GetExperimentForBaselineModel(modelName string) *Experiment
	GetExperimentForBaselinePipeline(pipelineName string) *Experiment
	SetStatus(experimentName string, active bool, reason string) error
}

type ExperimentStore struct {
	logger             logrus.FieldLogger
	mu                 sync.RWMutex
	experiments        map[string]*Experiment
	modelBaselines     map[string]*Experiment            // model name to the single baseline experiment it appears in
	modelReferences    map[string]map[string]*Experiment // model name to experiments it appears in
	pipelineBaselines  map[string]*Experiment            // pipeline name to the single baseline experiment it appears in
	pipelineReferences map[string]map[string]*Experiment // pipeline name to experiments it appears in
	eventHub           *coordinator.EventHub
	store              store.ModelStore
	pipelineStore      pipeline.PipelineHandler
	db                 *ExperimentDBManager
}

func NewExperimentServer(logger logrus.FieldLogger, eventHub *coordinator.EventHub, store store.ModelStore, pipelineStore pipeline.PipelineHandler) *ExperimentStore {
	es := &ExperimentStore{
		logger:             logger.WithField("source", "experimentServer"),
		experiments:        make(map[string]*Experiment),
		modelBaselines:     make(map[string]*Experiment),
		modelReferences:    make(map[string]map[string]*Experiment),
		pipelineBaselines:  make(map[string]*Experiment),
		pipelineReferences: make(map[string]map[string]*Experiment),
		eventHub:           eventHub,
		store:              store,
		pipelineStore:      pipelineStore,
	}

	if eventHub != nil {
		eventHub.RegisterModelEventHandler(
			modelEventHandlerName,
			pendingSyncsQueueSize,
			es.logger,
			es.handleModelEvents,
		)
		eventHub.RegisterPipelineEventHandler(
			pipelineEventHandlerName,
			pendingSyncsQueueSize,
			es.logger,
			es.handlePipelineEvents,
		)
	}

	return es
}

func getExperimentDbFolder(basePath string) string {
	return filepath.Join(basePath, experimentDbFolder)
}

// we just add a reference to the experiment in the memory store
// so that we can keep track of it in case we need to replay the event (to the controller)
// we do not trigger an event though as envoy has a clean state when the scheduler restarts
func (es *ExperimentStore) AddExperimentInMap(experiment *Experiment) error {
	es.mu.Lock()
	defer es.mu.Unlock()
	if _, ok := es.experiments[experiment.Name]; !ok {
		es.experiments[experiment.Name] = experiment
		return nil
	} else {
		return fmt.Errorf("Experiment %s already exists", experiment.Name)
	}
}

func (es *ExperimentStore) InitialiseOrRestoreDB(path string) error {
	logger := es.logger.WithField("func", "initialiseDB")
	experimentDbPath := getExperimentDbFolder(path)
	logger.Infof("Initialise DB at %s", experimentDbPath)
	err := os.MkdirAll(experimentDbPath, os.ModePerm)
	if err != nil {
		return err
	}
	db, err := newExperimentDbManager(experimentDbPath, es.logger)
	if err != nil {
		return err
	}
	es.db = db
	// If database already existed we can restore else this is a noop
	err = es.db.restore(es.StartExperiment, es.AddExperimentInMap)
	if err != nil {
		return err
	}
	return nil
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
	logger.Debugf("Received event %s", event.String())

	go func() {
		var updatedExperiments []*Experiment
		es.mu.Lock()
		refs := es.modelReferences[event.ModelName]
		for _, experiment := range refs {
			for _, candidate := range experiment.Candidates {
				if candidate.Name == event.ModelName {
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
			if experiment.Mirror != nil {
				if experiment.Mirror.Name == event.ModelName {
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

// This function will publish experiment events for any pipeline that has changed which is part of an existing experiment
func (es *ExperimentStore) handlePipelineEvents(event coordinator.PipelineEventMsg) {
	logger := es.logger.WithField("func", "handlePipelineEvents")
	logger.Debugf("Received event %s", event.String())

	go func() {
		var updatedExperiments []*Experiment
		es.mu.Lock()
		refs := es.pipelineReferences[event.PipelineName]
		for _, experiment := range refs {
			for _, candidate := range experiment.Candidates {
				if candidate.Name == event.PipelineName {
					p, err := es.pipelineStore.GetPipeline(event.PipelineName)
					if err != nil {
						logger.WithError(err).Warnf("Failed to get pipeline %s for candidate check for experiment %s", event.PipelineName, experiment.Name)
					} else {
						if p.GetLatestPipelineVersion() != nil && p.GetLatestPipelineVersion().State.Status == pipeline.PipelineReady {
							candidate.Ready = true
						} else {
							candidate.Ready = false
						}
					}
				}
			}
			if experiment.Mirror != nil {
				if experiment.Mirror.Name == event.PipelineName {
					p, err := es.pipelineStore.GetPipeline(event.PipelineName)
					if err != nil {
						logger.WithError(err).Warnf("Failed to get pipeline %s for mirror check for experiment %s", event.PipelineName, experiment.Name)
					} else {
						if p.GetLatestPipelineVersion() != nil && p.GetLatestPipelineVersion().State.Status == pipeline.PipelineReady {
							experiment.Mirror.Ready = true
						} else {
							experiment.Mirror.Ready = false
						}
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
	return es.modelBaselines[modelName]
}

func (es *ExperimentStore) GetExperimentForBaselinePipeline(pipelineName string) *Experiment {
	es.mu.RLock()
	defer es.mu.RUnlock()
	return es.pipelineBaselines[pipelineName]
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
		if !experiment.Deleted || !active { // can't reactivate a deleted experiment
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
	expEvt, modelEvt, pipelineEvt, err := es.startExperimentImpl(experiment)
	if err != nil {
		return err
	}
	if es.eventHub != nil {
		if modelEvt != nil {
			es.eventHub.PublishModelEvent(experimentStateEventSource, *modelEvt)
		}
		if pipelineEvt != nil {
			es.eventHub.PublishPipelineEvent(experimentStartEventSource, *pipelineEvt)
		}
		if expEvt != nil {
			es.eventHub.PublishExperimentEvent(experimentStartEventSource, *expEvt)
		}
	}
	return nil
}

func (es *ExperimentStore) startExperimentImpl(experiment *Experiment) (*coordinator.ExperimentEventMsg, *coordinator.ModelEventMsg, *coordinator.PipelineEventMsg, error) {
	es.mu.Lock()
	defer es.mu.Unlock()
	var modelEvt *coordinator.ModelEventMsg
	var pipelineEvt *coordinator.PipelineEventMsg
	logger := es.logger.WithField("func", "StartExperiment")
	logger.Infof("Start %s", experiment.Name)
	err := es.validate(experiment)
	if err != nil {
		return nil, nil, nil, err
	}
	if resourceName := es.cleanExperimentState(experiment); resourceName != nil {
		switch experiment.ResourceType {
		case PipelineResourceType:
			pipelineEvt = &coordinator.PipelineEventMsg{
				PipelineName:     *resourceName,
				ExperimentUpdate: true,
			}
		case ModelResourceType:
			modelEvt = &coordinator.ModelEventMsg{
				ModelName: *resourceName,
			}
		default:
			return nil, nil, nil, fmt.Errorf("unknown resource type %v", experiment.ResourceType)
		}
	}
	es.updateExperimentState(experiment)
	if es.db != nil {
		err := es.db.save(experiment, nil)
		if err != nil {
			return nil, nil, nil, err
		}
	}
	es.experiments[experiment.Name] = experiment
	return es.createExperimentEventMsg(experiment, true), modelEvt, pipelineEvt, nil
}

func (es *ExperimentStore) StopExperiment(experimentName string) error {
	expEvt, modelEvt, pipelineEvt, err := es.stopExperimentImpl(experimentName)
	if err != nil {
		return err
	}
	if es.eventHub != nil {
		if modelEvt != nil {
			es.eventHub.PublishModelEvent(experimentStopEventSource, *modelEvt)
		}
		if pipelineEvt != nil {
			es.eventHub.PublishPipelineEvent(experimentStopEventSource, *pipelineEvt)
		}
		if expEvt != nil {
			es.eventHub.PublishExperimentEvent(experimentStopEventSource, *expEvt)
		}
	}
	return nil
}

func (es *ExperimentStore) stopExperimentImpl(experimentName string) (*coordinator.ExperimentEventMsg, *coordinator.ModelEventMsg, *coordinator.PipelineEventMsg, error) {
	logger := es.logger.WithField("func", "StopExperiment")
	logger.Infof("Stop %s", experimentName)
	es.mu.Lock()
	defer es.mu.Unlock()
	if experiment, ok := es.experiments[experimentName]; ok {
		var modelEvt *coordinator.ModelEventMsg
		var pipelineEvt *coordinator.PipelineEventMsg
		experiment.Deleted = true
		experiment.Active = false
		es.cleanExperimentState(experiment)
		if experiment.Default != nil {
			switch experiment.ResourceType {
			case PipelineResourceType:
				pipelineEvt = &coordinator.PipelineEventMsg{
					PipelineName:     *experiment.Default,
					ExperimentUpdate: true,
				}
			case ModelResourceType:
				modelEvt = &coordinator.ModelEventMsg{
					ModelName: *experiment.Default,
				}
			default:
				return nil, nil, nil, fmt.Errorf("unknown resource type %v", experiment.ResourceType)
			}
		}
		if es.db != nil {
			ttl := deletedExperimentTTL
			err := es.db.save(experiment, &ttl)
			if err != nil {
				return nil, nil, nil, err
			}
		}
		return es.createExperimentEventMsg(experiment, true), modelEvt, pipelineEvt, nil
	} else {
		return nil, nil, nil, &ExperimentNotFound{
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
		copied, err := copystructure.Copy(e)
		if err != nil {
			return nil, err
		}

		foundExperiments = append(foundExperiments, copied.(*Experiment))
	}
	return foundExperiments, nil
}
