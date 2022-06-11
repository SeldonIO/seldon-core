package store

import (
	"fmt"
	"strings"
	"sync"
	"time"

	pba "github.com/seldonio/seldon-core/scheduler/apis/mlops/agent"
	pb "github.com/seldonio/seldon-core/scheduler/apis/mlops/scheduler"
	"google.golang.org/protobuf/proto"
)

type LocalSchedulerStore struct {
	servers                map[string]*Server
	models                 map[string]*Model
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
	versions []*ModelVersion
	deleted  bool
}

type ModelVersionID struct {
	Name    string
	Version uint32
}

func (mv *ModelVersionID) String() string {
	return fmt.Sprintf("%s:%d", mv.Name, mv.Version)
}

type ModelVersion struct {
	modelDefn *pb.Model
	version   uint32
	server    string
	replicas  map[int]ReplicaStatus
	deleted   bool
	state     ModelStatus
	mu        sync.RWMutex
}

type ModelStatus struct {
	State               ModelState
	Reason              string
	AvailableReplicas   uint32
	UnavailableReplicas uint32
	Timestamp           time.Time
}

type ReplicaStatus struct {
	State     ModelReplicaState
	Reason    string
	Timestamp time.Time
}

func NewDefaultModelVersion(model *pb.Model, version uint32) *ModelVersion {
	return &ModelVersion{
		version:   version,
		modelDefn: model,
		replicas:  make(map[int]ReplicaStatus),
		deleted:   false,
		state:     ModelStatus{State: ModelStateUnknown},
		mu:        sync.RWMutex{},
	}
}

func NewModelVersion(model *pb.Model, version uint32, server string, replicas map[int]ReplicaStatus, deleted bool, state ModelState) *ModelVersion {
	return &ModelVersion{
		version:   version,
		modelDefn: model,
		server:    server,
		replicas:  replicas,
		deleted:   deleted,
		state:     ModelStatus{State: state},
		mu:        sync.RWMutex{},
	}
}

type Server struct {
	name             string
	replicas         map[int]*ServerReplica
	shared           bool
	expectedReplicas int
	kubernetesMeta   *pb.KubernetesMeta
}

func (s *Server) CreateSnapshot(shallow bool) *ServerSnapshot {
	// TODO: this is considered interface leakage if we do shallow copy by allowing
	// callers to access and change this structure
	// perhaps we consider returning back only what callers need
	var replicas map[int]*ServerReplica
	if !shallow {
		replicas = make(map[int]*ServerReplica, len(s.replicas))
		for k, v := range s.replicas {
			replicas[k] = v.createSnapshot()
		}
	} else {
		replicas = s.replicas
	}
	return &ServerSnapshot{
		Name:             s.name,
		Replicas:         replicas,
		Shared:           s.shared,
		ExpectedReplicas: s.expectedReplicas,
		KubernetesMeta:   proto.Clone(s.kubernetesMeta).(*pb.KubernetesMeta),
	}
}

func (s *Server) SetExpectedReplicas(replicas int) {
	s.expectedReplicas = replicas
}

func (s *Server) SetKubernetesMeta(meta *pb.KubernetesMeta) {
	s.kubernetesMeta = meta
}

func NewServer(name string, shared bool) *Server {
	return &Server{
		name:             name,
		replicas:         make(map[int]*ServerReplica),
		shared:           shared,
		expectedReplicas: -1,
	}
}

type ServerReplica struct {
	muReservedMemory     sync.RWMutex
	muLoadedModels       sync.RWMutex
	inferenceSvc         string
	inferenceHttpPort    int32
	inferenceGrpcPort    int32
	replicaIdx           int
	server               *Server
	capabilities         []string
	memory               uint64
	availableMemory      uint64
	loadedModels         map[ModelVersionID]bool
	overCommitPercentage uint32
	reservedMemory       uint64          // while loading models, internal to scheduler
	uniqueLoadedModels   map[string]bool // precomputed values to speed up ops on scheduler
}

