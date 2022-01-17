package processor

import (
	"testing"

	. "github.com/onsi/gomega"
	"github.com/seldonio/seldon-core/scheduler/apis/mlops/scheduler"
	"github.com/seldonio/seldon-core/scheduler/pkg/envoy/xdscache"
	"github.com/seldonio/seldon-core/scheduler/pkg/store"
	log "github.com/sirupsen/logrus"
)

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
					1: store.NewServerReplica("host1", 8080, 5000, 1, nil, nil, 100, 100, nil, false),
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
					1: store.NewServerReplica("host1", 8080, 5000, 1, nil, nil, 100, 100, nil, false),
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
					1: store.NewServerReplica("host1", 8080, 5000, 1, nil, nil, 100, 100, nil, false),
					2: store.NewServerReplica("host2", 8080, 5000, 1, nil, nil, 100, 100, nil, false),
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
				xdsCache: xdscache.NewSeldonXDSCache(),
			}
			for _, mv := range test.modelVersions {
				inc.updateEnvoyForModelVersion(mv.GetMeta().GetName(), mv, test.server, test.traffic)
			}

			g.Expect(len(inc.xdsCache.Routes)).To(Equal(test.expectedRoutes))
			g.Expect(len(inc.xdsCache.Clusters)).To(Equal(test.expectedClusters))
		})
	}
}
