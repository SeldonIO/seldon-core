package filters

import "github.com/seldonio/seldon-core/scheduler/pkg/store"

type AvailableMemoryFilter struct{}

func (r AvailableMemoryFilter) Filter(model *store.ModelVersion, replica *store.ServerReplica) bool {
	return model.GetRequiredMemory() <= replica.GetAvailableMemory()
}



