package processor

import (
	"fmt"
	"testing"

	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	pba "github.com/seldonio/seldon-core/scheduler/apis/mlops/agent"
	"github.com/seldonio/seldon-core/scheduler/pkg/store/experiment"
	"github.com/seldonio/seldon-core/scheduler/pkg/store/pipeline"

	. "github.com/onsi/gomega"
	"github.com/seldonio/seldon-core/scheduler/apis/mlops/scheduler"
	"github.com/seldonio/seldon-core/scheduler/pkg/envoy/xdscache"
	"github.com/seldonio/seldon-core/scheduler/pkg/store"
	log "github.com/sirupsen/logrus"
)

func TestGetTrafficShare(t *testing.T) {
	g := NewGomegaWithT(t)
	type test struct {
		name                             string
		latestModel                      *store.ModelVersion
		lastAvailableModel               *store.ModelVersion
		weight                           uint32
		expectedLatestModelWeight        uint32
		expectedLastAvailableModelWeight uint32
	}

	tests := []test{
		{
			name: "50 - 50",
			latestModel: store.NewModelVersion(nil, 1, "server", map[int]store.ReplicaStatus{
				1: {
					State: store.Available,
				},
			}, false, store.ModelAvailable),
			lastAvailableModel: store.NewModelVersion(nil, 1, "server", map[int]store.ReplicaStatus{
				1: {
					State: store.Available,
				},
			}, false, store.ModelAvailable),
			weight:                           100,
			expectedLatestModelWeight:        50,
			expectedLastAvailableModelWeight: 50,
		},
		{
			name: "2 latest replicas to 1 last available",
			latestModel: store.NewModelVersion(nil, 1, "server", map[int]store.ReplicaStatus{
				1: {
					State: store.Available,
				},
				2: {
					State: store.Available,
				},
			}, false, store.ModelAvailable),
			lastAvailableModel: store.NewModelVersion(nil, 1, "server", map[int]store.ReplicaStatus{
				1: {
					State: store.Available,
				},
			}, false, store.ModelAvailable),
			weight:                           100,
			expectedLatestModelWeight:        67,
			expectedLastAvailableModelWeight: 33,
		},
		{
			name: "1 latest replicas to 2 last available",
			latestModel: store.NewModelVersion(nil, 1, "server", map[int]store.ReplicaStatus{
				1: {
					State: store.Available,
				},
			}, false, store.ModelAvailable),
			lastAvailableModel: store.NewModelVersion(nil, 1, "server", map[int]store.ReplicaStatus{
				1: {
					State: store.Available,
				},
				2: {
					State: store.Available,
				},
			}, false, store.ModelAvailable),
			weight:                           100,
			expectedLatestModelWeight:        34,
			expectedLastAvailableModelWeight: 66,
		},
		{
			name: "model failed so all to latest",
			latestModel: store.NewModelVersion(nil, 1, "server", map[int]store.ReplicaStatus{
				1: {
					State: store.Available,
				},
				2: {
					State: store.Available,
				},
			}, false, store.ModelAvailable),
			lastAvailableModel: store.NewModelVersion(nil, 1, "server", map[int]store.ReplicaStatus{
				1: {
					State: store.LoadFailed,
				},
			}, false, store.ModelAvailable),
			weight:                           100,
			expectedLatestModelWeight:        100,
			expectedLastAvailableModelWeight: 0,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			lastestModelTraffic, lastAvailableModelTraffic := getTrafficShare(test.latestModel, test.lastAvailableModel, test.weight)
			g.Expect(lastestModelTraffic).To(Equal(test.expectedLatestModelWeight))
			g.Expect(lastAvailableModelTraffic).To(Equal(test.expectedLastAvailableModelWeight))
		})
	}
}

