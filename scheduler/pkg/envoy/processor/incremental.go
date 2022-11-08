package processor

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"sync"
	"time"

	"github.com/seldonio/seldon-core/scheduler/pkg/envoy/resources"

	"github.com/seldonio/seldon-core/scheduler/pkg/store/pipeline"

	"github.com/seldonio/seldon-core/scheduler/pkg/store/experiment"
	"github.com/seldonio/seldon-core/scheduler/pkg/util"

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
	pipelineEventHandlerName       = "incremental.processor.pipelines"
)

type IncrementalProcessor struct {
	cache  cache.SnapshotCache
	nodeID string
	// snapshotVersion holds the current version of the snapshot.
	snapshotVersion      int64
	logger               logrus.FieldLogger
	xdsCache             *xdscache.SeldonXDSCache
	mu                   sync.RWMutex
	modelStore           store.ModelStore
	experimentServer     experiment.ExperimentServer
	pipelineHandler      pipeline.PipelineHandler
	batchTrigger         *time.Timer
	batchWaitMillis      time.Duration
	pendingModelVersions []*pendingModelVersion
}

type pendingModelVersion struct {
	name    string
	version uint32
}

func NewIncrementalProcessor(
	cache cache.SnapshotCache,
	nodeID string,
	log logrus.FieldLogger,
	modelStore store.ModelStore,
	experimentServer experiment.ExperimentServer,
	pipelineHandler pipeline.PipelineHandler,
	hub *coordinator.EventHub,
	pipelineGatewayDetails *xdscache.PipelineGatewayDetails,
) (*IncrementalProcessor, error) {
	ip := &IncrementalProcessor{
		cache:            cache,
		nodeID:           nodeID,
		snapshotVersion:  rand.Int63n(1000),
		logger:           log.WithField("source", "EnvoyServer"),
		xdsCache:         xdscache.NewSeldonXDSCache(log, pipelineGatewayDetails),
		modelStore:       modelStore,
		experimentServer: experimentServer,
		pipelineHandler:  pipelineHandler,
		batchTrigger:     nil,
		batchWaitMillis:  util.EnvoyUpdateDefaultBatchWaitMillis,
	}

	err := ip.setListeners()
	if err != nil {
		return nil, err
	}

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
	hub.RegisterPipelineEventHandler(
		pipelineEventHandlerName,
		pendingSyncsQueueSize,
		ip.logger,
		ip.handlePipelinesEvents,
	)

	err = ip.updateEnvoy()
	if err != nil {
		return nil, err
	}
	return ip, nil
}

func (p *IncrementalProcessor) handlePipelinesEvents(event coordinator.PipelineEventMsg) {
	logger := p.logger.WithField("func", "handleExperimentEvents")
	go func() {
		err := p.addPipeline(event.PipelineName)
		if err != nil {
			logger.WithError(err).Errorf("Failed to add pipeline %s", event.PipelineName)
		}
	}()
}

