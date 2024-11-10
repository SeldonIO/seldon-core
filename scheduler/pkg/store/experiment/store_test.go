/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package experiment

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
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
			path := fmt.Sprintf("%s/db", t.TempDir())

			logger := logrus.New()
			eventHub, err := coordinator.NewEventHub(logger)
			g.Expect(err).To(BeNil())
			server := NewExperimentServer(logger, eventHub, fakeModelStore{}, fakePipelineStore{})
			// init db
			_ = server.InitialiseOrRestoreDB(path, 10)
			for _, ea := range test.experiments {
				err := server.StartExperiment(ea.experiment)
				if ea.fail {
					g.Expect(err).ToNot(BeNil())
				} else {
					g.Expect(err).To(BeNil())
					// check db
					experimentFromDB, _ := server.db.get(ea.experiment.Name)
					g.Expect(experimentFromDB.Deleted).To(BeFalse())
					g.Expect(experimentFromDB.Active).To(BeFalse()) // by default experiments are not active
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
					"a": {Name: "a"},
				},
			},
			experimentName: "a",
		},
		{
			name: "remove not existing",
			store: &ExperimentStore{
				logger: logrus.New(),
				experiments: map[string]*Experiment{
					"b": {Name: "b"},
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
					"a": {Name: "a"},
					"b": {Name: "b"},
				},
			},
			experimentName: "a",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			path := fmt.Sprintf("%s/db", t.TempDir())

			// init db
			err := test.store.InitialiseOrRestoreDB(path, 1)
			g.Expect(err).To(BeNil())
			for _, p := range test.store.experiments {
				err := test.store.db.save(p)
				g.Expect(err).To(BeNil())
			}

			err = test.store.StopExperiment(test.experimentName)
			if test.err != nil {
				_, ok := err.(*ExperimentNotFound)
				g.Expect(ok).To(BeTrue())

				// check db
				experimentFromDB, _ := test.store.db.get(test.experimentName)
				g.Expect(experimentFromDB).To(BeNil())
			} else {
				g.Expect(err).To(BeNil())

				// check experiment in store marked as deleted
				experiment, err := test.store.GetExperiment(test.experimentName)
				g.Expect(err).To(BeNil())
				g.Expect(experiment.Deleted).To(BeTrue())

				// check db
				experimentFromDB, _ := test.store.db.get(test.experimentName)
				g.Expect(experimentFromDB.Deleted).To(BeTrue())

				time.Sleep(1 * time.Second)
				test.store.cleanupDeletedExperiments()
				g.Expect(test.store.experiments[test.experimentName]).To(BeNil())
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

func TestRestoreExperiments(t *testing.T) {
	g := NewGomegaWithT(t)
	type test struct {
		name        string
		experiments map[string]*Experiment
	}

	tests := []test{
		{
			name: "running experiment",
			experiments: map[string]*Experiment{
				"a": {
					Name: "a",
					Candidates: []*Candidate{
						{
							Name: "model1",
						},
						{
							Name: "model2",
						},
					},
					Deleted: false,
				},
			},
		},
		{
			name: "deleted experiment",
			experiments: map[string]*Experiment{
				"b": {Name: "b", Deleted: true},
			},
		},
		{
			name: "mix of experiments",
			experiments: map[string]*Experiment{
				"a": {
					Name: "a",
					Candidates: []*Candidate{
						{
							Name: "model1",
						},
						{
							Name: "model2",
						},
					},
					Deleted: false,
				},
				"b": {Name: "b", Deleted: true},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			path := fmt.Sprintf("%s/db", t.TempDir())

			store := &ExperimentStore{
				logger:          logrus.New(),
				modelReferences: make(map[string]map[string]*Experiment),
				experiments:     make(map[string]*Experiment),
			}
			// init db
			err := store.InitialiseOrRestoreDB(path, 10)
			g.Expect(err).To(BeNil())
			for _, p := range test.experiments {
				err := store.db.save(p)
				g.Expect(err).To(BeNil())
			}
			_ = store.db.Stop()

			// restore from db now that we have state on disk
			_ = store.InitialiseOrRestoreDB(path, 10)

			for _, p := range test.experiments {
				experimentFromDB, _ := store.db.get(p.Name)
				g.Expect(experimentFromDB.Deleted).To(Equal(p.Deleted))
			}
			// check store
			for _, p := range store.experiments {
				expectedExperiment, ok := test.experiments[p.Name]
				g.Expect(ok).To(BeTrue())
				g.Expect(expectedExperiment.Deleted).To(Equal(p.Deleted))
				g.Expect(cmp.Equal(p.Name, expectedExperiment.Name)).To(BeTrue())
				if expectedExperiment.Deleted {
					g.Expect(expectedExperiment.DeletedAt.Before(time.Now())).To(BeTrue())
				}

				g.Expect(len(store.experiments)).To(Equal(len(test.experiments)))
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

func (f fakeModelStore) ServerNotify(request *scheduler.ServerNotify) error {
	panic("implement me")
}

func (f fakeModelStore) RemoveServerReplica(serverName string, replicaIdx int) ([]string, error) {
	panic("implement me")
}

func (f fakeModelStore) DrainServerReplica(serverName string, replicaIdx int) ([]string, error) {
	panic("implement me")
}

func (f fakeModelStore) FailedScheduling(modelVersion *store.ModelVersion, reason string, reset bool) {
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
