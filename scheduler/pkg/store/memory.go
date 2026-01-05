/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package store

import (
	"errors"
	"fmt"
	"sort"
	"sync"

	"github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler/db"
	log "github.com/sirupsen/logrus"

	"github.com/seldonio/seldon-core/apis/go/v2/mlops/agent"
	pb "github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/coordinator"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store/utils"
)

type ModelServerStore struct {
	mu      sync.RWMutex
	opLocks sync.Map
	store   struct {
		models  Storage[*db.Model]
		servers Storage[*db.Server]
	}
	logger   log.FieldLogger
	eventHub *coordinator.EventHub
}

var _ ModelServerAPI = &ModelServerStore{}

func NewMemoryStore(
	logger log.FieldLogger,
	modelStore Storage[*db.Model],
	serverStore Storage[*db.Server],
	eventHub *coordinator.EventHub,
) *ModelServerStore {
	return &ModelServerStore{
		store: struct {
			models  Storage[*db.Model]
			servers Storage[*db.Server]
		}{models: modelStore, servers: serverStore},
		logger:   logger.WithField("source", "ModelServerStore"),
		eventHub: eventHub,
	}
}

func (m *ModelServerStore) GetAllModels() ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var modelNames []string

	models, err := m.store.models.List()
	if err != nil {
		return nil, err
	}

	for _, model := range models {
		modelNames = append(modelNames, model.GetName())
	}
	return modelNames, nil
}

func (m *ModelServerStore) GetModels() ([]*db.Model, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.store.models.List()
}

func (m *ModelServerStore) addModelVersionIfNotExists(req *agent.ModelVersion) (*db.Model, *db.ModelVersion, error) {
	modelName := req.GetModel().GetMeta().GetName()
	model, err := m.store.models.Get(modelName)
	if err != nil {
		if !errors.Is(err, ErrNotFound) {
			return nil, nil, err
		}
		model = &db.Model{Name: modelName}
		if err := m.store.models.Insert(model); err != nil {
			return nil, nil, err
		}
	}

	var existingModelVersion *db.ModelVersion
	if existingModelVersion = model.GetVersion(req.GetVersion()); existingModelVersion == nil {
		modelVersion := NewDefaultModelVersion(req.GetModel(), req.GetVersion())
		model.Versions = append(model.Versions, modelVersion)
		sort.SliceStable(model.Versions, func(i, j int) bool { // resort model versions based on version number
			return model.Versions[i].GetVersion() < model.Versions[j].GetVersion()
		})
		return model, modelVersion, nil
	}

	return model, existingModelVersion, nil
}

func (m *ModelServerStore) addNextModelVersion(model *db.Model, pbmodel *pb.Model) {
	// if we start from a clean state, lets use the generation id as the starting version
	// this is to ensure that we have monotonic increasing version numbers
	// and we never reset back to 1
	generation := pbmodel.GetMeta().GetKubernetesMeta().GetGeneration()
	version := max(uint32(1), uint32(generation))
	if model.Latest() != nil {
		version = model.Versions[len(model.Versions)-1].Version + 1
	}
	modelVersion := NewDefaultModelVersion(pbmodel, version)

	model.Versions = append(model.Versions, modelVersion)
	sort.SliceStable(model.Versions, func(i, j int) bool { // resort model versions based on version number
		return model.Versions[i].GetVersion() < model.Versions[j].GetVersion()
	})
}

func (m *ModelServerStore) UpdateModel(req *pb.LoadModelRequest) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	logger := m.logger.WithField("func", "UpdateModel")
	modelName := req.GetModel().GetMeta().GetName()
	validName := utils.CheckName(modelName)
	if !validName {
		return fmt.Errorf(
			"Model %s does not have a valid name - it must be alphanumeric and not contains dots (.)",
			modelName,
		)
	}
	model, err := m.store.models.Get(modelName)
	if err != nil {
		if !errors.Is(err, ErrNotFound) {
			return fmt.Errorf("failed to get model %s: %w", modelName, err)
		}
		model = &db.Model{Name: modelName}
		m.addNextModelVersion(model, req.GetModel())
		if err := m.store.models.Insert(model); err != nil {
			return fmt.Errorf("failed to update model %s: %w", modelName, err)
		}
		return nil
	}

	if model.Deleted {
		if model.Inactive() {
			m.addNextModelVersion(model, req.GetModel())
			if err := m.store.models.Update(model); err != nil {
				return fmt.Errorf("failed to update model %s: %w", modelName, err)
			}
			return nil
		}
		return fmt.Errorf(
			"model %s is in process of deletion - new model can not be created",
			modelName,
		)

	}

	latestModel := model.Latest()
	if latestModel == nil {
		return fmt.Errorf("model %s has no latest version", modelName)
	}

	meq := ModelEqualityCheck(latestModel.ModelDefn, req.GetModel())

	switch {
	case meq.Equal:
		logger.Debugf("Model %s semantically equal - doing nothing", modelName)
		return nil
	case meq.ModelSpecDiffers:
		logger.Debugf("Model %s model spec differs - adding new version of model", modelName)
		m.addNextModelVersion(model, req.GetModel())
	case meq.DeploymentSpecDiffers:
		logger.Debugf(
			"Model %s deployment spec differs - updating latest model version with new spec",
			modelName,
		)
		latestModel.ModelDefn.DeploymentSpec = req.GetModel().GetDeploymentSpec()
	}

	if meq.MetaDiffers {
		// Update just kubernetes meta
		latestModel.ModelDefn.Meta.KubernetesMeta = req.GetModel().GetMeta().GetKubernetesMeta()
	}

	if err := m.store.models.Update(model); err != nil {
		return fmt.Errorf("failed to update model %s: %w", modelName, err)
	}
	return nil
}

