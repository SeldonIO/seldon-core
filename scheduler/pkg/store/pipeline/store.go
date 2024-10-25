/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package pipeline

import (
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/mitchellh/copystructure"
	"github.com/sirupsen/logrus"

	"github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/coordinator"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store/utils"
)

const (
	pendingSyncsQueueSize             int = 1000
	addPipelineEventSource                = "pipeline.store.addpipeline"
	removePipelineEventSource             = "pipeline.store.removepipeline"
	setStatusPipelineEventSource          = "pipeline.store.setstatus"
	SetModelStatusPipelineEventSource     = "pipeline.store.setmodelstatus"
	pipelineDbFolder                      = "pipelinedb"
	modelEventHandlerName                 = "pipeline.store.models"
)

type PipelineHandler interface {
	AddPipeline(pipeline *scheduler.Pipeline) error
	RemovePipeline(name string) error
	GetPipelineVersion(name string, version uint32, uid string) (*PipelineVersion, error)
	GetPipeline(name string) (*Pipeline, error)
	GetPipelines() ([]*Pipeline, error)
	SetPipelineState(name string, version uint32, uid string, state PipelineStatus, reason string, source string) error
	GetAllRunningPipelineVersions() []coordinator.PipelineEventMsg
}

type PipelineStore struct {
	logger             logrus.FieldLogger
	mu                 sync.RWMutex
	eventHub           *coordinator.EventHub
	pipelines          map[string]*Pipeline
	db                 *PipelineDBManager
	modelStatusHandler ModelStatusHandler
}

func NewPipelineStore(logger logrus.FieldLogger, eventHub *coordinator.EventHub, store store.ModelStore) *PipelineStore {
	ps := &PipelineStore{
		logger:    logger.WithField("source", "pipelineStore"),
		eventHub:  eventHub,
		pipelines: make(map[string]*Pipeline),
		db:        nil,
		modelStatusHandler: ModelStatusHandler{
			logger:          logger.WithField("source", "PipelineModelStatusHandler"),
			store:           store,
			modelReferences: map[string]map[string]void{},
		},
	}
	if eventHub != nil {
		eventHub.RegisterModelEventHandler(
			modelEventHandlerName,
			pendingSyncsQueueSize,
			ps.logger,
			ps.handleModelEvents,
		)
	}
	return ps
}

func getPipelineDbFolder(basePath string) string {
	return filepath.Join(basePath, pipelineDbFolder)
}

func (ps *PipelineStore) InitialiseOrRestoreDB(path string, deletedResourceTTL uint) error {
	logger := ps.logger.WithField("func", "initialiseDB")
	pipelineDbPath := getPipelineDbFolder(path)
	logger.Infof("Initialise DB at %s", pipelineDbPath)
	err := os.MkdirAll(pipelineDbPath, os.ModePerm)
	if err != nil {
		return err
	}
	db, err := newPipelineDbManager(pipelineDbPath, ps.logger, deletedResourceTTL)
	if err != nil {
		return err
	}
	ps.db = db
	// If database already existed we can restore else this is a noop
	err = ps.db.restore(ps.restorePipeline)
	if err != nil {
		return err
	}

	go func() {
		ticker := time.NewTicker(utils.DeletedResourceCleanupFrequency)
		for range ticker.C {
			ps.cleanupDeletedPipelines()
		}
	}()

	return nil
}

// note: we do not validate the pipeline when we restore it from the db as we assume it was validated when it was added
func (ps *PipelineStore) restorePipeline(pipeline *Pipeline) {
	logger := ps.logger.WithField("func", "restorePipeline")
	logger.Infof("Adding pipeline %s with state %s", pipeline.GetLatestPipelineVersion().String(), pipeline.GetLatestPipelineVersion().State.Status.String())
	ps.mu.Lock()
	err := ps.modelStatusHandler.addPipelineModelStatus(pipeline)
	if err != nil {
		logger.WithError(err).Errorf("Failed to set pipeline state for pipeline %s", pipeline.Name)
	}
	ps.pipelines[pipeline.Name] = pipeline
	ps.mu.Unlock()
	pv := pipeline.GetLatestPipelineVersion()
	if ps.eventHub != nil {
		ps.eventHub.PublishPipelineEvent(addPipelineEventSource, coordinator.PipelineEventMsg{
			PipelineName:    pv.Name,
			PipelineVersion: pv.Version,
			UID:             pv.UID,
		})
	}
}

