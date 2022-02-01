package filters

import "github.com/seldonio/seldon-core/scheduler/pkg/store"

type AvailableMemoryReplicaFilter struct{}

func (r AvailableMemoryReplicaFilter) Name() string {
	return "AvailableMemoryReplicaFilter"
}

func (r AvailableMemoryReplicaFilter) Filter(model *store.ModelVersion, replica *store.ServerReplica) bool {
	return model.GetRequiredMemory() <= replica.GetAvailableMemory()
}
