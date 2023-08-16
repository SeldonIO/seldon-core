/*
Copyright 2022 Seldon Technologies Ltd.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package store

import (
	pba "github.com/seldonio/seldon-core/apis/go/v2/mlops/agent"
	pb "github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"
)

type ServerSnapshot struct {
	Name             string
	Replicas         map[int]*ServerReplica
	Shared           bool
	ExpectedReplicas int
	KubernetesMeta   *pb.KubernetesMeta
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

type ModelStore interface {
	UpdateModel(config *pb.LoadModelRequest) error
	GetModel(key string) (*ModelSnapshot, error)
	GetModels() ([]*ModelSnapshot, error)
	LockModel(modelId string)
	UnlockModel(modelId string)
	RemoveModel(req *pb.UnloadModelRequest) error
	GetServers(shallow bool, modelDetails bool) ([]*ServerSnapshot, error)
	GetServer(serverKey string, shallow bool, modelDetails bool) (*ServerSnapshot, error)
	UpdateLoadedModels(modelKey string, version uint32, serverKey string, replicas []*ServerReplica) error
	UnloadVersionModels(modelKey string, version uint32) (bool, error)
	UpdateModelState(modelKey string, version uint32, serverKey string, replicaIdx int, availableMemory *uint64, expectedState, desiredState ModelReplicaState, reason string) error
	AddServerReplica(request *pba.AgentSubscribeRequest) error
	ServerNotify(request *pb.ServerNotifyRequest) error
	RemoveServerReplica(serverName string, replicaIdx int) ([]string, error) // return previously loaded models
	DrainServerReplica(serverName string, replicaIdx int) ([]string, error)  // return previously loaded models
	FailedScheduling(modelVersion *ModelVersion, reason string, reset bool)
	GetAllModels() []string
}
