/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package processor

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"testing"
	"time"

	clusterv3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	endpointv3 "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
	log "github.com/sirupsen/logrus"
	"google.golang.org/protobuf/encoding/protojson"

	pba "github.com/seldonio/seldon-core/apis/go/v2/mlops/agent"
	"github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/coordinator"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/envoy/xdscache"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store/experiment"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store/pipeline"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/util"
)

// Set this flag if you want to regenerate all of the snapshot files.
// It should always default to false.
var generateSnapshots bool = *flag.Bool("generate.envoy.snapshot.files", false, "Regenerate the snapshots of the envoy configs")

const snapshots_directory_name = "snapshots_testdata"

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
					1: store.NewServerReplica("host1", 8080, 5000, 1, store.NewServer("server", false), nil, 100, 100, 0, nil, 100),
				},
			},
			traffic:          100,
			expectedRoutes:   1,
			expectedClusters: 2,
		},
		{
			name: "With one replica unloading",
			modelVersions: []*store.ModelVersion{
				store.NewModelVersion(
					&scheduler.Model{
						Meta:           &scheduler.MetaData{Name: "foo"},
						DeploymentSpec: &scheduler.DeploymentSpec{LogPayloads: false},
					},
					2,
					"server",
					map[int]store.ReplicaStatus{
						1: {State: store.Loaded},
						2: {State: store.UnloadEnvoyRequested},
					},
					false,
					store.ModelAvailable,
				),
			},
			server: &store.ServerSnapshot{
				Name: "server",
				Replicas: map[int]*store.ServerReplica{
					1: store.NewServerReplica("host1", 8080, 5000, 1, store.NewServer("server", false), nil, 100, 100, 0, nil, 100),
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
					1: store.NewServerReplica("host1", 8080, 5000, 1, store.NewServer("server", false), nil, 100, 100, 0, nil, 100),
				},
			},
			traffic:          100,
			expectedRoutes:   2,
			expectedClusters: 4,
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
					1: store.NewServerReplica("host1", 8080, 5000, 1, store.NewServer("server", false), nil, 100, 100, 0, nil, 100),
					2: store.NewServerReplica("host2", 8080, 5000, 1, store.NewServer("server", false), nil, 100, 100, 0, nil, 100),
				},
			},
			traffic:          100,
			expectedRoutes:   2,
			expectedClusters: 4,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			xdsCache, err := xdscache.NewSeldonXDSCache(log.New(), &xdscache.PipelineGatewayDetails{Host: "pipeline", GrpcPort: 1, HttpPort: 2}, nil)
			g.Expect(err).To(BeNil())
			inc := IncrementalProcessor{
				logger:   log.New(),
				xdsCache: xdsCache,
			}
			for _, mv := range test.modelVersions {
				inc.updateEnvoyForModelVersion(mv.GetMeta().GetName(), mv, test.server, test.traffic, false)
			}

			g.Expect(inc.xdsCache.Routes.Length()).To(Equal(test.expectedRoutes))
			g.Expect(inc.xdsCache.Clusters.Length()).To(Equal(test.expectedClusters))
		})
	}
}

func TestRollingUpdate(t *testing.T) {
	g := NewGomegaWithT(t)
	type test struct {
		name                string
		ops                 []func(proc *IncrementalProcessor, g *WithT)
		numExpectedClusters int
		numExpectedRoutes   int
		numTrafficSplits    map[string]int
	}

	tests := []test{
		{
			name: "Rolling update in progress",
			ops: []func(inc *IncrementalProcessor, g *WithT){
				createTestServer("server", 2),
				createTestModel("model", "server", 1, []int{0}, 1, []store.ModelReplicaState{store.Available}),
				createTestModel("model", "server", 2, []int{0, 1}, 2, []store.ModelReplicaState{store.Available, store.Loading}),
			},
			numExpectedClusters: 4,
			numExpectedRoutes:   1,
			numTrafficSplits:    map[string]int{"model": 2},
		},
		{
			name: "Rolling update complete",
			ops: []func(inc *IncrementalProcessor, g *WithT){
				createTestServer("server", 1),
				createTestModel("model", "server", 1, []int{0}, 1, []store.ModelReplicaState{store.Available}),
				createTestModel("model", "server", 1, []int{0}, 2, []store.ModelReplicaState{store.Available}),
			},
			numExpectedClusters: 2,
			numExpectedRoutes:   1,
			numTrafficSplits:    map[string]int{"model": 1},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			modelStore := store.NewMemoryStore(log.New(), store.NewLocalSchedulerStore(), nil)
			xdsCache, err := xdscache.NewSeldonXDSCache(log.New(), &xdscache.PipelineGatewayDetails{Host: "pipeline", GrpcPort: 1, HttpPort: 2}, nil)
			g.Expect(err).To(BeNil())
			inc := &IncrementalProcessor{
				logger:           log.New(),
				xdsCache:         xdsCache,
				modelStore:       modelStore,
				experimentServer: experiment.NewExperimentServer(log.New(), nil, nil, nil),
				pipelineHandler:  pipeline.NewPipelineStore(log.New(), nil, modelStore),
			}
			for _, op := range test.ops {
				op(inc, g)
			}
			g.Expect(inc.xdsCache.Clusters.Length()).To(Equal(test.numExpectedClusters))
			g.Expect(inc.xdsCache.Routes.Length()).To(Equal(test.numExpectedRoutes))
			for modelName, trafficSplits := range test.numTrafficSplits {
				g.Expect(len(mustFindVal(inc.xdsCache.Routes, modelName).Clusters)).To(Equal(trafficSplits))
			}
		})
	}
}

