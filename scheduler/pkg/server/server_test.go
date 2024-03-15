/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package server

import (
	"context"
	"fmt"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pba "github.com/seldonio/seldon-core/apis/go/v2/mlops/agent"
	pb "github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/coordinator"
	scheduler2 "github.com/seldonio/seldon-core/scheduler/v2/pkg/scheduler"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store/experiment"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store/pipeline"
)

func stringPtr(s string) *string {
	return &s
}

type mockAgentHandler struct {
	numSyncs int
}

func (m *mockAgentHandler) SendAgentSync(modelName string) {
	m.numSyncs++
}

func TestLoadModel(t *testing.T) {
	g := NewGomegaWithT(t)

	createTestScheduler := func() (*SchedulerServer, *mockAgentHandler) {
		logger := log.New()
		logger.SetLevel(log.WarnLevel)

		eventHub, err := coordinator.NewEventHub(logger)
		g.Expect(err).To(BeNil())

		schedulerStore := store.NewMemoryStore(logger, store.NewLocalSchedulerStore(), eventHub)
		experimentServer := experiment.NewExperimentServer(logger, eventHub, nil, nil)
		pipelineServer := pipeline.NewPipelineStore(logger, eventHub, schedulerStore)

		scheduler := scheduler2.NewSimpleScheduler(
			logger,
			schedulerStore,
			scheduler2.DefaultSchedulerConfig(schedulerStore),
		)
		s := NewSchedulerServer(logger, schedulerStore, experimentServer, pipelineServer, scheduler, eventHub)
		mockAgent := &mockAgentHandler{}

		return s, mockAgent
	}

	smallMemory := uint64(100)
	largeMemory := uint64(2000)

	type test struct {
		name  string
		req   []*pba.AgentSubscribeRequest
		model *pb.Model
		code  codes.Code
	}

	tests := []test{
		{
			name: "Simple",
			req: []*pba.AgentSubscribeRequest{
				{
					ServerName:           "server1",
					ReplicaIdx:           0,
					Shared:               true,
					AvailableMemoryBytes: 1000,
					ReplicaConfig: &pba.ReplicaConfig{
						InferenceSvc:      "server1",
						InferenceHttpPort: 1,
						MemoryBytes:       1000,
						Capabilities:      []string{"sklearn"},
					},
				},
			},
			model: &pb.Model{
				Meta: &pb.MetaData{Name: "model1"},
				ModelSpec: &pb.ModelSpec{
					Uri:          "gs://model",
					Requirements: []string{"sklearn"},
					MemoryBytes:  &smallMemory,
				},
				DeploymentSpec: &pb.DeploymentSpec{Replicas: 1},
			},
			code: codes.OK,
		},
		{
			name: "TooManyReplicas",
			req: []*pba.AgentSubscribeRequest{
				{
					ServerName:           "server1",
					ReplicaIdx:           0,
					Shared:               true,
					AvailableMemoryBytes: 1000,
					ReplicaConfig: &pba.ReplicaConfig{
						InferenceSvc:      "server1",
						InferenceHttpPort: 1,
						MemoryBytes:       1000,
						Capabilities:      []string{"sklearn"},
					},
				},
			},
			model: &pb.Model{Meta: &pb.MetaData{Name: "model1"},
				ModelSpec: &pb.ModelSpec{
					Uri:          "gs://model",
					Requirements: []string{"sklearn"},
					MemoryBytes:  &smallMemory,
				},
				DeploymentSpec: &pb.DeploymentSpec{Replicas: 2},
			},
			code: codes.FailedPrecondition,
		},
		{
			name: "TooMuchMemory",
			req: []*pba.AgentSubscribeRequest{
				{
					ServerName:           "server1",
					ReplicaIdx:           0,
					Shared:               true,
					AvailableMemoryBytes: 1000,
					ReplicaConfig: &pba.ReplicaConfig{
						InferenceSvc:      "server1",
						InferenceHttpPort: 1,
						MemoryBytes:       1000,
						Capabilities:      []string{"sklearn"},
					},
				},
			},
			model: &pb.Model{
				Meta: &pb.MetaData{Name: "model1"},
				ModelSpec: &pb.ModelSpec{
					Uri:          "gs://model",
					Requirements: []string{"sklearn"},
					MemoryBytes:  &largeMemory,
				},
				DeploymentSpec: &pb.DeploymentSpec{Replicas: 1},
			},
			code: codes.FailedPrecondition,
		},
		{
			name: "FailedRequirements",
			req: []*pba.AgentSubscribeRequest{
				{
					ServerName:           "server1",
					ReplicaIdx:           0,
					Shared:               true,
					AvailableMemoryBytes: 1000,
					ReplicaConfig: &pba.ReplicaConfig{
						InferenceSvc:      "server1",
						InferenceHttpPort: 1,
						MemoryBytes:       1000,
						Capabilities:      []string{"sklearn"},
					},
				},
			},
			model: &pb.Model{
				Meta: &pb.MetaData{Name: "model1"},
				ModelSpec: &pb.ModelSpec{
					Uri:          "gs://model",
					Requirements: []string{"xgboost"},
					MemoryBytes:  &smallMemory,
				},
				DeploymentSpec: &pb.DeploymentSpec{Replicas: 1},
			},
			code: codes.FailedPrecondition,
		},
		{
			name: "MultipleRequirements",
			req: []*pba.AgentSubscribeRequest{
				{
					ServerName:           "server1",
					ReplicaIdx:           0,
					Shared:               true,
					AvailableMemoryBytes: 1000,
					ReplicaConfig: &pba.ReplicaConfig{
						InferenceSvc:      "server1",
						InferenceHttpPort: 1,
						MemoryBytes:       1000,
						Capabilities:      []string{"sklearn", "xgboost"},
					},
				},
			},
			model: &pb.Model{
				Meta: &pb.MetaData{Name: "model1"},
				ModelSpec: &pb.ModelSpec{
					Uri:          "gs://model",
					Requirements: []string{"sklearn", "xgboost"},
					MemoryBytes:  &smallMemory,
				},
				DeploymentSpec: &pb.DeploymentSpec{Replicas: 1},
			},
			code: codes.OK,
		},
		{
			name: "TwoReplicas",
			req: []*pba.AgentSubscribeRequest{
				{
					ServerName:           "server1",
					ReplicaIdx:           0,
					Shared:               true,
					AvailableMemoryBytes: 1000,
					ReplicaConfig: &pba.ReplicaConfig{
						InferenceSvc:      "server1",
						InferenceHttpPort: 1,
						MemoryBytes:       1000,
						Capabilities:      []string{"sklearn"},
					},
				},
				{
					ServerName:           "server1",
					ReplicaIdx:           1,
					Shared:               true,
					AvailableMemoryBytes: 1000,
					ReplicaConfig: &pba.ReplicaConfig{
						InferenceSvc:      "server1",
						InferenceHttpPort: 1,
						MemoryBytes:       1000,
						Capabilities:      []string{"sklearn"},
					},
				},
			},
			model: &pb.Model{
				Meta: &pb.MetaData{Name: "model1"},
				ModelSpec: &pb.ModelSpec{
					Uri:          "gs://model",
					Requirements: []string{"sklearn"},
					MemoryBytes:  &smallMemory,
				},
				DeploymentSpec: &pb.DeploymentSpec{Replicas: 2},
			},
			code: codes.OK,
		},
		{
			name: "TwoReplicasFail",
			req: []*pba.AgentSubscribeRequest{
				{
					ServerName:           "server1",
					ReplicaIdx:           0,
					Shared:               true,
					AvailableMemoryBytes: 1000,
					ReplicaConfig: &pba.ReplicaConfig{
						InferenceSvc:      "server1",
						InferenceHttpPort: 1,
						MemoryBytes:       1000,
						Capabilities:      []string{"sklearn"},
					},
				},
				{
					ServerName:           "server1",
					ReplicaIdx:           1,
					Shared:               true,
					AvailableMemoryBytes: 0,
					ReplicaConfig: &pba.ReplicaConfig{
						InferenceSvc:      "server1",
						InferenceHttpPort: 1,
						MemoryBytes:       1000,
						Capabilities:      []string{"foo"},
					},
				},
			},
			model: &pb.Model{
				Meta: &pb.MetaData{Name: "model1"},
				ModelSpec: &pb.ModelSpec{
					Uri:          "gs://model",
					Requirements: []string{"sklearn"},
					MemoryBytes:  &smallMemory,
				},
				DeploymentSpec: &pb.DeploymentSpec{Replicas: 2},
			},
			code: codes.FailedPrecondition,
		}, // schedule to 2 replicas but 1 fails
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given
			s, _ := createTestScheduler()
			for _, repReq := range test.req {
				err := s.modelStore.AddServerReplica(repReq)
				g.Expect(err).To(BeNil())
			}

			// When
			lm := pb.LoadModelRequest{
				Model: test.model,
			}
			r, err := s.LoadModel(context.Background(), &lm)

			// Then
			if test.code != codes.OK {
				g.Expect(err).ToNot(BeNil())
				e, ok := status.FromError(err)
				g.Expect(ok).To(BeTrue())
				g.Expect(e.Code()).To(Equal(test.code))
			} else {
				g.Expect(err).To(BeNil())
				g.Expect(r).ToNot(BeNil())
			}
		})
	}
}

