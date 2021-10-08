package processor

import (
	"context"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	rsrc "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/seldonio/seldon-core/scheduler/apis/mesh"
	"github.com/seldonio/seldon-core/scheduler/pkg/envoy/resources"
	"github.com/seldonio/seldon-core/scheduler/pkg/envoy/xdscache"
	"github.com/sirupsen/logrus"
	"math"
	"math/rand"
	"strconv"
)

type IncrementalProcessor struct {
	cache  cache.SnapshotCache
	nodeID string

	// snapshotVersion holds the current version of the snapshot.
	snapshotVersion int64

	logger logrus.FieldLogger

	xdsCache xdscache.SeldonXDSCache
}

func NewIncrementalProcessor(cache cache.SnapshotCache, nodeID string, log logrus.FieldLogger) *IncrementalProcessor {
	return &IncrementalProcessor{
		cache:           cache,
		nodeID:          nodeID,
		snapshotVersion: rand.Int63n(1000),
		logger:   log,
		xdsCache: xdscache.SeldonXDSCache{
			Listeners: make(map[string]resources.Listener),
			Clusters:  make(map[string]resources.Cluster),
			Routes:    make(map[string]resources.Route),
			Endpoints: make(map[string]resources.Endpoint),
		},
	}
}

func (p *IncrementalProcessor) SetListener(listenerName string) {
	p.xdsCache.AddListener(listenerName)
}

// newSnapshotVersion increments the current snapshotVersion
// and returns as a string.
func (p *IncrementalProcessor) newSnapshotVersion() string {

	// Reset the snapshotVersion if it ever hits max size.
	if p.snapshotVersion == math.MaxInt64 {
		p.snapshotVersion = 0
	}

	// Increment the snapshot version & return as string.
	p.snapshotVersion++
	return strconv.FormatInt(p.snapshotVersion, 10)
}

func (p *IncrementalProcessor) SetModelForServerInEnvoy(model *mesh.ModelAssignment, server *mesh.ServerAssignment) error {
	clusterName := server.Server.Name + "_" + computeHashKeyForList(model.Assignment)
	p.xdsCache.AddRoute(model.Model.Name,model.Model.Name,clusterName)
	if !p.xdsCache.HasCluster(clusterName) {
		p.xdsCache.AddCluster(clusterName)
		for _,serverIdx := range model.Assignment {
			serverInstance := server.Server.Replicas[serverIdx]
			p.xdsCache.AddEndpoint(clusterName, serverInstance.InferenceSvc, uint32(serverInstance.InferencePort))
		}
	}

	// Create the snapshot that we'll serve to Envoy
	snapshot,err := cache.NewSnapshot(
		p.newSnapshotVersion(), // version
		map[rsrc.Type][]types.Resource{
			rsrc.ClusterType: p.xdsCache.ClusterContents(),   // clusters
			rsrc.RouteType: p.xdsCache.RouteContents(),     // routes
			rsrc.ListenerType: p.xdsCache.ListenerContents(),  // listeners
		})
	if err != nil {
		return err
	}

	if err := snapshot.Consistent(); err != nil {
		return err
	}
	p.logger.Debugf("will serve snapshot %+v", snapshot)

	// Add the snapshot to the cache
	if err := p.cache.SetSnapshot(context.Background(), p.nodeID, snapshot); err != nil {
		return err
	}
	return nil
}