package agent

import (
	"context"
	"fmt"
	"net"
	"testing"

	"google.golang.org/grpc/credentials/insecure"

	"github.com/jarcoal/httpmock"
	. "github.com/onsi/gomega"
	pb "github.com/seldonio/seldon-core/scheduler/apis/mlops/agent"
	pbs "github.com/seldonio/seldon-core/scheduler/apis/mlops/scheduler"
	"github.com/seldonio/seldon-core/scheduler/pkg/agent/k8s"
	"github.com/seldonio/seldon-core/scheduler/pkg/util"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

type mockAgentV2Server struct {
	models []string
	pb.UnimplementedAgentServiceServer
	loadedEvents       int
	loadFailedEvents   int
	unloadedEvents     int
	unloadFailedEvents int
	otherEvents        int
	errors             int
}

type FakeModelRepository struct {
	err error
}

func (f FakeModelRepository) RemoveModelVersion(modelName string, version uint32) (int, error) {
	return 0, nil
}

func (f FakeModelRepository) DownloadModelVersion(modelName string, version uint32, artifactVersion *uint32, srcUri string, config []byte) (*string, error) {
	if f.err != nil {
		return nil, f.err
	}
	path := "path"
	return &path, nil
}

func (f FakeModelRepository) Ready() error {
	return nil
}

type FakeClientService struct {
	err error
}

func (f FakeClientService) SetState(state *LocalStateManager) {
}

func (f FakeClientService) Start() error {
	return f.err
}

func (f FakeClientService) Ready() bool {
	return f.err == nil
}

func (f FakeClientService) Stop() error {
	return f.err
}

func (f FakeClientService) Name() string {
	return "FakeService"
}

func addVerionToModels(models []string, version uint32) []string {
	modelsWithVersions := make([]string, len(models))
	for i := 0; i < len(models); i++ {
		modelsWithVersions[i] = util.GetVersionedModelName(models[i], version)
	}
	return modelsWithVersions
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

func (m *mockAgentV2Server) loaded() int {
	return m.loadedEvents
}

func (m *mockAgentV2Server) loadFailed() int {
	return m.loadFailedEvents
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

func (m *mockAgentV2Server) Subscribe(request *pb.AgentSubscribeRequest, server pb.AgentService_SubscribeServer) error {
	for _, model := range m.models {
		err := server.Send(&pb.ModelOperationMessage{
			Operation: pb.ModelOperationMessage_LOAD_MODEL,
			ModelVersion: &pb.ModelVersion{
				Model: &pbs.Model{
					Meta: &pbs.MetaData{
						Name: model,
					},
					ModelSpec: &pbs.ModelSpec{Uri: "gs://model"},
				},
			},
		})
		if err != nil {
			m.errors++
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
		name          string
		models        []string
		replicaConfig *pb.ReplicaConfig
		v2Status      int
		modelRepoErr  error
	}
	tests := []test{
		{name: "v2Response200", models: []string{"model"}, replicaConfig: &pb.ReplicaConfig{}, v2Status: 200},
		{name: "v2Response400", models: []string{"model"}, replicaConfig: &pb.ReplicaConfig{}, v2Status: 400, modelRepoErr: fmt.Errorf("repo err")},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			httpmock.Activate()
			defer httpmock.DeactivateAndReset()
			v2Client := createTestV2Client(addVerionToModels(test.models, 0), test.v2Status)
			modelRepository := FakeModelRepository{err: test.modelRepoErr}
			rpHTTP := FakeClientService{err: nil}
			rpGRPC := FakeClientService{err: nil}
			clientDebug := FakeClientService{err: nil}
			client := NewClient("mlserver", 1, "scheduler", 9002, logger, modelRepository, v2Client, test.replicaConfig, "0.0.0.0", "default", rpHTTP, rpGRPC, clientDebug)
			mockAgentV2Server := &mockAgentV2Server{models: test.models}
			conn, err := grpc.DialContext(context.Background(), "", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithContextDialer(dialerv2(mockAgentV2Server)))
			g.Expect(err).To(BeNil())
			client.conn = conn
			err = client.Start()
			g.Expect(err).To(BeNil())
			if test.v2Status == 200 && test.modelRepoErr == nil {
				g.Eventually(mockAgentV2Server.loaded).Should(Equal(1))
				g.Eventually(mockAgentV2Server.loadFailed).Should(Equal(0))
			} else {
				g.Eventually(mockAgentV2Server.loaded).Should(Equal(0))
				g.Eventually(mockAgentV2Server.loadFailed).Should(Equal(1))
			}
			err = conn.Close()
			g.Expect(err).To(BeNil())
		})
	}
}

func TestLoadModel(t *testing.T) {
	t.Logf("Started")
	logger := log.New()
	log.SetLevel(log.DebugLevel)
	g := NewGomegaWithT(t)

	type test struct {
		name                    string
		models                  []string
		replicaConfig           *pb.ReplicaConfig
		op                      *pb.ModelOperationMessage
		expectedAvailableMemory uint64
		v2Status                int
		modelRepoErr            error
		success                 bool
	}
	smallMemory := uint64(500)
	largeMemory := uint64(2000)
	tests := []test{
		{
			name:   "simple",
			models: []string{"iris"},
			op: &pb.ModelOperationMessage{
				Operation: pb.ModelOperationMessage_LOAD_MODEL,
				ModelVersion: &pb.ModelVersion{
					Model: &pbs.Model{
						Meta: &pbs.MetaData{
							Name: "iris",
						},
						ModelSpec: &pbs.ModelSpec{Uri: "gs://model", MemoryBytes: &smallMemory},
					},
				},
			},
			replicaConfig:           &pb.ReplicaConfig{MemoryBytes: 1000},
			expectedAvailableMemory: 500,
			v2Status:                200,
			success:                 true}, // Success
		{
			name:   "V2Fail",
			models: []string{"iris"},
			op: &pb.ModelOperationMessage{
				Operation: pb.ModelOperationMessage_LOAD_MODEL,
				ModelVersion: &pb.ModelVersion{
					Model: &pbs.Model{
						Meta: &pbs.MetaData{
							Name: "iris",
						},
						ModelSpec: &pbs.ModelSpec{Uri: "gs://model", MemoryBytes: &smallMemory},
					},
				},
			},
			replicaConfig:           &pb.ReplicaConfig{MemoryBytes: 1000},
			expectedAvailableMemory: 1000,
			v2Status:                400,
			success:                 false}, // Fail as V2 fail
		{
			name:   "MemoryAvailableFail",
			models: []string{"iris"},
			op: &pb.ModelOperationMessage{
				Operation: pb.ModelOperationMessage_LOAD_MODEL,
				ModelVersion: &pb.ModelVersion{
					Model: &pbs.Model{
						Meta: &pbs.MetaData{
							Name: "iris",
						},
						ModelSpec: &pbs.ModelSpec{Uri: "gs://model", MemoryBytes: &largeMemory},
					},
				},
			},
			replicaConfig:           &pb.ReplicaConfig{MemoryBytes: 1000},
			expectedAvailableMemory: 1000,
			v2Status:                200,
			success:                 false}, // Fail due to too much memory required
	}

	for tidx, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Logf("Test #%d", tidx)
			httpmock.Activate()
			defer httpmock.DeactivateAndReset()
			v2Client := createTestV2Client(addVerionToModels(test.models, 0), test.v2Status)
			modelRepository := FakeModelRepository{err: test.modelRepoErr}
			rpHTTP := FakeClientService{err: nil}
			rpGRPC := FakeClientService{err: nil}
			clientDebug := FakeClientService{err: nil}
			client := NewClient("mlserver", 1, "scheduler", 9002, logger, modelRepository, v2Client, test.replicaConfig, "0.0.0.0", "default", rpHTTP, rpGRPC, clientDebug)
			mockAgentV2Server := &mockAgentV2Server{models: []string{}}
			conn, cerr := grpc.DialContext(context.Background(), "", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithContextDialer(dialerv2(mockAgentV2Server)))
			g.Expect(cerr).To(BeNil())
			client.conn = conn
			err := client.Start()
			g.Expect(err).To(BeNil())
			err = client.LoadModel(test.op)
			if test.success {
				g.Expect(err).To(BeNil())
				g.Expect(mockAgentV2Server.loadedEvents).To(Equal(1))
				g.Expect(mockAgentV2Server.loadFailedEvents).To(Equal(0))
				g.Expect(client.stateManager.GetAvailableMemoryBytes()).To(Equal(test.expectedAvailableMemory))
			} else {
				g.Expect(err).ToNot(BeNil())
				g.Expect(mockAgentV2Server.loadedEvents).To(Equal(0))
				g.Expect(mockAgentV2Server.loadFailedEvents).To(Equal(1))
				g.Expect(client.stateManager.GetAvailableMemoryBytes()).To(Equal(test.expectedAvailableMemory))
			}
			err = conn.Close()
			g.Expect(err).To(BeNil())
		})
	}
}

