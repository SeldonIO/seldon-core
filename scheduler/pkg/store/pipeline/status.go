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

package pipeline

import (
	"sync"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/coordinator"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store"
	"github.com/sirupsen/logrus"
)

type void struct{}

var member void

// Handle model status updates that affect Pipelines
type ModelStatusHandler struct {
	mu              sync.RWMutex
	logger          logrus.FieldLogger
	store           store.ModelStore
	modelReferences map[string]map[string]void
}

// Set pipeline model readiness
// Setup references so we can update when model status' change
func (ms *ModelStatusHandler) addPipelineModelStatus(pipeline *Pipeline) error {
	err := ms.setPipelineModelsReady(pipeline.GetLatestPipelineVersion())
	if err != nil {
		return err
	}
	ms.addModelReferences(pipeline)
	return nil
}

// Change a pipeline model readiness based on a new model status
// 1. if model is not ready pipeline can't be
// 2. if model is ready then check all models are ready
func updatePipelineModelsReady(latestPipeline *PipelineVersion, modelAvailable bool, loggerIn logrus.FieldLogger) {
	logger := loggerIn.WithField("func", "updatePipelineModelsReady")
	existingState := latestPipeline.State.ModelsReady
	if latestPipeline.State.ModelsReady && !modelAvailable {
		latestPipeline.State.ModelsReady = false
	} else if !latestPipeline.State.ModelsReady && modelAvailable {
		stepsReady := true
		for _, pstep := range latestPipeline.Steps {
			if !pstep.Available {
				stepsReady = false
				break
			}
		}
		latestPipeline.State.ModelsReady = stepsReady
	}
	logger.Debugf("Pipeline %s models ready was %v and is now %v", latestPipeline.Name, existingState, latestPipeline.State.ModelsReady)
}

// Set pipeline models ready by finding out if all models are ready
func (ms *ModelStatusHandler) setPipelineModelsReady(pipelineVersion *PipelineVersion) error {
	modelsReady := true
	if pipelineVersion != nil && ms.store != nil {
		for stepName, step := range pipelineVersion.Steps {
			model, err := ms.store.GetModel(stepName)
			if err != nil {
				return err
			}
			step.Available = model != nil && model.GetLastAvailableModel() != nil
			if !step.Available {
				modelsReady = false
			}
		}
		pipelineVersion.State.ModelsReady = modelsReady
	}
	return nil
}

// Find and set Pipeline Model Ready due to a Model whose status has changed
func updatePipelinesFromModelAvailability(references map[string]void, modelName string, modelAvailable bool, pipelines map[string]*Pipeline, loggerIn logrus.FieldLogger) []*coordinator.PipelineEventMsg {
	logger := loggerIn.WithField("func", "updatePipelinesFromModelAvailability")
	logger.Debugf("Updating pipeline state from model %s available:%v", modelName, modelAvailable)
	var changedPipelines []*coordinator.PipelineEventMsg
	for pipelineName := range references {
		if pipeline, ok := pipelines[pipelineName]; ok {
			latestPipeline := pipeline.GetLatestPipelineVersion()
			if latestPipeline != nil {
				if step, ok := latestPipeline.Steps[modelName]; ok {
					if step.Available != modelAvailable {
						changedPipelines = append(changedPipelines, &coordinator.PipelineEventMsg{
							PipelineName:      latestPipeline.Name,
							PipelineVersion:   latestPipeline.Version,
							UID:               latestPipeline.UID,
							ModelStatusChange: true,
						})
						step.Available = modelAvailable
						logger.Debugf("Updating pipeline overall state from model %s available:%v", modelName, modelAvailable)
						updatePipelineModelsReady(latestPipeline, modelAvailable, logger)
					} else {
						logger.Debugf("Ignore update to step %s for pipeline %s as already %v", modelName, pipelineName, modelAvailable)
					}
				} else {
					logger.Debugf("Failed to find step %s in pipeline %s", modelName, pipelineName)
				}
			} else {
				logger.Debugf("Failed to find latest pipeline %s", pipelineName)
			}
		} else {
			logger.Debugf("Failed to find latest pipeline %s", pipelineName)
		}
	}
	return changedPipelines
}

func (ms *ModelStatusHandler) removeModelReferences(pv *PipelineVersion) {
	if pv != nil {
		for step := range pv.Steps {
			if members, ok := ms.modelReferences[step]; ok {
				delete(members, pv.Name)
				if len(members) == 0 {
					delete(ms.modelReferences, step)
				}
			}
		}
	}
}

func (ms *ModelStatusHandler) addModelReferences(pipeline *Pipeline) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	if pipeline != nil {
		// remove any existing references from previous version if it exists
		ms.removeModelReferences(pipeline.GetPreviousPipelineVersion())
		latestVersion := pipeline.GetLatestPipelineVersion()
		if latestVersion != nil {
			for step := range latestVersion.Steps {
				if members, ok := ms.modelReferences[step]; !ok {
					members = make(map[string]void)
					members[pipeline.Name] = member
					ms.modelReferences[step] = members
				} else {
					members[pipeline.Name] = member
				}
			}
		}
	}
}
