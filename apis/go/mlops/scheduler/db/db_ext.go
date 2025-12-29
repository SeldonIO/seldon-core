package db

import (
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func (m *ModelVersion) GetAssignment() []int {
	var assignment []int
	var draining []int

	for k, v := range m.Replicas {
		if v.Status.State == ModelReplicaState_MODEL_REPLICA_STATE_LOADED ||
			v.Status.State == ModelReplicaState_MODEL_REPLICA_STATE_AVAILABLE ||
			v.Status.State == ModelReplicaState_MODEL_REPLICA_STATE_LOADED_UNAVAILABLE {
			assignment = append(assignment, k)
		}
		if v.Status.State == ModelReplicaState_MODEL_REPLICA_STATE_DRAINING {
			draining = append(draining, k)
		}
	}

	// prefer assignments that are not draining as envoy is eventual consistent
	if len(assignment) > 0 {
		return assignment
	}
	if len(draining) > 0 {
		return draining
	}
	return nil
}

func (m *ModelVersion) SetReplicaState(replicaIdx int, state ModelReplicaState, reason string) {
	m.Replicas[replicaIdx] = &ReplicaStatusEntry{State: state, Timestamp: timestamppb.Now(), Reason: reason}
}

func (m *Model) Latest() *ModelVersion {
	if len(m.Versions) > 0 {
		return m.Versions[len(m.Versions)-1]
	}
	return nil
}

func (m *Model) HasLatest() bool {
	return len(m.Versions) > 0
}

func (m *Model) GetVersion(version uint32) *ModelVersion {
	for _, mv := range m.Versions {
		if mv.GetVersion() == version {
			return mv
		}
	}
	return nil
}

// TODO do we need to consider previous versions?
func (m *Model) Inactive() bool {
	return m.Latest().Inactive()
}

func (m *ModelVersion) Inactive() bool {
	for _, v := range m.Replicas {
		if !v.Status.State.Inactive() {
			return false
		}
	}
	return true
}

func (m ModelReplicaState) CanReceiveTraffic() bool {
	return m == ModelReplicaState_MODEL_REPLICA_STATE_LOADED ||
		m == ModelReplicaState_MODEL_REPLICA_STATE_AVAILABLE ||
		m == ModelReplicaState_MODEL_REPLICA_STATE_LOADED_UNAVAILABLE ||
		m == ModelReplicaState_MODEL_REPLICA_STATE_DRAINING
}

func (m ModelReplicaState) AlreadyLoadingOrLoaded() bool {
	return m == ModelReplicaState_MODEL_REPLICA_STATE_LOADING ||
		m == ModelReplicaState_MODEL_REPLICA_STATE_LOADED ||
		m == ModelReplicaState_MODEL_REPLICA_STATE_AVAILABLE ||
		m == ModelReplicaState_MODEL_REPLICA_STATE_LOADED_UNAVAILABLE
}

func (m ModelReplicaState) UnloadingOrUnloaded() bool {
	return m == ModelReplicaState_MODEL_REPLICA_STATE_UNLOAD_ENVOY_REQUESTED ||
		m == ModelReplicaState_MODEL_REPLICA_STATE_UNLOAD_REQUESTED ||
		m == ModelReplicaState_MODEL_REPLICA_STATE_LOADING ||
		m == ModelReplicaState_MODEL_REPLICA_STATE_UNLOADED ||
		m == ModelReplicaState_MODEL_REPLICA_STATE_UNKNOWN
}

func (m ModelReplicaState) Inactive() bool {
	return m == ModelReplicaState_MODEL_REPLICA_STATE_UNLOADED ||
		m == ModelReplicaState_MODEL_REPLICA_STATE_UNLOAD_FAILED ||
		m == ModelReplicaState_MODEL_REPLICA_STATE_UNKNOWN ||
		m == ModelReplicaState_MODEL_REPLICA_STATE_LOAD_FAILED
}

func (m ModelReplicaState) IsLoadingOrLoaded() bool {
	return m == ModelReplicaState_MODEL_REPLICA_STATE_LOADED ||
		m == ModelReplicaState_MODEL_REPLICA_STATE_LOAD_REQUESTED ||
		m == ModelReplicaState_MODEL_REPLICA_STATE_LOADING ||
		m == ModelReplicaState_MODEL_REPLICA_STATE_AVAILABLE || m == ModelReplicaState_MODEL_REPLICA_STATE_LOADED_UNAVAILABLE
}

//
// Servers
//

func (s *ServerReplica) GetLoadedOrLoadingModelVersions() []*ModelVersionID {
	var models []*ModelVersionID
	for _, model := range s.LoadingModels {
		models = append(models, model.ModelVersionId)
	}
	for _, model := range s.LoadingModels {
		models = append(models, model.ModelVersionId)
	}
	return models
}
