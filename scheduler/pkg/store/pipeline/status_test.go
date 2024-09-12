/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package pipeline

import (
	"testing"

	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"

	"github.com/seldonio/seldon-core/apis/go/v2/mlops/agent"
	"github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store"
)

type fakeModelStore struct {
	status map[string]store.ModelState
}

var _ store.ModelStore = (*fakeModelStore)(nil)

func (f fakeModelStore) UpdateModel(config *scheduler.LoadModelRequest) error {
	panic("implement me")
}

func (f fakeModelStore) GetModel(key string) (*store.ModelSnapshot, error) {
	return &store.ModelSnapshot{
		Name: key,
		Versions: []*store.ModelVersion{
			store.NewModelVersion(nil, 1, "server", nil, false, f.status[key]),
		},
	}, nil
}

func (f fakeModelStore) GetModels() ([]*store.ModelSnapshot, error) {
	panic("implement me")
}

func (f fakeModelStore) LockModel(modelId string) {
	panic("implement me")
}

func (f fakeModelStore) UnlockModel(modelId string) {
	panic("implement me")
}

func (f fakeModelStore) RemoveModel(req *scheduler.UnloadModelRequest) error {
	panic("implement me")
}

func (f fakeModelStore) GetServers(shallow bool, modelDetails bool) ([]*store.ServerSnapshot, error) {
	panic("implement me")
}

func (f fakeModelStore) GetServer(serverKey string, shallow bool, modelDetails bool) (*store.ServerSnapshot, error) {
	panic("implement me")
}

func (f fakeModelStore) UpdateLoadedModels(modelKey string, version uint32, serverKey string, replicas []*store.ServerReplica) error {
	panic("implement me")
}

func (f fakeModelStore) UnloadVersionModels(modelKey string, version uint32) (bool, error) {
	panic("implement me")
}

func (f fakeModelStore) UpdateModelState(modelKey string, version uint32, serverKey string, replicaIdx int, availableMemory *uint64, expectedState, desiredState store.ModelReplicaState, reason string) error {
	panic("implement me")
}

func (f fakeModelStore) AddServerReplica(request *agent.AgentSubscribeRequest) error {
	panic("implement me")
}

func (f fakeModelStore) ServerNotify(request *scheduler.ServerNotify) error {
	panic("implement me")
}

func (f fakeModelStore) RemoveServerReplica(serverName string, replicaIdx int) ([]string, error) {
	panic("implement me")
}

func (f fakeModelStore) FailedScheduling(modelVersion *store.ModelVersion, reason string, reset bool) {
	panic("implement me")
}

func (f fakeModelStore) GetAllModels() []string {
	panic("implement me")
}

