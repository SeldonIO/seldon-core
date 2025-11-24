package filter

import (
	"fmt"
	"math"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/fsm/state_machine/model"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/fsm/state_machine/server"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store"
)

type AvailableMemoryReplicaFilter struct{}

func (r AvailableMemoryReplicaFilter) Name() string {
	return "AvailableMemoryReplicaFilter"
}

func isModelReplicaLoadedOnServerReplica(modelVersion *model.VersionStatus, replica *server.Replica) bool {
	if modelVersion.HasServer() {

		rep, ok := modelVersion.GetModelReplicaState()[int32(replica.ReplicaIdx)]
		if !ok {
			return false
		}

		return modelVersion.ServerName == replica.ServerName && model.NewModelReplicaStatus(rep).AlreadyLoadingOrLoaded()
	}

	return false
}

func (r AvailableMemoryReplicaFilter) Filter(modelVersion *model.VersionStatus, replica *server.Replica) bool {
	mem := math.Max(0, float64(replica.Memory)-float64(replica.ReservedMemory))
	return modelVersion.GetRequiredMemory() <= uint64(mem) || isModelReplicaLoadedOnServerReplica(modelVersion, replica)
}

func (r AvailableMemoryReplicaFilter) Description(model *store.ModelVersion, replica *store.ServerReplica) string {
	return fmt.Sprintf("model memory %d replica memory %d", model.GetRequiredMemory(), replica.GetAvailableMemory())
}