func TestUpdateEnvoyForModelVersion(t *testing.T) {
	g := NewGomegaWithT(t)
	type test struct {
		name             string
		modelVersions    []*store.ModelVersion
		server           *store.ServerSnapshot
		traffic          uint32
		expectedRoutes   int
		expectedClusters int
	}

	tests := []test{
		{
			name: "Simple",
			modelVersions: []*store.ModelVersion{
				store.NewModelVersion(
					&scheduler.Model{
						Meta:           &scheduler.MetaData{Name: "foo"},
						DeploymentSpec: &scheduler.DeploymentSpec{LogPayloads: false},
					},
					1,
					"server",
					map[int]store.ReplicaStatus{
						1: {State: store.Loaded},
					},
					false,
					store.ModelAvailable,
				),
			},
			server: &store.ServerSnapshot{
				Name: "server",
				Replicas: map[int]*store.ServerReplica{
					1: store.NewServerReplica("host1", 8080, 5000, 1, nil, nil, 100, 100, nil, 100),
				},
			},
			traffic:          100,
			expectedRoutes:   1,
			expectedClusters: 2,
		},
		{
			name: "TwoRoutesSameCluster",
			modelVersions: []*store.ModelVersion{
				store.NewModelVersion(
					&scheduler.Model{
						Meta:           &scheduler.MetaData{Name: "foo"},
						DeploymentSpec: &scheduler.DeploymentSpec{LogPayloads: false},
					},
					1,
					"server",
					map[int]store.ReplicaStatus{
						1: {State: store.Loaded},
					},
					false,
					store.ModelAvailable,
				),
				store.NewModelVersion(
					&scheduler.Model{
						Meta:           &scheduler.MetaData{Name: "bar"},
						DeploymentSpec: &scheduler.DeploymentSpec{LogPayloads: false},
					},
					1,
					"server",
					map[int]store.ReplicaStatus{
						1: {State: store.Loaded},
					},
					false,
					store.ModelAvailable,
				),
			},
			server: &store.ServerSnapshot{
				Name: "server",
				Replicas: map[int]*store.ServerReplica{
					1: store.NewServerReplica("host1", 8080, 5000, 1, nil, nil, 100, 100, nil, 100),
				},
			},
			traffic:          100,
			expectedRoutes:   2,
			expectedClusters: 2,
		},
		{
			name: "TwoRoutesDifferentClusters",
			modelVersions: []*store.ModelVersion{
				store.NewModelVersion(
					&scheduler.Model{
						Meta:           &scheduler.MetaData{Name: "foo"},
						DeploymentSpec: &scheduler.DeploymentSpec{LogPayloads: false},
					},
					1,
					"server",
					map[int]store.ReplicaStatus{
						1: {State: store.Loaded},
					},
					false,
					store.ModelAvailable,
				),
				store.NewModelVersion(
					&scheduler.Model{
						Meta:           &scheduler.MetaData{Name: "bar"},
						DeploymentSpec: &scheduler.DeploymentSpec{LogPayloads: false},
					},
					1,
					"server",
					map[int]store.ReplicaStatus{
						2: {State: store.Loaded},
					},
					false,
					store.ModelAvailable,
				),
			},
			server: &store.ServerSnapshot{
				Name: "server",
				Replicas: map[int]*store.ServerReplica{
					1: store.NewServerReplica("host1", 8080, 5000, 1, nil, nil, 100, 100, nil, 100),
					2: store.NewServerReplica("host2", 8080, 5000, 1, nil, nil, 100, 100, nil, 100),
				},
			},
			traffic:          100,
			expectedRoutes:   2,
			expectedClusters: 4,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			inc := IncrementalProcessor{
				logger:   log.New(),
				xdsCache: xdscache.NewSeldonXDSCache(log.New(), &xdscache.PipelineGatewayDetails{}),
			}
			for _, mv := range test.modelVersions {
				inc.updateEnvoyForModelVersion(mv.GetMeta().GetName(), mv, test.server, test.traffic)
			}

			g.Expect(len(inc.xdsCache.Routes)).To(Equal(test.expectedRoutes))
			g.Expect(len(inc.xdsCache.Clusters)).To(Equal(test.expectedClusters))
		})
	}
}

func createTestServer(serverName string, numReplicas uint32) func(inc *IncrementalProcessor, g *WithT) {
	f := func(inc *IncrementalProcessor, g *WithT) {
		for i := uint32(0); i < numReplicas; i++ {
			err := inc.modelStore.AddServerReplica(&pba.AgentSubscribeRequest{
				ServerName: serverName,
				ReplicaIdx: i,
				ReplicaConfig: &pba.ReplicaConfig{
					InferenceSvc:      fmt.Sprintf("%s.%d", serverName, i),
					InferenceHttpPort: 1234,
				},
			})
			g.Expect(err).To(BeNil())
		}
	}
	return f
}

func createTestModel(modelName string, serverName string, replicas []int) func(inc *IncrementalProcessor, g *WithT) {
	f := func(inc *IncrementalProcessor, g *WithT) {
		model := &scheduler.Model{
			Meta: &scheduler.MetaData{
				Name: modelName,
			},
			ModelSpec:      &scheduler.ModelSpec{},
			DeploymentSpec: &scheduler.DeploymentSpec{},
		}
		err := inc.modelStore.UpdateModel(&scheduler.LoadModelRequest{Model: model})
		g.Expect(err).To(BeNil())
		replicaMap := make(map[int]store.ReplicaStatus)
		for _, replicaIdx := range replicas {
			err := inc.modelStore.UpdateLoadedModels(modelName, 1, serverName, []*store.ServerReplica{store.NewServerReplica("", 1, 2, replicaIdx, nil, nil, 1000, 1000, nil, 0)})
			g.Expect(err).To(BeNil())
			replicaMap[replicaIdx] = store.ReplicaStatus{State: store.Available}
			err = inc.modelStore.UpdateModelState(modelName, 1, serverName, replicaIdx, nil, store.LoadRequested, store.Loaded, "")
			g.Expect(err).To(BeNil())
		}
		err = inc.addTraffic(&store.ModelSnapshot{
			Name: modelName,
			Versions: []*store.ModelVersion{
				store.NewModelVersion(model, 1, serverName, replicaMap, false, store.ModelAvailable),
			},
		})
		g.Expect(err).To(BeNil())
	}
	return f
}

