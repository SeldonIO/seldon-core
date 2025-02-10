/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package store

import (
	"fmt"
	"sort"
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/seldonio/seldon-core/apis/go/v2/mlops/agent"
	pb "github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/coordinator"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store/utils"
)

type MemoryStore struct {
	mu       sync.RWMutex
	opLocks  sync.Map
	store    *LocalSchedulerStore
	logger   log.FieldLogger
	eventHub *coordinator.EventHub
}

func NewMemoryStore(
	logger log.FieldLogger,
	store *LocalSchedulerStore,
	eventHub *coordinator.EventHub,
) *MemoryStore {
	return &MemoryStore{
		store:    store,
		logger:   logger.WithField("source", "MemoryStore"),
		eventHub: eventHub,
	}
}

func (m *MemoryStore) GetAllModels() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var modelNames []string
	for modelName := range m.store.models {
		modelNames = append(modelNames, modelName)
	}
	return modelNames
}

func (m *MemoryStore) GetModels() ([]*ModelSnapshot, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	foundModels := []*ModelSnapshot{}
	for name, model := range m.store.models {
		snapshot := &ModelSnapshot{
			Name:     name,
			Deleted:  model.IsDeleted(),
			Versions: model.versions,
		}
		foundModels = append(foundModels, snapshot)
	}
	return foundModels, nil
}

func (m *MemoryStore) addModelVersionIfNotExists(req *agent.ModelVersion) (*Model, *ModelVersion) {
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
		return model, modelVersion
	} else {
		return model, existingModelVersion
	}
}

func (m *MemoryStore) addNextModelVersion(model *Model, pbmodel *pb.Model) {
	// if we start from a clean state, lets use the generation id as the starting version
	// this is to ensure that we have monotonic increasing version numbers
	// and we never reset back to 1
	generation := pbmodel.GetMeta().GetKubernetesMeta().GetGeneration()
	version := max(uint32(1), uint32(generation))
	if model.Latest() != nil {
		version = model.Latest().GetVersion() + 1
	}
	modelVersion := NewDefaultModelVersion(pbmodel, version)

	model.versions = append(model.versions, modelVersion)
	sort.SliceStable(model.versions, func(i, j int) bool { // resort model versions based on version number
		return model.versions[i].GetVersion() < model.versions[j].GetVersion()
	})
}

func (m *MemoryStore) UpdateModel(req *pb.LoadModelRequest) error {
	logger := m.logger.WithField("func", "UpdateModel")
	m.mu.Lock()
	defer m.mu.Unlock()
	modelName := req.GetModel().GetMeta().GetName()
	validName := utils.CheckName(modelName)
	if !validName {
		return fmt.Errorf(
			"Model %s does not have a valid name - it must be alphanumeric and not contains dots (.)",
			modelName,
		)
	}
	model, ok := m.store.models[modelName]
	if !ok {
		model = &Model{}
		m.store.models[modelName] = model
		m.addNextModelVersion(model, req.GetModel())
	} else if model.IsDeleted() {
		if model.Inactive() {
			model = &Model{}
			m.store.models[modelName] = model
			m.addNextModelVersion(model, req.GetModel())
		} else {
			return fmt.Errorf(
				"Model %s is in process of deletion - new model can not be created",
				modelName,
			)
		}
	} else {
		meq := ModelEqualityCheck(model.Latest().modelDefn, req.GetModel())
		if meq.Equal {
			logger.Debugf("Model %s semantically equal - doing nothing", modelName)
			return nil
		} else if meq.ModelSpecDiffers {
			logger.Debugf("Model %s model spec differs - adding new version of model", modelName)
			m.addNextModelVersion(model, req.GetModel())
			return nil
		} else if meq.DeploymentSpecDiffers {
			logger.Debugf(
				"Model %s deployment spec differs - updating latest model version with new spec",
				modelName,
			)
			model.Latest().SetDeploymentSpec(req.GetModel().GetDeploymentSpec())
		}
		if meq.MetaDiffers {
			// Update just kubernetes meta
			model.Latest().UpdateKubernetesMeta(req.GetModel().GetMeta().GetKubernetesMeta())
		}
	}
	return nil
}

