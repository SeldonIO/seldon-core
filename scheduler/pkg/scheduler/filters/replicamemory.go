package filters

import (
	"fmt"

	"github.com/seldonio/seldon-core/scheduler/pkg/store"
)

type AvailableMemoryReplicaFilter struct{}

func (r AvailableMemoryReplicaFilter) Name() string {
	return "AvailableMemoryReplicaFilter"
}

func (r AvailableMemoryReplicaFilter) Filter(model *store.ModelVersion, replica *store.ServerReplica) bool {
	return model.GetRequiredMemory() <= replica.GetAvailableMemory()
}

func (s AvailableMemoryReplicaFilter) Description(model *store.ModelVersion, replica *store.ServerReplica) string {
	return fmt.Sprintf("model memory %d replica memory %d", model.GetRequiredMemory(), replica.GetAvailableMemory())
}
