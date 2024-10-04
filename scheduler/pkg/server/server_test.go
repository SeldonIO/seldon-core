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
	"sort"
	"sync/atomic"
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
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/synchroniser"
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

	createTestScheduler := func() (*SchedulerServer, *mockAgentHandler, *coordinator.EventHub) {
		logger := log.New()
		logger.SetLevel(log.WarnLevel)

		eventHub, err := coordinator.NewEventHub(logger)
		g.Expect(err).To(BeNil())

		schedulerStore := store.NewMemoryStore(logger, store.NewLocalSchedulerStore(), eventHub)
		experimentServer := experiment.NewExperimentServer(logger, eventHub, nil, nil)
		pipelineServer := pipeline.NewPipelineStore(logger, eventHub, schedulerStore)
		sync := synchroniser.NewSimpleSynchroniser(time.Duration(10 * time.Millisecond))
		scheduler := scheduler2.NewSimpleScheduler(
			logger,
			schedulerStore,
			scheduler2.DefaultSchedulerConfig(schedulerStore),
			sync,
		)
		s := NewSchedulerServer(logger, schedulerStore, experimentServer, pipelineServer, scheduler, eventHub, sync)
		sync.Signals(1)
		mockAgent := &mockAgentHandler{}

		return s, mockAgent, eventHub
	}

	smallMemory := uint64(100)
	largeMemory := uint64(2000)

	type test struct {
		name           string
		req            []*pba.AgentSubscribeRequest
		model          *pb.Model
		scheduleFailed bool
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
			scheduleFailed: false,
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
			scheduleFailed: true,
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
			scheduleFailed: true,
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
			scheduleFailed: true,
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
			scheduleFailed: false,
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
			scheduleFailed: false,
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
			scheduleFailed: true,
		}, // schedule to 2 replicas but 1 fails
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given
			s, _, h := createTestScheduler()
			for _, repReq := range test.req {
				err := s.modelStore.AddServerReplica(repReq)
				g.Expect(err).To(BeNil())
			}

			scheduledFailed := atomic.Bool{}

			// Subscribe to model events
			h.RegisterModelEventHandler(
				"handler-model",
				10,
				log.New(),
				func(event coordinator.ModelEventMsg) {
					if event.ModelName != test.model.Meta.Name {
						return
					}
					model, _ := s.modelStore.GetModel(event.ModelName)
					latest := model.GetLatest()
					if latest.ModelState().State == store.ScheduleFailed {
						scheduledFailed.Store(true)
					} else {
						scheduledFailed.Store(false)
					}
				},
			)

			// When
			lm := pb.LoadModelRequest{
				Model: test.model,
			}
			r, err := s.LoadModel(context.Background(), &lm)

			time.Sleep(100 * time.Millisecond)

			// Then
			g.Expect(r).ToNot(BeNil())
			g.Expect(err).To(BeNil())
			if test.scheduleFailed {
				g.Expect(scheduledFailed.Load()).To(BeTrueBecause("schedule failed"))
			} else {
				g.Expect(scheduledFailed.Load()).To(BeFalseBecause("schedule ok"))
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
		sync := synchroniser.NewSimpleSynchroniser(time.Duration(10 * time.Millisecond))
		scheduler := scheduler2.NewSimpleScheduler(logger,
			schedulerStore,
			scheduler2.DefaultSchedulerConfig(schedulerStore),
			sync,
		)
		s := NewSchedulerServer(logger, schedulerStore, experimentServer, pipelineServer, scheduler, eventHub, sync)
		sync.Signals(1)
		return s, mockAgent, eventHub
	}

	type test struct {
		name       string
		req        []*pba.AgentSubscribeRequest
		model      *pb.Model
		code       codes.Code
		modelState store.ModelState
	}
	modelName := "model1"
	smallMemory := uint64(100)
	tests := []test{
		{
			name: "Simple",
			req: []*pba.AgentSubscribeRequest{
				{ServerName: "server1", ReplicaIdx: 0, Shared: true, AvailableMemoryBytes: 1000,
					ReplicaConfig: &pba.ReplicaConfig{InferenceSvc: "server1", InferenceHttpPort: 1, Capabilities: []string{"sklearn"}}}},
			model:      &pb.Model{Meta: &pb.MetaData{Name: "model1"}, ModelSpec: &pb.ModelSpec{Uri: "gs://model", Requirements: []string{"sklearn"}, MemoryBytes: &smallMemory}, DeploymentSpec: &pb.DeploymentSpec{Replicas: 1}},
			code:       codes.OK,
			modelState: store.ModelTerminated,
		},
		{
			name: "Multiple",
			req: []*pba.AgentSubscribeRequest{
				{ServerName: "server1", ReplicaIdx: 0, Shared: true, AvailableMemoryBytes: 1000,
					ReplicaConfig: &pba.ReplicaConfig{InferenceSvc: "server1", InferenceHttpPort: 1, Capabilities: []string{"sklearn", "xgboost"}}}},
			model:      &pb.Model{Meta: &pb.MetaData{Name: "model1"}, ModelSpec: &pb.ModelSpec{Uri: "gs://model", Requirements: []string{"sklearn", "xgboost"}, MemoryBytes: &smallMemory}, DeploymentSpec: &pb.DeploymentSpec{Replicas: 1}},
			code:       codes.OK,
			modelState: store.ModelTerminated,
		},
		{
			name: "TwoReplicas",
			req: []*pba.AgentSubscribeRequest{
				{ServerName: "server1", ReplicaIdx: 0, Shared: true, AvailableMemoryBytes: 1000,
					ReplicaConfig: &pba.ReplicaConfig{InferenceSvc: "server1", InferenceHttpPort: 1, Capabilities: []string{"sklearn"}}},
				{ServerName: "server1", ReplicaIdx: 1, Shared: true, AvailableMemoryBytes: 1000,
					ReplicaConfig: &pba.ReplicaConfig{InferenceSvc: "server1", InferenceHttpPort: 1, Capabilities: []string{"sklearn"}}}},
			model:      &pb.Model{Meta: &pb.MetaData{Name: "model1"}, ModelSpec: &pb.ModelSpec{Uri: "gs://model", Requirements: []string{"sklearn"}, MemoryBytes: &smallMemory}, DeploymentSpec: &pb.DeploymentSpec{Replicas: 2}},
			code:       codes.OK,
			modelState: store.ModelTerminated,
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
				g.Expect(ms.GetLatest().ModelState().State).To(Equal(test.modelState))

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

func TestServerNotify(t *testing.T) {
	g := NewGomegaWithT(t)

	createTestScheduler := func() (*SchedulerServer, synchroniser.Synchroniser) {
		logger := log.New()
		log.SetLevel(log.DebugLevel)
		eventHub, err := coordinator.NewEventHub(logger)
		g.Expect(err).To(BeNil())
		schedulerStore := store.NewMemoryStore(logger, store.NewLocalSchedulerStore(), eventHub)
		sync := synchroniser.NewSimpleSynchroniser(time.Duration(10 * time.Millisecond))
		scheduler := scheduler2.NewSimpleScheduler(logger,
			schedulerStore,
			scheduler2.DefaultSchedulerConfig(schedulerStore),
			sync,
		)
		s := NewSchedulerServer(logger, schedulerStore, nil, nil, scheduler, eventHub, sync)
		return s, sync
	}

	type test struct {
		name                 string
		req                  *pb.ServerNotifyRequest
		expectedServerStates []*store.ServerSnapshot
		signalTriggered      bool
	}
	tests := []test{
		{
			name: "Initial sync",
			req: &pb.ServerNotifyRequest{
				Servers: []*pb.ServerNotify{
					{
						Name:             "server1",
						ExpectedReplicas: 2,
						Shared:           true,
					},
					{
						Name:             "server2",
						ExpectedReplicas: 3,
						Shared:           true,
					},
				},
				IsFirstSync: true,
			},
			expectedServerStates: []*store.ServerSnapshot{
				{
					Name:             "server1",
					ExpectedReplicas: 2,
					Shared:           true,
					Replicas:         map[int]*store.ServerReplica{},
				},
				{
					Name:             "server2",
					ExpectedReplicas: 3,
					Shared:           true,
					Replicas:         map[int]*store.ServerReplica{},
				},
			},
			signalTriggered: true,
		},
		{
			// this should not trigger sync.Signal()
			name: "normal sync",
			req: &pb.ServerNotifyRequest{
				Servers: []*pb.ServerNotify{
					{
						Name:             "server1",
						ExpectedReplicas: 2,
						Shared:           true,
					},
				},
				IsFirstSync: false,
			},
			expectedServerStates: []*store.ServerSnapshot{
				{
					Name:             "server1",
					ExpectedReplicas: 2,
					Shared:           true,
					Replicas:         map[int]*store.ServerReplica{},
				},
			},
			signalTriggered: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s, sync := createTestScheduler()

			_, _ = s.ServerNotify(context.Background(), test.req)

			time.Sleep(50 * time.Millisecond) // allow events to be processed

			actualServers, err := s.modelStore.GetServers(true, false)
			g.Expect(err).To(BeNil())
			sort.Slice(actualServers, func(i, j int) bool {
				return actualServers[i].Name < actualServers[j].Name
			})
			sort.Slice(test.expectedServerStates, func(i, j int) bool {
				return test.expectedServerStates[i].Name < test.expectedServerStates[j].Name
			})
			g.Expect(actualServers).To(Equal(test.expectedServerStates))

			g.Expect(sync.IsTriggered()).To(Equal(test.signalTriggered))
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

type stubControlPlaneServer struct {
	msgs      chan *pb.ControlPlaneResponse
	sleepTime time.Duration
	grpc.ServerStream
}

var _ pb.Scheduler_SubscribeControlPlaneServer = (*stubControlPlaneServer)(nil)

func newStubControlPlaneServer(capacity int, sleepTime time.Duration) *stubControlPlaneServer {
	return &stubControlPlaneServer{
		msgs:      make(chan *pb.ControlPlaneResponse, capacity),
		sleepTime: sleepTime,
	}
}

func (s *stubControlPlaneServer) Send(r *pb.ControlPlaneResponse) error {
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
