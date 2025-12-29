package db

import (
	pb "github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (m *ModelVersion) GetAssignment() []int {
	var assignment []int
	var draining []int

	for k, v := range m.Replicas {
		if v.State == ModelReplicaState_MODEL_REPLICA_STATE_LOADED ||
			v.State == ModelReplicaState_MODEL_REPLICA_STATE_AVAILABLE ||
			v.State == ModelReplicaState_MODEL_REPLICA_STATE_LOADED_UNAVAILABLE {
			assignment = append(assignment, int(k))
		}
		if v.State == ModelReplicaState_MODEL_REPLICA_STATE_DRAINING {
			draining = append(draining, int(k))
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

func (m *ModelVersion) DesiredReplicas() int {
	return int(m.ModelDefn.DeploymentSpec.Replicas)
}

func (m *ModelVersion) GetReplicaForState(state ModelReplicaState) []int {
	var assignment []int
	for k, v := range m.Replicas {
		if v.State == state {
			assignment = append(assignment, int(k))
		}
	}
	return assignment
}

func (m *ModelVersion) HasServer() bool {
	return m.Server != ""
}

func (m *ModelVersion) GetRequiredMemory() uint64 {
	var multiplier uint64 = 1
	if m.ModelDefn != nil && m.ModelDefn.ModelSpec != nil &&
		m.ModelDefn.ModelSpec.ModelRuntimeInfo != nil &&
		m.ModelDefn.ModelSpec.ModelRuntimeInfo.ModelRuntimeInfo != nil {
		multiplier = getInstanceCount(m.ModelDefn.GetModelSpec().ModelRuntimeInfo)
	}
	return m.ModelDefn.GetModelSpec().GetMemoryBytes() * multiplier
}

func getInstanceCount(modelRuntimeInfo *pb.ModelRuntimeInfo) uint64 {
	switch modelRuntimeInfo.ModelRuntimeInfo.(type) {
	case *pb.ModelRuntimeInfo_Mlserver:
		return uint64(modelRuntimeInfo.GetMlserver().ParallelWorkers)
	case *pb.ModelRuntimeInfo_Triton:
		return uint64(modelRuntimeInfo.GetTriton().Cpu[0].InstanceCount)
	default:
		return 1
	}
}

func (m *ModelVersion) SetReplicaState(replicaIdx int, state ModelReplicaState, reason string) {
	m.Replicas[int32(replicaIdx)] = &ReplicaStatus{State: state, Timestamp: timestamppb.Now(), Reason: reason}
}

func (m *ModelVersion) UpdateRuntimeInfo(runtimeInfo *pb.ModelRuntimeInfo) {
	if m.ModelDefn.ModelSpec != nil && m.ModelDefn.ModelSpec.ModelRuntimeInfo == nil && runtimeInfo != nil {
		m.ModelDefn.ModelSpec.ModelRuntimeInfo = runtimeInfo
	}
}

func (m *ModelVersion) GetModelReplicaState(replicaIdx int) ModelReplicaState {
	state, ok := m.Replicas[int32(replicaIdx)]
	if !ok {
		return ModelReplicaState_MODEL_REPLICA_STATE_UNKNOWN
	}
	return state.State
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

func (m *Model) getLastAvailableModelVersionIdx() int {
	lastAvailableIdx := -1
	for idx, mv := range m.Versions {
		if mv.State.State == ModelState_MODEL_STATE_AVAILABLE {
			lastAvailableIdx = idx
		}
	}
	return lastAvailableIdx
}

func (m *Model) GetLastAvailableModelVersion() *ModelVersion {
	lastAvailableIdx := m.getLastAvailableModelVersionIdx()
	if lastAvailableIdx != -1 {
		return m.Versions[lastAvailableIdx]
	}
	return nil
}

func (m *ModelVersion) Inactive() bool {
	for _, v := range m.Replicas {
		if !v.State.Inactive() {
			return false
		}
	}
	return true
}

func (m *ModelVersion) DeleteReplica(replicaIdx int) {
	delete(m.Replicas, int32(replicaIdx))
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
		m == ModelReplicaState_MODEL_REPLICA_STATE_AVAILABLE ||
		m == ModelReplicaState_MODEL_REPLICA_STATE_LOADED_UNAVAILABLE
}
