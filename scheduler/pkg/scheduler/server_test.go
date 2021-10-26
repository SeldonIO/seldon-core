package scheduler

import (
	"context"
	. "github.com/onsi/gomega"
	pba "github.com/seldonio/seldon-core/scheduler/apis/mlops/agent"
	pb "github.com/seldonio/seldon-core/scheduler/apis/mlops/scheduler"
	"github.com/seldonio/seldon-core/scheduler/pkg/store"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"testing"
)

func createTestScheduler() (*SchedulerServer){
	logger := log.New()
	log.SetLevel(log.DebugLevel)
	// Create a cache
	agentChan := make(chan string, 10)
	envoyChan := make(chan string, 10)
	schedulerStore := store.NewMemoryScheduler(logger, agentChan, envoyChan)
	s := NewScheduler(logger, schedulerStore)
	return s
}


func TestAddRemoveServerReplica(t *testing.T) {
	t.Logf("Started")
	g := NewGomegaWithT(t)
	s := createTestScheduler()

	err := s.store.UpdateServerReplica(&pba.AgentSubscribeRequest{ServerName:
		"server1", ReplicaIdx: 0, ReplicaConfig: &pba.ReplicaConfig{InferenceSvc: "server1", InferencePort: 1, Memory: 1000, Capabilities: []string{"sklearn"}}})
	g.Expect(err).To(BeNil())
	server, err := s.store.GetServer("server1")
	g.Expect(err).To(BeNil())
	g.Expect(server).ToNot(BeNil())
	err = s.store.RemoveServerReplicaAndRedeployModels("server1",0)
	g.Expect(err).To(BeNil())
}


func TestLoadModel(t *testing.T) {
	t.Logf("Started")
	g := NewGomegaWithT(t)

	type test struct {
		req []*pba.AgentSubscribeRequest
		model *pb.ModelDetails
		code codes.Code
	}
	smallMemory := uint64(100)
	largeMemory := uint64(2000)
	tests := []test {
		{req: []*pba.AgentSubscribeRequest{{ServerName: "server1", ReplicaIdx: 0,
			ReplicaConfig: &pba.ReplicaConfig{InferenceSvc: "server1", InferencePort: 1, Memory: 1000, AvailableMemory: 1000, Capabilities: []string{"sklearn"}}}},
			model: &pb.ModelDetails{Name: "model1", Uri: "gs://model", Requirements: []string{"sklearn"}, Memory:&smallMemory, Replicas: 1},
			code: codes.OK},
		{req: []*pba.AgentSubscribeRequest{{ServerName: "server1", ReplicaIdx: 0,
			ReplicaConfig: &pba.ReplicaConfig{InferenceSvc: "server1", InferencePort: 1, Memory: 1000, AvailableMemory: 1000, Capabilities: []string{"sklearn"}}}},
			model: &pb.ModelDetails{Name: "model1", Uri: "gs://model", Requirements: []string{"sklearn"}, Memory:&smallMemory, Replicas: 2},
			code: codes.FailedPrecondition}, // ask for too many replicas
		{req: []*pba.AgentSubscribeRequest{{ServerName: "server1", ReplicaIdx: 0,
			ReplicaConfig: &pba.ReplicaConfig{InferenceSvc: "server1", InferencePort: 1, Memory: 1000, AvailableMemory: 1000, Capabilities: []string{"sklearn"}}}},
			model: &pb.ModelDetails{Name: "model1", Uri: "gs://model", Requirements: []string{"sklearn"}, Memory:&largeMemory, Replicas: 1},
			code: codes.FailedPrecondition}, // ask for too much memory
		{req: []*pba.AgentSubscribeRequest{{ServerName: "server1", ReplicaIdx: 0,
			ReplicaConfig: &pba.ReplicaConfig{InferenceSvc: "server1", InferencePort: 1, Memory: 1000, AvailableMemory: 1000, Capabilities: []string{"sklearn"}}}},
			model: &pb.ModelDetails{Name: "model1", Uri: "gs://model", Requirements: []string{"xgboost"}, Memory:&smallMemory, Replicas: 1},
			code: codes.FailedPrecondition}, // unable to find requirements
		{req: []*pba.AgentSubscribeRequest{{ServerName: "server1", ReplicaIdx: 0,
			ReplicaConfig: &pba.ReplicaConfig{InferenceSvc: "server1", InferencePort: 1, Memory: 1000, AvailableMemory: 1000, Capabilities: []string{"sklearn","xgboost"}}}},
			model: &pb.ModelDetails{Name: "model1", Uri: "gs://model", Requirements: []string{"xgboost","sklearn"}, Memory:&smallMemory, Replicas: 1},
			code: codes.OK}, // multiple requirements
		{req: []*pba.AgentSubscribeRequest{{ServerName: "server1", ReplicaIdx: 0,
			ReplicaConfig: &pba.ReplicaConfig{InferenceSvc: "server1", InferencePort: 1, Memory: 1000, AvailableMemory: 1000, Capabilities: []string{"sklearn"}}},
			{ServerName: "server1", ReplicaIdx: 1,
				ReplicaConfig: &pba.ReplicaConfig{InferenceSvc: "server1", InferencePort: 1, Memory: 1000, AvailableMemory: 1000, Capabilities: []string{"sklearn"}}}},
			model: &pb.ModelDetails{Name: "model1", Uri: "gs://model", Requirements: []string{"sklearn"}, Memory:&smallMemory, Replicas: 2},
			code: codes.OK}, // schedule to 2 replicas
		{req: []*pba.AgentSubscribeRequest{{ServerName: "server1", ReplicaIdx: 0,
			ReplicaConfig: &pba.ReplicaConfig{InferenceSvc: "server1", InferencePort: 1, Memory: 1000, AvailableMemory: 1000, Capabilities: []string{"sklearn"}}},
			{ServerName: "server1", ReplicaIdx: 1,
				ReplicaConfig: &pba.ReplicaConfig{InferenceSvc: "server1", InferencePort: 1, Memory: 1000, AvailableMemory: 1000, Capabilities: []string{"foo"}}}},
			model: &pb.ModelDetails{Name: "model1", Uri: "gs://model", Requirements: []string{"sklearn"}, Memory:&smallMemory, Replicas: 2},
			code: codes.FailedPrecondition}, // schedule to 2 replicas but 1 fails
	}
	for tidx,test := range tests {
		t.Logf("start test %d",tidx)
		s := createTestScheduler()
		for _, repReq := range test.req {
			err := s.store.UpdateServerReplica(repReq) // Create server and replicas
			g.Expect(err).To(BeNil())
		}
		lm := pb.LoadModelRequest{
			Model: test.model,
		}
		r,err := s.LoadModel(context.Background(), &lm)
		if test.code != codes.OK {
			g.Expect(err).ToNot(BeNil())
			e, ok := status.FromError(err)
			g.Expect(ok).To(BeTrue())
			g.Expect(e.Code()).To(Equal(test.code))
		} else {
			g.Expect(err).To(BeNil())
			g.Expect(r).ToNot(BeNil())
		}

	}
}