func validateAndAddPipelineVersion(req *scheduler.Pipeline, pipeline *Pipeline) error {
	pv, err := CreatePipelineVersionFromProto(req)
	if err != nil {
		return err
	}
	pv.Version = pipeline.LastVersion + 1
	err = validate(pv)
	if err != nil {
		return err
	}
	pv.State.setState(PipelineCreate, "")
	pipeline.LastVersion = pipeline.LastVersion + 1
	pipeline.Versions = append(pipeline.Versions, pv)
	return nil
}

func (ps *PipelineStore) AddPipeline(req *scheduler.Pipeline) error {
	evt, err := ps.addPipelineImpl(req)
	if err != nil {
		return err
	}
	if ps.eventHub != nil && evt != nil {
		ps.eventHub.PublishPipelineEvent(addPipelineEventSource, *evt)
	}
	return nil
}

func (ps *PipelineStore) addPipelineImpl(req *scheduler.Pipeline) (*coordinator.PipelineEventMsg, error) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	var pipeline *Pipeline
	var ok bool
	if pipeline, ok = ps.pipelines[req.Name]; !ok {
		pipeline = &Pipeline{
			Name:        req.Name,
			LastVersion: 0,
		}
	} else {
		latestPipeline := pipeline.GetLatestPipelineVersion()

		switch latestPipeline.State.Status {
		case PipelineTerminate, PipelineTerminating, PipelineTerminated:
			pipeline = &Pipeline{
				Name:        req.Name,
				LastVersion: 0,
			}
		default:
			// Handle repeat Kubernetes resource calls for same generation
			if ps.generationMatches(req, latestPipeline) {
				return nil, nil
			}
		}
	}
	err := validateAndAddPipelineVersion(req, pipeline)
	if err != nil {
		return nil, err
	}
	err = ps.modelStatusHandler.addPipelineModelStatus(pipeline)
	if err != nil {
		return nil, err
	}
	ps.pipelines[req.Name] = pipeline
	if ps.db != nil {
		err = ps.db.save(pipeline)
		if err != nil {
			return nil, err
		}
	}
	pv := pipeline.GetLatestPipelineVersion()
	return &coordinator.PipelineEventMsg{
		PipelineName:    pv.Name,
		PipelineVersion: pv.Version,
		UID:             pv.UID,
	}, nil
}

func (ps *PipelineStore) generationMatches(req *scheduler.Pipeline, lastPipeline *PipelineVersion) bool {
	logger := ps.logger.WithField("func", "generationMatches")
	if req.GetKubernetesMeta() != nil &&
		lastPipeline.KubernetesMeta != nil &&
		req.GetKubernetesMeta().Generation > 0 &&
		lastPipeline.KubernetesMeta.Generation > 0 {
		if req.GetKubernetesMeta().Generation == lastPipeline.KubernetesMeta.Generation {
			logger.Infof("Pipeline %s kubernetes meta generation matches %d so will ignore", req.Name, req.KubernetesMeta.Generation)
			return true
		}
	}
	return false
}

func (ps *PipelineStore) RemovePipeline(name string) error {
	logger := ps.logger.WithField("func", "RemovePipeline")
	logger.Debugf("Attempt to remove pipeline %s", name)
	evt, err := ps.removePipelineImpl(name)
	if err != nil {
		return err
	}
	if ps.eventHub != nil && evt != nil {
		ps.eventHub.PublishPipelineEvent(removePipelineEventSource, *evt)
	}
	return nil
}

