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

package scheduler

import (
	"sort"
	"testing"

	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"

	"github.com/seldonio/seldon-core/apis/go/v2/mlops/agent"
	pb "github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store"
)

type mockStore struct {
	models            map[string]*store.ModelSnapshot
	servers           []*store.ServerSnapshot
	scheduledServer   string
	scheduledReplicas []int
}

var _ store.ModelStore = (*mockStore)(nil)

func (f mockStore) FailedScheduling(modelVersion *store.ModelVersion, reason string, reset bool) {
}

func (f mockStore) UnloadVersionModels(modelKey string, version uint32) (bool, error) {
	return true, nil
}

func (f mockStore) ServerNotify(request *pb.ServerNotifyRequest) error {
	return nil
}

func (f mockStore) RemoveModel(req *pb.UnloadModelRequest) error {
	return nil
}

func (f mockStore) UpdateModel(config *pb.LoadModelRequest) error {
	return nil
}

func (f mockStore) GetModel(key string) (*store.ModelSnapshot, error) {
	return f.models[key], nil
}

func (f mockStore) GetModels() ([]*store.ModelSnapshot, error) {
	models := []*store.ModelSnapshot{}
	for _, m := range f.models {
		models = append(models, m)
	}
	return models, nil
}

func (f mockStore) LockModel(modelId string) {
}

func (f mockStore) UnlockModel(modelId string) {
}

func (f mockStore) ExistsModelVersion(key string, version uint32) bool {
	return false
}

func (f mockStore) GetServers(shallow bool, modelDetails bool) ([]*store.ServerSnapshot, error) {
	return f.servers, nil
}

func (f mockStore) GetServer(serverKey string, shallow bool, modelDetails bool) (*store.ServerSnapshot, error) {
	panic("implement me")
}

func (m *mockStore) GetAllModels() []string {
	var modelNames []string
	for modelName := range m.models {
		modelNames = append(modelNames, modelName)
	}
	return modelNames
}

func (f *mockStore) UpdateLoadedModels(modelKey string, version uint32, serverKey string, replicas []*store.ServerReplica) error {
	f.scheduledServer = serverKey
	var replicaIdxs []int
	for _, rep := range replicas {
		replicaIdxs = append(replicaIdxs, rep.GetReplicaIdx())
	}
	f.scheduledReplicas = replicaIdxs
	return nil
}

func (f mockStore) UpdateModelState(modelKey string, version uint32, serverKey string, replicaIdx int, availableMemory *uint64, expectedState, desiredState store.ModelReplicaState, reason string) error {
	panic("implement me")
}

func (f mockStore) AddServerReplica(request *agent.AgentSubscribeRequest) error {
	panic("implement me")
}

func (f mockStore) RemoveServerReplica(serverName string, replicaIdx int) ([]string, error) {
	panic("implement me")
}

func (f mockStore) DrainServerReplica(serverName string, replicaIdx int) ([]string, error) {
	panic("implement me")
}

func (f mockStore) AddModelEventListener(c chan *store.ModelSnapshot) {

}

func (f mockStore) AddServerEventListener(c chan string) {

}

