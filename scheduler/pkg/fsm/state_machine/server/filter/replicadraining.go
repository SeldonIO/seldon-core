package filter

import (
	"fmt"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/fsm/state_machine/model"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/fsm/state_machine/server"
)

type ReplicaDrainingFilter struct{}

func (r ReplicaDrainingFilter) Name() string {
	return "ReplicaDrainingFilter"
}

func (r ReplicaDrainingFilter) Filter(modelVersion *model.VersionStatus, replica *server.Replica) bool {
	return !replica.IsDraining
}

func (r ReplicaDrainingFilter) Description(modelVersion *model.VersionStatus, replica *server.Replica) string {
	return fmt.Sprintf(
		"Replica server %d is draining check %t",
		replica.ReplicaIdx, replica.IsDraining)
}
