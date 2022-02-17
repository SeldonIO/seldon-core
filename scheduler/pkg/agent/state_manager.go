package agent

import (
	"fmt"
	"net/url"
	"sync"

	"github.com/seldonio/seldon-core/scheduler/apis/mlops/agent"
	cache "github.com/seldonio/seldon-core/scheduler/pkg/agent/cache"
	log "github.com/sirupsen/logrus"
)

const (
	DefaultModelMemoryBytes = uint64(1 * 1024 * 1024) // set a 1MB default
)

// manages the state associated with models on local agent
type LocalStateManager struct {
	v2Client      *V2Client
	logger        log.FieldLogger
	modelVersions *ModelState
	cache         cache.CacheManager
	// because of race conditions we might occassionatly go into negative memory
	availableMemoryBytes int64
	// lock for `state.loadedModels`
	mu sync.RWMutex
	// semas for model reloading synchronisation
	semas sync.Map
	// locks for model operation on control plane
	opLocks sync.Map
}

func (manager *LocalStateManager) GetBackEndPath() *url.URL {
	return manager.v2Client.getUrl("/")
}

// this should be called from control plane (if directly)
// the load request will always come with a version attached and the actual load
// is done for all versions, whether this is the first or subequent versions
func (manager *LocalStateManager) LoadModelVersion(modelVersionDetails *agent.ModelVersion) error {
	modelId := modelVersionDetails.GetModel().GetMeta().GetName()
	modelVersion := modelVersionDetails.GetVersion()

	manager.logger.Debugf("Loading model %s, version %d", modelId, modelVersion)

	// model version already assigned to this instance (although could be not in main memory),
	// so do nothing
	if manager.modelVersions.versionExists(modelId, modelVersion) {
		manager.logger.Infof(
			"Model version already exists for %s:%d",
			modelId, modelVersion)
		return nil
	}

	manager.modelVersions.addModelVersion(modelVersionDetails)

	var memBytesToLoad uint64
	var err error
	memBytesToLoad, err = manager.getMemoryDelta(modelId, modelVersionDetails)
	if err != nil {
		return err
	}

	if manager.cache.Exists(modelId) {
		// this is a bit hacky, we bump up priority so it is less likely to be evicted
		// TODO: fixme
		// this is only in case of non-flattened versions
		if err := manager.cache.UpdateDefault(modelId); err != nil {
			return err
		}
	}

	if err := manager.makeRoomIfNeeded(modelId, memBytesToLoad); err != nil {
		manager.modelVersions.removeModelVersion(modelVersionDetails)
		return err
	}

	// be optimistic and mark model memory
	if err := manager.updateAvailableMemory(memBytesToLoad, true); err != nil {
		manager.modelVersions.removeModelVersion(modelVersionDetails)
		return err
	}

	if err := manager.v2Client.LoadModel(modelId); err != nil {
		manager.modelVersions.removeModelVersion(modelVersionDetails)
		if err := manager.updateAvailableMemory(memBytesToLoad, false); err != nil {
			manager.logger.Warn("Could not update memory %s", err)
		}
		return err
	}

	if manager.cache.Exists(modelId) {
		err = manager.cache.UpdateDefault(modelId)
	} else {
		err = manager.cache.AddDefault(modelId)
	}
	if err != nil {
		manager.modelVersions.removeModelVersion(modelVersionDetails)
		if err := manager.updateAvailableMemory(memBytesToLoad, false); err != nil {
			manager.logger.Warn("Could not update memory %s", err)
		}
		return err
	}

	manager.logger.Debugf("Load model %s version %d success, available memory is %d",
		modelId, modelVersion, manager.GetAvailableMemoryBytes())
	return nil
}

