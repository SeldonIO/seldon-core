/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package scheduler

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"

	"github.com/seldonio/seldon-core/apis/go/v2/mlops/agent"
	pb "github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/coordinator"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/scheduler/filters"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/synchroniser"
)

type mockStore struct {
	models            map[string]*store.ModelSnapshot
	servers           []*store.ServerSnapshot
	scheduledServer   string
	scheduledReplicas []int
	unloadedModels    map[string]uint32
}

var _ store.ModelStore = (*mockStore)(nil)

func (f mockStore) FailedScheduling(modelVersion *store.ModelVersion, reason string, reset bool) {
}

func (f mockStore) UnloadVersionModels(modelKey string, version uint32) (bool, error) {
	if f.unloadedModels != nil {
		f.unloadedModels[modelKey] = version
	}
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

func TestRemoveAllVersions(t *testing.T) {
	logger := log.New()
	g := NewGomegaWithT(t)

	newTestModel := func(name string, requirements []string, loadedModels []int, scheduledServer string, numVersions int) *store.ModelSnapshot {
		config := &pb.Model{Meta: &pb.MetaData{Name: t.Name()}, ModelSpec: &pb.ModelSpec{Requirements: requirements}}
		rmap := make(map[int]store.ReplicaStatus)
		for _, ridx := range loadedModels {
			rmap[ridx] = store.ReplicaStatus{State: store.Loaded}
		}

		versions := []*store.ModelVersion{}
		for i := 1; i <= numVersions; i++ {
			versions = append(versions, store.NewModelVersion(config, uint32(i), scheduledServer, rmap, false, store.ModelAvailable))
		}
		// load a bad version - this should not get unloaded by the test
		versions = append(versions, store.NewModelVersion(config, uint32(numVersions+1), scheduledServer, map[int]store.ReplicaStatus{}, false, store.ScheduleFailed))

		return &store.ModelSnapshot{
			Name:     name,
			Versions: versions,
			Deleted:  true,
		}
	}

	gsr := func(replicaIdx int, availableMemory uint64, capabilities []string, serverName string) *store.ServerReplica {
		replica := store.NewServerReplica("svc", 8080, 5001, replicaIdx, store.NewServer(serverName, true), capabilities, availableMemory, availableMemory, 0, nil, 100)
		return replica
	}

	newMockStore := func(model *store.ModelSnapshot, servers []*store.ServerSnapshot) *mockStore {
		modelMap := make(map[string]*store.ModelSnapshot)
		modelMap[model.Name] = model
		return &mockStore{
			models:         modelMap,
			servers:        servers,
			unloadedModels: make(map[string]uint32),
		}
	}

	type test struct {
		name        string
		model       *store.ModelSnapshot
		servers     []*store.ServerSnapshot
		numVersions int
	}

	tests := []test{
		{
			name:  "Allversions - 1",
			model: newTestModel("model1", []string{"sklearn"}, []int{0, 1}, "server", 1),
			servers: []*store.ServerSnapshot{
				{
					Name: "server2",
					Replicas: map[int]*store.ServerReplica{
						0: gsr(0, 200, []string{"sklearn"}, "server"),
						1: gsr(1, 200, []string{"sklearn"}, "server"),
					},
					Shared:           true,
					ExpectedReplicas: -1,
				},
			},
			numVersions: 1,
		},
		{
			name:  "Allversions - > 1",
			model: newTestModel("model1", []string{"sklearn"}, []int{0, 1}, "server", 10),
			servers: []*store.ServerSnapshot{
				{
					Name: "server",
					Replicas: map[int]*store.ServerReplica{
						0: gsr(0, 200, []string{"sklearn"}, "server"),
						1: gsr(1, 200, []string{"sklearn"}, "server"),
					},
					Shared:           true,
					ExpectedReplicas: -1,
				},
			},
			numVersions: 10,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, _ = coordinator.NewEventHub(logger)
			mockStore := newMockStore(test.model, test.servers)
			scheduler := NewSimpleScheduler(logger, mockStore, DefaultSchedulerConfig(mockStore), synchroniser.NewSimpleSynchroniser(time.Duration(10*time.Millisecond)), nil)
			err := scheduler.Schedule(test.model.Name)
			g.Expect(err).To(BeNil())

			g.Expect(mockStore.unloadedModels[test.model.Name]).To(Equal(uint32(test.numVersions)))
		})
	}
}

