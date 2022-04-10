package server

import (
	"context"
	"testing"

	"github.com/seldonio/seldon-core/scheduler/pkg/store/pipeline"

	"github.com/seldonio/seldon-core/scheduler/pkg/store/experiment"

	"github.com/seldonio/seldon-core/scheduler/pkg/coordinator"

	. "github.com/onsi/gomega"
	pba "github.com/seldonio/seldon-core/scheduler/apis/mlops/agent"
	pb "github.com/seldonio/seldon-core/scheduler/apis/mlops/scheduler"
	scheduler2 "github.com/seldonio/seldon-core/scheduler/pkg/scheduler"
	"github.com/seldonio/seldon-core/scheduler/pkg/store"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

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
		log.SetLevel(log.DebugLevel)
		eventHub, err := coordinator.NewEventHub(logger)
		g.Expect(err).To(BeNil())
		schedulerStore := store.NewMemoryStore(logger, store.NewLocalSchedulerStore(), eventHub)
		experimentServer := experiment.NewExperimentServer(logger, eventHub)
		pipelineServer := pipeline.NewPipelineStore(logger, eventHub)
		mockAgent := &mockAgentHandler{}
		scheduler := scheduler2.NewSimpleScheduler(logger,
			schedulerStore,
			scheduler2.DefaultSchedulerConfig())
		s := NewSchedulerServer(logger, schedulerStore, experimentServer, pipelineServer, scheduler, eventHub)
		return s, mockAgent
	}

	type test struct {
		name  string
		req   []*pba.AgentSubscribeRequest
		model *pb.Model
		code  codes.Code
	}
	smallMemory := uint64(100)
	largeMemory := uint64(2000)
	tests := []test{
		{
			name: "Simple",
			req: []*pba.AgentSubscribeRequest{
				{ServerName: "server1", ReplicaIdx: 0, Shared: true, AvailableMemoryBytes: 1000,
					ReplicaConfig: &pba.ReplicaConfig{InferenceSvc: "server1", InferenceHttpPort: 1, MemoryBytes: 1000, Capabilities: []string{"sklearn"}}}},
			model: &pb.Model{Meta: &pb.MetaData{Name: "model1"}, ModelSpec: &pb.ModelSpec{Uri: "gs://model", Requirements: []string{"sklearn"}, MemoryBytes: &smallMemory}, DeploymentSpec: &pb.DeploymentSpec{Replicas: 1}},
			code:  codes.OK},
		{
			name: "TooManyReplicas",
			req: []*pba.AgentSubscribeRequest{
				{ServerName: "server1", ReplicaIdx: 0, Shared: true, AvailableMemoryBytes: 1000,
					ReplicaConfig: &pba.ReplicaConfig{InferenceSvc: "server1", InferenceHttpPort: 1, MemoryBytes: 1000, Capabilities: []string{"sklearn"}}}},
			model: &pb.Model{Meta: &pb.MetaData{Name: "model1"}, ModelSpec: &pb.ModelSpec{Uri: "gs://model", Requirements: []string{"sklearn"}, MemoryBytes: &smallMemory}, DeploymentSpec: &pb.DeploymentSpec{Replicas: 2}},
			code:  codes.FailedPrecondition},
		{
			name: "TooMuchMemory",
			req: []*pba.AgentSubscribeRequest{
				{ServerName: "server1", ReplicaIdx: 0, Shared: true, AvailableMemoryBytes: 1000,
					ReplicaConfig: &pba.ReplicaConfig{InferenceSvc: "server1", InferenceHttpPort: 1, MemoryBytes: 1000, Capabilities: []string{"sklearn"}}}},
			model: &pb.Model{Meta: &pb.MetaData{Name: "model1"}, ModelSpec: &pb.ModelSpec{Uri: "gs://model", Requirements: []string{"sklearn"}, MemoryBytes: &largeMemory}, DeploymentSpec: &pb.DeploymentSpec{Replicas: 1}},
			code:  codes.FailedPrecondition},
		{
			name: "FailedRequirements",
			req: []*pba.AgentSubscribeRequest{
				{ServerName: "server1", ReplicaIdx: 0, Shared: true, AvailableMemoryBytes: 1000,
					ReplicaConfig: &pba.ReplicaConfig{InferenceSvc: "server1", InferenceHttpPort: 1, MemoryBytes: 1000, Capabilities: []string{"sklearn"}}}},
			model: &pb.Model{Meta: &pb.MetaData{Name: "model1"}, ModelSpec: &pb.ModelSpec{Uri: "gs://model", Requirements: []string{"xgboost"}, MemoryBytes: &smallMemory}, DeploymentSpec: &pb.DeploymentSpec{Replicas: 1}},
			code:  codes.FailedPrecondition},
		{
			name: "MultipleRequirements",
			req: []*pba.AgentSubscribeRequest{
				{ServerName: "server1", ReplicaIdx: 0, Shared: true, AvailableMemoryBytes: 1000,
					ReplicaConfig: &pba.ReplicaConfig{InferenceSvc: "server1", InferenceHttpPort: 1, MemoryBytes: 1000, Capabilities: []string{"sklearn", "xgboost"}}}},
			model: &pb.Model{Meta: &pb.MetaData{Name: "model1"}, ModelSpec: &pb.ModelSpec{Uri: "gs://model", Requirements: []string{"sklearn", "xgboost"}, MemoryBytes: &smallMemory}, DeploymentSpec: &pb.DeploymentSpec{Replicas: 1}},
			code:  codes.OK},
		{
			name: "TwoReplicas",
			req: []*pba.AgentSubscribeRequest{
				{ServerName: "server1", ReplicaIdx: 0, Shared: true, AvailableMemoryBytes: 1000,
					ReplicaConfig: &pba.ReplicaConfig{InferenceSvc: "server1", InferenceHttpPort: 1, MemoryBytes: 1000, Capabilities: []string{"sklearn"}}},
				{ServerName: "server1", ReplicaIdx: 1, Shared: true, AvailableMemoryBytes: 1000,
					ReplicaConfig: &pba.ReplicaConfig{InferenceSvc: "server1", InferenceHttpPort: 1, MemoryBytes: 1000, Capabilities: []string{"sklearn"}}}},
			model: &pb.Model{Meta: &pb.MetaData{Name: "model1"}, ModelSpec: &pb.ModelSpec{Uri: "gs://model", Requirements: []string{"sklearn"}, MemoryBytes: &smallMemory}, DeploymentSpec: &pb.DeploymentSpec{Replicas: 2}},
			code:  codes.OK},
		{
			name: "TwoReplicasFail",
			req: []*pba.AgentSubscribeRequest{
				{ServerName: "server1", ReplicaIdx: 0, Shared: true, AvailableMemoryBytes: 1000,
					ReplicaConfig: &pba.ReplicaConfig{InferenceSvc: "server1", InferenceHttpPort: 1, MemoryBytes: 1000, Capabilities: []string{"sklearn"}}},
				{ServerName: "server1", ReplicaIdx: 1, Shared: true, AvailableMemoryBytes: 1000,
					ReplicaConfig: &pba.ReplicaConfig{InferenceSvc: "server1", InferenceHttpPort: 1, MemoryBytes: 1000, Capabilities: []string{"foo"}}}},
			model: &pb.Model{Meta: &pb.MetaData{Name: "model1"}, ModelSpec: &pb.ModelSpec{Uri: "gs://model", Requirements: []string{"sklearn"}, MemoryBytes: &smallMemory}, DeploymentSpec: &pb.DeploymentSpec{Replicas: 2}},
			code:  codes.FailedPrecondition}, // schedule to 2 replicas but 1 fails
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s, _ := createTestScheduler()
			for _, repReq := range test.req {
				err := s.modelStore.AddServerReplica(repReq)
				g.Expect(err).To(BeNil())
			}
			lm := pb.LoadModelRequest{
				Model: test.model,
			}
			r, err := s.LoadModel(context.Background(), &lm)
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
		experimentServer := experiment.NewExperimentServer(logger, eventHub)
		pipelineServer := pipeline.NewPipelineStore(logger, eventHub)
		mockAgent := &mockAgentHandler{}
		scheduler := scheduler2.NewSimpleScheduler(logger,
			schedulerStore,
			scheduler2.DefaultSchedulerConfig())
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
			modelReplicaStates: map[int]store.ModelReplicaState{0: store.UnloadRequested},
		},
		{
			name: "Multiple",
			req: []*pba.AgentSubscribeRequest{
				{ServerName: "server1", ReplicaIdx: 0, Shared: true, AvailableMemoryBytes: 1000,
					ReplicaConfig: &pba.ReplicaConfig{InferenceSvc: "server1", InferenceHttpPort: 1, Capabilities: []string{"sklearn", "xgboost"}}}},
			model:              &pb.Model{Meta: &pb.MetaData{Name: "model1"}, ModelSpec: &pb.ModelSpec{Uri: "gs://model", Requirements: []string{"sklearn", "xgboost"}, MemoryBytes: &smallMemory}, DeploymentSpec: &pb.DeploymentSpec{Replicas: 1}},
			code:               codes.OK,
			modelReplicaStates: map[int]store.ModelReplicaState{0: store.UnloadRequested},
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
			modelReplicaStates: map[int]store.ModelReplicaState{0: store.UnloadRequested, 1: store.UnloadRequested},
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
			server: &SchedulerServer{pipelineHandler: pipeline.NewPipelineStore(log.New(), nil)},
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
			server: &SchedulerServer{pipelineHandler: pipeline.NewPipelineStore(log.New(), nil)},
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
			server:    &SchedulerServer{pipelineHandler: pipeline.NewPipelineStore(log.New(), nil)},
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
			server:    &SchedulerServer{pipelineHandler: pipeline.NewPipelineStore(log.New(), nil)},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
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
			statusReq: &pb.PipelineStatusRequest{Name: "foo"},
			server:    &SchedulerServer{pipelineHandler: pipeline.NewPipelineStore(log.New(), nil)},
			err:       true,
		},
		{
			name: "pipeline status",
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
			statusReq: &pb.PipelineStatusRequest{Name: "foo"},
			statusRes: &pb.PipelineStatusResponse{
				PipelineName: "foo",
				Versions: []*pb.PipelineWithState{
					{
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
						State: &pb.PipelineVersionState{
							PipelineVersion: 1,
							Status:          pb.PipelineVersionState_PipelineCreate,
						},
					},
				},
			},
			server: &SchedulerServer{pipelineHandler: pipeline.NewPipelineStore(log.New(), nil)},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.loadReq != nil {
				err := test.server.pipelineHandler.AddPipeline(test.loadReq.Pipeline)
				g.Expect(err).To(BeNil())
			}
			psr, err := test.server.PipelineStatus(context.Background(), test.statusReq)
			if test.err {
				g.Expect(err).ToNot(BeNil())
			} else {
				g.Expect(err).To(BeNil())
				// clear timestamps before checking equality
				for _, pv := range psr.Versions {
					pv.State.LastChangeTimestamp = nil
				}
				g.Expect(psr).To(Equal(test.statusRes))
			}
		})
	}

}