// this should be called from control plane (if directly)
func (manager *LocalStateManager) UnloadModelVersion(modelVersionDetails *agent.ModelVersion) error {
	modelId := modelVersionDetails.GetModel().GetMeta().GetName()
	modelVersion := modelVersionDetails.GetVersion()

	manager.logger.Debugf("Unloading model %s, version %d", modelId, modelVersion)

	if !manager.modelVersions.versionExists(modelId, modelVersion) {
		err := fmt.Errorf("Model %s version %d does not exist locally", modelId, modelVersion)
		manager.logger.Error(err)
		return err
	}

	numVersions, err := manager.modelVersions.numVersions(modelId)
	if err != nil {
		return err
	}

	if manager.cache.Exists(modelId) {
		if numVersions > 1 {
			// we assume here that after we remove below the model version there is still
			// other versions of this model that exist, therefore we reload the model
			// to sync the remaining versions from disk.
			if err := manager.v2Client.LoadModel(modelId); err != nil {
				return err
			}
		} else {
			// otherwise we remove the model from server
			if err := manager.v2Client.UnloadModel(modelId); err != nil {
				return err
			}
		}

		updateMemoryFlag := true
		if numVersions == 1 {
			// this is the last version of the model so delete from cache
			if err := manager.cache.Delete(modelId); err != nil {
				manager.logger.Warnf("Delete model %s from cache failed", modelId)
				updateMemoryFlag = false
			}
			manager.logger.Infof("Removed last version of model %s", modelId)
		}

		memBytes, err := manager.modelVersions.getModelVersionMemoryBytes(modelId, modelVersion)
		if err != nil {
			manager.logger.Warnf("Failed to get memory details for model %s version %d", modelId, modelVersion)
			updateMemoryFlag = false
		}

		if updateMemoryFlag {
			// if the model has been removed from the cache and/or server (by a concurrent request),
			// skip updateing memory
			if err := manager.updateAvailableMemory(memBytes, false); err != nil {
				manager.logger.Warn("Could not update memory %s", err)
			}
		} else {
			manager.logger.Warnf("Race condition with %s version %d", modelId, modelVersion)
		}

	}

	modelRemoved := manager.modelVersions.removeModelVersion(modelVersionDetails)

	noRemainingVersions := (numVersions - 1) == 0
	if noRemainingVersions != modelRemoved {
		manager.logger.Warnf("Mismatch in state. Removed all versions from state is [%v] but model repo says remaining versions [%d] for %s:%d",
			modelRemoved, (numVersions - 1), modelId, modelVersion)
	}

	manager.logger.Debugf("Unload model %s version %d success, available memory is %d",
		modelId, modelVersion, manager.GetAvailableMemoryBytes())
	return nil
}

// this should be called from data plane (on incoming inference)
func (manager *LocalStateManager) EnsureLoadModel(modelId string) error {
	// wait for any concurrent operations from the control plane on the same model

	// this is the crux of the logic to go
	manager.logger.Debugf("Ensure that model %s is loaded in memory", modelId)
	manager.modelReadLoadLockCreate(modelId)

	if !manager.cache.Exists(modelId) {
		// wait if the model is currently being reloaded (by another request)
		waiter := manager.modelReloadLockWaitOrCreate(modelId)

		if waiter {
			// concurrent reload is done, retry now
			manager.logger.Warnf("Model %s is reloading concurrently, retry", modelId)

			manager.modelReadLoadUnlock(modelId)
			return manager.EnsureLoadModel(modelId)
		} else {
			// mark sema as done to signal others to go ahead
			defer manager.modelReloadLockDelete(modelId)
		}

		defer manager.modelReadLoadUnlock(modelId)
		manager.logger.Debugf("Making room for %s", modelId)
		// here we need to make sure that we can load the models
		// we also assume that the model exists aleady in the models maps
		modelMemoryBytes, err := manager.modelVersions.getModelTotalMemoryBytes(modelId)
		if err != nil {
			manager.logger.Errorf("Error getting memory for model %s - %s (concurency)", err, modelId)
			return err
		}
		if err := manager.makeRoomIfNeeded(modelId, modelMemoryBytes); err != nil {
			manager.logger.Errorf("No room %s - %s", err, modelId)
			return err
		}

		// be optimistic before actual load
		if err := manager.updateAvailableMemory(modelMemoryBytes, true); err != nil {
			manager.logger.Errorf("Cannot update memory for model %s, %s", modelId, err)
			return err
		}

		if err := manager.v2Client.LoadModel(modelId); err != nil {
			manager.logger.Errorf("Cannot reload %s - %s", err, modelId)
			if err := manager.updateAvailableMemory(modelMemoryBytes, false); err != nil {
				manager.logger.Warn("Could not update memory %s", err)
			}
			return err
		}

		// worse case scenario here is that we have a model in cache that we unloaded and taking a slot
		// however this should (quickly?) go down in priority
		if err := manager.cache.AddDefault(modelId); err != nil {
			manager.logger.Errorf("Cannot re-add to cache %s - %s", err, modelId)
			if err := manager.v2Client.UnloadModel(modelId); err != nil {
				manager.logger.Warnf("Failed to unload model %s with %err", modelId, err)
			}
			if err := manager.updateAvailableMemory(modelMemoryBytes, false); err != nil {
				manager.logger.Warn("Could not update memory %s", err)
			}
			return err
		}

		manager.logger.Infof("Reload model %s success, available memory is %d",
			modelId, manager.GetAvailableMemoryBytes())
	} else {
		defer manager.modelReadLoadUnlock(modelId)
		manager.logger.Debugf("Model exsits in cache %s", modelId)
		if err := manager.cache.UpdateDefault(modelId); err != nil {
			// we try to be speculative here perhaps and still try the inference request
			// this should be very rare in practice
			manager.logger.Warnf("Model %s has been unloaded by a concurrent request (error %s)", modelId, err)
		}
	}
	return nil
}

