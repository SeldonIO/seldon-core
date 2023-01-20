/*
Copyright 2022 Seldon Technologies Ltd.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package agent

import (
	"context"
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/proto"

	"github.com/jarcoal/httpmock"
	. "github.com/onsi/gomega"
	pb "github.com/seldonio/seldon-core/apis/go/v2/mlops/agent"
	pbs "github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/drainservice"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/interfaces"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/k8s"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/modelscaling"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/util"
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

func (f FakeModelRepository) RemoveModelVersion(modelName string) error {
	return nil
}

func (f FakeModelRepository) DownloadModelVersion(modelName string, version uint32, artifactVersion *uint32, srcUri string, config []byte, explainer *pbs.ExplainerSpec, parameters []*pbs.ParameterSpec) (*string, error) {
	if f.err != nil {
		return nil, f.err
	}
	path := "path"
	return &path, nil
}

func (f FakeModelRepository) Ready() error {
	return f.err
}

type FakeDependencyService struct {
	err error
}

func (f FakeDependencyService) SetState(state interface{}) {
}

func (f FakeDependencyService) Start() error {
	return f.err
}

func (f FakeDependencyService) Ready() bool {
	return f.err == nil
}

func (f FakeDependencyService) Stop() error {
	return f.err
}

func (f FakeDependencyService) Name() string {
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
			AutoscalingEnabled: false,
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
			v2Client := createTestV2Client(addVerionToModels(test.models, 0), test.v2Status)
			httpmock.ActivateNonDefault(v2Client.httpClient)
			modelRepository := FakeModelRepository{err: test.modelRepoErr}
			rpHTTP := FakeDependencyService{err: nil}
			rpGRPC := FakeDependencyService{err: nil}
			agentDebug := FakeDependencyService{err: nil}
			lags := modelscaling.ModelScalingStatsWrapper{
				Stats:     modelscaling.NewModelReplicaLagsKeeper(),
				Operator:  interfaces.Gte,
				Threshold: 10,
				Reset:     true,
				EventType: modelscaling.ScaleUpEvent,
			}
			lastUsed := modelscaling.ModelScalingStatsWrapper{
				Stats:     modelscaling.NewModelReplicaLastUsedKeeper(),
				Operator:  interfaces.Gte,
				Threshold: 10,
				Reset:     false,
				EventType: modelscaling.ScaleDownEvent,
			}
			modelScalingService := modelscaling.NewStatsAnalyserService(
				[]modelscaling.ModelScalingStatsWrapper{lags, lastUsed}, logger, 10)
			drainerServicePort, _ := getFreePort()
			drainerService := drainservice.NewDrainerService(logger, uint(drainerServicePort))
			client := NewClient(
				NewClientSettings("mlserver", 1, "scheduler", 9002, 9055, 1*time.Minute, 1*time.Minute, 1*time.Minute),
				logger, modelRepository, v2Client,
				test.replicaConfig, "default",
				rpHTTP, rpGRPC, agentDebug, modelScalingService, drainerService, newFakeMetricsHandler())
			mockAgentV2Server := &mockAgentV2Server{models: test.models}
			conn, err := grpc.DialContext(
				context.Background(), "", grpc.WithTransportCredentials(insecure.NewCredentials()),
				grpc.WithContextDialer(dialerv2(mockAgentV2Server)))
			g.Expect(err).To(BeNil())
			client.conn = conn
			go func() {
				_ = client.Start()
			}()
			time.Sleep(10 * time.Millisecond)
			if test.v2Status == 200 && test.modelRepoErr == nil {
				g.Eventually(mockAgentV2Server.loaded).Should(BeNumerically(">", 1))
				g.Eventually(mockAgentV2Server.loadFailed).Should(Equal(0))
			} else {
				g.Eventually(mockAgentV2Server.loaded).Should(Equal(0))
				g.Eventually(mockAgentV2Server.loadFailed).Should(BeNumerically(">", 1))
			}
			client.Stop()
			httpmock.DeactivateAndReset()
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
		autoscalingEnabled      bool
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
			name:   "simple - autoscaling enabled",
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
				AutoscalingEnabled: true,
			},
			replicaConfig:           &pb.ReplicaConfig{MemoryBytes: 1000},
			expectedAvailableMemory: 500,
			v2Status:                200,
			success:                 true,
			autoscalingEnabled:      true}, // Success
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
			v2Client := createTestV2Client(addVerionToModels(test.models, 0), test.v2Status)
			httpmock.ActivateNonDefault(v2Client.httpClient)
			modelRepository := FakeModelRepository{err: test.modelRepoErr}
			rpHTTP := FakeDependencyService{err: nil}
			rpGRPC := FakeDependencyService{err: nil}
			agentDebug := FakeDependencyService{err: nil}
			lags := modelscaling.ModelScalingStatsWrapper{
				Stats:     modelscaling.NewModelReplicaLagsKeeper(),
				Operator:  interfaces.Gte,
				Threshold: 10,
				Reset:     true,
				EventType: modelscaling.ScaleUpEvent,
			}
			lastUsed := modelscaling.ModelScalingStatsWrapper{
				Stats:     modelscaling.NewModelReplicaLastUsedKeeper(),
				Operator:  interfaces.Gte,
				Threshold: 10,
				Reset:     false,
				EventType: modelscaling.ScaleDownEvent,
			}
			modelScalingService := modelscaling.NewStatsAnalyserService(
				[]modelscaling.ModelScalingStatsWrapper{lags, lastUsed}, logger, 10)
			drainerServicePort, _ := getFreePort()
			drainerService := drainservice.NewDrainerService(logger, uint(drainerServicePort))
			client := NewClient(
				NewClientSettings("mlserver", 1, "scheduler", 9002, 9055, 1*time.Minute, 1*time.Minute, 1*time.Minute),
				logger, modelRepository, v2Client, test.replicaConfig, "default",
				rpHTTP, rpGRPC, agentDebug, modelScalingService, drainerService, newFakeMetricsHandler())
			mockAgentV2Server := &mockAgentV2Server{models: []string{}}
			conn, cerr := grpc.DialContext(context.Background(), "", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithContextDialer(dialerv2(mockAgentV2Server)))
			g.Expect(cerr).To(BeNil())
			client.conn = conn
			go func() {
				_ = client.Start()
			}()
			time.Sleep(50 * time.Millisecond)
			err := client.LoadModel(test.op)
			if test.success {
				g.Expect(err).To(BeNil())
				g.Expect(mockAgentV2Server.loadedEvents).To(Equal(1))
				g.Expect(mockAgentV2Server.loadFailedEvents).To(Equal(0))
				g.Expect(client.stateManager.GetAvailableMemoryBytes()).To(Equal(test.expectedAvailableMemory))
				loadedVersions := client.stateManager.modelVersions.getVersionsForAllModels()
				// we have only one version in the test
				g.Expect(proto.Clone(loadedVersions[0])).To(Equal(proto.Clone(test.op.ModelVersion)))
				// we have set model stats state if autoscaling is enabled
				versionedModelName := util.GetVersionedModelName(test.op.GetModelVersion().Model.Meta.Name, test.op.GetModelVersion().GetVersion())
				if test.autoscalingEnabled {
					_, err := lags.Stats.Get(versionedModelName)
					g.Expect(err).To(BeNil())
					_, err = lastUsed.Stats.Get(versionedModelName)
					g.Expect(err).To(BeNil())
				} else {
					_, err := lags.Stats.Get(versionedModelName)
					g.Expect(err).ToNot(BeNil())
					_, err = lastUsed.Stats.Get(versionedModelName)
					g.Expect(err).ToNot(BeNil())
				}
			} else {
				g.Expect(err).ToNot(BeNil())
				g.Expect(mockAgentV2Server.loadedEvents).To(Equal(0))
				g.Expect(mockAgentV2Server.loadFailedEvents).To(Equal(1))
				g.Expect(client.stateManager.GetAvailableMemoryBytes()).To(Equal(test.expectedAvailableMemory))
			}
			client.Stop()
			httpmock.DeactivateAndReset()
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
			v2Client := createTestV2Client(addVerionToModels(test.models, 0), test.v2Status)
			httpmock.ActivateNonDefault(v2Client.httpClient)
			modelRepository := FakeModelRepository{}
			rpHTTP := FakeDependencyService{err: nil}
			rpGRPC := FakeDependencyService{err: nil}
			agentDebug := FakeDependencyService{err: nil}
			lags := modelscaling.ModelScalingStatsWrapper{
				Stats:     modelscaling.NewModelReplicaLagsKeeper(),
				Operator:  interfaces.Gte,
				Threshold: 10,
				Reset:     true,
				EventType: modelscaling.ScaleUpEvent,
			}
			lastUsed := modelscaling.ModelScalingStatsWrapper{
				Stats:     modelscaling.NewModelReplicaLastUsedKeeper(),
				Operator:  interfaces.Gte,
				Threshold: 10,
				Reset:     false,
				EventType: modelscaling.ScaleDownEvent,
			}
			modelScalingService := modelscaling.NewStatsAnalyserService(
				[]modelscaling.ModelScalingStatsWrapper{lags, lastUsed}, logger, 10)
			drainerServicePort, _ := getFreePort()
			drainerService := drainservice.NewDrainerService(logger, uint(drainerServicePort))
			client := NewClient(
				NewClientSettings("mlserver", 1, "scheduler", 9002, 9055, 1*time.Minute, 1*time.Minute, 1*time.Minute),
				logger, modelRepository,
				v2Client, test.replicaConfig, "default",
				rpHTTP, rpGRPC, agentDebug, modelScalingService, drainerService,
				newFakeMetricsHandler())
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
			go func() {
				_ = client.Start()
			}()
			time.Sleep(50 * time.Millisecond)
			err := client.LoadModel(test.op)
			if test.success {
				g.Expect(err).To(BeNil())
				g.Expect(mockAgentV2Server.loadedEvents).To(Equal(1))
				g.Expect(mockAgentV2Server.loadFailedEvents).To(Equal(0))
				g.Expect(client.stateManager.GetAvailableMemoryBytes()).To(Equal(test.expectedAvailableMemory))
			} else {
				g.Expect(err).ToNot(BeNil())
				g.Expect(mockAgentV2Server.loadedEvents).To(Equal(0))
			}
			client.Stop()
			httpmock.DeactivateAndReset()
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
			name:   "UnknownModel - unload ok",
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
			success:                 true},
	}

	for tidx, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Logf("Test #%d", tidx)
			v2Client := createTestV2Client(addVerionToModels(test.models, 0), test.v2Status)
			httpmock.ActivateNonDefault(v2Client.httpClient)
			modelRepository := FakeModelRepository{}
			rpHTTP := FakeDependencyService{err: nil}
			rpGRPC := FakeDependencyService{err: nil}
			agentDebug := FakeDependencyService{err: nil}
			lags := modelscaling.ModelScalingStatsWrapper{
				Stats:     modelscaling.NewModelReplicaLagsKeeper(),
				Operator:  interfaces.Gte,
				Threshold: 10,
				Reset:     true,
				EventType: modelscaling.ScaleUpEvent,
			}
			lastUsed := modelscaling.ModelScalingStatsWrapper{
				Stats:     modelscaling.NewModelReplicaLastUsedKeeper(),
				Operator:  interfaces.Gte,
				Threshold: 10,
				Reset:     false,
				EventType: modelscaling.ScaleDownEvent,
			}
			modelScalingService := modelscaling.NewStatsAnalyserService(
				[]modelscaling.ModelScalingStatsWrapper{lags, lastUsed}, logger, 10)
			drainerServicePort, _ := getFreePort()
			drainerService := drainservice.NewDrainerService(logger, uint(drainerServicePort))
			client := NewClient(
				NewClientSettings("mlserver", 1, "scheduler", 9002, 9055, 1*time.Minute, 1*time.Minute, 1*time.Minute),
				logger, modelRepository, v2Client, test.replicaConfig, "default",
				rpHTTP, rpGRPC, agentDebug, modelScalingService, drainerService, newFakeMetricsHandler())
			mockAgentV2Server := &mockAgentV2Server{models: []string{}}
			conn, cerr := grpc.DialContext(context.Background(), "", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithContextDialer(dialerv2(mockAgentV2Server)))
			g.Expect(cerr).To(BeNil())
			client.conn = conn
			go func() {
				_ = client.Start()
			}()
			time.Sleep(50 * time.Millisecond)
			err := client.LoadModel(test.loadOp)
			g.Expect(err).To(BeNil())
			err = client.UnloadModel(test.unloadOp)
			if test.success {
				g.Expect(err).To(BeNil())
				g.Expect(mockAgentV2Server.loadedEvents).To(Equal(1))
				g.Expect(mockAgentV2Server.unloadedEvents).To(Equal(1))
				g.Expect(mockAgentV2Server.loadFailedEvents).To(Equal(0))
				g.Expect(mockAgentV2Server.unloadFailedEvents).To(Equal(0))
				g.Expect(client.stateManager.GetAvailableMemoryBytes()).To(Equal(test.expectedAvailableMemory))
				// check model scaling stats removed
				versionedModelName := util.GetVersionedModelName(test.unloadOp.GetModelVersion().Model.Meta.Name, test.unloadOp.GetModelVersion().GetVersion())
				_, err := lags.Stats.Get(versionedModelName)
				g.Expect(err).ToNot(BeNil())
				_, err = lastUsed.Stats.Get(versionedModelName)
				g.Expect(err).ToNot(BeNil())
			} else {
				g.Expect(err).ToNot(BeNil())
				g.Expect(mockAgentV2Server.loadedEvents).To(Equal(1))
				g.Expect(mockAgentV2Server.unloadedEvents).To(Equal(0))
				g.Expect(mockAgentV2Server.loadFailedEvents).To(Equal(0))
				g.Expect(mockAgentV2Server.unloadFailedEvents).To(Equal(1))
			}
			client.Stop()
			httpmock.DeactivateAndReset()
		})
	}
}

func TestClientClose(t *testing.T) {
	t.Logf("Started")
	logger := log.New()
	log.SetLevel(log.DebugLevel)
	g := NewGomegaWithT(t)

	v2Client := createTestV2Client(nil, 200)
	httpmock.ActivateNonDefault(v2Client.httpClient)
	defer httpmock.DeactivateAndReset()
	modelRepository := FakeModelRepository{}
	rpHTTP := FakeDependencyService{err: nil}
	rpGRPC := FakeDependencyService{err: nil}
	agentDebug := FakeDependencyService{err: nil}
	lags := modelscaling.ModelScalingStatsWrapper{
		Stats:     modelscaling.NewModelReplicaLagsKeeper(),
		Operator:  interfaces.Gte,
		Threshold: 10,
		Reset:     true,
		EventType: modelscaling.ScaleUpEvent,
	}
	lastUsed := modelscaling.ModelScalingStatsWrapper{
		Stats:     modelscaling.NewModelReplicaLastUsedKeeper(),
		Operator:  interfaces.Gte,
		Threshold: 10,
		Reset:     false,
		EventType: modelscaling.ScaleDownEvent,
	}
	modelScalingService := modelscaling.NewStatsAnalyserService(
		[]modelscaling.ModelScalingStatsWrapper{lags, lastUsed}, logger, 10)
	drainerServicePort, _ := getFreePort()
	drainerService := drainservice.NewDrainerService(logger, uint(drainerServicePort))
	client := NewClient(
		NewClientSettings("mlserver", 1, "scheduler", 9002, 9055, 1*time.Minute, 1*time.Minute, 1*time.Minute),
		logger, modelRepository, v2Client,
		&pb.ReplicaConfig{MemoryBytes: 1000}, "default",
		rpHTTP, rpGRPC, agentDebug, modelScalingService, drainerService, newFakeMetricsHandler())
	mockAgentV2Server := &mockAgentV2Server{}
	conn, err := grpc.DialContext(
		context.Background(), "", grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(dialerv2(mockAgentV2Server)))
	g.Expect(err).To(BeNil())
	client.conn = conn

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		err = client.Start()
		g.Expect(err).To(BeNil())
		wg.Done()
	}()
	client.Stop()
	wg.Wait()
	g.Expect(client.conn.GetState()).To(Equal(connectivity.Shutdown))
	g.Expect(client.stop.Load()).To(BeTrue())
}

func TestClientCloseWithFailure(t *testing.T) {
	t.Logf("Started")
	logger := log.New()
	log.SetLevel(log.DebugLevel)
	g := NewGomegaWithT(t)

	const (
		inference string = "inference"
		drain     string = "drain"
		scale     string = "scale"
	)

	type test struct {
		name        string
		isError     bool
		serviceName string
	}
	tests := []test{
		{
			name:    "no-error",
			isError: false,
		},
		{
			name:        "error-" + inference,
			isError:     true,
			serviceName: inference,
		},
		{
			name:        "error-" + scale,
			isError:     true,
			serviceName: scale,
		},
		{
			name:        "error-" + drain,
			isError:     true,
			serviceName: drain,
		},
	}

	period := 10 * time.Millisecond
	maxTimeBeforeStart := 1 * time.Millisecond // not used in test
	maxTimeAfterStart := 1 * time.Millisecond
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			v2Client := createTestV2Client([]string{}, 200)
			httpmock.ActivateNonDefault(v2Client.httpClient)

			modelRepository := FakeModelRepository{}
			rpHTTP := FakeDependencyService{err: nil}
			rpGRPC := FakeDependencyService{err: nil}
			agentDebug := FakeDependencyService{err: nil}
			lags := modelscaling.ModelScalingStatsWrapper{
				Stats:     modelscaling.NewModelReplicaLagsKeeper(),
				Operator:  interfaces.Gte,
				Threshold: 10,
				Reset:     true,
				EventType: modelscaling.ScaleUpEvent,
			}
			lastUsed := modelscaling.ModelScalingStatsWrapper{
				Stats:     modelscaling.NewModelReplicaLastUsedKeeper(),
				Operator:  interfaces.Gte,
				Threshold: 10,
				Reset:     false,
				EventType: modelscaling.ScaleDownEvent,
			}
			modelScalingService := modelscaling.NewStatsAnalyserService(
				[]modelscaling.ModelScalingStatsWrapper{lags, lastUsed}, logger, 10)
			go func() {
				_ = modelScalingService.Start()
			}()
			drainerServicePort, _ := getFreePort()
			drainerService := drainservice.NewDrainerService(logger, uint(drainerServicePort))
			go func() {
				_ = drainerService.Start()
			}()
			client := NewClient(
				NewClientSettings("mlserver", 1, "scheduler", 9002, 9055, period, maxTimeBeforeStart, maxTimeAfterStart),
				logger, modelRepository, v2Client,
				&pb.ReplicaConfig{MemoryBytes: 1000}, "default",
				rpHTTP, rpGRPC, agentDebug, modelScalingService, drainerService, newFakeMetricsHandler())

			mockAgentV2Server := &mockAgentV2Server{}
			conn, err := grpc.DialContext(
				context.Background(), "", grpc.WithTransportCredentials(insecure.NewCredentials()),
				grpc.WithContextDialer(dialerv2(mockAgentV2Server)))
			g.Expect(err).To(BeNil())
			client.conn = conn

			if test.isError {
				go func() {
					time.Sleep(100 * time.Millisecond)
					// induce a failure in one of the sub services
					if test.serviceName == drain {
						_ = drainerService.Stop()
					} else if test.serviceName == scale {
						_ = modelScalingService.Stop()
					} else if test.serviceName == inference {
						httpmock.DeactivateAndReset()
					}
				}()
				err = client.Start()
				g.Expect(err).To(BeNil()) //  we are here it means agent has stopped
				g.Expect(client.stop.Load()).To(BeTrue())
			} else {
				go func() {
					_ = client.Start()
				}()
				time.Sleep(period + maxTimeAfterStart)
				g.Expect(client.stop.Load()).To(BeFalse())
				client.Stop()
			}

			httpmock.DeactivateAndReset()
		})
	}
}
