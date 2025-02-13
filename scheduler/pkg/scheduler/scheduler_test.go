/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package scheduler

import (
	"sort"
	"sync/atomic"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"

	"github.com/seldonio/seldon-core/apis/go/v2/mlops/agent"
	pb "github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/coordinator"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/synchroniser"
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

func (f mockStore) ServerNotify(request *pb.ServerNotify) error {
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

func (f mockStore) UpdateModelState(modelKey string, version uint32, serverKey string, replicaIdx int, availableMemory *uint64, expectedState, desiredState store.ModelReplicaState, reason string, runtimeInfo *pb.ModelRuntimeInfo) error {
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

	newTestModel := func(name string, requiredMemory uint64, requirements []string, replicas, minReplicas uint32, maxReplicas uint32, loadedModels []int, deleted bool, scheduledServer string, drainedModels []int) *store.ModelSnapshot {
		config := &pb.Model{Meta: &pb.MetaData{Name: t.Name()}, ModelSpec: &pb.ModelSpec{MemoryBytes: &requiredMemory, Requirements: requirements}, DeploymentSpec: &pb.DeploymentSpec{Replicas: replicas, MinReplicas: minReplicas, MaxReplicas: maxReplicas}}
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
		name                 string
		model                *store.ModelSnapshot
		servers              []*store.ServerSnapshot
		scheduled            bool
		scheduledServer      string
		scheduledReplicas    []int
		expectedServerEvents int
	}

	tests := []test{
		{
			name:  "SmokeTest",
			model: newTestModel("model1", 100, []string{"sklearn"}, 1, 0, 1, []int{}, false, "", nil),
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
			model: newTestModel("model1", 100, []string{"sklearn"}, 2, 0, 2, []int{}, false, "", nil),
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
			model: newTestModel("model1", 100, []string{"sklearn"}, 2, 0, 2, []int{}, false, "", nil),
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
			name:  "NotEnoughReplicas - schedule min replicas",
			model: newTestModel("model1", 100, []string{"sklearn"}, 3, 2, 3, []int{}, false, "server2", nil),
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
			scheduled:            true, // not here that we still trying to mark the model as Available
			scheduledServer:      "server2",
			scheduledReplicas:    []int{0, 1},
			expectedServerEvents: 1,
		},
		{
			name:  "MemoryOneServer",
			model: newTestModel("model1", 100, []string{"sklearn"}, 1, 0, 1, []int{}, false, "", nil),
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
			model: newTestModel("model1", 100, []string{"sklearn"}, 2, 0, 2, []int{1}, false, "", nil),
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
			model: newTestModel("model1", 100, []string{"sklearn"}, 2, 0, 2, []int{1}, true, "server2", nil),
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
			model: newTestModel("model1", 100, []string{"sklearn"}, 1, 0, 1, []int{}, false, "", nil),
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
			model: newTestModel("model1", 100, []string{"sklearn"}, 1, 0, 1, []int{0}, false, "server1", nil),
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
			model: newTestModel("model1", 100, []string{"sklearn"}, 1, 0, 1, []int{1}, false, "", nil),
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
			model: newTestModel("model1", 100, []string{"sklearn"}, 1, 0, 1, []int{1}, false, "", nil),
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
			model: newTestModel("model1", 100, []string{"sklearn"}, 2, 0, 1, []int{1}, false, "", nil),
			servers: []*store.ServerSnapshot{
				{
					Name: "server2",
					Replicas: map[int]*store.ServerReplica{
						0: gsr(0, 150, []string{"sklearn"}, "server2", true, false),
						1: gsr(1, 200, []string{"sklearn"}, "server2", true, false), // expect schedule here
						2: gsr(2, 175, []string{"sklearn"}, "server2", true, false), // expect schedule here
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
			model: newTestModel("model1", 100, []string{"sklearn"}, 3, 0, 3, []int{1, 2}, false, "server1", nil),
			servers: []*store.ServerSnapshot{
				{
					Name: "server1",
					Replicas: map[int]*store.ServerReplica{
						0: gsr(0, 50, []string{"sklearn"}, "server1", true, false),
						1: gsr(1, 200, []string{"sklearn"}, "server1", true, false), // expect schedule here - nop
						2: gsr(2, 175, []string{"sklearn"}, "server1", true, false), // expect schedule here - nop
						3: gsr(3, 100, []string{"sklearn"}, "server1", true, false), // expect schedule here
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
			model: newTestModel("model1", 100, []string{"sklearn"}, 1, 0, 1, []int{1, 2}, false, "server1", nil),
			servers: []*store.ServerSnapshot{
				{
					Name: "server1",
					Replicas: map[int]*store.ServerReplica{
						0: gsr(0, 50, []string{"sklearn"}, "server1", true, false),
						1: gsr(1, 200, []string{"sklearn"}, "server1", true, false), // expect schedule here - nop
						2: gsr(2, 175, []string{"sklearn"}, "server1", true, false), // expect schedule here - nop
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
			name:  "Scale up - not enough replicas use max of the server",
			model: newTestModel("model1", 100, []string{"sklearn"}, 5, 3, 5, []int{1, 2}, false, "server1", nil),
			servers: []*store.ServerSnapshot{
				{
					Name: "server1",
					Replicas: map[int]*store.ServerReplica{
						0: gsr(0, 100, []string{"sklearn"}, "server1", true, false), // expect schedule here
						1: gsr(1, 100, []string{"sklearn"}, "server1", true, false), // expect schedule here - nop
						2: gsr(2, 100, []string{"sklearn"}, "server1", true, false), // expect schedule here - nop
						3: gsr(3, 100, []string{"sklearn"}, "server1", true, false), // expect schedule here
					},
					Shared:           true,
					ExpectedReplicas: -1,
				},
			},
			scheduled:            true, // note that we are still trying to make the model as Available
			scheduledServer:      "server1",
			scheduledReplicas:    []int{0, 1, 2, 3}, // used all replicas
			expectedServerEvents: 1,
		},
		{
			name:  "Scale up - no capacity on loaded replica servers, should still go there",
			model: newTestModel("model1", 100, []string{"sklearn"}, 3, 0, 3, []int{1, 2}, false, "server1", nil),
			servers: []*store.ServerSnapshot{
				{
					Name: "server1",
					Replicas: map[int]*store.ServerReplica{
						0: gsr(0, 50, []string{"sklearn"}, "server1", true, false),
						1: gsr(1, 0, []string{"sklearn"}, "server1", true, false),   // expect schedule here - nop
						2: gsr(2, 0, []string{"sklearn"}, "server1", true, false),   // expect schedule here - nop
						3: gsr(3, 100, []string{"sklearn"}, "server1", true, false), // expect schedule here
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
			model: newTestModel("model1", 100, []string{"sklearn"}, 1, 0, 1, []int{1, 2}, false, "server1", nil),
			servers: []*store.ServerSnapshot{
				{
					Name: "server1",
					Replicas: map[int]*store.ServerReplica{
						0: gsr(0, 50, []string{"sklearn"}, "server1", true, false),
						1: gsr(1, 0, []string{"sklearn"}, "server1", true, false), // expect schedule here - nop
						2: gsr(2, 0, []string{"sklearn"}, "server1", true, false), // expect schedule here - nop
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
			model: newTestModel("model1", 100, []string{"sklearn"}, 2, 0, 2, []int{1}, false, "server1", []int{2}),
			servers: []*store.ServerSnapshot{
				{
					Name: "server1",
					Replicas: map[int]*store.ServerReplica{
						0: gsr(0, 50, []string{"sklearn"}, "server1", true, false),
						1: gsr(1, 200, []string{"sklearn"}, "server1", true, false), // expect schedule here - nop
						2: gsr(2, 175, []string{"sklearn"}, "server1", true, true),  // drain - should not be returned
						3: gsr(3, 100, []string{"sklearn"}, "server1", true, false), // expect schedule here new replica
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
			eventHub, _ := coordinator.NewEventHub(logger)

			serverEvents := int64(0)
			eventHub.RegisterServerEventHandler(
				"handler-server",
				10,
				logger,
				func(event coordinator.ServerEventMsg) { atomic.AddInt64(&serverEvents, 1) },
			)

			mockStore := newMockStore(test.model, test.servers)
			scheduler := NewSimpleScheduler(logger, mockStore, DefaultSchedulerConfig(mockStore), synchroniser.NewSimpleSynchroniser(time.Duration(10*time.Millisecond)), eventHub)
			err := scheduler.Schedule(test.model.Name)
			if test.scheduled {
				g.Expect(err).To(BeNil())
			} else {
				g.Expect(err).ToNot(BeNil())
			}
			if test.expectedServerEvents > 0 { // wait for event
				time.Sleep(500 * time.Millisecond)
			}
			if test.scheduledServer != "" {
				g.Expect(test.scheduledServer).To(Equal(mockStore.scheduledServer))
				sort.Ints(test.scheduledReplicas)
				sort.Ints(mockStore.scheduledReplicas)
				g.Expect(test.scheduledReplicas).To(Equal(mockStore.scheduledReplicas))
				g.Expect(serverEvents).To(Equal(int64(test.expectedServerEvents)))
			}
		})
	}
}

func TestFailedModels(t *testing.T) {
	logger := log.New()
	g := NewGomegaWithT(t)

	type modelStateWithMetadata struct {
		state             store.ModelState
		deploymentSpec    *pb.DeploymentSpec
		availableReplicas uint32
	}

	newMockStore := func(models map[string]modelStateWithMetadata) *mockStore {
		snapshots := map[string]*store.ModelSnapshot{}
		for name, state := range models {
			mv := store.NewModelVersion(&pb.Model{DeploymentSpec: state.deploymentSpec}, 1, "", map[int]store.ReplicaStatus{}, false, state.state)
			mv.SetModelState(store.ModelStatus{
				State:             state.state,
				AvailableReplicas: state.availableReplicas,
			})
			snapshot := &store.ModelSnapshot{
				Name:     name,
				Versions: []*store.ModelVersion{mv},
			}
			snapshots[name] = snapshot
		}
		return &mockStore{
			models: snapshots,
		}
	}

	type test struct {
		name                 string
		models               map[string]modelStateWithMetadata
		expectedFailedModels []string
	}

	tests := []test{
		{
			name: "SmokeTest",
			models: map[string]modelStateWithMetadata{
				"model1": {store.ScheduleFailed, &pb.DeploymentSpec{Replicas: 1}, 0},
				"model2": {store.ModelFailed, &pb.DeploymentSpec{Replicas: 1}, 0},
				"model3": {store.ModelAvailable, &pb.DeploymentSpec{Replicas: 1}, 1},
				"model4": {store.ModelAvailable, &pb.DeploymentSpec{Replicas: 2, MinReplicas: 1, MaxReplicas: 2}, 1}, // retry models that have not reached desired replicas
			},
			expectedFailedModels: []string{"model1", "model2", "model4"},
		},
		{
			name: "SmokeTest",
			models: map[string]modelStateWithMetadata{
				"model3": {store.ModelAvailable, &pb.DeploymentSpec{Replicas: 1}, 1},
			},
			expectedFailedModels: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			eventHub, _ := coordinator.NewEventHub(logger)

			mockStore := newMockStore(test.models)
			scheduler := NewSimpleScheduler(logger, mockStore, DefaultSchedulerConfig(mockStore), synchroniser.NewSimpleSynchroniser(time.Duration(10*time.Millisecond)), eventHub)
			failedModels, err := scheduler.getFailedModels()
			g.Expect(err).To(BeNil())
			sort.Strings(failedModels)
			sort.Strings(test.expectedFailedModels)
			g.Expect(failedModels).To(Equal(test.expectedFailedModels))
		})
	}
}
