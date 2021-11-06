package store

import (
	pba "github.com/seldonio/seldon-core/scheduler/apis/mlops/agent"
	pb "github.com/seldonio/seldon-core/scheduler/apis/mlops/scheduler"
	"google.golang.org/protobuf/proto"
)

type LocalSchedulerStore struct {
	servers map[string]*Server
	models map[string]*Model
	failedToScheduleModels map[string]bool
}

func NewLocalSchedulerStore() *LocalSchedulerStore {
	m := LocalSchedulerStore{}
	m.servers = make(map[string]*Server)
	m.models = make(map[string]*Model)
	m.failedToScheduleModels = make(map[string]bool)
	return &m
}

type Model struct {
	versionMap map[string]*ModelVersion
	versions []*ModelVersion
	deleted bool
}

func NewModel() *Model {
	return &Model{
		versionMap: make(map[string]*ModelVersion),
	}
}

type ModelVersion struct {
	config *pb.ModelDetails
	server string
	replicas map[int]ModelReplicaState
	deleted bool
	state ModelState
}

func NewDefaultModelVersion(config *pb.ModelDetails) *ModelVersion {
	return &ModelVersion{
		config: config,
		replicas: make(map[int]ModelReplicaState),
		deleted: false,
		state: ModelStateUnknown,
	}
}

func NewModelVersion(config *pb.ModelDetails,
	server string,
	replicas map[int]ModelReplicaState,
	deleted bool,
	state ModelState) *ModelVersion {
	return &ModelVersion{
		config: config,
		server: server,
		replicas: replicas,
		deleted: deleted,
		state: state,
	}
}

type Server struct {
	name string
	replicas map[int]*ServerReplica
	shared bool
}

