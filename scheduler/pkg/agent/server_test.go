package agent

import (
	"context"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	pb "github.com/seldonio/seldon-core/scheduler/apis/mlops/agent"
	"github.com/seldonio/seldon-core/scheduler/pkg/store"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/metadata"
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
	ctx, _ := context.WithTimeout(context.Background(), time.Millisecond*80) // nolint
	return ctx
}

func (m MockAgentServer) SendMsg(message interface{}) error {
	panic("implement me")
}

func (m MockAgentServer) RecvMsg(message interface{}) error {
	panic("implement me")
}

type mockEnvoyHandler struct {
	sentSyncs int
}

func (m *mockEnvoyHandler) SendEnvoySync(modelName string) {
	m.sentSyncs++
}

type mockScheduler struct {
	numSchedules int
}

func (m *mockScheduler) ScheduleFailedModels() ([]string, error) {
	return nil, nil
}

func (m *mockScheduler) Schedule(modelKey string) error {
	m.numSchedules++
	return nil
}

func setupTestAgent() (*Server, *store.MemoryStore) {
	logger := log.New()
	log.SetLevel(log.DebugLevel)
	schedulerStore := store.NewMemoryStore(logger, store.NewLocalSchedulerStore())
	mockEnvoyHandler := &mockEnvoyHandler{}
	mockSched := &mockScheduler{}
	as := NewAgentServer(logger, schedulerStore, mockEnvoyHandler, mockSched)
	go as.ListenForSyncs()
	return as, schedulerStore
}

func TestSubscribe(t *testing.T) {
	g := NewGomegaWithT(t)
	as, _ := setupTestAgent()
	mockStream := NewMockAgentServer()
	err := as.Subscribe(&pb.AgentSubscribeRequest{ServerName: "test", ReplicaIdx: 0, ReplicaConfig: &pb.ReplicaConfig{Capabilities: []string{"sklearn"}, MemoryBytes: 1000}}, mockStream)
	g.Expect(err).To(BeNil())
	g.Expect(mockStream.sentMessages).To(Equal(0))
}