func TestUnloadModel(t *testing.T) {
	g := NewGomegaWithT(t)

	createTestScheduler := func() (*SchedulerServer, *mockAgentHandler, *coordinator.EventHub) {
		logger := log.New()
		log.SetLevel(log.DebugLevel)
		eventHub, err := coordinator.NewEventHub(logger)
		g.Expect(err).To(BeNil())
		schedulerStore := store.NewMemoryStore(logger, store.NewLocalSchedulerStore(), eventHub)
		experimentServer := experiment.NewExperimentServer(logger, eventHub, nil, nil)
		pipelineServer := pipeline.NewPipelineStore(logger, eventHub, schedulerStore)
		mockAgent := &mockAgentHandler{}
		scheduler := scheduler2.NewSimpleScheduler(logger,
			schedulerStore,
			scheduler2.DefaultSchedulerConfig(schedulerStore))
		s := NewSchedulerServer(logger, schedulerStore, experimentServer, pipelineServer, scheduler, eventHub)
		return s, mockAgent, eventHub
	}

	type test struct {
		name               string
		req                []*pba.AgentSubscribeRequest
		model              *pb.Model
		code               codes.Code
		modelReplicaStates map[int]store.ModelReplicaState
	}
	modelName := "model1"
	smallMemory := uint64(100)
	tests := []test{
		{
			name: "Simple",
			req: []*pba.AgentSubscribeRequest{
				{ServerName: "server1", ReplicaIdx: 0, Shared: true, AvailableMemoryBytes: 1000,
					ReplicaConfig: &pba.ReplicaConfig{InferenceSvc: "server1", InferenceHttpPort: 1, Capabilities: []string{"sklearn"}}}},
			model:              &pb.Model{Meta: &pb.MetaData{Name: "model1"}, ModelSpec: &pb.ModelSpec{Uri: "gs://model", Requirements: []string{"sklearn"}, MemoryBytes: &smallMemory}, DeploymentSpec: &pb.DeploymentSpec{Replicas: 1}},
			code:               codes.OK,
			modelReplicaStates: map[int]store.ModelReplicaState{0: store.UnloadEnvoyRequested},
		},
		{
			name: "Multiple",
			req: []*pba.AgentSubscribeRequest{
				{ServerName: "server1", ReplicaIdx: 0, Shared: true, AvailableMemoryBytes: 1000,
					ReplicaConfig: &pba.ReplicaConfig{InferenceSvc: "server1", InferenceHttpPort: 1, Capabilities: []string{"sklearn", "xgboost"}}}},
			model:              &pb.Model{Meta: &pb.MetaData{Name: "model1"}, ModelSpec: &pb.ModelSpec{Uri: "gs://model", Requirements: []string{"sklearn", "xgboost"}, MemoryBytes: &smallMemory}, DeploymentSpec: &pb.DeploymentSpec{Replicas: 1}},
			code:               codes.OK,
			modelReplicaStates: map[int]store.ModelReplicaState{0: store.UnloadEnvoyRequested},
		},
		{
			name: "TwoReplicas",
			req: []*pba.AgentSubscribeRequest{
				{ServerName: "server1", ReplicaIdx: 0, Shared: true, AvailableMemoryBytes: 1000,
					ReplicaConfig: &pba.ReplicaConfig{InferenceSvc: "server1", InferenceHttpPort: 1, Capabilities: []string{"sklearn"}}},
				{ServerName: "server1", ReplicaIdx: 1, Shared: true, AvailableMemoryBytes: 1000,
					ReplicaConfig: &pba.ReplicaConfig{InferenceSvc: "server1", InferenceHttpPort: 1, Capabilities: []string{"sklearn"}}}},
			model:              &pb.Model{Meta: &pb.MetaData{Name: "model1"}, ModelSpec: &pb.ModelSpec{Uri: "gs://model", Requirements: []string{"sklearn"}, MemoryBytes: &smallMemory}, DeploymentSpec: &pb.DeploymentSpec{Replicas: 2}},
			code:               codes.OK,
			modelReplicaStates: map[int]store.ModelReplicaState{0: store.UnloadEnvoyRequested, 1: store.UnloadEnvoyRequested},
		},
		{
			name: "NotExist",
			req: []*pba.AgentSubscribeRequest{
				{ServerName: "server1", ReplicaIdx: 0, Shared: true, AvailableMemoryBytes: 1000,
					ReplicaConfig: &pba.ReplicaConfig{InferenceSvc: "server1", InferenceHttpPort: 1, Capabilities: []string{"sklearn"}}}},
			model: nil,
			code:  codes.FailedPrecondition},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s, _, _ := createTestScheduler()
			defer s.StopSendModelEvents()
			defer s.StopSendServerEvents()

			for _, repReq := range test.req {
				err := s.modelStore.AddServerReplica(repReq)
				g.Expect(err).To(BeNil())
			}

			if test.model != nil {
				lm := pb.LoadModelRequest{
					Model: test.model,
				}
				r, err := s.LoadModel(context.Background(), &lm)
				g.Expect(err).To(BeNil())
				g.Expect(r).ToNot(BeNil())
			}
			rm := &pb.UnloadModelRequest{Model: &pb.ModelReference{Name: modelName}}
			r, err := s.UnloadModel(context.Background(), rm)
			if test.code != codes.OK {
				g.Expect(err).ToNot(BeNil())
				e, ok := status.FromError(err)
				g.Expect(ok).To(BeTrue())
				g.Expect(e.Code()).To(Equal(test.code))
			} else {
				g.Expect(err).To(BeNil())
				g.Expect(r).ToNot(BeNil())
				ms, err := s.modelStore.GetModel(modelName)
				g.Expect(err).To(BeNil())
				for replicaIdx, state := range test.modelReplicaStates {
					g.Expect(ms.GetLatest().GetModelReplicaState(replicaIdx)).To(Equal(state))
				}
			}
		})
	}
}