func mustFindVal[T any](c *util.CountedSyncMap[T], key string) T {
	val, ok := c.Load(key)
	if !ok {
		panic("failed to find key")
	}
	return *val
}

func TestDraining(t *testing.T) {
	g := NewGomegaWithT(t)
	type test struct {
		name                string
		ops                 []func(proc *IncrementalProcessor, g *WithT)
		numExpectedClusters int
		numExpectedRoutes   int
		numTrafficSplits    map[string]int
		expectedModelState  map[string]store.ModelState
	}

	tests := []test{
		{
			name: "Model with draining and available replicas",
			ops: []func(inc *IncrementalProcessor, g *WithT){
				createTestServer("server", 2),
				createTestModel("model", "server", 1, []int{0, 1}, 1, []store.ModelReplicaState{store.Available, store.Draining}),
			},
			numExpectedClusters: 2,
			numExpectedRoutes:   1,
			numTrafficSplits:    map[string]int{"model": 1},
			expectedModelState:  map[string]store.ModelState{"model": store.ModelAvailable},
		},
		{
			name: "Model with draining and loading replicas",
			ops: []func(inc *IncrementalProcessor, g *WithT){
				createTestServer("server", 2),
				createTestModel("model", "server", 1, []int{0, 1}, 1, []store.ModelReplicaState{store.Loading, store.Draining}),
			},
			numExpectedClusters: 2,
			numExpectedRoutes:   1,
			numTrafficSplits:    map[string]int{"model": 1},
			expectedModelState:  map[string]store.ModelState{"model": store.ModelProgressing},
		},
		{
			name: "Model load failed during draining so failed",
			ops: []func(inc *IncrementalProcessor, g *WithT){
				createTestServer("server", 2),
				createTestModel("model", "server", 1, []int{0, 1}, 1, []store.ModelReplicaState{store.LoadFailed, store.Draining}),
			},
			numExpectedClusters: 2,
			numExpectedRoutes:   1,
			numTrafficSplits:    map[string]int{"model": 1},
			expectedModelState:  map[string]store.ModelState{"model": store.ModelFailed},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			modelStore := store.NewMemoryStore(log.New(), store.NewLocalSchedulerStore(), nil)
			xdsCache, err := xdscache.NewSeldonXDSCache(log.New(), &xdscache.PipelineGatewayDetails{Host: "pipeline", GrpcPort: 1, HttpPort: 2}, nil)
			g.Expect(err).To(BeNil())
			inc := &IncrementalProcessor{
				logger:           log.New(),
				xdsCache:         xdsCache,
				modelStore:       modelStore,
				experimentServer: experiment.NewExperimentServer(log.New(), nil, nil, nil),
				pipelineHandler:  pipeline.NewPipelineStore(log.New(), nil, modelStore),
			}
			for _, op := range test.ops {
				op(inc, g)
			}
			g.Expect(inc.xdsCache.Clusters.Length()).To(Equal(test.numExpectedClusters))
			g.Expect(inc.xdsCache.Routes.Length()).To(Equal(test.numExpectedRoutes))
			for modelName, trafficSplits := range test.numTrafficSplits {
				g.Expect(len(mustFindVal(inc.xdsCache.Routes, modelName).Clusters)).To(Equal(trafficSplits))
			}
			for modelName, modelState := range test.expectedModelState {
				model, err := inc.modelStore.GetModel(modelName)
				g.Expect(err).To(BeNil())
				g.Expect(model.GetLatest().ModelState().State).To(Equal(modelState))
			}
		})
	}
}