func TestScheduler(t *testing.T) {
	logger := log.New()
	g := NewGomegaWithT(t)

	newTestModel := func(name string, requiredMemory uint64, requirements []string, server *string, replicas uint32, loadedModels []int, deleted bool, scheduledServer string, drainedModels []int) *store.ModelSnapshot {
		config := &pb.Model{ModelSpec: &pb.ModelSpec{MemoryBytes: &requiredMemory, Requirements: requirements, Server: server}, DeploymentSpec: &pb.DeploymentSpec{Replicas: replicas}}
		rmap := make(map[int]store.ReplicaStatus)
		for _, ridx := range loadedModels {
			rmap[ridx] = store.ReplicaStatus{State: store.Loaded}
		}
		for _, ridx := range drainedModels {
			rmap[ridx] = store.ReplicaStatus{State: store.Draining}
		}
		return &store.ModelSnapshot{
			Name:     name,
			Versions: []*store.ModelVersion{store.NewModelVersion(config, 1, scheduledServer, rmap, false, store.ModelProgressing)},
			Deleted:  deleted,
		}
	}

	gsr := func(replicaIdx int, availableMemory uint64, capabilities []string, serverName string, shared, isDraining bool) *store.ServerReplica {
		replica := store.NewServerReplica("svc", 8080, 5001, replicaIdx, store.NewServer(serverName, shared), capabilities, availableMemory, availableMemory, 0, nil, 100)
		if isDraining {
			replica.SetIsDraining()
		}
		return replica
	}

	type test struct {
		name              string
		model             *store.ModelSnapshot
		servers           []*store.ServerSnapshot
		scheduled         bool
		scheduledServer   string
		scheduledReplicas []int
	}

	tests := []test{
		{
			name:  "SmokeTest",
			model: newTestModel("model1", 100, []string{"sklearn"}, nil, 1, []int{}, false, "", nil),
			servers: []*store.ServerSnapshot{
				{
					Name:             "server1",
					Replicas:         map[int]*store.ServerReplica{0: gsr(0, 200, []string{"sklearn"}, "server1", true, false)}, // expect schedule here
					Shared:           true,
					ExpectedReplicas: -1,
				},
			},
			scheduled:         true,
			scheduledServer:   "server1",
			scheduledReplicas: []int{0},
		},
		{
			name:  "ReplicasTwo",
			model: newTestModel("model1", 100, []string{"sklearn"}, nil, 2, []int{}, false, "", nil),
			servers: []*store.ServerSnapshot{
				{
					Name:             "server1",
					Replicas:         map[int]*store.ServerReplica{0: gsr(0, 200, []string{"sklearn"}, "server1", true, false)},
					Shared:           true,
					ExpectedReplicas: -1,
				},
				{
					Name: "server2",
					Replicas: map[int]*store.ServerReplica{
						0: gsr(0, 200, []string{"sklearn"}, "server2", true, false), // expect schedule here
						1: gsr(1, 200, []string{"sklearn"}, "server2", true, false), // expect schedule here
					},
					Shared:           true,
					ExpectedReplicas: -1,
				},
			},
			scheduled:         true,
			scheduledServer:   "server2",
			scheduledReplicas: []int{0, 1},
		},
		{
			name:  "NotEnoughReplicas",
			model: newTestModel("model1", 100, []string{"sklearn"}, nil, 2, []int{}, false, "", nil),
			servers: []*store.ServerSnapshot{
				{
					Name:             "server1",
					Replicas:         map[int]*store.ServerReplica{0: gsr(0, 200, []string{"sklearn"}, "server1", true, false)},
					Shared:           true,
					ExpectedReplicas: -1,
				},
				{
					Name: "server2",
					Replicas: map[int]*store.ServerReplica{
						0: gsr(0, 200, []string{"sklearn"}, "server2", true, false),
						1: gsr(1, 0, []string{"sklearn"}, "server2", true, false),
					},
					Shared:           true,
					ExpectedReplicas: -1,
				},
			},
			scheduled: false,
		},
		{
			name:  "MemoryOneServer",
			model: newTestModel("model1", 100, []string{"sklearn"}, nil, 1, []int{}, false, "", nil),
			servers: []*store.ServerSnapshot{
				{
					Name:             "server1",
					Replicas:         map[int]*store.ServerReplica{0: gsr(0, 50, []string{"sklearn"}, "server1", true, false)},
					Shared:           true,
					ExpectedReplicas: -1,
				},
				{
					Name: "server2",
					Replicas: map[int]*store.ServerReplica{
						0: gsr(0, 200, []string{"sklearn"}, "server2", true, false), // expect schedule here
					},
					Shared:           true,
					ExpectedReplicas: -1,
				},
			},
			scheduled:         true,
			scheduledServer:   "server2",
			scheduledReplicas: []int{0},
		},
		{
			name:  "ModelsLoaded",
			model: newTestModel("model1", 100, []string{"sklearn"}, nil, 2, []int{1}, false, "", nil),
			servers: []*store.ServerSnapshot{
				{
					Name:             "server1",
					Replicas:         map[int]*store.ServerReplica{0: gsr(0, 50, []string{"sklearn"}, "server1", true, false)},
					Shared:           true,
					ExpectedReplicas: -1,
				},
				{
					Name: "server2",
					Replicas: map[int]*store.ServerReplica{
						0: gsr(0, 200, []string{"sklearn"}, "server2", true, false), // expect schedule here
						1: gsr(1, 200, []string{"sklearn"}, "server2", true, false), // expect schedule here
					},
					Shared:           true,
					ExpectedReplicas: -1,
				},
			},
			scheduled:         true,
			scheduledServer:   "server2",
			scheduledReplicas: []int{1, 0},
		},
		{
			name:  "ModelUnLoaded",
			model: newTestModel("model1", 100, []string{"sklearn"}, nil, 2, []int{1}, true, "server2", nil),
			servers: []*store.ServerSnapshot{
				{
					Name: "server2",
					Replicas: map[int]*store.ServerReplica{
						0: gsr(0, 200, []string{"sklearn"}, "server2", true, false),
						1: gsr(1, 200, []string{"sklearn"}, "server2", true, false),
					},
					Shared:           true,
					ExpectedReplicas: -1,
				},
			},
			scheduled:         true,
			scheduledServer:   "server2",
			scheduledReplicas: nil,
		},
		{
			name:  "DeletedServer",
			model: newTestModel("model1", 100, []string{"sklearn"}, nil, 1, []int{}, false, "", nil),
			servers: []*store.ServerSnapshot{
				{
					Name:             "server1",
					Replicas:         map[int]*store.ServerReplica{0: gsr(0, 200, []string{"sklearn"}, "server1", true, false)},
					Shared:           true,
					ExpectedReplicas: 0,
				},
				{
					Name: "server2",
					Replicas: map[int]*store.ServerReplica{
						0: gsr(0, 200, []string{"sklearn"}, "server2", true, false), // expect schedule here
						1: gsr(1, 200, []string{"sklearn"}, "server2", true, false),
					},
					Shared:           true,
					ExpectedReplicas: -1,
				},
			},
			scheduled:         true,
			scheduledServer:   "server2",
			scheduledReplicas: []int{0},
		},
		{
			name:  "Reschedule",
			model: newTestModel("model1", 100, []string{"sklearn"}, nil, 1, []int{0}, false, "server1", nil),
			servers: []*store.ServerSnapshot{
				{
					Name:             "server1",
					Replicas:         map[int]*store.ServerReplica{0: gsr(0, 200, []string{"sklearn"}, "server1", true, false)},
					Shared:           true,
					ExpectedReplicas: 0,
				},
				{
					Name: "server2",
					Replicas: map[int]*store.ServerReplica{
						0: gsr(0, 200, []string{"sklearn"}, "server2", true, false), // expect schedule here
						1: gsr(1, 200, []string{"sklearn"}, "server2", true, false),
					},
					Shared:           true,
					ExpectedReplicas: -1,
				},
			},
			scheduled:         true,
			scheduledServer:   "server2",
			scheduledReplicas: []int{0},
		},
		{
			name:  "DeletedServerFail",
			model: newTestModel("model1", 100, []string{"sklearn"}, nil, 1, []int{1}, false, "", nil),
			servers: []*store.ServerSnapshot{
				{
					Name:             "server1",
					Replicas:         map[int]*store.ServerReplica{0: gsr(0, 200, []string{"sklearn"}, "server1", true, false)},
					Shared:           true,
					ExpectedReplicas: 0,
				},
			},
			scheduled: false,
		},
		{
			name:  "Available memory sorting",
			model: newTestModel("model1", 100, []string{"sklearn"}, nil, 1, []int{1}, false, "", nil),
			servers: []*store.ServerSnapshot{
				{
					Name: "server2",
					Replicas: map[int]*store.ServerReplica{
						0: gsr(0, 150, []string{"sklearn"}, "server2", true, false),
						1: gsr(1, 200, []string{"sklearn"}, "server2", true, false), // expect schedule here
					},
					Shared:           true,
					ExpectedReplicas: -1,
				},
			},
			scheduled:         true,
			scheduledServer:   "server2",
			scheduledReplicas: []int{1},
		},
		{
			name:  "Available memory sorting with multiple replicas",
			model: newTestModel("model1", 100, []string{"sklearn"}, nil, 2, []int{1}, false, "", nil),
			servers: []*store.ServerSnapshot{
				{
					Name: "server2",
					Replicas: map[int]*store.ServerReplica{
						0: gsr(0, 150, []string{"sklearn"}, "server2", true, false),
						1: gsr(1, 200, []string{"sklearn"}, "server2", true, false), //expect schedule here
						2: gsr(2, 175, []string{"sklearn"}, "server2", true, false), //expect schedule here
					},
					Shared:           true,
					ExpectedReplicas: -1,
				},
			},
			scheduled:         true,
			scheduledServer:   "server2",
			scheduledReplicas: []int{1, 2},
		},
		{
			name:  "Scale up",
			model: newTestModel("model1", 100, []string{"sklearn"}, nil, 3, []int{1, 2}, false, "server1", nil),
			servers: []*store.ServerSnapshot{
				{
					Name: "server1",
					Replicas: map[int]*store.ServerReplica{
						0: gsr(0, 50, []string{"sklearn"}, "server1", true, false),
						1: gsr(1, 200, []string{"sklearn"}, "server1", true, false), //expect schedule here - nop
						2: gsr(2, 175, []string{"sklearn"}, "server1", true, false), //expect schedule here - nop
						3: gsr(3, 100, []string{"sklearn"}, "server1", true, false), //expect schedule here
					},
					Shared:           true,
					ExpectedReplicas: -1,
				},
			},
			scheduled:         true,
			scheduledServer:   "server1",
			scheduledReplicas: []int{1, 2, 3},
		},
		{
			name:  "Scale down",
			model: newTestModel("model1", 100, []string{"sklearn"}, nil, 1, []int{1, 2}, false, "server1", nil),
			servers: []*store.ServerSnapshot{
				{
					Name: "server1",
					Replicas: map[int]*store.ServerReplica{
						0: gsr(0, 50, []string{"sklearn"}, "server1", true, false),
						1: gsr(1, 200, []string{"sklearn"}, "server1", true, false), //expect schedule here - nop
						2: gsr(2, 175, []string{"sklearn"}, "server1", true, false), //expect schedule here - nop
						3: gsr(3, 100, []string{"sklearn"}, "server1", true, false),
					},
					Shared:           true,
					ExpectedReplicas: -1,
				},
			},
			scheduled:         true,
			scheduledServer:   "server1",
			scheduledReplicas: []int{1},
		},
		{
			name:  "Scale up - no capacity on loaded replica servers, should still go there",
			model: newTestModel("model1", 100, []string{"sklearn"}, nil, 3, []int{1, 2}, false, "server1", nil),
			servers: []*store.ServerSnapshot{
				{
					Name: "server1",
					Replicas: map[int]*store.ServerReplica{
						0: gsr(0, 50, []string{"sklearn"}, "server1", true, false),
						1: gsr(1, 0, []string{"sklearn"}, "server1", true, false),   //expect schedule here - nop
						2: gsr(2, 0, []string{"sklearn"}, "server1", true, false),   //expect schedule here - nop
						3: gsr(3, 100, []string{"sklearn"}, "server1", true, false), //expect schedule here
					},
					Shared:           true,
					ExpectedReplicas: -1,
				},
			},
			scheduled:         true,
			scheduledServer:   "server1",
			scheduledReplicas: []int{1, 2, 3},
		},
		{
			name:  "Scale down - no capacity on loaded replica servers, should still go there",
			model: newTestModel("model1", 100, []string{"sklearn"}, nil, 1, []int{1, 2}, false, "server1", nil),
			servers: []*store.ServerSnapshot{
				{
					Name: "server1",
					Replicas: map[int]*store.ServerReplica{
						0: gsr(0, 50, []string{"sklearn"}, "server1", true, false),
						1: gsr(1, 0, []string{"sklearn"}, "server1", true, false), //expect schedule here - nop
						2: gsr(2, 0, []string{"sklearn"}, "server1", true, false), //expect schedule here - nop
						3: gsr(3, 100, []string{"sklearn"}, "server1", true, false),
					},
					Shared:           true,
					ExpectedReplicas: -1,
				},
			},
			scheduled:         true,
			scheduledServer:   "server1",
			scheduledReplicas: []int{1},
		},
		{
			name:  "Drain",
			model: newTestModel("model1", 100, []string{"sklearn"}, nil, 2, []int{1}, false, "server1", []int{2}),
			servers: []*store.ServerSnapshot{
				{
					Name: "server1",
					Replicas: map[int]*store.ServerReplica{
						0: gsr(0, 50, []string{"sklearn"}, "server1", true, false),
						1: gsr(1, 200, []string{"sklearn"}, "server1", true, false), //expect schedule here - nop
						2: gsr(2, 175, []string{"sklearn"}, "server1", true, true),  //drain - should not be returned
						3: gsr(3, 100, []string{"sklearn"}, "server1", true, false), //expect schedule here new replica
					},
					Shared:           true,
					ExpectedReplicas: -1,
				},
			},
			scheduled:         true,
			scheduledServer:   "server1",
			scheduledReplicas: []int{1, 3},
		},
	}

	newMockStore := func(model *store.ModelSnapshot, servers []*store.ServerSnapshot) *mockStore {
		modelMap := make(map[string]*store.ModelSnapshot)
		modelMap[model.Name] = model
		return &mockStore{
			models:  modelMap,
			servers: servers,
		}
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockStore := newMockStore(test.model, test.servers)
			scheduler := NewSimpleScheduler(logger, mockStore, DefaultSchedulerConfig(mockStore))
			err := scheduler.Schedule(test.model.Name)
			if test.scheduled {
				g.Expect(err).To(BeNil())
				g.Expect(test.scheduledServer).To(Equal(mockStore.scheduledServer))
				sort.Ints(test.scheduledReplicas)
				sort.Ints(mockStore.scheduledReplicas)
				g.Expect(test.scheduledReplicas).To(Equal(mockStore.scheduledReplicas))
			} else {
				g.Expect(err).ToNot(BeNil())
			}
		})

	}

}

