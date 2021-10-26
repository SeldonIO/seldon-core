package store

import (
	"errors"
	"fmt"
	pba "github.com/seldonio/seldon-core/scheduler/apis/mlops/agent"
	pb "github.com/seldonio/seldon-core/scheduler/apis/mlops/scheduler"
	log "github.com/sirupsen/logrus"
	"math/rand"
	"sort"
	"sync"
)

var (
	DecreasedCapabilitiesErr = errors.New("Decreased capabilities")
	DecreasedMemoryErr = errors.New("Decreased memory")
	OverCommitChangeErr = errors.New("Overcommit change")
	ModelNotFoundErr = errors.New("Model not found")
	ModelExistsErr = errors.New("Model exists")
	ServerNotFoundErr = errors.New("Server not found")
	ServerReplicaNotFoundErr = errors.New("Server replica not found")
	ServerReplicaAlreadyExistsErr = errors.New("Server replica already exists")
	RescheduleModelsFromReplicaErr = errors.New("Rechesule of models from replica failed")
	ModelServerIncompatibleErr = errors.New("Model and server are incompatible")
)

type MemorySchedulerStore struct {
	mu     sync.RWMutex
	store  *LocalSchedulerStore
	logger log.FieldLogger
	agentChan chan string
	envoyChan chan string
}



func NewMemoryScheduler(logger log.FieldLogger, agentChan chan string, envoyChan chan string) *MemorySchedulerStore {
	return &MemorySchedulerStore{
		store:  NewLocalSchedulerStore(),
		logger: logger,
		agentChan: agentChan,
		envoyChan: envoyChan,
	}
}

func (m *MemorySchedulerStore) syncAgentAndEnvoy(modelName string) {
	m.agentChan <- modelName
	m.envoyChan <- modelName
}

func createServerReplicaFromProto(replicaConfig *pba.ReplicaConfig, replicaIdx int, server *Server) *ServerReplica {
	 return &ServerReplica{
		inferenceSvc: replicaConfig.GetInferenceSvc(),
		inferencePort: replicaConfig.GetInferencePort(),
		replicaIdx: replicaIdx,
		server: server,
		memory: replicaConfig.GetMemory(),
		capabilities: replicaConfig.GetCapabilities(),
		availableMemory: replicaConfig.GetMemory(),
		loadedModels: make(map[string]bool),
		overCommit: replicaConfig.GetOverCommit(),
	}
}