func TestModelSync(t *testing.T) {
	g := NewGomegaWithT(t)
	type test struct {
		name                 string
		ops                  []func(proc *IncrementalProcessor, g *WithT)
		pendingModelVersions []*pendingModelVersion
		expectedReplicaStats map[string]map[int]store.ModelReplicaState
		expectedModelState   map[string]store.ModelState
	}

	tests := []test{
		{
			name: "test loaded",
			ops: []func(inc *IncrementalProcessor, g *WithT){
				createTestServer("server", 2),
				createTestModel("model", "server", 1, []int{0}, 1, []store.ModelReplicaState{store.Loaded}),
			},
			pendingModelVersions: []*pendingModelVersion{
				{name: "model", version: 1},
			},
			expectedReplicaStats: map[string]map[int]store.ModelReplicaState{"model": {0: store.Available}},
			expectedModelState:   map[string]store.ModelState{"model": store.ModelAvailable},
		},
		{
			name: "test draining",
			ops: []func(inc *IncrementalProcessor, g *WithT){
				createTestServer("server", 2),
				createTestModel("model", "server", 1, []int{0}, 1, []store.ModelReplicaState{store.Draining}),
			},
			expectedReplicaStats: map[string]map[int]store.ModelReplicaState{"model": {0: store.Draining}},
			expectedModelState:   map[string]store.ModelState{"model": store.ModelProgressing},
		},
		{
			name: "test draining multiple replicas with other loaded",
			ops: []func(inc *IncrementalProcessor, g *WithT){
				createTestServer("server", 2),
				createTestModel("model", "server", 1, []int{0, 1}, 1, []store.ModelReplicaState{store.Draining, store.Loaded}),
			},
			expectedReplicaStats: map[string]map[int]store.ModelReplicaState{"model": {0: store.Draining, 1: store.Available}},
			expectedModelState:   map[string]store.ModelState{"model": store.ModelAvailable},
		},
		{
			name: "test draining multiple replicas with other loading",
			ops: []func(inc *IncrementalProcessor, g *WithT){
				createTestServer("server", 2),
				createTestModel("model", "server", 1, []int{0, 1}, 1, []store.ModelReplicaState{store.Draining, store.Loading}),
			},
			expectedReplicaStats: map[string]map[int]store.ModelReplicaState{"model": {0: store.Draining, 1: store.Loading}},
			expectedModelState:   map[string]store.ModelState{"model": store.ModelProgressing},
		},
		{
			name: "loaded unavailable turns to available",
			ops: []func(inc *IncrementalProcessor, g *WithT){
				createTestServer("server", 2),
				createTestModel("model", "server", 1, []int{0}, 1, []store.ModelReplicaState{store.LoadedUnavailable}),
			},
			expectedReplicaStats: map[string]map[int]store.ModelReplicaState{"model": {0: store.Available}},
			expectedModelState:   map[string]store.ModelState{"model": store.ModelAvailable},
		},
		{
			name: "load failed",
			ops: []func(inc *IncrementalProcessor, g *WithT){
				createTestServer("server", 2),
				createTestModel("model", "server", 1, []int{0}, 1, []store.ModelReplicaState{store.LoadFailed}),
			},
			expectedReplicaStats: map[string]map[int]store.ModelReplicaState{"model": {0: store.LoadFailed}},
			expectedModelState:   map[string]store.ModelState{"model": store.ModelFailed},
		},
		{
			name: "loading - 1 of 2 replicas loaded",
			ops: []func(inc *IncrementalProcessor, g *WithT){
				createTestServer("server", 2),
				createTestModel("model", "server", 2, []int{0, 1}, 1, []store.ModelReplicaState{store.Loaded, store.Loading}),
			},
			expectedReplicaStats: map[string]map[int]store.ModelReplicaState{"model": {0: store.Available, 1: store.Loading}},
			expectedModelState:   map[string]store.ModelState{"model": store.ModelProgressing},
		},
		{
			name: "load failed on 1 replica",
			ops: []func(inc *IncrementalProcessor, g *WithT){
				createTestServer("server", 2),
				createTestModel("model", "server", 2, []int{0, 1}, 1, []store.ModelReplicaState{store.LoadFailed, store.Available}),
			},
			expectedReplicaStats: map[string]map[int]store.ModelReplicaState{"model": {0: store.LoadFailed, 1: store.Available}},
			expectedModelState:   map[string]store.ModelState{"model": store.ModelFailed},
		},
		{
			name: "unload failed on 1 replica",
			ops: []func(inc *IncrementalProcessor, g *WithT){
				createTestServer("server", 2),
				createTestModel("model", "server", 1, []int{0, 1}, 1, []store.ModelReplicaState{store.UnloadFailed, store.Available}),
			},
			expectedReplicaStats: map[string]map[int]store.ModelReplicaState{"model": {0: store.UnloadFailed, 1: store.Available}},
			expectedModelState:   map[string]store.ModelState{"model": store.ModelAvailable},
		},
		{
			name: "UnloadEnvoyRequest - model being deleted",
			ops: []func(inc *IncrementalProcessor, g *WithT){
				createTestServer("server", 2),
				createTestModel("model", "server", 0, []int{0}, 1, []store.ModelReplicaState{store.UnloadEnvoyRequested}),
			},
			expectedReplicaStats: map[string]map[int]store.ModelReplicaState{"model": {0: store.UnloadRequested}},
			// note: model state removed here as this case can only happen when model is deleted, which we cannot simulate in this test.
			expectedModelState: map[string]store.ModelState{},
		},
		{
			name: "UnloadEnvoyRequest - model available",
			ops: []func(inc *IncrementalProcessor, g *WithT){
				createTestServer("server", 2),
				createTestModel("model", "server", 1, []int{0, 1}, 1, []store.ModelReplicaState{store.UnloadEnvoyRequested, store.Available}),
			},
			expectedReplicaStats: map[string]map[int]store.ModelReplicaState{"model": {0: store.UnloadRequested, 1: store.Available}},
			expectedModelState:   map[string]store.ModelState{"model": store.ModelAvailable},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			modelStore := store.NewMemoryStore(log.New(), store.NewLocalSchedulerStore(), nil)
			xdsCache, err := xdscache.NewSeldonXDSCache(log.New(), &xdscache.PipelineGatewayDetails{Host: "pipeline", GrpcPort: 1, HttpPort: 2}, nil)
			g.Expect(err).To(BeNil())
			inc := &IncrementalProcessor{
				logger:               log.New(),
				xdsCache:             xdsCache,
				modelStore:           modelStore,
				experimentServer:     experiment.NewExperimentServer(log.New(), nil, nil, nil),
				pipelineHandler:      pipeline.NewPipelineStore(log.New(), nil, modelStore),
				pendingModelVersions: test.pendingModelVersions,
			}
			for _, op := range test.ops {
				op(inc, g)
			}
			inc.modelSyncWithLock()
			for modelName, modelReplicas := range test.expectedReplicaStats {
				model, err := inc.modelStore.GetModel(modelName)
				g.Expect(err).To(BeNil())
				for replicaIdx, replicaState := range modelReplicas {
					g.Expect(model.GetLatest().ReplicaState()[replicaIdx].State).To(Equal(replicaState))
				}
			}
			for modelName, modelState := range test.expectedModelState {
				model, err := inc.modelStore.GetModel(modelName)
				g.Expect(err).To(BeNil())
				g.Expect(model.GetLatest().ModelState().State).To(Equal(modelState))
			}
		})
	}
}