func TestUnloadModel(t *testing.T) {
	t.Logf("Started")
	g := NewGomegaWithT(t)

	type test struct {
		req []*pba.AgentSubscribeRequest
		model *pb.ModelDetails
		code codes.Code
	}
	modelName := "model1"
	smallMemory := uint64(100)
	tests := []test {
		{req: []*pba.AgentSubscribeRequest{{ServerName: "server1", ReplicaIdx: 0,
			ReplicaConfig: &pba.ReplicaConfig{InferenceSvc: "server1", InferencePort: 1, Memory: 1000, Capabilities: []string{"sklearn"}}}},
			model: &pb.ModelDetails{Name: modelName, Uri: "gs://model", Requirements: []string{"sklearn"}, Memory:&smallMemory, Replicas: 1},
			code: codes.OK}, // simple create
		{req: []*pba.AgentSubscribeRequest{{ServerName: "server1", ReplicaIdx: 0,
			ReplicaConfig: &pba.ReplicaConfig{InferenceSvc: "server1", InferencePort: 1, Memory: 1000, Capabilities: []string{"sklearn","xgboost"}}}},
			model: &pb.ModelDetails{Name: modelName, Uri: "gs://model", Requirements: []string{"xgboost","sklearn"}, Memory:&smallMemory, Replicas: 1},
			code: codes.OK}, // multiple requirements
		{req: []*pba.AgentSubscribeRequest{{ServerName: "server1", ReplicaIdx: 0,
			ReplicaConfig: &pba.ReplicaConfig{InferenceSvc: "server1", InferencePort: 1, Memory: 1000, Capabilities: []string{"sklearn"}}},
			{ServerName: "server1", ReplicaIdx: 1,
				ReplicaConfig: &pba.ReplicaConfig{InferenceSvc: "server1", InferencePort: 1, Memory: 1000, Capabilities: []string{"sklearn"}}}},
			model: &pb.ModelDetails{Name: modelName, Uri: "gs://model", Requirements: []string{"sklearn"}, Memory:&smallMemory, Replicas: 2},
			code: codes.OK}, // schedule to 2 replicas
		{req: []*pba.AgentSubscribeRequest{{ServerName: "server1", ReplicaIdx: 0,
			ReplicaConfig: &pba.ReplicaConfig{InferenceSvc: "server1", InferencePort: 1, Memory: 1000, Capabilities: []string{"sklearn"}}}},
			model: nil,
			code: codes.FailedPrecondition}, // Fail to unload model that does not exist
	}
	for _,test := range tests {
		s := createTestScheduler()
		for _, repReq := range test.req {
			err := s.store.UpdateServerReplica(repReq) // Create server and replicas
			g.Expect(err).To(BeNil())
		}

		if test.model != nil {
			lm := pb.LoadModelRequest{
				Model: test.model,
			}
			r,err := s.LoadModel(context.Background(), &lm)
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
		}
	}
}
