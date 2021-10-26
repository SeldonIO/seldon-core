package store

import (
	pba "github.com/seldonio/seldon-core/scheduler/apis/mlops/agent"
	pb "github.com/seldonio/seldon-core/scheduler/apis/mlops/scheduler"
)

//TODO remove pointer returns to make thread safe
type SchedulerStore interface {
	// Server
	UpdateServerReplica(request *pba.AgentSubscribeRequest) error
	RemoveServerReplicaAndRedeployModels(serverName string, replicaIdx int) error
	GetServer (key string) (*Server, error)
	GetServerReplica(key string, replicaIdx int) (*ServerReplica, error)

	// Models
	CreateModel(key string, config *pb.ModelDetails) error
	UpdateModel(key string, config *pb.ModelDetails) error
	GetModel(key string) (*Model, error)
	RemoveModel(key string) error

	// Assign model to server
	ScheduleModelToServer(modelKey string) error
	UpdateModelOnServer(modelKey string, serverKey string) error
	SetModelState(modelKey string, serverKey string, replicaIdx int, state ModelState, availableMemory *uint64) error
}