func TestEnvoySettings(t *testing.T) {
	g := NewGomegaWithT(t)
	// Disable max comparison length, because the json output is quite large
	format.MaxLength = 0
	type test struct {
		name                     string
		ops                      []func(proc *IncrementalProcessor, g *WithT)
		numExpectedClusters      int
		numExpectedRoutes        int
		numExpectedPipelines     int
		experimentActive         bool
		experimentExists         bool
		experimentDeleted        bool
		expectedVersionsInRoutes map[string]uint32
		snapshotFilename         string
	}

	getStrPtr := func(t string) *string { return &t }
	tests := []test{
		{
			name: "experiment with deleted model",
			ops: []func(inc *IncrementalProcessor, g *WithT){
				createTestServer("server", 2),
				createTestModel("model1", "server", 1, []int{0}, 1, []store.ModelReplicaState{store.Available}),
				createTestModel("model2", "server", 1, []int{1}, 1, []store.ModelReplicaState{store.Available}),
				createTestExperiment("exp", []string{"model1", "model2"}, getStrPtr("model1"), nil),
				removeTestModel("model2", 1, "server", 1),
			},
			numExpectedClusters: 2, // model2 should be removed from the clusters (server 1)
			numExpectedRoutes:   2, // model2 should be removed from the routes
			experimentActive:    false,
			experimentExists:    true,
			snapshotFilename:    "experiment-deleted-model",
		},
		{
			name: "One model",
			ops: []func(inc *IncrementalProcessor, g *WithT){
				createTestServer("server", 1),
				createTestModel("model", "server", 1, []int{0}, 1, []store.ModelReplicaState{store.Available}),
			},
			numExpectedClusters: 2,
			numExpectedRoutes:   1,
			snapshotFilename:    "one-model",
		},
		{
			name: "two models",
			ops: []func(inc *IncrementalProcessor, g *WithT){
				createTestServer("server", 2),
				createTestModel("model1", "server", 1, []int{0}, 1, []store.ModelReplicaState{store.Available}),
				createTestModel("model2", "server", 1, []int{1}, 1, []store.ModelReplicaState{store.Available}),
			},
			numExpectedClusters: 4,
			numExpectedRoutes:   2,
			snapshotFilename:    "two-models",
		},
		{
			name: "three models - 1 unloading",
			ops: []func(inc *IncrementalProcessor, g *WithT){
				createTestServer("server", 2),
				createTestModel("model1", "server", 1, []int{0}, 1, []store.ModelReplicaState{store.Available}),
				createTestModel("model2", "server", 1, []int{1}, 1, []store.ModelReplicaState{store.Available}),
				createTestModel("model3", "server", 1, []int{1}, 1, []store.ModelReplicaState{store.Unloading}),
			},
			numExpectedClusters: 4,
			numExpectedRoutes:   2,
			snapshotFilename:    "three-models",
		},
		{
			name: "experiment",
			ops: []func(inc *IncrementalProcessor, g *WithT){
				createTestServer("server", 2),
				createTestModel("model1", "server", 1, []int{0}, 1, []store.ModelReplicaState{store.Available}),
				createTestModel("model2", "server", 1, []int{1}, 1, []store.ModelReplicaState{store.Available}),
				createTestExperiment("exp", []string{"model1", "model2"}, getStrPtr("model1"), nil),
			},
			numExpectedClusters: 4,
			numExpectedRoutes:   3,
			experimentActive:    true,
			experimentExists:    true,
			snapshotFilename:    "experiment",
		},
		{
			name: "experiment - no default",
			ops: []func(inc *IncrementalProcessor, g *WithT){
				createTestServer("server", 2),
				createTestModel("model1", "server", 1, []int{0}, 1, []store.ModelReplicaState{store.Available}),
				createTestModel("model2", "server", 1, []int{1}, 1, []store.ModelReplicaState{store.Available}),
				createTestExperiment("exp", []string{"model1", "model2"}, nil, nil),
			},
			numExpectedClusters: 4,
			numExpectedRoutes:   3,
			experimentActive:    true,
			experimentExists:    true,
			expectedVersionsInRoutes: map[string]uint32{
				"model1": 1,
				"model2": 1,
			},
			snapshotFilename: "experiment-no-default",
		},
		{
			name: "experiment - new model version",
			ops: []func(inc *IncrementalProcessor, g *WithT){
				createTestServer("server", 2),
				createTestModel("model1", "server", 1, []int{0}, 1, []store.ModelReplicaState{store.Available}),
				createTestModel("model2", "server", 1, []int{1}, 1, []store.ModelReplicaState{store.Available}),
				createTestExperiment("exp", []string{"model1", "model2"}, nil, nil),
				// update model2 to version 2, will trigger change in routes / experiment
				createTestModel("model2", "server", 1, []int{1}, 2, []store.ModelReplicaState{store.Available}),
			},
			numExpectedClusters: 4,
			numExpectedRoutes:   3,
			experimentActive:    true,
			experimentExists:    true,
			expectedVersionsInRoutes: map[string]uint32{
				"model1": 1,
				"model2": 2,
			},
			snapshotFilename: "experiment-new-model-version",
		},

		{
			name: "delete experiment",
			ops: []func(inc *IncrementalProcessor, g *WithT){
				createTestServer("server", 2),
				createTestModel("model1", "server", 1, []int{0}, 1, []store.ModelReplicaState{store.Available}),
				createTestModel("model2", "server", 1, []int{1}, 1, []store.ModelReplicaState{store.Available}),
				createTestExperiment("exp", []string{"model1", "model2"}, getStrPtr("model1"), nil),
				deleteTestExperiment("exp"),
			},
			numExpectedClusters: 4,
			numExpectedRoutes:   2,
			experimentExists:    true, // exists but not active
			experimentDeleted:   true,
			snapshotFilename:    "delete-experiment",
		},
		{
			name: "mirror",
			ops: []func(inc *IncrementalProcessor, g *WithT){
				createTestServer("server", 2),
				createTestModel("model1", "server", 1, []int{0}, 1, []store.ModelReplicaState{store.Available}),
				createTestModel("model2", "server", 1, []int{1}, 1, []store.ModelReplicaState{store.Available}),
				createTestExperiment("exp", []string{"model1"}, getStrPtr("model1"), getStrPtr("model2")),
			},
			numExpectedClusters: 4,
			numExpectedRoutes:   3,
			experimentActive:    true,
			experimentExists:    true,
			snapshotFilename:    "mirror",
		},
		{
			name: "mirror, deleted model",
			ops: []func(inc *IncrementalProcessor, g *WithT){
				createTestServer("server", 2),
				createTestModel("model1", "server", 1, []int{0}, 1, []store.ModelReplicaState{store.Available}),
				createTestModel("model2", "server", 1, []int{1}, 1, []store.ModelReplicaState{store.Available}),
				createTestExperiment("exp", []string{"model1"}, getStrPtr("model1"), getStrPtr("model2")),
				removeTestModel("model2", 1, "server", 1),
			},
			numExpectedClusters: 2, // model2 should be removed from the clusters (server 1)
			numExpectedRoutes:   2, // model2 should be removed from the routes
			experimentActive:    false,
			experimentExists:    true,
			snapshotFilename:    "mirror-deleted-model",
		},
		{
			name: "experiment with candidate and mirror",
			ops: []func(inc *IncrementalProcessor, g *WithT){
				createTestServer("server", 2),
				createTestModel("model1", "server", 1, []int{0}, 1, []store.ModelReplicaState{store.Available}),
				createTestModel("model2", "server", 1, []int{1}, 1, []store.ModelReplicaState{store.Available}),
				createTestModel("model3", "server", 1, []int{1}, 1, []store.ModelReplicaState{store.Available}),
				createTestExperiment("exp", []string{"model1", "model2"}, getStrPtr("model1"), getStrPtr("model3")),
			},
			numExpectedClusters: 6,
			numExpectedRoutes:   4,
			experimentActive:    true,
			experimentExists:    true,
			snapshotFilename:    "experiment-candidate-mirror",
		},
		{
			name: "pipeline",
			ops: []func(inc *IncrementalProcessor, g *WithT){
				createTestServer("server", 2),
				createTestModel("model1", "server", 1, []int{0}, 1, []store.ModelReplicaState{store.Available}),
				createTestModel("model2", "server", 1, []int{1}, 1, []store.ModelReplicaState{store.Available}),
				createTestModel("model3", "server", 1, []int{1}, 1, []store.ModelReplicaState{store.Available}),
				createTestPipeline("pipe", []string{"model1", "model2", "model3"}, 1),
			},
			numExpectedClusters:  6,
			numExpectedRoutes:    3,
			numExpectedPipelines: 1,
			snapshotFilename:     "pipeline",
		},
		{
			name: "pipeline with removed model",
			ops: []func(inc *IncrementalProcessor, g *WithT){
				createTestServer("server", 2),
				createTestModel("model1", "server", 1, []int{0}, 1, []store.ModelReplicaState{store.Available}),
				createTestModel("model2", "server", 1, []int{1}, 1, []store.ModelReplicaState{store.Available}),
				createTestModel("model3", "server", 1, []int{1}, 1, []store.ModelReplicaState{store.Available}),
				createTestPipeline("pipe", []string{"model1", "model2", "model3"}, 1),
				removeTestModel("model2", 1, "server", 1),
			},
			numExpectedClusters:  4,
			numExpectedRoutes:    2, // model2 should be removed from the routes
			numExpectedPipelines: 1, // route to pipeline is till there
			snapshotFilename:     "removed-model",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			logger := log.New()
			eventHub, _ := coordinator.NewEventHub(logger)
			memoryStore := store.NewMemoryStore(log.New(), store.NewLocalSchedulerStore(), eventHub)
			xdsCache, err := xdscache.NewSeldonXDSCache(log.New(), &xdscache.PipelineGatewayDetails{Host: "pipeline", GrpcPort: 1, HttpPort: 2}, nil)
			g.Expect(err).To(BeNil())
			inc := &IncrementalProcessor{
				logger:           logger.WithField("source", "IncrementalProcessor"),
				xdsCache:         xdsCache,
				modelStore:       memoryStore,
				experimentServer: experiment.NewExperimentServer(log.New(), eventHub, memoryStore, nil),
				pipelineHandler:  pipeline.NewPipelineStore(log.New(), eventHub, memoryStore),
			}
			eventHub.RegisterModelEventHandler(
				modelEventHandlerName,
				pendingSyncsQueueSize,
				inc.logger,
				inc.handleModelEvents,
			)
			eventHub.RegisterExperimentEventHandler(
				experimentEventHandlerName,
				pendingSyncsQueueSize,
				inc.logger,
				inc.handleExperimentEvents,
			)
			eventHub.RegisterPipelineEventHandler(
				pipelineEventHandlerName,
				pendingSyncsQueueSize,
				inc.logger,
				inc.handlePipelinesEvents,
			)

			for _, op := range test.ops {
				op(inc, g)
				time.Sleep(50 * time.Millisecond) // to allow event handlers to process
			}

			// clusters won't be removed until the next time updateEnvoy is called
			err = inc.updateEnvoy()
			g.Expect(err).To(BeNil())

			g.Expect(inc.xdsCache.Clusters.Length()).To(Equal(test.numExpectedClusters))
			g.Expect(inc.xdsCache.Routes.Length()).To(Equal(test.numExpectedRoutes))
			g.Expect(inc.xdsCache.Pipelines.Length()).To(Equal(test.numExpectedPipelines))

			exp, err := inc.experimentServer.GetExperiment("exp")
			if test.experimentExists {
				g.Expect(err).To(BeNil())
				g.Expect(exp).NotTo(BeNil())
				g.Expect(exp.Active).To(Equal(test.experimentActive))
				g.Expect(exp.Deleted).To(Equal(test.experimentDeleted))
			} else {
				g.Expect(err).NotTo(BeNil())
				g.Expect(exp).To(BeNil())
			}
			for modelName, version := range test.expectedVersionsInRoutes {

				inc.xdsCache.Routes.Range(func(_ string, route xdscache.Route) bool {
					for _, cluster := range route.Clusters {
						if cluster.ModelName == modelName {
							g.Expect(cluster.ModelVersion).To(Equal(version))
						}
					}
					return true
				})
			}

			// Check snapshots

			routeFilename := test.snapshotFilename + "-routes.json"
			if generateSnapshots {
				createSnapshot(g, inc.xdsCache.RouteResources(), routeFilename)
			}

			resultingRoutes := getResultingRoutes(inc.xdsCache.RouteResources())

			data, err := os.ReadFile(snapshots_directory_name + "/" + routeFilename)
			g.Expect(err).To(BeNil())

			var rawMessages []json.RawMessage
			err = json.Unmarshal(data, &rawMessages)
			g.Expect(err).To(BeNil())

			count := 0
			for _, rawMessage := range rawMessages {
				snapshotRouteConfig := &routev3.RouteConfiguration{}
				err := protojson.Unmarshal(rawMessage, snapshotRouteConfig)
				g.Expect(err).To(BeNil())

				resultingRouteConfig := resultingRoutes[snapshotRouteConfig.Name]
				g.Expect(resultingRouteConfig).To(Not(BeNil()))
				g.Expect(len(snapshotRouteConfig.VirtualHosts)).Should(Equal(1))
				g.Expect(len(resultingRouteConfig.VirtualHosts)).Should(Equal(1))
				snapshotRoutes := getTrafficSplits(snapshotRouteConfig.VirtualHosts[0])
				resultingRoutes := getTrafficSplits(resultingRouteConfig.VirtualHosts[0])
				g.Expect(len(resultingRoutes)).Should(Equal(len(snapshotRoutes)))
				g.Expect(resultingRoutes).Should(ConsistOf(snapshotRoutes))
				count++
			}
			g.Expect(len(resultingRoutes)).To(Equal(count))

			clusterFilename := test.snapshotFilename + "-clusters.json"
			if generateSnapshots {
				createSnapshot(g, inc.xdsCache.ClusterResources(), clusterFilename)
			}

			resultingClusters := getResultingClusters(inc.xdsCache.ClusterResources())

			data, err = os.ReadFile(snapshots_directory_name + "/" + clusterFilename)
			g.Expect(err).To(BeNil())

			err = json.Unmarshal(data, &rawMessages)
			g.Expect(err).To(BeNil())

			count = 0
			for _, rawMessage := range rawMessages {
				snapshotCluster := &clusterv3.Cluster{}
				err := protojson.Unmarshal(rawMessage, snapshotCluster)
				g.Expect(err).To(BeNil())

				resultingCluster := resultingClusters[snapshotCluster.Name]
				g.Expect(resultingCluster).To(Not(BeNil()))
				snapshotEndpoints := getEndpoints(snapshotCluster.LoadAssignment)
				resultingEndpoints := getEndpoints(resultingCluster.LoadAssignment)
				g.Expect(len(resultingEndpoints)).Should(Equal(len(snapshotEndpoints)))
				g.Expect(resultingEndpoints).Should(ConsistOf(snapshotEndpoints))
				count++
			}
			g.Expect(len(resultingClusters)).To(Equal(count))
		})
	}
}

