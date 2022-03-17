package processor

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"sync"

	"github.com/seldonio/seldon-core/scheduler/pkg/store/experiment"

	"github.com/seldonio/seldon-core/scheduler/pkg/coordinator"

	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	rsrc "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/seldonio/seldon-core/scheduler/pkg/envoy/xdscache"
	"github.com/seldonio/seldon-core/scheduler/pkg/store"
	"github.com/sirupsen/logrus"
)

const (
	pendingSyncsQueueSize      int = 100
	modelEventHandlerName          = "incremental.processor.models"
	experimentEventHandlerName     = "incremental.processor.experiments"
)

type IncrementalProcessor struct {
	cache  cache.SnapshotCache
	nodeID string
	// snapshotVersion holds the current version of the snapshot.
	snapshotVersion  int64
	logger           logrus.FieldLogger
	xdsCache         *xdscache.SeldonXDSCache
	mu               sync.RWMutex
	modelStore       store.ModelStore
	experimentServer experiment.ExperimentServer
}

func NewIncrementalProcessor(
	cache cache.SnapshotCache,
	nodeID string,
	log logrus.FieldLogger,
	modelStore store.ModelStore,
	experimentServer experiment.ExperimentServer,
	hub *coordinator.EventHub,
) *IncrementalProcessor {
	ip := &IncrementalProcessor{
		cache:            cache,
		nodeID:           nodeID,
		snapshotVersion:  rand.Int63n(1000),
		logger:           log.WithField("source", "EnvoyServer"),
		xdsCache:         xdscache.NewSeldonXDSCache(log),
		modelStore:       modelStore,
		experimentServer: experimentServer,
	}

	ip.SetListener("seldon_http")
	hub.RegisterModelEventHandler(
		modelEventHandlerName,
		pendingSyncsQueueSize,
		ip.logger,
		ip.handleModelEvents,
	)
	hub.RegisterExperimentEventHandler(
		experimentEventHandlerName,
		pendingSyncsQueueSize,
		ip.logger,
		ip.handleExperimentEvents,
	)

	return ip
}

func (p *IncrementalProcessor) handleExperimentEvents(event coordinator.ExperimentEventMsg) {
	logger := p.logger.WithField("func", "handleExperimentEvents")
	logger.Debugf("Received sync for experiment %s", event.String())
	if event.Status == nil {
		err := p.experimentSync(event.ExperimentName)
		var err2 error
		if err != nil {
			logger.WithError(err).Errorf("Failed to process sync for experiment %s", event.String())
			err2 = p.experimentServer.SetStatus(event.ExperimentName, false, err.Error())
		} else {
			err2 = p.experimentServer.SetStatus(event.ExperimentName, true, "experiment active")
		}
		if err2 != nil {
			logger.WithError(err2).Errorf("Failed to set experiment activation")
		}
	}
}

