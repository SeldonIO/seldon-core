package server

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"
	pba "github.com/seldonio/seldon-core/scheduler/apis/mlops/agent"
	pb "github.com/seldonio/seldon-core/scheduler/apis/mlops/scheduler"
	scheduler2 "github.com/seldonio/seldon-core/scheduler/pkg/scheduler"
	"github.com/seldonio/seldon-core/scheduler/pkg/scheduler/filters"
	"github.com/seldonio/seldon-core/scheduler/pkg/scheduler/sorters"
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
	t.Logf("Started")
	g := NewGomegaWithT(t)

	createTestScheduler := func() (*SchedulerServer, *mockAgentHandler) {
		logger := log.New()
		log.SetLevel(log.DebugLevel)
		schedulerStore := store.NewMemoryStore(logger, store.NewLocalSchedulerStore())
		mockAgent := &mockAgentHandler{}
		scheduler := scheduler2.NewSimpleScheduler(logger,
			schedulerStore,
			[]scheduler2.ServerFilter{filters.SharingServerFilter{}},
			[]scheduler2.ReplicaFilter{filters.RequirementsReplicaFilter{}, filters.AvailableMemoryFilter{}},
			[]sorters.ServerSorter{},
			[]sorters.ReplicaSorter{sorters.ModelAlreadyLoadedSorter{}})
		s := NewSchedulerServer(logger, schedulerStore, scheduler, mockAgent)
		return s, mockAgent
	}

	type test struct {
		name  string
		req   []*pba.AgentSubscribeRequest
		model *pb.ModelDetails
		code  codes.Code
	}
	smallMemory := uint64(100)
	largeMemory := uint64(2000)
	tests := []test{
		{
			name: "Simple",
			req: []*pba.AgentSubscribeRequest{
				{ServerName: "server1", ReplicaIdx: 0, Shared: true,
					ReplicaConfig: &pba.ReplicaConfig{InferenceSvc: "server1", InferenceHttpPort: 1, MemoryBytes: 1000, AvailableMemoryBytes: 1000, Capabilities: []string{"sklearn"}}}},
			model: &pb.ModelDetails{Name: "model1", Uri: "gs://model", Requirements: []string{"sklearn"}, MemoryBytes: &smallMemory, Replicas: 1},
			code:  codes.OK},
		{
			name: "TooManyReplicas",
			req: []*pba.AgentSubscribeRequest{
				{ServerName: "server1", ReplicaIdx: 0, Shared: true,
					ReplicaConfig: &pba.ReplicaConfig{InferenceSvc: "server1", InferenceHttpPort: 1, MemoryBytes: 1000, AvailableMemoryBytes: 1000, Capabilities: []string{"sklearn"}}}},
			model: &pb.ModelDetails{Name: "model1", Uri: "gs://model", Requirements: []string{"sklearn"}, MemoryBytes: &smallMemory, Replicas: 2},
			code:  codes.FailedPrecondition},
		{
			name: "TooMuchMemory",
			req: []*pba.AgentSubscribeRequest{
				{ServerName: "server1", ReplicaIdx: 0, Shared: true,
					ReplicaConfig: &pba.ReplicaConfig{InferenceSvc: "server1", InferenceHttpPort: 1, MemoryBytes: 1000, AvailableMemoryBytes: 1000, Capabilities: []string{"sklearn"}}}},
			model: &pb.ModelDetails{Name: "model1", Uri: "gs://model", Requirements: []string{"sklearn"}, MemoryBytes: &largeMemory, Replicas: 1},
			code:  codes.FailedPrecondition},
		{
			name: "FailedRequirements",
			req: []*pba.AgentSubscribeRequest{
				{ServerName: "server1", ReplicaIdx: 0, Shared: true,
					ReplicaConfig: &pba.ReplicaConfig{InferenceSvc: "server1", InferenceHttpPort: 1, MemoryBytes: 1000, AvailableMemoryBytes: 1000, Capabilities: []string{"sklearn"}}}},
			model: &pb.ModelDetails{Name: "model1", Uri: "gs://model", Requirements: []string{"xgboost"}, MemoryBytes: &smallMemory, Replicas: 1},
			code:  codes.FailedPrecondition},
		{
			name: "MultipleRequirements",
			req: []*pba.AgentSubscribeRequest{
				{ServerName: "server1", ReplicaIdx: 0, Shared: true,
					ReplicaConfig: &pba.ReplicaConfig{InferenceSvc: "server1", InferenceHttpPort: 1, MemoryBytes: 1000, AvailableMemoryBytes: 1000, Capabilities: []string{"sklearn", "xgboost"}}}},
			model: &pb.ModelDetails{Name: "model1", Uri: "gs://model", Requirements: []string{"xgboost", "sklearn"}, MemoryBytes: &smallMemory, Replicas: 1},
			code:  codes.OK},
		{
			name: "TwoReplicas",
			req: []*pba.AgentSubscribeRequest{
				{ServerName: "server1", ReplicaIdx: 0, Shared: true,
					ReplicaConfig: &pba.ReplicaConfig{InferenceSvc: "server1", InferenceHttpPort: 1, MemoryBytes: 1000, AvailableMemoryBytes: 1000, Capabilities: []string{"sklearn"}}},
				{ServerName: "server1", ReplicaIdx: 1, Shared: true,
					ReplicaConfig: &pba.ReplicaConfig{InferenceSvc: "server1", InferenceHttpPort: 1, MemoryBytes: 1000, AvailableMemoryBytes: 1000, Capabilities: []string{"sklearn"}}}},
			model: &pb.ModelDetails{Name: "model1", Uri: "gs://model", Requirements: []string{"sklearn"}, MemoryBytes: &smallMemory, Replicas: 2},
			code:  codes.OK},
		{
			name: "TwoReplicasFail",
			req: []*pba.AgentSubscribeRequest{
				{ServerName: "server1", ReplicaIdx: 0, Shared: true,
					ReplicaConfig: &pba.ReplicaConfig{InferenceSvc: "server1", InferenceHttpPort: 1, MemoryBytes: 1000, AvailableMemoryBytes: 1000, Capabilities: []string{"sklearn"}}},
				{ServerName: "server1", ReplicaIdx: 1, Shared: true,
					ReplicaConfig: &pba.ReplicaConfig{InferenceSvc: "server1", InferenceHttpPort: 1, MemoryBytes: 1000, AvailableMemoryBytes: 1000, Capabilities: []string{"foo"}}}},
			model: &pb.ModelDetails{Name: "model1", Uri: "gs://model", Requirements: []string{"sklearn"}, MemoryBytes: &smallMemory, Replicas: 2},
			code:  codes.FailedPrecondition}, // schedule to 2 replicas but 1 fails
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s, mockAgent := createTestScheduler()
			for _, repReq := range test.req {
				err := s.store.AddServerReplica(repReq)
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
				g.Expect(mockAgent.numSyncs).To(Equal(1))
			}
		})
	}
}