func (manager *LocalStateManager) GetAvailableMemoryBytes() uint64 {
	manager.mu.RLock()
	defer manager.mu.RUnlock()
	if manager.availableMemoryBytes < 0 {
		return 0
	} else {
		return uint64(manager.availableMemoryBytes)
	}
}

func (manager *LocalStateManager) getMemoryDelta(modelId string, modelVersionDetails *agent.ModelVersion) (uint64, error) {
	if manager.cache.Exists(modelId) {
		// if the model exists in cache and we are loading a new version of it then we only consider the delta of this
		// new version
		memBytesToLoad := modelVersionDetails.GetModel().GetModelSpec().GetMemoryBytes()
		return memBytesToLoad, nil
	} else {
		// otherwise we consider the entirety of a new model:
		// - if this model has other versions (this means that it has been evicted)
		// - if this is the only version then the delta is just this version anyway
		var err error
		memBytesToLoad, err := manager.modelVersions.getModelTotalMemoryBytes(modelId)
		if err != nil {
			return 0, err
		}
		return memBytesToLoad, nil
	}
}

func (manager *LocalStateManager) updateAvailableMemory(memBytes uint64, isLoadingModel bool) error {
	manager.mu.Lock()
	defer manager.mu.Unlock()
	manager.logger.Debugf("Before memory update %d, %d", manager.availableMemoryBytes, memBytes)
	if isLoadingModel {
		if manager.availableMemoryBytes-int64(memBytes) < 0 {
			return fmt.Errorf("Memory will go below zero %d", manager.availableMemoryBytes)
		}
		manager.availableMemoryBytes -= int64(memBytes)
	} else {
		manager.availableMemoryBytes += int64(memBytes)
	}
	manager.logger.Debugf("After memory update %d, %d", manager.availableMemoryBytes, memBytes)
	return nil
}

func (manager *LocalStateManager) modelLoadLockCreate(modelId string) bool {
	var lock sync.RWMutex

	// TODO: find a way to recycle / trim used locks
	existingLock, loaded := manager.opLocks.LoadOrStore(modelId, &lock)

	if !loaded {
		manager.logger.Debugf("Creating load/unload lock for model %s", modelId)
	}
	existingLock.(*sync.RWMutex).Lock()
	manager.logger.Debugf("MODEL %s +WC", modelId)
	manager.logger.Debugf("After lock for model %s", modelId)
	return loaded
}

func (manager *LocalStateManager) modelReadLoadLockCreate(modelId string) bool {
	var lock sync.RWMutex

	// TODO: find a way to recycle / trim used locks
	existingLock, loaded := manager.opLocks.LoadOrStore(modelId, &lock)

	if !loaded {
		manager.logger.Debugf("Creating read load/unload lock for model %s", modelId)
	}
	existingLock.(*sync.RWMutex).RLock()
	manager.logger.Debugf("MODEL %s +RC", modelId)
	manager.logger.Debugf("After read lock for model %s", modelId)
	return loaded
}

func (manager *LocalStateManager) modelReadLoadUnlock(modelId string) {
	manager.logger.Debugf("Deleting read load/unload lock for model %s", modelId)
	lock, loaded := manager.opLocks.Load(modelId)
	if loaded {
		lock.(*sync.RWMutex).RUnlock()
		manager.logger.Debugf("MODEL %s -RC", modelId)
	} else {
		manager.logger.Warnf("Model %s state is inconsistent", modelId)
	}
}

func (manager *LocalStateManager) modelLoadUnlock(modelId string) {
	manager.logger.Debugf("Deleting load/unload lock for model %s", modelId)
	lock, loaded := manager.opLocks.Load(modelId)
	if loaded {
		lock.(*sync.RWMutex).Unlock()
		manager.logger.Debugf("MODEL %s -WC", modelId)
	} else {
		manager.logger.Warnf("Model %s state is inconsistent", modelId)
	}
}

