package store

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/seldonio/seldon-core/scheduler/apis/mlops/agent"
	pb "github.com/seldonio/seldon-core/scheduler/apis/mlops/scheduler"
	log "github.com/sirupsen/logrus"
)

type MemoryStore struct {
	mu                  sync.RWMutex
	store               *LocalSchedulerStore
	logger              log.FieldLogger
	modelEventListeners []chan<- *ModelSnapshot
}

func NewMemoryStore(logger log.FieldLogger, store *LocalSchedulerStore) *MemoryStore {
	return &MemoryStore{
		store:  store,
		logger: logger.WithField("source", "MemoryStore"),
	}
}

func (m *MemoryStore) addModelVersionIfNotExists(req *agent.ModelVersion) *ModelVersion {
	modelName := req.GetModel().GetMeta().GetName()
	model, ok := m.store.models[modelName]
	if !ok {
		model = &Model{}
		m.store.models[modelName] = model
	}
	if existingModelVersion := model.GetVersion(req.GetVersion()); existingModelVersion == nil {
		modelVersion := NewDefaultModelVersion(req.GetModel(), req.GetVersion())
		model.versions = append(model.versions, modelVersion)
		sort.SliceStable(model.versions, func(i, j int) bool { // resort model versions based on version number
			return model.versions[i].GetVersion() < model.versions[j].GetVersion()
		})
		return modelVersion
	} else {
		return existingModelVersion
	}
}

func (m *MemoryStore) addNextModelVersion(model *Model, pbmodel *pb.Model) {
	version := uint32(1)
	if model.Latest() != nil {
		version = model.Latest().GetVersion() + 1
	}
	modelVersion := NewDefaultModelVersion(pbmodel, version)

	model.versions = append(model.versions, modelVersion)
	sort.SliceStable(model.versions, func(i, j int) bool { // resort model versions based on version number
		return model.versions[i].GetVersion() < model.versions[j].GetVersion()
	})
}

func (m *MemoryStore) UpdateModel(req *pb.LoadModelRequest) {
	logger := m.logger.WithField("func", "UpdateModel")
	m.mu.Lock()
	defer m.mu.Unlock()
	modelName := req.GetModel().GetMeta().GetName()
	model, ok := m.store.models[modelName]
	if !ok {
		model = &Model{}
		m.store.models[modelName] = model
		m.addNextModelVersion(model, req.GetModel())
	} else {
		meq := ModelEqualityCheck(model.Latest().modelDefn, req.GetModel())
		if meq.Equal {
			logger.Debugf("Model %s semantically equal - doing nothing", modelName)
			return
		} else if meq.ModelSpecDiffers {
			logger.Debugf("Model %s model spec differs - adding new version of model", modelName)
			m.addNextModelVersion(model, req.GetModel())
			return
		} else if meq.DeploymentSpecDiffers {
			logger.Debugf("Model %s deployment spec differs - updating latest model version with new spec", modelName)
			model.Latest().SetDeploymentSpec(req.GetModel().GetDeploymentSpec())
		}
		if meq.MetaDiffers {
			// Update just kubernetes meta
			model.Latest().UpdateKubernetesMeta(req.GetModel().GetMeta().GetKubernetesMeta())
		}
	}
}

func (m *MemoryStore) getModelImpl(key string) *ModelSnapshot {
	model, ok := m.store.models[key]
	if ok {
		return &ModelSnapshot{
			Name:     key,
			Versions: model.versions, //TODO make a copy for safety?
			Deleted:  model.deleted,
		}
	} else {
		return &ModelSnapshot{
			Name:     key,
			Versions: nil,
		}
	}
}

func (m *MemoryStore) GetModel(key string) (*ModelSnapshot, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.getModelImpl(key), nil
}

func (m *MemoryStore) RemoveModel(req *pb.UnloadModelRequest) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	model, ok := m.store.models[req.GetModel().GetName()]
	if ok {
		// Updating the k8s meta is required to be updated so status updates back (to manager)
		// will match latest generation value. Previous generation values might be ignored by manager.
		model.Latest().UpdateKubernetesMeta(req.GetKubernetesMeta())
		model.deleted = true
	}
	return nil
}

func createServerSnapshot(server *Server) *ServerSnapshot {
	return &ServerSnapshot{
		Name:     server.name,
		Replicas: server.replicas,
		Shared:   server.shared,
	}
}

func (m *MemoryStore) GetServers() ([]*ServerSnapshot, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var servers []*ServerSnapshot
	for _, server := range m.store.servers {
		servers = append(servers, createServerSnapshot(server))
	}
	return servers, nil
}

func (m *MemoryStore) GetServer(serverKey string) (*ServerSnapshot, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	server := m.store.servers[serverKey]
	return createServerSnapshot(server), nil
}

func (m *MemoryStore) getModelServer(modelKey string, version uint32, serverKey string) (*Model, *ModelVersion, *Server, error) {
	// Validate
	model, ok := m.store.models[modelKey]
	if !ok {
		return nil, nil, nil, fmt.Errorf("failed to find model %s", modelKey)
	}
	modelVersion := model.GetVersion(version)
	if modelVersion == nil {
		return nil, nil, nil, fmt.Errorf("Version not found for model %s, version %d", modelKey, version)
	}
	server, ok := m.store.servers[serverKey]
	if !ok {
		return nil, nil, nil, fmt.Errorf("failed to find server %s", serverKey)
	}
	return model, modelVersion, server, nil
}

