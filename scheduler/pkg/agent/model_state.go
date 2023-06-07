/*
Copyright 2022 Seldon Technologies Ltd.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package agent

import (
	"fmt"
	"sync"

	"github.com/seldonio/seldon-core/apis/go/v2/mlops/agent"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/util"
)

type ModelState struct {
	mu                      sync.RWMutex
	loadedModels            map[string]*modelVersion
	totalMemoryForAllModels uint64
}

func NewModelState() *ModelState {
	return &ModelState{
		mu:                      sync.RWMutex{},
		loadedModels:            make(map[string]*modelVersion),
		totalMemoryForAllModels: 0,
	}
}

// Add model version, will return true if new version is added
func (modelState *ModelState) addModelVersion(modelVersionDetails *agent.ModelVersion) (bool, error) {
	// TODO: any error management here?
	modelState.mu.Lock()
	defer modelState.mu.Unlock()
	return modelState.addModelVersionImpl(modelVersionDetails)
}

func (modelState *ModelState) addModelVersionImpl(modelVersionDetails *agent.ModelVersion) (bool, error) {
	modelName := modelVersionDetails.GetModel().GetMeta().GetName()
	versionId := modelVersionDetails.GetVersion()

	exsistingVersion, ok := modelState.loadedModels[modelName]
	if !ok {
		newVersionDetails := &modelVersion{versionInfo: modelVersionDetails}
		modelState.totalMemoryForAllModels += newVersionDetails.getVersionMemory()
		modelState.loadedModels[modelName] = newVersionDetails
		return true, nil
	} else {
		if exsistingVersion.getVersion() == versionId {
			return false, nil
		} else {
			return false, fmt.Errorf(
				"Version number mismatch for model %s (%d vs %d)",
				modelName, versionId, exsistingVersion.getVersion())
		}
	}

}

// Remove model version and return true if no versions left (in which case we remove from map)
func (modelState *ModelState) removeModelVersion(modelVersionDetails *agent.ModelVersion) (bool, error) {
	modelState.mu.Lock()
	defer modelState.mu.Unlock()
	return modelState.removeModelVersionImpl(modelVersionDetails)
}

func (modelState *ModelState) removeModelVersionImpl(modelVersionDetails *agent.ModelVersion) (bool, error) {

	modelName := modelVersionDetails.GetModel().GetMeta().GetName()
	versionId := modelVersionDetails.GetVersion()

	exsistingVersion, ok := modelState.loadedModels[modelName]
	if !ok {
		return true, nil
	}
	if exsistingVersion.getVersion() == versionId {
		modelState.totalMemoryForAllModels -= exsistingVersion.getVersionMemory()
		delete(modelState.loadedModels, modelName)
		return true, nil
	} else {
		return false, fmt.Errorf(
			"Version number mismatch for model %s (%d vs %d)",
			modelName, versionId, exsistingVersion.getVersion())
	}
}

func (modelState *ModelState) getModelMemoryBytes(modelId string) (uint64, error) {
	modelState.mu.RLock()
	defer modelState.mu.RUnlock()
	exsistingVersion, ok := modelState.loadedModels[modelId]
	if !ok {
		return 0, fmt.Errorf("No details for model %s", modelId)
	}
	return exsistingVersion.getVersionMemory(), nil
}

func (modelState *ModelState) versionExists(modelId string, versionId uint32) bool {
	modelState.mu.RLock()
	defer modelState.mu.RUnlock()
	version, ok := modelState.loadedModels[modelId]
	if !ok {
		return false
	}
	return version.getVersion() == versionId
}

// this includes the ones that are evicted as we do not differentiate at this level
func (modelState *ModelState) getTotalMemoryBytesForAllModels() uint64 {
	modelState.mu.RLock()
	defer modelState.mu.RUnlock()
	return modelState.totalMemoryForAllModels
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
	for _, version := range modelState.loadedModels {
		mv := version.get()
		versionedModelName := mv.Model.GetMeta().Name
		originalModelName, originalModelVersion, _ := util.GetOrignalModelNameAndVersion(versionedModelName)
		loadedModels = append(loadedModels, getModifiedModelVersion(originalModelName, originalModelVersion, mv))
	}
	return loadedModels
}

type modelVersion struct {
	versionInfo *agent.ModelVersion
}

func (version *modelVersion) getVersionMemory() uint64 {
	return version.versionInfo.GetModel().GetModelSpec().GetMemoryBytes()
}

func (version *modelVersion) getVersion() uint32 {
	return version.versionInfo.GetVersion()
}

func (version *modelVersion) get() *agent.ModelVersion {
	return version.versionInfo
}
