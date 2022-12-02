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
	"testing"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store/pipeline"

	. "github.com/onsi/gomega"
)

func TestPipelineStatusManagerUpdate(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name            string
		statusManager   *PipelineStatusManager
		pipelineVersion *pipeline.PipelineVersion
		added           bool
		deleted         bool
	}

	tests := []test{
		{
			name: "pipeline version added",
			statusManager: &PipelineStatusManager{
				pipelines: map[string]*pipeline.PipelineVersion{
					"test": {Name: "foo", State: &pipeline.PipelineState{Status: pipeline.PipelineReady}},
				},
			},
			pipelineVersion: &pipeline.PipelineVersion{Name: "bar", State: &pipeline.PipelineState{Status: pipeline.PipelineReady}},
			added:           true,
			deleted:         false,
		},
		{
			name: "pipeline version added over existing",
			statusManager: &PipelineStatusManager{
				pipelines: map[string]*pipeline.PipelineVersion{
					"test": {Name: "foo", Version: 1, State: &pipeline.PipelineState{Status: pipeline.PipelineReady}},
				},
			},
			pipelineVersion: &pipeline.PipelineVersion{Name: "foo", Version: 1, State: &pipeline.PipelineState{Status: pipeline.PipelineReady}},
			added:           true,
			deleted:         false,
		},
		{
			name: "pipeline version not added over existing",
			statusManager: &PipelineStatusManager{
				pipelines: map[string]*pipeline.PipelineVersion{
					"test": {Name: "foo", Version: 2, State: &pipeline.PipelineState{Status: pipeline.PipelineReady}},
				},
			},
			pipelineVersion: &pipeline.PipelineVersion{Name: "foo", Version: 1, State: &pipeline.PipelineState{Status: pipeline.PipelineReady}},
			added:           true,
			deleted:         false,
		},
		{
			name: "pipeline deleted",
			statusManager: &PipelineStatusManager{
				pipelines: map[string]*pipeline.PipelineVersion{
					"test": {Name: "foo", Version: 1, State: &pipeline.PipelineState{Status: pipeline.PipelineReady}},
				},
			},
			pipelineVersion: &pipeline.PipelineVersion{Name: "foo", Version: 1, State: &pipeline.PipelineState{Status: pipeline.PipelineTerminated}},
			added:           false,
			deleted:         true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.statusManager.Update(test.pipelineVersion)
			if test.deleted {
				g.Expect(test.statusManager.Get(test.pipelineVersion.Name)).To(BeNil())
			} else {
				if test.added {
					g.Expect(test.pipelineVersion).To(Equal(test.statusManager.Get(test.pipelineVersion.Name)))
				} else {
					g.Expect(test.pipelineVersion).ToNot(Equal(test.statusManager.Get(test.pipelineVersion.Name)))
				}
			}
		})
	}
}

func TestPipelineStatusManagerGet(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name          string
		statusManager *PipelineStatusManager
		pipelineName  string
		found         bool
	}

	tests := []test{
		{
			name: "exists",
			statusManager: &PipelineStatusManager{
				pipelines: map[string]*pipeline.PipelineVersion{
					"test": &pipeline.PipelineVersion{},
				},
			},
			pipelineName: "test",
			found:        true,
		},
		{
			name: "not exists",
			statusManager: &PipelineStatusManager{
				pipelines: map[string]*pipeline.PipelineVersion{
					"test": &pipeline.PipelineVersion{},
				},
			},
			pipelineName: "test1",
			found:        false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			pv := test.statusManager.Get(test.pipelineName)
			if test.found {
				g.Expect(pv).ToNot(BeNil())
			} else {
				g.Expect(pv).To(BeNil())
			}
		})
	}
}
