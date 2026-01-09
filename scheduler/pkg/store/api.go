/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package store

import (
	pba "github.com/seldonio/seldon-core/apis/go/v2/mlops/agent"
	pb "github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"
	"github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler/db"
)

type ServerSnapshot struct {
	Name             string
	Replicas         map[int]*ServerReplica
	Shared           bool
	ExpectedReplicas int
	MinReplicas      int
	MaxReplicas      int
	KubernetesMeta   *pb.KubernetesMeta
	Stats            *ServerStats
}

type ServerStats struct {
	NumEmptyReplicas          uint32
	MaxNumReplicaHostedModels uint32
}

func (s *ServerSnapshot) String() string {
	return s.Name
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
	if m == nil { // TODO Make safe by not working on actual object
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

func (m *ModelSnapshot) getLastModelGwAvailableModelIdx() int {
	if m == nil { // TODO Make safe by not working on actual object
		return -1
	}
	lastAvailableIdx := -1
	for idx, mv := range m.Versions {
		if mv.state.ModelGwState == ModelAvailable {
			lastAvailableIdx = idx
		}
	}
	return lastAvailableIdx
}

func (m *ModelSnapshot) CanReceiveTraffic() bool {
	if m.GetLastAvailableModel() != nil {
		return true
	}
	latestVersion := m.GetLatest()
	if latestVersion != nil && latestVersion.HasLiveReplicas() {
		return true
	}
	return false
}

func (m *ModelSnapshot) GetLastAvailableModel() *ModelVersion {
	if m == nil { // TODO Make safe by not working on actual object
		return nil
	}
	lastAvailableIdx := m.getLastAvailableModelIdx()
	if lastAvailableIdx != -1 {
		return m.Versions[lastAvailableIdx]
	}
	return nil
}

func (m *ModelSnapshot) GetVersionsBeforeLastAvailable() []*ModelVersion {
	if m == nil { // TODO Make safe by not working on actual object
		return nil
	}
	lastAvailableIdx := m.getLastAvailableModelIdx()
	if lastAvailableIdx != -1 {
		return m.Versions[0:lastAvailableIdx]
	}
	return nil
}

func (m *ModelSnapshot) GetVersionsBeforeLastModelGwAvailable() []*ModelVersion {
	if m == nil { // TODO Make safe by not working on actual object
		return nil
	}
	lastAvailableIdx := m.getLastModelGwAvailableModelIdx()
	if lastAvailableIdx != -1 {
		return m.Versions[0:lastAvailableIdx]
	}
	return nil
}

//go:generate go tool mockgen -source=./api.go -destination=./mock/store.go -package=mock ModelServerAPI
type ModelServerAPI interface {
	UpdateModel(config *pb.LoadModelRequest) error
	GetModel(key string) (*db.Model, error)
	GetModels() ([]*db.Model, error)
	LockModel(modelName string)
	UnlockModel(modelName string)
	LockServer(serverName string)
	UnlockServer(serverName string)
	RemoveModel(req *pb.UnloadModelRequest) error
	GetServers() ([]*db.Server, error)
	GetServer(serverName string, modelDetails bool) (*db.Server, *ServerStats, error)
	UpdateLoadedModels(modelName string, version uint32, serverKey string, replicas []*db.ServerReplica) error
	UnloadVersionModels(modelName string, version uint32) (bool, error)
	UnloadModelGwVersionModels(modelName string, version uint32) (bool, error)
	UpdateModelState(modelName string, version uint32, serverKey string, replicaIdx int, availableMemory *uint64, expectedState, desiredState db.ModelReplicaState, reason string, runtimeInfo *pb.ModelRuntimeInfo) error
	AddServerReplica(request *pba.AgentSubscribeRequest) error
	ServerNotify(request *pb.ServerNotify) error
	RemoveServerReplica(serverName string, replicaIdx int) ([]string, error) // return previously loaded models
	DrainServerReplica(serverName string, replicaIdx int) ([]string, error)  // return previously loaded models
	FailedScheduling(modelName string, version uint32, reason string, reset bool) error
	GetAllModels() ([]string, error)
	SetModelGwModelState(modelName string, versionNumber uint32, status db.ModelState, reason string, source string) error
	// TODO better name... should it even be on tihs interface?
	EmitEvents() error
}
