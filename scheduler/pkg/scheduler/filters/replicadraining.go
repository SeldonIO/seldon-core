package filters

import (
	"fmt"

	"github.com/seldonio/seldon-core/scheduler/pkg/store"
)

type ReplicaDrainingFilter struct{}

func (r ReplicaDrainingFilter) Name() string {
	return "ReplicaDrainingFilter"
}

func (r ReplicaDrainingFilter) Filter(model *store.ModelVersion, replica *store.ServerReplica) bool {
	return !replica.GetIsDraining()
}

func (r ReplicaDrainingFilter) Description(model *store.ModelVersion, replica *store.ServerReplica) string {
	return fmt.Sprintf(
		"Replica server %d is draining check %t",
		replica.GetReplicaIdx(), replica.GetIsDraining())
}
