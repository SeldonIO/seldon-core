/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package store

import (
	"fmt"
	"time"

	"github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler/db"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/coordinator"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	modelFailureEventSource = "memory.status.scheduling.failed"
	modelUpdateEventSource  = "memory.status.model.update"
	serverUpdateEventSource = "memory.status.server.update"
)

type modelVersionStateStatistics struct {
	replicasAvailable    uint32
	replicasLoading      uint32
	replicasLoadFailed   uint32
	replicasUnloading    uint32
	replicasUnloaded     uint32
	replicasUnloadFailed uint32
	replicasDraining     uint32
	lastFailedStateTime  time.Time
	latestTime           time.Time
	lastFailedReason     string
}

func calcModelVersionStatistics(modelVersion *db.ModelVersion, deleted bool) *modelVersionStateStatistics {
	s := modelVersionStateStatistics{}

	for _, replicaState := range modelVersion.Replicas {
		switch replicaState.State {
		case db.ModelReplicaState_MODEL_REPLICA_STATE_AVAILABLE:
			s.replicasAvailable++
		case db.ModelReplicaState_MODEL_REPLICA_STATE_LOAD_REQUESTED,
			db.ModelReplicaState_MODEL_REPLICA_STATE_LOADING, db.ModelReplicaState_MODEL_REPLICA_STATE_LOADED: // unavailable but OK
			s.replicasLoading++
		case db.ModelReplicaState_MODEL_REPLICA_STATE_LOAD_FAILED, db.ModelReplicaState_MODEL_REPLICA_STATE_LOADED_UNAVAILABLE: // unavailable but not OK
			s.replicasLoadFailed++
			if !deleted && replicaState.Timestamp.AsTime().After(s.lastFailedStateTime) {
				s.lastFailedStateTime = replicaState.Timestamp.AsTime()
				s.lastFailedReason = replicaState.Reason
			}
		case db.ModelReplicaState_MODEL_REPLICA_STATE_UNLOAD_ENVOY_REQUESTED,
			db.ModelReplicaState_MODEL_REPLICA_STATE_UNLOAD_REQUESTED, db.ModelReplicaState_MODEL_REPLICA_STATE_UNLOADING:
			s.replicasUnloading++
		case db.ModelReplicaState_MODEL_REPLICA_STATE_UNLOADED:
			s.replicasUnloaded++
		case db.ModelReplicaState_MODEL_REPLICA_STATE_UNLOAD_FAILED:
			s.replicasUnloadFailed++
			if deleted && replicaState.Timestamp.AsTime().After(s.lastFailedStateTime) {
				s.lastFailedStateTime = replicaState.Timestamp.AsTime()
				s.lastFailedReason = replicaState.Reason
			}
		case db.ModelReplicaState_MODEL_REPLICA_STATE_DRAINING:
			s.replicasDraining++
		}
		if replicaState.Timestamp.AsTime().After(s.latestTime) {
			s.latestTime = replicaState.Timestamp.AsTime()
		}
	}
	return &s
}

func updateModelState(isLatest bool, modelVersion *db.ModelVersion, prevModelVersion *db.ModelVersion, stats *modelVersionStateStatistics, deleted bool) {
	var modelState db.ModelState
	var modelReason string

	modelTimestamp := stats.latestTime
	if deleted || !isLatest {
		if stats.replicasUnloadFailed > 0 {
			modelState = db.ModelState_MODEL_STATE_TERMINATED
			modelReason = stats.lastFailedReason
			modelTimestamp = stats.lastFailedStateTime
		} else if stats.replicasUnloading > 0 || stats.replicasAvailable > 0 || stats.replicasLoading > 0 {
			modelState = db.ModelState_MODEL_STATE_TERMINATING
		} else {
			modelState = db.ModelState_MODEL_STATE_TERMINATED
		}
	} else {
		if stats.replicasLoadFailed > 0 {
			modelState = db.ModelState_MODEL_STATE_FAILED
			modelReason = stats.lastFailedReason
			modelTimestamp = stats.lastFailedStateTime
		} else if modelVersion.ModelDefn.DeploymentSpec != nil && stats.replicasAvailable == 0 &&
			modelVersion.ModelDefn.DeploymentSpec.Replicas == 0 && modelVersion.ModelDefn.DeploymentSpec.MinReplicas == 0 {
			modelState = db.ModelState_MODEL_STATE_SCALED_DOWN
		} else if (modelVersion.ModelDefn.DeploymentSpec != nil &&
			stats.replicasAvailable == modelVersion.ModelDefn.DeploymentSpec.Replicas) || // equal to desired replicas
			(modelVersion.ModelDefn.DeploymentSpec != nil && stats.replicasAvailable >= modelVersion.ModelDefn.DeploymentSpec.MinReplicas &&
				modelVersion.ModelDefn.DeploymentSpec.MinReplicas > 0) || // min replicas is set and available replicas are greater than or equal to min replicas
			(stats.replicasAvailable > 0 && prevModelVersion != nil && modelVersion != prevModelVersion &&
				prevModelVersion.State.State == db.ModelState_MODEL_STATE_AVAILABLE) {
			modelState = db.ModelState_MODEL_STATE_AVAILABLE
		} else {
			modelState = db.ModelState_MODEL_STATE_PROGRESSING
		}
	}

	modelVersion.State = &db.ModelStatus{
		State:               modelState,
		ModelGwState:        modelVersion.State.ModelGwState,
		Reason:              modelReason,
		ModelGwReason:       modelVersion.State.ModelGwReason,
		Timestamp:           timestamppb.New(modelTimestamp),
		AvailableReplicas:   stats.replicasAvailable,
		UnavailableReplicas: stats.replicasLoading + stats.replicasLoadFailed,
		DrainingReplicas:    stats.replicasDraining,
	}
}