func getResultingClusters(resources []types.Resource) map[string]*clusterv3.Cluster {
	clusters := make(map[string]*clusterv3.Cluster)

	for _, resource := range resources {
		cluster := resource.(*clusterv3.Cluster)
		clusters[cluster.Name] = cluster
	}

	return clusters
}

func getResultingRoutes(resources []types.Resource) map[string]*routev3.RouteConfiguration {
	routes := make(map[string]*routev3.RouteConfiguration)

	for _, resource := range resources {
		route := resource.(*routev3.RouteConfiguration)
		routes[route.Name] = route
	}

	return routes
}

func createSnapshot(g Gomega, resources []types.Resource, filename string) {
	// Use this to regenerate the snapshot files
	jsonData := []byte("[")
	for i, resource := range resources {
		data, err := protojson.Marshal(resource)
		g.Expect(err).To(BeNil())
		jsonData = append(jsonData, data...)
		if i < len(resources)-1 {
			jsonData = append(jsonData, ',')
		}
	}
	jsonData = append(jsonData, ']')

	// Write the JSON data to a file
	file, err := os.Create(snapshots_directory_name + "/" + filename)
	g.Expect(err).To(BeNil())

	defer file.Close()

	_, err = file.Write(jsonData)
	g.Expect(err).To(BeNil())
}

