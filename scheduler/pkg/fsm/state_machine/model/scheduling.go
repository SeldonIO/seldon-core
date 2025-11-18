package model

import pb "github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"

func (mrs *ModelReplicaStatus) CanReceiveTraffic() bool {
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

func (mrs *ModelReplicaStatus) AlreadyLoadingOrLoaded() bool {
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

func (mrs *ModelReplicaStatus) UnloadingOrUnloaded() bool {
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

func (mrs *ModelReplicaStatus) Inactive() bool {
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

func (mrs *ModelReplicaStatus) Active() bool {
	if mrs.Inactive() {
		return false
	}
	return true
}

func (mrs *ModelReplicaStatus) IsLoadingOrLoaded() bool {
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
func (ms *ModelSnapshot) IsModelFullyInactive() bool {
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

func (ms *ModelSnapshot) GetLatestModelVersionStatus() *ModelVersionStatus {
	if ms == nil || len(ms.Versions) == 0 {
		return nil
	}

	return NewModelVersion(ms.Versions[len(ms.Versions)-1])
}
