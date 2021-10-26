package agent

import (
	"context"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	. "github.com/onsi/gomega"
	pb "github.com/seldonio/seldon-core/scheduler/apis/mlops/agent"
	"github.com/seldonio/seldon-core/scheduler/pkg/envoy/processor"
	"github.com/seldonio/seldon-core/scheduler/pkg/store"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/metadata"
	"testing"
	"time"
)

type MockAgentServer struct {
	sentMessages int
}

func NewMockAgentServer() *MockAgentServer {
	return &MockAgentServer{}
}

func (m *MockAgentServer) Send(message *pb.ModelOperationMessage) error {
	m.sentMessages++
	return nil
}

func (m MockAgentServer) SetHeader(md metadata.MD) error {
	panic("implement me")
}

func (m MockAgentServer) SendHeader(md metadata.MD) error {
	panic("implement me")
}

func (m MockAgentServer) SetTrailer(md metadata.MD) {
	panic("implement me")
}

func (m MockAgentServer) Context() context.Context {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(time.Millisecond*80))
	return ctx
}

func (m MockAgentServer) SendMsg(message interface{}) error {
	panic("implement me")
}

func (m MockAgentServer) RecvMsg(message interface{}) error {
	panic("implement me")
}

func setupTestAgent() (*Server, *store.MemorySchedulerStore, chan string) {
	logger := log.New()
	log.SetLevel(log.DebugLevel)
	agentChan := make(chan string, 10)
	envoyChan := make(chan string, 10)
	schedulerStore := store.NewMemoryScheduler(logger, agentChan, envoyChan)
	cache := cache.NewSnapshotCache(false, cache.IDHash{}, logger)
	es := processor.NewIncrementalProcessor(cache, "nodeID", logger, schedulerStore, envoyChan)
	as := NewAgentServer(logger, schedulerStore, es, agentChan)
	go as.ListenForSyncs()
	return as, schedulerStore, agentChan
}

func TestSubscribe(t *testing.T) {
	g := NewGomegaWithT(t)
	as,_ ,ch := setupTestAgent()
	defer close(ch)
	mockStream := NewMockAgentServer()
	err := as.Subscribe(&pb.AgentSubscribeRequest{ServerName: "test", ReplicaIdx: 0, ReplicaConfig: &pb.ReplicaConfig{Capabilities: []string{"sklearn"}, Memory: 1000}}, mockStream)
	g.Expect(err).To(BeNil())
	g.Expect(mockStream.sentMessages).To(Equal(0))
}