func getTrafficSplits(virtualHost *routev3.VirtualHost) []xdscache.Route {
	trafficSplits := make([]xdscache.Route, 0)

	for _, route := range virtualHost.Routes {
		trafficSplit := xdscache.Route{
			RouteName: route.Name,
			Clusters:  make([]xdscache.TrafficSplit, 0),
		}

		clusterSpecificer := route.GetRoute().GetClusterSpecifier()

		fmt.Printf("%v", clusterSpecificer)

		switch route.GetRoute().GetClusterSpecifier().(type) {
		case *routev3.RouteAction_WeightedClusters:
			weightedClusters := route.GetRoute().GetClusterSpecifier().(*routev3.RouteAction_WeightedClusters)

			for _, weightedCluster := range weightedClusters.WeightedClusters.Clusters {
				trafficSplit.Clusters = append(trafficSplit.Clusters, xdscache.TrafficSplit{
					ModelName:     weightedCluster.Name,
					TrafficWeight: weightedCluster.Weight.Value,
				})
			}
		case *routev3.RouteAction_Cluster:
			cluster := route.GetRoute().GetClusterSpecifier().(*routev3.RouteAction_Cluster)
			trafficSplit.Clusters = append(trafficSplit.Clusters, xdscache.TrafficSplit{
				ModelName:     cluster.Cluster,
				TrafficWeight: 100,
			})

		}

		if len(route.GetRoute().RequestMirrorPolicies) > 0 {
			mirror := route.GetRoute().RequestMirrorPolicies[0]
			trafficSplit.Mirror = &xdscache.TrafficSplit{ModelName: mirror.Cluster, TrafficWeight: mirror.RuntimeFraction.DefaultValue.Numerator}
		}

		trafficSplits = append(trafficSplits, trafficSplit)

	}

	return trafficSplits
}