func TestLoadPipeline(t *testing.T) {
	g := NewGomegaWithT(t)
	type test struct {
		name   string
		req    *pb.LoadPipelineRequest
		server *SchedulerServer
		err    bool
	}

	tests := []test{
		{
			name: "pipeline with no output",
			req: &pb.LoadPipelineRequest{
				Pipeline: &pb.Pipeline{
					Name: "p1",
					Steps: []*pb.PipelineStep{
						{
							Name: "a",
						},
						{
							Name:   "b",
							Inputs: []string{"a.inputs"},
						},
					},
				},
			},
			server: &SchedulerServer{pipelineHandler: pipeline.NewPipelineStore(log.New(), nil, nil)},
		},
		{
			name: "pipeline with output",
			req: &pb.LoadPipelineRequest{
				Pipeline: &pb.Pipeline{
					Name: "p1",
					Steps: []*pb.PipelineStep{
						{
							Name: "a",
						},
						{
							Name:   "b",
							Inputs: []string{"a.inputs"},
						},
					},
					Output: &pb.PipelineOutput{
						Steps: []string{"b.outputs"},
					},
				},
			},
			server: &SchedulerServer{pipelineHandler: pipeline.NewPipelineStore(log.New(), nil, nil)},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := test.server.LoadPipeline(context.Background(), test.req)
			if test.err {
				g.Expect(err).ToNot(BeNil())
			} else {
				g.Expect(err).To(BeNil())
			}
		})
	}

}