// C2s must be superset of c1s
func extendedCapabilities(c1s []string, c2s []string) bool {
	for _,c1 := range c1s {
		found := false
		for _,c2 := range c2s {
			if c1 == c2 {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func (m *MemorySchedulerStore) UpdateServerReplica(request *pba.AgentSubscribeRequest) error {
	err := m.updateServerReplicaInternal(request)
	if err != nil {
		return err
	}
	m.attemptRescheduleFailedModels()
	return nil
}

func (m *MemorySchedulerStore) updateServerReplicaInternal(request *pba.AgentSubscribeRequest) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	serverName := request.GetServerName()
	replicaIdx := int(request.ReplicaIdx)
	replicaConfig := request.ReplicaConfig
	server, ok := m.store.servers[serverName]
	if !ok {
		// Create Server and new replica
		server = &Server{
			name: serverName,
			replicas: make(map[int]*ServerReplica),
		}
		server.replicas[replicaIdx] = createServerReplicaFromProto(replicaConfig, replicaIdx, server)
		m.store.servers[serverName] = server
	} else {
		// Update Server and new replica
		_, ok := server.replicas[replicaIdx]
		if !ok {
			// Create replica
			server.replicas[replicaIdx] = createServerReplicaFromProto(replicaConfig, replicaIdx, server)
		} else {
			return fmt.Errorf("Server replica already exists %s:%d. %w",serverName, replicaIdx, ServerReplicaAlreadyExistsErr)
		}
	}
	// Add loaded models
	replica := server.replicas[replicaIdx]
	for _,details := range request.LoadedModels {
		model, ok := m.store.models[details.Name]
		if !ok {
			//Add model
			model = &Model{
				config: details,
				replicas: make(map[int]ModelState),
			}
			m.store.models[details.Name] = model
		}
		if model.config.Version != details.Version {
			m.logger.Infof("Old version of model %s current %s agent supplied %s",details.Name,model.config.Version, details.Version)
			model.replicas[replicaIdx] = UnloadRequested
			model.server = server.name
		} else {
			replica.loadedModels[model.Key()] = true
			model.replicas[replicaIdx] = Loaded
			model.server = server.name
		}

	}
	return nil
}

func (m *MemorySchedulerStore) attemptRescheduleFailedModels()  {
	for _,key := range m.getFailedModels() {
		model, ok := m.store.models[key]
		if !ok {
			m.logger.Errorf("Failed to find model in failed model list with key %s",key)
			m.removeFailedModel(key)
			continue
		}
		if model.config.Server != nil {
			err := m.UpdateModelOnServer(model.Key(), model.config.GetServer())
			if err != nil {
				m.logger.WithError(err).Errorf("Failed to reschedule model %s to required server %s",key, model.config.GetServer())
				continue
			}
		} else if model.HasServer() {
			err := m.UpdateModelOnServer(model.Key(), model.Server())
			if err != nil {
				m.logger.WithError(err).Errorf("Failed to reschedule model %s to server %s",key,model.Server())
				continue
			}
		} else {
			err := m.ScheduleModelToServer(key)
			if err != nil {
				m.logger.WithError(err).Errorf("Failed to reschedule model %s",key)
				continue
			}
		}
		m.removeFailedModel(key)
	}
}

func (m *MemorySchedulerStore) getFailedModels() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var models []string
	for model := range m.store.failedToScheduleModels {
		models = append(models, model)
	}
	return models
}

func (m *MemorySchedulerStore) addFailedModel(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.store.failedToScheduleModels[key] = true
}

func (m *MemorySchedulerStore) removeFailedModel(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.store.failedToScheduleModels,key)
}

func (m *MemorySchedulerStore) isFailedModel(key string) bool {
	_, ok := m.store.failedToScheduleModels[key]
	return ok
}

func (m *MemorySchedulerStore) RemoveServerReplicaAndRedeployModels(serverName string, replicaIdx int) error {
	modelsNeedingRedploy, err := m.removeServerReplica(serverName, replicaIdx)
	if err != nil {
		return err
	}
	rescheduleFailed := false
	for _,modelName := range modelsNeedingRedploy {
		err := m.UpdateModelOnServer(modelName, serverName)
		if err != nil {
			m.logger.WithError(err).Errorf("Failed to update model %s on server %s after remove of replica %d", modelName, serverName, replicaIdx)
			m.addFailedModel(modelName)
			rescheduleFailed = true
		}
	}
	if rescheduleFailed {
		return RescheduleModelsFromReplicaErr
	}
	return nil
}

func (m *MemorySchedulerStore) removeServerReplica(serverName string, replicaIdx int) ([]string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var modelsNeedingRedeploy []string
	server, ok := m.store.servers[serverName]
	if !ok {
		return modelsNeedingRedeploy, fmt.Errorf("No server with key server %s", serverName)
	}
	serverReplica, ok := server.replicas[replicaIdx]
	if !ok {
		return modelsNeedingRedeploy, fmt.Errorf("No server replica with index %d for server %s", replicaIdx, serverName)
	}
	// If models deployed on server replica decide what to do

	for k,_ := range serverReplica.loadedModels  {
		model, ok := m.store.models[k]
		if !ok {
			m.logger.Errorf("Failed to find model %s when removing replica %s:%d",k,serverName,replicaIdx)
			continue
		}
		if model.isLiveReplica(replicaIdx) {
			modelsNeedingRedeploy = append(modelsNeedingRedeploy, k)
		}
		delete(model.replicas,replicaIdx)
	}
	delete(server.replicas, replicaIdx)
	if len(server.replicas) == 0 && len(modelsNeedingRedeploy) == 0 {
		m.logger.Infof("Deleting server %s as no replicas left",serverName)
		delete(m.store.servers, serverName)
	}
	return modelsNeedingRedeploy, nil
}


func (m *MemorySchedulerStore) GetServer(key string) (*Server, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	server, ok := m.store.servers[key]
	if !ok {
		return nil, fmt.Errorf("Server with key %s does not exist %w",key, ServerNotFoundErr)
	}
	return server, nil
}


func (m*MemorySchedulerStore) GetServerReplica(key string, replicaIdx int) (*ServerReplica, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	server, ok := m.store.servers[key]
	if !ok {
		return nil, fmt.Errorf("Server with key %s does not exist. %w",key, ServerReplicaNotFoundErr)
	}
	serverReplica, ok := server.replicas[replicaIdx]
	if !ok {
		return nil, fmt.Errorf("Server with key %s and replica %d does not exist. %w",key, replicaIdx, ServerReplicaNotFoundErr)
	}
	return serverReplica, nil
}

func checkModelRequirements(capabilities []string, requirements []string, availableMemory uint64, memory uint64) bool {
	for _,requirement := range requirements {
		requirementFound := false
		for _, capability := range capabilities {
			if requirement == capability {
				requirementFound = true
				break
			}
		}
		if !requirementFound {
			return false
		}
	}
	if memory <= availableMemory {
		return true
	}
	return false
}

func checkMatchingReplicas(server *Server, requirements []string, replicas uint32, memory uint64) bool {
	numMatchingReplicas := uint32(0)
	for _,serverReplica := range server.replicas{
		if checkModelRequirements(serverReplica.capabilities, requirements, serverReplica.availableMemory, memory) {
			numMatchingReplicas++
		}
	}
	if numMatchingReplicas >= replicas {
		return true
	}
	return false
}

func  getMatchingServers(servers map[string]*Server, requirements []string, replicas uint32, memory uint64) []string {
	var foundServers []string
	for serverName, server := range servers {
		if checkMatchingReplicas(server, requirements, replicas, memory) {
			foundServers = append(foundServers, serverName)
		}
	}
	sort.Strings(foundServers)
	return foundServers
}

func (m *MemorySchedulerStore) CreateModel(key string, config *pb.ModelDetails) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, ok := m.store.models[key]
	if ok {
		return fmt.Errorf("Model with key %s already exists. %w", key, ModelExistsErr)
	}
	m.store.models[key] = &Model{
		config: config,
		replicas: make(map[int]ModelState),
	}
	return nil
}

