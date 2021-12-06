package store

import (
	"errors"

	pba "github.com/seldonio/seldon-core/scheduler/apis/mlops/agent"
	pb "github.com/seldonio/seldon-core/scheduler/apis/mlops/scheduler"
)

type ServerSnapshot struct {
	Name     string
	Replicas map[int]*ServerReplica
	Shared   bool
}

type ModelSnapshot struct {
	Name     string
	Versions []*ModelVersion
	Deleted  bool
}

func (m *ModelSnapshot) GetLatest() *ModelVersion {
	if len(m.Versions) > 0 {
		return m.Versions[len(m.Versions)-1]
	} else {
		return nil
	}

}

var (
	ModelVersionExistsErr          = errors.New("model version already exists")
	ModelNotLatestVersionRejectErr = errors.New("Model version is not latest. Rejecting update.")
)

type SchedulerStore interface {
	UpdateModel(config *pb.ModelDetails) error
	GetModel(key string) (*ModelSnapshot, error)
	ExistsModelVersion(key string, version string) bool
	RemoveModel(modelKey string) error
	GetServers() ([]*ServerSnapshot, error)
	GetServer(serverKey string) (*ServerSnapshot, error)
	UpdateLoadedModels(modelKey string, version string, serverKey string, replicas []*ServerReplica) error
	UpdateModelState(modelKey string, version string, serverKey string, replicaIdx int, availableMemory *uint64, state ModelReplicaState, reason string) error
	AddServerReplica(request *pba.AgentSubscribeRequest) error
	RemoveServerReplica(serverName string, replicaIdx int) ([]string, error) // return previously loaded models
}
