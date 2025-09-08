/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package agent

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jarcoal/httpmock"
	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/proto"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	pb "github.com/seldonio/seldon-core/apis/go/v2/mlops/agent"
	pbs "github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"
	"github.com/seldonio/seldon-core/components/tls/v2/pkg/tls"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/drainservice"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/interfaces"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/internal/testing_utils"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/k8s"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/k8s/mocks"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/modelscaling"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/modelserver_controlplane/oip"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/readyservice"
	testing_utils2 "github.com/seldonio/seldon-core/scheduler/v2/pkg/internal/testing_utils"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/util"
)

type mockAgentV2Server struct {
	models []string
	pb.UnimplementedAgentServiceServer
	loadedEvents       atomic.Int64
	loadFailedEvents   atomic.Int64
	unloadedEvents     int
	unloadFailedEvents int
	otherEvents        int
	errors             int
	events             []*pb.ModelEventMessage
}

type FakeModelRepository struct {
	err            error
	modelRemovals  int
	modelDownloads int
}

func (f *FakeModelRepository) RemoveModelVersion(modelName string) error {
	f.modelRemovals++
	return nil
}

func (f *FakeModelRepository) GetModelRuntimeInfo(modelName string) (*pbs.ModelRuntimeInfo, error) {
	return &pbs.ModelRuntimeInfo{ModelRuntimeInfo: &pbs.ModelRuntimeInfo_Mlserver{Mlserver: &pbs.MLServerModelSettings{ParallelWorkers: uint32(1)}}}, nil
}

func (f *FakeModelRepository) DownloadModelVersion(modelName string, version uint32, modelSpec *pbs.ModelSpec, config []byte) (*string, error) {
	f.modelDownloads++
	if f.err != nil {
		return nil, f.err
	}
	path := "path"
	return &path, nil
}

func (f *FakeModelRepository) Ready() error {
	return f.err
}

type FakeDependencyService struct {
	name           string
	err            error
	skipErrOnStart bool
	subServiceType interfaces.SubServiceType
}

func (f FakeDependencyService) SetState(state any) {
}

func (f FakeDependencyService) Start() error {
	if f.skipErrOnStart {
		return nil
	}
	return f.err
}

func (f FakeDependencyService) Ready() bool {
	return f.err == nil
}

func (f FakeDependencyService) Stop() error {
	return f.err
}

func (f FakeDependencyService) Name() string {
	return fmt.Sprintf("Fake-%s-Service", f.name)
}

func (f FakeDependencyService) GetType() interfaces.SubServiceType {
	return f.subServiceType
}

func addVerionToModels(models []string, version uint32) []string {
	modelsWithVersions := make([]string, len(models))
	for i := range len(models) {
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
	return int(m.loadedEvents.Load())
}

func (m *mockAgentV2Server) loadFailed() int {
	return int(m.loadFailedEvents.Load())
}

func (m *mockAgentV2Server) AgentEvent(ctx context.Context, message *pb.ModelEventMessage) (*pb.ModelEventResponse, error) {
	switch message.Event {
	case pb.ModelEventMessage_LOADED:
		m.loadedEvents.Add(1)
	case pb.ModelEventMessage_UNLOADED:
		m.unloadedEvents++
	case pb.ModelEventMessage_LOAD_FAILED:
		m.loadFailedEvents.Add(1)
	case pb.ModelEventMessage_UNLOAD_FAILED:
		m.unloadFailedEvents++
	default:
		m.otherEvents++
	}
	m.events = append(m.events, message)
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
	ctx := server.Context()
	<-ctx.Done()

	return nil
}

func createTestV2Client(models []string, status int) interfaces.ModelServerControlPlaneClient {
	v2, _ := testing_utils.CreateTestV2ClientwithState(models, status)
	return v2
}

func TestAgentServiceManagerCreate(t *testing.T) {
	t.Logf("Started")
	logger := log.New()
	log.SetLevel(log.DebugLevel)
	g := NewWithT(t)

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
			httpmock.ActivateNonDefault(v2Client.(*testing_utils.V2RestClientForTest).HttpClient)
			modelRepository := &FakeModelRepository{err: test.modelRepoErr}
			rpHTTP := FakeDependencyService{err: nil}
			rpGRPC := FakeDependencyService{err: nil}
			agentDebug := FakeDependencyService{err: nil}
			modelScalingService := modelscaling.NewStatsAnalyserService(
				[]modelscaling.ModelScalingStatsWrapper{}, logger, 10)
			drainerServicePort, _ := testing_utils2.GetFreePortForTest()
			drainerService := drainservice.NewDrainerService(logger, uint(drainerServicePort))
			readyServicePort, _ := testing_utils2.GetFreePortForTest()
			readinessService := readyservice.NewReadyService(logger, uint(readyServicePort))

			ctrl := gomock.NewController(t)
			k8sExtendedClient := mocks.NewMockExtendedClient(ctrl)
			k8sExtendedClient.EXPECT().HasPublishedIP(gomock.Any(), "mlserver-1", "").Return(nil)

			asm := NewAgentServiceManager(
				NewAgentServiceConfig("mlserver",
					1,
					"scheduler",
					9002,
					9055, 1*time.Minute,
					1*time.Minute,
					1*time.Minute,
					1*time.Minute,
					1*time.Minute, 1, 1, 1, true, tls.TLSOptions{}),
				logger, modelRepository, v2Client,
				test.replicaConfig, "default",
				rpHTTP, rpGRPC, agentDebug, modelScalingService, drainerService, readinessService, newFakeMetricsHandler(), k8sExtendedClient)
			mockAgentV2Server := &mockAgentV2Server{models: test.models}
			conn, err := grpc.NewClient("passthrough://", grpc.WithTransportCredentials(insecure.NewCredentials()),
				grpc.WithContextDialer(dialerv2(mockAgentV2Server)))
			g.Expect(err).To(BeNil())
			asm.schedulerConn = conn
			go func() {
				_ = asm.StartControlLoop()
			}()
			time.Sleep(10 * time.Millisecond)
			if test.v2Status == 200 && test.modelRepoErr == nil {
				g.Eventually(mockAgentV2Server.loaded).Should(BeNumerically(">=", 1))
				g.Eventually(mockAgentV2Server.loadFailed).Should(Equal(0))
			} else {
				g.Eventually(mockAgentV2Server.loaded).Should(Equal(0))
				g.Eventually(mockAgentV2Server.loadFailed).Should(BeNumerically(">=", 1))
			}
			asm.StopControlLoop()
			httpmock.DeactivateAndReset()
		})
	}
}