func (p *IncrementalProcessor) handleModelEvents(event coordinator.ModelEventMsg) {
	logger := p.logger.WithField("func", "handleModelEvents")
	logger.Debugf("Received sync for model %s", event.String())

	err := p.modelSync(event.ModelName)
	if err != nil {
		logger.WithError(err).Errorf("Failed to process sync for model %s", event.String())
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

func (p *IncrementalProcessor) removeRouteForServerInEnvoy(routeName string) error {
	logger := p.logger.WithField("func", "removeModelForServerInEnvoy")
	err := p.xdsCache.RemoveRoute(routeName)
	if err != nil {
		logger.Debugf("Failed to remove route for %s", routeName)
		return err
	}
	return p.updateEnvoy()
}

func (p *IncrementalProcessor) updateEnvoyForModelVersion(modelRouteName string, modelVersion *store.ModelVersion, server *store.ServerSnapshot, trafficPercent uint32) {
	logger := p.logger.WithField("func", "updateEnvoyForModelVersion")
	assignment := modelVersion.GetAssignment() // Get loaded replicas for model
	if len(assignment) == 0 {
		logger.Debugf("No assigned replicas so returning for %s", modelRouteName)
		return
	}
	clusterNameBase := server.Name + "_" + computeHashKeyForList(assignment)
	httpClusterName := clusterNameBase + "_http"
	grpcClusterName := clusterNameBase + "_grpc"
	p.xdsCache.AddCluster(httpClusterName, modelVersion.GetModel().GetMeta().GetName(), modelVersion.GetVersion(), false)
	for _, replicaIdx := range assignment {
		replica, ok := server.Replicas[replicaIdx]
		if !ok {
			logger.Warnf("Invalid replica index %d for server %s", replicaIdx, server.Name)
		} else {
			p.xdsCache.AddEndpoint(httpClusterName, replica.GetInferenceSvc(), uint32(replica.GetInferenceHttpPort()))
		}
	}
	p.xdsCache.AddCluster(grpcClusterName, modelVersion.GetModel().GetMeta().GetName(), modelVersion.GetVersion(), true)
	for _, replicaIdx := range assignment {
		replica, ok := server.Replicas[replicaIdx]
		if !ok {
			logger.Warnf("Invalid replica index %d for server %s", replicaIdx, server.Name)
		} else {
			p.xdsCache.AddEndpoint(grpcClusterName, replica.GetInferenceSvc(), uint32(replica.GetInferenceGrpcPort()))
		}
	}
	p.xdsCache.AddRouteClusterTraffic(modelRouteName, modelVersion.GetModel().GetMeta().GetName(), modelVersion.GetVersion(), trafficPercent, httpClusterName, grpcClusterName, modelVersion.GetDeploymentSpec().LogPayloads)
}

func getTrafficShare(latestModel *store.ModelVersion, lastAvailableModelVersion *store.ModelVersion, weight uint32) (uint32, uint32) {
	lastAvailableReplicas := len(lastAvailableModelVersion.GetAssignment())
	latestReplicas := len(latestModel.GetAssignment())
	totalReplicas := lastAvailableReplicas + latestReplicas
	trafficLastAvailableModel := uint32((lastAvailableReplicas * int(weight)) / totalReplicas)
	trafficLatestModel := weight - trafficLastAvailableModel
	return trafficLatestModel, trafficLastAvailableModel
}

func (p *IncrementalProcessor) addModelTraffic(routeName string, model *store.ModelSnapshot, weight uint32) error {
	logger := p.logger.WithField("func", "addModelTraffic")
	modelName := model.Name
	latestModel := model.GetLatest()
	if latestModel == nil || latestModel.NoLiveReplica() {
		return fmt.Errorf("No live replica for model %s for model route %s", model.Name, routeName)
	}
	server, err := p.modelStore.GetServer(latestModel.Server())
	if err != nil {
		return err
	}
	lastAvailableModelVersion := model.GetLastAvailableModel()
	if lastAvailableModelVersion != nil && latestModel.GetVersion() != lastAvailableModelVersion.GetVersion() {
		trafficLatestModel, trafficLastAvailableModel := getTrafficShare(latestModel, lastAvailableModelVersion, weight)
		lastAvailableServer, err := p.modelStore.GetServer(lastAvailableModelVersion.Server())
		if err != nil {
			logger.WithError(err).Errorf("Failed to find server %s for last available model %s", lastAvailableModelVersion.Server(), modelName)
			return err
		}
		logger.Debugf("Splitting traffic between latest %s:%d %d percent and %s:%d %d percent",
			modelName,
			latestModel.GetVersion(),
			trafficLatestModel,
			modelName,
			lastAvailableModelVersion.GetVersion(),
			trafficLastAvailableModel)
		p.updateEnvoyForModelVersion(routeName, lastAvailableModelVersion, lastAvailableServer, trafficLastAvailableModel)
		p.updateEnvoyForModelVersion(routeName, latestModel, server, trafficLatestModel)
	} else {
		p.updateEnvoyForModelVersion(routeName, latestModel, server, weight)
	}
	return nil
}

func (p *IncrementalProcessor) addExperimentBaselineTraffic(model *store.ModelSnapshot, exp *experiment.Experiment) error {
	logger := p.logger.WithField("func", "addExperimentTraffic")
	logger.Infof("Trying to setup experiment for %s", model.Name)
	if exp.DefaultModel == nil {
		return fmt.Errorf("Didn't find baseline in experiment for model %s", model.Name)
	}
	if *exp.DefaultModel != model.Name {
		return fmt.Errorf("Didn't find expected model name baseline in experiment for model found %s but expected %s", *exp.DefaultModel, model.Name)
	}
	if exp.Deleted {
		return fmt.Errorf("Experiment on model %s, but %s is deleted", model.Name, *exp.DefaultModel)
	}
	for _, candidate := range exp.Candidates {
		candidateModel, err := p.modelStore.GetModel(candidate.ModelName)
		if err != nil {
			return err
		}
		err = p.addModelTraffic(model.Name, candidateModel, candidate.Weight)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *IncrementalProcessor) addTraffic(model *store.ModelSnapshot) error {
	logger := p.logger.WithField("func", "addTraffic")
	exp := p.experimentServer.GetExperimentForBaselineModel(model.Name)
	if exp != nil {
		err := p.addExperimentBaselineTraffic(model, exp)
		if err != nil {
			logger.WithError(err).Debugf("Revert experiment traffic to just model %s", model.Name)
			err = p.removeRouteForServerInEnvoy(model.Name)
			if err != nil {
				return err
			}
			return p.addModelTraffic(model.Name, model, 100)
		}
	} else {
		logger.Infof("Handle vanilla no experiment traffic for %s", model.Name)
		return p.addModelTraffic(model.Name, model, 100)
	}
	return nil
}

func (p *IncrementalProcessor) addTrafficForExperiment(routeName string, exp *experiment.Experiment) error {
	for _, candidate := range exp.Candidates {
		candidateModel, err := p.modelStore.GetModel(candidate.ModelName)
		if err != nil {
			return err
		}
		err = p.addModelTraffic(routeName, candidateModel, candidate.Weight)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *IncrementalProcessor) addExperiment(exp *experiment.Experiment) error {
	logger := p.logger.WithField("func", "addExperiment")
	p.mu.Lock()
	defer p.mu.Unlock()
	routeName := fmt.Sprintf("%s.experiment", exp.Name)
	err := p.addTrafficForExperiment(routeName, exp)
	if err != nil {
		logger.WithError(err).Errorf("Failed to add traffic for experiment %s", routeName)
		return p.removeRouteForServerInEnvoy(routeName)
	}
	return p.updateEnvoy()
}

func (p *IncrementalProcessor) experimentSync(experimentName string) error {
	logger := p.logger.WithField("func", "experimentSync")
	exp, err := p.experimentServer.GetExperiment(experimentName)
	if err != nil {
		return err
	}
	if exp.DefaultModel != nil {
		logger.Infof("Experiment %s sync - calling for model %s", experimentName, *exp.DefaultModel)
		err := p.modelSync(*exp.DefaultModel)
		if err != nil {
			return err
		}
	}
	return p.addExperiment(exp)
}

func (p *IncrementalProcessor) modelSync(modelName string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.modelStore.LockModel(modelName)
	defer p.modelStore.UnlockModel(modelName)

	logger := p.logger.WithField("func", "Sync")
	model, err := p.modelStore.GetModel(modelName)
	if err != nil {
		logger.WithError(err).Errorf("Failed to sync model %s", modelName)
		return p.removeRouteForServerInEnvoy(modelName)
	}
	if model == nil {
		logger.Debugf("sync: No model - removing for %s", modelName)
		return p.removeRouteForServerInEnvoy(modelName)
	}
	latestModel := model.GetLatest()
	if latestModel == nil {
		logger.Debugf("sync: No latest model - removing for %s", modelName)
		return p.removeRouteForServerInEnvoy(modelName)
	}
	if latestModel.NoLiveReplica() {
		logger.Debugf("sync: No live model - removing for %s", modelName)
		return p.removeRouteForServerInEnvoy(modelName)
	}
	server, err := p.modelStore.GetServer(latestModel.Server())
	if err != nil || server == nil {
		logger.Debugf("sync: No server - removing for %s", modelName)
		return p.removeRouteForServerInEnvoy(modelName)
	}

	// Remove routes before we recreate
	err = p.xdsCache.RemoveRoute(modelName)
	if err != nil {
		logger.Debugf("Failed to remove route before starting update for %s", modelName)
		return err
	}

	err = p.addTraffic(model)
	if err != nil {
		logger.WithError(err).Errorf("Failed to add traffic for model %s", modelName)
		return p.removeRouteForServerInEnvoy(modelName)
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
		err2 := p.modelStore.UpdateModelState(modelName, latestModel.GetVersion(), server.Name, replicaIdx, nil, expectedState, state, reason)
		if err2 != nil {
			return err2
		}
	}
	return err
}