func (p *IncrementalProcessor) handleExperimentEvents(event coordinator.ExperimentEventMsg) {
	logger := p.logger.WithField("func", "handleExperimentEvents")
	logger.Debugf("Received sync for experiment %s", event.String())
	go func() {
		exp, err := p.experimentServer.GetExperiment(event.ExperimentName)
		if err != nil {
			logger.WithError(err).Errorf("Failed to get experiment %s", event.ExperimentName)
		} else {
			if exp.Deleted {
				err := p.removeExperiment(exp)
				if err != nil {
					logger.WithError(err).Errorf("Failed to get experiment %s", event.ExperimentName)
				}
			} else {
				if event.UpdatedExperiment {
					err := p.experimentUpdate(exp)
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
		}
	}()
}

func (p *IncrementalProcessor) handleModelEvents(event coordinator.ModelEventMsg) {
	logger := p.logger.WithField("func", "handleModelEvents")
	logger.Debugf("Received sync for model %s", event.String())

	go func() {
		err := p.modelUpdate(event.ModelName)
		if err != nil {
			logger.WithError(err).Errorf("Failed to process sync for model %s", event.String())
		}
	}()
}

func (p *IncrementalProcessor) setListeners() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	err := p.xdsCache.SetupTLS()
	if err != nil {
		return err
	}
	p.xdsCache.AddListeners()
	return nil
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
			rsrc.SecretType:   p.xdsCache.SecretContents(),   // Secrets
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

func (p *IncrementalProcessor) updateEnvoyForModelVersion(modelRouteName string, modelVersion *store.ModelVersion, server *store.ServerSnapshot, trafficPercent uint32, isMirror bool) {
	logger := p.logger.WithField("func", "updateEnvoyForModelVersion")

	assignment := modelVersion.GetAssignment() // Get loaded replicas for model
	if len(assignment) == 0 {
		logger.Debugf("No assigned replicas so returning for %s", modelRouteName)
		return
	}

	clusterNameBase := server.Name + "_" + computeHashKeyForList(assignment)
	httpClusterName := clusterNameBase + "_http"
	grpcClusterName := clusterNameBase + "_grpc"
	p.xdsCache.AddCluster(httpClusterName, modelRouteName, modelVersion.GetModel().GetMeta().GetName(), modelVersion.GetVersion(), false)
	for _, replicaIdx := range assignment {
		replica, ok := server.Replicas[replicaIdx]
		if !ok {
			logger.Warnf("Invalid replica index %d for server %s", replicaIdx, server.Name)
		} else {
			p.xdsCache.AddEndpoint(httpClusterName, replica.GetInferenceSvc(), uint32(replica.GetInferenceHttpPort()))
		}
	}
	p.xdsCache.AddCluster(grpcClusterName, modelRouteName, modelVersion.GetModel().GetMeta().GetName(), modelVersion.GetVersion(), true)
	for _, replicaIdx := range assignment {
		replica, ok := server.Replicas[replicaIdx]
		if !ok {
			logger.Warnf("Invalid replica index %d for server %s", replicaIdx, server.Name)
		} else {
			p.xdsCache.AddEndpoint(grpcClusterName, replica.GetInferenceSvc(), uint32(replica.GetInferenceGrpcPort()))
		}
	}

	logPayloads := false
	if modelVersion.GetDeploymentSpec() != nil {
		logPayloads = modelVersion.GetDeploymentSpec().LogPayloads
	} else {
		logger.Warnf("model %s has not deployment spec", modelVersion.GetModel().GetMeta().GetName())
	}

	p.xdsCache.AddRouteClusterTraffic(modelRouteName, modelVersion.GetModel().GetMeta().GetName(), modelVersion.GetVersion(), trafficPercent, httpClusterName, grpcClusterName, logPayloads, isMirror)
}

func getTrafficShare(latestModel *store.ModelVersion, lastAvailableModelVersion *store.ModelVersion, weight uint32) (uint32, uint32) {
	lastAvailableReplicas := len(lastAvailableModelVersion.GetAssignment())
	latestReplicas := len(latestModel.GetAssignment())
	totalReplicas := lastAvailableReplicas + latestReplicas
	trafficLastAvailableModel := uint32((lastAvailableReplicas * int(weight)) / totalReplicas)
	trafficLatestModel := weight - trafficLastAvailableModel
	return trafficLatestModel, trafficLastAvailableModel
}

func (p *IncrementalProcessor) addModelTraffic(routeName string, model *store.ModelSnapshot, weight uint32, isMirror bool) error {
	logger := p.logger.WithField("func", "addModelTraffic")

	modelName := model.Name
	latestModel := model.GetLatest()
	if latestModel == nil || latestModel.NoLiveReplica() {
		return fmt.Errorf("No live replica for model %s for model route %s", model.Name, routeName)
	}

	server, err := p.modelStore.GetServer(latestModel.Server(), false, false)
	if err != nil {
		return err
	}

	lastAvailableModelVersion := model.GetLastAvailableModel()
	if lastAvailableModelVersion != nil && latestModel.GetVersion() != lastAvailableModelVersion.GetVersion() {
		trafficLatestModel, trafficLastAvailableModel := getTrafficShare(latestModel, lastAvailableModelVersion, weight)
		lastAvailableServer, err := p.modelStore.GetServer(lastAvailableModelVersion.Server(), false, false)
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
		p.updateEnvoyForModelVersion(routeName, lastAvailableModelVersion, lastAvailableServer, trafficLastAvailableModel, isMirror)
		p.updateEnvoyForModelVersion(routeName, latestModel, server, trafficLatestModel, isMirror)
	} else {
		p.updateEnvoyForModelVersion(routeName, latestModel, server, weight, isMirror)
	}
	return nil
}

func (p *IncrementalProcessor) addExperimentModelBaselineTraffic(model *store.ModelSnapshot, exp *experiment.Experiment) error {
	logger := p.logger.WithField("func", "addExperimentModelBaselineTraffic")
	logger.Infof("Trying to setup experiment for %s", model.Name)
	if exp.Default == nil {
		return fmt.Errorf("Didn't find baseline in experiment for model %s", model.Name)
	}
	if *exp.Default != model.Name {
		return fmt.Errorf("Didn't find expected model name baseline in experiment for model found %s but expected %s", *exp.Default, model.Name)
	}
	if exp.Deleted {
		return fmt.Errorf("Experiment on model %s, but %s is deleted", model.Name, *exp.Default)
	}
	for _, candidate := range exp.Candidates {
		candidateModel, err := p.modelStore.GetModel(candidate.Name)
		if err != nil {
			return err
		}
		err = p.addModelTraffic(model.Name, candidateModel, candidate.Weight, false)
		if err != nil {
			return err
		}
	}
	if exp.Mirror != nil {
		mirrorModel, err := p.modelStore.GetModel(exp.Mirror.Name)
		if err != nil {
			return err
		}
		logger.Infof("Getting mirror model %s to add to model %s", mirrorModel.Name, model.Name)
		err = p.addModelTraffic(model.Name, mirrorModel, exp.Mirror.Percent, true)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *IncrementalProcessor) addModel(model *store.ModelSnapshot) error {
	logger := p.logger.WithField("func", "addTraffic")
	exp := p.experimentServer.GetExperimentForBaselineModel(model.Name)
	if exp != nil {
		err := p.addExperimentModelBaselineTraffic(model, exp)
		if err != nil {
			logger.WithError(err).Debugf("Revert experiment traffic to just model %s", model.Name)
			err = p.removeRouteForServerInEnvoy(model.Name)
			if err != nil {
				return err
			}
			return p.addModelTraffic(model.Name, model, 100, false)
		}
	} else {
		logger.Infof("Handle vanilla no experiment traffic for %s", model.Name)
		return p.addModelTraffic(model.Name, model, 100, false)
	}
	return nil
}

func (p *IncrementalProcessor) addTrafficForExperiment(routeName string, exp *experiment.Experiment) error {
	switch exp.ResourceType {
	case experiment.PipelineResourceType:
		for _, candidate := range exp.Candidates {
			p.xdsCache.AddPipelineRoute(routeName, candidate.Name, candidate.Weight, false)
		}
		if exp.Mirror != nil {
			p.xdsCache.AddPipelineRoute(routeName, exp.Mirror.Name, exp.Mirror.Percent, true)
		}
	case experiment.ModelResourceType:
		for _, candidate := range exp.Candidates {
			candidateModel, err := p.modelStore.GetModel(candidate.Name)
			if err != nil {
				return err
			}
			err = p.addModelTraffic(routeName, candidateModel, candidate.Weight, false)
			if err != nil {
				return err
			}
		}
		if exp.Mirror != nil {
			mirrorModel, err := p.modelStore.GetModel(exp.Mirror.Name)
			if err != nil {
				return err
			}
			err = p.addModelTraffic(routeName, mirrorModel, exp.Mirror.Percent, true)
			if err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("Unknown resource type %v", exp.ResourceType)
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

func (p *IncrementalProcessor) removeExperiment(exp *experiment.Experiment) error {
	logger := p.logger.WithField("func", "addExperiment")
	p.mu.Lock()
	defer p.mu.Unlock()
	routeName := fmt.Sprintf("%s.experiment", exp.Name)
	logger.Debugf("Remove experiment route %s", routeName)
	return p.removeRouteForServerInEnvoy(routeName)
}

func getPipelineRouteName(pipelineName string) string {
	return fmt.Sprintf("%s.%s", pipelineName, resources.SeldonPipelineHeaderSuffix)
}

func (p *IncrementalProcessor) addPipeline(pipelineName string) error {
	logger := p.logger.WithField("func", "addPipeline")
	p.mu.Lock()
	defer p.mu.Unlock()
	pip, err := p.pipelineHandler.GetPipeline(pipelineName)
	if err != nil {
		logger.WithError(err).Errorf("Failed to get pipeline %s", pipelineName)
		return err
	} else {
		if pip.Deleted {
			return p.removePipeline(pip)
		}
	}
	routeName := getPipelineRouteName(pip.Name)
	p.xdsCache.RemovePipelineRoute(routeName)
	exp := p.experimentServer.GetExperimentForBaselinePipeline(pip.Name)
	logger.Infof("getting experiment for baseline %s returned %v", pip.Name, exp)
	// This experiment must have a default for this pipeline
	if exp != nil {
		if exp.Default == nil {
			return fmt.Errorf("Didn't find baseline in experiment for pipeline %s", pip.Name)
		}
		if *exp.Default != pip.Name {
			return fmt.Errorf("Didn't find expected pipeline name baseline in experiment for pipeline found %s but expected %s", *exp.Default, pip.Name)
		}
		if exp.Deleted {
			return fmt.Errorf("Experiment on pipeline %s, but %s is deleted", pip.Name, *exp.Default)
		}
		for _, candidate := range exp.Candidates {
			logger.Infof("Adding pipeline experiment candidate %s %s %d", routeName, candidate.Name, candidate.Weight)
			p.xdsCache.AddPipelineRoute(routeName, candidate.Name, candidate.Weight, false)
		}
		if exp.Mirror != nil {
			logger.Infof("Adding pipeline experiment mirror %s %s %d", routeName, exp.Mirror.Name, exp.Mirror.Percent)
			p.xdsCache.AddPipelineRoute(routeName, exp.Mirror.Name, exp.Mirror.Percent, true)
		}
	} else {
		logger.Infof("Adding normal pipeline route %s", routeName)
		p.xdsCache.AddPipelineRoute(routeName, pip.Name, 100, false)
	}

	return p.updateEnvoy()
}

func (p *IncrementalProcessor) removePipeline(pip *pipeline.Pipeline) error {
	p.xdsCache.RemovePipelineRoute(getPipelineRouteName(pip.Name))
	return p.updateEnvoy()
}

func (p *IncrementalProcessor) experimentUpdate(exp *experiment.Experiment) error {
	logger := p.logger.WithField("func", "experimentSync")
	if exp.Default != nil {
		switch exp.ResourceType {
		case experiment.PipelineResourceType:
			return p.addPipeline(*exp.Default)
		case experiment.ModelResourceType:
			logger.Infof("Experiment %s sync - calling for model %s", exp.Name, *exp.Default)
			err := p.modelUpdate(*exp.Default)
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("Unknown resource type %v", exp.ResourceType)
		}
	}
	return p.addExperiment(exp)
}

func (p *IncrementalProcessor) modelUpdate(modelName string) error {
	logger := p.logger.WithField("func", "modelUpdate")
	p.mu.Lock()
	defer p.mu.Unlock()
	p.modelStore.LockModel(modelName)
	defer p.modelStore.UnlockModel(modelName)

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

	server, err := p.modelStore.GetServer(latestModel.Server(), false, false)
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

	err = p.addModel(model)
	if err != nil {
		logger.WithError(err).Errorf("Failed to add traffic for model %s", modelName)
		return p.removeRouteForServerInEnvoy(modelName)
	}

	// Add to batch pending Envoy sync
	p.pendingModelVersions = append(
		p.pendingModelVersions,
		&pendingModelVersion{
			name:    modelName,
			version: latestModel.GetVersion(),
		},
	)

	if p.batchTrigger == nil {
		p.batchTrigger = time.AfterFunc(p.batchWaitMillis, p.modelSync)
	}

	return nil
}

func (p *IncrementalProcessor) modelSync() {
	logger := p.logger.WithField("func", "modelSync")
	p.mu.Lock()
	defer p.mu.Unlock()

	err := p.updateEnvoy()
	serverReplicaState := store.Available
	reason := ""
	if err != nil {
		serverReplicaState = store.LoadedUnavailable
		reason = err.Error()
	}

	for _, mv := range p.pendingModelVersions {
		p.modelStore.LockModel(mv.name)

		m, err := p.modelStore.GetModel(mv.name)
		if err != nil {
			logger.Debugf("Failed to get model %s", mv.name)
			p.modelStore.UnlockModel(mv.name)
			continue
		}

		v := m.GetVersion(mv.version)
		if v == nil {
			logger.Debugf("Failed to get version for model %s version %d", mv.name, mv.version)
			p.modelStore.UnlockModel(mv.name)
			continue
		}

		s, err := p.modelStore.GetServer(v.Server(), false, false)
		if err != nil || s == nil {
			logger.Debugf("Failed to get server for model %s server %s", mv.name, v.Server())
			p.modelStore.UnlockModel(mv.name)
			continue
		}

		vs := v.ReplicaState()
		for _, replicaIdx := range v.GetAssignment() {
			serverReplicaExpectedState := vs[replicaIdx].State
			if serverReplicaExpectedState != store.Draining {
				err2 := p.modelStore.UpdateModelState(
					mv.name,
					v.GetVersion(),
					s.Name,
					replicaIdx,
					nil,
					serverReplicaExpectedState,
					serverReplicaState,
					reason,
				)
				if err2 != nil {
					logger.WithError(err2).Warnf("Failed to update state for model %s", mv.name)
					break
				}
			} else {
				logger.Debugf(
					"Skipping draining server for model %s server replica %s%d", mv.name, v.Server(), replicaIdx)
			}
		}
		p.modelStore.UnlockModel(mv.name)
	}

	// Reset
	p.batchTrigger = nil
	p.pendingModelVersions = nil
}