func TestNotInStartUpPhaseIfSchedulerConnLost(t *testing.T) {
	logger := log.New()
	log.SetLevel(log.DebugLevel)
	g := NewWithT(t)

	v2Client := createTestV2Client(addVerionToModels([]string{"model"}, 0), 200)
	httpmock.ActivateNonDefault(v2Client.(*testing_utils.V2RestClientForTest).HttpClient)
	modelRepository := &FakeModelRepository{err: nil}
	rpHTTP := FakeDependencyService{err: nil}
	rpGRPC := FakeDependencyService{err: nil}
	agentDebug := FakeDependencyService{err: nil}
	modelScalingService := modelscaling.NewStatsAnalyserService(
		[]modelscaling.ModelScalingStatsWrapper{}, logger, 10)
	drainerServicePort, _ := testing_utils2.GetFreePortForTest()
	drainerService := drainservice.NewDrainerService(logger, uint(drainerServicePort))
	readyServicePort, _ := testing_utils2.GetFreePortForTest()
	readinessService := readyservice.NewReadyService(logger, uint(readyServicePort))

	ctrl := gomock.NewController(t)
	k8sExtendedClient := mocks.NewMockExtendedClient(ctrl)
	k8sExtendedClient.EXPECT().HasPublishedIP(gomock.Any(), "mlserver-1", "").Return(nil)

	asm := NewAgentServiceManager(
		NewAgentServiceConfig("mlserver",
			1,
			"scheduler",
			9002,
			9055, 1*time.Minute,
			1*time.Minute,
			1*time.Minute,
			1*time.Minute,
			1*time.Minute, 1, 1, 1, true, tls.TLSOptions{}),
		logger, modelRepository, v2Client,
		&pb.ReplicaConfig{}, "default",
		rpHTTP, rpGRPC, agentDebug, modelScalingService, drainerService, readinessService, newFakeMetricsHandler(), k8sExtendedClient)

	mockAgentV2Server := &mockAgentV2Server{models: []string{"model"}}
	conn, err := grpc.NewClient("passthrough://", grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(dialerv2(mockAgentV2Server)))
	g.Expect(err).To(BeNil())
	asm.schedulerConn = conn

	defer func() {
		asm.StopControlLoop()
		httpmock.DeactivateAndReset()
	}()

	go func() {
		time.Sleep(time.Millisecond * 100)
		// will cause handleSchedulerSubscription below to exit
		conn.Close()
	}()

	g.Expect(asm.isStartup.Load()).To(BeTrue())
	err = asm.handleSchedulerSubscription()
	g.Expect(err).To(BeNil())

	// losing scheduler conn should not cause agent to fail readiness check as it can still serve incoming reqs for
	// models which are already loaded, it just can't load/unload any models at current time.
	g.Expect(asm.isStartup.Load()).To(BeFalse())
}

func TestHandleSchedulerSubscription(t *testing.T) {
	t.Parallel()

	logger := log.New()
	log.SetLevel(log.DebugLevel)
	g := NewWithT(t)

	type test struct {
		name          string
		expectErr     string
		mockK8sClient func(c *mocks.MockExtendedClient)
	}
	tests := []test{
		{
			name: "success - has IP",
			mockK8sClient: func(c *mocks.MockExtendedClient) {
				c.EXPECT().HasPublishedIP(gomock.Any(), "mlserver-1", "").Return(nil)
			},
		},
		{
			name: "failure - IP not yet published to endpoints",
			mockK8sClient: func(c *mocks.MockExtendedClient) {
				c.EXPECT().HasPublishedIP(gomock.Any(), "mlserver-1", "").Return(errors.New("ip not found"))
			},
			expectErr: "failed waiting to check if pod's IP is published to endpoints: ip not found",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			v2Client := createTestV2Client(addVerionToModels([]string{"model"}, 0), 200)
			httpmock.ActivateNonDefault(v2Client.(*testing_utils.V2RestClientForTest).HttpClient)
			modelRepository := &FakeModelRepository{err: nil}
			rpHTTP := FakeDependencyService{err: nil}
			rpGRPC := FakeDependencyService{err: nil}
			agentDebug := FakeDependencyService{err: nil}
			modelScalingService := modelscaling.NewStatsAnalyserService(
				[]modelscaling.ModelScalingStatsWrapper{}, logger, 10)
			drainerServicePort, _ := testing_utils2.GetFreePortForTest()
			drainerService := drainservice.NewDrainerService(logger, uint(drainerServicePort))
			readyServicePort, _ := testing_utils2.GetFreePortForTest()
			readinessService := readyservice.NewReadyService(logger, uint(readyServicePort))

			ctrl := gomock.NewController(t)
			k8sExtendedClient := mocks.NewMockExtendedClient(ctrl)
			test.mockK8sClient(k8sExtendedClient)

			asm := NewAgentServiceManager(
				NewAgentServiceConfig("mlserver",
					1,
					"scheduler",
					9002,
					9055, 1*time.Minute,
					1*time.Minute,
					1*time.Minute,
					1*time.Minute,
					1*time.Minute, 1, 1, 1, true, tls.TLSOptions{}),
				logger, modelRepository, v2Client,
				&pb.ReplicaConfig{}, "default",
				rpHTTP, rpGRPC, agentDebug, modelScalingService, drainerService, readinessService, newFakeMetricsHandler(), k8sExtendedClient)

			mockAgentV2Server := &mockAgentV2Server{models: []string{"model"}}
			conn, err := grpc.NewClient("passthrough://", grpc.WithTransportCredentials(insecure.NewCredentials()),
				grpc.WithContextDialer(dialerv2(mockAgentV2Server)))
			g.Expect(err).To(BeNil())
			asm.schedulerConn = conn

			defer func() {
				asm.StopControlLoop()
				httpmock.DeactivateAndReset()
			}()

			errChan := make(chan error)
			go func() {
				err = asm.handleSchedulerSubscription()
				errChan <- err
			}()

			if test.expectErr != "" {
				err := <-errChan
				g.Expect(err).ToNot(BeNil())
				g.Expect(err.Error()).To(Equal(test.expectErr))
				return
			}

			select {
			case <-time.After(100 * time.Millisecond):
			case err := <-errChan:
				g.Expect(err).To(BeNil())
			}
		})
	}
}

