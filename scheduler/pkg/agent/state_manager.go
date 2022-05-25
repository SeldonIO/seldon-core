package agent

import (
	"fmt"
	"net/url"
	"sync"

	"github.com/seldonio/seldon-core/scheduler/apis/mlops/agent"
	cache "github.com/seldonio/seldon-core/scheduler/pkg/agent/cache"
	"github.com/seldonio/seldon-core/scheduler/pkg/metrics"
	log "github.com/sirupsen/logrus"
)

const (
	DefaultModelMemoryBytes = uint64(1 * 1024 * 1024) // use 1MB default
)

// manages the state associated with models on local agent
type LocalStateManager struct {
	v2Client             *V2Client
	logger               log.FieldLogger
	modelVersions        *ModelState
	cache                *cache.CacheTransactionManager
	totalMainMemoryBytes uint64
	overCommitPercentage uint32
	// because of race conditions we might occassionatly go into negative memory
	availableMainMemoryBytes int64
	// lock for `availableMainMemoryBytes`
	mu      sync.RWMutex
	metrics metrics.MetricsHandler
}

func (manager *LocalStateManager) GetBackEndPath() *url.URL {
	return manager.v2Client.getUrl("/")
}

// this should be called from control plane (if directly)
// the load request will always come with versioned model name (only one version)
func (manager *LocalStateManager) LoadModelVersion(modelVersionDetails *agent.ModelVersion) error {
	// note this function is guarded by per-model lock from the caller

	modelId := modelVersionDetails.GetModel().GetMeta().GetName()
	modelVersion := modelVersionDetails.GetVersion()

	manager.logger.Debugf("Loading model %s", modelId)

	// model version already assigned to this instance (although could be not in main memory),
	// so do nothing
	if manager.modelVersions.versionExists(modelId, modelVersion) {
		manager.logger.Infof(
			"Model version already exists for %s:%d",
			modelId, modelVersion)
		return nil
	}

	if _, err := manager.modelVersions.addModelVersion(modelVersionDetails); err != nil {
		return err
	}

	var memBytesToLoad uint64
	var err error
	memBytesToLoad, err = manager.getMemoryDelta(modelId)
	if err != nil {
		if _, err := manager.modelVersions.removeModelVersion(modelVersionDetails); err != nil {
			manager.logger.WithError(err).Warnf("Model removing failed %s", modelId)
		}
		return err
	}

	err = manager.makeRoomIfNeeded(modelId, memBytesToLoad)
	if err != nil {
		if _, err := manager.modelVersions.removeModelVersion(modelVersionDetails); err != nil {
			manager.logger.WithError(err).Warnf("Model removing failed %s", modelId)
		}
		return err
	}

	// be optimistic and mark model memory
	if err := manager.updateAvailableMemory(memBytesToLoad, true); err != nil {
		if _, err := manager.modelVersions.removeModelVersion(modelVersionDetails); err != nil {
			manager.logger.WithError(err).Warnf("Model removing failed %s", modelId)
		}
		return err
	}

	if err := manager.v2Client.LoadModel(modelId); err != nil {
		if _, err := manager.modelVersions.removeModelVersion(modelVersionDetails); err != nil {
			manager.logger.WithError(err).Warnf("Model removing failed %s", modelId)
		}
		if err := manager.updateAvailableMemory(memBytesToLoad, false); err != nil {
			manager.logger.WithError(err).Warnf("Could not update memory for model %s", modelId)
		}
		return err.err
	}

	if err := manager.cache.AddDefault(modelId); err != nil {
		manager.logger.WithError(err).Infof("Cannot load model %s, aborting", modelId)
		if err := manager.v2Client.UnloadModel(modelId); err != nil {
			manager.logger.WithError(err.err).Warnf("Model unload failed %s", modelId)
		}
		if err := manager.updateAvailableMemory(memBytesToLoad, false); err != nil {
			manager.logger.WithError(err).Warnf("Could not update memory %s", modelId)
		}
		if _, err := manager.modelVersions.removeModelVersion(modelVersionDetails); err != nil {
			manager.logger.WithError(err).Warnf("Model removing failed %s", modelId)
		}
		return err
	}

	go manager.metrics.AddLoadedModelMetrics(modelId, memBytesToLoad, true, false)

	manager.logger.Debugf("Load model %s success, available memory is %d",
		modelId, manager.GetAvailableMemoryBytes())
	return nil
}

