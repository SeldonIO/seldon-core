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
	"github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler/db"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/internal/testing_utils"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/scheduler/mock"
	log "github.com/sirupsen/logrus"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/seldonio/seldon-core/apis/go/v2/mlops/agent"
	pbs "github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"
	"github.com/seldonio/seldon-core/components/tls/v2/pkg/tls"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/coordinator"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/util"
)

type mockGrpcStream struct {
	err error
	grpc.ServerStream
	ctx context.Context
}

func (ms *mockGrpcStream) Send(msg *pb.ModelOperationMessage) error {
	return ms.err
}

func (ms *mockGrpcStream) Context() context.Context {
	return ms.ctx
}

func TestSync(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	g := NewGomegaWithT(t)
	cancelledCtx, cancel := context.WithTimeout(context.Background(), 0)
	defer cancel()

	type ExpectedVersionState struct {
		version        uint32
		expectedStates map[int]db.ReplicaStatus
	}
	type test struct {
		name                  string
		agents                map[ServerKey]*AgentSubscriber
		models                []*db.Model
		servers               []*db.Server
		modelName             string
		expectedVersionStates []ExpectedVersionState
	}
	tests := []test{
		{
			name:      "success - simple",
			modelName: "iris",
			agents: map[ServerKey]*AgentSubscriber{
				{serverName: "server1", replicaIdx: 1}: {stream: &mockGrpcStream{ctx: context.Background()}},
			},
			models: []*db.Model{
				{
					Name: "iris",
					Versions: []*db.ModelVersion{
						util.NewTestModelVersion(&pbs.Model{Meta: &pbs.MetaData{Name: "iris"}}, 1, "server1",
							map[int32]*db.ReplicaStatus{
								1: {State: db.ModelReplicaState_LoadRequested},
							}, db.ModelState_ModelProgressing),
					},
				},
			},
			servers: []*db.Server{
				{
					Name: "server1",
					Replicas: map[int32]*db.ServerReplica{
						1: {},
					},
				},
			},
			expectedVersionStates: []ExpectedVersionState{
				{
					version: 1,
					expectedStates: map[int]db.ReplicaStatus{
						1: {State: db.ModelReplicaState_Loading},
					},
				},
			},
		},
		{
			name:      "failure - stream ctx cancelled",
			modelName: "iris",
			agents: map[ServerKey]*AgentSubscriber{
				{serverName: "server1", replicaIdx: 1}: {stream: &mockGrpcStream{ctx: cancelledCtx}},
			},
			models: []*db.Model{
				{
					Name: "iris",
					Versions: []*db.ModelVersion{
						util.NewTestModelVersion(&pbs.Model{Meta: &pbs.MetaData{Name: "iris"}}, 1, "server1",
							map[int32]*db.ReplicaStatus{
								1: {State: db.ModelReplicaState_LoadRequested},
							}, db.ModelState_ModelProgressing),
					},
				},
			},
			servers: []*db.Server{
				{
					Name: "server1",
					Replicas: map[int32]*db.ServerReplica{
						1: {},
					},
				},
			},
			expectedVersionStates: []ExpectedVersionState{
				{
					version: 1,
					expectedStates: map[int]db.ReplicaStatus{
						1: {State: db.ModelReplicaState_LoadFailed},
					},
				},
			},
		},
		{
			name:      "failure - simple - error load",
			modelName: "iris",
			agents: map[ServerKey]*AgentSubscriber{
				{serverName: "server1", replicaIdx: 1}: {stream: &mockGrpcStream{ctx: context.Background(), err: fmt.Errorf("error send")}},
			},
			models: []*db.Model{
				{
					Name: "iris",
					Versions: []*db.ModelVersion{
						util.NewTestModelVersion(&pbs.Model{Meta: &pbs.MetaData{Name: "iris"}}, 1, "server1",
							map[int32]*db.ReplicaStatus{
								1: {State: db.ModelReplicaState_LoadRequested},
							}, db.ModelState_ModelProgressing),
					},
				},
			},
			servers: []*db.Server{
				{
					Name: "server1",
					Replicas: map[int32]*db.ServerReplica{
						1: {},
					},
				},
			},
			expectedVersionStates: []ExpectedVersionState{
				{
					version: 1,
					expectedStates: map[int]db.ReplicaStatus{
						1: {State: db.ModelReplicaState_LoadFailed},
					},
				},
			},
		},
		{
			name:      "success - simple - unload",
			modelName: "iris",
			agents: map[ServerKey]*AgentSubscriber{
				{serverName: "server1", replicaIdx: 1}: {stream: &mockGrpcStream{ctx: context.Background()}},
			},
			models: []*db.Model{
				{
					Name: "iris",
					Versions: []*db.ModelVersion{
						util.NewTestModelVersion(&pbs.Model{Meta: &pbs.MetaData{Name: "iris"}}, 1, "server1",
							map[int32]*db.ReplicaStatus{
								1: {State: db.ModelReplicaState_UnloadRequested},
							}, db.ModelState_ModelTerminating),
					},
				},
			},
			servers: []*db.Server{
				{
					Name: "server1",
					Replicas: map[int32]*db.ServerReplica{
						1: {},
					},
				},
			},
			expectedVersionStates: []ExpectedVersionState{
				{
					version: 1,
					expectedStates: map[int]db.ReplicaStatus{
						1: {State: db.ModelReplicaState_Unloading},
					},
				},
			},
		},
		{
			name:      "failure - simple - error unload",
			modelName: "iris",
			agents: map[ServerKey]*AgentSubscriber{
				{serverName: "server1", replicaIdx: 1}: {stream: &mockGrpcStream{ctx: context.Background(), err: fmt.Errorf("error send")}},
			},
			models: []*db.Model{
				{
					Name: "iris",
					Versions: []*db.ModelVersion{
						util.NewTestModelVersion(&pbs.Model{Meta: &pbs.MetaData{Name: "iris"}}, 1, "server1",
							map[int32]*db.ReplicaStatus{
								1: {State: db.ModelReplicaState_UnloadRequested},
							}, db.ModelState_ModelTerminating),
					},
				},
			},
			servers: []*db.Server{
				{
					Name: "server1",
					Replicas: map[int32]*db.ServerReplica{
						1: {},
					},
				},
			},
			expectedVersionStates: []ExpectedVersionState{
				{
					version: 1,
					expectedStates: map[int]db.ReplicaStatus{
						1: {State: db.ModelReplicaState_UnloadFailed},
					},
				},
			},
		},
		{
			name:      "success - OlderVersions",
			modelName: "iris",
			agents: map[ServerKey]*AgentSubscriber{
				{serverName: "server1", replicaIdx: 1}: {stream: &mockGrpcStream{ctx: context.Background()}},
			},
			models: []*db.Model{
				{
					Name: "iris",
					Versions: []*db.ModelVersion{
						util.NewTestModelVersion(&pbs.Model{Meta: &pbs.MetaData{Name: "iris"}}, 1, "server1",
							map[int32]*db.ReplicaStatus{
								1: {State: db.ModelReplicaState_UnloadRequested},
							}, db.ModelState_ModelProgressing),
						util.NewTestModelVersion(&pbs.Model{Meta: &pbs.MetaData{Name: "iris"}}, 2, "server1",
							map[int32]*db.ReplicaStatus{
								1: {State: db.ModelReplicaState_LoadRequested},
							}, db.ModelState_ModelProgressing),
					},
				},
			},
			servers: []*db.Server{
				{
					Name: "server1",
					Replicas: map[int32]*db.ServerReplica{
						1: {},
					},
				},
			},
			expectedVersionStates: []ExpectedVersionState{
				{
					version: 1,
					expectedStates: map[int]db.ReplicaStatus{
						1: {State: db.ModelReplicaState_Unloading},
					},
				},
				{
					version: 2,
					expectedStates: map[int]db.ReplicaStatus{
						1: {State: db.ModelReplicaState_Loading},
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

			// Create storage instances
			modelStorage := store.NewInMemoryStorage[*db.Model]()
			serverStorage := store.NewInMemoryStorage[*db.Server]()

			// Populate storage with test data
			for _, model := range test.models {
				err := modelStorage.Insert(context.TODO(), model)
				g.Expect(err).To(BeNil())
			}
			for _, server := range test.servers {
				err := serverStorage.Insert(context.TODO(), server)
				g.Expect(err).To(BeNil())
			}

			// Create MemoryStore with populated storage
			ms := store.NewModelServerStore(logger, modelStorage, serverStorage, eventHub)

			server := NewAgentServer(logger, ms, nil, eventHub, false, tls.TLSOptions{})
			server.agents = test.agents
			server.Sync(test.modelName)
			model, err := modelStorage.Get(context.TODO(), test.modelName)
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
		models              []*db.Model
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
			models: []*db.Model{
				{
					Name: "iris",
					Versions: []*db.ModelVersion{
						util.NewTestModelVersion(
							&pbs.Model{
								Meta:           &pbs.MetaData{Name: "iris"},
								DeploymentSpec: &pbs.DeploymentSpec{Replicas: 1},
							},
							1, "server1",
							map[int32]*db.ReplicaStatus{
								1: {State: db.ModelReplicaState_Available},
							}, db.ModelState_ModelAvailable),
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
			models: []*db.Model{
				{
					Name: "iris",
					Versions: []*db.ModelVersion{
						util.NewTestModelVersion(
							&pbs.Model{
								Meta:           &pbs.MetaData{Name: "iris"},
								DeploymentSpec: &pbs.DeploymentSpec{Replicas: 1, MinReplicas: 1},
							},
							1, "server1",
							map[int32]*db.ReplicaStatus{
								1: {State: db.ModelReplicaState_Available},
							}, db.ModelState_ModelAvailable),
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
			models: []*db.Model{
				{
					Name: "iris",
					Versions: []*db.ModelVersion{
						util.NewTestModelVersion(
							&pbs.Model{
								Meta:           &pbs.MetaData{Name: "iris"},
								DeploymentSpec: &pbs.DeploymentSpec{Replicas: 1, MinReplicas: 1, MaxReplicas: 2},
							},
							1, "server1",
							map[int32]*db.ReplicaStatus{
								1: {State: db.ModelReplicaState_Available},
							}, db.ModelState_ModelAvailable),
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
			models: []*db.Model{
				{
					Name: "iris",
					Versions: []*db.ModelVersion{
						util.NewTestModelVersion(
							&pbs.Model{
								Meta:           &pbs.MetaData{Name: "iris"},
								DeploymentSpec: &pbs.DeploymentSpec{Replicas: 1, MinReplicas: 1, MaxReplicas: 1},
							},
							1, "server1",
							map[int32]*db.ReplicaStatus{
								1: {State: db.ModelReplicaState_Available},
							}, db.ModelState_ModelAvailable),
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
			models: []*db.Model{
				{
					Name: "iris",
					Versions: []*db.ModelVersion{
						util.NewTestModelVersion(
							&pbs.Model{
								Meta:           &pbs.MetaData{Name: "iris"},
								DeploymentSpec: &pbs.DeploymentSpec{Replicas: 2, MinReplicas: 1, MaxReplicas: 2},
							},
							1, "server1",
							map[int32]*db.ReplicaStatus{
								1: {State: db.ModelReplicaState_Available}, 2: {State: db.ModelReplicaState_Available},
							}, db.ModelState_ModelAvailable),
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
			models: []*db.Model{
				{
					Name: "iris",
					Versions: []*db.ModelVersion{
						util.NewTestModelVersion(
							&pbs.Model{
								Meta:           &pbs.MetaData{Name: "iris"},
								DeploymentSpec: &pbs.DeploymentSpec{Replicas: 2, MinReplicas: 2, MaxReplicas: 3},
							},
							1, "server1",
							map[int32]*db.ReplicaStatus{
								1: {State: db.ModelReplicaState_Available}, 2: {State: db.ModelReplicaState_Available},
							}, db.ModelState_ModelAvailable),
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
			models: []*db.Model{
				{
					Name: "iris",
					Versions: []*db.ModelVersion{
						util.NewTestModelVersion(
							&pbs.Model{
								Meta:           &pbs.MetaData{Name: "iris"},
								DeploymentSpec: &pbs.DeploymentSpec{Replicas: 2},
							},
							1, "server1",
							map[int32]*db.ReplicaStatus{
								1: {State: db.ModelReplicaState_Available}, 2: {State: db.ModelReplicaState_Available},
							}, db.ModelState_ModelAvailable),
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
			models: []*db.Model{
				{
					Name: "iris",
					Versions: []*db.ModelVersion{
						util.NewTestModelVersion(
							&pbs.Model{
								Meta:           &pbs.MetaData{Name: "iris"},
								DeploymentSpec: &pbs.DeploymentSpec{Replicas: 2, MinReplicas: 1},
							},
							1, "server1",
							map[int32]*db.ReplicaStatus{
								1: {State: db.ModelReplicaState_Available}, 2: {State: db.ModelReplicaState_Available},
							}, db.ModelState_ModelAvailable),
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
			models: []*db.Model{
				{
					Name: "iris",
					Versions: []*db.ModelVersion{
						util.NewTestModelVersion(
							&pbs.Model{
								Meta:           &pbs.MetaData{Name: "iris"},
								DeploymentSpec: &pbs.DeploymentSpec{Replicas: 2, MinReplicas: 1},
							},
							1, "server1",
							map[int32]*db.ReplicaStatus{
								1: {State: db.ModelReplicaState_Available}, 2: {State: db.ModelReplicaState_Available},
							}, db.ModelState_ModelAvailable),
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
			models: []*db.Model{
				{
					Name: "iris",
					Versions: []*db.ModelVersion{
						util.NewTestModelVersion(
							&pbs.Model{
								Meta:           &pbs.MetaData{Name: "iris"},
								DeploymentSpec: &pbs.DeploymentSpec{Replicas: 2, MinReplicas: 1},
							},
							1, "server1",
							map[int32]*db.ReplicaStatus{
								1: {State: db.ModelReplicaState_Available}, 2: {State: db.ModelReplicaState_LoadFailed},
							}, db.ModelState_ScheduleFailed),
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
			models: []*db.Model{
				{
					Name: "iris",
					Versions: []*db.ModelVersion{
						util.NewTestModelVersion(
							&pbs.Model{
								Meta:           &pbs.MetaData{Name: "iris"},
								DeploymentSpec: &pbs.DeploymentSpec{Replicas: 2, MinReplicas: 1},
							},
							1, "server1",
							map[int32]*db.ReplicaStatus{
								1: {State: db.ModelReplicaState_Available}, 2: {State: db.ModelReplicaState_LoadFailed},
							}, db.ModelState_ScheduleFailed),
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
			models: []*db.Model{
				{
					Name: "iris",
					Versions: []*db.ModelVersion{
						util.NewTestModelVersion(
							&pbs.Model{
								Meta:           &pbs.MetaData{Name: "iris"},
								DeploymentSpec: &pbs.DeploymentSpec{Replicas: 2, MinReplicas: 1},
							},
							1, "server1",
							map[int32]*db.ReplicaStatus{
								1: {State: db.ModelReplicaState_Available}, 2: {State: db.ModelReplicaState_Available},
							}, db.ModelState_ModelAvailable),
						util.NewTestModelVersion(
							&pbs.Model{
								Meta:           &pbs.MetaData{Name: "iris"},
								DeploymentSpec: &pbs.DeploymentSpec{Replicas: 2, MinReplicas: 1},
							},
							2, "server1",
							map[int32]*db.ReplicaStatus{
								1: {State: db.ModelReplicaState_Available}, 2: {State: db.ModelReplicaState_Loading},
							}, db.ModelState_ModelProgressing),
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
			models: []*db.Model{
				{
					Name: "iris",
					Versions: []*db.ModelVersion{
						util.NewTestModelVersion(
							&pbs.Model{
								Meta:           &pbs.MetaData{Name: "iris"},
								DeploymentSpec: &pbs.DeploymentSpec{Replicas: 2, MinReplicas: 1},
							},
							1, "server1",
							map[int32]*db.ReplicaStatus{
								1: {State: db.ModelReplicaState_Available}, 2: {State: db.ModelReplicaState_Available},
							}, db.ModelState_ModelAvailable),
						util.NewTestModelVersion(
							&pbs.Model{
								Meta:           &pbs.MetaData{Name: "iris"},
								DeploymentSpec: &pbs.DeploymentSpec{Replicas: 2, MinReplicas: 1},
							},
							2, "server1",
							map[int32]*db.ReplicaStatus{
								1: {State: db.ModelReplicaState_Available}, 2: {State: db.ModelReplicaState_Available},
							}, db.ModelState_ModelAvailable),
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
			name:                "model does not exist in scheduler state",
			trigger:             pb.ModelScalingTriggerMessage_SCALE_UP,
			triggerModelName:    "iris",
			triggerModelVersion: 1,
			expectedReplicas:    2,
			isError:             true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			// Create storage instances
			modelStorage := store.NewInMemoryStorage[*db.Model]()

			// Populate storage with test data
			for _, model := range test.models {
				err := modelStorage.Insert(context.TODO(), model)
				g.Expect(err).To(BeNil())
			}

			model, _ := modelStorage.Get(context.TODO(), test.triggerModelName)
			if model != nil { // in the cases where the model is not in the scheduler state yet
				lastestModel := model.Latest()
				state := lastestModel.State
				state.Timestamp = timestamppb.New(test.lastUpdate)
				lastestModel.State = state
			} else {
				model = &db.Model{
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
		setupMock                     func(s *mock.MockScheduler)
	}
	tests := []test{
		{
			name: "simple",
			agents: []ag{
				{1, true},
				{2, true},
			},
			expectedAgentsCount:           2,
			expectedAgentsCountAfterClose: 0,
			setupMock: func(s *mock.MockScheduler) {
				s.EXPECT().ScheduleFailedModels().Return([]string{}, nil).MinTimes(2)
			},
		},
		{
			name: "simple - no close",
			agents: []ag{
				{1, true},
				{2, false},
			},
			expectedAgentsCount:           2,
			expectedAgentsCountAfterClose: 1,
			setupMock: func(s *mock.MockScheduler) {
				s.EXPECT().ScheduleFailedModels().Return([]string{}, nil).MinTimes(2)
			},
		},
		{
			name: "duplicates",
			agents: []ag{
				{1, true},
				{1, false},
			},
			expectedAgentsCount:           1,
			expectedAgentsCountAfterClose: 1,
			setupMock: func(s *mock.MockScheduler) {
				s.EXPECT().ScheduleFailedModels().Return([]string{}, nil).MinTimes(1)
			},
		},
		{
			name: "duplicates with all close",
			agents: []ag{
				{1, true},
				{1, true},
				{1, true},
			},
			expectedAgentsCount:           1,
			expectedAgentsCountAfterClose: 0,
			setupMock: func(s *mock.MockScheduler) {
				s.EXPECT().ScheduleFailedModels().Return([]string{}, nil).MinTimes(3)
			},
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
			ctrl := gomock.NewController(t)
			mockScheduler := mock.NewMockScheduler(ctrl)
			test.setupMock(mockScheduler)

			logger := log.New()
			eventHub, err := coordinator.NewEventHub(logger)
			g.Expect(err).To(BeNil())

			// Create storage instances
			modelStorage := store.NewInMemoryStorage[*db.Model]()
			serverStorage := store.NewInMemoryStorage[*db.Server]()
			ms := store.NewModelServerStore(logger, modelStorage, serverStorage, eventHub)

			server := NewAgentServer(logger, ms, mockScheduler, eventHub, false, tls.TLSOptions{})
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

			for count < maxCount {
				server.mutex.RLock()
				if len(server.agents) == test.expectedAgentsCount {
					server.mutex.RUnlock()
					break
				}
				server.mutex.RUnlock()
				time.Sleep(100 * time.Millisecond)
				count++
			}

			server.mutex.RLock()
			g.Expect(len(server.agents)).To(Equal(test.expectedAgentsCount))
			server.mutex.RUnlock()

			mu.Lock()
			for idx, s := range streams {
				go func(idx int, s *grpc.ClientConn) {
					if test.agents[idx].doClose {
						s.Close()
					}
				}(idx, s)
			}
			mu.Unlock()

			count = 0

			for count < maxCount {
				server.mutex.RLock()
				if len(server.agents) == test.expectedAgentsCountAfterClose {
					server.mutex.RUnlock()
					break
				}
				server.mutex.RUnlock()
				time.Sleep(100 * time.Millisecond)
				count++
			}

			server.mutex.RLock()
			g.Expect(len(server.agents)).To(Equal(test.expectedAgentsCountAfterClose))
			server.mutex.RUnlock()

			server.StopAgentStreams()
		})
	}
}