func TestLoadModel(t *testing.T) {
	t.Logf("Started")
	logger := log.New()
	log.SetLevel(log.DebugLevel)
	g := NewWithT(t)

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

	memory500 := uint64(500)
	memory2000 := uint64(2000)

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
						ModelSpec: &pbs.ModelSpec{Uri: "gs://model", MemoryBytes: &memory500, ModelRuntimeInfo: getModelRuntimeInfo(1)},
					},
				},
			},
			replicaConfig:           &pb.ReplicaConfig{MemoryBytes: 1000},
			expectedAvailableMemory: 500,
			v2Status:                200,
			success:                 true,
		}, // Success
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
						ModelSpec: &pbs.ModelSpec{Uri: "gs://model", MemoryBytes: &memory500, ModelRuntimeInfo: getModelRuntimeInfo(1)},
					},
				},
				AutoscalingEnabled: true,
			},
			replicaConfig:           &pb.ReplicaConfig{MemoryBytes: 1000},
			expectedAvailableMemory: 500,
			v2Status:                200,
			success:                 true,
			autoscalingEnabled:      true,
		}, // Success
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
						ModelSpec: &pbs.ModelSpec{Uri: "gs://model", MemoryBytes: &memory500, ModelRuntimeInfo: getModelRuntimeInfo(1)},
					},
				},
			},
			replicaConfig:           &pb.ReplicaConfig{MemoryBytes: 1000},
			expectedAvailableMemory: 1000,
			v2Status:                400,
			success:                 false,
		}, // Fail as V2 fail
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
						ModelSpec: &pbs.ModelSpec{Uri: "gs://model", MemoryBytes: &memory2000, ModelRuntimeInfo: getModelRuntimeInfo(1)},
					},
				},
			},
			replicaConfig:           &pb.ReplicaConfig{MemoryBytes: 1000},
			expectedAvailableMemory: 1000,
			v2Status:                200,
			success:                 false,
		}, // Fail due to too much memory required
	}

	for tidx, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Logf("Test #%d", tidx)

			// Set up dependencies
			ctrl := gomock.NewController(t)
			k8sExtendedClient := mocks.NewMockExtendedClient(ctrl)
			k8sExtendedClient.EXPECT().HasPublishedIP(gomock.Any(), gomock.Any(), "").Return(nil)

			v2Client := createTestV2Client(addVerionToModels(test.models, 0), test.v2Status)
			httpmock.ActivateNonDefault(v2Client.(*testing_utils.V2RestClientForTest).HttpClient)
			modelRepository := &FakeModelRepository{err: test.modelRepoErr}
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

			drainerServicePort, _ := testing_utils2.GetFreePortForTest()
			drainerService := drainservice.NewDrainerService(logger, uint(drainerServicePort))
			readyServicePort, _ := testing_utils2.GetFreePortForTest()
			readinessService := readyservice.NewReadyService(logger, uint(readyServicePort))

			asm := NewAgentServiceManager(
				NewAgentServiceConfig("mlserver", 1, "scheduler",
					9002,
					9055,
					1*time.Minute,
					1*time.Minute,
					1*time.Minute,
					1*time.Minute,
					1*time.Minute, 1, 1, 1, true, tls.TLSOptions{}),
				logger, modelRepository, v2Client, test.replicaConfig, "default",
				rpHTTP, rpGRPC, agentDebug, modelScalingService, drainerService, readinessService, newFakeMetricsHandler(), k8sExtendedClient)

			mockAgentV2Server := &mockAgentV2Server{models: []string{}}
			conn, cerr := grpc.NewClient("passthrough://", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithContextDialer(dialerv2(mockAgentV2Server)))
			g.Expect(cerr).To(BeNil())

			asm.schedulerConn = conn

			go func() {
				err := asm.StartControlLoop()
				// Regardless if this is/isn't a success test case, the client should've started without error
				g.Expect(err).To(BeNil())
			}()

			// Give the client time to start (?)
			time.Sleep(50 * time.Millisecond)

			// Do the actual function call that is being tested
			err := asm.LoadModel(test.op, 1)

			if test.success {
				g.Expect(err).To(BeNil())
				g.Expect(mockAgentV2Server.loaded()).To(Equal(1))
				g.Expect(mockAgentV2Server.loadFailed()).To(Equal(0))
				g.Expect(len(mockAgentV2Server.events)).To(Equal(1))
				g.Expect(mockAgentV2Server.events[0].RuntimeInfo).ToNot(BeNil())
				g.Expect(mockAgentV2Server.events[0].RuntimeInfo.GetMlserver().ParallelWorkers).To(Equal(uint32(1)))
				g.Expect(asm.stateManager.GetAvailableMemoryBytes()).To(Equal(test.expectedAvailableMemory))
				g.Expect(modelRepository.modelRemovals).To(Equal(0))
				loadedVersions := asm.stateManager.modelVersions.getVersionsForAllModels()
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
				g.Expect(mockAgentV2Server.loaded()).To(Equal(0))
				g.Expect(mockAgentV2Server.loadFailed()).To(Equal(1))
				g.Expect(asm.stateManager.GetAvailableMemoryBytes()).To(Equal(test.expectedAvailableMemory))
				g.Expect(modelRepository.modelRemovals).To(Equal(1))
			}
			asm.StopControlLoop()
			httpmock.DeactivateAndReset()
		})
	}
}

