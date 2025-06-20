/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package agent

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/seldonio/seldon-core/apis/go/v2/mlops/agent"
	pbs "github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/coordinator"
	testing_utils "github.com/seldonio/seldon-core/scheduler/v2/pkg/internal/testing_utils"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/scheduler"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/util"
)

type mockScheduler struct{}

var _ scheduler.Scheduler = (*mockScheduler)(nil)

func (s mockScheduler) Schedule(_ string) error {
	return nil
}

func (s mockScheduler) ScheduleFailedModels() ([]string, error) {
	return nil, nil
}

type mockStore struct {
	models map[string]*store.ModelSnapshot
}

var _ store.ModelStore = (*mockStore)(nil)

func (m *mockStore) FailedScheduling(modelVersion *store.ModelVersion, reason string, reset bool) {
}

func (m *mockStore) UpdateModel(config *pbs.LoadModelRequest) error {
	panic("implement me")
}

func (m *mockStore) GetModel(key string) (*store.ModelSnapshot, error) {
	return m.models[key], nil
}

func (f mockStore) GetModels() ([]*store.ModelSnapshot, error) {
	models := []*store.ModelSnapshot{}
	for _, m := range f.models {
		models = append(models, m)
	}
	return models, nil
}

func (m mockStore) LockModel(modelId string) {
}

func (m mockStore) UnlockModel(modelId string) {
}

func (m *mockStore) RemoveModel(req *pbs.UnloadModelRequest) error {
	panic("implement me")
}

func (m *mockStore) GetServers(shallow bool, modelDetails bool) ([]*store.ServerSnapshot, error) {
	panic("implement me")
}