func TestLoadModelWithAuth(t *testing.T) {
	t.Logf("Started")
	logger := log.New()
	log.SetLevel(log.DebugLevel)
	g := NewGomegaWithT(t)

	type test struct {
		name                    string
		models                  []string
		replicaConfig           *pb.ReplicaConfig
		op                      *pb.ModelOperationMessage
		secretData              string
		expectedAvailableMemory uint64
		v2Status                int
		success                 bool
	}
	rcloneConfig := `{"type":"s3","name":"s3","parameters":{"provider":"minio","env_auth":"false","access_key_id":"minioadmin","secret_access_key":"minioadmin","endpoint":"http://172.18.255.2:9000"}}`
	rcloneSecret := "minio-secret"
	yamlSecretDataOK := `
type: s3                                                                                                  
name: s3                                                                                                  
parameters:                                                                                               
   provider: minio                                                                                         
   env_auth: false                                                                                         
   access_key_id: minioadmin                                                                               
   secret_access_key: minioadmin                                                                           
   endpoint: http://172.18.255.2:9000
`
	smallMemory := uint64(500)
	tests := []test{
		{
			name:   "rclongConfig",
			models: []string{"iris"},
			op: &pb.ModelOperationMessage{
				Operation: pb.ModelOperationMessage_LOAD_MODEL,
				ModelVersion: &pb.ModelVersion{
					Model: &pbs.Model{
						Meta: &pbs.MetaData{
							Name: "iris",
						},
						ModelSpec: &pbs.ModelSpec{
							Uri:           "gs://model",
							MemoryBytes:   &smallMemory,
							StorageConfig: &pbs.StorageConfig{Config: &pbs.StorageConfig_StorageRcloneConfig{StorageRcloneConfig: rcloneConfig}},
						},
					},
				},
			},
			replicaConfig:           &pb.ReplicaConfig{MemoryBytes: 1000},
			expectedAvailableMemory: 500,
			v2Status:                200,
			success:                 true,
		},
		{
			name:   "secretConfig",
			models: []string{"iris"},
			op: &pb.ModelOperationMessage{
				Operation: pb.ModelOperationMessage_LOAD_MODEL,
				ModelVersion: &pb.ModelVersion{
					Model: &pbs.Model{
						Meta: &pbs.MetaData{
							Name: "iris",
						},
						ModelSpec: &pbs.ModelSpec{
							Uri:           "gs://model",
							MemoryBytes:   &smallMemory,
							StorageConfig: &pbs.StorageConfig{Config: &pbs.StorageConfig_StorageSecretName{StorageSecretName: rcloneSecret}},
						},
					},
				},
			},
			secretData:              yamlSecretDataOK,
			replicaConfig:           &pb.ReplicaConfig{MemoryBytes: 1000},
			expectedAvailableMemory: 500,
			v2Status:                200,
			success:                 true,
		},
	}

	for tidx, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Logf("Test #%d", tidx)
			httpmock.Activate()
			defer httpmock.DeactivateAndReset()
			v2Client := createTestV2Client(addVerionToModels(test.models, 0), test.v2Status)
			modelRepository := FakeModelRepository{}
			rpHTTP := FakeClientService{err: nil}
			rpGRPC := FakeClientService{err: nil}
			clientDebug := FakeClientService{err: nil}
			client := NewClient("mlserver", 1, "scheduler", 9002, logger, modelRepository, v2Client, test.replicaConfig, "0.0.0.0", "default", rpHTTP, rpGRPC, clientDebug)
			switch x := test.op.GetModelVersion().GetModel().GetModelSpec().StorageConfig.Config.(type) {
			case *pbs.StorageConfig_StorageSecretName:
				secret := &v1.Secret{ObjectMeta: metav1.ObjectMeta{Name: x.StorageSecretName, Namespace: client.namespace}, StringData: map[string]string{"mys3": test.secretData}}
				fakeClientset := fake.NewSimpleClientset(secret)
				s := k8s.NewSecretsHandler(fakeClientset, client.namespace)
				client.secretsHandler = s
			}
			mockAgentV2Server := &mockAgentV2Server{models: []string{}}
			conn, cerr := grpc.DialContext(context.Background(), "", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithContextDialer(dialerv2(mockAgentV2Server)))
			g.Expect(cerr).To(BeNil())
			client.conn = conn
			err := client.Start()
			g.Expect(err).To(BeNil())
			err = client.LoadModel(test.op)
			if test.success {
				g.Expect(err).To(BeNil())
				g.Expect(mockAgentV2Server.loadedEvents).To(Equal(1))
				g.Expect(mockAgentV2Server.loadFailedEvents).To(Equal(0))
				g.Expect(client.stateManager.GetAvailableMemoryBytes()).To(Equal(test.expectedAvailableMemory))
			} else {
				g.Expect(err).ToNot(BeNil())
				g.Expect(mockAgentV2Server.loadedEvents).To(Equal(0))
			}
			err = conn.Close()
			g.Expect(err).To(BeNil())
		})
	}
}