func TestLoadModelWithAuth(t *testing.T) {
	t.Logf("Started")
	logger := log.New()
	log.SetLevel(log.DebugLevel)
	g := NewWithT(t)

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

			ctrl := gomock.NewController(t)
			k8sExtendedClient := mocks.NewMockExtendedClient(ctrl)
			k8sExtendedClient.EXPECT().HasPublishedIP(gomock.Any(), gomock.Any(), "").Return(nil)

			v2Client := createTestV2Client(addVerionToModels(test.models, 0), test.v2Status)
			httpmock.ActivateNonDefault(v2Client.(*testing_utils.V2RestClientForTest).HttpClient)
			modelRepository := &FakeModelRepository{}
			rpHTTP := FakeDependencyService{err: nil}
			rpGRPC := FakeDependencyService{err: nil}
			agentDebug := FakeDependencyService{err: nil}
			modelScalingService := modelscaling.NewStatsAnalyserService(
				[]modelscaling.ModelScalingStatsWrapper{}, logger, 10)
			drainerServicePort, _ := testing_utils2.GetFreePortForTest()
			drainerService := drainservice.NewDrainerService(logger, uint(drainerServicePort))
			readyServicePort, _ := testing_utils2.GetFreePortForTest()
			readinessService := readyservice.NewReadyService(logger, uint(readyServicePort))
			asm := NewAgentServiceManager(
				NewAgentServiceConfig(
					"mlserver",
					1,
					"scheduler",
					9002,
					9055,
					1*time.Minute,
					1*time.Minute,
					1*time.Minute,
					1*time.Minute,
					1*time.Minute,
					1, 1, 1, true, tls.TLSOptions{}),
				logger, modelRepository,
				v2Client, test.replicaConfig, "default",
				rpHTTP, rpGRPC, agentDebug, modelScalingService, drainerService, readinessService,
				newFakeMetricsHandler(), k8sExtendedClient)
			switch x := test.op.GetModelVersion().GetModel().GetModelSpec().StorageConfig.Config.(type) {
			case *pbs.StorageConfig_StorageSecretName:
				secret := &v1.Secret{ObjectMeta: metav1.ObjectMeta{Name: x.StorageSecretName, Namespace: asm.namespace}, StringData: map[string]string{"mys3": test.secretData}}
				fakeClientset := fake.NewSimpleClientset(secret)
				s := k8s.NewSecretsHandler(fakeClientset, asm.namespace)
				asm.secretsHandler = s
			}
			mockAgentV2Server := &mockAgentV2Server{models: []string{}}
			conn, cerr := grpc.NewClient("passthrough://", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithContextDialer(dialerv2(mockAgentV2Server)))
			g.Expect(cerr).To(BeNil())
			asm.schedulerConn = conn
			go func() {
				_ = asm.StartControlLoop()
			}()
			// Give the client time to start (?)
			time.Sleep(50 * time.Millisecond)

			err := asm.LoadModel(test.op, 1)
			if test.success {
				g.Expect(err).To(BeNil())
				g.Expect(mockAgentV2Server.loaded()).To(Equal(1))
				g.Expect(mockAgentV2Server.loadFailed()).To(Equal(0))
				g.Expect(asm.stateManager.GetAvailableMemoryBytes()).To(Equal(test.expectedAvailableMemory))
				g.Expect(modelRepository.modelRemovals).To(Equal(0))
			} else {
				g.Expect(err).ToNot(BeNil())
				g.Expect(mockAgentV2Server.loaded()).To(Equal(0))
				g.Expect(modelRepository.modelRemovals).To(Equal(1))
			}
			asm.StopControlLoop()
			httpmock.DeactivateAndReset()
		})
	}
}

