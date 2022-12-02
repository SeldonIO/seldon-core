/*
Copyright 2022 Seldon Technologies Ltd.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package store

import (
	"time"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/coordinator"
)

const (
	modelFailureEventSource = "memory.status.scheduling.failed"
	modelUpdateEventSource  = "memory.status.model.update"
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
			(stats.replicasAvailable > 0 && prevModelVersion != nil && modelVersion != prevModelVersion && prevModelVersion.state.State == ModelAvailable) { //TODO In future check if available replicas is > minReplicas
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

func (m *MemoryStore) FailedScheduling(modelVersion *ModelVersion, reason string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	modelVersion.state = ModelStatus{
		State:               ScheduleFailed,
		Reason:              reason,
		Timestamp:           time.Now(),
		AvailableReplicas:   modelVersion.state.AvailableReplicas,
		UnavailableReplicas: modelVersion.GetModel().GetDeploymentSpec().GetReplicas() - modelVersion.state.AvailableReplicas,
	}

	m.eventHub.PublishModelEvent(
		modelFailureEventSource,
		coordinator.ModelEventMsg{
			ModelName:    modelVersion.GetMeta().GetName(),
			ModelVersion: modelVersion.GetVersion(),
		},
	)
}

func (m *MemoryStore) updateModelStatus(isLatest bool, deleted bool, modelVersion *ModelVersion, prevModelVersion *ModelVersion) {
	logger := m.logger.WithField("func", "updateModelStatus")
	stats := calcModelVersionStatistics(modelVersion, deleted)
	logger.Debugf("Stats %+v modelVersion %+v prev model %+v", stats, modelVersion, prevModelVersion)

	updateModelState(isLatest, modelVersion, prevModelVersion, stats, deleted)
}