func NewServerReplica(inferenceSvc string,
	inferenceHttpPort int32,
	inferenceGrpcPort int32,
	replicaIdx int,
	server *Server,
	capabilities []string,
	memory uint64,
	availableMemory uint64,
	loadedModels map[ModelVersionID]bool,
	overCommitPercentage uint32) *ServerReplica {
	return &ServerReplica{
		inferenceSvc:         inferenceSvc,
		inferenceHttpPort:    inferenceHttpPort,
		inferenceGrpcPort:    inferenceGrpcPort,
		replicaIdx:           replicaIdx,
		server:               server,
		capabilities:         cleanCapabilities(capabilities),
		memory:               memory,
		availableMemory:      availableMemory,
		loadedModels:         loadedModels,
		overCommitPercentage: overCommitPercentage,
		uniqueLoadedModels:   toUniqueModels(loadedModels),
	}
}

func NewServerReplicaFromConfig(server *Server, replicaIdx int, loadedModels map[ModelVersionID]bool, config *pba.ReplicaConfig, availableMemoryBytes uint64) *ServerReplica {
	return &ServerReplica{
		inferenceSvc:         config.GetInferenceSvc(),
		inferenceHttpPort:    config.GetInferenceHttpPort(),
		inferenceGrpcPort:    config.GetInferenceGrpcPort(),
		replicaIdx:           replicaIdx,
		server:               server,
		capabilities:         cleanCapabilities(config.GetCapabilities()),
		memory:               config.GetMemoryBytes(),
		availableMemory:      availableMemoryBytes,
		loadedModels:         loadedModels,
		overCommitPercentage: config.GetOverCommitPercentage(),
		uniqueLoadedModels:   toUniqueModels(loadedModels),
	}
}

func cleanCapabilities(capabilities []string) []string {
	var cleaned []string
	for _, cap := range capabilities {
		cleaned = append(cleaned, strings.TrimSpace(cap))
	}
	return cleaned
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
	ScheduleFailed
)

func (m ModelState) String() string {
	return [...]string{"ModelStateUnknown", "ModelProgressing", "ModelAvailable", "ModelFailed", "ModelTerminating", "ModelTerminated", "ModelTerminateFailed", "ScheduleFailed"}[m]
}

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
	Available
	LoadedUnavailable
)

var replicaStates = []ModelReplicaState{
	ModelReplicaStateUnknown,
	LoadRequested,
	Loading,
	Loaded,
	LoadFailed,
	UnloadRequested,
	Unloading,
	Unloaded,
	UnloadFailed,
	Available,
	LoadedUnavailable,
}

func (m ModelReplicaState) NoProgressingEndpoint() bool {
	return (m == Unloaded || m == ModelReplicaStateUnknown || m == UnloadFailed || m == Unloading || m == UnloadRequested)
}

func (m ModelReplicaState) AlreadyLoadingOrLoaded() bool {
	return (m == Loading || m == Loaded || m == Available || m == LoadedUnavailable)
}

func (m ModelReplicaState) UnloadingOrUnloaded() bool {
	return (m == UnloadRequested || m == Unloading || m == Unloaded || m == ModelReplicaStateUnknown)
}

func (m ModelReplicaState) Inactive() bool {
	return (m == Unloaded || m == UnloadFailed || m == ModelReplicaStateUnknown)
}

func (m ModelReplicaState) IsLoadingOrLoaded() bool {
	return (m == Loaded || m == LoadRequested || m == Loading || m == Available || m == LoadedUnavailable)
}

func (m ModelReplicaState) IsLoading() bool {
	return (m == LoadRequested || m == Loading)
}

func (me ModelReplicaState) String() string {
	return [...]string{"ModelReplicaStateUnknown", "LoadRequested", "Loading", "Loaded", "LoadFailed", "UnloadRequested", "Unloading", "Unloaded", "UnloadFailed", "Available", "LoadedUnavailable"}[me]
}

func (m *Model) HasLatest() bool {
	return len(m.versions) > 0
}

func (m *Model) Latest() *ModelVersion {
	if len(m.versions) > 0 {
		return m.versions[len(m.versions)-1]
	} else {
		return nil
	}
}

func (m *Model) GetVersion(version uint32) *ModelVersion {
	for _, mv := range m.versions {
		if mv.GetVersion() == version {
			return mv
		}
	}
	return nil
}

func (m *Model) GetVersions() []uint32 {
	versions := make([]uint32, len(m.versions))
	for idx, v := range m.versions {
		versions[idx] = v.version
	}
	return versions
}