func TestUnloadModel(t *testing.T) {
	t.Logf("Started")
	logger := log.New()
	log.SetLevel(log.DebugLevel)
	g := NewWithT(t)

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
				Operation: pb.ModelOperationMessage_UNLOAD_MODEL,
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
			success:                 true,
		}, // Success
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
				Operation: pb.ModelOperationMessage_UNLOAD_MODEL,
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
			success:                 true,
		},
	}

	for tidx, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Logf("Test #%d", tidx)

			ctrl := gomock.NewController(t)
			k8sExtendedClient := mocks.NewMockExtendedClient(ctrl)
			k8sExtendedClient.EXPECT().HasPublishedIP(gomock.Any(), gomock.Any(), "").Return(nil)

			v2Client := createTestV2Client(addVerionToModels(test.models, 0), test.v2Status)
			httpmock.ActivateNonDefault(v2Client.(*testing_utils.V2RestClientForTest).HttpClient)
			modelRepository := &FakeModelRepository{}
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
			drainerServicePort, _ := testing_utils2.GetFreePortForTest()
			drainerService := drainservice.NewDrainerService(logger, uint(drainerServicePort))
			readyServicePort, _ := testing_utils2.GetFreePortForTest()
			readinessService := readyservice.NewReadyService(logger, uint(readyServicePort))
			asm := NewAgentServiceManager(
				NewAgentServiceConfig("mlserver",
					1,
					"scheduler",
					9002,
					9055,
					1*time.Minute,
					1*time.Minute,
					1*time.Minute,
					1*time.Minute,
					1*time.Minute, 1, 1, 1, true, tls.TLSOptions{}),
				logger, modelRepository, v2Client, test.replicaConfig, "default",
				rpHTTP, rpGRPC, agentDebug, modelScalingService, drainerService, readinessService, newFakeMetricsHandler(), k8sExtendedClient)
			mockAgentV2Server := &mockAgentV2Server{models: []string{}}
			conn, cerr := grpc.NewClient("passthrough://", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithContextDialer(dialerv2(mockAgentV2Server)))
			g.Expect(cerr).To(BeNil())
			asm.schedulerConn = conn
			go func() {
				_ = asm.StartControlLoop()
			}()
			// Give the client time to start (?)
			time.Sleep(50 * time.Millisecond)

			err := asm.LoadModel(test.loadOp, 1)
			g.Expect(err).To(BeNil())
			err = asm.UnloadModel(test.unloadOp, 2)
			if test.success {
				g.Expect(err).To(BeNil())
				g.Expect(mockAgentV2Server.loaded()).To(Equal(1))
				g.Expect(mockAgentV2Server.unloadedEvents).To(Equal(1))
				g.Expect(mockAgentV2Server.loadFailed()).To(Equal(0))
				g.Expect(mockAgentV2Server.unloadFailedEvents).To(Equal(0))
				g.Expect(asm.stateManager.GetAvailableMemoryBytes()).To(Equal(test.expectedAvailableMemory))
				// check model scaling stats removed
				versionedModelName := util.GetVersionedModelName(test.unloadOp.GetModelVersion().Model.Meta.Name, test.unloadOp.GetModelVersion().GetVersion())
				_, err := lags.Stats.Get(versionedModelName)
				g.Expect(err).ToNot(BeNil())
				_, err = lastUsed.Stats.Get(versionedModelName)
				g.Expect(err).ToNot(BeNil())
			} else {
				g.Expect(err).ToNot(BeNil())
				g.Expect(mockAgentV2Server.loaded()).To(Equal(1))
				g.Expect(mockAgentV2Server.unloadedEvents).To(Equal(0))
				g.Expect(mockAgentV2Server.loadFailed()).To(Equal(0))
				g.Expect(mockAgentV2Server.unloadFailedEvents).To(Equal(1))
			}
			asm.StopControlLoop()
			httpmock.DeactivateAndReset()
		})
	}
}

func TestAgentServiceManagerClose(t *testing.T) {
	t.Logf("Started")
	logger := log.New()
	log.SetLevel(log.DebugLevel)
	g := NewWithT(t)

	ctrl := gomock.NewController(t)
	k8sExtendedClient := mocks.NewMockExtendedClient(ctrl)

	v2Client := createTestV2Client(nil, 200)
	httpmock.ActivateNonDefault(v2Client.(*testing_utils.V2RestClientForTest).HttpClient)
	defer httpmock.DeactivateAndReset()
	modelRepository := &FakeModelRepository{}
	rpHTTP := FakeDependencyService{err: nil}
	rpGRPC := FakeDependencyService{err: nil}
	agentDebug := FakeDependencyService{err: nil}
	modelScalingService := modelscaling.NewStatsAnalyserService(
		[]modelscaling.ModelScalingStatsWrapper{}, logger, 10)
	drainerServicePort, _ := testing_utils2.GetFreePortForTest()
	drainerService := drainservice.NewDrainerService(logger, uint(drainerServicePort))
	readyServicePort, _ := testing_utils2.GetFreePortForTest()
	readinessService := readyservice.NewReadyService(logger, uint(readyServicePort))
	asm := NewAgentServiceManager(
		NewAgentServiceConfig("mlserver",
			1,
			"scheduler",
			9002,
			9055,
			1*time.Minute,
			1*time.Minute,
			1*time.Minute,
			1*time.Minute,
			1*time.Minute, 1, 1, 1, true, tls.TLSOptions{}),
		logger, modelRepository, v2Client,
		&pb.ReplicaConfig{MemoryBytes: 1000}, "default",
		rpHTTP, rpGRPC, agentDebug, modelScalingService, drainerService, readinessService, newFakeMetricsHandler(), k8sExtendedClient)
	mockAgentV2Server := &mockAgentV2Server{}
	conn, err := grpc.NewClient("passthrough://", grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(dialerv2(mockAgentV2Server)))
	g.Expect(err).To(BeNil())
	asm.schedulerConn = conn

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		err = asm.StartControlLoop()
		g.Expect(err).To(BeNil())
		wg.Done()
	}()
	asm.StopControlLoop()
	wg.Wait()
	g.Expect(asm.schedulerConn.GetState()).To(Equal(connectivity.Shutdown))
	g.Expect(asm.stop.Load()).To(BeTrue())
}

