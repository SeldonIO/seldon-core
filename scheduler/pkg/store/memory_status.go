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

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/coordinator"
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

func calcModelVersionStatistics(modelVersion *ModelVersion, deleted bool) *modelVersionStateStatistics {
	s := modelVersionStateStatistics{}
	for _, replicaState := range modelVersion.ReplicaState() {
		switch replicaState.State {
		case Available:
			s.replicasAvailable++
		case LoadRequested, Loading, Loaded: // unavailable but OK
			s.replicasLoading++
		case LoadFailed, LoadedUnavailable: // unavailable but not OK
			s.replicasLoadFailed++
			if !deleted && replicaState.Timestamp.After(s.lastFailedStateTime) {
				s.lastFailedStateTime = replicaState.Timestamp
				s.lastFailedReason = replicaState.Reason
			}
		case UnloadEnvoyRequested, UnloadRequested, Unloading:
			s.replicasUnloading++
		case Unloaded:
			s.replicasUnloaded++
		case UnloadFailed:
			s.replicasUnloadFailed++
			if deleted && replicaState.Timestamp.After(s.lastFailedStateTime) {
				s.lastFailedStateTime = replicaState.Timestamp
				s.lastFailedReason = replicaState.Reason
			}
		case Draining:
			s.replicasDraining++
		}
		if replicaState.Timestamp.After(s.latestTime) {
			s.latestTime = replicaState.Timestamp
		}
	}
	return &s
}

func updateModelState(isLatest bool, modelVersion *ModelVersion, prevModelVersion *ModelVersion, stats *modelVersionStateStatistics, deleted bool) {
	var modelState ModelState
	var modelReason string
	modelTimestamp := stats.latestTime
	if deleted || !isLatest {
		if stats.replicasUnloadFailed > 0 {
			modelState = ModelTerminateFailed
			modelReason = stats.lastFailedReason
			modelTimestamp = stats.lastFailedStateTime
		} else if stats.replicasUnloading > 0 || stats.replicasAvailable > 0 || stats.replicasLoading > 0 {
			modelState = ModelTerminating
		} else {
			modelState = ModelTerminated
		}
	} else {
		if stats.replicasLoadFailed > 0 {
			modelState = ModelFailed
			modelReason = stats.lastFailedReason
			modelTimestamp = stats.lastFailedStateTime
		} else if (modelVersion.GetDeploymentSpec() != nil && stats.replicasAvailable == modelVersion.GetDeploymentSpec().Replicas) || // equal to desired replicas
			(modelVersion.GetDeploymentSpec() != nil && stats.replicasAvailable >= modelVersion.GetDeploymentSpec().MinReplicas && modelVersion.GetDeploymentSpec().MinReplicas > 0) || // min replicas is set and available replicas are greater than or equal to min replicas
			(stats.replicasAvailable > 0 && prevModelVersion != nil && modelVersion != prevModelVersion && prevModelVersion.state.State == ModelAvailable) {
			modelState = ModelAvailable
		} else {
			modelState = ModelProgressing
		}
	}
	modelVersion.state = ModelStatus{
		State:               modelState,
		Reason:              modelReason,
		Timestamp:           modelTimestamp,
		AvailableReplicas:   stats.replicasAvailable,
		UnavailableReplicas: stats.replicasLoading + stats.replicasLoadFailed,
		DrainingReplicas:    stats.replicasDraining,
	}
}

func (m *MemoryStore) FailedScheduling(modelID string, version uint32, reason string, reset bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	model, ok := m.store.models[modelID]
	if !ok {
		return fmt.Errorf("model %s not found", modelID)
	}

	// likely the failed model version is the latest, so we loop through in reverse order
	for i := len(model.versions) - 1; i >= 0; i-- {
		modelVersion := model.versions[i]
		
		if modelVersion.version == version {
			// we use len of GetAssignment instead of .state.AvailableReplicas as it is more accurate in this context
			availableReplicas := uint32(len(modelVersion.GetAssignment()))

			modelVersion.state = ModelStatus{
				State:               ScheduleFailed,
				Reason:              reason,
				Timestamp:           time.Now(),
				AvailableReplicas:   availableReplicas,
				UnavailableReplicas: modelVersion.GetModel().GetDeploymentSpec().GetReplicas() - availableReplicas,
			}
			// make sure we reset server but only if there are no available replicas
			if reset {
				modelVersion.SetServer("")
			}

			m.eventHub.PublishModelEvent(
				modelFailureEventSource,
				coordinator.ModelEventMsg{
					ModelName:    modelVersion.GetMeta().GetName(),
					ModelVersion: modelVersion.GetVersion(),
				},
			)

			return nil
		}
	}

	return fmt.Errorf("model %s found, version %d not found", modelID, version)
}

func (m *MemoryStore) updateModelStatus(isLatest bool, deleted bool, modelVersion *ModelVersion, prevModelVersion *ModelVersion) {
	logger := m.logger.WithField("func", "updateModelStatus")
	stats := calcModelVersionStatistics(modelVersion, deleted)
	logger.Debugf("Stats %+v modelVersion %+v prev model %+v", stats, modelVersion, prevModelVersion)

	updateModelState(isLatest, modelVersion, prevModelVersion, stats, deleted)
}
