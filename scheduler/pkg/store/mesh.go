package store

import (
	pb "github.com/seldonio/seldon-core/scheduler/apis/mlops/scheduler"
	"google.golang.org/protobuf/proto"
)

type LocalSchedulerStore struct {
	servers map[string]*Server
	models map[string]*Model
	failedToScheduleModels map[string]bool
}

type Model struct {
	config *pb.ModelDetails
	server string
	replicas map[int]ModelState
	deleted bool
}

type Server struct {
	name string
	replicas map[int]*ServerReplica
}

type ServerReplica struct {
	inferenceSvc string
	inferencePort int32
	replicaIdx int
	server *Server
	capabilities []string
	memory uint64
	availableMemory uint64
	loadedModels map[string]bool
	overCommit bool
}

type ModelState uint32

const (
	Unknown ModelState = iota
	LoadRequested
	Loading
	Loaded
	LoadFailed
	UnloadRequested
	Unloading
	Unloaded
	UnloadFailed
)

func (m *Model) Details() *pb.ModelDetails {
	return proto.Clone(m.config).(*pb.ModelDetails)
}

func (m *Model) Server() string {
	return m.server
}

func (m *Model) ReplicaState() map[int32]string {
	replicaState := make(map[int32]string)
	for k,v:= range  m.replicas {
		replicaState[int32(k)] = v.String()
	}
	return replicaState
}


func (m *Model) GetModelReplicaState(replicaIdx int) ModelState {
	state, ok := m.replicas[replicaIdx]
	if !ok {
		return Unknown
	}
	return state
}

func (m *Model) GetReplicaForState(state ModelState) []int {
	var assignment []int
	for k, v := range m.replicas {
		if v == state {
			assignment = append(assignment, k)
		}
	}
	return assignment
}

func (m *Model) HasServer() bool {
	return m.server != ""
}

func (m *Model) NumReplicas() int {
	return len(m.replicas)
}

func (m *Model) NumActiveReplicas() uint32 {
	count := uint32(0)
	for _,v := range m.replicas {
		if v.CanUnload() {
			count++
		}
	}
	return count;
}

func (m *Model) CanRemove() bool {
	if !m.deleted {
		return false
	}
	for _,v := range m.replicas {
		if !v.CanRemove() {
			return false
		}
	}
	return true;
}

func (m *Model) isLiveReplica(replicaIdx int) bool {
	for r,v := range m.replicas {
		if r == replicaIdx && v == Loaded {
			return true
		}
	}
	return false
}

func (m *Model) NoLiveReplica() bool {
	for _,v := range m.replicas {
		if !v.NoEndpoint() {
			return false
		}
	}
	return true;
}

func (m *Model) GetAssignment() []int {
	var assignment []int
	for k,v := range m.replicas {
		if v == Loaded {
			assignment = append(assignment,k)
		}
	}
	return assignment
}

func (m *Model) Key() string {
	return m.config.Name
}

func (m *Model) isDeleted() bool {
	return m.deleted
}

func (m *Model) GetUri() string {
	return m.config.Uri
}

func (m *Model) GetStorageSecretName() *string {
	return m.config.StorageSecretName
}

func (s* Server) maxReplicas() uint32 {
	if len(s.replicas) == 0 {
		return 0
	}
	maxIdx := 0;
	for k:= range s.replicas {
		if k>maxIdx {
			maxIdx = k
		}
	}
	return uint32(maxIdx+1)
}

func (s *Server) Key() string {
	return s.name
}

func (s *Server) NumReplicas() int {
	return len(s.replicas)
}

func (s *Server) GetAvailableMemory(idx int) uint64 {
	if s != nil && idx < len(s.replicas) {
		return s.replicas[idx].availableMemory
	}
	return 0
}

func (s *Server) GetMemory(idx int) uint64 {
	if s != nil && idx < len(s.replicas) {
		return s.replicas[idx].memory
	}
	return 0
}

func (s *Server) GetReplicaInferenceSvc(idx int) string {
	return s.replicas[idx].inferenceSvc
}

func (s *Server) GetReplicaInferencePort(idx int) int32 {
	return s.replicas[idx].inferencePort
}



func (s *ServerReplica) GetLoadedModels() []string {
	var models []string
	for model := range s.loadedModels {
		models = append(models, model)
	}
	return models
}

func NewLocalSchedulerStore() *LocalSchedulerStore {
	m := LocalSchedulerStore{}
	m.servers = make(map[string]*Server)
	m.models = make(map[string]*Model)
	m.failedToScheduleModels = make(map[string]bool)
	return &m
}

func (m ModelState) CanLoad() bool {
	return !(m == LoadRequested || m == Loading || m == Loaded || m == Unknown )
}

func (m ModelState) CanUnload() bool {
	return !(m == UnloadRequested || m == Unloading || m == Unloaded || m == Unknown)
}

func (m ModelState) CanRemove() bool {
	return (m == Unloaded || m == Unknown || m == UnloadFailed)
}

func (m ModelState) NoEndpoint() bool {
	return (m == Unloaded || m == Unknown || m == UnloadFailed || m == Unloading || m == UnloadRequested)
}

func (me ModelState) String() string {
	return [...]string{"Unknown", "LoadRequested", "Loading", "Loaded", "LoadFailed", "UnloadRequested", "Unloading", "Unloaded", "UnloadFailed"}[me]
}




