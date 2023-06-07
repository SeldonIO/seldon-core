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

package experiment

import (
	"errors"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"

	"github.com/seldonio/seldon-core/apis/go/v2/mlops/agent"
	"github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/coordinator"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store"
)

func TestStartExperiment(t *testing.T) {
	g := NewGomegaWithT(t)

	type experimentAddition struct {
		experiment *Experiment
		fail       bool
	}

	type test struct {
		name           string
		experiments    []*experimentAddition
		expectedNumExp int
	}

	getStrPtr := func(val string) *string { return &val }
	tests := []test{
		{
			name: "add one",
			experiments: []*experimentAddition{
				{
					experiment: &Experiment{
						Name: "a",
						Candidates: []*Candidate{
							{
								Name: "model1",
							},
						},
					},
				},
			},
			expectedNumExp: 1,
		},
		{
			name: "add two",
			experiments: []*experimentAddition{
				{
					experiment: &Experiment{
						Name: "a",
						Candidates: []*Candidate{
							{
								Name: "model1",
							},
						},
					},
				},
				{
					experiment: &Experiment{
						Name: "b",
						Candidates: []*Candidate{
							{
								Name: "model1",
							},
						},
					},
				},
			},
			expectedNumExp: 2,
		},
		{
			name: "add duplicates",
			experiments: []*experimentAddition{
				{
					experiment: &Experiment{
						Name: "a",
						Candidates: []*Candidate{
							{
								Name: "model1",
							},
						},
					},
				},
				{
					experiment: &Experiment{
						Name: "a",
						Candidates: []*Candidate{
							{
								Name: "model1",
							},
						},
					},
				},
			},
			expectedNumExp: 1,
		},
		{
			name: "add baseline experiment but no model exists",
			experiments: []*experimentAddition{
				{
					experiment: &Experiment{
						Name:    "a",
						Default: getStrPtr("model1"),
						Candidates: []*Candidate{
							{
								Name: "model1",
							},
							{
								Name: "model2",
							},
						},
					},
				},
			},
			expectedNumExp: 1,
		},
		{
			name: "add baseline experiment",
			experiments: []*experimentAddition{
				{
					experiment: &Experiment{
						Name:    "a",
						Default: getStrPtr("model1"),
						Candidates: []*Candidate{
							{
								Name: "model1",
							},
							{
								Name: "model2",
							},
						},
					},
				},
			},
			expectedNumExp: 1,
		},
		{
			name: "add baseline experiment twice to same model - not allowed",
			experiments: []*experimentAddition{
				{
					experiment: &Experiment{
						Name:    "a",
						Default: getStrPtr("model1"),
						Candidates: []*Candidate{
							{
								Name: "model1",
							},
							{
								Name: "model2",
							},
						},
					},
				},
				{
					experiment: &Experiment{
						Name:    "b",
						Default: getStrPtr("model1"),
						Candidates: []*Candidate{
							{
								Name: "model1",
							}, {
								Name: "model2",
							},
						},
					},
					fail: true,
				},
			},
			expectedNumExp: 1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			logger := logrus.New()
			eventHub, err := coordinator.NewEventHub(logger)
			g.Expect(err).To(BeNil())
			server := NewExperimentServer(logger, eventHub, fakeModelStore{}, fakePipelineStore{})
			for _, ea := range test.experiments {
				err := server.StartExperiment(ea.experiment)
				if ea.fail {
					g.Expect(err).ToNot(BeNil())
				} else {
					g.Expect(err).To(BeNil())
				}
			}
			g.Expect(len(server.experiments)).To(Equal(test.expectedNumExp))
		})
	}
}

func TestStopExperiment(t *testing.T) {
	g := NewGomegaWithT(t)
	type test struct {
		name           string
		store          *ExperimentStore
		experimentName string
		err            error
	}

	tests := []test{
		{
			name: "remove existing",
			store: &ExperimentStore{
				logger: logrus.New(),
				experiments: map[string]*Experiment{
					"a": {},
				},
			},
			experimentName: "a",
		},
		{
			name: "remove not existing",
			store: &ExperimentStore{
				logger: logrus.New(),
				experiments: map[string]*Experiment{
					"b": {},
				},
			},
			experimentName: "a",
			err:            &ExperimentNotFound{experimentName: "a"},
		},
		{
			name: "remove existing multiple",
			store: &ExperimentStore{
				logger: logrus.New(),
				experiments: map[string]*Experiment{
					"a": {},
					"b": {},
				},
			},
			experimentName: "a",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.store.StopExperiment(test.experimentName)
			if test.err != nil {
				_, ok := err.(*ExperimentNotFound)
				g.Expect(ok).To(BeTrue())
			} else {
				g.Expect(err).To(BeNil())
				experiment, err := test.store.GetExperiment(test.experimentName)
				g.Expect(err).To(BeNil())
				g.Expect(experiment.Deleted).To(BeTrue())
			}
		})
	}
}