func TestReadinessServiceAgentSync(t *testing.T) {
	t.Logf("Started")
	logger := log.New()
	log.SetLevel(log.DebugLevel)
	g := NewWithT(t)

	serviceStatusCheckEvery := 10 * time.Second
	maxTimeBeforeStart := 2 * time.Second
	maxTimeAfterStart := 500 * time.Millisecond

	ctrl := gomock.NewController(t)
	k8sExtendedClient := mocks.NewMockExtendedClient(ctrl)
	k8sExtendedClient.EXPECT().HasPublishedIP(gomock.Any(), gomock.Any(), "").Return(nil)

	readyServicePort, _ := testing_utils2.GetFreePortForTest()
	readinessService := readyservice.NewReadyService(logger, uint(readyServicePort))
	err := readinessService.Start()
	g.Expect(err).To(BeNil(), "Readiness service shouldn't fail")

	time.Sleep(50 * time.Millisecond)
	readinessHandlerPtr := readinessService.GetHTTPHandler()
	g.Expect(readinessHandlerPtr).ToNot(BeNil(), "Readiness service should have a valid HTTP handler")
	readinessHandler := *readinessHandlerPtr

	req := httptest.NewRequest("GET", "/ready", nil)
	w := httptest.NewRecorder()
	readinessHandler.ServeHTTP(w, req)
	g.Expect(w.Code).To(Equal(http.StatusNotFound), "Readiness service should return 404 before AgentServiceManager is initialized")

	mockMLServer := &testing_utils.MockGRPCMLServer{}
	backEndGRPCPort, err := testing_utils2.GetFreePortForTest()
	if err != nil {
		t.Fatal(err)
	}
	_ = mockMLServer.Setup(uint(backEndGRPCPort))
	go func() {
		_ = mockMLServer.Start()
	}()

	time.Sleep(50 * time.Millisecond)

	v2Client := oip.NewV2Client(
		oip.GetV2ConfigWithDefaults("", backEndGRPCPort), log.New())

	// Set up dependencies
	modelRepository := &FakeModelRepository{}
	rpHTTP := FakeDependencyService{
		name:           "HTTP-proxy",
		subServiceType: (&reverseHTTPProxy{}).GetType(),
	}
	rpGRPC := FakeDependencyService{
		name:           "gRPC-proxy",
		subServiceType: (&reverseGRPCProxy{}).GetType(),
	}
	agentDebug := FakeDependencyService{
		name:           "agentDebug",
		subServiceType: (&AgentDebug{}).GetType(),
	}
	modelScalingService := modelscaling.NewStatsAnalyserService(
		[]modelscaling.ModelScalingStatsWrapper{}, logger, 10)
	drainerServicePort, _ := testing_utils2.GetFreePortForTest()
	drainerService := drainservice.NewDrainerService(logger, uint(drainerServicePort))

	asm := NewAgentServiceManager(
		NewAgentServiceConfig(
			"mlserver",
			1,
			"scheduler",
			9002,
			9055,
			serviceStatusCheckEvery,
			maxTimeBeforeStart,
			maxTimeAfterStart,
			1*time.Minute,
			1*time.Minute,
			1, 1, 1, true, tls.TLSOptions{}),
		logger, modelRepository,
		v2Client,
		&pb.ReplicaConfig{MemoryBytes: 1000}, "default",
		rpHTTP, rpGRPC, agentDebug, modelScalingService, drainerService, readinessService,
		newFakeMetricsHandler(), k8sExtendedClient)
	mockAgentV2Server := &mockAgentV2Server{models: []string{}}
	conn, cerr := grpc.NewClient("passthrough://", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithContextDialer(dialerv2(mockAgentV2Server)))
	g.Expect(cerr).To(BeNil())
	asm.schedulerConn = conn
	w = httptest.NewRecorder()
	readinessHandler.ServeHTTP(w, req)
	g.Expect(w.Code).To(Equal(http.StatusServiceUnavailable), "Readiness service should return StatusServiceUnavailable (503) before AgentServiceManager is ready")

	err = asm.WaitReadySubServices(true)
	g.Expect(err).To(BeNil(), "All sub-services should be ready")
	w = httptest.NewRecorder()
	readinessHandler.ServeHTTP(w, req)
	g.Expect(w.Code).To(Equal(http.StatusServiceUnavailable), "Readiness service should return StatusServiceUnavailable (503) before AgentServiceManager connects to scheduler for the first time")

	wgAsmStopped := sync.WaitGroup{}
	wgAsmStopped.Add(1)
	wgAsmStartupDone := sync.WaitGroup{}
	wgAsmStartupDone.Add(1)

	go func() {
		err = asm.StartControlLoop()
		g.Expect(err).To(BeNil())
		wgAsmStopped.Done()
	}()

	go func() {
		for asm.isStartup.Load() == true {
			time.Sleep(10 * time.Millisecond)
		}
		wgAsmStartupDone.Done()
	}()

	wgAsmStartupDone.Wait()
	w = httptest.NewRecorder()
	readinessHandler.ServeHTTP(w, req)
	// All services are ready and AgentServiceManager has connected to scheduler successfully
	g.Expect(w.Code).To(Equal(http.StatusOK), "Readiness service should return StatusOK (200) after AgentServiceManager is ready and connected to scheduler")

	asm.StopControlLoop()
	wgAsmStopped.Wait()
	w = httptest.NewRecorder()
	readinessHandler.ServeHTTP(w, req)
	g.Expect(w.Code).To(Equal(http.StatusNotFound), "Readiness service should return StatusNotFound (404) after the AgentServiceManager ControlLoop has stopped")

}

