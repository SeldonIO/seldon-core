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

package status

import (
	"sync"

	pipeline "github.com/seldonio/seldon-core/scheduler/v2/pkg/store/pipeline"
)

type PipelineStatusUpdater interface {
	Update(version *pipeline.PipelineVersion)
}

type PipelineStatusProvider interface {
	Get(name string) *pipeline.PipelineVersion
}

type PipelineStatusManager struct {
	mu        sync.RWMutex
	pipelines map[string]*pipeline.PipelineVersion
}

func NewPipelineStatusManager() *PipelineStatusManager {
	return &PipelineStatusManager{
		pipelines: make(map[string]*pipeline.PipelineVersion),
	}
}

func (sm *PipelineStatusManager) Update(version *pipeline.PipelineVersion) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	if pv, ok := sm.pipelines[version.Name]; ok {
		if version.Version < pv.Version { //ignore older versions
			return
		}
	}
	// Remove or update the pipeline
	if version.State != nil && version.State.Status == pipeline.PipelineTerminated {
		delete(sm.pipelines, version.Name)
	} else {
		sm.pipelines[version.Name] = version
	}
}

func (sm *PipelineStatusManager) Get(name string) *pipeline.PipelineVersion {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.pipelines[name]
}