// Test the getModelRequirementsStr helper function
func TestGetModelRequirementsStr(t *testing.T) {
	tests := []struct {
		name     string
		model    *store.ModelVersion
		expected string
	}{
		{
			name: "model with all requirements",
			model: store.NewDefaultModelVersion(&pb.Model{
				Meta: &pb.MetaData{Name: "test-model"},
				ModelSpec: &pb.ModelSpec{
					Server:       proto.String("mlserver"),
					MemoryBytes:  proto.Uint64(1024 * 1024 * 512), // 512MB
					Requirements: []string{"pytorch", "numpy"},
				},
				DeploymentSpec: &pb.DeploymentSpec{
					Replicas:    3,
					MinReplicas: 1,
				},
			}, 1),
			expected: "replicas=3, min_replicas=1, memory=512MB, server=mlserver, capabilities=[pytorch numpy]",
		},
		{
			name: "model with no requirements",
			model: store.NewDefaultModelVersion(&pb.Model{
				Meta: &pb.MetaData{Name: "test-model"},
				ModelSpec:      &pb.ModelSpec{},
				DeploymentSpec: &pb.DeploymentSpec{},
			}, 1),
			expected: "none specified",
		},
		{
			name: "model with only memory requirement",
			model: store.NewDefaultModelVersion(&pb.Model{
				Meta: &pb.MetaData{Name: "test-model"},
				ModelSpec: &pb.ModelSpec{
					MemoryBytes: proto.Uint64(1024 * 1024 * 1024), // 1GB
				},
				DeploymentSpec: &pb.DeploymentSpec{
					Replicas: 1,
				},
			}, 1),
			expected: "replicas=1, memory=1024MB",
		},
		{
			name: "model with server and capabilities only",
			model: store.NewDefaultModelVersion(&pb.Model{
				Meta: &pb.MetaData{Name: "test-model"},
				ModelSpec: &pb.ModelSpec{
					Server:       proto.String("triton"),
					Requirements: []string{"tensorflow"},
				},
				DeploymentSpec: &pb.DeploymentSpec{},
			}, 1),
			expected: "server=triton, capabilities=[tensorflow]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getModelRequirementsStr(tt.model)
			if result != tt.expected {
				t.Errorf("getModelRequirementsStr() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

// Test the storeRejectionReasons helper function
func TestStoreRejectionReasons(t *testing.T) {
	// Create a test logger that captures log entries
	logBuffer := &bytes.Buffer{}
	logger := log.New()
	logger.SetOutput(logBuffer)
	logger.SetLevel(log.InfoLevel)

	scheduler := &SimpleScheduler{
		logger: logger,
	}

	model := store.NewDefaultModelVersion(&pb.Model{
		Meta: &pb.MetaData{
			Name: "test-model",
		},
	}, 1)

	reasons := []string{
		"Server 'server1' rejected by MemoryFilter: insufficient memory",
		"Server 'server2' rejected by CapabilityFilter: missing pytorch",
	}

	scheduler.storeRejectionReasons(model, reasons)

	logOutput := logBuffer.String()
	
	// Verify the log contains the model name
	if !strings.Contains(logOutput, "test-model") {
		t.Error("Log output should contain model name")
	}

	// Verify the log contains rejection reasons
	if !strings.Contains(logOutput, "insufficient memory") {
		t.Error("Log output should contain detailed rejection reasons")
	}
	
	// Verify the log contains the joined reasons
	expectedReasons := strings.Join(reasons, "; ")
	if !strings.Contains(logOutput, expectedReasons) {
		t.Error("Log output should contain joined rejection reasons")
	}
}

// Mock implementations for testing verbose logging

type mockVerboseMemoryStore struct {
	servers       []*store.ServerSnapshot
	modelVersions map[string]*store.ModelVersion
	failedModels  map[string]string // modelName -> reason
}

func (m *mockVerboseMemoryStore) GetServers(includeScalingDownServers bool, includeScalingUpServers bool) ([]*store.ServerSnapshot, error) {
	return m.servers, nil
}

func (m *mockVerboseMemoryStore) GetModelVersion(name string) (*store.ModelVersion, error) {
	if model, exists := m.modelVersions[name]; exists {
		return model, nil
	}
	return nil, fmt.Errorf("model not found")
}

func (m *mockVerboseMemoryStore) FailedScheduling(modelVersion *store.ModelVersion, reason string, reset bool) {
	if m.failedModels == nil {
		m.failedModels = make(map[string]string)
	}
	m.failedModels[modelVersion.GetModel().GetMeta().GetName()] = reason
}

// Implement other required methods as no-ops for testing
func (m *mockVerboseMemoryStore) GetModel(key string) (*store.ModelSnapshot, error) {
	if version, exists := m.modelVersions[key]; exists {
		// Create a minimal ModelSnapshot with the version
		return &store.ModelSnapshot{
			Name: key,
			Versions: []*store.ModelVersion{version},
		}, nil
	}
	return nil, fmt.Errorf("Unable to find model")
}
func (m *mockVerboseMemoryStore) GetModels() ([]*store.ModelSnapshot, error) { return nil, nil }
func (m *mockVerboseMemoryStore) LockModel(modelId string) {}
func (m *mockVerboseMemoryStore) UnlockModel(modelId string) {}
func (m *mockVerboseMemoryStore) ExistsModelVersion(key string, version uint32) bool { return false }
func (m *mockVerboseMemoryStore) GetServer(serverKey string, shallow bool, modelDetails bool) (*store.ServerSnapshot, error) { return nil, nil }
func (m *mockVerboseMemoryStore) GetAllModels() []string { return nil }
func (m *mockVerboseMemoryStore) UpdateLoadedModels(modelKey string, version uint32, serverKey string, replicas []*store.ServerReplica) error { return nil }
func (m *mockVerboseMemoryStore) UnloadVersionModels(modelKey string, version uint32) (bool, error) { return true, nil }
func (m *mockVerboseMemoryStore) ServerNotify(request *pb.ServerNotify) error { return nil }
func (m *mockVerboseMemoryStore) RemoveModel(req *pb.UnloadModelRequest) error { return nil }
func (m *mockVerboseMemoryStore) UpdateModel(config *pb.LoadModelRequest) error { return nil }
func (m *mockVerboseMemoryStore) AddServerReplica(request *agent.AgentSubscribeRequest) error { return nil }
func (m *mockVerboseMemoryStore) DrainServerReplica(serverKey string, replicaIdx int) ([]string, error) { return nil, nil }
func (m *mockVerboseMemoryStore) RemoveServerReplica(serverKey string, replicaIdx int) ([]string, error) { return nil, nil }
func (m *mockVerboseMemoryStore) UpdateModelState(modelKey string, version uint32, serverKey string, replicaIdx int, availableMemory *uint64, expectedState, desiredState store.ModelReplicaState, reason string, runtimeInfo *pb.ModelRuntimeInfo) error { return nil }

type mockServerFilter struct {
	name        string
	shouldPass  bool
	description string
}

func (f *mockServerFilter) Filter(model *store.ModelVersion, server *store.ServerSnapshot) bool {
	return f.shouldPass
}

func (f *mockServerFilter) Name() string {
	return f.name
}

func (f *mockServerFilter) Description(model *store.ModelVersion, server *store.ServerSnapshot) string {
	return f.description
}

type mockReplicaFilter struct {
	name        string
	shouldPass  bool
	description string
}

func (f *mockReplicaFilter) Filter(model *store.ModelVersion, replica *store.ServerReplica) bool {
	return f.shouldPass
}

func (f *mockReplicaFilter) Name() string {
	return f.name
}

func (f *mockReplicaFilter) Description(model *store.ModelVersion, replica *store.ServerReplica) string {
	return f.description
}

// Test enhanced error messages in scheduleToServer
func TestScheduleToServerEnhancedErrorMessages(t *testing.T) {
	// Create a test logger that captures log entries
	logBuffer := &bytes.Buffer{}
	logger := log.New()
	logger.SetOutput(logBuffer)
	logger.SetLevel(log.InfoLevel)

	// Create a mock store that returns empty servers
	mockStore := &mockVerboseMemoryStore{
		servers: []*store.ServerSnapshot{}, // No servers available
	}

	scheduler := &SimpleScheduler{
		logger: logger,
		store:  mockStore,
		SchedulerConfig: SchedulerConfig{
			serverFilters: []filters.ServerFilter{},
		},
	}

	// Test model
	modelName := "test-model"
	mockStore.modelVersions = map[string]*store.ModelVersion{
		modelName: store.NewDefaultModelVersion(&pb.Model{
			Meta: &pb.MetaData{
				Name: modelName,
			},
			ModelSpec: &pb.ModelSpec{
				Server:       proto.String("mlserver"),
				MemoryBytes:  proto.Uint64(1024 * 1024 * 256), // 256MB
				Requirements: []string{"pytorch"},
			},
			DeploymentSpec: &pb.DeploymentSpec{
				Replicas:    2,
				MinReplicas: 1,
			},
		}, 1),
	}

	// Call scheduleToServer which should fail with enhanced error message
	_, err := scheduler.scheduleToServer(modelName)

	// Verify error is returned
	if err == nil {
		t.Error("Expected error when no servers available")
	}

	logOutput := logBuffer.String()

	// Verify enhanced error message components
	if !strings.Contains(logOutput, "test-model") {
		t.Error("Log should contain model name")
	}

	if !strings.Contains(logOutput, "checked 0 servers") {
		t.Error("Log should contain server count")
	}

	if !strings.Contains(logOutput, "model_requirements") {
		t.Error("Log should contain model requirements field")
	}

	if !strings.Contains(logOutput, "mlserver") {
		t.Error("Log should contain model server requirement")
	}

	// Check that the reason was stored properly
	if reason, exists := mockStore.failedModels[modelName]; !exists || !strings.Contains(reason, "test-model") {
		t.Error("Enhanced error message should be stored in failed scheduling")
	}
}

// Test server filter rejection reason collection
func TestFilterServersRejectionReasons(t *testing.T) {
	// Create a test logger that captures log entries
	logBuffer := &bytes.Buffer{}
	logger := log.New()
	logger.SetOutput(logBuffer)
	logger.SetLevel(log.InfoLevel)

	// Create a mock filter that always rejects
	mockFilter := &mockServerFilter{
		name:        "TestFilter",
		shouldPass:  false,
		description: "Test rejection reason",
	}

	scheduler := &SimpleScheduler{
		logger: logger,
		SchedulerConfig: SchedulerConfig{
			serverFilters: []filters.ServerFilter{mockFilter},
		},
	}

	// Test data
	model := store.NewDefaultModelVersion(&pb.Model{
		Meta: &pb.MetaData{
			Name: "test-model",
		},
	}, 1)

	servers := []*store.ServerSnapshot{
		{Name: "server1"},
		{Name: "server2"},
	}

	// Call filterServers
	result := scheduler.filterServers(model, servers)

	// Verify no servers pass the filter
	if len(result) != 0 {
		t.Error("Expected no servers to pass the filter")
	}

	logOutput := logBuffer.String()

	// Verify rejection reasons are logged
	if !strings.Contains(logOutput, "Server 'server1' rejected by TestFilter: Test rejection reason") {
		t.Error("Log should contain detailed rejection reason for server1")
	}

	if !strings.Contains(logOutput, "Server 'server2' rejected by TestFilter: Test rejection reason") {
		t.Error("Log should contain detailed rejection reason for server2")
	}

	if !strings.Contains(logOutput, "All servers rejected for model scheduling") {
		t.Error("Log should contain summary message about all servers being rejected")
	}
}

// Test replica filter rejection reason collection
func TestFilterReplicasRejectionReasons(t *testing.T) {
	// Create a test logger that captures log entries
	logBuffer := &bytes.Buffer{}
	logger := log.New()
	logger.SetOutput(logBuffer)
	logger.SetLevel(log.InfoLevel)

	// Create a mock filter that always rejects
	mockFilter := &mockReplicaFilter{
		name:        "TestReplicaFilter",
		shouldPass:  false,
		description: "Test replica rejection reason",
	}

	scheduler := &SimpleScheduler{
		logger: logger,
		SchedulerConfig: SchedulerConfig{
			replicaFilters: []filters.ReplicaFilter{mockFilter},
		},
	}

	// Test data
	model := store.NewDefaultModelVersion(&pb.Model{
		Meta: &pb.MetaData{
			Name: "test-model",
		},
	}, 1)

	server := &store.ServerSnapshot{
		Name: "test-server",
		Replicas: map[int]*store.ServerReplica{
			0: store.NewServerReplica("test-inference-svc", 8000, 9000, 0, store.NewServer("test-server", true), []string{}, 1024, 1024, 0, map[store.ModelVersionID]bool{}, 0),
			1: store.NewServerReplica("test-inference-svc", 8000, 9000, 1, store.NewServer("test-server", true), []string{}, 1024, 1024, 0, map[store.ModelVersionID]bool{}, 0),
		},
	}

	// Call filterReplicas
	result := scheduler.filterReplicas(model, server)

	// Verify no replicas pass the filter
	if len(result.ChosenReplicas) != 0 {
		t.Error("Expected no replicas to pass the filter")
	}

	logOutput := logBuffer.String()

	// Verify replica rejection reasons are logged
	if !strings.Contains(logOutput, "Replica 0 rejected by TestReplicaFilter: Test replica rejection reason") {
		t.Error("Log should contain detailed rejection reason for replica 0")
	}

	if !strings.Contains(logOutput, "Replica 1 rejected by TestReplicaFilter: Test replica rejection reason") {
		t.Error("Log should contain detailed rejection reason for replica 1")
	}

	if !strings.Contains(logOutput, "No suitable replicas found on server for model") {
		t.Error("Log should contain summary message about no suitable replicas")
	}
}
