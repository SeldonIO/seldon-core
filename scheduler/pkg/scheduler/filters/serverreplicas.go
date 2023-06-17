package filters

import (
	"fmt"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store"
)

type ServerReplicaFilter struct{}

func (r ServerReplicaFilter) Name() string {
	return "ServerReplicaFilter"
}

func (r ServerReplicaFilter) Filter(model *store.ModelVersion, server *store.ServerSnapshot) bool {
	return len(server.Replicas) > 0
}

func (r ServerReplicaFilter) Description(model *store.ModelVersion, server *store.ServerSnapshot) string {
	return fmt.Sprintf("expected replicas %d > 0", len(server.Replicas))
}
