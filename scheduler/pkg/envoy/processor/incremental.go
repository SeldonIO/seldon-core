package processor

import (
	"context"
	"math"
	"math/rand"
	"strconv"
	"sync"

	"github.com/seldonio/seldon-core/scheduler/pkg/coordinator"

	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	rsrc "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/seldonio/seldon-core/scheduler/pkg/envoy/xdscache"
	"github.com/seldonio/seldon-core/scheduler/pkg/store"
	"github.com/sirupsen/logrus"
)

type IncrementalProcessor struct {
	cache  cache.SnapshotCache
	nodeID string
	// snapshotVersion holds the current version of the snapshot.
	snapshotVersion int64
	logger          logrus.FieldLogger
	xdsCache        *xdscache.SeldonXDSCache
	mu              sync.RWMutex
	store           store.SchedulerStore
	source          chan coordinator.ModelEventMsg
}

func NewIncrementalProcessor(cache cache.SnapshotCache, nodeID string, log logrus.FieldLogger, store store.SchedulerStore, hub *coordinator.ModelEventHub) *IncrementalProcessor {
	ip := &IncrementalProcessor{
		cache:           cache,
		nodeID:          nodeID,
		snapshotVersion: rand.Int63n(1000),
		logger:          log.WithField("source", "EnvoyServer"),
		xdsCache:        xdscache.NewSeldonXDSCache(log),
		store:           store,
		source:          make(chan coordinator.ModelEventMsg, 1),
	}
	ip.SetListener("seldon_http")
	hub.AddListener(ip.source)
	return ip
}