func (m *MemoryStore) UpdateLoadedModels(modelKey string, version uint32, serverKey string, replicas []*ServerReplica) error {
	logger := m.logger.WithField("func", "UpdateLoadedModels")
	m.mu.Lock()
	defer m.mu.Unlock()

	// Validate
	model, modelVersion, server, err := m.getModelServer(modelKey, version, serverKey)
	if err != nil {
		return err
	}
	for _, replica := range replicas {
		_, ok := server.replicas[replica.GetReplicaIdx()]
		if !ok {
			return fmt.Errorf("Failed to reserve replica %d as it does not exist on server %s", replica.GetReplicaIdx(), serverKey)
		}
	}

	// Update
	updatedReplicas := make(map[int]bool)
	for _, replica := range replicas {
		existingState := modelVersion.replicas[replica.GetReplicaIdx()]
		if !existingState.State.AlreadyLoadingOrLoaded() {
			logger.Debugf("Setting model %s version %d on server %s replica %d to LoadRequested", modelKey, version, serverKey, replica.GetReplicaIdx())
			modelVersion.replicas[replica.GetReplicaIdx()] = ReplicaStatus{State: LoadRequested}
		} else {
			logger.Debugf("model %s on server %s replica %d already loaded", modelKey, serverKey, replica.GetReplicaIdx())
		}
		updatedReplicas[replica.GetReplicaIdx()] = true
	}
	for replicaIdx, existingState := range modelVersion.replicas {
		logger.Debugf("Looking at replicaidx %d with state %s but ignoring processed %v", replicaIdx, existingState.State.String(), updatedReplicas)
		if _, ok := updatedReplicas[replicaIdx]; !ok {
			if !existingState.State.AlreadyUnloadingOrUnloaded() {
				logger.Debugf("Setting model %s version %d on server %s replica %d to UnloadRequested", modelKey, version, serverKey, replicaIdx)
				modelVersion.replicas[replicaIdx] = ReplicaStatus{State: UnloadRequested}
			} else {
				logger.Debugf("model %s on server %s replica %d already unloaded", modelKey, serverKey, replicaIdx)
			}
		}
	}
	modelVersion.server = serverKey
	latestModel := model.Latest()
	isLatest := latestModel.GetVersion() == modelVersion.GetVersion()
	m.updateModelStatus(isLatest, model.IsDeleted(), modelVersion, model.GetLastAvailableModelVersion())
	return nil
}

func (m *MemoryStore) UpdateModelState(modelKey string, version uint32, serverKey string, replicaIdx int, availableMemory *uint64, state ModelReplicaState, reason string) error {
	logger := m.logger.WithField("func", "UpdateModelState")
	m.mu.Lock()
	defer m.mu.Unlock()

	// Validate
	model, modelVersion, server, err := m.getModelServer(modelKey, version, serverKey)
	if err != nil {
		return err
	}

	modelVersion.replicas[replicaIdx] = ReplicaStatus{State: state, Reason: reason, Timestamp: time.Now()}
	logger.Debugf("Setting model %s version %d on server %s replica %d to %s", modelKey, version, serverKey, replicaIdx, state.String())
	// Update models loaded onto replica if loaded or unloaded is state
	if state == Loaded || state == Unloaded {
		server, ok := m.store.servers[serverKey]
		if ok {
			replica, ok := server.replicas[replicaIdx]
			if ok {
				if state == Loaded {
					replica.loadedModels[modelKey] = true
				} else {
					delete(replica.loadedModels, modelKey)
				}
			}
		}
	}
	if availableMemory != nil {
		server.replicas[replicaIdx].availableMemory = *availableMemory
	}
	latestModel := model.Latest()
	isLatest := latestModel.GetVersion() == modelVersion.GetVersion()

	m.updateModelStatus(isLatest, model.IsDeleted(), modelVersion, model.GetLastAvailableModelVersion())
	logger.Infof("Model %s deleted %v active %v", modelKey, model.IsDeleted(), model.Inactive())
	if model.IsDeleted() && model.Inactive() { //TODO probably needs further work when versions of models correctly handled
		logger.Debugf("Deleting model %s as inactive", modelKey)
		delete(m.store.models, modelKey)
	}
	return nil
}

func (m *MemoryStore) AddServerReplica(request *agent.AgentSubscribeRequest) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	server, ok := m.store.servers[request.ServerName]
	if !ok {
		server = NewServer(request.ServerName, request.Shared)
		m.store.servers[request.ServerName] = server
	}
	loadedModels := make(map[string]bool)
	for _, modelVersionReq := range request.LoadedModels {
		modelVersion := m.addModelVersionIfNotExists(modelVersionReq)
		loadedModels[modelVersionReq.GetModel().GetMeta().GetName()] = true
		modelVersion.replicas[int(request.ReplicaIdx)] = ReplicaStatus{State: Loaded}
	}
	serverReplica := NewServerReplicaFromConfig(server, int(request.ReplicaIdx), loadedModels, request.ReplicaConfig, request.AvailableMemoryBytes)
	server.replicas[int(request.ReplicaIdx)] = serverReplica
	return nil
}

func (m *MemoryStore) RemoveServerReplica(serverName string, replicaIdx int) ([]string, error) {
	server, ok := m.store.servers[serverName]
	if !ok {
		return nil, fmt.Errorf("Failed to find server %s", serverName)
	}
	serverReplica, ok := server.replicas[replicaIdx]
	if !ok {
		return nil, fmt.Errorf("Failed to find replica %d for server %s", replicaIdx, serverName)
	}
	delete(server.replicas, replicaIdx)
	if len(server.replicas) == 0 {
		delete(m.store.servers, serverName)
	}
	var modelNames []string
	for modelName := range serverReplica.loadedModels {
		model, ok := m.store.models[modelName]
		if ok {
			latestVersion := model.Latest()
			if latestVersion != nil {
				//latestVersion.replicas[replicaIdx] = ModelReplicaStateUnknown
				delete(latestVersion.replicas, replicaIdx)
			}
		}
		modelNames = append(modelNames, modelName)
	}
	return modelNames, nil
}