func TestAgentReadiness(t *testing.T) {
	t.Logf("Started")
	logger := log.New()
	log.SetLevel(log.DebugLevel)
	g := NewWithT(t)

	serviceStatusCheckEvery := 10 * time.Second
	maxTimeBeforeStart := 2 * time.Second
	maxTimeAfterStart := 500 * time.Millisecond

	type test struct {
		name                 string
		isFirstStart         bool
		withErrorSvcCritical bool
		withErrorSvcAux      bool
		withErrorSvcOptional bool
	}
	baseTests := []test{
		{
			name:         "on-startup",
			isFirstStart: true,
		},
		{
			name:         "periodic-check",
			isFirstStart: false,
		},
	}
	// Generate matrix tests for all possible combinations of withError* flags.
	var tests []test
	for _, baseTest := range baseTests {
		// errCase from 0 to 7, for which the binary representation covers all the possible
		// combinations for the presence of the three types of errors.
		for errCase := range 8 {
			test := baseTest
			test.name = fmt.Sprintf("%s-errCritAuxOpt-%03b", baseTest.name, errCase)
			test.withErrorSvcCritical = (errCase & 0b100) != 0
			test.withErrorSvcAux = (errCase & 0b010) != 0
			test.withErrorSvcOptional = (errCase & 0b001) != 0
			tests = append(tests, test)
		}
	}

	mockMLServer := &testing_utils.MockGRPCMLServer{}
	backEndGRPCPort, err := testing_utils2.GetFreePortForTest()
	if err != nil {
		t.Fatal(err)
	}
	_ = mockMLServer.Setup(uint(backEndGRPCPort))
	go func() {
		_ = mockMLServer.Start()
	}()

	time.Sleep(50 * time.Millisecond)

	v2Client := oip.NewV2Client(
		oip.GetV2ConfigWithDefaults("", backEndGRPCPort), log.New())
	readyServicePort, _ := testing_utils2.GetFreePortForTest()
	readinessService := readyservice.NewReadyService(logger, uint(readyServicePort))
	err = readinessService.Start()
	g.Expect(err).To(BeNil(), "Readiness service shouldn't fail")

	time.Sleep(50 * time.Millisecond)

	// Set up dependencies
	modelRepository := &FakeModelRepository{}
	rpHTTP := FakeDependencyService{
		name:           "HTTP-proxy",
		skipErrOnStart: true,
		subServiceType: (&reverseHTTPProxy{}).GetType(),
	}
	rpGRPC := FakeDependencyService{
		name:           "gRPC-proxy",
		err:            nil,
		subServiceType: (&reverseGRPCProxy{}).GetType(),
	}
	agentDebug := FakeDependencyService{
		name:           "agentDebug",
		skipErrOnStart: true,
		subServiceType: (&AgentDebug{}).GetType(),
	}
	modelScalingService := FakeDependencyService{
		name:           "modelScaling",
		skipErrOnStart: true,
		subServiceType: (&modelscaling.StatsAnalyserService{}).GetType(),
	}
	drainerService := FakeDependencyService{
		name:           "drainer",
		err:            nil,
		subServiceType: (&drainservice.DrainerService{}).GetType(),
	}

	g.Expect(rpHTTP.GetType()).To(Equal(interfaces.CriticalDataPlaneService))
	g.Expect(rpGRPC.GetType()).To(Equal(interfaces.CriticalDataPlaneService))
	g.Expect(agentDebug.GetType()).To(Equal(interfaces.OptionalService))
	g.Expect(modelScalingService.GetType()).To(Equal(interfaces.AuxControlPlaneService))
	g.Expect(drainerService.GetType()).To(Equal(interfaces.CriticalControlPlaneService))

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			k8sExtendedClient := mocks.NewMockExtendedClient(ctrl)
			var maybeCriticalServiceError, maybeOptionalServiceError, maybeAuxServiceError error
			hasErrors := false
			if test.withErrorSvcCritical {
				maybeCriticalServiceError = fmt.Errorf("critical service error")
				hasErrors = true
			}
			if test.withErrorSvcAux {
				maybeAuxServiceError = fmt.Errorf("auxiliary service error")
				hasErrors = true
			}
			if test.withErrorSvcOptional {
				maybeOptionalServiceError = fmt.Errorf("optional service error")
				hasErrors = true
			}
			// Test case specific copies, we do not modify the originals so that we can run
			// tests in parallel
			rpHTTP := rpHTTP
			modelScalingService := modelScalingService
			agentDebug := agentDebug

			rpHTTP.err = maybeCriticalServiceError
			modelScalingService.err = maybeAuxServiceError
			agentDebug.err = maybeOptionalServiceError

			asm := NewAgentServiceManager(
				NewAgentServiceConfig("mlserver",
					1,
					"scheduler",
					9002,
					9055,
					serviceStatusCheckEvery,
					maxTimeBeforeStart,
					maxTimeAfterStart,
					1*time.Minute,
					1*time.Minute, 1, 1, 1, true, tls.TLSOptions{}),
				logger, modelRepository,
				v2Client,
				&pb.ReplicaConfig{MemoryBytes: 1000}, "default",
				rpHTTP, rpGRPC, agentDebug, modelScalingService, drainerService, readinessService,
				newFakeMetricsHandler(), k8sExtendedClient)

			start := time.Now()
			err := asm.WaitReadySubServices(test.isFirstStart)
			endTime := time.Now()
			elapsed := endTime.Sub(start)

			if test.isFirstStart {
				if test.withErrorSvcAux || test.withErrorSvcCritical {
					g.Expect(err).ToNot(BeNil(), "Expected agent error when starting with errors in non-optional subservices")
				} else {
					g.Expect(err).To(BeNil())
				}
			} else {
				if test.withErrorSvcCritical {
					g.Expect(err).ToNot(BeNil(), "Expected agent error when starting with critical errors in non-optional subservices")
				} else {
					g.Expect(err).To(BeNil())
				}
			}

			if hasErrors {
				if test.isFirstStart {
					g.Expect(elapsed.Milliseconds()).To(BeNumerically("~", maxTimeBeforeStart.Milliseconds(), 50))
				} else {
					g.Expect(elapsed.Milliseconds()).To(BeNumerically("~", maxTimeAfterStart.Milliseconds(), 50))
				}
			} else {
				// If there are no errors, we expect the service to start quickly
				g.Expect(elapsed.Milliseconds()).To(BeNumerically("<", 150))
			}
		})
	}
}

