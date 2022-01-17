package store

import "time"

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

func (m *MemoryStore) AddListener(c chan *ModelSnapshot) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.modelEventListeners = append(m.modelEventListeners, c)
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
		} else if (modelVersion.GetModelSpec() != nil && stats.replicasAvailable == modelVersion.GetDeploymentSpec().Replicas) ||
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

func (m *MemoryStore) updateModelStatus(isLatest bool, deleted bool, modelVersion *ModelVersion, prevModelVersion *ModelVersion) {
	stats := calcReplicaStateStatistics(modelVersion, deleted)
	m.logger.Debugf("Stats %+v modelVersion %+v prev model %+v", stats, modelVersion, prevModelVersion)
	updateModelState(isLatest, modelVersion, prevModelVersion, stats, deleted)

	model := m.getModelImpl(modelVersion.GetMeta().GetName())
	if isLatest {
		for _, listener := range m.modelEventListeners {
			listener <- model
		}
	}
}