func TestUnloadPipeline(t *testing.T) {
	g := NewGomegaWithT(t)
	type test struct {
		name      string
		loadReq   *pb.LoadPipelineRequest
		unloadReq *pb.UnloadPipelineRequest
		server    *SchedulerServer
		err       bool
	}

	tests := []test{
		{
			name:      "pipeline does not exist",
			unloadReq: &pb.UnloadPipelineRequest{Name: "foo"},
			server:    &SchedulerServer{pipelineHandler: pipeline.NewPipelineStore(log.New(), nil, nil)},
			err:       true,
		},
		{
			name: "pipeline removed",
			loadReq: &pb.LoadPipelineRequest{
				Pipeline: &pb.Pipeline{
					Name: "foo",
					Steps: []*pb.PipelineStep{
						{
							Name: "a",
						},
						{
							Name:   "b",
							Inputs: []string{"a.outputs"},
						},
					},
				},
			},
			unloadReq: &pb.UnloadPipelineRequest{Name: "foo"},
			server:    &SchedulerServer{pipelineHandler: pipeline.NewPipelineStore(log.New(), nil, nil)},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			path := fmt.Sprintf("%s/db", t.TempDir())
			_ = test.server.pipelineHandler.(*pipeline.PipelineStore).InitialiseOrRestoreDB(path)
			if test.loadReq != nil {
				err := test.server.pipelineHandler.AddPipeline(test.loadReq.Pipeline)
				g.Expect(err).To(BeNil())
			}
			_, err := test.server.UnloadPipeline(context.Background(), test.unloadReq)
			if test.err {
				g.Expect(err).ToNot(BeNil())
			} else {
				g.Expect(err).To(BeNil())
			}
		})
	}

}