func (m *ModelServerStore) deepCopy(model *Model, key string) *ModelSnapshot {
	snapshot := &ModelSnapshot{
		Name:    key,
		Deleted: model.IsDeleted(),
	}

	snapshot.Versions = make([]*ModelVersion, len(model.versions))
	for i, version := range model.versions {
		snapshot.Versions[i] = version.DeepCopy()
	}
	return snapshot
}

func (m *ModelServerStore) LockModel(modelId string) {
	var lock sync.RWMutex
	existingLock, _ := m.opLocks.LoadOrStore(modelId, &lock)
	existingLock.(*sync.RWMutex).Lock()
}

func (m *ModelServerStore) UnlockModel(modelId string) {
	logger := m.logger.WithField("func", "UnlockModel")
	lock, loaded := m.opLocks.Load(modelId)
	if loaded {
		lock.(*sync.RWMutex).Unlock()
	} else {
		logger.Warnf("Trying to unlock model %s that was not locked.", modelId)
	}
}

func (m *ModelServerStore) GetModel(key string) (*db.Model, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	model, err := m.store.models.Get(key)
	if err != nil {
		return nil, fmt.Errorf("failed to get model %s: %w", key, err)
	}
	return model, nil
}

func (m *ModelServerStore) RemoveModel(req *pb.UnloadModelRequest) error {
	err := m.removeModelImpl(req)
	if err != nil {
		return err
	}
	return nil
}

func (m *ModelServerStore) removeModelImpl(req *pb.UnloadModelRequest) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	modelName := req.GetModel().GetName()
	model, err := m.store.models.Get(modelName)
	if err == nil {
		// Updating the k8s meta is required to be updated so status updates back (to manager)
		// will match latest generation value. Previous generation values might be ignored by manager.
		if req.GetKubernetesMeta() != nil { // k8s meta can be nil if unload is called directly using scheduler grpc api
			model.Latest().ModelDefn.Meta.KubernetesMeta = req.GetKubernetesMeta()
		}

		model.Deleted = true
		m.setModelGwStatusToTerminate(true, model.Latest())
		return m.updateModelStatus(true, true, model.Latest(), model.GetLastAvailableModelVersion(), model)
	}

	if errors.Is(err, ErrNotFound) {
		return fmt.Errorf("model %s not found", req.GetModel().GetName())
	}
	return err
}

func (m *ModelServerStore) GetServers(shallow bool, modelDetails bool) ([]*db.Server, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.store.servers.List()
}

func (m *ModelServerStore) GetServer(serverKey string, _ bool, modelDetails bool) (*db.Server, *ServerStats, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	server, err := m.store.servers.Get(serverKey)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, nil, fmt.Errorf("server [%s] not found", serverKey)
		}
		return nil, nil, err
	}

	if modelDetails {
		// this is a hint to the caller that the server is in a state where it can be scaled down
		stats, err := m.getServerStats(serverKey)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get stats: %w", err)
		}
		return server, stats, nil
	}

	return server, nil, nil
}

func (m *ModelServerStore) getServerStats(serverKey string) (*ServerStats, error) {
	emptyReplicas, err := m.numEmptyServerReplicas(serverKey)
	if err != nil {
		return nil, err
	}

	maxModelReplicas, err := m.maxNumModelReplicasForServer(serverKey)
	if err != nil {
		return nil, err
	}

	return &ServerStats{
		NumEmptyReplicas:          emptyReplicas,
		MaxNumReplicaHostedModels: maxModelReplicas,
	}, nil
}

func (m *ModelServerStore) getModelServer(
	modelKey string,
	version uint32,
	serverKey string,
) (*db.Model, *db.ModelVersion, *db.Server, error) {
	// Validate
	model, err := m.store.models.Get(modelKey)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to find model %s", modelKey)
	}
	modelVersion := model.GetVersion(version)
	if modelVersion == nil {
		return nil, nil, nil, fmt.Errorf("Version not found for model %s, version %d", modelKey, version)
	}
	server, err := m.store.servers.Get(serverKey)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to find server %s", serverKey)
	}
	return model, modelVersion, server, nil
}