func getEndpoints(loadAssignment *endpointv3.ClusterLoadAssignment) []xdscache.Endpoint {
	endpoints := make([]xdscache.Endpoint, 0)
	for _, localityLbEndpoint := range loadAssignment.Endpoints {
		for _, lbEndpoint := range localityLbEndpoint.LbEndpoints {
			endpointEndpoint := lbEndpoint.HostIdentifier.(*endpointv3.LbEndpoint_Endpoint)
			endpoints = append(endpoints, xdscache.Endpoint{
				UpstreamHost: endpointEndpoint.Endpoint.Address.GetSocketAddress().Address,
				UpstreamPort: endpointEndpoint.Endpoint.Address.GetSocketAddress().GetPortValue(),
			})
		}
	}
	return endpoints
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

func createTestModel(modelName string,
	serverName string,
	desiredReplicas uint32,
	replicas []int,
	version uint32,
	replicaStates []store.ModelReplicaState,
) func(inc *IncrementalProcessor, g *WithT) {
	f := func(inc *IncrementalProcessor, g *WithT) {
		model := &scheduler.Model{
			Meta: &scheduler.MetaData{
				Name: modelName,
			},
			ModelSpec: &scheduler.ModelSpec{
				Uri: "gs://" + util.CreateRequestId(), // Create a random uri
			},
			DeploymentSpec: &scheduler.DeploymentSpec{
				Replicas: desiredReplicas,
			},
		}
		err := inc.modelStore.UpdateModel(&scheduler.LoadModelRequest{Model: model})
		g.Expect(err).To(BeNil())
		var serverReplicas []*store.ServerReplica
		for _, replicaIdx := range replicas {
			var serverReplica *store.ServerReplica
			server, err := inc.modelStore.GetServer(serverName, false, true)
			g.Expect(err).To(BeNil())
			if server != nil {
				if sr, ok := server.Replicas[replicaIdx]; ok {
					serverReplica = sr
				}
			}
			if serverReplica == nil {
				serverReplica = store.NewServerReplica("", 1, 2, replicaIdx, nil, nil, 1000, 1000, 0, nil, 0)
			}
			serverReplicas = append(serverReplicas, serverReplica)
		}

		// this adds all model replicas as `LoadRequested`
		err = inc.modelStore.UpdateLoadedModels(modelName, version, serverName, serverReplicas)
		g.Expect(err).To(BeNil())

		for idx, replicaIdx := range replicas {
			err = inc.modelStore.UpdateModelState(modelName, version, serverName, replicaIdx, nil, store.LoadRequested, replicaStates[idx], "", nil)
			g.Expect(err).To(BeNil())
		}

		err = inc.modelUpdate(modelName)
		g.Expect(err).To(BeNil())
	}
	return f
}

func removeTestModel(
	modelName string,
	version uint32,
	serverName string,
	serverIdx int,
) func(inc *IncrementalProcessor, g *WithT) {
	f := func(inc *IncrementalProcessor, g *WithT) {
		err := inc.modelStore.RemoveModel(&scheduler.UnloadModelRequest{Model: &scheduler.ModelReference{Name: "model1", Version: &version}})
		g.Expect(err).To(BeNil())
		err = inc.modelStore.UpdateModelState(modelName, version, serverName, serverIdx, nil, store.Available, store.Unloaded, "", nil)
		g.Expect(err).To(BeNil())
	}
	return f
}

func createTestExperiment(experimentName string, modelNames []string, defaultModel *string, mirrorName *string) func(inc *IncrementalProcessor, g *WithT) {
	f := func(inc *IncrementalProcessor, g *WithT) {
		var candidates []*experiment.Candidate
		var mirror *experiment.Mirror
		for _, modelName := range modelNames {
			candidates = append(candidates, &experiment.Candidate{Name: modelName, Weight: 1})
		}
		if mirrorName != nil {
			mirror = &experiment.Mirror{
				Name:    *mirrorName,
				Percent: 100,
			}
		}
		exp := &experiment.Experiment{
			Name:       experimentName,
			Default:    defaultModel,
			Candidates: candidates,
			Mirror:     mirror,
		}
		err := inc.experimentServer.StartExperiment(exp)
		g.Expect(err).To(BeNil())
		err = inc.experimentUpdate(exp)
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

func createTestPipeline(pipelineName string, modelNames []string, version uint32) func(inc *IncrementalProcessor, g *WithT) {
	f := func(inc *IncrementalProcessor, g *WithT) {
		steps := []*scheduler.PipelineStep{}
		for _, modelName := range modelNames {
			steps = append(steps, &scheduler.PipelineStep{
				Name: modelName,
			})
		}
		pipe := &scheduler.Pipeline{
			Name:    pipelineName,
			Version: version,
			Steps:   steps,
			Uid:     "uid",
		}
		err := inc.pipelineHandler.AddPipeline(pipe)
		g.Expect(err).To(BeNil())
		err = inc.pipelineHandler.SetPipelineState(pipelineName, version, "uid", pipeline.PipelineReady, "", "")
		g.Expect(err).To(BeNil())
	}
	return f
}