func (m *MemoryStore) getModelImpl(key string) *ModelSnapshot {
	model, ok := m.store.models[key]
	if ok {
		return &ModelSnapshot{
			Name:     key,
			Versions: model.versions, // TODO make a copy for safety?
			Deleted:  model.IsDeleted(),
		}
	} else {
		return &ModelSnapshot{
			Name:     key,
			Versions: nil,
		}
	}
}

func (m *MemoryStore) LockModel(modelId string) {
	var lock sync.RWMutex
	existingLock, _ := m.opLocks.LoadOrStore(modelId, &lock)
	existingLock.(*sync.RWMutex).Lock()
}

func (m *MemoryStore) UnlockModel(modelId string) {
	logger := m.logger.WithField("func", "UnlockModel")
	lock, loaded := m.opLocks.Load(modelId)
	if loaded {
		lock.(*sync.RWMutex).Unlock()
	} else {
		logger.Warnf("Trying to unlock model %s that was not locked.", modelId)
	}
}

func (m *MemoryStore) GetModel(key string) (*ModelSnapshot, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.getModelImpl(key), nil
}

func (m *MemoryStore) RemoveModel(req *pb.UnloadModelRequest) error {
	err := m.removeModelImpl(req)
	if err != nil {
		return err
	}
	return nil
}

func (m *MemoryStore) removeModelImpl(req *pb.UnloadModelRequest) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	model, ok := m.store.models[req.GetModel().GetName()]
	if ok {
		// Updating the k8s meta is required to be updated so status updates back (to manager)
		// will match latest generation value. Previous generation values might be ignored by manager.
		if req.GetKubernetesMeta() != nil { // k8s meta can be nil if unload is called directly using scheduler grpc api
			model.Latest().UpdateKubernetesMeta(req.GetKubernetesMeta())
		}
		model.SetDeleted()
		m.updateModelStatus(true, true, model.Latest(), model.GetLastAvailableModelVersion())
		return nil
	} else {
		return fmt.Errorf("Model %s not found", req.GetModel().GetName())
	}
}

func (m *MemoryStore) GetServers(shallow bool, modelDetails bool) ([]*ServerSnapshot, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var servers []*ServerSnapshot
	for _, server := range m.store.servers {
		servers = append(servers, server.CreateSnapshot(shallow, modelDetails))
	}
	return servers, nil
}

func (m *MemoryStore) GetServer(serverKey string, shallow bool, modelDetails bool) (*ServerSnapshot, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	server := m.store.servers[serverKey]
	if server == nil {
		return nil, fmt.Errorf("Server [%s] not found", serverKey)
	} else {
		// TODO: refactor cleanly
		snapshot := server.CreateSnapshot(shallow, modelDetails)
		if modelDetails {
			// this is a hint to the caller that the server is in a state where it can be scaled down
			snapshot.Stats = &ServerStats{
				ScaleDownFlag: m.numEmptyServerReplicas(serverKey) > 0 || m.maxNumModelReplicasForServer(serverKey) < server.NumReplicas(),
			}
		}
		return snapshot, nil
	}
}