func TestPipelineStatus(t *testing.T) {
	g := NewGomegaWithT(t)
	type test struct {
		name      string
		loadReq   *pb.LoadPipelineRequest
		statusReq *pb.PipelineStatusRequest
		statusRes *pb.PipelineStatusResponse
		server    *SchedulerServer
		err       bool
	}

	tests := []test{
		{
			name:      "pipeline does not exist",
			statusReq: &pb.PipelineStatusRequest{Name: stringPtr("foo")},
			server: &SchedulerServer{
				pipelineHandler: pipeline.NewPipelineStore(log.New(), nil, nil),
				logger:          log.New(),
			},
			err: true,
		},
		{
			name: "pipeline status",
			loadReq: &pb.LoadPipelineRequest{
				Pipeline: &pb.Pipeline{
					Name:    "foo",
					Version: 1,
					Uid:     "x",
					Steps: []*pb.PipelineStep{
						{
							Name: "a",
						},
						{
							Name:   "b",
							Inputs: []string{"a.outputs"},
						},
					},
				},
			},
			statusReq: &pb.PipelineStatusRequest{Name: stringPtr("foo")},
			statusRes: &pb.PipelineStatusResponse{
				PipelineName: "foo",
				Versions: []*pb.PipelineWithState{
					{
						Pipeline: &pb.Pipeline{
							Name:    "foo",
							Version: 1,
							Uid:     "x",
							Steps: []*pb.PipelineStep{
								{
									Name: "a",
								},
								{
									Name:   "b",
									Inputs: []string{"a.outputs"},
								},
							},
						},
						State: &pb.PipelineVersionState{
							PipelineVersion: 1,
							Status:          pb.PipelineVersionState_PipelineCreate,
							ModelsReady:     false,
						},
					},
				},
			},
			server: &SchedulerServer{
				pipelineHandler: pipeline.NewPipelineStore(log.New(), nil, nil),
				logger:          log.New(),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.loadReq != nil {
				err := test.server.pipelineHandler.AddPipeline(test.loadReq.Pipeline)
				g.Expect(err).To(BeNil())
			}

			stream := newStubPipelineStatusServer(1, 1*time.Millisecond)
			err := test.server.PipelineStatus(test.statusReq, stream)
			if test.err {
				g.Expect(err).ToNot(BeNil())
			} else {
				g.Expect(err).To(BeNil())

				var psr *pb.PipelineStatusResponse
				select {
				case next := <-stream.msgs:
					psr = next
				default:
					t.Fail()
				}

				g.Expect(psr).ToNot(BeNil())
				// clear timestamps before checking equality
				for _, pv := range psr.Versions {
					pv.State.LastChangeTimestamp = nil
				}
				g.Expect(psr).To(Equal(test.statusRes))
			}
		})
	}
}