func (ps *PipelineStore) removePipelineImpl(name string) (*coordinator.PipelineEventMsg, error) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	if pipeline, ok := ps.pipelines[name]; ok {
		lastPipelineVersion := pipeline.GetLatestPipelineVersion()
		if lastPipelineVersion == nil {
			return nil, &PipelineVersionNotFoundErr{pipeline: name, version: pipeline.LastVersion - 1}
		}
		lastState := lastPipelineVersion.State
		switch lastState.Status {
		case PipelineTerminating:
			return nil, &PipelineTerminatingErr{pipeline: name}
		case PipelineTerminated:
			return nil, &PipelineAlreadyTerminatedErr{pipeline: name}
		default:
			pipeline.Deleted = true
			pipeline.DeletedAt = time.Now()
			lastPipelineVersion.State.setState(PipelineTerminate, "pipeline removed")
			if err := ps.db.save(pipeline); err != nil {
				ps.logger.WithError(err).Errorf("Failed to save pipeline %s", name)
				return nil, err
			}
			return &coordinator.PipelineEventMsg{
				PipelineName:    lastPipelineVersion.Name,
				PipelineVersion: lastPipelineVersion.Version,
				UID:             lastPipelineVersion.UID,
			}, nil
		}
	} else {
		return nil, &PipelineNotFoundErr{pipeline: name}
	}
}

func (ps *PipelineStore) GetPipelineVersion(name string, versionNumber uint32, uid string) (*PipelineVersion, error) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	if pipeline, ok := ps.pipelines[name]; ok {
		if pipelineVersion := pipeline.GetPipelineVersion(versionNumber); pipelineVersion != nil {
			if pipelineVersion.UID == uid {
				copiedPipelineVersion, err := copystructure.Copy(pipelineVersion)
				if err != nil {
					return nil, err
				}
				return copiedPipelineVersion.(*PipelineVersion), nil
			} else {
				return nil, &PipelineVersionUidMismatchErr{pipeline: name, version: versionNumber, uidActual: pipelineVersion.UID, uidExpected: uid}
			}
		} else {
			return nil, &PipelineVersionNotFoundErr{pipeline: name, version: versionNumber}
		}
	} else {
		return nil, &PipelineNotFoundErr{pipeline: name}
	}
}

func (ps *PipelineStore) GetPipeline(name string) (*Pipeline, error) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	if pipeline, ok := ps.pipelines[name]; ok {
		copiedPipeline, err := copystructure.Copy(pipeline)
		if err != nil {
			return nil, err
		}
		return copiedPipeline.(*Pipeline), nil
	} else {
		return nil, &PipelineNotFoundErr{pipeline: name}
	}
}

func (ps *PipelineStore) GetAllRunningPipelineVersions() []coordinator.PipelineEventMsg {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	var events []coordinator.PipelineEventMsg
	for _, p := range ps.pipelines {
		pv := p.GetLatestPipelineVersion()
		switch pv.State.Status {
		// we consider PipelineTerminating as running as it is still active
		case PipelineCreate, PipelineCreating, PipelineReady, PipelineTerminating:
			events = append(events, coordinator.PipelineEventMsg{
				PipelineName:    pv.Name,
				PipelineVersion: pv.Version,
				UID:             pv.UID,
			})
		}
	}
	return events
}

func (ps *PipelineStore) GetPipelines() ([]*Pipeline, error) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	foundPipelines := []*Pipeline{}
	for _, p := range ps.pipelines {
		copied, err := copystructure.Copy(p)
		if err != nil {
			return nil, err
		}
		foundPipelines = append(foundPipelines, copied.(*Pipeline))
	}

	return foundPipelines, nil
}

func (ps *PipelineStore) terminateOldUnterminatedPipelinesIfNeeded(pipeline *Pipeline) []*coordinator.PipelineEventMsg {
	var evts []*coordinator.PipelineEventMsg
	for _, pv := range pipeline.Versions {
		if pv.Version != pipeline.LastVersion {
			switch pv.State.Status {
			case PipelineTerminating, PipelineTerminate, PipelineTerminated:
				continue
			default:
				pv.State.setState(PipelineTerminate, "")
				evts = append(evts, &coordinator.PipelineEventMsg{
					PipelineName:    pv.Name,
					PipelineVersion: pv.Version,
					UID:             pv.UID,
				})
			}
		}
	}
	return evts
}