func (manager *LocalStateManager) modelReloadLockWaitOrCreate(modelId string) bool {
	var lock sync.RWMutex

	existingLock, loaded := manager.semas.LoadOrStore(modelId, &lock)

	if !loaded {
		// because of race conditions we need to recheck model in cache again here
		// this is because the initial branch is on model not existing in cache
		// and if we wait for the first concurrent operation to reload the model
		// if will be put back in the cache
		if manager.cache.Exists(modelId) {
			manager.logger.Warnf("Model is already in cache %s", modelId)
			manager.modelReloadLockDelete(modelId)
			return true // treat it as model loaded, TODO: can we simplify logic?
		}
	}

	if loaded {
		manager.logger.Debugf("Waiting for concurrent reload of model %s", modelId)
		existingLock.(*sync.RWMutex).RLock()
		manager.logger.Infof("MODEL %s +RD", modelId)
		defer existingLock.(*sync.RWMutex).RUnlock()
		manager.logger.Infof("MODEL %s -RD", modelId)
	} else {
		manager.logger.Debugf("Creating reload lock for model %s", modelId)
		existingLock.(*sync.RWMutex).Lock()
		manager.logger.Infof("MODEL %s +WD", modelId)
	}
	return loaded
}

func (manager *LocalStateManager) modelReloadLockDelete(modelId string) {
	manager.logger.Debugf("Deleting reload lock for model %s", modelId)
	existingLock, loaded := manager.semas.LoadAndDelete(modelId)
	if loaded {
		existingLock.(*sync.RWMutex).Unlock()
		manager.logger.Infof("MODEL %s -WD", modelId)
	} else {
		manager.logger.Warnf("Model %s state is inconsistent", modelId)
	}

}

func (manager *LocalStateManager) makeRoomIfNeeded(modelId string, modelMemoryBytes uint64) error {
	// note that we might evict a lot of models to make room and we do not know until
	// we have evicted them all.
	// note also that modelMemoryBytes can be just for the new version to be loaded
	// this is provided by the caller

	manager.logger.Debugf("model: %s, avail: %d, required %d",
		modelId, manager.GetAvailableMemoryBytes(), modelMemoryBytes)

	for manager.GetAvailableMemoryBytes() < modelMemoryBytes {
		// note there is a race condition here between the above and below statement
		// this can happen if there is a parallel unload / evict in between so the
		// cache and memory is not reflected
		evictedModelId, evictedModelValue, err := manager.cache.StartEvict()

		if err != nil {
			// due to race condition this could be a false error in the cases we have room
			// so test again memory space
			if manager.GetAvailableMemoryBytes() >= modelMemoryBytes {
				manager.logger.Warnf("Model %s has room now", modelId)
				return nil
			} else {
				return err
			}
		}

		if evictedModelId == modelId {
			// we are trying to load a new version of a model that has a previous version in memory
			// as we cannot unload the previous version on its own we put the model back in the
			// cache with latest timestamp and try the next model to evict if any
			// if this is the last model in the cache break and err
			manager.logger.Warnf(
				"Re-adding model %s to cache as it is the same as new model (different versions?)",
				modelId)
			_ = manager.cache.EndEvict(evictedModelId, evictedModelValue, true) // rollback
			continue
		}

		// note that we unload here all versions of the same model
		if err := manager.v2Client.UnloadModel(evictedModelId); err != nil {
			// if we get an error here, proceed and assume that the model has been unloaded
			// by a concurrent request!
			// TODO: what can we really do about that?
			// is memory calculationn correct, we decide to be conservative and not update memory
			manager.logger.Warnf("Cannot unload model %s from server", evictedModelId)
			_ = manager.cache.EndEvict(evictedModelId, evictedModelValue, false) // should we rollback??
			continue
		}

		// we evict all versions of a model
		evictedModelMemoryBytes, err := manager.modelVersions.getModelTotalMemoryBytes(evictedModelId)
		if err != nil {
			manager.logger.Warn(
				"Could not get memory details for model %s with error %s", evictedModelId, err)
			_ = manager.cache.EndEvict(evictedModelId, evictedModelValue, false)
			continue
		}
		if err := manager.updateAvailableMemory(evictedModelMemoryBytes, false); err != nil {
			manager.logger.Warn("Could not update memory %s", err)
		}
		_ = manager.cache.EndEvict(evictedModelId, evictedModelValue, false)
		manager.logger.Infof(
			"model %s %d evicted (for model %s %d), available memory bytes: %d",
			evictedModelId, evictedModelMemoryBytes, modelId, modelMemoryBytes, manager.GetAvailableMemoryBytes())
	}
	return nil
}

func NewLocalStateManager(
	modelVersions *ModelState,
	logger log.FieldLogger,
	v2Client *V2Client,
	availableMemoryBytes int64,
) *LocalStateManager {
	// if we are here it means that it is a fresh instance with no state yet
	// i.e. should not have any models loaded / cache is empty etc.
	var cache cache.CacheManager = cache.MakeLRU(map[string]int64{})
	return &LocalStateManager{
		v2Client:             v2Client,
		logger:               logger,
		modelVersions:        modelVersions,
		cache:                cache,
		availableMemoryBytes: availableMemoryBytes,
		mu:                   sync.RWMutex{},
		semas:                sync.Map{},
	}
}