func TestFailedModels(t *testing.T) {
	logger := log.New()
	g := NewGomegaWithT(t)

	newMockStore := func(models map[string]store.ModelState) *mockStore {
		snapshots := map[string]*store.ModelSnapshot{}
		for name, state := range models {
			snapshot := &store.ModelSnapshot{
				Name:     name,
				Versions: []*store.ModelVersion{store.NewModelVersion(&pb.Model{}, 1, "", map[int]store.ReplicaStatus{}, false, state)},
			}
			snapshots[name] = snapshot
		}
		return &mockStore{
			models: snapshots,
		}
	}

	type test struct {
		name                 string
		models               map[string]store.ModelState
		expectedFailedModels []string
	}

	tests := []test{
		{
			name: "SmokeTest",
			models: map[string]store.ModelState{
				"model1": store.ScheduleFailed,
				"model2": store.ModelFailed,
				"model3": store.ModelAvailable,
			},
			expectedFailedModels: []string{"model1", "model2"},
		},
		{
			name: "SmokeTest",
			models: map[string]store.ModelState{
				"model3": store.ModelAvailable,
			},
			expectedFailedModels: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockStore := newMockStore(test.models)
			scheduler := NewSimpleScheduler(logger, mockStore, DefaultSchedulerConfig(mockStore))
			failedMoels, err := scheduler.getFailedModels()
			g.Expect(err).To(BeNil())
			sort.Strings(failedMoels)
			sort.Strings(test.expectedFailedModels)
			g.Expect(failedMoels).To(Equal(test.expectedFailedModels))
		})

	}

}
