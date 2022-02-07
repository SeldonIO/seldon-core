package store

import (
	"time"

	"github.com/seldonio/seldon-core/scheduler/pkg/coordinator"
)

type replicaStateStatistics struct {
	replicasAvailable    uint32
	replicasLoading      uint32
	replicasLoadFailed   uint32
	replicasUnloading    uint32
	replicasUnloaded     uint32
	replicasUnloadFailed uint32
	lastFailedStateTime  time.Time
	latestTime           time.Time
	lastFailedReason     string
}

func calcReplicaStateStatistics(modelVersion *ModelVersion, deleted bool) *replicaStateStatistics {
	s := replicaStateStatistics{}
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
		case UnloadRequested, Unloading:
			s.replicasUnloading++
		case Unloaded:
			s.replicasUnloaded++
		case UnloadFailed:
			s.replicasUnloadFailed++
			if deleted && replicaState.Timestamp.After(s.lastFailedStateTime) {
				s.lastFailedStateTime = replicaState.Timestamp
				s.lastFailedReason = replicaState.Reason
			}
		}
		if replicaState.Timestamp.After(s.latestTime) {
			s.latestTime = replicaState.Timestamp
		}
	}
	return &s
}

func updateModelState(isLatest bool, modelVersion *ModelVersion, prevModelVersion *ModelVersion, stats *replicaStateStatistics, deleted bool) {
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
	}
}

func (m *MemoryStore) FailedScheduling(modelVersion *ModelVersion, reason string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	modelVersion.state = ModelStatus{
		State:               ScheduleFailed,
		Reason:              reason,
		Timestamp:           time.Now(),
		AvailableReplicas:   0,
		UnavailableReplicas: modelVersion.GetModel().GetDeploymentSpec().GetReplicas(),
	}
	go m.eventHub.TriggerModelEvent(coordinator.ModelEventMsg{ModelName: modelVersion.GetMeta().GetName(), ModelVersion: modelVersion.GetVersion()})
}

func (m *MemoryStore) updateModelStatus(isLatest bool, deleted bool, modelVersion *ModelVersion, prevModelVersion *ModelVersion) {
	logger := m.logger.WithField("func", "updateModelStatus")
	stats := calcReplicaStateStatistics(modelVersion, deleted)
	logger.Debugf("Stats %+v modelVersion %+v prev model %+v", stats, modelVersion, prevModelVersion)
	updateModelState(isLatest, modelVersion, prevModelVersion, stats, deleted)

	logger.Debugf("Trigger event for model %s:%d", modelVersion.GetMeta().GetName(), modelVersion.GetVersion())
	go m.eventHub.TriggerModelEvent(coordinator.ModelEventMsg{ModelName: modelVersion.GetMeta().GetName(), ModelVersion: modelVersion.GetVersion()})
}
