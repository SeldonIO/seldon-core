/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
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