func (m *MemoryStore) getModelServer(
	modelKey string,
	version uint32,
	serverKey string,
) (*Model, *ModelVersion, *Server, error) {
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

func (m *MemoryStore) UpdateLoadedModels(
	modelKey string,
	version uint32,
	serverKey string,
	replicas []*ServerReplica,
) error {
	m.mu.Lock()
	modelEvt, serverEvt, err := m.updateLoadedModelsImpl(modelKey, version, serverKey, replicas)
	m.mu.Unlock()
	if err != nil {
		return err
	}
	if m.eventHub != nil && modelEvt != nil {
		m.eventHub.PublishModelEvent(
			modelUpdateEventSource,
			*modelEvt,
		)
	}
	if m.eventHub != nil && serverEvt != nil {
		m.eventHub.PublishServerEvent(
			serverUpdateEventSource,
			*serverEvt,
		)
	}
	return nil
}

func (m *MemoryStore) updateLoadedModelsImpl(
	modelKey string,
	version uint32,
	serverKey string,
	replicas []*ServerReplica,
) (*coordinator.ModelEventMsg, *coordinator.ServerEventMsg, error) {
	logger := m.logger.WithField("func", "updateLoadedModelsImpl")

	// Validate
	model, ok := m.store.models[modelKey]
	if !ok {
		return nil, nil, fmt.Errorf("failed to find model %s", modelKey)
	}

	modelVersion := model.Latest()
	if version != modelVersion.GetVersion() {
		return nil, nil, fmt.Errorf(
			"Model version mismatch for %s got %d but latest version is now %d",
			modelKey, version, modelVersion.GetVersion(),
		)
	}

	if serverKey == "" {
		// nothing to do for a model that doesn't have a server, proceed with sending an event for downstream
		return &coordinator.ModelEventMsg{
			ModelName:    modelVersion.GetMeta().GetName(),
			ModelVersion: modelVersion.GetVersion(),
		}, nil, nil
	}

	server, ok := m.store.servers[serverKey]
	if !ok {
		return nil, nil, fmt.Errorf("failed to find server %s", serverKey)
	}

	assignedReplicaIds := make(map[int]struct{})
	for _, replica := range replicas {
		if _, ok := server.replicas[replica.replicaIdx]; !ok {
			return nil, nil, fmt.Errorf(
				"failed to reserve replica %d as it does not exist on server %s",
				replica.replicaIdx, serverKey,
			)
		}
		assignedReplicaIds[replica.replicaIdx] = struct{}{}
	}

	for modelVersion.HasServer() && modelVersion.Server() != serverKey {
		logger.Debugf("Adding new version as server changed to %s from %s", modelVersion.Server(), serverKey)
		m.addNextModelVersion(model, model.Latest().modelDefn)
		modelVersion = model.Latest()
	}

	//  reserve memory for existing replicas that are not already loading or loaded
	replicaStateUpdated := false
	unloadRequested := false
	for replicaIdx := range assignedReplicaIds {
		if existingState, ok := modelVersion.replicas[replicaIdx]; !ok {
			logger.Debugf(
				"Model %s version %d state %s on server %s replica %d does not exist yet and should be loaded",
				modelKey, modelVersion.version, existingState.State.String(), serverKey, replicaIdx,
			)
			modelVersion.SetReplicaState(replicaIdx, LoadRequested, "")
			m.updateReservedMemory(LoadRequested, serverKey, replicaIdx, modelVersion.GetRequiredMemory())
			replicaStateUpdated = true
		} else {
			logger.Debugf(
				"Checking if model %s version %d state %s on server %s replica %d should be loaded",
				modelKey, modelVersion.version, existingState.State.String(), serverKey, replicaIdx,
			)
			if !existingState.State.AlreadyLoadingOrLoaded() {
				modelVersion.SetReplicaState(replicaIdx, LoadRequested, "")
				m.updateReservedMemory(LoadRequested, serverKey, replicaIdx, modelVersion.GetRequiredMemory())
				replicaStateUpdated = true
			}
		}
	}

	// Unload any existing model replicas assignments that are no longer part of the replica set
	for replicaIdx, existingState := range modelVersion.ReplicaState() {
		if _, ok := assignedReplicaIds[replicaIdx]; !ok {
			logger.Debugf(
				"Checking if replicaidx %d with state %s should be unloaded",
				replicaIdx, existingState.State.String(),
			)
			if !existingState.State.UnloadingOrUnloaded() && existingState.State != Draining {
				modelVersion.SetReplicaState(replicaIdx, UnloadEnvoyRequested, "")
				replicaStateUpdated = true
				unloadRequested = true
			}
		}
	}

	// in cases where we did have a previous ScheduleFailed, we need to reflect the change here
	// this could be in the cases where we are scaling down a model and the new replica count can be all deployed
	// and always send an update for deleted models, so the operator will remove them from k8s
	// also send an update for progressing models so the operator can update the status in the case of a network glitch where the model generation has been updated
	// also send an update if the model is not yet at desired replicas, if we have partial scheduling

	// note that we use len(modelVersion.GetAssignment()) to calculate the number of replicas as the status of the model at this point might not reflect the actual number of replicas
	// in modelVersion.state.AvailableReplicas (we call updateModelStatus later)

	// TODO: the conditions here keep growing, refactor or consider a simpler check.
	if replicaStateUpdated || modelVersion.state.State == ScheduleFailed || model.IsDeleted() || modelVersion.state.State == ModelProgressing ||
		(modelVersion.state.State == ModelAvailable && len(modelVersion.GetAssignment()) < modelVersion.DesiredReplicas()) {
		logger.Debugf("Updating model status for model %s server %s", modelKey, serverKey)
		modelVersion.server = serverKey
		m.updateModelStatus(true, model.IsDeleted(), modelVersion, model.GetLastAvailableModelVersion())
		if unloadRequested {
			return &coordinator.ModelEventMsg{
					ModelName:    modelVersion.GetMeta().GetName(),
					ModelVersion: modelVersion.GetVersion(),
				},
				&coordinator.ServerEventMsg{
					ServerName:    serverKey,
					UpdateContext: coordinator.SERVER_SCALE,
				},
				nil
		} else {
			return &coordinator.ModelEventMsg{
					ModelName:    modelVersion.GetMeta().GetName(),
					ModelVersion: modelVersion.GetVersion(),
				},
				nil,
				nil
		}
	} else {
		logger.Debugf("Model status update not required for model %s server %s as no replicas were updated", modelKey, serverKey)
		return nil, nil, nil
	}
}

func (m *MemoryStore) UnloadVersionModels(modelKey string, version uint32) (bool, error) {
	evt, updated, err := m.unloadVersionModelsImpl(modelKey, version)
	if err != nil {
		return updated, err
	}
	if m.eventHub != nil && evt != nil {
		m.eventHub.PublishModelEvent(
			modelUpdateEventSource,
			*evt,
		)
	}
	return updated, nil
}

func (m *MemoryStore) unloadVersionModelsImpl(modelKey string, version uint32) (*coordinator.ModelEventMsg, bool, error) {
	logger := m.logger.WithField("func", "UnloadVersionModels")
	m.mu.Lock()
	defer m.mu.Unlock()

	// Validate
	model, ok := m.store.models[modelKey]
	if !ok {
		return nil, false, fmt.Errorf("failed to find model %s", modelKey)
	}
	modelVersion := model.GetVersion(version)
	if modelVersion == nil {
		return nil, false, fmt.Errorf("version not found for model %s, version %d", modelKey, version)
	}

	updated := false
	for replicaIdx, existingState := range modelVersion.ReplicaState() {
		if !existingState.State.UnloadingOrUnloaded() {
			logger.Debugf(
				"Setting model %s version %d on server %s replica %d to UnloadRequested was %s",
				modelKey,
				modelVersion.version,
				modelVersion.Server(),
				replicaIdx,
				existingState.State.String(),
			)
			modelVersion.SetReplicaState(replicaIdx, UnloadRequested, "")
			updated = true
		} else {
			logger.Debugf(
				"model %s on server %s replica %d already unloaded",
				modelKey, modelVersion.Server(), replicaIdx,
			)
		}
	}
	if updated {
		logger.Debugf("Calling update model status for model %s version %d", modelKey, version)
		m.updateModelStatus(false, model.IsDeleted(), modelVersion, model.GetLastAvailableModelVersion())
		return &coordinator.ModelEventMsg{
			ModelName:    modelVersion.GetMeta().GetName(),
			ModelVersion: modelVersion.GetVersion(),
		}, true, nil
	}
	return nil, false, nil
}

func (m *MemoryStore) UpdateModelState(
	modelKey string,
	version uint32,
	serverKey string,
	replicaIdx int,
	availableMemory *uint64,
	expectedState ModelReplicaState,
	desiredState ModelReplicaState,
	reason string,
	runtimeInfo *pb.ModelRuntimeInfo,
) error {
	evt, err := m.updateModelStateImpl(modelKey, version, serverKey, replicaIdx, availableMemory, expectedState, desiredState, reason, runtimeInfo)
	if err != nil {
		return err
	}
	if m.eventHub != nil && evt != nil {
		m.eventHub.PublishModelEvent(
			modelUpdateEventSource,
			*evt,
		)
	}
	return nil
}

func (m *MemoryStore) updateModelStateImpl(
	modelKey string,
	version uint32,
	serverKey string,
	replicaIdx int,
	availableMemory *uint64,
	expectedState ModelReplicaState,
	desiredState ModelReplicaState,
	reason string,
	runtimeInfo *pb.ModelRuntimeInfo,
) (*coordinator.ModelEventMsg, error) {
	logger := m.logger.WithField("func", "updateModelStateImpl")
	m.mu.Lock()
	defer m.mu.Unlock()

	// Validate
	model, modelVersion, server, err := m.getModelServer(modelKey, version, serverKey)
	if err != nil {
		return nil, err
	}

	modelVersion.UpdateRuntimeInfo(runtimeInfo)

	existingState := modelVersion.GetModelReplicaState(replicaIdx)

	if existingState != expectedState {
		return nil, fmt.Errorf(
			"State mismatch for %s:%d expected state %s but was %s when trying to move to state %s",
			modelKey, version, expectedState.String(), existingState.String(), desiredState.String(),
		)
	}

	m.updateReservedMemory(desiredState, serverKey, replicaIdx, modelVersion.GetRequiredMemory())

	if existingState != desiredState {
		latestModel := model.Latest()
		isLatest := latestModel.GetVersion() == modelVersion.GetVersion()

		modelVersion.SetReplicaState(replicaIdx, desiredState, reason)
		logger.Debugf(
			"Setting model %s version %d on server %s replica %d to %s",
			modelKey, version, serverKey, replicaIdx, desiredState.String(),
		)

		// Update models loaded onto replica for relevant state
		if desiredState == Loaded || desiredState == Loading || desiredState == Unloaded || desiredState == LoadFailed {
			server, ok := m.store.servers[serverKey]
			if ok {
				replica, ok := server.replicas[replicaIdx]
				if ok {
					if desiredState == Loaded || desiredState == Loading {
						logger.Infof(
							"Adding model %s(%d) to server %s replica %d list of loaded / loading models",
							modelKey, version, serverKey, replicaIdx,
						)
						replica.addModelVersion(modelKey, version, desiredState) // we need to distinguish between loaded and loading
					} else {
						logger.Infof(
							"Removing model %s(%d) from server %s replica %d list of loaded / loading models",
							modelKey, version, serverKey, replicaIdx,
						)
						// we could go from loaded -> unloaded or loading -> failed. in the case we have a failure then we just remove from loading
						replica.deleteModelVersion(modelKey, version)
					}
				}
			}
		}
		if availableMemory != nil {
			server.replicas[replicaIdx].availableMemory = *availableMemory
		}

		m.updateModelStatus(isLatest, model.IsDeleted(), modelVersion, model.GetLastAvailableModelVersion())
		return &coordinator.ModelEventMsg{ModelName: modelVersion.GetMeta().GetName(), ModelVersion: modelVersion.GetVersion()}, nil
	}

	return nil, nil
}

func (m *MemoryStore) updateReservedMemory(
	modelReplicaState ModelReplicaState, serverKey string, replicaIdx int, memBytes uint64,
) {
	// update reserved memory that is being used for sorting replicas
	// do we need to lock replica update?
	server, ok := m.store.servers[serverKey]
	if ok {
		replica, okReplica := server.replicas[replicaIdx]
		if okReplica {
			if modelReplicaState == LoadRequested {
				replica.UpdateReservedMemory(memBytes, true)
			} else if modelReplicaState == LoadFailed || modelReplicaState == Loaded {
				replica.UpdateReservedMemory(memBytes, false)
			}
		}
	}
}

func (m *MemoryStore) AddServerReplica(request *agent.AgentSubscribeRequest) error {
	evts, serverEvt, err := m.addServerReplicaImpl(request)
	if err != nil {
		return err
	}
	if m.eventHub != nil {
		for _, evt := range evts {
			m.eventHub.PublishModelEvent(
				modelUpdateEventSource,
				evt,
			)
		}
		m.eventHub.PublishServerEvent(
			serverUpdateEventSource,
			serverEvt,
		)
	}

	return nil
}

func (m *MemoryStore) addServerReplicaImpl(request *agent.AgentSubscribeRequest) ([]coordinator.ModelEventMsg, coordinator.ServerEventMsg, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	server, ok := m.store.servers[request.ServerName]
	if !ok {
		server = NewServer(request.ServerName, request.Shared)
		m.store.servers[request.ServerName] = server
	}
	server.shared = request.Shared

	loadedModels := toSchedulerLoadedModels(request.LoadedModels)

	serverReplica := NewServerReplicaFromConfig(
		server,
		int(request.ReplicaIdx),
		loadedModels,
		request.ReplicaConfig,
		request.AvailableMemoryBytes,
	)
	server.replicas[int(request.ReplicaIdx)] = serverReplica

	var evts []coordinator.ModelEventMsg
	for _, modelVersionReq := range request.LoadedModels {
		model, modelVersion := m.addModelVersionIfNotExists(modelVersionReq)
		modelVersion.replicas[int(request.ReplicaIdx)] = ReplicaStatus{State: Loaded}
		modelVersion.server = request.ServerName
		m.updateModelStatus(true, false, modelVersion, model.GetLastAvailableModelVersion())
		evts = append(evts, coordinator.ModelEventMsg{
			ModelName:    modelVersion.GetMeta().GetName(),
			ModelVersion: modelVersion.GetVersion(),
		})
	}

	serverEvt := coordinator.ServerEventMsg{
		ServerName:    request.ServerName,
		ServerIdx:     uint32(request.ReplicaIdx),
		UpdateContext: coordinator.SERVER_REPLICA_CONNECTED,
	}

	return evts, serverEvt, nil
}

func (m *MemoryStore) RemoveServerReplica(serverName string, replicaIdx int) ([]string, error) {
	models, evts, err := m.removeServerReplicaImpl(serverName, replicaIdx)
	if err != nil {
		return nil, err
	}
	if m.eventHub != nil {
		for _, evt := range evts {
			m.eventHub.PublishModelEvent(
				modelUpdateEventSource,
				evt,
			)
		}
	}
	return models, nil
}

func (m *MemoryStore) removeServerReplicaImpl(serverName string, replicaIdx int) ([]string, []coordinator.ModelEventMsg, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	server, ok := m.store.servers[serverName]
	if !ok {
		return nil, nil, fmt.Errorf("Failed to find server %s", serverName)
	}
	serverReplica, ok := server.replicas[replicaIdx]
	if !ok {
		return nil, nil, fmt.Errorf("Failed to find replica %d for server %s", replicaIdx, serverName)
	}
	delete(server.replicas, replicaIdx)
	// TODO we should not reschedule models on servers with dedicated models, e.g. non shareable servers
	if len(server.replicas) == 0 {
		delete(m.store.servers, serverName)
	}
	loadedModelsRemoved, loadedEvts := m.removeModelfromServerReplica(serverReplica.loadedModels, replicaIdx)
	loadingModelsRemoved, loadingEtvs := m.removeModelfromServerReplica(serverReplica.loadingModels, replicaIdx)

	modelsRemoved := append(loadedModelsRemoved, loadingModelsRemoved...)
	evts := append(loadedEvts, loadingEtvs...)

	return modelsRemoved, evts, nil
}

func (m *MemoryStore) removeModelfromServerReplica(lModels map[ModelVersionID]bool, replicaIdx int) ([]string, []coordinator.ModelEventMsg) {
	logger := m.logger.WithField("func", "RemoveServerReplica")
	var modelNames []string
	var evts []coordinator.ModelEventMsg
	// Find models to reschedule due to this server replica being removed
	for modelVersionID := range lModels {
		model, ok := m.store.models[modelVersionID.Name]
		if ok {
			modelVersion := model.GetVersion(modelVersionID.Version)
			if modelVersion != nil {
				modelVersion.DeleteReplica(replicaIdx)
				if model.IsDeleted() || model.Latest().GetVersion() != modelVersion.GetVersion() {
					// In some cases we found that the user can ask for a model to be deleted and the model replica
					// is still in the process of being loaded. In this case we should not reschedule the model.
					logger.Debugf(
						"Model %s is being deleted and server replica %d is disconnected, skipping",
						modelVersionID.Name, replicaIdx,
					)
					modelVersion.SetReplicaState(replicaIdx, Unloaded, "model is removed when server replica was removed")
					m.LockModel(modelVersionID.Name)
					m.updateModelStatus(
						model.Latest().GetVersion() == modelVersion.GetVersion(),
						model.IsDeleted(), modelVersion, model.GetLastAvailableModelVersion())
					m.UnlockModel(modelVersionID.Name)
					// send an event to progress the deletion
					evts = append(
						evts,
						coordinator.ModelEventMsg{
							ModelName:    modelVersion.GetMeta().GetName(),
							ModelVersion: modelVersion.GetVersion(),
						},
					)
				} else {
					modelNames = append(modelNames, modelVersionID.Name)
				}
			} else {
				logger.Warnf("Can't find model version %s", modelVersionID.String())
			}
		}
	}
	return modelNames, evts
}

func (m *MemoryStore) DrainServerReplica(serverName string, replicaIdx int) ([]string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.drainServerReplicaImpl(serverName, replicaIdx)
}

func (m *MemoryStore) drainServerReplicaImpl(serverName string, replicaIdx int) ([]string, error) {
	logger := m.logger.WithField("func", "DrainServerReplica")
	server, ok := m.store.servers[serverName]
	if !ok {
		return nil, fmt.Errorf("Failed to find server %s", serverName)
	}
	serverReplica, ok := server.replicas[replicaIdx]
	if !ok {
		return nil, fmt.Errorf("Failed to find replica %d for server %s", replicaIdx, serverName)
	}

	// we mark this server replica as draining so should not be used in future scheduling decisions
	serverReplica.SetIsDraining()

	var modelNames []string
	// Find models to reschedule due to this server replica being removed
	for modelVersionID := range serverReplica.loadedModels {
		model, ok := m.store.models[modelVersionID.Name]
		if ok {
			modelVersion := model.GetVersion(modelVersionID.Version)
			if modelVersion != nil {
				modelVersion.SetReplicaState(replicaIdx, Draining, "trigger to drain")
				modelNames = append(modelNames, modelVersionID.Name)
			} else {
				logger.Warnf("Can't find model version %s", modelVersionID.String())
			}
		}
	}
	return modelNames, nil
}

func (m *MemoryStore) ServerNotify(request *pb.ServerNotify) error {
	logger := m.logger.WithField("func", "MemoryServerNotify")
	m.mu.Lock()
	defer m.mu.Unlock()

	logger.Debugf("ServerNotify %v", request)

	server, ok := m.store.servers[request.Name]
	if !ok {
		server = NewServer(request.Name, request.Shared)
		m.store.servers[request.Name] = server
	}
	server.SetExpectedReplicas(int(request.ExpectedReplicas))
	server.SetMinReplicas(int(request.MinReplicas))
	server.SetMaxReplicas(int(request.MaxReplicas))
	server.SetKubernetesMeta(request.KubernetesMeta)
	return nil
}

func (m *MemoryStore) numEmptyServerReplicas(serverName string) uint32 {
	emptyReplicas := uint32(0)
	server, ok := m.store.servers[serverName]
	if !ok {
		return emptyReplicas
	}
	for _, replica := range server.replicas {
		if len(replica.GetLoadedOrLoadingModelVersions()) == 0 {
			emptyReplicas++
		}
	}
	return emptyReplicas
}

func (m *MemoryStore) maxNumModelReplicasForServer(serverName string) uint32 {
	maxNumModels := uint32(0)
	for _, model := range m.store.models {
		latest := model.Latest()
		if latest != nil && latest.Server() == serverName {
			maxNumModels = max(maxNumModels, uint32(latest.DesiredReplicas()))
		}
	}
	return maxNumModels
}

func toSchedulerLoadedModels(agentLoadedModels []*agent.ModelVersion) map[ModelVersionID]bool {
	loadedModels := make(map[ModelVersionID]bool)
	for _, modelVersionReq := range agentLoadedModels {
		key := ModelVersionID{
			Name:    modelVersionReq.GetModel().GetMeta().GetName(),
			Version: modelVersionReq.GetVersion(),
		}
		loadedModels[key] = true
	}
	return loadedModels
}
