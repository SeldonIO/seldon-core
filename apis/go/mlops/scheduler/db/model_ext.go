/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package db

import (
	pb "github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (m *ModelVersion) GetAssignment() []int {
	var assignment []int
	var draining []int

	for k, v := range m.Replicas {
		if v.State == ModelReplicaState_Loaded ||
			v.State == ModelReplicaState_Available ||
			v.State == ModelReplicaState_LoadedUnavailable {
			assignment = append(assignment, int(k))
		}
		if v.State == ModelReplicaState_Draining {
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

func (m *ModelVersion) ReplicaState() map[int]*ReplicaStatus {
	copyReplicas := make(map[int]*ReplicaStatus, len(m.Replicas))
	for idx, r := range m.Replicas {
		copyReplicas[int(idx)] = r
	}
	return copyReplicas
}

func (m *ModelVersion) GetRequestedServer() *string {
	return m.ModelDefn.GetModelSpec().Server
}

func (m *ModelVersion) GetRequirements() []string {
	return m.ModelDefn.GetModelSpec().GetRequirements()
}

func (m *ModelVersion) IsLoadingOrLoaded(server string, replicaIdx int) bool {
	if server != m.Server {
		return false
	}
	for r, v := range m.Replicas {
		if int(r) == replicaIdx && v.State.IsLoadingOrLoaded() {
			return true
		}
	}
	return false
}

func (m *ModelVersion) DesiredReplicas() int {
	return int(m.ModelDefn.DeploymentSpec.Replicas)
}

func (m *ModelVersion) ModelName() string {
	return m.ModelDefn.GetMeta().GetName()
}

func (m *ModelVersion) IsLoadingOrLoadedOnServer() bool {
	for _, v := range m.Replicas {
		if v.State.AlreadyLoadingOrLoaded() {
			return true
		}
	}
	return false
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

func (m *Model) GetLastAvailableModel() *ModelVersion {
	if m == nil { // TODO Make safe by not working on actual object
		return nil
	}
	lastAvailableIdx := m.getLastAvailableModelIdx()
	if lastAvailableIdx != -1 {
		return m.Versions[lastAvailableIdx]
	}
	return nil
}

func (m *Model) CanReceiveTraffic() bool {
	if m.GetLastAvailableModel() != nil {
		return true
	}
	latestVersion := m.Latest()
	if latestVersion != nil && latestVersion.HasLiveReplicas() {
		return true
	}
	return false
}

func (m *Model) getLastModelGwAvailableModelIdx() int {
	if m == nil { // TODO Make safe by not working on actual object
		return -1
	}
	lastAvailableIdx := -1
	for idx, mv := range m.Versions {
		if mv.State.ModelGwState == ModelState_ModelAvailable {
			lastAvailableIdx = idx
		}
	}
	return lastAvailableIdx
}

func (m *Model) GetVersionsBeforeLastModelGwAvailable() []*ModelVersion {
	if m == nil { // TODO Make safe by not working on actual object
		return nil
	}
	lastAvailableIdx := m.getLastModelGwAvailableModelIdx()
	if lastAvailableIdx != -1 {
		return m.Versions[0:lastAvailableIdx]
	}
	return nil
}

func (m *Model) getLastAvailableModelIdx() int {
	if m == nil { // TODO Make safe by not working on actual object
		return -1
	}
	lastAvailableIdx := -1
	for idx, mv := range m.Versions {
		if mv.State.State == ModelState_ModelAvailable {
			lastAvailableIdx = idx
		}
	}
	return lastAvailableIdx
}

func (m *Model) GetVersionsBeforeLastAvailable() []*ModelVersion {
	if m == nil {
		return nil
	}
	lastAvailableIdx := m.getLastAvailableModelIdx()
	if lastAvailableIdx != -1 {
		return m.Versions[0:lastAvailableIdx]
	}
	return nil
}

func (m *ModelVersion) HasLiveReplicas() bool {
	for _, v := range m.Replicas {
		if v.State.CanReceiveTraffic() {
			return true
		}
	}
	return false
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
	m.initReplicasIfEmpty()
	m.Replicas[int32(replicaIdx)] = &ReplicaStatus{State: state, Timestamp: timestamppb.Now(), Reason: reason}
}

func (m *ModelVersion) UpdateRuntimeInfo(runtimeInfo *pb.ModelRuntimeInfo) {
	if m.ModelDefn.ModelSpec != nil && m.ModelDefn.ModelSpec.ModelRuntimeInfo == nil && runtimeInfo != nil {
		m.ModelDefn.ModelSpec.ModelRuntimeInfo = runtimeInfo
	}
}

func (m *ModelVersion) initReplicasIfEmpty() {
	if m.Replicas == nil {
		m.Replicas = make(map[int32]*ReplicaStatus)
	}
}

func (m *ModelVersion) GetModelReplicaState(replicaIdx int) ModelReplicaState {
	m.initReplicasIfEmpty()
	state, ok := m.Replicas[int32(replicaIdx)]
	if !ok {
		return ModelReplicaState_ModelReplicaStateUnknown
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
		if mv.State.State == ModelState_ModelAvailable {
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
	return m == ModelReplicaState_Loaded ||
		m == ModelReplicaState_Available ||
		m == ModelReplicaState_LoadedUnavailable ||
		m == ModelReplicaState_Draining
}

func (m ModelReplicaState) AlreadyLoadingOrLoaded() bool {
	return m == ModelReplicaState_Loading ||
		m == ModelReplicaState_Loaded ||
		m == ModelReplicaState_Available ||
		m == ModelReplicaState_LoadedUnavailable
}

func (m ModelReplicaState) UnloadingOrUnloaded() bool {
	return m == ModelReplicaState_UnloadEnvoyRequested ||
		m == ModelReplicaState_UnloadRequested ||
		m == ModelReplicaState_Unloading ||
		m == ModelReplicaState_Unloaded ||
		m == ModelReplicaState_ModelReplicaStateUnknown
}

func (m ModelReplicaState) Inactive() bool {
	return m == ModelReplicaState_Unloaded ||
		m == ModelReplicaState_UnloadFailed ||
		m == ModelReplicaState_ModelReplicaStateUnknown ||
		m == ModelReplicaState_LoadFailed
}

func (m ModelReplicaState) IsLoadingOrLoaded() bool {
	return m == ModelReplicaState_Loaded ||
		m == ModelReplicaState_LoadRequested ||
		m == ModelReplicaState_Loading ||
		m == ModelReplicaState_Available ||
		m == ModelReplicaState_LoadedUnavailable
}