func (ps *PipelineStore) SetPipelineState(name string, versionNumber uint32, uid string, status PipelineStatus, reason string, source string) error {
	logger := ps.logger.WithField("func", "SetPipelineState")
	logger.Debugf("Attempt to set state on pipeline %s:%d status:%s", name, versionNumber, status.String())
	evts, err := ps.setPipelineStateImpl(name, versionNumber, uid, status, reason, source)
	if err != nil {
		return err
	}
	if ps.eventHub != nil {
		for _, evt := range evts {
			ps.eventHub.PublishPipelineEvent(setStatusPipelineEventSource, *evt)
		}
	}
	return nil
}

func (ps *PipelineStore) setPipelineStateImpl(name string, versionNumber uint32, uid string, status PipelineStatus, reason, source string) ([]*coordinator.PipelineEventMsg, error) {
	var evts []*coordinator.PipelineEventMsg
	ps.mu.Lock()
	defer ps.mu.Unlock()
	if pipeline, ok := ps.pipelines[name]; ok {
		if pipelineVersion := pipeline.GetPipelineVersion(versionNumber); pipelineVersion != nil {
			if pipelineVersion.UID == uid {
				pipelineVersion.State.setState(status, reason)
				evts = append(evts, &coordinator.PipelineEventMsg{
					PipelineName:    pipelineVersion.Name,
					PipelineVersion: pipelineVersion.Version,
					UID:             pipelineVersion.UID,
					Source:          source,
				})
				if status == PipelineReady {
					// note that we are not setting the source for these events as we do not want to discard them in the chainer service
					// i.e in handlePipelineEvent we want to process the termination events even though they are triggered by the chainer
					// to set the status of the new version of the pipeline to ready
					evts = append(evts, ps.terminateOldUnterminatedPipelinesIfNeeded(pipeline)...)
				}
				if !pipeline.Deleted && ps.db != nil {
					err := ps.db.save(pipeline)
					if err != nil {
						return evts, err
					}
				}
				return evts, nil
			} else {
				return evts, &PipelineVersionUidMismatchErr{
					pipeline:    name,
					version:     versionNumber,
					uidActual:   pipelineVersion.UID,
					uidExpected: uid,
				}
			}
		} else {
			return evts, &PipelineVersionNotFoundErr{pipeline: name, version: versionNumber}
		}
	} else {
		return evts, &PipelineNotFoundErr{pipeline: name}
	}
}

func (ps *PipelineStore) handleModelEvents(event coordinator.ModelEventMsg) {
	logger := ps.logger.WithField("func", "handleModelEvents")
	logger.Infof("Received event %s", event.String())

	go func() {
		ps.modelStatusHandler.mu.RLock()
		defer ps.modelStatusHandler.mu.RUnlock()
		refs := ps.modelStatusHandler.modelReferences[event.ModelName]
		if len(refs) > 0 {
			model, err := ps.modelStatusHandler.store.GetModel(event.ModelName)
			if err != nil {
				logger.Warningf("Failed to get model %s from store", event.ModelName)
				return
			}
			ps.mu.Lock()
			evts := updatePipelinesFromModelAvailability(refs, event.ModelName, model != nil && model.GetLastAvailableModel() != nil, ps.pipelines, ps.logger)
			ps.mu.Unlock()
			// Publish events for modified pipelines
			if ps.eventHub != nil {
				for _, evt := range evts {
					ps.eventHub.PublishPipelineEvent(SetModelStatusPipelineEventSource, *evt)
				}
			}
		} else {
			logger.Debugf("No references in pipelines for model %s", event.ModelName)
		}
	}()
}

func (ps *PipelineStore) cleanupDeletedPipelines() {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.logger.Info("cleaning up deleted pipelines")
	for _, pipeline := range ps.pipelines {
		if pipeline.Deleted {
			if pipeline.DeletedAt.IsZero() {
				pipeline.DeletedAt = time.Now()
				if ps.db != nil {
					err := ps.db.save(pipeline)
					if err != nil {
						ps.logger.Warnf("could not update DB TTL for pipeline: %s", pipeline.Name)
					}
				}
			} else if pipeline.DeletedAt.Add(utils.DeletedResourceTTL).Before(time.Now()) {
				delete(ps.pipelines, pipeline.Name)
			}
		}
	}
}