func (f fakeModelStore) DrainServerReplica(serverName string, replicaIdx int) ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func TestUpdatePipelineModelAvailable(t *testing.T) {
	g := NewGomegaWithT(t)
	type test struct {
		name                  string
		references            map[string]void
		pipelines             map[string]*Pipeline
		modelName             string
		available             bool
		expectedChanged       map[string]void
		expectedPipelineReady bool
	}

	tests := []test{
		{
			name:       "single pipeline changed - model available",
			references: map[string]void{"p1": member},
			pipelines: map[string]*Pipeline{"p1": {
				Name: "p1",
				Versions: []*PipelineVersion{
					{
						Name: "p1",
						Steps: map[string]*PipelineStep{
							"model1": {Available: false},
						},
						State: &PipelineState{ModelsReady: false},
					},
				},
			}},
			modelName:             "model1",
			available:             true,
			expectedChanged:       map[string]void{"p1": member},
			expectedPipelineReady: true,
		},
		{
			name:       "single pipeline changed - model unavailable",
			references: map[string]void{"p1": member},
			pipelines: map[string]*Pipeline{"p1": {
				Name: "p1",
				Versions: []*PipelineVersion{
					{
						Name: "p1",
						Steps: map[string]*PipelineStep{
							"model1": {Available: true},
						},
						State: &PipelineState{ModelsReady: true},
					},
				},
			}},
			modelName:             "model1",
			available:             false,
			expectedChanged:       map[string]void{"p1": member},
			expectedPipelineReady: false,
		},
		{
			name:       "single pipeline not changed - model available",
			references: map[string]void{"p1": member},
			pipelines: map[string]*Pipeline{"p1": {
				Name: "p1",
				Versions: []*PipelineVersion{
					{
						Name: "p1",
						Steps: map[string]*PipelineStep{
							"model1": {Available: true},
						},
						State: &PipelineState{ModelsReady: true},
					},
				},
			}},
			modelName:             "model1",
			available:             true,
			expectedChanged:       map[string]void{},
			expectedPipelineReady: true,
		},
		{
			name:       "single pipeline changed - two steps - model available - pipeline not available",
			references: map[string]void{"p1": member},
			pipelines: map[string]*Pipeline{"p1": {
				Name: "p1",
				Versions: []*PipelineVersion{
					{
						Name: "p1",
						Steps: map[string]*PipelineStep{
							"model1": {Available: false},
							"model2": {Available: false},
						},
						State: &PipelineState{ModelsReady: false},
					},
				},
			}},
			modelName:             "model1",
			available:             true,
			expectedChanged:       map[string]void{"p1": member},
			expectedPipelineReady: false,
		},
		{
			name:       "single pipeline changed - two steps - model available - pipeline available",
			references: map[string]void{"p1": member},
			pipelines: map[string]*Pipeline{"p1": {
				Name: "p1",
				Versions: []*PipelineVersion{
					{
						Name: "p1",
						Steps: map[string]*PipelineStep{
							"model1": {Available: false},
							"model2": {Available: true},
						},
						State: &PipelineState{ModelsReady: false},
					},
				},
			}},
			modelName:             "model1",
			available:             true,
			expectedChanged:       map[string]void{"p1": member},
			expectedPipelineReady: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			changed := updatePipelinesFromModelAvailability(test.references, test.modelName, test.available, test.pipelines, logrus.New())
			for _, evt := range changed {
				_, ok := test.expectedChanged[evt.PipelineName]
				g.Expect(ok).To(BeTrue())
				g.Expect(test.pipelines[evt.PipelineName].GetLatestPipelineVersion().Steps[test.modelName].Available).To(Equal(test.available))
				g.Expect(test.pipelines[evt.PipelineName].GetLatestPipelineVersion().State.ModelsReady).To(Equal(test.expectedPipelineReady))
			}
			g.Expect(len(changed)).To(Equal(len(test.expectedChanged)))
		})
	}
}

