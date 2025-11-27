package filter

import (
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/fsm/state_machine/model"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/fsm/state_machine/server"
)

type ServerFilter interface {
	Name() string
	Filter(model *model.VersionStatus, server *server.Snapshot) bool
	Description(model *model.VersionStatus, server *server.Snapshot) string
}

type ReplicaFilter interface {
	Name() string
	Filter(modelVersion *model.VersionStatus, replica *server.Replica) bool
	Description(modelVersion *model.VersionStatus, replica *server.Replica) string
}