func (m *MemorySchedulerStore) UpdateModel(key string, config *pb.ModelDetails) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	model, ok := m.store.models[key]
	if !ok {
		return fmt.Errorf("Model with key %s does not exist. %w", key, ModelNotFoundErr)
	}
	model.config = config
	return nil
}

func (m *MemorySchedulerStore) GetModel(key string) (*Model, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	model, ok := m.store.models[key]
	if !ok {
		return nil, fmt.Errorf("Model with key %s does not exist. %w",key, ModelNotFoundErr)
	}
	return model, nil
}

func (m *MemorySchedulerStore) RemoveModel(key string) error {
	m.mu.Lock()
	defer m.syncAgentAndEnvoy(key)
	defer m.mu.Unlock()
	model, ok := m.store.models[key]
	if !ok {
		return fmt.Errorf("Model with key %s does not exist. %w", key, ModelNotFoundErr)
	}
	for k := range model.replicas {
		model.replicas[k] = UnloadRequested
	}
	model.deleted = true
	return nil
}

// "Atomic" assignment of replica updates for model
// In future would call external store and possibly fail
func (m *MemorySchedulerStore) addModelToServerReplicaMemory(model *Model, server *Server, replicas []int) error {
	for _,replicaIdx := range replicas {
		replica, ok  := server.replicas[replicaIdx]
		if !ok {
			return fmt.Errorf("Failed to find replica %s:%d. %w",server.name, replicaIdx, ServerReplicaNotFoundErr)
		}
		if model.config.Memory != nil { // can be an estimate as will be updated by agent on next load/unload to true value
			if model.config.GetMemory() > replica.availableMemory {
				return fmt.Errorf("Not enough memory on replica for model %s available %d requested %d",model.Key(), replica.availableMemory, model.config.GetMemory())
			}
			replica.availableMemory = replica.availableMemory - *model.config.Memory
		}
		replica.loadedModels[model.Key()] = true
		model.replicas[replicaIdx] = LoadRequested
		model.server = server.name
	}
	return nil
}

var (
	FailedSchedulingErr = errors.New("Failed to schedule model")
)

