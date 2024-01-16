/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package status

import (
	"testing"

	. "github.com/onsi/gomega"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store/pipeline"
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
