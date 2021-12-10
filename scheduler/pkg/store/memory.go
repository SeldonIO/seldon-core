package store

import (
	"fmt"
	"sync"
	"time"

	"github.com/seldonio/seldon-core/scheduler/apis/mlops/agent"
	pb "github.com/seldonio/seldon-core/scheduler/apis/mlops/scheduler"
	log "github.com/sirupsen/logrus"
)

type MemoryStore struct {
	mu     sync.RWMutex
	store  *LocalSchedulerStore
	logger log.FieldLogger
	modelEventListeners  []chan<- string
}

func NewMemoryStore(logger log.FieldLogger, store *LocalSchedulerStore) *MemoryStore {
	return &MemoryStore{
		store:  store,
		logger: logger.WithField("source", "MemoryStore"),
	}
}

func (m *MemoryStore) AddListener(c chan string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.modelEventListeners = append(m.modelEventListeners, c)
}


func (m *MemoryStore) updateModelImpl(config *pb.ModelDetails, addAsLatest bool) (*ModelVersion, error) {
	model, ok := m.store.models[config.Name]
	if !ok {
		model = NewModel()
		m.store.models[config.Name] = model
	}
	if existingModelVersion, ok := model.versionMap[config.Version]; !ok {
		modelVersion := NewDefaultModelVersion(config)
		model.versionMap[config.Version] = modelVersion
		if addAsLatest {
			model.versions = append(model.versions, modelVersion)
		} else {
			model.versions = append([]*ModelVersion{modelVersion}, model.versions...)
		}
		return modelVersion, nil
	} else {
		if addAsLatest {
			return nil, fmt.Errorf("Model %s version %s already exists. %w", config.Name, config.Version, ModelVersionExistsErr)
		} else {
			return existingModelVersion, nil
		}

	}
}

func (m *MemoryStore) UpdateModel(config *pb.ModelDetails) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, err := m.updateModelImpl(config, true)
	return err
}

func (m *MemoryStore) GetModel(key string) (*ModelSnapshot, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	model, ok := m.store.models[key]
	if ok {
		return &ModelSnapshot{
			Name:     key,
			Versions: model.versions, //TODO make a copy for safety?
			Deleted:  model.deleted,
		}, nil
	} else {
		return &ModelSnapshot{
			Name:     key,
			Versions: nil,
		}, nil
	}
}

func (m *MemoryStore) ExistsModelVersion(key string, version string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if model, ok := m.store.models[key]; ok {
		for _,mv := range model.versions {
			if mv.Details().Version == version {
				return true
			}
		}
	}
	return false
}

func (m *MemoryStore) RemoveModel(modelKey string) error {
	model, ok := m.store.models[modelKey]
	if ok {
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

func (m *MemoryStore) getModelServer(modelKey string, version string, serverKey string) (*Model, *ModelVersion, *Server, error) {
	// Validate
	model, ok := m.store.models[modelKey]
	if !ok {
		return nil, nil, nil, fmt.Errorf("failed to find model %s", modelKey)
	}
	modelVersion := model.Latest()
	if modelVersion == nil {
		return nil, nil, nil, fmt.Errorf("No latest version for model %s", modelKey)
	}
	if modelVersion.config.Version != version {
		return nil, nil, nil, fmt.Errorf("Model version is not matching. Found %s but was trying to update %s. %w", modelVersion.config.Version, version, ModelNotLatestVersionRejectErr)
	}
	server, ok := m.store.servers[serverKey]
	if !ok {
		return nil, nil, nil, fmt.Errorf("failed to find server %s", serverKey)
	}
	return model, modelVersion, server, nil
}

func (m *MemoryStore) UpdateLoadedModels(modelKey string, version string, serverKey string, replicas []*ServerReplica) error {
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
			logger.Debugf("Setting model %s on server %s replica %d to LoadRequested", modelKey, serverKey, replica.GetReplicaIdx())
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
				logger.Debugf("Setting model %s on server %s replica %d to UnloadRequested", modelKey, serverKey, replicaIdx)
				modelVersion.replicas[replicaIdx] = ReplicaStatus{State: UnloadRequested}
			} else {
				logger.Debugf("model %s on server %s replica %d already unloaded", modelKey, serverKey, replicaIdx)
			}
		}
	}
	modelVersion.server = serverKey
	m.updateModelStatus(model.isDeleted(), modelVersion, model.Previous())
	return nil
}