func (m *mockStore) GetServer(serverKey string, shallow bool, modelDetails bool) (*store.ServerSnapshot, error) {
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

func (m *mockStore) UpdateModelState(modelKey string, version uint32, serverKey string, replicaIdx int, availableMemory *uint64, expectedState, desiredState store.ModelReplicaState, reason string, runtimeInfo *pbs.ModelRuntimeInfo) error {
	model := m.models[modelKey]
	for _, mv := range model.Versions {
		if mv.GetVersion() == version {
			mv.SetReplicaState(replicaIdx, desiredState, reason)
		}
	}
	return nil
}

func (m *mockStore) AddServerReplica(request *pb.AgentSubscribeRequest) error {
	return nil
}

func (m *mockStore) ServerNotify(request *pbs.ServerNotify) error {
	panic("implement me")
}

func (m *mockStore) RemoveServerReplica(serverName string, replicaIdx int) ([]string, error) {
	return nil, nil
}

func (m *mockStore) DrainServerReplica(serverName string, replicaIdx int) ([]string, error) {
	panic("implement me")
}

func (m *mockStore) GetAllModels() []string {
	var modelNames []string
	for modelName := range m.models {
		modelNames = append(modelNames, modelName)
	}
	return modelNames
}

type mockGrpcStream struct {
	err error
	grpc.ServerStream
}

func (ms *mockGrpcStream) Send(msg *pb.ModelOperationMessage) error {
	return ms.err
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
				{serverName: "server1", replicaIdx: 1}: {stream: &mockGrpcStream{}},
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
			name:      "simple - error load",
			modelName: "iris",
			agents: map[ServerKey]*AgentSubscriber{
				{serverName: "server1", replicaIdx: 1}: {stream: &mockGrpcStream{err: fmt.Errorf("error send")}},
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
						1: {State: store.LoadFailed},
					},
				},
			},
		},
		{
			name:      "simple - unload",
			modelName: "iris",
			agents: map[ServerKey]*AgentSubscriber{
				{serverName: "server1", replicaIdx: 1}: {stream: &mockGrpcStream{}},
			},
			store: &mockStore{
				models: map[string]*store.ModelSnapshot{
					"iris": {
						Name: "iris",
						Versions: []*store.ModelVersion{
							store.NewModelVersion(&pbs.Model{Meta: &pbs.MetaData{Name: "iris"}}, 1, "server1",
								map[int]store.ReplicaStatus{
									1: {State: store.UnloadRequested},
								}, false, store.ModelTerminating),
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
			},
		},
		{
			name:      "simple - error unload",
			modelName: "iris",
			agents: map[ServerKey]*AgentSubscriber{
				{serverName: "server1", replicaIdx: 1}: {stream: &mockGrpcStream{err: fmt.Errorf("error send")}},
			},
			store: &mockStore{
				models: map[string]*store.ModelSnapshot{
					"iris": {
						Name: "iris",
						Versions: []*store.ModelVersion{
							store.NewModelVersion(&pbs.Model{Meta: &pbs.MetaData{Name: "iris"}}, 1, "server1",
								map[int]store.ReplicaStatus{
									1: {State: store.UnloadRequested},
								}, false, store.ModelTerminating),
						},
					},
				},
			},
			expectedVersionStates: []ExpectedVersionState{
				{
					version: 1,
					expectedStates: map[int]store.ReplicaStatus{
						1: {State: store.UnloadFailed},
					},
				},
			},
		},
		{
			name:      "OlderVersions",
			modelName: "iris",
			agents: map[ServerKey]*AgentSubscriber{
				{serverName: "server1", replicaIdx: 1}: {stream: &mockGrpcStream{}},
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
			server := NewAgentServer(logger, test.store, nil, eventHub, false)
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

func TestCalculateDesiredReplicas(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	g := NewGomegaWithT(t)

	type test struct {
		name                string
		trigger             pb.ModelScalingTriggerMessage_Trigger
		previousNumReplicas int
		minNumReplicas      int
		maxNumReplicas      int
		expectedNumReplicas int
		err                 bool
	}
	tests := []test{
		{
			name:                "scale up",
			trigger:             pb.ModelScalingTriggerMessage_SCALE_UP,
			previousNumReplicas: 1,
			minNumReplicas:      1,
			maxNumReplicas:      0,
			expectedNumReplicas: 2,
			err:                 false,
		},
		{
			name:                "scale down",
			trigger:             pb.ModelScalingTriggerMessage_SCALE_DOWN,
			previousNumReplicas: 2,
			minNumReplicas:      1,
			maxNumReplicas:      0,
			expectedNumReplicas: 1,
			err:                 false,
		},
		{
			name:                "cannot scale down",
			trigger:             pb.ModelScalingTriggerMessage_SCALE_DOWN,
			previousNumReplicas: 1,
			minNumReplicas:      0,
			maxNumReplicas:      2,
			expectedNumReplicas: 0,
			err:                 true,
		},
		{
			name:                "scaling not enabled",
			trigger:             pb.ModelScalingTriggerMessage_SCALE_DOWN,
			previousNumReplicas: 2,
			minNumReplicas:      0,
			maxNumReplicas:      0,
			expectedNumReplicas: 0,
			err:                 true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dummyModel := pbs.Model{
				Meta:       nil,
				ModelSpec:  nil,
				StreamSpec: nil,
				DeploymentSpec: &pbs.DeploymentSpec{
					Replicas:    uint32(test.previousNumReplicas),
					MinReplicas: uint32(test.minNumReplicas),
					MaxReplicas: uint32(test.maxNumReplicas),
				},
			}
			numReplicas, err := calculateDesiredNumReplicas(
				&dummyModel, test.trigger, test.previousNumReplicas)
			if test.err {
				g.Expect(err).ToNot(BeNil())
			} else {
				g.Expect(err).To(BeNil())
				g.Expect(numReplicas).To(Equal(test.expectedNumReplicas))
			}
		})
	}
}

func TestModelScalingProtos(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	g := NewGomegaWithT(t)

	type test struct {
		name                string
		store               *mockStore
		trigger             pb.ModelScalingTriggerMessage_Trigger
		triggerModelName    string
		triggerModelVersion int
		expectedReplicas    uint32
		lastUpdate          time.Time
		isError             bool
	}
	tests := []test{
		{
			name: "scale up not enabled",
			store: &mockStore{
				models: map[string]*store.ModelSnapshot{
					"iris": {
						Name: "iris",
						Versions: []*store.ModelVersion{
							store.NewModelVersion(
								&pbs.Model{
									Meta:           &pbs.MetaData{Name: "iris"},
									DeploymentSpec: &pbs.DeploymentSpec{Replicas: 1},
								},
								1, "server1",
								map[int]store.ReplicaStatus{
									1: {State: store.Available},
								}, false, store.ModelAvailable),
						},
					},
				},
			},
			trigger:             pb.ModelScalingTriggerMessage_SCALE_UP,
			triggerModelName:    "iris",
			triggerModelVersion: 1,
			expectedReplicas:    1,
			isError:             true,
		},
		{
			name: "scale up within range no max",
			store: &mockStore{
				models: map[string]*store.ModelSnapshot{
					"iris": {
						Name: "iris",
						Versions: []*store.ModelVersion{
							store.NewModelVersion(
								&pbs.Model{
									Meta:           &pbs.MetaData{Name: "iris"},
									DeploymentSpec: &pbs.DeploymentSpec{Replicas: 1, MinReplicas: 1},
								},
								1, "server1",
								map[int]store.ReplicaStatus{
									1: {State: store.Available},
								}, false, store.ModelAvailable),
						},
					},
				},
			},
			trigger:             pb.ModelScalingTriggerMessage_SCALE_UP,
			triggerModelName:    "iris",
			triggerModelVersion: 1,
			expectedReplicas:    2,
			isError:             false,
		},
		{
			name: "scale up within range",
			store: &mockStore{
				models: map[string]*store.ModelSnapshot{
					"iris": {
						Name: "iris",
						Versions: []*store.ModelVersion{
							store.NewModelVersion(
								&pbs.Model{
									Meta:           &pbs.MetaData{Name: "iris"},
									DeploymentSpec: &pbs.DeploymentSpec{Replicas: 1, MinReplicas: 1, MaxReplicas: 2},
								},
								1, "server1",
								map[int]store.ReplicaStatus{
									1: {State: store.Available},
								}, false, store.ModelAvailable),
						},
					},
				},
			},
			trigger:             pb.ModelScalingTriggerMessage_SCALE_UP,
			triggerModelName:    "iris",
			triggerModelVersion: 1,
			expectedReplicas:    2,
			isError:             false,
		},
		{
			name: "scale up not within range",
			store: &mockStore{
				models: map[string]*store.ModelSnapshot{
					"iris": {
						Name: "iris",
						Versions: []*store.ModelVersion{
							store.NewModelVersion(
								&pbs.Model{
									Meta:           &pbs.MetaData{Name: "iris"},
									DeploymentSpec: &pbs.DeploymentSpec{Replicas: 1, MinReplicas: 1, MaxReplicas: 1},
								},
								1, "server1",
								map[int]store.ReplicaStatus{
									1: {State: store.Available},
								}, false, store.ModelAvailable),
						},
					},
				},
			},
			trigger:             pb.ModelScalingTriggerMessage_SCALE_UP,
			triggerModelName:    "iris",
			triggerModelVersion: 1,
			expectedReplicas:    1,
			isError:             true,
		},
		{
			name: "scale down within range",
			store: &mockStore{
				models: map[string]*store.ModelSnapshot{
					"iris": {
						Name: "iris",
						Versions: []*store.ModelVersion{
							store.NewModelVersion(
								&pbs.Model{
									Meta:           &pbs.MetaData{Name: "iris"},
									DeploymentSpec: &pbs.DeploymentSpec{Replicas: 2, MinReplicas: 1, MaxReplicas: 2},
								},
								1, "server1",
								map[int]store.ReplicaStatus{
									1: {State: store.Available}, 2: {State: store.Available},
								}, false, store.ModelAvailable),
						},
					},
				},
			},
			trigger:             pb.ModelScalingTriggerMessage_SCALE_DOWN,
			triggerModelName:    "iris",
			triggerModelVersion: 1,
			expectedReplicas:    1,
			isError:             false,
		},
		{
			name: "scale down not within range",
			store: &mockStore{
				models: map[string]*store.ModelSnapshot{
					"iris": {
						Name: "iris",
						Versions: []*store.ModelVersion{
							store.NewModelVersion(
								&pbs.Model{
									Meta:           &pbs.MetaData{Name: "iris"},
									DeploymentSpec: &pbs.DeploymentSpec{Replicas: 2, MinReplicas: 2, MaxReplicas: 3},
								},
								1, "server1",
								map[int]store.ReplicaStatus{
									1: {State: store.Available}, 2: {State: store.Available},
								}, false, store.ModelAvailable),
						},
					},
				},
			},
			trigger:             pb.ModelScalingTriggerMessage_SCALE_DOWN,
			triggerModelName:    "iris",
			triggerModelVersion: 1,
			expectedReplicas:    2,
			isError:             true,
		},
		{
			name: "scale down not enabled",
			store: &mockStore{
				models: map[string]*store.ModelSnapshot{
					"iris": {
						Name: "iris",
						Versions: []*store.ModelVersion{
							store.NewModelVersion(
								&pbs.Model{
									Meta:           &pbs.MetaData{Name: "iris"},
									DeploymentSpec: &pbs.DeploymentSpec{Replicas: 2},
								},
								1, "server1",
								map[int]store.ReplicaStatus{
									1: {State: store.Available}, 2: {State: store.Available},
								}, false, store.ModelAvailable),
						},
					},
				},
			},
			trigger:             pb.ModelScalingTriggerMessage_SCALE_DOWN,
			triggerModelName:    "iris",
			triggerModelVersion: 1,
			expectedReplicas:    2,
			isError:             true,
		},
		{
			name: "model not stable, scale down - should not proceed",
			store: &mockStore{
				models: map[string]*store.ModelSnapshot{
					"iris": {
						Name: "iris",
						Versions: []*store.ModelVersion{
							store.NewModelVersion(
								&pbs.Model{
									Meta:           &pbs.MetaData{Name: "iris"},
									DeploymentSpec: &pbs.DeploymentSpec{Replicas: 2, MinReplicas: 1},
								},
								1, "server1",
								map[int]store.ReplicaStatus{
									1: {State: store.Available}, 2: {State: store.Available},
								}, false, store.ModelAvailable),
						},
					},
				},
			},
			trigger:             pb.ModelScalingTriggerMessage_SCALE_DOWN,
			triggerModelName:    "iris",
			triggerModelVersion: 1,
			expectedReplicas:    2,
			lastUpdate:          time.Now(),
			isError:             true,
		},
		{
			name: "model not stable, scale up - should proceed",
			store: &mockStore{
				models: map[string]*store.ModelSnapshot{
					"iris": {
						Name: "iris",
						Versions: []*store.ModelVersion{
							store.NewModelVersion(
								&pbs.Model{
									Meta:           &pbs.MetaData{Name: "iris"},
									DeploymentSpec: &pbs.DeploymentSpec{Replicas: 2, MinReplicas: 1},
								},
								1, "server1",
								map[int]store.ReplicaStatus{
									1: {State: store.Available}, 2: {State: store.Available},
								}, false, store.ModelAvailable),
						},
					},
				},
			},
			trigger:             pb.ModelScalingTriggerMessage_SCALE_UP,
			triggerModelName:    "iris",
			triggerModelVersion: 1,
			expectedReplicas:    3,
			lastUpdate:          time.Now(),
			isError:             false,
		},
		{
			name: "model not available, scale up - should not proceed",
			store: &mockStore{
				models: map[string]*store.ModelSnapshot{
					"iris": {
						Name: "iris",
						Versions: []*store.ModelVersion{
							store.NewModelVersion(
								&pbs.Model{
									Meta:           &pbs.MetaData{Name: "iris"},
									DeploymentSpec: &pbs.DeploymentSpec{Replicas: 2, MinReplicas: 1},
								},
								1, "server1",
								map[int]store.ReplicaStatus{
									1: {State: store.Available}, 2: {State: store.LoadFailed},
								}, false, store.ScheduleFailed),
						},
					},
				},
			},
			trigger:             pb.ModelScalingTriggerMessage_SCALE_UP,
			triggerModelName:    "iris",
			triggerModelVersion: 1,
			expectedReplicas:    2,
			isError:             true,
		},
		{
			name: "model not available, scale down - should proceed",
			store: &mockStore{
				models: map[string]*store.ModelSnapshot{
					"iris": {
						Name: "iris",
						Versions: []*store.ModelVersion{
							store.NewModelVersion(
								&pbs.Model{
									Meta:           &pbs.MetaData{Name: "iris"},
									DeploymentSpec: &pbs.DeploymentSpec{Replicas: 2, MinReplicas: 1},
								},
								1, "server1",
								map[int]store.ReplicaStatus{
									1: {State: store.Available}, 2: {State: store.LoadFailed},
								}, false, store.ScheduleFailed),
						},
					},
				},
			},
			trigger:             pb.ModelScalingTriggerMessage_SCALE_DOWN,
			triggerModelName:    "iris",
			triggerModelVersion: 1,
			expectedReplicas:    1,
			isError:             false,
		},
		{
			name: "model available is not latest, scale up - should not proceed",
			store: &mockStore{
				models: map[string]*store.ModelSnapshot{
					"iris": {
						Name: "iris",
						Versions: []*store.ModelVersion{
							store.NewModelVersion(
								&pbs.Model{
									Meta:           &pbs.MetaData{Name: "iris"},
									DeploymentSpec: &pbs.DeploymentSpec{Replicas: 2, MinReplicas: 1},
								},
								1, "server1",
								map[int]store.ReplicaStatus{
									1: {State: store.Available}, 2: {State: store.Available},
								}, false, store.ModelAvailable),
							store.NewModelVersion(
								&pbs.Model{
									Meta:           &pbs.MetaData{Name: "iris"},
									DeploymentSpec: &pbs.DeploymentSpec{Replicas: 2, MinReplicas: 1},
								},
								2, "server1",
								map[int]store.ReplicaStatus{
									1: {State: store.Available}, 2: {State: store.Loading},
								}, false, store.ModelProgressing),
						},
					},
				},
			},
			trigger:             pb.ModelScalingTriggerMessage_SCALE_UP,
			triggerModelName:    "iris",
			triggerModelVersion: 2,
			expectedReplicas:    2,
			lastUpdate:          time.Now(),
			isError:             true,
		},
		{
			name: "model versions mismatch - should not proceed",
			store: &mockStore{
				models: map[string]*store.ModelSnapshot{
					"iris": {
						Name: "iris",
						Versions: []*store.ModelVersion{
							store.NewModelVersion(
								&pbs.Model{
									Meta:           &pbs.MetaData{Name: "iris"},
									DeploymentSpec: &pbs.DeploymentSpec{Replicas: 2, MinReplicas: 1},
								},
								1, "server1",
								map[int]store.ReplicaStatus{
									1: {State: store.Available}, 2: {State: store.Available},
								}, false, store.ModelAvailable),
							store.NewModelVersion(
								&pbs.Model{
									Meta:           &pbs.MetaData{Name: "iris"},
									DeploymentSpec: &pbs.DeploymentSpec{Replicas: 2, MinReplicas: 1},
								},
								2, "server1",
								map[int]store.ReplicaStatus{
									1: {State: store.Available}, 2: {State: store.Available},
								}, false, store.ModelState(store.Available)),
						},
					},
				},
			},
			trigger:             pb.ModelScalingTriggerMessage_SCALE_UP,
			triggerModelName:    "iris",
			triggerModelVersion: 1,
			expectedReplicas:    2,
			isError:             true,
		},
		{
			name: "model does not exist in scheduler state",
			store: &mockStore{
				models: map[string]*store.ModelSnapshot{},
			},
			trigger:             pb.ModelScalingTriggerMessage_SCALE_UP,
			triggerModelName:    "iris",
			triggerModelVersion: 1,
			expectedReplicas:    2,
			isError:             true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			model, _ := test.store.GetModel(test.triggerModelName)
			if model != nil { // in the cases where the model is not in the scheduler state yet
				lastestModel := model.GetLatest()
				state := lastestModel.ModelState()
				state.Timestamp = test.lastUpdate
				lastestModel.SetModelState(state)
			} else {
				model = &store.ModelSnapshot{
					Name: test.triggerModelName,
				}
			}

			protos, err := createScalingPseudoRequest(&pb.ModelScalingTriggerMessage{
				ModelName:    test.triggerModelName,
				ModelVersion: uint32(test.triggerModelVersion),
				Trigger:      test.trigger,
			}, model)
			if !test.isError {
				g.Expect(err).To(BeNil())
				g.Expect(protos.GetDeploymentSpec().GetReplicas()).To(Equal(test.expectedReplicas))
			} else {
				g.Expect(err).NotTo(BeNil())
			}
		})
	}
}

func TestModelRelocatedWaiterSmoke(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	g := NewGomegaWithT(t)

	type serverReplica struct {
		serverName string
		serverIdx  int
	}
	type in struct {
		serverReplica serverReplica
		models        []string
	}
	type test struct {
		name            string
		input           []in
		serverUnderTest int
	}
	tests := []test{
		{
			name: "simple",
			input: []in{
				{
					serverReplica: serverReplica{
						serverName: "server",
						serverIdx:  1,
					},
					models: []string{"model1", "model2"},
				},
			},
			serverUnderTest: 0,
		},
		{
			name: "simple - no models",
			input: []in{
				{
					serverReplica: serverReplica{
						serverName: "server",
						serverIdx:  1,
					},
					models: []string{},
				},
			},
			serverUnderTest: 0,
		},
		{
			name: "twoservers",
			input: []in{
				{
					serverReplica: serverReplica{
						serverName: "server",
						serverIdx:  1,
					},
					models: []string{"model1", "model2"},
				},
				{
					serverReplica: serverReplica{
						serverName: "server",
						serverIdx:  2,
					},
					models: []string{"model1", "model2", "model3"},
				},
			},
			serverUnderTest: 0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			waiter := newModelRelocatedWaiter()
			for _, i := range test.input {
				waiter.registerServerReplica(i.serverReplica.serverName, i.serverReplica.serverIdx, i.models)
			}
			modelsToDrain := test.input[test.serverUnderTest].models
			for _, model := range modelsToDrain {
				waiter.signalModel(model)
			}
			serverToDrain := test.input[test.serverUnderTest].serverReplica
			waiter.wait(serverToDrain.serverName, serverToDrain.serverIdx)
			// make sure we did clean up
			_, ok := waiter.serverReplicaModels[waiter.getServerReplicaName(serverToDrain.serverName, serverToDrain.serverIdx)]
			g.Expect(ok).To(BeFalse())
			if len(test.input) > 1 { // we have more than one server replica to drain
				size := len(waiter.serverReplicaModels)
				g.Expect(size).To(BeNumerically(">", 0))
			}
			// test signal random model, working fine
			waiter.signalModel("dummy")
		})
	}
}

func TestAutoscalingEnabled(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	g := NewGomegaWithT(t)

	dummyModelName := "iris"

	type test struct {
		name    string
		model   *pbs.Model
		enabled bool
	}
	tests := []test{
		{
			name: "enabled - minreplica set",
			model: &pbs.Model{
				Meta:           &pbs.MetaData{Name: dummyModelName},
				DeploymentSpec: &pbs.DeploymentSpec{Replicas: 2, MinReplicas: 1},
			},
			enabled: true,
		},
		{
			name: "enabled - maxreplica set",
			model: &pbs.Model{
				Meta:           &pbs.MetaData{Name: dummyModelName},
				DeploymentSpec: &pbs.DeploymentSpec{Replicas: 2, MaxReplicas: 3},
			},
			enabled: true,
		},
		{
			name: "disabled",
			model: &pbs.Model{
				Meta:           &pbs.MetaData{Name: dummyModelName},
				DeploymentSpec: &pbs.DeploymentSpec{Replicas: 2},
			},
			enabled: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			enabled := util.AutoscalingEnabled(test.model.DeploymentSpec.MinReplicas, test.model.DeploymentSpec.MaxReplicas)
			g.Expect(enabled).To(Equal(test.enabled))
		})
	}
}

func TestSubscribe(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	g := NewGomegaWithT(t)

	type ag struct {
		id      uint32
		doClose bool
	}
	type test struct {
		name                          string
		agents                        []ag
		expectedAgentsCount           int
		expectedAgentsCountAfterClose int
	}
	tests := []test{
		{
			name: "simple",
			agents: []ag{
				{1, true}, {2, true},
			},
			expectedAgentsCount:           2,
			expectedAgentsCountAfterClose: 0,
		},
		{
			name: "simple - no close",
			agents: []ag{
				{1, true}, {2, false},
			},
			expectedAgentsCount:           2,
			expectedAgentsCountAfterClose: 1,
		},
		{
			name: "duplicates",
			agents: []ag{
				{1, true}, {1, false},
			},
			expectedAgentsCount:           1,
			expectedAgentsCountAfterClose: 1,
		},
		{
			name: "duplicates with all close",
			agents: []ag{
				{1, true}, {1, true}, {1, true},
			},
			expectedAgentsCount:           1,
			expectedAgentsCountAfterClose: 0,
		},
	}

	getStream := func(id uint32, context context.Context, port int) *grpc.ClientConn {
		conn, _ := grpc.NewClient(fmt.Sprintf(":%d", port), grpc.WithTransportCredentials(insecure.NewCredentials()))
		grpcClient := pb.NewAgentServiceClient(conn)
		_, _ = grpcClient.Subscribe(
			context,
			&pb.AgentSubscribeRequest{
				ServerName:           "dummy",
				ReplicaIdx:           id,
				ReplicaConfig:        &pb.ReplicaConfig{},
				Shared:               true,
				AvailableMemoryBytes: 0,
			},
		)
		return conn
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			logger := log.New()
			eventHub, err := coordinator.NewEventHub(logger)
			g.Expect(err).To(BeNil())
			server := NewAgentServer(logger, &mockStore{}, mockScheduler{}, eventHub, false)
			port, err := testing_utils.GetFreePortForTest()
			if err != nil {
				t.Fatal(err)
			}
			err = server.startServer(uint(port), false)
			if err != nil {
				t.Fatal(err)
			}
			time.Sleep(100 * time.Millisecond)

			mu := sync.Mutex{}
			streams := make([]*grpc.ClientConn, 0)
			for _, a := range test.agents {
				go func(id uint32) {
					conn := getStream(id, context.Background(), port)
					mu.Lock()
					streams = append(streams, conn)
					mu.Unlock()
				}(a.id)
			}

			maxCount := 10
			count := 0
			for len(server.agents) != test.expectedAgentsCount && count < maxCount {
				time.Sleep(100 * time.Millisecond)
				count++
			}
			g.Expect(len(server.agents)).To(Equal(test.expectedAgentsCount))

			for idx, s := range streams {
				go func(idx int, s *grpc.ClientConn) {
					if test.agents[idx].doClose {
						s.Close()
					}
				}(idx, s)
			}

			count = 0
			for len(server.agents) != test.expectedAgentsCountAfterClose && count < maxCount {
				time.Sleep(100 * time.Millisecond)
				count++
			}
			g.Expect(len(server.agents)).To(Equal(test.expectedAgentsCountAfterClose))

			server.StopAgentStreams()
		})
	}
}
