package agent

import (
	"context"
	"net"
	"testing"

	"github.com/jarcoal/httpmock"
	. "github.com/onsi/gomega"
	pb "github.com/seldonio/seldon-core/scheduler/apis/mlops/agent"
	pbs "github.com/seldonio/seldon-core/scheduler/apis/mlops/scheduler"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

type mockAgentV2Server struct {
	models []string
	pb.UnimplementedAgentServiceServer
	loadedEvents       int
	loadFailedEvents   int
	unloadedEvents     int
	unloadFailedEvents int
	otherEvents        int
}

func dialerv2(mockAgentV2Server *mockAgentV2Server) func(context.Context, string) (net.Conn, error) {
	listener := bufconn.Listen(1024 * 1024)

	server := grpc.NewServer()

	pb.RegisterAgentServiceServer(server, mockAgentV2Server)

	go func() {
		if err := server.Serve(listener); err != nil {
			log.Fatal(err)
		}
	}()

	return func(context.Context, string) (net.Conn, error) {
		return listener.Dial()
	}
}

func (m *mockAgentV2Server) AgentEvent(ctx context.Context, message *pb.ModelEventMessage) (*pb.ModelEventResponse, error) {
	switch message.Event {
	case pb.ModelEventMessage_LOADED:
		m.loadedEvents++
	case pb.ModelEventMessage_UNLOADED:
		m.unloadedEvents++
	case pb.ModelEventMessage_LOAD_FAILED:
		m.loadFailedEvents++
	case pb.ModelEventMessage_UNLOAD_FAILED:
		m.unloadFailedEvents++
	default:
		m.otherEvents++
	}
	return &pb.ModelEventResponse{}, nil
}

func (m mockAgentV2Server) Subscribe(request *pb.AgentSubscribeRequest, server pb.AgentService_SubscribeServer) error {
	for _, model := range m.models {
		err := server.Send(&pb.ModelOperationMessage{
			Operation: pb.ModelOperationMessage_LOAD_MODEL,
			Details: &pbs.ModelDetails{
				Name: model,
				Uri:  "gs://model",
			},
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func TestClientCreate(t *testing.T) {
	t.Logf("Started")
	logger := log.New()
	log.SetLevel(log.DebugLevel)
	g := NewGomegaWithT(t)

	type test struct {
		models        []string
		replicaConfig *pb.ReplicaConfig
		v2Status      int
		rsStatus      int
	}
	tests := []test{
		{models: []string{"model"}, replicaConfig: &pb.ReplicaConfig{}, v2Status: 200, rsStatus: 200},
		{models: []string{"model"}, replicaConfig: &pb.ReplicaConfig{}, v2Status: 400, rsStatus: 200},
	}

	for _, test := range tests {
		httpmock.Activate()
		v2Client := createTestV2Client(test.models, test.v2Status)
		rcloneClient := createTestRCloneClient(test.rsStatus)
		client, err := NewClient("mlserver", 1, "scheduler", 9002, logger, rcloneClient, v2Client, test.replicaConfig, "0.0.0.0")
		g.Expect(err).To(BeNil())
		mockAgentV2Server := &mockAgentV2Server{models: test.models}
		conn, err := grpc.DialContext(context.Background(), "", grpc.WithInsecure(), grpc.WithContextDialer(dialerv2(mockAgentV2Server)))
		g.Expect(err).To(BeNil())
		client.conn = conn
		err = client.Start()
		g.Expect(err).To(BeNil())
		if test.v2Status == 200 && test.rsStatus == 200 {
			g.Expect(mockAgentV2Server.loadedEvents).To(Equal(1))
			g.Expect(mockAgentV2Server.loadFailedEvents).To(Equal(0))
		} else {
			g.Expect(mockAgentV2Server.loadedEvents).To(Equal(0))
			g.Expect(mockAgentV2Server.loadFailedEvents).To(Equal(1))
		}
		httpmock.DeactivateAndReset()
		err = conn.Close()
		g.Expect(err).To(BeNil())
	}
}

func TestLoadModel(t *testing.T) {
	t.Logf("Started")
	logger := log.New()
	log.SetLevel(log.DebugLevel)
	g := NewGomegaWithT(t)

	type test struct {
		models                  []string
		replicaConfig           *pb.ReplicaConfig
		op                      *pb.ModelOperationMessage
		expectedAvailableMemory uint64
		v2Status                int
		rsStatus                int
		success                 bool
	}
	smallMemory := uint64(500)
	largeMemory := uint64(2000)
	tests := []test{
		{models: []string{"iris"},
			op:                      &pb.ModelOperationMessage{Details: &pbs.ModelDetails{Name: "iris", Uri: "gs://models/iris", MemoryBytes: &smallMemory}},
			replicaConfig:           &pb.ReplicaConfig{MemoryBytes: 1000},
			expectedAvailableMemory: 500,
			v2Status:                200,
			rsStatus:                200,
			success:                 true}, // Success
		{models: []string{"iris"},
			op:                      &pb.ModelOperationMessage{Details: &pbs.ModelDetails{Name: "iris", Uri: "gs://models/iris", MemoryBytes: &smallMemory}},
			replicaConfig:           &pb.ReplicaConfig{MemoryBytes: 1000},
			expectedAvailableMemory: 1000,
			v2Status:                400,
			rsStatus:                200,
			success:                 false}, // Fail as V2 fail
		{models: []string{"iris"},
			op:                      &pb.ModelOperationMessage{Details: &pbs.ModelDetails{Name: "iris", Uri: "gs://models/iris", MemoryBytes: &largeMemory}},
			replicaConfig:           &pb.ReplicaConfig{MemoryBytes: 1000},
			expectedAvailableMemory: 500,
			v2Status:                200,
			rsStatus:                200,
			success:                 false}, // Fail due to too much memory required
	}

	for tidx, test := range tests {
		t.Logf("Test #%d", tidx)
		httpmock.Activate()
		v2Client := createTestV2Client(test.models, test.v2Status)
		rcloneClient := createTestRCloneClient(test.rsStatus)
		client, err := NewClient("mlserver", 1, "scheduler", 9002, logger, rcloneClient, v2Client, test.replicaConfig, "0.0.0.0")
		g.Expect(err).To(BeNil())
		mockAgentV2Server := &mockAgentV2Server{models: []string{}}
		conn, cerr := grpc.DialContext(context.Background(), "", grpc.WithInsecure(), grpc.WithContextDialer(dialerv2(mockAgentV2Server)))
		g.Expect(cerr).To(BeNil())
		client.conn = conn
		err = client.Start()
		g.Expect(err).To(BeNil())
		err = client.LoadModel(test.op)
		if test.success {
			g.Expect(err).To(BeNil())
			g.Expect(mockAgentV2Server.loadedEvents).To(Equal(1))
			g.Expect(mockAgentV2Server.loadFailedEvents).To(Equal(0))
			g.Expect(client.replicaConfig.AvailableMemoryBytes).To(Equal(test.expectedAvailableMemory))
		} else {
			g.Expect(err).ToNot(BeNil())
			g.Expect(mockAgentV2Server.loadedEvents).To(Equal(0))
			g.Expect(mockAgentV2Server.loadFailedEvents).To(Equal(1))
		}
		httpmock.DeactivateAndReset()
		err = conn.Close()
		g.Expect(err).To(BeNil())
	}
}

func TestUnloadModel(t *testing.T) {
	t.Logf("Started")
	logger := log.New()
	log.SetLevel(log.DebugLevel)
	g := NewGomegaWithT(t)

	type test struct {
		models                  []string
		replicaConfig           *pb.ReplicaConfig
		loadOp                  *pb.ModelOperationMessage
		unloadOp                *pb.ModelOperationMessage
		expectedAvailableMemory uint64
		v2Status                int
		success                 bool
	}
	smallMemory := uint64(500)
	tests := []test{
		{models: []string{"iris"},
			loadOp:                  &pb.ModelOperationMessage{Details: &pbs.ModelDetails{Name: "iris", Uri: "gs://models/iris", MemoryBytes: &smallMemory}},
			unloadOp:                &pb.ModelOperationMessage{Details: &pbs.ModelDetails{Name: "iris", Uri: "gs://models/iris", MemoryBytes: &smallMemory}},
			replicaConfig:           &pb.ReplicaConfig{MemoryBytes: 1000},
			expectedAvailableMemory: 1000,
			v2Status:                200,
			success:                 true}, // Success
		{models: []string{"iris"},
			loadOp:                  &pb.ModelOperationMessage{Details: &pbs.ModelDetails{Name: "iris", Uri: "gs://models/iris", MemoryBytes: &smallMemory}},
			unloadOp:                &pb.ModelOperationMessage{Details: &pbs.ModelDetails{Name: "iris2", Uri: "gs://models/iris", MemoryBytes: &smallMemory}},
			replicaConfig:           &pb.ReplicaConfig{MemoryBytes: 1000},
			expectedAvailableMemory: 500,
			v2Status:                200,
			success:                 false}, // Fail to unload unknown model
	}

	for tidx, test := range tests {
		t.Logf("Test #%d", tidx)
		httpmock.Activate()
		v2Client := createTestV2Client(test.models, test.v2Status)
		rcloneClient := createTestRCloneClient(200)
		client, err := NewClient("mlserver", 1, "scheduler", 9002, logger, rcloneClient, v2Client, test.replicaConfig, "0.0.0.0")
		g.Expect(err).To(BeNil())
		mockAgentV2Server := &mockAgentV2Server{models: []string{}}
		conn, cerr := grpc.DialContext(context.Background(), "", grpc.WithInsecure(), grpc.WithContextDialer(dialerv2(mockAgentV2Server)))
		g.Expect(cerr).To(BeNil())
		client.conn = conn
		err = client.Start()
		g.Expect(err).To(BeNil())
		err = client.LoadModel(test.loadOp)
		g.Expect(err).To(BeNil())
		err = client.UnloadModel(test.unloadOp)
		if test.success {
			g.Expect(err).To(BeNil())
			g.Expect(mockAgentV2Server.loadedEvents).To(Equal(1))
			g.Expect(mockAgentV2Server.unloadedEvents).To(Equal(1))
			g.Expect(mockAgentV2Server.loadFailedEvents).To(Equal(0))
			g.Expect(mockAgentV2Server.unloadFailedEvents).To(Equal(0))
			g.Expect(client.replicaConfig.AvailableMemoryBytes).To(Equal(test.expectedAvailableMemory))
		} else {
			g.Expect(err).ToNot(BeNil())
			g.Expect(mockAgentV2Server.loadedEvents).To(Equal(1))
			g.Expect(mockAgentV2Server.unloadedEvents).To(Equal(0))
			g.Expect(mockAgentV2Server.loadFailedEvents).To(Equal(0))
			g.Expect(mockAgentV2Server.unloadFailedEvents).To(Equal(1))
		}
		httpmock.DeactivateAndReset()
		err = conn.Close()
		g.Expect(err).To(BeNil())
	}
}