func (m *MemoryStore) UpdateModelState(modelKey string, version string, serverKey string, replicaIdx int, availableMemory *uint64, state ModelReplicaState, reason string) error {
	logger := m.logger.WithField("func", "UpdateModelState")
	m.mu.Lock()
	defer m.mu.Unlock()

	// Validate
	model, modelVersion, server, err := m.getModelServer(modelKey, version, serverKey)
	if err != nil {
		return err
	}

	modelVersion.replicas[replicaIdx] = ReplicaStatus{State: state, Reason: reason, Timestamp: time.Now()}
	logger.Debugf("Setting model %s version %s on server %s replica %d to %s", modelKey, version, serverKey, replicaIdx, state.String())
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
	m.updateModelStatus(model.isDeleted(), modelVersion, model.Previous())
	logger.Infof("Model %s deleted %v active %v",modelKey, model.isDeleted(), model.Inactive())
	if model.isDeleted() && model.Inactive() {
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
	for _, modelConfig := range request.LoadedModels {
		modelVersion, err := m.updateModelImpl(modelConfig, false)
		if err != nil {
			return err
		}
		loadedModels[modelConfig.Name] = true
		modelVersion.replicas[int(request.ReplicaIdx)] = ReplicaStatus{State: Loaded}
	}
	serverReplica := NewServerReplicaFromConfig(server, int(request.ReplicaIdx), loadedModels, request.ReplicaConfig)
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

func (m *MemoryStore) updateModelStatus(deleted bool, modelVersion *ModelVersion, prevModelVersion *ModelVersion)  {
	logger := m.logger.WithField("func","updateModelStatus")
	var replicasAvailable, replicasLoading, replicasLoadFailed, replicasUnloading, replicasUnloaded, replicasUnloadFailed  uint32
	var lastFailedReason string
	lastFailedStateTime := time.Time{}
	latestTime := time.Time{}
	for _,replicaState := range modelVersion.ReplicaState() {
		switch replicaState.State {
		case Available:
			replicasAvailable++
		case LoadRequested, Loading, Loaded: // unavailable but OK
			replicasLoading++
		case LoadFailed, LoadedUnavailable: // unavailable but not OK
			replicasLoadFailed++
			if !deleted && replicaState.Timestamp.After(lastFailedStateTime) {
				lastFailedStateTime = replicaState.Timestamp
				lastFailedReason = replicaState.Reason
			}
		case UnloadRequested, Unloading:
			replicasUnloading++
		case Unloaded:
			replicasUnloaded++
		case UnloadFailed:
			replicasUnloadFailed++
			if deleted && replicaState.Timestamp.After(lastFailedStateTime) {
				lastFailedStateTime = replicaState.Timestamp
				lastFailedReason = replicaState.Reason
			}
		}
		if replicaState.Timestamp.After(latestTime) {
			latestTime = replicaState.Timestamp
		}
	}
	var modelState ModelState
	var modelReason string
	modelTimestamp := latestTime
	logger.Infof("Model details %v, replicasAvailable %d, deleted %v, prev model %v",modelVersion.Details(),replicasAvailable,deleted,prevModelVersion)
	if deleted {
		if replicasUnloadFailed > 0 {
			modelState = ModelTerminateFailed
			modelReason = lastFailedReason
			modelTimestamp = lastFailedStateTime
		}  else if replicasUnloading > 0 || replicasAvailable > 0{
			modelState = ModelTerminating
		} else {
			modelState = ModelTerminated
		}
	} else {
		if replicasLoadFailed > 0 {
			modelState = ModelFailed
			modelReason = lastFailedReason
			modelTimestamp = lastFailedStateTime
		} else if (modelVersion.Details() != nil && replicasAvailable == modelVersion.Details().Replicas && prevModelVersion == nil) ||
			(replicasAvailable > 0 && prevModelVersion != nil && prevModelVersion.state.State == ModelAvailable) { //TODO In future check if available replicas is > minReplicas
			modelState = ModelAvailable
		} else {
			modelState = ModelProgressing
		}
	}
	modelVersion.state = ModelStatus{
		State:               modelState,
		Reason:              modelReason,
		Timestamp:           modelTimestamp,
		AvailableReplicas:   replicasAvailable,
		UnavailableReplicas: replicasLoading,
	}
	for _,listener := range m.modelEventListeners {
		listener <- modelVersion.Details().Name
	}
}