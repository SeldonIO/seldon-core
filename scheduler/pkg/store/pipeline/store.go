package pipeline

import (
	"os"
	"sync"

	"github.com/mitchellh/copystructure"
	"github.com/seldonio/seldon-core/scheduler/apis/mlops/scheduler"
	"github.com/seldonio/seldon-core/scheduler/pkg/coordinator"
	"github.com/sirupsen/logrus"
)

const (
	addPipelineEventSource       = "pipeline.store.addpipeline"
	removePipelineEventSource    = "pipeline.store.removepipeline"
	setStatusPipelineEventSource = "pipeline.store.setstatus"
)

type PipelineHandler interface {
	AddPipeline(pipeline *scheduler.Pipeline) error
	RemovePipeline(name string) error
	GetPipelineVersion(name string, version uint32, uid string) (*PipelineVersion, error)
	GetPipeline(name string) (*Pipeline, error)
	GetPipelines() ([]*Pipeline, error)
	SetPipelineState(name string, version uint32, uid string, state PipelineStatus, reason string) error
	GetAllRunningPipelineVersions() []coordinator.PipelineEventMsg
}

type PipelineStore struct {
	logger    logrus.FieldLogger
	mu        sync.RWMutex
	eventHub  *coordinator.EventHub
	pipelines map[string]*Pipeline
	db        *PipelineDB
}

func NewPipelineStore(logger logrus.FieldLogger, eventHub *coordinator.EventHub) *PipelineStore {
	ps := &PipelineStore{
		logger:    logger,
		eventHub:  eventHub,
		pipelines: make(map[string]*Pipeline),
		db:        nil,
	}
	return ps
}

func (ps *PipelineStore) InitialiseDB(path string) error {
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return err
	}
	db, err := NewPipelineDb(path)
	if err != nil {
		return err
	}
	ps.db = db
	err = ps.db.restore(ps)
	if err != nil {
		return err
	}
	return nil
}

func (ps *PipelineStore) restorePipeline(pipeline *Pipeline) {
	logger := ps.logger.WithField("func", "addPipeline")
	logger.Infof("Adding pipeline %s with state %s", pipeline.GetLatestPipelineVersion().String(), pipeline.GetLatestPipelineVersion().State.Status.String())
	ps.pipelines[pipeline.Name] = pipeline
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
	logger := ps.logger.WithField("func", "AddPipeline")
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
		lastPipeline := pipeline.GetLatestPipelineVersion()

		switch lastPipeline.State.Status {
		case PipelineTerminate:
			return &PipelineTerminatingErr{pipeline: req.Name}
		case PipelineTerminating, PipelineTerminated:
			pipeline = &Pipeline{
				Name:        req.Name,
				LastVersion: 0,
			}
		default:
			// Handle repeat Kubernetes resource calls for same generation
			if req.GetKubernetesMeta() != nil && lastPipeline.KubernetesMeta != nil {
				if req.GetKubernetesMeta().Generation == lastPipeline.KubernetesMeta.Generation {
					logger.Infof("Pipeline %s kubernetes meta generation matches %d so will ignore", req.Name, req.KubernetesMeta.Generation)
					return nil
				}
			}
		}
	}
	err := validateAndAddPipelineVersion(req, pipeline)
	if err != nil {
		return err
	}
	ps.pipelines[req.Name] = pipeline
	if ps.db != nil {
		err = ps.db.save(pipeline)
		if err != nil {
			return err
		}
	}
	pv := pipeline.GetLatestPipelineVersion()
	if ps.eventHub != nil {
		ps.eventHub.PublishPipelineEvent(addPipelineEventSource, coordinator.PipelineEventMsg{
			PipelineName:    pv.Name,
			PipelineVersion: pv.Version,
			UID:             pv.UID,
		})
	}
	return nil
}

func (ps *PipelineStore) RemovePipeline(name string) error {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	if pipeline, ok := ps.pipelines[name]; ok {
		lastPipelineVersion := pipeline.GetLatestPipelineVersion()
		if lastPipelineVersion == nil {
			return &PipelineVersionNotFoundErr{pipeline: name, version: pipeline.LastVersion - 1}
		}
		lastState := lastPipelineVersion.State
		switch lastState.Status {
		case PipelineTerminate:
			return &PipelineTerminatingErr{pipeline: name}
		case PipelineTerminated:
			return &PipelineAlreadyTerminatedErr{pipeline: name}
		default:
			if ps.db != nil {
				err := ps.db.delete(pipeline)
				if err != nil {
					return err
				}
			}
			pipeline.Deleted = true
			lastPipelineVersion.State.setState(PipelineTerminate, "pipeline removed")
			if ps.eventHub != nil {
				ps.eventHub.PublishPipelineEvent(removePipelineEventSource, coordinator.PipelineEventMsg{
					PipelineName:    lastPipelineVersion.Name,
					PipelineVersion: lastPipelineVersion.Version,
					UID:             lastPipelineVersion.UID,
				})
			}
		}
		return nil
	} else {
		return &PipelineNotFoundErr{pipeline: name}
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
		case PipelineCreate, PipelineCreating, PipelineReady:
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

func (ps *PipelineStore) terminateOldUnterminatedPipelinesIfNeeded(pipeline *Pipeline) {
	for _, pv := range pipeline.Versions {
		if pv.Version != pipeline.LastVersion {
			switch pv.State.Status {
			case PipelineTerminating, PipelineTerminate, PipelineTerminated:
				continue
			default:
				pv.State.setState(PipelineTerminate, "")
				if ps.eventHub != nil {
					ps.eventHub.PublishPipelineEvent(setStatusPipelineEventSource, coordinator.PipelineEventMsg{
						PipelineName:    pv.Name,
						PipelineVersion: pv.Version,
						UID:             pv.UID,
					})
				}
			}
		}
	}
}

func (ps *PipelineStore) SetPipelineState(name string, versionNumber uint32, uid string, status PipelineStatus, reason string) error {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	if pipeline, ok := ps.pipelines[name]; ok {
		if pipelineVersion := pipeline.GetPipelineVersion(versionNumber); pipelineVersion != nil {
			if pipelineVersion.UID == uid {
				pipelineVersion.State.setState(status, reason)
				if ps.eventHub != nil {
					ps.eventHub.PublishPipelineEvent(setStatusPipelineEventSource, coordinator.PipelineEventMsg{
						PipelineName:    pipelineVersion.Name,
						PipelineVersion: pipelineVersion.Version,
						UID:             pipelineVersion.UID,
					})
				}
				if status == PipelineReady {
					ps.terminateOldUnterminatedPipelinesIfNeeded(pipeline)
				}
				if ps.db != nil {
					err := ps.db.save(pipeline)
					if err != nil {
						return err
					}
				}
				return nil
			} else {
				return &PipelineVersionUidMismatchErr{
					pipeline:    name,
					version:     versionNumber,
					uidActual:   pipelineVersion.UID,
					uidExpected: uid,
				}
			}
		} else {
			return &PipelineVersionNotFoundErr{pipeline: name, version: versionNumber}
		}
	} else {
		return &PipelineNotFoundErr{pipeline: name}
	}
}