func TestUnloadModel(t *testing.T) {
	t.Logf("Started")
	logger := log.New()
	log.SetLevel(log.DebugLevel)
	g := NewGomegaWithT(t)

	type test struct {
		name                    string
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
		{
			name:   "simple",
			models: []string{"iris"},
			loadOp: &pb.ModelOperationMessage{
				Operation: pb.ModelOperationMessage_LOAD_MODEL,
				ModelVersion: &pb.ModelVersion{
					Model: &pbs.Model{
						Meta: &pbs.MetaData{
							Name: "iris",
						},
						ModelSpec: &pbs.ModelSpec{Uri: "gs://model", MemoryBytes: &smallMemory},
					},
				},
			},
			unloadOp: &pb.ModelOperationMessage{
				Operation: pb.ModelOperationMessage_LOAD_MODEL,
				ModelVersion: &pb.ModelVersion{
					Model: &pbs.Model{
						Meta: &pbs.MetaData{
							Name: "iris",
						},
						ModelSpec: &pbs.ModelSpec{Uri: "gs://model", MemoryBytes: &smallMemory},
					},
				},
			},
			replicaConfig:           &pb.ReplicaConfig{MemoryBytes: 1000},
			expectedAvailableMemory: 1000,
			v2Status:                200,
			success:                 true}, // Success
		{
			name:   "FailUnknownModel",
			models: []string{"iris"},
			loadOp: &pb.ModelOperationMessage{
				Operation: pb.ModelOperationMessage_LOAD_MODEL,
				ModelVersion: &pb.ModelVersion{
					Model: &pbs.Model{
						Meta: &pbs.MetaData{
							Name: "iris",
						},
						ModelSpec: &pbs.ModelSpec{Uri: "gs://model", MemoryBytes: &smallMemory},
					},
				},
			},
			unloadOp: &pb.ModelOperationMessage{
				Operation: pb.ModelOperationMessage_LOAD_MODEL,
				ModelVersion: &pb.ModelVersion{
					Model: &pbs.Model{
						Meta: &pbs.MetaData{
							Name: "iris2",
						},
						ModelSpec: &pbs.ModelSpec{Uri: "gs://model", MemoryBytes: &smallMemory},
					},
				},
			},
			replicaConfig:           &pb.ReplicaConfig{MemoryBytes: 1000},
			expectedAvailableMemory: 500,
			v2Status:                200,
			success:                 false},
	}

	for tidx, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Logf("Test #%d", tidx)
			httpmock.Activate()
			defer httpmock.DeactivateAndReset()
			v2Client := createTestV2Client(addVerionToModels(test.models, 0), test.v2Status)
			modelRepository := FakeModelRepository{}
			rpHTTP := FakeClientService{err: nil}
			rpGRPC := FakeClientService{err: nil}
			clientDebug := FakeClientService{err: nil}
			client := NewClient("mlserver", 1, "scheduler", 9002, logger, modelRepository, v2Client, test.replicaConfig, "0.0.0.0", "default", rpHTTP, rpGRPC, clientDebug)
			mockAgentV2Server := &mockAgentV2Server{models: []string{}}
			conn, cerr := grpc.DialContext(context.Background(), "", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithContextDialer(dialerv2(mockAgentV2Server)))
			g.Expect(cerr).To(BeNil())
			client.conn = conn
			err := client.Start()
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
				g.Expect(client.stateManager.GetAvailableMemoryBytes()).To(Equal(test.expectedAvailableMemory))
			} else {
				g.Expect(err).ToNot(BeNil())
				g.Expect(mockAgentV2Server.loadedEvents).To(Equal(1))
				g.Expect(mockAgentV2Server.unloadedEvents).To(Equal(0))
				g.Expect(mockAgentV2Server.loadFailedEvents).To(Equal(0))
				g.Expect(mockAgentV2Server.unloadFailedEvents).To(Equal(1))
			}
			err = conn.Close()
			g.Expect(err).To(BeNil())
		})
	}
}
