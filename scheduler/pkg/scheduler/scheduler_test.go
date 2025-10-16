/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package scheduler

import (
	"errors"
	"maps"
	"slices"
	"sort"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gotidy/ptr"
	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/seldonio/seldon-core/apis/go/v2/mlops/agent"
	pb "github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/coordinator"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store/mock"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/synchroniser"
	mock2 "github.com/seldonio/seldon-core/scheduler/v2/pkg/synchroniser/mock"
)

type mockStore struct {
	models            map[string]*store.ModelSnapshot
	servers           []*store.ServerSnapshot
	scheduledServer   string
	scheduledReplicas []int
	unloadedModels    map[string]uint32
}

var _ store.ModelStore = (*mockStore)(nil)

func (f mockStore) FailedScheduling(modelID string, version uint32, reason string, reset bool) error {
	return nil
}

func (f mockStore) UnloadVersionModels(modelKey string, version uint32) (bool, error) {
	if f.unloadedModels != nil {
		f.unloadedModels[modelKey] = version
	}
	return true, nil
}

func (f mockStore) UnloadModelGwVersionModels(modelKey string, version uint32) (bool, error) {
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

func (f mockStore) SetModelGwModelState(name string, versionNumber uint32, status store.ModelState, reason string, source string) error {
	panic("implement me")
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
				g.Expect(atomic.LoadInt64(&serverEvents)).To(Equal(int64(test.expectedServerEvents)))
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

func TestScheduleFailedModels(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		setupMocks     func(*mock.MockModelStore, *mock2.MockSynchroniser)
		expectedModels []string
		expectError    bool
		errorContains  string
	}{
		{
			name: "success - schedules single failed model",
			setupMocks: func(ms *mock.MockModelStore, sync *mock2.MockSynchroniser) {
				sync.EXPECT().IsReady().Return(true)

				model1 := &store.ModelSnapshot{
					Name: "model1",
					Versions: []*store.ModelVersion{store.NewModelVersion(&pb.Model{
						Meta: &pb.MetaData{
							Name:           "model1",
							Kind:           nil,
							Version:        nil,
							KubernetesMeta: nil,
						},
						ModelSpec: &pb.ModelSpec{
							Uri:              "",
							ArtifactVersion:  nil,
							StorageConfig:    nil,
							Requirements:     nil,
							MemoryBytes:      nil,
							Server:           ptr.String("server1"),
							Parameters:       nil,
							ModelRuntimeInfo: nil,
							ModelSpec:        nil,
						},
						DeploymentSpec: &pb.DeploymentSpec{
							Replicas:    1,
							MinReplicas: 0,
							MaxReplicas: 0,
							LogPayloads: false,
						},
						StreamSpec:   nil,
						DataflowSpec: nil,
					}, 1, "server1", map[int]store.ReplicaStatus{}, false, store.ScheduleFailed)},
				}

				ms.EXPECT().GetModels().Return([]*store.ModelSnapshot{model1}, nil)

				ms.EXPECT().LockModel("model1")
				ms.EXPECT().UnlockModel("model1")
				ms.EXPECT().GetModel("model1").Return(model1, nil)

				servers := []*store.ServerSnapshot{
					createServerSnapshot("server1", 1, 16000),
				}
				ms.EXPECT().GetServers(false, true).Return(servers, nil)
				ms.EXPECT().UpdateLoadedModels("model1", uint32(1),
					"server1", slices.Collect(maps.Values(servers[0].Replicas))).Return(nil)
			},
			expectedModels: []string{"model1"},
			expectError:    false,
		},
		{
			name: "success - schedules 2 failed models",
			setupMocks: func(ms *mock.MockModelStore, sync *mock2.MockSynchroniser) {
				sync.EXPECT().IsReady().Return(true)

				model1 := &store.ModelSnapshot{
					Name: "model1",
					Versions: []*store.ModelVersion{store.NewModelVersion(&pb.Model{
						Meta: &pb.MetaData{
							Name:           "model1",
							Kind:           nil,
							Version:        nil,
							KubernetesMeta: nil,
						},
						ModelSpec: &pb.ModelSpec{
							Uri:              "",
							ArtifactVersion:  nil,
							StorageConfig:    nil,
							Requirements:     nil,
							MemoryBytes:      nil,
							Server:           ptr.String("server1"),
							Parameters:       nil,
							ModelRuntimeInfo: nil,
							ModelSpec:        nil,
						},
						DeploymentSpec: &pb.DeploymentSpec{
							Replicas:    1,
							MinReplicas: 0,
							MaxReplicas: 0,
							LogPayloads: false,
						},
						StreamSpec:   nil,
						DataflowSpec: nil,
					}, 1, "server1", map[int]store.ReplicaStatus{}, false, store.ScheduleFailed)},
				}

				model2 := &store.ModelSnapshot{
					Name: "model2",
					Versions: []*store.ModelVersion{store.NewModelVersion(&pb.Model{
						Meta: &pb.MetaData{
							Name:           "model2",
							Kind:           nil,
							Version:        nil,
							KubernetesMeta: nil,
						},
						ModelSpec: &pb.ModelSpec{
							Uri:              "",
							ArtifactVersion:  nil,
							StorageConfig:    nil,
							Requirements:     nil,
							MemoryBytes:      nil,
							Server:           ptr.String("server1"),
							Parameters:       nil,
							ModelRuntimeInfo: nil,
							ModelSpec:        nil,
						},
						DeploymentSpec: &pb.DeploymentSpec{
							Replicas:    1,
							MinReplicas: 0,
							MaxReplicas: 0,
							LogPayloads: false,
						},
						StreamSpec:   nil,
						DataflowSpec: nil,
					}, 1, "server1", map[int]store.ReplicaStatus{}, false, store.ScheduleFailed)},
				}

				ms.EXPECT().GetModels().Return([]*store.ModelSnapshot{model1, model2}, nil)

				// model1
				ms.EXPECT().LockModel("model1")
				ms.EXPECT().UnlockModel("model1")
				ms.EXPECT().GetModel("model1").Return(model1, nil)

				servers := []*store.ServerSnapshot{
					createServerSnapshot("server1", 1, 16000),
				}
				ms.EXPECT().GetServers(false, true).Return(servers, nil)
				ms.EXPECT().UpdateLoadedModels("model1", uint32(1),
					"server1", slices.Collect(maps.Values(servers[0].Replicas))).Return(nil)

				// model2

				ms.EXPECT().LockModel("model2")
				ms.EXPECT().UnlockModel("model2")
				ms.EXPECT().GetModel("model2").Return(model2, nil)

				ms.EXPECT().GetServers(false, true).Return(servers, nil)
				ms.EXPECT().UpdateLoadedModels("model2", uint32(1),
					"server1", slices.Collect(maps.Values(servers[0].Replicas))).Return(nil)
			},
			expectedModels: []string{"model1", "model2"},
			expectError:    false,
		},
		{
			name: "failure - unable to schedule model on desired replicas or min replicas",
			setupMocks: func(ms *mock.MockModelStore, sync *mock2.MockSynchroniser) {
				sync.EXPECT().IsReady().Return(true)

				model1 := &store.ModelSnapshot{
					Name: "model1",
					Versions: []*store.ModelVersion{store.NewModelVersion(&pb.Model{
						Meta: &pb.MetaData{
							Name:           "model1",
							Kind:           nil,
							Version:        nil,
							KubernetesMeta: nil,
						},
						ModelSpec: &pb.ModelSpec{
							Uri:              "",
							ArtifactVersion:  nil,
							StorageConfig:    nil,
							Requirements:     nil,
							MemoryBytes:      nil,
							Server:           ptr.String("server1"),
							Parameters:       nil,
							ModelRuntimeInfo: nil,
							ModelSpec:        nil,
						},
						DeploymentSpec: &pb.DeploymentSpec{
							Replicas:    3,
							MinReplicas: 2,
							MaxReplicas: 0,
							LogPayloads: false,
						},
						StreamSpec:   nil,
						DataflowSpec: nil,
					}, 1, "server1", map[int]store.ReplicaStatus{}, false, store.ScheduleFailed)},
				}

				ms.EXPECT().GetModels().Return([]*store.ModelSnapshot{model1}, nil)

				ms.EXPECT().LockModel("model1")
				ms.EXPECT().UnlockModel("model1")
				ms.EXPECT().GetModel("model1").Return(model1, nil)

				servers := []*store.ServerSnapshot{
					createServerSnapshot("server1", 1, 16000),
				}
				ms.EXPECT().GetServers(false, true).Return(servers, nil)
				ms.EXPECT().FailedScheduling("model1", uint32(1),
					"Failed to schedule model as no matching server had enough suitable replicas", true).Return(nil)
			},
			expectedModels: []string{},
		},
		{
			name: "failure - failed getting models",
			setupMocks: func(ms *mock.MockModelStore, sync *mock2.MockSynchroniser) {
				sync.EXPECT().IsReady().Return(true)
				ms.EXPECT().GetModels().Return(nil, errors.New("some error"))
			},
			expectedModels: []string{},
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			mockStore := mock.NewMockModelStore(ctrl)
			mockSync := mock2.NewMockSynchroniser(ctrl)

			tt.setupMocks(mockStore, mockSync)

			eventHub, err := coordinator.NewEventHub(log.New())
			require.NoError(t, err)

			scheduler := NewSimpleScheduler(
				log.New(),
				mockStore,
				DefaultSchedulerConfig(mockStore),
				mockSync,
				eventHub)

			updatedModels, err := scheduler.ScheduleFailedModels()

			if tt.expectError {
				require.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				return
			}

			require.NoError(t, err)
			assert.ElementsMatch(t, tt.expectedModels, updatedModels)
		})
	}
}

func createServerSnapshot(name string, numReplicas int, availableMemory uint64) *store.ServerSnapshot {
	replicas := make(map[int]*store.ServerReplica, numReplicas)
	server := store.NewServer(name, false)

	for i := 0; i < numReplicas; i++ {
		replicas[i] = store.NewServerReplica(name+"-svc",
			4000, 5000, i, server, nil,
			availableMemory, availableMemory, availableMemory, nil, 0)
	}

	return &store.ServerSnapshot{
		Name:             name,
		Replicas:         replicas,
		ExpectedReplicas: numReplicas,
	}
}
