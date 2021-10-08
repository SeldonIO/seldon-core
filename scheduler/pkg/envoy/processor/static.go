package processor

import (
	"context"
	"github.com/seldonio/seldon-core/scheduler/apis/mesh"
	"math"
	"math/rand"
	"os"
	"strconv"

	"github.com/seldonio/seldon-core/scheduler/pkg/envoy/resources"

	"github.com/envoyproxy/go-control-plane/pkg/cache/types"

	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	rsrc "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/seldonio/seldon-core/scheduler/pkg/envoy/watcher"
	"github.com/seldonio/seldon-core/scheduler/pkg/envoy/xdscache"
	"github.com/sirupsen/logrus"
)

type SeldonProcessor struct {
	cache  cache.SnapshotCache
	nodeID string

	// snapshotVersion holds the current version of the snapshot.
	snapshotVersion int64

	logrus.FieldLogger

	xdsCache xdscache.SeldonXDSCache
}

func NewSeldonProcessor(cache cache.SnapshotCache, nodeID string, log logrus.FieldLogger) *SeldonProcessor {
	return &SeldonProcessor{
		cache:           cache,
		nodeID:          nodeID,
		snapshotVersion: rand.Int63n(1000),
		FieldLogger:     log,
		xdsCache: xdscache.SeldonXDSCache{
			Listeners: make(map[string]resources.Listener),
			Clusters:  make(map[string]resources.Cluster),
			Routes:    make(map[string]resources.Route),
			Endpoints: make(map[string]resources.Endpoint),
		},
	}
}

// newSnapshotVersion increments the current snapshotVersion
// and returns as a string.
func (p *SeldonProcessor) newSnapshotVersion() string {

	// Reset the snapshotVersion if it ever hits max size.
	if p.snapshotVersion == math.MaxInt64 {
		p.snapshotVersion = 0
	}

	// Increment the snapshot version & return as string.
	p.snapshotVersion++
	return strconv.FormatInt(p.snapshotVersion, 10)
}

// ProcessFile takes a file and generates an xDS snapshot
func (p *SeldonProcessor) ProcessFile(msg watcher.NotifyMessage) {

	// Parse file into object
	seldonConfig, err := parseSeldonYaml(msg.Contents)
	if err != nil {
		p.Errorf("error parsing yaml file: %+v", err)
		return
	}

	//modelNames := make([]string, len(seldonConfig.Models))
	servers := make(map[string]map[int]mesh.Replica)
	for _,s := range seldonConfig.Servers {
		for i,r := range s.Replicas {
			if _, ok := servers[s.Name]; !ok {
				servers[s.Name] = make(map[int]mesh.Replica)
			}
			servers[s.Name][i] = r
		}
	}
	clustersAdded := make(map[string]bool)
	for _, m := range seldonConfig.Models {
		//modelNames[i] = m.Name
		serverSubsetHash := computeHashKeyForList(m.Servers)
		clusterName := m.ModelServer+"_"+serverSubsetHash
		p.xdsCache.AddRoute(m.Name,m.Name,clusterName)
		if _,ok := clustersAdded[clusterName]; !ok {
			p.xdsCache.AddCluster(clusterName)
			for _,serverIdx := range m.Servers {
				replica := servers[m.ModelServer][serverIdx]
				p.xdsCache.AddEndpoint(clusterName, replica.Address, replica.Port)
			}
		}
	}
	p.xdsCache.AddListener(seldonConfig.Name)

	// Create the snapshot that we'll serve to Envoy
	snapshot,err := cache.NewSnapshot(
		p.newSnapshotVersion(), // version
		map[rsrc.Type][]types.Resource{
			//rsrc.EndpointType: p.xdsCache.EndpointsContents(), // endpoints
			rsrc.ClusterType: p.xdsCache.ClusterContents(),   // clusters
			rsrc.RouteType: p.xdsCache.RouteContents(),     // routes
			rsrc.ListenerType: p.xdsCache.ListenerContents(),  // listeners
		})
	if err != nil {
		p.Errorf("new snapshot failed: %+v\n\n\n%+v", snapshot, err)
		return
	}

	if err := snapshot.Consistent(); err != nil {
		p.Errorf("snapshot inconsistency: %+v\n\n\n%+v", snapshot, err)
		return
	}
	p.Debugf("will serve snapshot %+v", snapshot)

	// Add the snapshot to the cache
	if err := p.cache.SetSnapshot(context.Background(), p.nodeID, snapshot); err != nil {
		p.Errorf("snapshot error %q for %+v", err, snapshot)
		os.Exit(1)
	}
}
