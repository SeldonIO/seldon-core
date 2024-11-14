package processor

import (
	"context"
	"sync"
	"testing"
	"time"

	clusterv3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	client "github.com/envoyproxy/go-control-plane/pkg/client/sotw/v3"
	resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	envoyServerControlPlaneV3 "github.com/envoyproxy/go-control-plane/pkg/server/v3"
	. "github.com/onsi/gomega"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/envoy/xdscache"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/scheduler/cleaner"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store/experiment"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store/pipeline"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

func TestFetch(t *testing.T) {
	g := NewGomegaWithT(t)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	snapCache := cache.NewSnapshotCache(true, cache.IDHash{}, nil)
	go func() {
		err := startAdsServer(ctx, snapCache, 18001)
		g.Expect(err).To(BeNil())
	}()

	logger := log.New()
	memoryStore := store.NewMemoryStore(logger, nil, nil)
	pipelineHandler := pipeline.NewPipelineStore(logger, nil, memoryStore)
	inc, err := NewIncrementalProcessor(
		snapCache,
		"test-node",
		logger,
		memoryStore,
		experiment.NewExperimentServer(logger, nil, memoryStore, pipelineHandler),
		pipelineHandler,
		nil,
		&xdscache.PipelineGatewayDetails{},
		cleaner.NewVersionCleaner(memoryStore, logger),
	)
	g.Expect(err).To(BeNil())

	conn, err := grpc.NewClient(":18001", grpc.WithTransportCredentials(insecure.NewCredentials()))
	g.Expect(err).To(BeNil())
	defer conn.Close()

	c := client.NewADSClient(ctx, &core.Node{Id: "node_1"}, resource.ClusterType)
	err = c.InitConnect(conn)
	g.Expect(err).To(BeNil())

	t.Run("Test initial fetch", testInitialFetch(ctx, g, inc, snapCache, c))
	// t.Run("Test next fetch", testNextFetch(ctx, g, inc, snapCache, c))
}

func testInitialFetch(ctx context.Context, g Gomega, inc *IncrementalProcessor, snapCache cache.SnapshotCache, c client.ADSClient) func(t *testing.T) {
	return func(t *testing.T) {
		wg := sync.WaitGroup{}
		wg.Add(1)

		inc.modelUpdate("test")
		go func() {
			// watch for configs
			resp, err := c.Fetch()
			g.Expect(err).To(BeNil())
			for _, r := range resp.Resources {
				cluster := &clusterv3.Cluster{}
				err := anypb.UnmarshalTo(r, cluster, proto.UnmarshalOptions{})
				g.Expect(err).To(BeNil())
			}

			err = c.Ack()
			g.Expect(err).To(BeNil())
			wg.Done()
		}()

		snapshot, err := cache.NewSnapshot("1", map[resource.Type][]types.Resource{
			resource.ClusterType: {
				&clusterv3.Cluster{Name: "cluster_1"},
				&clusterv3.Cluster{Name: "cluster_2"},
				&clusterv3.Cluster{Name: "cluster_3"},
			},
		})
		require.NoError(t, err)

		err = snapshot.Consistent()
		g.Expect(err).To(BeNil())
		err = snapCache.SetSnapshot(ctx, "node_1", snapshot)
		g.Expect(err).To(BeNil())
		wg.Wait()
	}
}

func testNextFetch(ctx context.Context, g Gomega, inc *IncrementalProcessor, snapCache cache.SnapshotCache, c client.ADSClient) func(t *testing.T) {
	return func(t *testing.T) {
		wg := sync.WaitGroup{}
		wg.Add(1)

		go func() {
			// watch for configs
			resp, err := c.Fetch()
			require.NoError(t, err)
			g.Expect(err).To(BeNil())
			// assert.Len(t, resp.Resources, 2)
			for _, r := range resp.Resources {
				cluster := &clusterv3.Cluster{}
				err = anypb.UnmarshalTo(r, cluster, proto.UnmarshalOptions{})
				require.NoError(t, err)
				g.Expect(err).To(BeNil())
			}

			err = c.Ack()
			require.NoError(t, err)
			wg.Done()
		}()

		snapshot, err := cache.NewSnapshot("2", map[resource.Type][]types.Resource{
			resource.ClusterType: {
				&clusterv3.Cluster{Name: "cluster_2"},
				&clusterv3.Cluster{Name: "cluster_4"},
			},
		})
		require.NoError(t, err)

		err = snapshot.Consistent()
		g.Expect(err).To(BeNil())
		err = snapCache.SetSnapshot(ctx, "node_1", snapshot)
		g.Expect(err).To(BeNil())
		wg.Wait()
	}
}

func startAdsServer(ctx context.Context, snapCache cache.SnapshotCache, port uint) error {
	logger := log.New()
	// Start envoy xDS server, this is done after the scheduler is ready
	// so that the xDS server can start sending valid updates to envoy.
	srv := envoyServerControlPlaneV3.NewServer(ctx, snapCache, nil)
	xdsServer := NewXDSServer(srv, logger)
	err := xdsServer.StartXDSServer(port)
	if err != nil {
		log.WithError(err).Fatalf("Failed to start envoy xDS server")
	}

	return err
}
