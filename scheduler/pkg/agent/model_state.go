package agent

import (
	"fmt"
	"sync"

	"github.com/seldonio/seldon-core/scheduler/apis/mlops/agent"
)

type ModelState struct {
	mu           sync.RWMutex
	loadedModels map[string]*ModelVersions
}

func NewModelState() *ModelState {
	return &ModelState{
		loadedModels: make(map[string]*ModelVersions),
	}
}

// Add model version, will return true if new version is added
func (modelState *ModelState) addModelVersion(modelVersionDetails *agent.ModelVersion) bool {
	// TODO: any error management here?
	modelState.mu.Lock()
	defer modelState.mu.Unlock()
	return modelState.addModelVersionImpl(modelVersionDetails)
}

func (modelState *ModelState) addModelVersionImpl(modelVersionDetails *agent.ModelVersion) bool {
	modelName := modelVersionDetails.GetModel().GetMeta().GetName()
	versionId := modelVersionDetails.GetVersion()

	existingVersions := modelState.loadedModels[modelName]
	if existingVersions == nil {
		existingVersions = NewModelVersions() // empty versions
		modelState.loadedModels[modelName] = existingVersions
	}
	if existingVersions.getModelVersionDetails(versionId) != nil {
		// model version already exist, do nothing
		// do we need to raise an error / warning ?
		return false
	}
	existingVersions.addModelVersion(modelVersionDetails)
	return true
}

// Remove model version and return true if no versions left (in which case we remove from map)
func (modelState *ModelState) removeModelVersion(modelVersionDetails *agent.ModelVersion) bool {
	modelState.mu.Lock()
	defer modelState.mu.Unlock()
	return modelState.removeModelVersionImpl(modelVersionDetails)
}

func (modelState *ModelState) removeModelVersionImpl(modelVersionDetails *agent.ModelVersion) bool {

	modelName := modelVersionDetails.GetModel().GetMeta().GetName()
	versions := modelState.loadedModels[modelName]
	if versions == nil {
		return true
	}
	versions.removeModelVersion(modelVersionDetails)

	if versions.numVersions() == 0 {
		delete(modelState.loadedModels, modelName)
		return true
	}
	return false
}

func (modelState *ModelState) getModelTotalMemoryBytes(modelId string) (uint64, error) {
	modelState.mu.RLock()
	defer modelState.mu.RUnlock()
	versions, ok := modelState.loadedModels[modelId]
	if !ok {
		return 0, fmt.Errorf("No details for model %s", modelId)
	}
	return versions.totalMemoryBytes, nil
}

func (modelState *ModelState) getModelVersionMemoryBytes(modelId string, versionId uint32) (uint64, error) {
	modelState.mu.RLock()
	defer modelState.mu.RUnlock()
	versions, ok := modelState.loadedModels[modelId]
	if !ok {
		return 0, fmt.Errorf("Model %s details not found", modelId)
	}
	versionDetails := versions.getModelVersionDetails(versionId)
	if versionDetails == nil {
		return 0, fmt.Errorf("Model %s (version %d) details not found", modelId, versionId)
	} else {
		return versionDetails.GetModel().GetModelSpec().GetMemoryBytes(), nil
	}
}

func (modelState *ModelState) versionExists(modelId string, versionId uint32) bool {
	modelState.mu.RLock()
	defer modelState.mu.RUnlock()
	versions, ok := modelState.loadedModels[modelId]
	if !ok {
		return false
	}
	versionDetails := versions.getModelVersionDetails(versionId)
	return versionDetails != nil
}

func (modelState *ModelState) numVersions(modelId string) (int, error) {
	modelState.mu.RLock()
	defer modelState.mu.RUnlock()
	versions, ok := modelState.loadedModels[modelId]
	if ok {
		return versions.numVersions(), nil
	} else {
		return 0, fmt.Errorf("Model %s details not found", modelId)
	}
}

func (modelState *ModelState) numModels() int {
	modelState.mu.RLock()
	defer modelState.mu.RUnlock()
	return len(modelState.loadedModels)
}

func (modelState *ModelState) modelNames() []string {
	modelState.mu.RLock()
	defer modelState.mu.RUnlock()
	numModels := len(modelState.loadedModels)
	models := make([]string, numModels)
	i := 0
	for name := range modelState.loadedModels {
		models[i] = name
		i++
	}
	return models
}

func (modelState *ModelState) getVersionsForAllModels() []*agent.ModelVersion {
	modelState.mu.RLock()
	defer modelState.mu.RUnlock()
	var loadedModels []*agent.ModelVersion
	for _, model := range modelState.loadedModels {
		for _, version := range model.versions {
			loadedModels = append(loadedModels, version)
		}
	}
	return loadedModels
}

// note: this should not be accessed directly
// TODO: make it lower case?
type ModelVersions struct {
	versions         map[uint32]*agent.ModelVersion
	totalMemoryBytes uint64
}

func NewModelVersions() *ModelVersions {
	return &ModelVersions{
		versions: make(map[uint32]*agent.ModelVersion),
	}
}

func (versions *ModelVersions) numVersions() int {
	return len(versions.versions)
}

func (versions *ModelVersions) computeTotalMemory() uint64 {
	var total uint64
	for _, versionDetails := range versions.versions {
		total += versionDetails.GetModel().GetModelSpec().GetMemoryBytes()
	}
	return total
}

func (versions *ModelVersions) getModelVersionDetails(versionId uint32) *agent.ModelVersion {
	versionDetails, ok := versions.versions[versionId]
	if ok {
		return versionDetails
	} else {
		return nil
	}
}

func (versions *ModelVersions) addModelVersion(versionDetails *agent.ModelVersion) {
	versions.versions[versionDetails.Version] = versionDetails
	versions.totalMemoryBytes = versions.computeTotalMemory()
}

func (versions *ModelVersions) removeModelVersion(versionDetails *agent.ModelVersion) {
	if _, ok := versions.versions[versionDetails.Version]; !ok {
		return
	}
	delete(versions.versions, versionDetails.GetVersion())
	versions.totalMemoryBytes = versions.computeTotalMemory()
}