// this should be called from control plane (if directly)
func (manager *LocalStateManager) UnloadModelVersion(modelVersionDetails *agent.ModelVersion) error {
	// note this function is guarded by per-model lock from the caller

	modelId := modelVersionDetails.GetModel().GetMeta().GetName()
	modelVersion := modelVersionDetails.GetVersion()

	manager.logger.Debugf("Unloading model %s, version %d", modelId, modelVersion)

	if !manager.modelVersions.versionExists(modelId, modelVersion) {
		err := fmt.Errorf("Model %s version %d does not exist locally", modelId, modelVersion)
		manager.logger.Error(err)
		return err
	}

	if manager.cache.Exists(modelId, false) {

		if err := manager.v2Client.UnloadModel(modelId); err != nil {
			if err.IsNotFound() {
				manager.logger.Warnf("Model is not found on server %s", modelId)
			} else {
				manager.logger.WithError(err.err).Errorf("Cannot unload model %s from server", modelId)
				return err.err
			}
		}

		// TODO: should we reload if we get an error here?
		if err := manager.cache.Delete(modelId); err != nil {
			manager.logger.WithError(err).Errorf("Delete model %s from cache failed", modelId)
			return err
		}
		memBytes, err := manager.modelVersions.getModelMemoryBytes(modelId)
		if err != nil {
			manager.logger.WithError(err).Errorf("Failed to get memory details for model %s", modelId)
		}

		if err := manager.updateAvailableMemory(memBytes, false); err != nil {
			manager.logger.WithError(err).Errorf("Could not update memory for model %s", modelId)
			return err
		}

		manager.logger.Infof("Removed model from cache %s", modelId)
		go manager.metrics.AddLoadedModelMetrics(modelId, memBytes, false, false)
	} else {
		go manager.metrics.AddLoadedModelMetrics(modelId, 0, false, false) // model already out of memory
	}

	if _, err := manager.modelVersions.removeModelVersion(modelVersionDetails); err != nil {
		manager.logger.WithError(err).Errorf("Model removing failed for %s", modelId)
		return err
	}

	manager.logger.Debugf("Unload model %s success, available memory is %d",
		modelId, manager.GetAvailableMemoryBytes())
	return nil
}

// this should be called from data plane (on incoming inference)
func (manager *LocalStateManager) EnsureLoadModel(modelId string) error {
	manager.logger.Debugf("Ensure that model %s is loaded in memory", modelId)

	endReloadFn, modelExists := manager.cache.StartReloadIfNotExists(modelId)
	defer endReloadFn()
	if !modelExists {
		manager.logger.Debugf("Making room for %s", modelId)
		// here we need to make sure that we can load the models
		// we also assume that the model exists aleady in the models maps
		modelMemoryBytes, err := manager.modelVersions.getModelMemoryBytes(modelId)
		if err != nil {
			manager.logger.WithError(err).Errorf("Error getting memory for model %s, model unloaded?", modelId)
			return err
		}
		err = manager.makeRoomIfNeeded(modelId, modelMemoryBytes)
		if err != nil {
			manager.logger.Errorf("No room %s - %s", err, modelId)
			return err
		}

		// be optimistic before actual load
		if err := manager.updateAvailableMemory(modelMemoryBytes, true); err != nil {
			manager.logger.WithError(err).Errorf("Cannot update memory for model %s", modelId)
			return err
		}

		if err := manager.v2Client.LoadModel(modelId); err != nil {
			manager.logger.WithError(err.err).Errorf("Cannot reload %s", modelId)
			if err := manager.updateAvailableMemory(modelMemoryBytes, false); err != nil {
				manager.logger.WithError(err).Warnf("Could not update memory %s", modelId)
			}
			return err.err
		}

		if err := manager.cache.AddDefault(modelId); err != nil {
			// we were not too quick and the model has been added by a concurrent request
			// this should not happen in practice, but we need to confirm
			manager.logger.WithError(err).Warnf("Cannot re-add to cache %s", modelId)
			if err := manager.updateAvailableMemory(modelMemoryBytes, false); err != nil {
				manager.logger.WithError(err).Warnf("Could not update memory %s", modelId)
			}
		}

		go manager.metrics.AddLoadedModelMetrics(modelId, modelMemoryBytes, true, true)
		manager.logger.Infof("Reload model %s success, available memory is %d",
			modelId, manager.GetAvailableMemoryBytes())

		return nil
	} else {
		if err := manager.cache.UpdateDefault(modelId); err != nil {
			manager.logger.WithError(err).Errorf(
				"Model %s has been unloaded by a concurrent request", modelId)
			return err
		}
		manager.logger.Debugf("Model exists in cache %s", modelId)
		return nil
	}
}

