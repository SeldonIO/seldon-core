package agent

import (
	"testing"

	. "github.com/onsi/gomega"
	"github.com/seldonio/seldon-core/scheduler/pkg/coordinator"
	"google.golang.org/grpc"

	pb "github.com/seldonio/seldon-core/scheduler/apis/mlops/agent"
	pbs "github.com/seldonio/seldon-core/scheduler/apis/mlops/scheduler"
	"github.com/seldonio/seldon-core/scheduler/pkg/store"
	log "github.com/sirupsen/logrus"
)

type mockStore struct {
	models map[string]*store.ModelSnapshot
}

func (m *mockStore) FailedScheduling(modelVersion *store.ModelVersion, reason string) {
}

func (m *mockStore) UpdateModel(config *pbs.LoadModelRequest) error {
	panic("implement me")
}

func (m *mockStore) GetModel(key string) (*store.ModelSnapshot, error) {
	return m.models[key], nil
}

func (m mockStore) LockModel(modelId string) {
}

func (m mockStore) UnlockModel(modelId string) {
}

func (m *mockStore) RemoveModel(req *pbs.UnloadModelRequest) error {
	panic("implement me")
}

func (m *mockStore) GetServers() ([]*store.ServerSnapshot, error) {
	panic("implement me")
}

func (m *mockStore) GetServer(serverKey string) (*store.ServerSnapshot, error) {
	panic("implement me")
}

func (m *mockStore) AddNewModelVersion(modelName string) error {
	panic("implement me")
}

func (m *mockStore) UpdateLoadedModels(modelKey string, version uint32, serverKey string, replicas []*store.ServerReplica) error {
	panic("implement me")
}

func (m *mockStore) UnloadVersionModels(modelKey string, version uint32) (bool, error) {
	panic("implement me")
}

func (m *mockStore) UpdateModelState(modelKey string, version uint32, serverKey string, replicaIdx int, availableMemory *uint64, expectedState, desiredState store.ModelReplicaState, reason string) error {
	model := m.models[modelKey]
	for _, mv := range model.Versions {
		if mv.GetVersion() == version {
			mv.SetReplicaState(replicaIdx, store.ReplicaStatus{State: desiredState, Reason: reason})
		}
	}
	return nil
}

func (m *mockStore) AddServerReplica(request *pb.AgentSubscribeRequest) error {
	panic("implement me")
}

func (m *mockStore) ServerNotify(request *pbs.ServerNotifyRequest) error {
	panic("implement me")
}

func (m *mockStore) RemoveServerReplica(serverName string, replicaIdx int) ([]string, error) {
	panic("implement me")
}

type mockGrpcStream struct {
	grpc.ServerStream
}

func (ms *mockGrpcStream) Send(msg *pb.ModelOperationMessage) error {
	return nil
}

func TestSync(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	g := NewGomegaWithT(t)

	type ExpectedVersionState struct {
		version        uint32
		expectedStates map[int]store.ReplicaStatus
	}
	type test struct {
		name                  string
		agents                map[ServerKey]*AgentSubscriber
		store                 *mockStore
		modelName             string
		expectedVersionStates []ExpectedVersionState
	}
	tests := []test{
		{
			name:      "simple",
			modelName: "iris",
			agents: map[ServerKey]*AgentSubscriber{
				ServerKey{serverName: "server1", replicaIdx: 1}: {stream: &mockGrpcStream{}},
			},
			store: &mockStore{
				models: map[string]*store.ModelSnapshot{
					"iris": {
						Name: "iris",
						Versions: []*store.ModelVersion{
							store.NewModelVersion(&pbs.Model{Meta: &pbs.MetaData{Name: "iris"}}, 1, "server1",
								map[int]store.ReplicaStatus{
									1: {State: store.LoadRequested},
								}, false, store.ModelProgressing),
						},
					},
				},
			},
			expectedVersionStates: []ExpectedVersionState{
				{
					version: 1,
					expectedStates: map[int]store.ReplicaStatus{
						1: {State: store.Loading},
					},
				},
			},
		},
		{
			name:      "OlderVersions",
			modelName: "iris",
			agents: map[ServerKey]*AgentSubscriber{
				ServerKey{serverName: "server1", replicaIdx: 1}: {stream: &mockGrpcStream{}},
			},
			store: &mockStore{
				models: map[string]*store.ModelSnapshot{
					"iris": {
						Name: "iris",
						Versions: []*store.ModelVersion{
							store.NewModelVersion(&pbs.Model{Meta: &pbs.MetaData{Name: "iris"}}, 1, "server1",
								map[int]store.ReplicaStatus{
									1: {State: store.UnloadRequested},
								}, false, store.ModelProgressing),
							store.NewModelVersion(&pbs.Model{Meta: &pbs.MetaData{Name: "iris"}}, 2, "server1",
								map[int]store.ReplicaStatus{
									1: {State: store.LoadRequested},
								}, false, store.ModelProgressing),
						},
					},
				},
			},
			expectedVersionStates: []ExpectedVersionState{
				{
					version: 1,
					expectedStates: map[int]store.ReplicaStatus{
						1: {State: store.Unloading},
					},
				},
				{
					version: 2,
					expectedStates: map[int]store.ReplicaStatus{
						1: {State: store.Loading},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			logger := log.New()
			eventHub, err := coordinator.NewEventHub(logger)
			g.Expect(err).To(BeNil())
			server := NewAgentServer(logger, test.store, nil, eventHub)
			server.agents = test.agents
			server.Sync(test.modelName)
			model, err := test.store.GetModel(test.modelName)
			g.Expect(err).To(BeNil())
			for _, expectedVersionState := range test.expectedVersionStates {
				mv := model.GetVersion(expectedVersionState.version)
				for replicaIdx, rs := range expectedVersionState.expectedStates {
					g.Expect(mv.ReplicaState()[replicaIdx].State).To(Equal(rs.State))
				}
			}
		})
	}
}
