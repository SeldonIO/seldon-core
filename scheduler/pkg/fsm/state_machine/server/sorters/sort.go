package sorters

import (
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/fsm/state_machine/model"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/fsm/state_machine/server"
)

type CandidateServer struct {
	Model          *model.VersionStatus
	Server         *server.Snapshot
	ChosenReplicas []*server.Replica
}

type CandidateReplica struct {
	Model   *model.VersionStatus
	Server  *server.Snapshot
	Replica *server.Replica
}

type ServerSorter interface {
	Name() string
	IsLess(i *CandidateServer, j *CandidateServer) bool
}

type ReplicaSorter interface {
	Name() string
	IsLess(i *CandidateReplica, j *CandidateReplica) bool
}