func createTestExperiment(experimentName string, modelNames []string, defaultModel *string) func(inc *IncrementalProcessor, g *WithT) {
	f := func(inc *IncrementalProcessor, g *WithT) {
		var candidates []*experiment.Candidate
		for _, modelName := range modelNames {
			candidates = append(candidates, &experiment.Candidate{ModelName: modelName, Weight: 1})
		}
		exp := &experiment.Experiment{
			Name:         experimentName,
			DefaultModel: defaultModel,
			Candidates:   candidates,
		}
		err := inc.experimentServer.StartExperiment(exp)
		g.Expect(err).To(BeNil())
		err = inc.experimentSync(exp)
		g.Expect(err).To(BeNil())
	}
	return f
}

func deleteTestExperiment(experimentName string) func(inc *IncrementalProcessor, g *WithT) {
	f := func(inc *IncrementalProcessor, g *WithT) {
		err := inc.experimentServer.StopExperiment(experimentName)
		g.Expect(err).To(BeNil())
		exp, err := inc.experimentServer.GetExperiment(experimentName)
		g.Expect(err).To(BeNil())
		err = inc.removeExperiment(exp)
		g.Expect(err).To(BeNil())
	}
	return f
}

func TestExperiments(t *testing.T) {
	g := NewGomegaWithT(t)
	type test struct {
		name                string
		ops                 []func(proc *IncrementalProcessor, g *WithT)
		numExpectedClusters int
		numExpectedRoutes   int
	}

	getStrPtr := func(t string) *string { return &t }
	tests := []test{
		{
			name: "One model",
			ops: []func(inc *IncrementalProcessor, g *WithT){
				createTestServer("server", 1),
				createTestModel("model", "server", []int{0}),
			},
			numExpectedClusters: 2,
			numExpectedRoutes:   1,
		},
		{
			name: "two models",
			ops: []func(inc *IncrementalProcessor, g *WithT){
				createTestServer("server", 2),
				createTestModel("model1", "server", []int{0}),
				createTestModel("model2", "server", []int{1}),
			},
			numExpectedClusters: 4,
			numExpectedRoutes:   2,
		},
		{
			name: "experiment",
			ops: []func(inc *IncrementalProcessor, g *WithT){
				createTestServer("server", 2),
				createTestModel("model1", "server", []int{0}),
				createTestModel("model2", "server", []int{1}),
				createTestExperiment("exp", []string{"model1", "model2"}, getStrPtr("model1")),
			},
			numExpectedClusters: 4,
			numExpectedRoutes:   3,
		},
		{
			name: "delete experiment",
			ops: []func(inc *IncrementalProcessor, g *WithT){
				createTestServer("server", 2),
				createTestModel("model1", "server", []int{0}),
				createTestModel("model2", "server", []int{1}),
				createTestExperiment("exp", []string{"model1", "model2"}, getStrPtr("model1")),
				deleteTestExperiment("exp"),
			},
			numExpectedClusters: 4,
			numExpectedRoutes:   2,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			inc := &IncrementalProcessor{
				cache:            cache.NewSnapshotCache(false, cache.IDHash{}, log.New()),
				logger:           log.New(),
				xdsCache:         xdscache.NewSeldonXDSCache(log.New(), &xdscache.PipelineGatewayDetails{Host: "pipeline", GrpcPort: 1, HttpPort: 2}),
				modelStore:       store.NewMemoryStore(log.New(), store.NewLocalSchedulerStore(), nil),
				experimentServer: experiment.NewExperimentServer(log.New(), nil, nil),
				pipelineHandler:  pipeline.NewPipelineStore(log.New(), nil),
			}
			inc.xdsCache.AddListener("listener_0")
			for _, op := range test.ops {
				op(inc, g)
			}
			g.Expect(len(inc.xdsCache.Clusters)).To(Equal(test.numExpectedClusters))
			g.Expect(len(inc.xdsCache.Routes)).To(Equal(test.numExpectedRoutes))
		})
	}

}
