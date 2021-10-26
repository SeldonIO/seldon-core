package processor

import (
	"context"
	"errors"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	rsrc "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/seldonio/seldon-core/scheduler/pkg/envoy/resources"
	"github.com/seldonio/seldon-core/scheduler/pkg/envoy/xdscache"
	"github.com/seldonio/seldon-core/scheduler/pkg/store"
	"github.com/sirupsen/logrus"
	"math"
	"math/rand"
	"strconv"
	"sync"
)

type IncrementalProcessor struct {
	cache  cache.SnapshotCache
	nodeID string
	// snapshotVersion holds the current version of the snapshot.
	snapshotVersion int64
	logger logrus.FieldLogger
	xdsCache xdscache.SeldonXDSCache
	mu sync.RWMutex
	store store.SchedulerStore
	source chan string
}

func NewIncrementalProcessor(cache cache.SnapshotCache, nodeID string, log logrus.FieldLogger, store store.SchedulerStore, source chan string) *IncrementalProcessor {
	ip := &IncrementalProcessor{
		cache:           cache,
		nodeID:          nodeID,
		snapshotVersion: rand.Int63n(1000),
		logger:   log.WithField("Source","EnvoyServer"),
		xdsCache: xdscache.SeldonXDSCache{
			Listeners: make(map[string]resources.Listener),
			Clusters:  make(map[string]resources.Cluster),
			Routes:    make(map[string]resources.Route),
			Endpoints: make(map[string]resources.Endpoint),
		},
		store: store,
		source: source,
	}
	ip.SetListener("seldon_http")
	return ip
}

func (s *IncrementalProcessor) ListenForSyncs() {
	for msg := range s.source {
		s.logger.Infof("Received sync for model %s",msg)
		err := s.Sync(msg)
		if err != nil {
			s.logger.Errorf("Failed to process sync")
		}
	}
}


func (p *IncrementalProcessor) SetListener(listenerName string) {
	p.mu.Lock()
	defer p.mu.Unlock()
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

func (p *IncrementalProcessor) updateEnvoy() error {
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

func (p *IncrementalProcessor) removeModelForServerInEnvoy(modelName string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.xdsCache.RemoveRoute(modelName)
	return p.updateEnvoy()
}

func (p *IncrementalProcessor) Sync(modelName string) error {
	model, err := p.store.GetModel(modelName)
	if err != nil {
		if errors.Is(err, store.ModelNotFoundErr) {
			return p.removeModelForServerInEnvoy(modelName)
		}
		return nil
	}
	server, err := p.store.GetServer(model.Server())
	if err != nil {
		if errors.Is(err, store.ServerNotFoundErr) {
			return p.removeModelForServerInEnvoy(modelName)
		}
		return nil
	}
	if model.NoLiveReplica() {
		return p.removeModelForServerInEnvoy(modelName)
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	assignment := model.GetAssignment() // Get loaded replicas for model
	clusterName := server.Key() + "_" + computeHashKeyForList(assignment)
	p.xdsCache.AddRoute(modelName,modelName,clusterName)
	if !p.xdsCache.HasCluster(clusterName) {
		p.xdsCache.AddCluster(clusterName, modelName)
		for _,serverIdx := range assignment {
			p.xdsCache.AddEndpoint(clusterName, server.GetReplicaInferenceSvc(serverIdx), uint32(server.GetReplicaInferencePort(serverIdx)))
		}
	}
	return p.updateEnvoy()
}