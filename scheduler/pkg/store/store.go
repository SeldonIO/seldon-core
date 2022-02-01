package store

import (
	pba "github.com/seldonio/seldon-core/scheduler/apis/mlops/agent"
	pb "github.com/seldonio/seldon-core/scheduler/apis/mlops/scheduler"
)

type ServerSnapshot struct {
	Name             string
	Replicas         map[int]*ServerReplica
	Shared           bool
	ExpectedReplicas int
	KubernetesMeta   *pb.KubernetesMeta
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

func (m *ModelSnapshot) GetVersion(version uint32) *ModelVersion {
	for _, mv := range m.Versions {
		if mv.GetVersion() == version {
			return mv
		}
	}
	return nil
}

func (m *ModelSnapshot) GetPrevious() *ModelVersion {
	if len(m.Versions) > 1 {
		return m.Versions[len(m.Versions)-2]
	} else {
		return nil
	}
}

func (m *ModelSnapshot) getLastAvailableModelIdx() int {
	if m == nil { //TODO Make safe by not working on actual object
		return -1
	}
	lastAvailableIdx := -1
	for idx, mv := range m.Versions {
		if mv.state.State == ModelAvailable {
			lastAvailableIdx = idx
		}
	}
	return lastAvailableIdx
}

func (m *ModelSnapshot) GetLastAvailableModel() *ModelVersion {
	if m == nil { //TODO Make safe by not working on actual object
		return nil
	}
	lastAvailableIdx := m.getLastAvailableModelIdx()
	if lastAvailableIdx != -1 {
		return m.Versions[lastAvailableIdx]
	}
	return nil
}

func (m *ModelSnapshot) GetVersionsBeforeLastAvailable() []*ModelVersion {
	if m == nil { //TODO Make safe by not working on actual object
		return nil
	}
	lastAvailableIdx := m.getLastAvailableModelIdx()
	if lastAvailableIdx != -1 {
		return m.Versions[0:lastAvailableIdx]
	}
	return nil
}

type SchedulerStore interface {
	UpdateModel(config *pb.LoadModelRequest) error
	GetModel(key string) (*ModelSnapshot, error)
	RemoveModel(req *pb.UnloadModelRequest) error
	GetServers() ([]*ServerSnapshot, error)
	GetServer(serverKey string) (*ServerSnapshot, error)
	UpdateLoadedModels(modelKey string, version uint32, serverKey string, replicas []*ServerReplica) error
	UnloadVersionModels(modelKey string, version uint32) (bool, error)
	UpdateModelState(modelKey string, version uint32, serverKey string, replicaIdx int, availableMemory *uint64, state ModelReplicaState, reason string) error
	AddServerReplica(request *pba.AgentSubscribeRequest) error
	ServerNotify(request *pb.ServerNotifyRequest) error
	RemoveServerReplica(serverName string, replicaIdx int) ([]string, error) // return previously loaded models
	FailedScheduling(modelVersion *ModelVersion, reason string)
}