func (p *IncrementalProcessor) ListenForSyncs() {
	logger := p.logger.WithField("func", "ListenForSyncs")
	for evt := range p.source {
		logger.Debugf("Received sync for model %s", evt.String())
		err := p.Sync(evt.ModelName)
		if err != nil {
			logger.WithError(err).Errorf("Failed to process sync for model %s", evt.String())
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
	logger := p.logger.WithField("func", "updateEnvoy")
	// Create the snapshot that we'll serve to Envoy
	snapshot, err := cache.NewSnapshot(
		p.newSnapshotVersion(), // version
		map[rsrc.Type][]types.Resource{
			rsrc.ClusterType:  p.xdsCache.ClusterContents(),  // clusters
			rsrc.RouteType:    p.xdsCache.RouteContents(),    // routes
			rsrc.ListenerType: p.xdsCache.ListenerContents(), // listeners
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
	logger := p.logger.WithField("func", "removeModelForServerInEnvoy")
	err := p.xdsCache.RemoveRoute(modelName)
	if err != nil {
		logger.Debugf("Failed to remove route for %s", modelName)
		return err
	}
	return p.updateEnvoy()
}

func (p *IncrementalProcessor) updateEnvoyForModelVersion(modelName string, modelVersion *store.ModelVersion, server *store.ServerSnapshot, trafficPercent uint32) {
	logger := p.logger.WithField("func", "updateEnvoyForModelVersion")
	assignment := modelVersion.GetAssignment() // Get loaded replicas for model
	if len(assignment) == 0 {
		logger.Debugf("No assigned replicas so returning for %s", modelName)
		return
	}
	clusterNameBase := server.Name + "_" + computeHashKeyForList(assignment)
	httpClusterName := clusterNameBase + "_http"
	grpcClusterName := clusterNameBase + "_grpc"
	p.xdsCache.AddCluster(httpClusterName, modelName, modelVersion.GetVersion(), false)
	for _, replicaIdx := range assignment {
		replica, ok := server.Replicas[replicaIdx]
		if !ok {
			logger.Warnf("Invalid replica index %d for server %s", replicaIdx, server.Name)
		} else {
			p.xdsCache.AddEndpoint(httpClusterName, replica.GetInferenceSvc(), uint32(replica.GetInferenceHttpPort()))
		}
	}
	p.xdsCache.AddCluster(grpcClusterName, modelName, modelVersion.GetVersion(), true)
	for _, replicaIdx := range assignment {
		replica, ok := server.Replicas[replicaIdx]
		if !ok {
			logger.Warnf("Invalid replica index %d for server %s", replicaIdx, server.Name)
		} else {
			p.xdsCache.AddEndpoint(grpcClusterName, replica.GetInferenceSvc(), uint32(replica.GetInferenceGrpcPort()))
		}
	}
	p.xdsCache.AddRouteClusterTraffic(modelName, modelVersion.GetVersion(), trafficPercent, httpClusterName, grpcClusterName, modelVersion.GetDeploymentSpec().LogPayloads)
}

func (p *IncrementalProcessor) Sync(modelName string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.store.LockModel(modelName)
	defer p.store.UnlockModel(modelName)

	logger := p.logger.WithField("func", "Sync")
	model, err := p.store.GetModel(modelName)
	if err != nil {
		logger.WithError(err).Errorf("Failed to sync model %s", modelName)
		return p.removeModelForServerInEnvoy(modelName)
	}
	if model == nil {
		logger.Debugf("sync: No model - removing for %s", modelName)
		return p.removeModelForServerInEnvoy(modelName)
	}
	latestModel := model.GetLatest()
	if latestModel == nil {
		logger.Debugf("sync: No latest model - removing for %s", modelName)
		return p.removeModelForServerInEnvoy(modelName)
	}
	if latestModel.NoLiveReplica() {
		logger.Debugf("sync: No live model - removing for %s", modelName)
		return p.removeModelForServerInEnvoy(modelName)
	}
	server, err := p.store.GetServer(latestModel.Server())
	if err != nil || server == nil {
		logger.Debugf("sync: No server - removing for %s", modelName)
		return p.removeModelForServerInEnvoy(modelName)
	}

	// Remove routes before we recreate
	// This could be inefficient
	// TODO make update incremental?
	err = p.xdsCache.RemoveRoute(modelName)
	if err != nil {
		logger.Debugf("Failed to remove route before starting update for %s", modelName)
		return err
	}
	// Update last Available version
	lastAvailableModelVersion := model.GetLastAvailableModel()
	if lastAvailableModelVersion != nil && latestModel.GetVersion() != lastAvailableModelVersion.GetVersion() {
		totalReplicas := len(lastAvailableModelVersion.GetAssignment()) + len(latestModel.GetAssignment())
		trafficLastAvailableModel := uint32((len(lastAvailableModelVersion.GetAssignment()) * 100 / totalReplicas))
		trafficLatestModel := 100 - trafficLastAvailableModel
		lastAvailableServer, err := p.store.GetServer(lastAvailableModelVersion.Server())
		if err != nil {
			logger.WithError(err).Errorf("Failed to find server %s for last available model for %s", lastAvailableModelVersion.Server(), modelName)
			return err
		}
		logger.Debugf("Splitting traffic between latest %s:%d %d percent and %s:%d %d percent",
			modelName,
			latestModel.GetVersion(),
			trafficLatestModel,
			modelName,
			lastAvailableModelVersion.GetVersion(),
			trafficLastAvailableModel)
		p.updateEnvoyForModelVersion(modelName, lastAvailableModelVersion, lastAvailableServer, trafficLastAvailableModel)
		p.updateEnvoyForModelVersion(modelName, latestModel, server, trafficLatestModel)
	} else {
		p.updateEnvoyForModelVersion(modelName, latestModel, server, 100)
	}

	err = p.updateEnvoy()

	// Update the state after the envoy sync depending on whether we got an error doing the sync
	state := store.Available
	reason := ""
	if err != nil {
		state = store.LoadedUnavailable
		reason = err.Error()
	}
	for _, replicaIdx := range latestModel.GetAssignment() {
		expectedState := latestModel.ReplicaState()[replicaIdx].State
		err2 := p.store.UpdateModelState(modelName, latestModel.GetVersion(), server.Name, replicaIdx, nil, expectedState, state, reason)
		if err2 != nil {
			return err2
		}
	}
	return err
}
