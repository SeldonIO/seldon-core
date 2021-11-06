package processor

import (
	"context"
	"fmt"
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

type EnvoyHandler interface {
	SendEnvoySync(modelName string)
}

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

func NewIncrementalProcessor(cache cache.SnapshotCache, nodeID string, log logrus.FieldLogger, store store.SchedulerStore) *IncrementalProcessor {
	ip := &IncrementalProcessor{
		cache:           cache,
		nodeID:          nodeID,
		snapshotVersion: rand.Int63n(1000),
		logger:   log.WithField("source","EnvoyServer"),
		xdsCache: xdscache.SeldonXDSCache{
			Listeners: make(map[string]resources.Listener),
			Clusters:  make(map[string]resources.Cluster),
			Routes:    make(map[string]resources.Route),
			Endpoints: make(map[string]resources.Endpoint),
		},
		store: store,
		source: make(chan string, 1),
	}
	ip.SetListener("seldon_http")
	return ip
}

func (s *IncrementalProcessor) SendEnvoySync(modelName string) {
	s.source <- modelName
}

func (s *IncrementalProcessor) StopEnvoySync() {
	close(s.source)
}

func (s *IncrementalProcessor) ListenForSyncs() {
	logger := s.logger.WithField("func","ListenForSyncs")
	for msg := range s.source {
		logger.Debugf("Received sync for model %s",msg)
		err := s.Sync(msg)
		if err != nil {
			logger.Errorf("Failed to process sync")
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
	logger := p.logger.WithField("func","updateEnvoy")
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
	logger.Debugf("will serve snapshot %+v", snapshot)

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
	logger := p.logger.WithField("func","Sync")
	model, err := p.store.GetModel(modelName)
	if err != nil {
		logger.WithError(err).Errorf("Failed to sync model %s", modelName)
		return p.removeModelForServerInEnvoy(modelName)
	}
	if model == nil {
		logger.Debugf("sync: No model - removing for %s",modelName)
		return p.removeModelForServerInEnvoy(modelName)
	}
	latestModel := model.GetLatest()
	if latestModel == nil {
		logger.Debugf("sync: No latest model - removing for %s",modelName)
		return p.removeModelForServerInEnvoy(modelName)
	}
	if latestModel.NoLiveReplica() {
		logger.Debugf("sync: No live model - removing for %s",modelName)
		return p.removeModelForServerInEnvoy(modelName)
	}
	server, err := p.store.GetServer(latestModel.Server())
	if server == nil {
		logger.Debugf("sync: No server - removing for %s",modelName)
		return p.removeModelForServerInEnvoy(modelName)
	}

	p.mu.Lock()
	defer p.mu.Unlock()
	assignment := latestModel.GetAssignment() // Get loaded replicas for model
	clusterName := server.Name + "_" + computeHashKeyForList(assignment)
	p.xdsCache.AddRoute(modelName,modelName,clusterName)
	if !p.xdsCache.HasCluster(clusterName) {
		p.xdsCache.AddCluster(clusterName, modelName)
		for _,serverIdx := range assignment {
			replica, ok := server.Replicas[serverIdx]
			if !ok {
				return fmt.Errorf("Invalid replica index %d for server %s",serverIdx, server.Name)
			}
			p.xdsCache.AddEndpoint(clusterName, replica.GetInferenceSvc(), uint32(replica.GetInferencePort()))
		}
	}
	return p.updateEnvoy()
}