func (manager *LocalStateManager) GetAvailableMemoryBytes() uint64 {
	manager.mu.RLock()
	defer manager.mu.RUnlock()
	if manager.availableMainMemoryBytes < 0 {
		return 0
	} else {
		return uint64(manager.availableMainMemoryBytes)
	}
}
func (manager *LocalStateManager) GetOverCommitMemoryBytes() float32 {
	return (float32(manager.overCommitPercentage) / 100.) * float32(manager.totalMainMemoryBytes)
}
func (manager *LocalStateManager) GetAvailableMemoryBytesWithOverCommit() uint64 {
	manager.mu.RLock()
	defer manager.mu.RUnlock()
	overCommitBytesMax := manager.GetOverCommitMemoryBytes()
	if manager.modelVersions.getTotalMemoryBytesForAllModels() > manager.totalMainMemoryBytes {
		// we are overcommiting already, so there should be no main memory available
		overCommitUsedBytes := manager.modelVersions.getTotalMemoryBytesForAllModels() - manager.totalMainMemoryBytes
		return uint64(overCommitBytesMax) - overCommitUsedBytes
	} else {
		return uint64(manager.availableMainMemoryBytes) + uint64(overCommitBytesMax)
	}
}

func (manager *LocalStateManager) getMemoryDelta(modelId string) (uint64, error) {
	if memBytesToLoad, err := manager.modelVersions.getModelMemoryBytes(modelId); err != nil {
		return 0, err
	} else {
		return memBytesToLoad, nil
	}
}

func (manager *LocalStateManager) updateAvailableMemory(memBytes uint64, isLoadingModel bool) error {
	manager.mu.Lock()
	defer manager.mu.Unlock()
	manager.logger.Debugf("Before memory update %d, %d", manager.availableMainMemoryBytes, memBytes)
	if isLoadingModel {
		if manager.availableMainMemoryBytes-int64(memBytes) < 0 {
			return fmt.Errorf("Memory will go below zero %d", manager.availableMainMemoryBytes)
		}
		manager.availableMainMemoryBytes -= int64(memBytes)
	} else {
		manager.availableMainMemoryBytes += int64(memBytes)
	}
	manager.logger.Debugf("After memory update %d, %d", manager.availableMainMemoryBytes, memBytes)
	return nil
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
		evictedModelId, evictedModelVal, err := manager.cache.Peek()
		if err != nil {
			manager.logger.WithError(err).Errorf("Cannot peek cache for model %s", modelId)
			return err
		}

		// note: we cannot do `defer endEvictFn` as we want to release the lock before next
		// iteration in the loop
		endEvictFn, err := manager.cache.StartEvict(evictedModelId)

		if err != nil {
			// due to race condition this could be a false error in the cases we have room
			// so test again memory space
			endEvictFn()
			if manager.GetAvailableMemoryBytes() >= modelMemoryBytes {
				manager.logger.WithError(err).Warnf("Model %s has room now", modelId)
				return nil
			} else {
				continue
			}
		}

		evictedModelMemoryBytes, err := manager.modelVersions.getModelMemoryBytes(evictedModelId)
		if err != nil {
			manager.logger.WithError(err).Warnf(
				"Could not get memory details for model %s (for reload model %s)", evictedModelId, modelId)
			endEvictFn()
			continue
		}

		// note that we unload here all versions of the same model
		if err := manager.v2Client.UnloadModel(evictedModelId); err != nil {
			// if we get 404 assume that the model has been unloaded
			// by a concurrent request!
			// otherwise assume that the unload operation failed and put back the model in cache

			if !err.IsNotFound() {
				errStr := fmt.Sprintf("Cannot unload model %s (for reload model %s)", evictedModelId, modelId)
				manager.logger.WithError(err.err).Error(errStr)
				endEvictFn()
				// add the model back to cache
				if err := manager.cache.Add(evictedModelId, evictedModelVal); err != nil {
					manager.logger.Warn(err)
				}
				return err.err
			} else {
				manager.logger.Warnf("Server does not have model %s", evictedModelId)
			}
		}

		if err := manager.updateAvailableMemory(evictedModelMemoryBytes, false); err != nil {
			manager.logger.WithError(err).Warnf("Could not update memory %s", evictedModelId)
		}

		go manager.metrics.AddLoadedModelMetrics(evictedModelId, evictedModelMemoryBytes, false, true)
		endEvictFn()
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
	totalMainMemoryBytes uint64,
	overCommitPercentage uint32,
	metrics metrics.MetricsHandler,
) *LocalStateManager {
	// if we are here it means that it is a fresh instance with no state yet
	// i.e. should not have any models loaded / cache is empty etc.
	cacheWithTransaction := cache.NewLRUCacheTransactionManager(logger)

	return &LocalStateManager{
		v2Client:                 v2Client,
		logger:                   logger.WithField("Source", "StateManager"),
		modelVersions:            modelVersions,
		cache:                    cacheWithTransaction,
		availableMainMemoryBytes: int64(totalMainMemoryBytes),
		mu:                       sync.RWMutex{},
		totalMainMemoryBytes:     totalMainMemoryBytes,
		overCommitPercentage:     overCommitPercentage,
		metrics:                  metrics,
	}
}