func TestGetExperiment(t *testing.T) {
	g := NewGomegaWithT(t)
	type test struct {
		name           string
		store          *ExperimentStore
		experimentName string
		err            error
	}
	tests := []test{
		{
			name: "experiment found",
			store: &ExperimentStore{
				logger: logrus.New(),
				experiments: map[string]*Experiment{
					"a": {},
				},
			},
			experimentName: "a",
		},
		{
			name: "experiment not found",
			store: &ExperimentStore{
				logger: logrus.New(),
				experiments: map[string]*Experiment{
					"b": {},
				},
			},
			experimentName: "a",
			err:            &ExperimentNotFound{},
		},
		{
			name: "deleted",
			store: &ExperimentStore{
				logger: logrus.New(),
				experiments: map[string]*Experiment{
					"a": {Deleted: true},
				},
			},
			experimentName: "a",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			experiment, err := test.store.GetExperiment(test.experimentName)
			if test.err != nil {
				g.Expect(errors.Is(err, test.err)).To(BeTrue())
			} else {
				g.Expect(err).To(BeNil())
				g.Expect(experiment).To(Equal(test.store.experiments[test.experimentName]))
				// Change store experiment and check its a deep copy
				newName := "123"
				test.store.experiments[test.experimentName].Name = newName
				g.Expect(experiment.Name).ToNot(Equal(newName))
			}
		})
	}
}

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

func (f fakeModelStore) ServerNotify(request *scheduler.ServerNotifyRequest) error {
	panic("implement me")
}

func (f fakeModelStore) RemoveServerReplica(serverName string, replicaIdx int) ([]string, error) {
	panic("implement me")
}

func (f fakeModelStore) DrainServerReplica(serverName string, replicaIdx int) ([]string, error) {
	panic("implement me")
}

func (f fakeModelStore) FailedScheduling(modelVersion *store.ModelVersion, reason string) {
	panic("implement me")
}

func (f fakeModelStore) GetAllModels() []string {
	panic("implement me")
}

func TestHandleModelEvents(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name                    string
		experiment              *Experiment
		modelStates             map[string]store.ModelState
		modelEventMsgs          []coordinator.ModelEventMsg
		expectedCandidatesReady bool
		expectedMirrorReady     bool
	}

	tests := []test{
		{
			name: "candidate ready as model is ready",
			experiment: &Experiment{
				Name: "a",
				Candidates: []*Candidate{
					{
						Name: "model1",
					},
				},
			},
			modelStates: map[string]store.ModelState{"model1": store.ModelAvailable},
			modelEventMsgs: []coordinator.ModelEventMsg{
				{
					ModelName: "model1",
				},
			},
			expectedCandidatesReady: true,
			expectedMirrorReady:     true,
		},
		{
			name: "candidates not ready as model is not ready",
			experiment: &Experiment{
				Name: "a",
				Candidates: []*Candidate{
					{
						Name: "model1",
					},
				},
			},
			modelStates: map[string]store.ModelState{"model1": store.ModelFailed},
			modelEventMsgs: []coordinator.ModelEventMsg{
				{
					ModelName: "model1",
				},
			},
			expectedCandidatesReady: false,
			expectedMirrorReady:     true,
		},
		{
			name: "multiple candidates only one ready",
			experiment: &Experiment{
				Name: "a",
				Candidates: []*Candidate{
					{
						Name: "model1",
					},
					{
						Name: "model2",
					},
				},
			},
			modelStates: map[string]store.ModelState{"model1": store.ModelAvailable},
			modelEventMsgs: []coordinator.ModelEventMsg{
				{
					ModelName: "model1",
				},
			},
			expectedCandidatesReady: false,
			expectedMirrorReady:     true,
		},
		{
			name: "multiple candidates all ready",
			experiment: &Experiment{
				Name: "a",
				Candidates: []*Candidate{
					{
						Name: "model1",
					},
					{
						Name: "model2",
					},
				},
			},
			modelStates: map[string]store.ModelState{"model1": store.ModelAvailable, "model2": store.ModelAvailable},
			modelEventMsgs: []coordinator.ModelEventMsg{
				{
					ModelName: "model1",
				},
				{
					ModelName: "model2",
				},
			},
			expectedCandidatesReady: true,
			expectedMirrorReady:     true,
		},
		{
			name: "mirror and candidate ready as model is ready",
			experiment: &Experiment{
				Name: "a",
				Candidates: []*Candidate{
					{
						Name: "model1",
					},
				},
				Mirror: &Mirror{
					Name: "model2",
				},
			},
			modelStates: map[string]store.ModelState{"model1": store.ModelAvailable, "model2": store.ModelAvailable},
			modelEventMsgs: []coordinator.ModelEventMsg{
				{
					ModelName: "model1",
				},
			},
			expectedCandidatesReady: true,
			expectedMirrorReady:     true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			logger := logrus.New()
			eventHub, err := coordinator.NewEventHub(logger)
			g.Expect(err).To(BeNil())
			server := NewExperimentServer(logger, eventHub, fakeModelStore{status: test.modelStates}, fakePipelineStore{})
			err = server.StartExperiment(test.experiment)
			g.Expect(err).To(BeNil())
			for _, event := range test.modelEventMsgs {
				server.handleModelEvents(event)
			}
			exp, err := server.GetExperiment(test.experiment.Name)
			g.Expect(err).To(BeNil())
			g.Expect(exp.AreCandidatesReady()).To(Equal(test.expectedCandidatesReady))
			g.Expect(exp.IsMirrorReady()).To(Equal(test.expectedMirrorReady))
		})
	}
}