func (m *MemorySchedulerStore) ScheduleModelToServer(modelKey string) error{
	m.mu.RLock()
	model, ok := m.store.models[modelKey]
	if !ok {
		m.mu.RUnlock()
		return fmt.Errorf("Model with key %s does not exist. %w",modelKey, ModelNotFoundErr)
	}
	possibleServers := getMatchingServers(m.store.servers, model.config.Requirements, model.config.Replicas, model.config.GetMemory())
	m.mu.RUnlock()
	for _, server := range possibleServers {
		err := m.UpdateModelOnServer(modelKey, server)
		if err != nil {
			m.logger.WithError(err).Errorf("Failed to assign model %s to server %s",modelKey, server)
			continue
		}
		return nil
	}
	return fmt.Errorf("Failed to find server for model %s from %v %w",modelKey, possibleServers, FailedSchedulingErr)

}

func (m *MemorySchedulerStore) UpdateModelOnServer(modelKey string, serverKey string) error {
	m.mu.Lock()
	defer m.syncAgentAndEnvoy(modelKey)
	defer m.mu.Unlock()

	model, ok := m.store.models[modelKey]
	if !ok {
		return fmt.Errorf("Model with key %s does not exist. %w",modelKey, ModelNotFoundErr)
	}
	server, ok := m.store.servers[serverKey]
	if !ok {
		return fmt.Errorf("Server with key %s does not exist. %w",serverKey, ServerNotFoundErr)
	}
	desiredReplicas := model.config.Replicas

	if !checkMatchingReplicas(server, model.config.Requirements, desiredReplicas, model.config.GetMemory()) {
		return fmt.Errorf("Model %s can't be scheduled on server %s. %w",modelKey,serverKey, ModelServerIncompatibleErr)
	}

	serverReplicas := server.maxReplicas()
	if serverReplicas < desiredReplicas {
		return fmt.Errorf("Server does not have enough replicas: server %d, requested %d",serverReplicas,desiredReplicas)
	}
	currentReplicas := model.NumActiveReplicas()
	assigned := uint32(0)
	if currentReplicas == desiredReplicas {
		return nil
	} else if currentReplicas < desiredReplicas {
		// change any unloading models to loadRequested
		var replicasChosen []int
		// Add new server replicas
		randomServerOrdering := rand.Perm(int(serverReplicas))
		for _,replicaIdx := range randomServerOrdering {
			serverReplica, ok := server.replicas[replicaIdx]
			if ok {
				state,ok := model.replicas[replicaIdx]
				m.logger.Infof("Scheduling %s replicaIdx:%d state:%s ok:%s",modelKey,replicaIdx,state,ok)
				if !ok || state.CanLoad() {
					if checkModelRequirements(serverReplica.capabilities,model.config.GetRequirements(), serverReplica.availableMemory, model.config.GetMemory()) {
						replicasChosen = append(replicasChosen, replicaIdx)
						assigned++
						if assigned + currentReplicas == desiredReplicas {
							return m.addModelToServerReplicaMemory(model, server, replicasChosen)
						}
					}
				}
			} else {
				m.logger.Warnf("Replica %d does not exist and is skipped for server %s during scheduling for model %s",replicaIdx, serverKey, modelKey)
			}
		}
		return fmt.Errorf("Failed scale up model")
	} else {
		toBeRemoved := currentReplicas - desiredReplicas
		removed := uint32(0)
		for k,currentState := range model.replicas {
			if currentState.CanUnload() {
				model.replicas[k] = UnloadRequested
				removed++
				if removed == toBeRemoved {
					return nil
				}
			}
		}
		return fmt.Errorf("Failed scale down model")
	}
}


func (m *MemorySchedulerStore) SetModelState(modelKey string, serverKey string, replicaIdx int, state ModelState, availableMemory *uint64) error {
	model, err := m.GetModel(modelKey)
	if err != nil {
		return err
	}
	server, err := m.GetServer(serverKey)
	if err != nil {
		return err
	}
	m.mu.Lock()
	defer m.syncAgentAndEnvoy(modelKey)
	defer m.mu.Unlock()
	serverReplica, ok := server.replicas[replicaIdx]
	if !ok {
		return fmt.Errorf("Server replica %d does not exist for server with key %s when adding model with key %s",replicaIdx,serverKey,modelKey)
	}
	if availableMemory != nil {
		serverReplica.availableMemory = *availableMemory
	}
	model.replicas[replicaIdx] = state

	// Remove model if all replicas are inactive and marked for deletion
	if model.CanRemove() {
		delete(m.store.models, modelKey)
	}
	return nil
}