func TestUnloadModel(t *testing.T) {
	t.Logf("Started")
	g := NewGomegaWithT(t)

	createTestScheduler := func() (*SchedulerServer, *mockAgentHandler) {
		logger := log.New()
		log.SetLevel(log.DebugLevel)
		schedulerStore := store.NewMemoryStore(logger, store.NewLocalSchedulerStore())
		mockAgent := &mockAgentHandler{}
		scheduler := scheduler2.NewSimpleScheduler(logger,
			schedulerStore,
			[]scheduler2.ServerFilter{filters.SharingServerFilter{}},
			[]scheduler2.ReplicaFilter{filters.RequirementsReplicaFilter{}, filters.AvailableMemoryFilter{}},
			[]sorters.ServerSorter{},
			[]sorters.ReplicaSorter{sorters.ModelAlreadyLoadedSorter{}})
		s := NewSchedulerServer(logger, schedulerStore, scheduler, mockAgent)
		return s, mockAgent
	}

	type test struct {
		name               string
		req                []*pba.AgentSubscribeRequest
		model              *pb.ModelDetails
		code               codes.Code
		modelReplicaStates map[int]store.ModelReplicaState
	}
	modelName := "model1"
	smallMemory := uint64(100)
	tests := []test{
		{
			name: "Simple",
			req: []*pba.AgentSubscribeRequest{
				{ServerName: "server1", ReplicaIdx: 0, Shared: true,
					ReplicaConfig: &pba.ReplicaConfig{InferenceSvc: "server1", InferenceHttpPort: 1, AvailableMemoryBytes: 1000, Capabilities: []string{"sklearn"}}}},
			model:              &pb.ModelDetails{Name: modelName, Uri: "gs://model", Requirements: []string{"sklearn"}, MemoryBytes: &smallMemory, Replicas: 1},
			code:               codes.OK,
			modelReplicaStates: map[int]store.ModelReplicaState{0: store.UnloadRequested},
		},
		{
			name: "Multiple",
			req: []*pba.AgentSubscribeRequest{
				{ServerName: "server1", ReplicaIdx: 0, Shared: true,
					ReplicaConfig: &pba.ReplicaConfig{InferenceSvc: "server1", InferenceHttpPort: 1, AvailableMemoryBytes: 1000, Capabilities: []string{"sklearn", "xgboost"}}}},
			model:              &pb.ModelDetails{Name: modelName, Uri: "gs://model", Requirements: []string{"xgboost", "sklearn"}, MemoryBytes: &smallMemory, Replicas: 1},
			code:               codes.OK,
			modelReplicaStates: map[int]store.ModelReplicaState{0: store.UnloadRequested},
		},
		{
			name: "TwoReplicas",
			req: []*pba.AgentSubscribeRequest{
				{ServerName: "server1", ReplicaIdx: 0, Shared: true,
					ReplicaConfig: &pba.ReplicaConfig{InferenceSvc: "server1", InferenceHttpPort: 1, AvailableMemoryBytes: 1000, Capabilities: []string{"sklearn"}}},
				{ServerName: "server1", ReplicaIdx: 1, Shared: true,
					ReplicaConfig: &pba.ReplicaConfig{InferenceSvc: "server1", InferenceHttpPort: 1, AvailableMemoryBytes: 1000, Capabilities: []string{"sklearn"}}}},
			model:              &pb.ModelDetails{Name: modelName, Uri: "gs://model", Requirements: []string{"sklearn"}, MemoryBytes: &smallMemory, Replicas: 2},
			code:               codes.OK,
			modelReplicaStates: map[int]store.ModelReplicaState{0: store.UnloadRequested, 1: store.UnloadRequested},
		},
		{
			name: "NotExist",
			req: []*pba.AgentSubscribeRequest{
				{ServerName: "server1", ReplicaIdx: 0, Shared: true,
					ReplicaConfig: &pba.ReplicaConfig{InferenceSvc: "server1", InferenceHttpPort: 1, AvailableMemoryBytes: 1000, Capabilities: []string{"sklearn"}}}},
			model: nil,
			code:  codes.FailedPrecondition},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s, mockAgent := createTestScheduler()
			for _, repReq := range test.req {
				err := s.store.AddServerReplica(repReq)
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
			rm := &pb.ModelReference{Name: modelName}
			r, err := s.UnloadModel(context.Background(), rm)
			if test.code != codes.OK {
				g.Expect(err).ToNot(BeNil())
				e, ok := status.FromError(err)
				g.Expect(ok).To(BeTrue())
				g.Expect(e.Code()).To(Equal(test.code))
			} else {
				g.Expect(err).To(BeNil())
				g.Expect(r).ToNot(BeNil())
				ms, err := s.store.GetModel(modelName)
				g.Expect(err).To(BeNil())
				g.Expect(mockAgent.numSyncs).To(Equal(2))
				for replicaIdx, state := range test.modelReplicaStates {
					g.Expect(ms.GetLatest().GetModelReplicaState(replicaIdx)).To(Equal(state))
				}
			}
		})
	}
}
