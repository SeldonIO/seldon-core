package scheduler

import (
	"sort"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/seldonio/seldon-core/scheduler/apis/mlops/agent"
	pb "github.com/seldonio/seldon-core/scheduler/apis/mlops/scheduler"
	"github.com/seldonio/seldon-core/scheduler/pkg/store"
	log "github.com/sirupsen/logrus"
)

type mockStore struct {
	models            map[string]*store.ModelSnapshot
	servers           []*store.ServerSnapshot
	scheduledServer   string
	scheduledReplicas []int
}

func (f mockStore) FailedScheduling(modelVersion *store.ModelVersion, reason string) {
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

func (f mockStore) ExistsModelVersion(key string, version uint32) bool {
	return false
}

func (f mockStore) GetServers() ([]*store.ServerSnapshot, error) {
	return f.servers, nil
}

func (f mockStore) GetServer(serverKey string) (*store.ServerSnapshot, error) {
	panic("implement me")
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

func (f mockStore) UpdateModelState(modelKey string, version uint32, serverKey string, replicaIdx int, availableMemory *uint64, state store.ModelReplicaState, reason string) error {
	panic("implement me")
}

func (f mockStore) AddServerReplica(request *agent.AgentSubscribeRequest) error {
	panic("implement me")
}

func (f mockStore) RemoveServerReplica(serverName string, replicaIdx int) ([]string, error) {
	panic("implement me")
}

func (f mockStore) AddModelEventListener(c chan *store.ModelSnapshot) {

}

func (f mockStore) AddServerEventListener(c chan string) {

}

func TestScheduler(t *testing.T) {
	logger := log.New()
	g := NewGomegaWithT(t)

	newTestModel := func(name string, requiredMemory uint64, requirements []string, server *string, replicas uint32, loadedModels []int, deleted bool, scheduledServer string) *store.ModelSnapshot {
		config := &pb.Model{ModelSpec: &pb.ModelSpec{MemoryBytes: &requiredMemory, Requirements: requirements, Server: server}, DeploymentSpec: &pb.DeploymentSpec{Replicas: replicas}}
		rmap := make(map[int]store.ReplicaStatus)
		for _, ridx := range loadedModels {
			rmap[ridx] = store.ReplicaStatus{State: store.Loaded}
		}
		return &store.ModelSnapshot{
			Name:     name,
			Versions: []*store.ModelVersion{store.NewModelVersion(config, 1, scheduledServer, rmap, false, store.ModelProgressing)},
			Deleted:  deleted,
		}
	}

	gsr := func(replicaIdx int, availableMemory uint64, capabilities []string) *store.ServerReplica {
		return store.NewServerReplica("svc", 8080, 5001, replicaIdx, nil, capabilities, availableMemory, availableMemory, nil, true)
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
			model: newTestModel("model1", 100, []string{"sklearn"}, nil, 1, []int{}, false, ""),
			servers: []*store.ServerSnapshot{
				{
					Name:             "server1",
					Replicas:         map[int]*store.ServerReplica{0: gsr(0, 200, []string{"sklearn"})},
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
			model: newTestModel("model1", 100, []string{"sklearn"}, nil, 2, []int{}, false, ""),
			servers: []*store.ServerSnapshot{
				{
					Name:             "server1",
					Replicas:         map[int]*store.ServerReplica{0: gsr(0, 200, []string{"sklearn"})},
					Shared:           true,
					ExpectedReplicas: -1,
				},
				{
					Name: "server2",
					Replicas: map[int]*store.ServerReplica{
						0: gsr(0, 200, []string{"sklearn"}),
						1: gsr(1, 200, []string{"sklearn"}),
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
			model: newTestModel("model1", 100, []string{"sklearn"}, nil, 2, []int{}, false, ""),
			servers: []*store.ServerSnapshot{
				{
					Name:             "server1",
					Replicas:         map[int]*store.ServerReplica{0: gsr(0, 200, []string{"sklearn"})},
					Shared:           true,
					ExpectedReplicas: -1,
				},
				{
					Name: "server2",
					Replicas: map[int]*store.ServerReplica{
						0: gsr(0, 200, []string{"sklearn"}),
						1: gsr(1, 200, []string{"foo"}),
					},
					Shared:           true,
					ExpectedReplicas: -1,
				},
			},
			scheduled: false,
		},
		{
			name:  "MemoryOneServer",
			model: newTestModel("model1", 100, []string{"sklearn"}, nil, 1, []int{}, false, ""),
			servers: []*store.ServerSnapshot{
				{
					Name:             "server1",
					Replicas:         map[int]*store.ServerReplica{0: gsr(0, 50, []string{"sklearn"})},
					Shared:           true,
					ExpectedReplicas: -1,
				},
				{
					Name: "server2",
					Replicas: map[int]*store.ServerReplica{
						0: gsr(0, 200, []string{"sklearn"}),
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
			model: newTestModel("model1", 100, []string{"sklearn"}, nil, 2, []int{1}, false, ""),
			servers: []*store.ServerSnapshot{
				{
					Name:             "server1",
					Replicas:         map[int]*store.ServerReplica{0: gsr(0, 50, []string{"sklearn"})},
					Shared:           true,
					ExpectedReplicas: -1,
				},
				{
					Name: "server2",
					Replicas: map[int]*store.ServerReplica{
						0: gsr(0, 200, []string{"sklearn"}),
						1: gsr(1, 200, []string{"sklearn"}),
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
			model: newTestModel("model1", 100, []string{"sklearn"}, nil, 2, []int{1}, true, "server2"),
			servers: []*store.ServerSnapshot{
				{
					Name: "server2",
					Replicas: map[int]*store.ServerReplica{
						0: gsr(0, 200, []string{"sklearn"}),
						1: gsr(1, 200, []string{"sklearn"}),
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
			model: newTestModel("model1", 100, []string{"sklearn"}, nil, 1, []int{}, false, ""),
			servers: []*store.ServerSnapshot{
				{
					Name:             "server1",
					Replicas:         map[int]*store.ServerReplica{0: gsr(0, 200, []string{"sklearn"})},
					Shared:           true,
					ExpectedReplicas: 0,
				},
				{
					Name: "server2",
					Replicas: map[int]*store.ServerReplica{
						0: gsr(0, 200, []string{"sklearn"}),
						1: gsr(1, 200, []string{"sklearn"}),
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
			model: newTestModel("model1", 100, []string{"sklearn"}, nil, 1, []int{0}, false, "server1"),
			servers: []*store.ServerSnapshot{
				{
					Name:             "server1",
					Replicas:         map[int]*store.ServerReplica{0: gsr(0, 200, []string{"sklearn"})},
					Shared:           true,
					ExpectedReplicas: 0,
				},
				{
					Name: "server2",
					Replicas: map[int]*store.ServerReplica{
						0: gsr(0, 200, []string{"sklearn"}),
						1: gsr(1, 200, []string{"sklearn"}),
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
			model: newTestModel("model1", 100, []string{"sklearn"}, nil, 1, []int{1}, false, ""),
			servers: []*store.ServerSnapshot{
				{
					Name:             "server1",
					Replicas:         map[int]*store.ServerReplica{0: gsr(0, 200, []string{"sklearn"})},
					Shared:           true,
					ExpectedReplicas: 0,
				},
			},
			scheduled: false,
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
			scheduler := NewSimpleScheduler(logger, mockStore, DefaultSchedulerConfig())
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