func (m *ModelServerStore) UpdateLoadedModels(
	modelKey string,
	version uint32,
	serverKey string,
	replicas []*db.ServerReplica,
) error {
	m.mu.Lock()
	modelEvt, err := m.updateLoadedModelsImpl(modelKey, version, serverKey, replicas)
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
	return nil
}

func (m *ModelServerStore) updateLoadedModelsImpl(
	modelKey string,
	version uint32,
	serverKey string,
	replicas []*db.ServerReplica,
) (*coordinator.ModelEventMsg, error) {
	logger := m.logger.WithField("func", "updateLoadedModelsImpl")

	// Validate
	model, err := m.store.models.Get(modelKey)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, fmt.Errorf("model [%s] not found", modelKey)
		}
		return nil, fmt.Errorf("failed to find model %s", modelKey)
	}

	modelVersion := model.Latest()
	if modelVersion == nil {
		return nil, fmt.Errorf("latest version not found for model %s", modelKey)
	}
	if version != modelVersion.GetVersion() {
		return nil, fmt.Errorf(
			"model version mismatch for %s got %d but latest version is now %d",
			modelKey, version, modelVersion.GetVersion(),
		)
	}

	if serverKey == "" {
		// nothing to do for a model that doesn't have a server, proceed with sending an event for downstream
		return &coordinator.ModelEventMsg{
			ModelName:    modelVersion.ModelDefn.Meta.Name,
			ModelVersion: modelVersion.GetVersion(),
		}, nil
	}

	server, err := m.store.servers.Get(serverKey)
	if err != nil {
		return nil, fmt.Errorf("failed to find server %s: %w", serverKey, err)
	}

	assignedReplicaIds := make(map[int32]struct{})
	for _, replica := range replicas {
		if _, ok := server.Replicas[replica.ReplicaIdx]; !ok {
			return nil, fmt.Errorf(
				"failed to reserve replica %d as it does not exist on server %s",
				replica.ReplicaIdx, serverKey,
			)
		}
		assignedReplicaIds[replica.ReplicaIdx] = struct{}{}
	}

	for modelVersion.HasServer() && modelVersion.Server != serverKey {
		logger.Debugf("Adding new version as server changed to %s from %s", modelVersion.Server, serverKey)
		m.addNextModelVersion(model, model.Latest().ModelDefn)
		if err := m.store.models.Update(model); err != nil {
			return nil, fmt.Errorf("failed to update model %s: %w", modelKey, err)
		}
		modelVersion = model.Latest()
	}

	//  reserve memory for existing replicas that are not already loading or loaded
	replicaStateUpdated := false
	for replicaIdx := range assignedReplicaIds {
		if existingState, ok := modelVersion.Replicas[replicaIdx]; !ok {
			logger.Debugf(
				"Model %s version %d on server %s replica %d does not exist yet and should be loaded",
				modelKey, modelVersion.Version, serverKey, replicaIdx,
			)
			modelVersion.SetReplicaState(int(replicaIdx), db.ModelReplicaState_UnloadRequested, "")
			if err := m.updateReservedMemory(db.ModelReplicaState_LoadRequested, serverKey,
				int(replicaIdx), modelVersion.GetRequiredMemory()); err != nil {
				return nil, fmt.Errorf("failed to update server %s replica %d: %w", serverKey, replicaIdx, err)
			}
			replicaStateUpdated = true
		} else {
			logger.Debugf(
				"Checking if model %s version %d state %s on server %s replica %d should be loaded",
				modelKey, modelVersion.Version, existingState.State.String(), serverKey, replicaIdx,
			)
			if !existingState.State.AlreadyLoadingOrLoaded() {
				modelVersion.SetReplicaState(int(replicaIdx), db.ModelReplicaState_LoadRequested, "")
				if err := m.updateReservedMemory(db.ModelReplicaState_LoadRequested, serverKey,
					int(replicaIdx), modelVersion.GetRequiredMemory()); err != nil {
					return nil, fmt.Errorf("failed to update server %s replica %d: %w", serverKey, replicaIdx, err)
				}
				replicaStateUpdated = true
			}
		}
	}

	// Unload any existing model replicas assignments that are no longer part of the replica set
	for replicaIdx, existingState := range modelVersion.Replicas {
		if _, ok := assignedReplicaIds[replicaIdx]; !ok {
			logger.Debugf(
				"Checking if replicaidx %d with state %s should be unloaded",
				replicaIdx, existingState.State.String(),
			)
			if !existingState.State.UnloadingOrUnloaded() && existingState.State != db.ModelReplicaState_Draining {
				modelVersion.SetReplicaState(int(replicaIdx), db.ModelReplicaState_UnloadEnvoyRequested, "")
				replicaStateUpdated = true
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
	if replicaStateUpdated || modelVersion.State.State == db.ModelState_ScheduleFailed || model.Deleted ||
		modelVersion.State.State == db.ModelState_ModelProgressing ||
		(modelVersion.State.State == db.ModelState_ModelAvailable && len(modelVersion.GetAssignment()) < modelVersion.DesiredReplicas()) {
		logger.Debugf("Updating model status for model %s server %s", modelKey, serverKey)
		modelVersion.Server = serverKey
		if err := m.updateModelStatus(true, model.Deleted, modelVersion, model.GetLastAvailableModelVersion(), model); err != nil {
			return nil, fmt.Errorf("failed to update model %s: %w", modelKey, err)
		}

		return &coordinator.ModelEventMsg{
			ModelName:    modelVersion.ModelDefn.Meta.Name,
			ModelVersion: modelVersion.GetVersion(),
		}, nil
	}

	logger.Debugf("Model status update not required for model %s server %s as no replicas were updated", modelKey, serverKey)
	return nil, nil
}

func (m *ModelServerStore) UnloadVersionModels(modelKey string, version uint32) (bool, error) {
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

func (m *ModelServerStore) unloadVersionModelsImpl(modelKey string, version uint32) (*coordinator.ModelEventMsg, bool, error) {
	logger := m.logger.WithField("func", "UnloadVersionModels")
	m.mu.Lock()
	defer m.mu.Unlock()

	// Validate
	model, err := m.store.models.Get(modelKey)
	if err != nil {
		return nil, false, fmt.Errorf("failed to find model %s", modelKey)
	}
	modelVersion := model.GetVersion(version)
	if modelVersion == nil {
		return nil, false, fmt.Errorf("version not found for model %s, version %d", modelKey, version)
	}

	updated := false
	for replicaIdx, existingState := range modelVersion.Replicas {
		if !existingState.State.UnloadingOrUnloaded() {
			logger.Debugf(
				"Setting model %s version %d on server %s replica %d to UnloadRequested was %s",
				modelKey,
				modelVersion.Version,
				modelVersion.Server,
				replicaIdx,
				existingState.State.String(),
			)
			modelVersion.SetReplicaState(int(replicaIdx), db.ModelReplicaState_UnloadRequested, "")
			updated = true
			continue
		}
		logger.Debugf(
			"model %s on server %s replica %d already unloaded",
			modelKey, modelVersion.Server, replicaIdx,
		)
	}

	if updated {
		logger.Debugf("Calling update model status for model %s version %d", modelKey, version)
		if err := m.updateModelStatus(false, model.Deleted, modelVersion, model.GetLastAvailableModelVersion(), model); err != nil {
			return nil, false, fmt.Errorf("failed to update model %s: %w", modelKey, err)
		}
		return &coordinator.ModelEventMsg{
			ModelName:    modelVersion.ModelDefn.Meta.Name,
			ModelVersion: modelVersion.GetVersion(),
		}, true, nil
	}

	return nil, false, nil
}

func (m *ModelServerStore) UpdateModelState(
	modelKey string,
	version uint32,
	serverKey string,
	replicaIdx int,
	availableMemory *uint64,
	expectedState db.ModelReplicaState,
	desiredState db.ModelReplicaState,
	reason string,
	runtimeInfo *pb.ModelRuntimeInfo,
) error {
	modelEvt, serverEvt, err := m.updateModelState(modelKey, version, serverKey, replicaIdx, availableMemory, expectedState, desiredState, reason, runtimeInfo)
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

type updateModelResult struct {
	eventModel  *coordinator.ModelEventMsg
	eventServer *coordinator.ServerEventMsg
	model       *db.Model
	server      *db.Server
}

func (m *ModelServerStore) updateModelState(
	modelKey string,
	version uint32,
	serverKey string,
	replicaIdx int,
	availableMemory *uint64,
	expectedState db.ModelReplicaState,
	desiredState db.ModelReplicaState,
	reason string,
	runtimeInfo *pb.ModelRuntimeInfo,
) (*coordinator.ModelEventMsg, *coordinator.ServerEventMsg, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	res, err := m.updateModelStateImpl(modelKey, version, serverKey, replicaIdx, availableMemory, expectedState, desiredState, reason, runtimeInfo)
	if err != nil {
		return nil, nil, err
	}

	// TODO should be in a transaction to rollback if updating servers fail
	if err := m.store.models.Update(res.model); err != nil {
		return nil, nil, fmt.Errorf("failed to update model %s: %w", modelKey, err)
	}
	if err := m.store.servers.Update(res.server); err != nil {
		return nil, nil, fmt.Errorf("failed to update server %s: %w", serverKey, err)
	}

	return res.eventModel, res.eventServer, nil
}

func (m *ModelServerStore) updateModelStateImpl(
	modelKey string,
	version uint32,
	serverKey string,
	replicaIdx int,
	availableMemory *uint64,
	expectedState db.ModelReplicaState,
	desiredState db.ModelReplicaState,
	reason string,
	runtimeInfo *pb.ModelRuntimeInfo,
) (*updateModelResult, error) {
	logger := m.logger.WithField("func", "updateModelStateImpl")

	// Validate
	model, modelVersion, server, err := m.getModelServer(modelKey, version, serverKey)
	if err != nil {
		return nil, err
	}

	modelVersion.UpdateRuntimeInfo(runtimeInfo)

	existingState := modelVersion.GetModelReplicaState(replicaIdx)
	if existingState != expectedState {
		return nil, fmt.Errorf(
			"state mismatch for %s:%d expected state %s but was %s when trying to move to state %s",
			modelKey, version, expectedState.String(), existingState.String(), desiredState.String(),
		)
	}

	if err := m.updateReservedMemory(desiredState, serverKey, replicaIdx, modelVersion.GetRequiredMemory()); err != nil {
		return nil, fmt.Errorf("failed to server %s replica %d: %w", serverKey, replicaIdx, err)
	}

	deletedModelReplica := false
	if existingState == desiredState {
		return &updateModelResult{
			model:  model,
			server: server,
		}, nil
	}

	latestModel := model.Latest()
	isLatest := latestModel.GetVersion() == modelVersion.GetVersion()

	modelVersion.SetReplicaState(replicaIdx, desiredState, reason)
	logger.Debugf(
		"Setting model %s version %d on server %s replica %d to %s",
		modelKey, version, serverKey, replicaIdx, desiredState.String(),
	)

	// Update models loaded onto replica for relevant state
	if desiredState == db.ModelReplicaState_Loaded ||
		desiredState == db.ModelReplicaState_Loading ||
		desiredState == db.ModelReplicaState_Unloaded ||
		desiredState == db.ModelReplicaState_LoadFailed {

		replica, ok := server.Replicas[int32(replicaIdx)]
		if ok {
			if desiredState == db.ModelReplicaState_Loaded || desiredState == db.ModelReplicaState_Loading {
				logger.Infof(
					"Adding model %s(%d) to server %s replica %d list of loaded / loading models",
					modelKey, version, serverKey, replicaIdx,
				)
				replica.AddModelVersion(modelKey, version, desiredState) // we need to distinguish between loaded and loading
			} else {
				logger.Infof(
					"Removing model %s(%d) from server %s replica %d list of loaded / loading models",
					modelKey, version, serverKey, replicaIdx,
				)
				// we could go from loaded -> unloaded or loading -> failed. in the case we have a failure then we just remove from loading
				deletedModelReplica = true
				replica.DeleteModelVersion(modelKey, version)
			}
		}
	}
	if availableMemory != nil {
		server.Replicas[int32(replicaIdx)].AvailableMemory = *availableMemory
	}

	if err := m.updateModelStatus(isLatest, model.Deleted, modelVersion, model.GetLastAvailableModelVersion(), model); err != nil {
		return nil, fmt.Errorf("update model status failed: %w", err)
	}

	if err := m.store.servers.Update(server); err != nil {
		return nil, fmt.Errorf("failed to update server %s: %w", serverKey, err)
	}

	modelEvt := &coordinator.ModelEventMsg{
		ModelName:    modelVersion.ModelDefn.Meta.Name,
		ModelVersion: modelVersion.GetVersion(),
	}
	if deletedModelReplica {
		return &updateModelResult{
			eventModel: modelEvt,
			eventServer: &coordinator.ServerEventMsg{
				ServerName:    serverKey,
				UpdateContext: coordinator.SERVER_SCALE_DOWN,
			},
			model:  model,
			server: server,
		}, nil
	}

	return &updateModelResult{
		eventModel: modelEvt,
		model:      model,
		server:     server,
	}, nil
}

func (m *ModelServerStore) updateReservedMemory(
	modelReplicaState db.ModelReplicaState, serverKey string, replicaIdx int, memBytes uint64,
) error {
	// update reserved memory that is being used for sorting replicas
	// do we need to lock replica update?
	server, err := m.store.servers.Get(serverKey)
	if err != nil {
		return fmt.Errorf("get server %s failed: %w", serverKey, err)
	}

	replica, okReplica := server.Replicas[int32(replicaIdx)]
	update := false
	if okReplica {
		if modelReplicaState == db.ModelReplicaState_LoadRequested {
			replica.UpdateReservedMemory(memBytes, true)
			update = true
		} else if modelReplicaState == db.ModelReplicaState_LoadFailed ||
			modelReplicaState == db.ModelReplicaState_Loaded {
			replica.UpdateReservedMemory(memBytes, false)
			update = true
		}
	}

	if update {
		return m.store.servers.Update(server)
	}
	return nil
}

func (m *ModelServerStore) AddServerReplica(request *agent.AgentSubscribeRequest) error {
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

func (m *ModelServerStore) addServerReplicaImpl(request *agent.AgentSubscribeRequest) ([]coordinator.ModelEventMsg, coordinator.ServerEventMsg, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	server, err := m.store.servers.Get(request.ServerName)
	if err != nil {
		if !errors.Is(err, ErrNotFound) {
			return nil, coordinator.ServerEventMsg{}, err
		}
		server = NewServer(request.ServerName, request.Shared)
		if err := m.store.servers.Insert(server); err != nil {
			return nil, coordinator.ServerEventMsg{}, fmt.Errorf("failed to add server %s: %w", request.ServerName, err)
		}
	}
	server.Shared = request.Shared

	loadedModels := toSchedulerLoadedModels(request.LoadedModels)

	serverReplica := NewServerReplicaFromConfig(
		server,
		int(request.ReplicaIdx),
		loadedModels,
		request.ReplicaConfig,
		request.AvailableMemoryBytes,
	)

	server.AddReplica(int32(request.ReplicaIdx), serverReplica)

	if err := m.store.servers.Update(server); err != nil {
		return nil, coordinator.ServerEventMsg{}, fmt.Errorf("failed to update server %s: %w", request.ServerName, err)
	}

	var evts []coordinator.ModelEventMsg
	for _, modelVersionReq := range request.LoadedModels {
		model, modelVersion, err := m.addModelVersionIfNotExists(modelVersionReq)
		if err != nil {
			return nil, coordinator.ServerEventMsg{}, err
		}
		modelVersion.Replicas[int32(request.ReplicaIdx)] = &db.ReplicaStatus{State: db.ModelReplicaState_Loaded}
		modelVersion.Server = request.ServerName
		if err := m.updateModelStatus(true, false, modelVersion, model.GetLastAvailableModelVersion(), model); err != nil {
			return nil, coordinator.ServerEventMsg{}, fmt.Errorf("failed to update model status: %w", err)
		}
		evts = append(evts, coordinator.ModelEventMsg{
			ModelName:    modelVersion.ModelDefn.Meta.Name,
			ModelVersion: modelVersion.GetVersion(),
		})
	}

	serverEvt := coordinator.ServerEventMsg{
		ServerName:    request.ServerName,
		ServerIdx:     request.ReplicaIdx,
		UpdateContext: coordinator.SERVER_REPLICA_CONNECTED,
	}

	return evts, serverEvt, nil
}

func (m *ModelServerStore) RemoveServerReplica(serverName string, replicaIdx int) ([]string, error) {
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

func (m *ModelServerStore) removeServerReplicaImpl(serverName string, replicaIdx int) ([]string, []coordinator.ModelEventMsg, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	server, err := m.store.servers.Get(serverName)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to find server %s: %w", serverName, err)
	}
	serverReplica, ok := server.Replicas[int32(replicaIdx)]
	if !ok {
		return nil, nil, fmt.Errorf("Failed to find replica %d for server %s", replicaIdx, serverName)
	}
	delete(server.Replicas, int32(replicaIdx))

	if err := m.store.servers.Update(server); err != nil {
		return nil, nil, fmt.Errorf("failed to update server %s: %w", serverName, err)
	}

	// TODO we should not reschedule models on servers with dedicated models, e.g. non shareable servers
	if len(server.Replicas) == 0 {
		if err := m.store.servers.Delete(serverName); err != nil {
			return nil, nil, fmt.Errorf("failed to delete server %s: %w", serverName, err)
		}
	}
	loadedModelsRemoved, loadedEvts, err := m.removeModelfromServerReplica(serverReplica.LoadedModels, replicaIdx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to remove loaded models from replica %d: %w", replicaIdx, err)
	}
	loadingModelsRemoved, loadingEtvs, err := m.removeModelfromServerReplica(serverReplica.LoadingModels, replicaIdx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to remove loading models from replica %d: %w", replicaIdx, err)
	}

	modelsRemoved := append(loadedModelsRemoved, loadingModelsRemoved...)
	evts := append(loadedEvts, loadingEtvs...)

	return modelsRemoved, evts, nil
}

func (m *ModelServerStore) removeModelfromServerReplica(models []*db.ModelVersionID, replicaIdx int) ([]string, []coordinator.ModelEventMsg, error) {
	logger := m.logger.WithField("func", "RemoveServerReplica")
	var modelNames []string
	var evts []coordinator.ModelEventMsg
	// Find models to reschedule due to this server replica being removed
	for _, v := range models {
		// TODO pointer issue
		model, err := m.store.models.Get(v.Name)
		if err == nil {
			modelVersion := model.GetVersion(v.Version)
			if modelVersion != nil {
				modelVersion.DeleteReplica(replicaIdx)
				if model.Deleted || model.Latest().GetVersion() != modelVersion.GetVersion() {
					// In some cases we found that the user can ask for a model to be deleted and the model replica
					// is still in the process of being loaded. In this case we should not reschedule the model.
					logger.Debugf(
						"Model %s is being deleted and server replica %d is disconnected, skipping",
						v.Name, replicaIdx,
					)
					modelVersion.SetReplicaState(replicaIdx, db.ModelReplicaState_Unloaded,
						"model is removed when server replica was removed")
					m.LockModel(v.Name)
					if err := m.updateModelStatus(
						model.Latest().GetVersion() == modelVersion.GetVersion(),
						model.Deleted, modelVersion, model.GetLastAvailableModelVersion(), model); err != nil {
						return nil, nil, fmt.Errorf("failed to update model %s status: %w", model.Name, err)
					}
					m.UnlockModel(v.Name)
					// send an event to progress the deletion
					evts = append(
						evts,
						coordinator.ModelEventMsg{
							ModelName:    modelVersion.ModelDefn.Meta.Name,
							ModelVersion: modelVersion.GetVersion(),
						},
					)
				} else {
					modelNames = append(modelNames, v.Name)
					if err := m.store.models.Update(model); err != nil {
						return nil, nil, fmt.Errorf("failed to update model %s status: %w", model.Name, err)
					}
				}
			} else {
				logger.Warnf("Can't find model version %s", v.String())
			}
		}
	}
	return modelNames, evts, nil
}

func (m *ModelServerStore) DrainServerReplica(serverName string, replicaIdx int) ([]string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.drainServerReplicaImpl(serverName, replicaIdx)
}

func (m *ModelServerStore) drainServerReplicaImpl(serverName string, replicaIdx int) ([]string, error) {
	logger := m.logger.WithField("func", "DrainServerReplica")
	server, err := m.store.servers.Get(serverName)
	if err != nil {
		return nil, fmt.Errorf("failed to find server %s", serverName)
	}
	serverReplica, ok := server.Replicas[int32(replicaIdx)]
	if !ok {
		return nil, fmt.Errorf("failed to find replica %d for server %s", replicaIdx, serverName)
	}

	// we mark this server replica as draining so should not be used in future scheduling decisions
	serverReplica.IsDraining = true

	if err := m.store.servers.Update(server); err != nil {
		return nil, fmt.Errorf("failed to update server %s: %w", serverName, err)
	}

	loadedModels, err := m.findModelsToReSchedule(serverReplica.LoadingModels, replicaIdx)
	if err != nil {
		return nil, err
	}
	if len(loadedModels) > 0 {
		logger.WithField("models", loadedModels).Debug("Found loaded models to re-schedule")
	}
	loadingModels, err := m.findModelsToReSchedule(serverReplica.LoadingModels, replicaIdx)
	if err != nil {
		return nil, err
	}
	if len(loadingModels) > 0 {
		logger.WithField("models", loadingModels).Debug("Found loading models to re-schedule")
	}

	return append(loadedModels, loadingModels...), nil
}

func (m *ModelServerStore) findModelsToReSchedule(models []*db.ModelVersionID, replicaIdx int) ([]string, error) {
	logger := m.logger.WithField("func", "DrainServerReplica")
	modelsReSchedule := make([]string, 0)

	for _, v := range models {
		model, err := m.store.models.Get(v.Name)
		if err == nil {
			modelVersion := model.GetVersion(v.Version)
			if modelVersion != nil {
				modelVersion.SetReplicaState(replicaIdx, db.ModelReplicaState_Draining, "trigger to drain")
				if err := m.store.models.Update(model); err != nil {
					return nil, fmt.Errorf("failed update model %s: %w", model.Name, err)
				}
				modelsReSchedule = append(modelsReSchedule, model.Name)
				continue
			}
			logger.Warnf("Can't find model version %s", model.String())
		}
	}

	return modelsReSchedule, nil
}

func (m *ModelServerStore) ServerNotify(request *pb.ServerNotify) error {
	logger := m.logger.WithField("func", "MemoryServerNotify")
	m.mu.Lock()
	defer m.mu.Unlock()

	logger.Debugf("ServerNotify %v", request)

	server, err := m.store.servers.Get(request.Name)
	if err != nil {
		if !errors.Is(err, ErrNotFound) {
			return fmt.Errorf("failed to find server %s: %w", request.Name, err)
		}
		server = NewServer(request.Name, request.Shared)
		server.ExpectedReplicas = int32(request.ExpectedReplicas)
		server.MinReplicas = int32(request.MinReplicas)
		server.MaxReplicas = int32(request.MaxReplicas)
		server.KubernetesMeta = request.KubernetesMeta
		if err := m.store.servers.Insert(server); err != nil {
			return err
		}
		return nil
	}

	server.ExpectedReplicas = int32(request.ExpectedReplicas)
	server.MinReplicas = int32(request.MinReplicas)
	server.MaxReplicas = int32(request.MaxReplicas)
	server.KubernetesMeta = request.KubernetesMeta

	if err := m.store.servers.Update(server); err != nil {
		return fmt.Errorf("failed to update server %s: %w", request.Name, err)
	}
	return nil
}

func (m *ModelServerStore) numEmptyServerReplicas(serverName string) (uint32, error) {
	emptyReplicas := uint32(0)
	server, err := m.store.servers.Get(serverName)
	if err != nil {
		if !errors.Is(err, ErrNotFound) {
			return 0, err
		}
		return emptyReplicas, nil
	}
	for _, replica := range server.Replicas {
		if len(replica.GetLoadedOrLoadingModelVersions()) == 0 {
			emptyReplicas++
		}
	}
	return emptyReplicas, nil
}

func (m *ModelServerStore) maxNumModelReplicasForServer(serverName string) (uint32, error) {
	models, err := m.store.models.List()
	if err != nil {
		return 0, err
	}

	maxNumModels := uint32(0)
	for _, model := range models {
		latest := model.Latest()
		if latest != nil && latest.Server == serverName {
			maxNumModels = max(maxNumModels, uint32(latest.DesiredReplicas()))
		}
	}
	return maxNumModels, nil
}

func toSchedulerLoadedModels(agentLoadedModels []*agent.ModelVersion) []*db.ModelVersionID {
	loadedModels := make([]*db.ModelVersionID, 0)
	for _, modelVersionReq := range agentLoadedModels {
		loadedModels = append(loadedModels, &db.ModelVersionID{
			Name:    modelVersionReq.GetModel().GetMeta().GetName(),
			Version: modelVersionReq.GetVersion(),
		})
	}
	return loadedModels
}

func (m *ModelServerStore) SetModelGwModelState(name string, versionNumber uint32, status db.ModelState, reason string, source string) error {
	logger := m.logger.WithField("func", "SetModelGwModelState")
	logger.Debugf("Attempt to set model-gw state on model %s:%d status:%s", name, versionNumber, status.String())

	evts, err := m.setModelGwModelStateImpl(name, versionNumber, status, reason, source)
	if err != nil {
		return err
	}

	if m.eventHub != nil {
		for _, evt := range evts {
			m.eventHub.PublishModelEvent(source, *evt)
		}
	}

	return nil
}

func (m *ModelServerStore) setModelGwModelStateImpl(name string, versionNumber uint32, status db.ModelState, reason, source string) ([]*coordinator.ModelEventMsg, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var evts []*coordinator.ModelEventMsg
	model, err := m.store.models.Get(name)
	if err != nil {
		return nil, fmt.Errorf("failed to find model %s", name)
	}
	modelVersion := model.GetVersion(versionNumber)
	if modelVersion == nil {
		return nil, fmt.Errorf("version not found for model %s, version %d", name, versionNumber)
	}

	if modelVersion.State.ModelGwState != status || modelVersion.State.ModelGwReason != reason {
		modelVersion.State.ModelGwState = status
		modelVersion.State.ModelGwReason = reason
		if err := m.store.models.Update(model); err != nil {
			return nil, fmt.Errorf("failed to update model %s: %w", name, err)
		}
		evt := &coordinator.ModelEventMsg{
			ModelName:    modelVersion.ModelDefn.Meta.Name,
			ModelVersion: modelVersion.GetVersion(),
			Source:       source,
		}
		evts = append(evts, evt)
	}
	return evts, nil
}

func (m *ModelServerStore) EmitEvents() error {
	servers, err := m.store.servers.List()
	if err != nil {
		return err
	}
	models, err := m.store.models.List()
	if err != nil {
		return err
	}

	for _, server := range servers {
		for id := range server.Replicas {
			m.eventHub.PublishServerEvent(serverUpdateEventSource, coordinator.ServerEventMsg{
				ServerName:    server.Name,
				ServerIdx:     uint32(id),
				Source:        serverUpdateEventSource,
				UpdateContext: coordinator.SERVER_REPLICA_CONNECTED, // TODO can we be confident of that?
			})
		}
	}

	for _, model := range models {
		latest := model.GetLastAvailableModelVersion()
		if latest == nil {
			continue
		}

		m.eventHub.PublishModelEvent(modelUpdateEventSource, coordinator.ModelEventMsg{
			ModelName:    model.Name,
			Source:       modelUpdateEventSource,
			ModelVersion: latest.Version,
		})
	}

	return nil
}