type stubPipelineStatusServer struct {
	msgs      chan *pb.PipelineStatusResponse
	sleepTime time.Duration
	grpc.ServerStream
}

var _ pb.Scheduler_PipelineStatusServer = (*stubPipelineStatusServer)(nil)

func newStubPipelineStatusServer(capacity int, sleepTime time.Duration) *stubPipelineStatusServer {
	return &stubPipelineStatusServer{
		msgs:      make(chan *pb.PipelineStatusResponse, capacity),
		sleepTime: sleepTime,
	}
}

func (s *stubPipelineStatusServer) Send(r *pb.PipelineStatusResponse) error {
	time.Sleep(s.sleepTime)
	s.msgs <- r
	return nil
}

type stubModelStatusServer struct {
	msgs      chan *pb.ModelStatusResponse
	sleepTime time.Duration
	grpc.ServerStream
}

var _ pb.Scheduler_ModelStatusServer = (*stubModelStatusServer)(nil)

func newStubModelStatusServer(capacity int, sleepTime time.Duration) *stubModelStatusServer {
	return &stubModelStatusServer{
		msgs:      make(chan *pb.ModelStatusResponse, capacity),
		sleepTime: sleepTime,
	}
}

func (s *stubModelStatusServer) Send(r *pb.ModelStatusResponse) error {
	time.Sleep(s.sleepTime)
	s.msgs <- r
	return nil
}

type stubServerStatusServer struct {
	msgs      chan *pb.ServerStatusResponse
	sleepTime time.Duration
	grpc.ServerStream
}

var _ pb.Scheduler_ServerStatusServer = (*stubServerStatusServer)(nil)

func newStubServerStatusServer(capacity int, sleepTime time.Duration) *stubServerStatusServer {
	return &stubServerStatusServer{
		msgs:      make(chan *pb.ServerStatusResponse, capacity),
		sleepTime: sleepTime,
	}
}

func (s *stubServerStatusServer) Send(r *pb.ServerStatusResponse) error {
	time.Sleep(s.sleepTime)
	s.msgs <- r
	return nil
}

type stubExperimentStatusServer struct {
	msgs      chan *pb.ExperimentStatusResponse
	sleepTime time.Duration
	grpc.ServerStream
}

var _ pb.Scheduler_ExperimentStatusServer = (*stubExperimentStatusServer)(nil)

func newStubExperimentStatusServer(capacity int, sleepTime time.Duration) *stubExperimentStatusServer {
	return &stubExperimentStatusServer{
		msgs:      make(chan *pb.ExperimentStatusResponse, capacity),
		sleepTime: sleepTime,
	}
}

func (s *stubExperimentStatusServer) Send(r *pb.ExperimentStatusResponse) error {
	time.Sleep(s.sleepTime)
	s.msgs <- r
	return nil
}