func (m *MemoryStore) FailedScheduling(modelID string, version uint32, reason string, reset bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	model, err := m.store.models.Get(modelID)
	if err != nil {
		return fmt.Errorf("model %s not found: %w", modelID, err)
	}

	// likely the failed model version is the latest, so we loop through in reverse order
	for i := len(model.Versions) - 1; i >= 0; i-- {
		modelVersion := model.Versions[i]

		if modelVersion.Version == version {
			// we use len of GetAssignment instead of .state.AvailableReplicas as it is more accurate in this context
			availableReplicas := uint32(len(modelVersion.GetAssignment()))
			modelVersion.State = &db.ModelStatus{
				State:               db.ModelState_MODEL_STATE_SCHEDULE_FAILED,
				ModelGwState:        modelVersion.State.ModelGwState,
				Reason:              reason,
				ModelGwReason:       modelVersion.State.ModelGwReason,
				Timestamp:           timestamppb.Now(),
				AvailableReplicas:   availableReplicas,
				UnavailableReplicas: modelVersion.ModelDefn.GetDeploymentSpec().GetReplicas() - availableReplicas,
			}
			// make sure we reset server but only if there are no available replicas
			if reset {
				modelVersion.Server = ""
			}

			if err := m.store.models.Update(model); err != nil {
				return fmt.Errorf("failed to update model %s: %w", modelID, err)
			}

			m.eventHub.PublishModelEvent(
				modelFailureEventSource,
				coordinator.ModelEventMsg{
					ModelName:    modelVersion.ModelDefn.Meta.Name,
					ModelVersion: modelVersion.GetVersion(),
				},
			)

			return nil
		}
	}

	return fmt.Errorf("model %s found, version %d not found", modelID, version)
}

func (m *MemoryStore) updateModelStatus(isLatest bool, deleted bool, modelVersion *db.ModelVersion, prevModelVersion *db.ModelVersion) {
	logger := m.logger.WithField("func", "updateModelStatus")
	stats := calcModelVersionStatistics(modelVersion, deleted)
	logger.Debugf("Stats %+v modelVersion %+v prev model %+v", stats, modelVersion, prevModelVersion)

	updateModelState(isLatest, modelVersion, prevModelVersion, stats, deleted)
}

func (m *MemoryStore) setModelGwStatusToTerminate(isLatest bool, modelVersion *db.ModelVersion) {
	if !isLatest {
		modelVersion.State.ModelGwState = db.ModelState_MODEL_STATE_TERMINATED
		modelVersion.State.ModelGwReason = "Not latest version"
		return
	}
	modelVersion.State.ModelGwState = db.ModelState_MODEL_STATE_TERMINATE
	modelVersion.State.ModelGwReason = "Model deleted"
}

func (m *MemoryStore) UnloadModelGwVersionModels(modelKey string, version uint32) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	model, err := m.store.models.Get(modelKey)
	if err != nil {
		return false, fmt.Errorf("failed to find model %s: %w", modelKey, err)
	}

	modelVersion := model.GetVersion(version)
	if modelVersion == nil {
		return false, fmt.Errorf("version not found for model %s, version %d", modelKey, version)
	}

	m.setModelGwStatusToTerminate(false, modelVersion)

	if err := m.store.models.Update(model); err != nil {
		return false, fmt.Errorf("failed to update model %s: %w", modelKey, err)
	}

	return true, nil
}