func (m *Model) getLastAvailableModelVersionIdx() int {
	lastAvailableIdx := -1
	for idx, mv := range m.versions {
		if mv.state.State == ModelAvailable {
			lastAvailableIdx = idx
		}
	}
	return lastAvailableIdx
}

func (m *Model) GetLastAvailableModelVersion() *ModelVersion {
	lastAvailableIdx := m.getLastAvailableModelVersionIdx()
	if lastAvailableIdx != -1 {
		return m.versions[lastAvailableIdx]
	}
	return nil
}

func (m *Model) Previous() *ModelVersion {
	if len(m.versions) > 1 {
		return m.versions[len(m.versions)-2]
	} else {
		return nil
	}
}

//TODO do we need to consider previous versions?
func (m *Model) Inactive() bool {
	return m.Latest().Inactive()
}

func (m *Model) IsDeleted() bool {
	return m.deleted
}

func (m *ModelVersion) GetVersion() uint32 {
	return m.version
}

func (m *ModelVersion) GetRequiredMemory() uint64 {
	return m.modelDefn.GetModelSpec().GetMemoryBytes()
}

func (m *ModelVersion) GetRequirements() []string {
	return m.modelDefn.GetModelSpec().GetRequirements()
}

func (m *ModelVersion) DesiredReplicas() int {
	return int(m.modelDefn.GetDeploymentSpec().GetReplicas())
}

func (m *ModelVersion) GetModel() *pb.Model {
	return proto.Clone(m.modelDefn).(*pb.Model)
}

func (m *ModelVersion) GetMeta() *pb.MetaData {
	return proto.Clone(m.modelDefn.GetMeta()).(*pb.MetaData)
}

func (m *ModelVersion) GetModelSpec() *pb.ModelSpec {
	return proto.Clone(m.modelDefn.GetModelSpec()).(*pb.ModelSpec)
}

func (m *ModelVersion) GetDeploymentSpec() *pb.DeploymentSpec {
	return proto.Clone(m.modelDefn.GetDeploymentSpec()).(*pb.DeploymentSpec)
}

func (m *ModelVersion) SetDeploymentSpec(spec *pb.DeploymentSpec) {
	m.modelDefn.DeploymentSpec = spec
}

func (m *ModelVersion) Server() string {
	return m.server
}

func (m *ModelVersion) ReplicaState() map[int]ReplicaStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()
	copy := make(map[int]ReplicaStatus, len(m.replicas))
	for idx, r := range m.replicas {
		copy[idx] = r
	}
	return copy
}

func (m *ModelVersion) ModelState() ModelStatus {
	return m.state
}

func (m *ModelVersion) GetModelReplicaState(replicaIdx int) ModelReplicaState {
	m.mu.RLock()
	defer m.mu.RUnlock()
	state, ok := m.replicas[replicaIdx]
	if !ok {
		return ModelReplicaStateUnknown
	}
	return state.State
}

func (m *ModelVersion) UpdateKubernetesMeta(meta *pb.KubernetesMeta) {
	m.modelDefn.Meta.KubernetesMeta = meta
}

func (m *ModelVersion) GetReplicaForState(state ModelReplicaState) []int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var assignment []int
	for k, v := range m.replicas {
		if v.State == state {
			assignment = append(assignment, k)
		}
	}
	return assignment
}

func (m *ModelVersion) GetRequestedServer() *string {
	return m.modelDefn.GetModelSpec().Server
}

func (m *ModelVersion) HasServer() bool {
	return m.server != ""
}

func (m *ModelVersion) Inactive() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, v := range m.replicas {
		if !v.State.Inactive() {
			return false
		}
	}
	return true
}

func (m *ModelVersion) IsLoadingOrLoaded(server string, replicaIdx int) bool {
	if server != m.server {
		return false
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	for r, v := range m.replicas {
		if r == replicaIdx && v.State.IsLoadingOrLoaded() {
			return true
		}
	}
	return false
}

func (m *ModelVersion) NoLiveReplica() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, v := range m.replicas {
		if !v.State.NoProgressingEndpoint() {
			return false
		}
	}
	return true
}

func (m *ModelVersion) GetAssignment() []int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var assignment []int
	for k, v := range m.replicas {
		if v.State == Loaded || v.State == Available || v.State == LoadedUnavailable {
			assignment = append(assignment, k)
		}
	}
	return assignment
}

