package filters

import (
	"fmt"

	"github.com/seldonio/seldon-core/scheduler/pkg/store"
)

type AvailableMemoryReplicaFilter struct{}

func (r AvailableMemoryReplicaFilter) Name() string {
	return "AvailableMemoryReplicaFilter"
}

func isModelReplicaLoadedOnServerReplica(model *store.ModelVersion, replica *store.ServerReplica) bool {
	if model.HasServer() {
		return model.Server() == replica.GetServerName() && model.GetModelReplicaState(replica.GetReplicaIdx()).AlreadyLoadingOrLoaded()
	}
	return false
}

func (r AvailableMemoryReplicaFilter) Filter(model *store.ModelVersion, replica *store.ServerReplica) bool {
	return model.GetRequiredMemory() <= replica.GetAvailableMemory() || isModelReplicaLoadedOnServerReplica(model, replica)
}

func (r AvailableMemoryReplicaFilter) Description(model *store.ModelVersion, replica *store.ServerReplica) string {
	return fmt.Sprintf("model memory %d replica memory %d", model.GetRequiredMemory(), replica.GetAvailableMemory())
}
