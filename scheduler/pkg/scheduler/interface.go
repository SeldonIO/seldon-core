package scheduler

import (
	"github.com/seldonio/seldon-core/scheduler/pkg/store"
)

type Scheduler interface {
	Schedule(modelKey string) error
	ScheduleFailedModels() ([]string, error)
}

type ReplicaFilter interface {
	Filter(model *store.ModelVersion, replica *store.ServerReplica) bool
}

type ServerFilter interface {
	Filter(model *store.ModelVersion, server *store.ServerSnapshot) bool
}