func TestAgentStopOnSubServicesFailure(t *testing.T) {
	t.Logf("Started")
	logger := log.New()
	log.SetLevel(log.DebugLevel)
	g := NewWithT(t)

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
			isError:     false, // failure in model scaling service not considered critical
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
			ctrl := gomock.NewController(t)
			k8sExtendedClient := mocks.NewMockExtendedClient(ctrl)
			k8sExtendedClient.EXPECT().HasPublishedIP(gomock.Any(), gomock.Any(), "").Return(nil)

			mockMLServer := &testing_utils.MockGRPCMLServer{}
			backEndGRPCPort, err := testing_utils2.GetFreePortForTest()
			if err != nil {
				t.Fatal(err)
			}
			_ = mockMLServer.Setup(uint(backEndGRPCPort))
			go func() {
				_ = mockMLServer.Start()
			}()

			time.Sleep(50 * time.Millisecond)

			v2Client := oip.NewV2Client(
				oip.GetV2ConfigWithDefaults("", backEndGRPCPort), log.New())

			modelRepository := &FakeModelRepository{}
			rpHTTP := FakeDependencyService{err: nil}
			rpGRPC := FakeDependencyService{err: nil}
			agentDebug := FakeDependencyService{err: nil}
			modelScalingService := modelscaling.NewStatsAnalyserService(
				[]modelscaling.ModelScalingStatsWrapper{}, logger, 10)
			go func() {
				_ = modelScalingService.Start()
			}()
			drainerServicePort, _ := testing_utils2.GetFreePortForTest()
			drainerService := drainservice.NewDrainerService(logger, uint(drainerServicePort))
			go func() {
				_ = drainerService.Start()
			}()
			readyServicePort, _ := testing_utils2.GetFreePortForTest()
			readinessService := readyservice.NewReadyService(logger, uint(readyServicePort))
			asm := NewAgentServiceManager(
				NewAgentServiceConfig("mlserver",
					1,
					"scheduler",
					9002,
					9055,
					period,
					maxTimeBeforeStart,
					maxTimeAfterStart,
					1*time.Minute,
					1*time.Minute, 1, 1, 1, true, tls.TLSOptions{}),
				logger, modelRepository, v2Client,
				&pb.ReplicaConfig{MemoryBytes: 1000}, "default",
				rpHTTP, rpGRPC, agentDebug, modelScalingService, drainerService, readinessService, newFakeMetricsHandler(), k8sExtendedClient)
			mockAgentV2Server := &mockAgentV2Server{}
			conn, err := grpc.NewClient("passthrough://", grpc.WithTransportCredentials(insecure.NewCredentials()),
				grpc.WithContextDialer(dialerv2(mockAgentV2Server)))
			g.Expect(err).To(BeNil())
			asm.schedulerConn = conn

			if test.isError {
				go func() {
					time.Sleep(100 * time.Millisecond)
					// induce a failure in one of the critical sub services
					switch sn := test.serviceName; sn {
					case drain:
						_ = drainerService.Stop()
					case scale:
						_ = modelScalingService.Stop()
					case inference:
						go mockMLServer.Stop()
					}
				}()
				err = asm.StartControlLoop()
				g.Expect(err).To(BeNil()) //  we are here it means agent has stopped
				g.Expect(asm.stop.Load()).To(BeTrue())
			} else {
				go func() {
					_ = asm.StartControlLoop()
				}()
				time.Sleep(period + maxTimeAfterStart)
				g.Expect(asm.stop.Load()).To(BeFalse())
				asm.StopControlLoop()
			}

			go mockMLServer.Stop()
		})
	}
}

func TestUnloadModelOutOfOrder(t *testing.T) {
	t.Logf("Started")
	logger := log.New()
	log.SetLevel(log.DebugLevel)
	g := NewWithT(t)

	type test struct {
		name        string
		models      []string
		loadOp      *pb.ModelOperationMessage
		loadTicks   int64
		unloadOp    *pb.ModelOperationMessage
		unloadTicks int64
		success     bool
	}
	smallMemory := uint64(500)
	tests := []test{
		{
			name:   "in-order",
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
			loadTicks: 1,
			unloadOp: &pb.ModelOperationMessage{
				Operation: pb.ModelOperationMessage_UNLOAD_MODEL,
				ModelVersion: &pb.ModelVersion{
					Model: &pbs.Model{
						Meta: &pbs.MetaData{
							Name: "iris",
						},
						ModelSpec: &pbs.ModelSpec{Uri: "gs://model", MemoryBytes: &smallMemory},
					},
				},
			},
			unloadTicks: 2,
			success:     true,
		},
		{
			name:   "out-of-order",
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
			loadTicks: 2,
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
			unloadTicks: 1,
			success:     false,
		},
	}

	for tidx, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Logf("Test #%d", tidx)

			ctrl := gomock.NewController(t)
			k8sExtendedClient := mocks.NewMockExtendedClient(ctrl)
			k8sExtendedClient.EXPECT().HasPublishedIP(gomock.Any(), gomock.Any(), "").Return(nil)

			v2Client := createTestV2Client(addVerionToModels(test.models, 0), 200)
			httpmock.ActivateNonDefault(v2Client.(*testing_utils.V2RestClientForTest).HttpClient)
			modelRepository := &FakeModelRepository{}
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
			drainerServicePort, _ := testing_utils2.GetFreePortForTest()
			drainerService := drainservice.NewDrainerService(logger, uint(drainerServicePort))
			readyServicePort, _ := testing_utils2.GetFreePortForTest()
			readinessService := readyservice.NewReadyService(logger, uint(readyServicePort))
			asm := NewAgentServiceManager(
				NewAgentServiceConfig("mlserver",
					1,
					"scheduler",
					9002,
					9055,
					1*time.Minute,
					1*time.Minute,
					1*time.Minute,
					1*time.Minute,
					1*time.Minute, 1, 1, 1, true, tls.TLSOptions{}),
				logger, modelRepository, v2Client, &pb.ReplicaConfig{MemoryBytes: 1000}, "default",
				rpHTTP, rpGRPC, agentDebug, modelScalingService, drainerService, readinessService, newFakeMetricsHandler(), k8sExtendedClient)
			mockAgentV2Server := &mockAgentV2Server{models: []string{}}
			conn, cerr := grpc.NewClient("passthrough://", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithContextDialer(dialerv2(mockAgentV2Server)))
			g.Expect(cerr).To(BeNil())
			asm.schedulerConn = conn
			go func() {
				_ = asm.StartControlLoop()
			}()
			// Give the client time to start (?)
			time.Sleep(50 * time.Millisecond)

			err := asm.LoadModel(test.loadOp, test.loadTicks)
			g.Expect(err).To(BeNil())
			err = asm.UnloadModel(test.unloadOp, test.unloadTicks)
			g.Expect(err).To(BeNil())
			if test.success {
				g.Expect(mockAgentV2Server.loaded()).To(Equal(1))
				g.Expect(mockAgentV2Server.unloadedEvents).To(Equal(1))
			} else {
				g.Expect(mockAgentV2Server.loaded()).To(Equal(1))
				g.Expect(mockAgentV2Server.unloadedEvents).To(Equal(0))
			}
			asm.StopControlLoop()
			httpmock.DeactivateAndReset()
		})
	}
}
