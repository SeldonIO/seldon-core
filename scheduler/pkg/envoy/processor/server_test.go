package processor

import (
	"context"
	"slices"
	"sync"
	"testing"
	"time"

	clusterv3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	client "github.com/envoyproxy/go-control-plane/pkg/client/sotw/v3"
	resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	envoyServerControlPlaneV3 "github.com/envoyproxy/go-control-plane/pkg/server/v3"
	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/envoy/xdscache"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store/experiment"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store/pipeline"
)

func TestFetch(t *testing.T) {
	g := NewGomegaWithT(t)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	logger := log.New()

	snapCache := cache.NewSnapshotCache(true, cache.IDHash{}, logger)
	go func() {
		err := startAdsServer(ctx, snapCache, 18001)
		g.Expect(err).To(BeNil())
	}()

	memoryStore := store.NewMemoryStore(logger, store.NewLocalSchedulerStore(), nil)
	pipelineHandler := pipeline.NewPipelineStore(logger, nil, memoryStore)

	inc := &IncrementalProcessor{
		cache:            snapCache,
		logger:           logger,
		xdsCache:         xdscache.NewSeldonXDSCache(log.New(), &xdscache.PipelineGatewayDetails{}),
		pipelineHandler:  pipelineHandler,
		modelStore:       memoryStore,
		experimentServer: experiment.NewExperimentServer(logger, nil, memoryStore, pipelineHandler),
		nodeID:           "node_1",
	}

	err := inc.setListeners()
	g.Expect(err).To(BeNil())

	conn, err := grpc.NewClient(":18001", grpc.WithTransportCredentials(insecure.NewCredentials()))
	g.Expect(err).To(BeNil())
	defer conn.Close()

	c := client.NewADSClient(ctx, &core.Node{Id: "node_1"}, resource.ClusterType)
	err = c.InitConnect(conn)
	g.Expect(err).To(BeNil())

	t.Run("Test initial fetch with model 1", testInitialFetch(g, inc, c))
}

func testInitialFetch(g *WithT, inc *IncrementalProcessor, c client.ADSClient) func(t *testing.T) {
	expectedClusterNames := []string{"pipelinegateway_http", "pipelinegateway_grpc", "mirror_http", "mirror_grpc", "model_1_grpc", "model_1_http"}

	return func(t *testing.T) {
		wg := sync.WaitGroup{}
		wg.Add(1)

		ops := []func(inc *IncrementalProcessor, g *WithT){
			createTestServer("server", 1),
			createTestModel("model", "server", 1, []int{0}, 1, []store.ModelReplicaState{store.Available}),
		}

		for _, op := range ops {
			op(inc, g)
		}

		go func() {
			// watch for configs
			resp, err := c.Fetch()
			g.Expect(err).To(BeNil())
			for _, r := range resp.Resources {
				cluster := &clusterv3.Cluster{}
				err := anypb.UnmarshalTo(r, cluster, proto.UnmarshalOptions{})
				g.Expect(err).To(BeNil())
				g.Expect(slices.Contains(expectedClusterNames, cluster.Name)).To(BeTrue())
			}

			err = c.Ack()
			g.Expect(err).To(BeNil())
			wg.Done()
		}()

		wg.Wait()
	}
}

func startAdsServer(ctx context.Context, snapCache cache.SnapshotCache, port uint) error {
	logger := log.New()
	srv := envoyServerControlPlaneV3.NewServer(ctx, snapCache, nil)
	xdsServer := NewXDSServer(srv, logger)
	err := xdsServer.StartXDSServer(port)
	if err != nil {
		log.WithError(err).Fatalf("Failed to start envoy xDS server")
	}

	return err
}