func (m *ModelVersion) Key() string {
	return m.modelDefn.GetMeta().GetName()
}

func (m *ModelVersion) IsDeleted() bool {
	return m.deleted
}

func (m *ModelVersion) SetReplicaState(replicaIdx int, state ReplicaStatus) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.replicas[replicaIdx] = state
}

func (m *ModelVersion) DeleteReplica(replicaIdx int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.replicas, replicaIdx)
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

func (s *Server) GetReplicaInferenceHttpPort(idx int) int32 {
	return s.replicas[idx].inferenceHttpPort
}

func (s *ServerReplica) createSnapshot() *ServerReplica {
	capabilities := make([]string, len(s.capabilities))
	copy(capabilities, s.capabilities)
	loadedModels := make(map[ModelVersionID]bool, len(s.loadedModels))
	for k, v := range s.loadedModels {
		loadedModels[k] = v
	}
	uniqueLoadedModels := make(map[string]bool, len(s.loadedModels))
	for k, v := range s.uniqueLoadedModels {
		uniqueLoadedModels[k] = v
	}
	return &ServerReplica{
		inferenceSvc:         s.inferenceSvc,
		inferenceHttpPort:    s.inferenceHttpPort,
		inferenceGrpcPort:    s.inferenceGrpcPort,
		replicaIdx:           s.replicaIdx,
		server:               nil, //TODO change ServerReplica to snapshot struct
		capabilities:         capabilities,
		memory:               s.memory,
		availableMemory:      s.availableMemory,
		loadedModels:         loadedModels,
		overCommitPercentage: s.overCommitPercentage,
		reservedMemory:       s.reservedMemory,
		uniqueLoadedModels:   uniqueLoadedModels,
	}
}

func (s *ServerReplica) GetLoadedModelVersions() []ModelVersionID {
	s.muLoadedModels.RLock()
	defer s.muLoadedModels.RUnlock()

	var models []ModelVersionID
	for model := range s.loadedModels {
		models = append(models, model)
	}
	return models
}

func (s *ServerReplica) GetNumLoadedModels() int {
	s.muLoadedModels.RLock()
	defer s.muLoadedModels.RUnlock()

	return len(s.uniqueLoadedModels)
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

func (s *ServerReplica) GetInferenceHttpPort() int32 {
	return s.inferenceHttpPort
}

func (s *ServerReplica) GetInferenceGrpcPort() int32 {
	return s.inferenceGrpcPort
}

func (s *ServerReplica) GetOverCommitPercentage() uint32 {
	return s.overCommitPercentage
}

func (s *ServerReplica) GetReservedMemory() uint64 {
	s.muReservedMemory.RLock()
	defer s.muReservedMemory.RUnlock()

	return s.reservedMemory
}

func (s *ServerReplica) UpdateReservedMemory(memBytes uint64, isAdd bool) {
	s.muReservedMemory.Lock()
	defer s.muReservedMemory.Unlock()

	if isAdd {
		s.reservedMemory += memBytes
	} else {
		if memBytes > s.reservedMemory {
			s.reservedMemory = 0
		} else {
			s.reservedMemory -= memBytes
		}
	}
}

func (s *ServerReplica) addModelVersion(modelName string, modelVersion uint32) {
	s.muLoadedModels.Lock()
	defer s.muLoadedModels.Unlock()

	s.loadedModels[ModelVersionID{Name: modelName, Version: modelVersion}] = true
	s.uniqueLoadedModels[modelName] = true
}

func (s *ServerReplica) deleteModelVersion(modelName string, modelVersion uint32) {
	s.muLoadedModels.Lock()
	defer s.muLoadedModels.Unlock()

	delete(s.loadedModels, ModelVersionID{Name: modelName, Version: modelVersion})
	if !modelExists(s.loadedModels, modelName) {
		delete(s.uniqueLoadedModels, modelName)
	}
}

func toUniqueModels(loadedModels map[ModelVersionID]bool) map[string]bool {
	uniqueModels := make(map[string]bool)
	for key := range loadedModels {
		uniqueModels[key.Name] = true
	}
	return uniqueModels
}

func modelExists(loadedModels map[ModelVersionID]bool, modelKey string) bool {
	found := false
	for key := range loadedModels {
		if key.Name == modelKey {
			found = true
			break
		}
	}
	return found
}