func TestAddModelReferences(t *testing.T) {
	g := NewGomegaWithT(t)
	type test struct {
		name                   string
		modelStatusHandler     *ModelStatusHandler
		pipeline               *Pipeline
		expectedPipelineCounts map[string]int
	}

	tests := []test{
		{
			name: "empty references",
			modelStatusHandler: &ModelStatusHandler{
				logger:          logrus.New(),
				modelReferences: map[string]map[string]void{},
			},
			pipeline: &Pipeline{
				Name: "test",
				Versions: []*PipelineVersion{
					{
						Name: "test",
						Steps: map[string]*PipelineStep{
							"model1": nil,
						},
					},
				},
			},
			expectedPipelineCounts: map[string]int{
				"model1": 1,
			},
		},
		{
			name: "new reference to model that already has one",
			modelStatusHandler: &ModelStatusHandler{
				logger: logrus.New(),
				modelReferences: map[string]map[string]void{
					"model1": {"p2": member},
				},
			},
			pipeline: &Pipeline{
				Name: "test",
				Versions: []*PipelineVersion{
					{
						Name: "test",
						Steps: map[string]*PipelineStep{
							"model1": nil,
						},
					},
				},
			},
			expectedPipelineCounts: map[string]int{
				"model1": 2,
			},
		},
		{
			name: "reference to model that already has one",
			modelStatusHandler: &ModelStatusHandler{
				logger: logrus.New(),
				modelReferences: map[string]map[string]void{
					"model1": {"test": member},
				},
			},
			pipeline: &Pipeline{
				Name: "test",
				Versions: []*PipelineVersion{
					{
						Name: "test",
						Steps: map[string]*PipelineStep{
							"model1": nil,
						},
					},
				},
			},
			expectedPipelineCounts: map[string]int{
				"model1": 1,
			},
		},
		{
			name: "multiple steps",
			modelStatusHandler: &ModelStatusHandler{
				logger: logrus.New(),
				modelReferences: map[string]map[string]void{
					"model1":   {"p2": member, "p3": member},
					"model100": {},
				},
			},
			pipeline: &Pipeline{
				Name: "test",
				Versions: []*PipelineVersion{
					{
						Name: "test",
						Steps: map[string]*PipelineStep{
							"model1": nil,
							"model2": nil,
							"model3": nil,
						},
					},
				},
			},
			expectedPipelineCounts: map[string]int{
				"model1": 3,
				"model2": 1,
				"model3": 1,
			},
		},
		{
			name: "multiple steps - previous version",
			modelStatusHandler: &ModelStatusHandler{
				logger: logrus.New(),
				modelReferences: map[string]map[string]void{
					"model1":   {"p2": member, "p3": member, "test": member},
					"model2":   {"test": member},
					"model3":   {"test": member},
					"model4":   {"test": member},
					"model100": {},
				},
			},
			pipeline: &Pipeline{
				Name: "test",
				Versions: []*PipelineVersion{
					{
						Name: "test",
						Steps: map[string]*PipelineStep{
							"model1": nil,
							"model2": nil,
							"model3": nil,
						},
					},
					{
						Name: "test",
						Steps: map[string]*PipelineStep{
							"model1": nil,
							"model2": nil,
							"model3": nil,
							"model4": nil,
						},
					},
				},
			},
			expectedPipelineCounts: map[string]int{
				"model1": 3,
				"model2": 1,
				"model3": 1,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.modelStatusHandler.addModelReferences(test.pipeline)
			for modelName, count := range test.expectedPipelineCounts {
				g.Expect(len(test.modelStatusHandler.modelReferences[modelName])).To(Equal(count))
			}
		})
	}
}

func TestRemoveModelReferences(t *testing.T) {
	g := NewGomegaWithT(t)
	type test struct {
		name                       string
		modelStatusHandler         *ModelStatusHandler
		pipelineVersion            *PipelineVersion
		expectedPipelineCounts     map[string]int
		expectedPipelineReferences int
	}

	tests := []test{
		{
			name: "pipeline removed - no references left",
			modelStatusHandler: &ModelStatusHandler{
				logger: logrus.New(),
				modelReferences: map[string]map[string]void{
					"model1": {"test": member},
				},
			},
			pipelineVersion:            nil,
			expectedPipelineCounts:     map[string]int{"model1": 1},
			expectedPipelineReferences: 1,
		},
		{
			name: "pipeline removed - no references left",
			modelStatusHandler: &ModelStatusHandler{
				logger: logrus.New(),
				modelReferences: map[string]map[string]void{
					"model1": {"test": member},
				},
			},
			pipelineVersion: &PipelineVersion{
				Name: "test",
				Steps: map[string]*PipelineStep{
					"model1": nil,
				},
			},
			expectedPipelineCounts:     map[string]int{},
			expectedPipelineReferences: 0,
		},
		{
			name: "pipeline removed - some references left",
			modelStatusHandler: &ModelStatusHandler{
				logger: logrus.New(),
				modelReferences: map[string]map[string]void{
					"model1": {"test": member, "test2": member},
				},
			},
			pipelineVersion: &PipelineVersion{
				Name: "test",
				Steps: map[string]*PipelineStep{
					"model1": nil,
				},
			},
			expectedPipelineCounts: map[string]int{
				"model1": 1,
			},
			expectedPipelineReferences: 1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.modelStatusHandler.removeModelReferences(test.pipelineVersion)
			for modelName, count := range test.expectedPipelineCounts {
				g.Expect(len(test.modelStatusHandler.modelReferences[modelName])).To(Equal(count))
			}
			g.Expect(len(test.modelStatusHandler.modelReferences)).To(Equal(test.expectedPipelineReferences))
		})
	}
}
