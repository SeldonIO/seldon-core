/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package model

import pb "github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"

func (mrs *ReplicaStatus) CanReceiveTraffic() bool {
	switch mrs.State {
	case
		pb.ModelReplicaStatus_Loaded,
		pb.ModelReplicaStatus_Available,
		pb.ModelReplicaStatus_LoadedUnavailable,
		pb.ModelReplicaStatus_Draining:
		return true
	default:
		return false
	}
}

func (mrs *ReplicaStatus) AlreadyLoadingOrLoaded() bool {
	switch mrs.State {
	case
		pb.ModelReplicaStatus_Loading,
		pb.ModelReplicaStatus_Loaded,
		pb.ModelReplicaStatus_Available,
		pb.ModelReplicaStatus_LoadedUnavailable:
		return true
	default:
		return false
	}
}

func (mrs *ReplicaStatus) UnloadingOrUnloaded() bool {
	switch mrs.State {
	case
		pb.ModelReplicaStatus_UnloadEnvoyRequested,
		pb.ModelReplicaStatus_UnloadRequested,
		pb.ModelReplicaStatus_Unloading,
		pb.ModelReplicaStatus_Unloaded,
		pb.ModelReplicaStatus_ModelReplicaStateUnknown:
		return true
	default:
		return false
	}
}

func (mrs *ReplicaStatus) Inactive() bool {
	switch mrs.State {
	case
		pb.ModelReplicaStatus_Unloaded,
		pb.ModelReplicaStatus_UnloadFailed,
		pb.ModelReplicaStatus_ModelReplicaStateUnknown,
		pb.ModelReplicaStatus_LoadFailed:
		return true
	default:
		return false
	}
}

func (mrs *ReplicaStatus) Active() bool {
	if mrs.Inactive() {
		return false
	}
	return true
}

func (mrs *ReplicaStatus) IsLoadingOrLoaded() bool {
	switch mrs.State {
	case
		pb.ModelReplicaStatus_Loaded,
		pb.ModelReplicaStatus_LoadRequested,
		pb.ModelReplicaStatus_Loading,
		pb.ModelReplicaStatus_Available,
		pb.ModelReplicaStatus_LoadedUnavailable:
		return true
	default:
		return false
	}
}

// IsModelFullyInactive checks if ALL versions of a model are inactive
func (ms *Snapshot) IsModelFullyInactive() bool {
	if ms == nil || len(ms.Versions) == 0 {
		return true
	}

	for _, version := range ms.Versions {
		if NewModelVersion(version).Active() == false {
			return true
		}
	}

	return false
}

func (ms *Snapshot) GetLatestModelVersionStatus() *VersionStatus {
	if ms == nil || len(ms.Versions) == 0 {
		return nil
	}

	return NewModelVersion(ms.Versions[len(ms.Versions)-1])
}
