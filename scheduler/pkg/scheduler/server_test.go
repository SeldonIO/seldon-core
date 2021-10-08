package scheduler

import (
	"context"
	"fmt"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	. "github.com/onsi/gomega"
	pb "github.com/seldonio/seldon-core/scheduler/apis/mlops/scheduler"
	log "github.com/sirupsen/logrus"
	"testing"
)

func createTestServer(name string, memory int, replicas int, capabilities []string) (*pb.ServerDetails) {
	s := &pb.ServerDetails{
		Name: name,
		Replicas: []*pb.ServerReplica{},
		Memory: 1e6,
		Capabilities: capabilities,
	}
	for i:=0; i < replicas; i++ {
		r := &pb.ServerReplica{
		InferenceSvc: fmt.Sprintf("server.default%d",i),
			InferencePort: 9000,
			AgentPort: 9001,
		}
		s.Replicas = append(s.Replicas,r)
	}
	return s
}


func createTestModel(name string, requirements []string, memory int32, replicas int32) (*pb.ModelDetails) {
	return &pb.ModelDetails{
		Name: name,
		Requirements: requirements,
		Memory: memory,
		Replicas: replicas,
	}
}


func createTestScheduler() *SchedulerServer{
	logger := log.New()
	log.SetLevel(log.DebugLevel)
	// Create a cache
	cache := cache.NewSnapshotCache(false, cache.IDHash{}, logger)
	s := NewScheduler(cache, "node1", logger)
	return s
}

func TestAddServer(t *testing.T) {
	t.Logf("Started")
	g := NewGomegaWithT(t)
	serverName := "test"
	s := createTestScheduler()
	sd := createTestServer(serverName, 1e6, 1, []string{"sklearn"})
	_, err := s.AddServer(context.Background(),sd)
	g.Expect(err).To(BeNil())
	ss, err := s.ServerStatus(context.Background(),&pb.ServerReference{Name: serverName})
	g.Expect(err).To(BeNil())
	g.Expect(len(ss.LoadedModels)).To(Equal(0))
	g.Expect(ss.ServerName).To(Equal(serverName))
}

func TestAddServerEmptyName(t *testing.T) {
	t.Logf("Started")
	g := NewGomegaWithT(t)

	s := createTestScheduler()
	sd := createTestServer("", 1e6, 1, []string{"sklearn"})
	_, err := s.AddServer(context.Background(),sd)
	g.Expect(err).ToNot(BeNil())
	g.Expect(err).To(Equal(ErrAddServerEmptyServerName))
}

func TestRemoveServerNotFound(t *testing.T) {
	t.Logf("Started")
	g := NewGomegaWithT(t)

	s := createTestScheduler()
	sr := &pb.ServerReference{Name: "foo"}
	_, err := s.RemoveServer(context.Background(), sr)
	g.Expect(err).ToNot(BeNil())
	g.Expect(err).To(Equal(ErrRemoveServerServerNotFound))
}

func TestLoadModel(t *testing.T) {
	t.Logf("Started")
	g := NewGomegaWithT(t)
	serverName := "test"
	modelName := "model"
	s := createTestScheduler()
	sd := createTestServer(serverName, 1e6, 2, []string{"sklearn"})
	md := createTestModel(modelName, []string{"sklearn"}, 1e5, 2)
	_, err := s.AddServer(context.Background(),sd)
	g.Expect(err).To(BeNil())
	_, err = s.LoadModel(context.Background(), md)
	g.Expect(err).To(BeNil())
	ss, err := s.ServerStatus(context.Background(),&pb.ServerReference{Name: serverName})
	g.Expect(err).To(BeNil())
	g.Expect(len(ss.LoadedModels)).To(Equal(1))
	g.Expect(ss.ServerName).To(Equal(serverName))
	ms, err := s.ModelStatus(context.Background(),&pb.ModelReference{Name: modelName})
	g.Expect(err).To(BeNil())
	g.Expect(ms.ServerName).To(Equal(serverName))
	g.Expect(ms.ModelName).To(Equal(modelName))
}

func TestLoadModelFailedRequirement(t *testing.T) {
	t.Logf("Started")
	g := NewGomegaWithT(t)

	s := createTestScheduler()
	sd := createTestServer("test", 1e6, 1, []string{"sklearn"})
	md := createTestModel("model", []string{"foo"}, 1e5, 1)
	_, err := s.AddServer(context.Background(),sd)
	g.Expect(err).To(BeNil())
	_, err = s.LoadModel(context.Background(), md)
	g.Expect(err).ToNot(BeNil())
	g.Expect(err).To(Equal(ErrLoadModelRequirementFailed))
}

func TestLoadModelFailedRequirementReplicas(t *testing.T) {
	t.Logf("Started")
	g := NewGomegaWithT(t)

	s := createTestScheduler()
	sd := createTestServer("test", 1e6, 1, []string{"sklearn"})
	md := createTestModel("model", []string{"sklearn"}, 1e5, 2)
	_, err := s.AddServer(context.Background(),sd)
	g.Expect(err).To(BeNil())
	_, err = s.LoadModel(context.Background(), md)
	g.Expect(err).ToNot(BeNil())
	g.Expect(err).To(Equal(ErrLoadModelRequirementFailed))
}

func TestLoadModelUpdates(t *testing.T) {
	modelName := "model"
	serverName := "server"
	replicas := int32(7)
	t.Logf("Started")
	g := NewGomegaWithT(t)

	s := createTestScheduler()
	sd := createTestServer(serverName, 1e6, 20, []string{"sklearn"})
	md := createTestModel(modelName, []string{"sklearn"}, 1e5, 1)
	// Add Server
	_, err := s.AddServer(context.Background(),sd)
	g.Expect(err).To(BeNil())
	// Add Model
	_, err = s.LoadModel(context.Background(), md)
	g.Expect(err).To(BeNil())
	md.Replicas = replicas
	// Update Model
	_, err = s.LoadModel(context.Background(), md)
	g.Expect(err).To(BeNil())
	// get Model status
	ms, err := s.ModelStatus(context.Background(), &pb.ModelReference{Name: modelName})
	g.Expect(err).To(BeNil())
	g.Expect(ms.ModelName).To(Equal(modelName))
	g.Expect(ms.ServerName).To(Equal(serverName))
	g.Expect(len(ms.Assignment)).To(Equal(int(replicas)))
}