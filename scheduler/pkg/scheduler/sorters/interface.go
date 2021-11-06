package sorters

import "github.com/seldonio/seldon-core/scheduler/pkg/store"

type CandidateServer struct {
	Model          *store.ModelVersion
	Server         *store.ServerSnapshot
	ChosenReplicas []*store.ServerReplica
}

type CandidateReplica struct {
	Model   *store.ModelVersion
	Server  *store.ServerSnapshot
	Replica *store.ServerReplica
}

type ServerSorter interface {
	IsLess(i *CandidateServer, j *CandidateServer) bool
}

type ReplicaSorter interface {
	IsLess(i *CandidateReplica, j *CandidateReplica) bool
}