func NewServer(name string, shared bool) *Server {
	return &Server{
		name: name,
		replicas: make(map[int]*ServerReplica),
		shared: shared,
	}
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

func NewServerReplica(inferenceSvc string,
	inferencePort int32,
	replicaIdx int,
	server *Server,
	capabilities []string,
	memory uint64,
	availableMemory uint64,
	loadedModels map[string]bool,
	overCommit bool) *ServerReplica {
	return &ServerReplica{
		inferenceSvc: inferenceSvc,
		inferencePort: inferencePort,
		replicaIdx: replicaIdx,
		server: server,
		capabilities: capabilities,
		memory: memory,
		availableMemory: availableMemory,
		loadedModels: loadedModels,
		overCommit: overCommit,
	}
}

func NewServerReplicaFromConfig (server *Server, replicaIdx int, loadedModels map[string]bool, config *pba.ReplicaConfig) *ServerReplica {
	return &ServerReplica{
		inferenceSvc: config.GetInferenceSvc(),
		inferencePort: config.GetInferencePort(),
		replicaIdx: replicaIdx,
		server: server,
		capabilities: config.GetCapabilities(),
		memory: config.GetMemoryBytes(),
		availableMemory: config.GetAvailableMemoryBytes(),
		loadedModels: loadedModels,
		overCommit: config.GetOverCommit(),
	}
}

type ModelState uint32

const (
	ModelStateUnknown ModelState = iota
	ModelProgressing
	ModelAvailable
	ModelFailed
	ModelTerminating
	ModelTerminated
	ModelTerminateFailed
)

type ModelReplicaState uint32

const (
	ModelReplicaStateUnknown ModelReplicaState = iota
	LoadRequested
	Loading
	Loaded
	LoadFailed
	UnloadRequested
	Unloading
	Unloaded
	UnloadFailed
)

func (m *Model) HasLatest() bool {
	return len(m.versions) > 0
}

func (m *Model) Latest() *ModelVersion {
	if len(m.versions) > 0 {
		return m.versions[len(m.versions) - 1]
	} else {
		return nil
	}
}

func (m *Model) Inactive() bool {
	for _,mv := range m.versions {
		if !mv.Inactive() {
			return false
		}
	}
	return true
}

func (m *Model) isDeleted() bool {
	return m.deleted
}

func (m *ModelVersion) GetVersion() string {
	return m.config.GetVersion()
}

func (m *ModelVersion) GetRequiredMemory() uint64 {
	return m.config.GetMemoryBytes()
}

func (m *ModelVersion) GetRequirements() []string {
	return m.config.GetRequirements()
}

func (m *ModelVersion) DesiredReplicas() int {
	return int(m.config.Replicas)
}

func (m *ModelVersion) Details() *pb.ModelDetails {
	return proto.Clone(m.config).(*pb.ModelDetails)
}

func (m *ModelVersion) Server() string {
	return m.server
}

func (m *ModelVersion) ReplicaState() map[int32]string {
	replicaState := make(map[int32]string)
	for k,v:= range  m.replicas {
		replicaState[int32(k)] = v.String()
	}
	return replicaState
}


func (m *ModelVersion) GetModelReplicaState(replicaIdx int) ModelReplicaState {
	state, ok := m.replicas[replicaIdx]
	if !ok {
		return ModelReplicaStateUnknown
	}
	return state
}

func (m *ModelVersion) GetReplicaForState(state ModelReplicaState) []int {
	var assignment []int
	for k, v := range m.replicas {
		if v == state {
			assignment = append(assignment, k)
		}
	}
	return assignment
}

func (m *ModelVersion) GetRequestedServer() *string {
	return m.config.Server
}

func (m *ModelVersion) HasServer() bool {
	return m.server != ""
}

func (m *ModelVersion) Inactive() bool {
	for _,v := range m.replicas {
		if !(v == Unloaded || v == UnloadFailed || v == ModelReplicaStateUnknown) {
			return false
		}
	}
	return true
}

func (m *ModelVersion) IsLoading(replicaIdx int) bool {
	for r,v := range m.replicas {
		if r == replicaIdx && (v == Loaded || v == LoadRequested || v == Loading) {
			return true
		}
	}
	return false
}

func (m *ModelVersion) NoLiveReplica() bool {
	for _,v := range m.replicas {
		if !v.NoEndpoint() {
			return false
		}
	}
	return true;
}

func (m *ModelVersion) GetAssignment() []int {
	var assignment []int
	for k,v := range m.replicas {
		if v == Loaded {
			assignment = append(assignment,k)
		}
	}
	return assignment
}

func (m *ModelVersion) Key() string {
	return m.config.Name
}

func (m *ModelVersion) IsDeleted() bool {
	return m.deleted
}

func (s *Server) Key() string {
	return s.name
}

func (s *Server) NumReplicas() uint32 {
	return uint32(len(s.replicas))
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

func (s *ServerReplica) GetAvailableMemory() uint64 {
	return s.availableMemory
}

func (s *ServerReplica) GetMemory() uint64 {
	return s.memory
}

func (s *ServerReplica) GetCapabilities() []string {
	return s.capabilities
}

func (s *ServerReplica) GetReplicaIdx() int {
	return s.replicaIdx
}

func (s *ServerReplica) GetInferenceSvc() string {
	return s.inferenceSvc
}

func (s *ServerReplica) GetInferencePort() int32 {
	return s.inferencePort
}

func (m ModelReplicaState) CanLoad() bool {
	return !(m == LoadRequested || m == Loading || m == Loaded || m == ModelReplicaStateUnknown)
}

func (m ModelReplicaState) CanUnload() bool {
	return !(m == UnloadRequested || m == Unloading || m == Unloaded || m == ModelReplicaStateUnknown)
}

func (m ModelReplicaState) CanRemove() bool {
	return (m == Unloaded || m == ModelReplicaStateUnknown || m == UnloadFailed)
}

func (m ModelReplicaState) NoEndpoint() bool {
	return (m == Unloaded || m == ModelReplicaStateUnknown || m == UnloadFailed || m == Unloading || m == UnloadRequested)
}

func (m ModelReplicaState) AlreadyLoadingOrLoaded() bool {
	return (m == Loading || m == Loaded)
}

func (m ModelReplicaState) AlreadyUnloadingOrUnloaded() bool {
	return (m == Unloading || m == Unloaded)
}

func (me ModelReplicaState) String() string {
	return [...]string{"Unknown", "LoadRequested", "Loading", "Loaded", "LoadFailed", "UnloadRequested", "Unloading", "Unloaded", "UnloadFailed"}[me]
}

func (m ModelReplicaState) IsLoadingState() bool {
	switch m {
	case LoadRequested, Loading, Loaded:
		return true
	default:
		return false
	}
}




