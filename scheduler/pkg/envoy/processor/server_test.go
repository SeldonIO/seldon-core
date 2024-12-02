package processor

import (
	"context"
	"reflect"
	"slices"
	"strconv"
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
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/internal/testing_utils"
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

	port, err := testing_utils.GetFreePortForTest()
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		err := startAdsServer(ctx, snapCache, uint(port))
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

	err = inc.setListeners()
	g.Expect(err).To(BeNil())

	conn, err := grpc.NewClient(":"+strconv.Itoa(port), grpc.WithTransportCredentials(insecure.NewCredentials()))
	g.Expect(err).To(BeNil())
	defer conn.Close()

	c := client.NewADSClient(ctx, &core.Node{Id: "node_1"}, resource.ClusterType)
	err = c.InitConnect(conn)
	g.Expect(err).To(BeNil())

	t.Run("Test initial fetch with model version 1", testInitialFetch(g, inc, c))
	t.Run("Test update model version to 2", testUpdateModelVersion(g, inc, c))
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
			resp, err := c.Fetch()
			g.Expect(err).To(BeNil())
			actualClusterNames := make([]string, 0)
			for _, r := range resp.Resources {
				cluster := &clusterv3.Cluster{}
				err := anypb.UnmarshalTo(r, cluster, proto.UnmarshalOptions{})
				g.Expect(err).To(BeNil())
				actualClusterNames = append(actualClusterNames, cluster.Name)
			}
			slices.Sort(actualClusterNames)
			slices.Sort(expectedClusterNames)
			g.Expect(reflect.DeepEqual(actualClusterNames, expectedClusterNames)).To(BeTrue())

			err = c.Ack()
			g.Expect(err).To(BeNil())
			wg.Done()
		}()

		wg.Wait()
	}
}

func testUpdateModelVersion(g *WithT, inc *IncrementalProcessor, c client.ADSClient) func(t *testing.T) {
	expectedFirstResponse := []string{"pipelinegateway_http", "pipelinegateway_grpc", "mirror_http", "mirror_grpc", "model_1_grpc", "model_1_http"}

	return func(t *testing.T) {
		wg := sync.WaitGroup{}
		wg.Add(2)

		go func() {
			resp, err := c.Fetch()
			g.Expect(err).To(BeNil())
			actualClusterNames := make([]string, 0)
			for _, r := range resp.Resources {
				cluster := &clusterv3.Cluster{}
				err := anypb.UnmarshalTo(r, cluster, proto.UnmarshalOptions{})
				g.Expect(err).To(BeNil())
				actualClusterNames = append(actualClusterNames, cluster.Name)
			}
			slices.Sort(actualClusterNames)
			slices.Sort(expectedFirstResponse)
			g.Expect(reflect.DeepEqual(expectedFirstResponse, expectedFirstResponse)).To(BeTrue())

			err = c.Ack()
			g.Expect(err).To(BeNil())
			wg.Done()
		}()

		expectedSecondResponse := slices.Concat(expectedFirstResponse, []string{"model_2_grpc", "model_2_http"})

		ops := []func(inc *IncrementalProcessor, g *WithT){
			createTestModel("model", "server", 1, []int{0}, 2, []store.ModelReplicaState{store.Available}),
		}

		for _, op := range ops {
			op(inc, g)
		}

		go func() {
			resp, err := c.Fetch()
			g.Expect(err).To(BeNil())
			actualClusterNames := make([]string, 0)
			for _, r := range resp.Resources {
				cluster := &clusterv3.Cluster{}
				err := anypb.UnmarshalTo(r, cluster, proto.UnmarshalOptions{})
				g.Expect(err).To(BeNil())
				actualClusterNames = append(actualClusterNames, cluster.Name)
			}
			slices.Sort(actualClusterNames)
			slices.Sort(expectedSecondResponse)
			g.Expect(reflect.DeepEqual(actualClusterNames, expectedSecondResponse)).To(BeTrue())

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
