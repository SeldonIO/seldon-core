package filter

import (
	"fmt"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/fsm/state_machine/model"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/fsm/state_machine/server"
)

type ServerReplicaFilter struct{}

func (r ServerReplicaFilter) Name() string {
	return "ServerReplicaFilter"
}

func (r ServerReplicaFilter) Filter(model *model.VersionStatus, server *server.Snapshot) bool {
	return len(server.Replicas) > 0
}

func (r ServerReplicaFilter) Description(model *model.VersionStatus, server *server.Snapshot) string {
	return fmt.Sprintf("%d server replicas (waiting for server replicas to connect)", len(server.Replicas))
}
