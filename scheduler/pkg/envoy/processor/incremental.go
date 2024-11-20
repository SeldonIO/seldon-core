/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package processor

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"sync"
	"time"

	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	rsrc "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/sirupsen/logrus"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/coordinator"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/envoy/resources"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/envoy/xdscache"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/scheduler/cleaner"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store/experiment"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store/pipeline"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/util"
)

const (
	pendingSyncsQueueSize      int = 1000
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
	runEnvoyBatchUpdates bool
	batchTrigger         *time.Timer
	batchWait            time.Duration
	pendingModelVersions []*pendingModelVersion
	versionCleaner       cleaner.ModelVersionCleaner
	batchTriggerManual   *time.Time
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
	versionCleaner cleaner.ModelVersionCleaner,
) (*IncrementalProcessor, error) {
	ip := &IncrementalProcessor{
		cache:                cache,
		nodeID:               nodeID,
		snapshotVersion:      rand.Int63n(1000),
		logger:               log.WithField("source", "EnvoyServer"),
		xdsCache:             xdscache.NewSeldonXDSCache(log, pipelineGatewayDetails),
		modelStore:           modelStore,
		experimentServer:     experimentServer,
		pipelineHandler:      pipelineHandler,
		runEnvoyBatchUpdates: true,
		batchTrigger:         nil,
		batchWait:            util.EnvoyUpdateDefaultBatchWait,
		versionCleaner:       versionCleaner,
		batchTriggerManual:   nil,
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
	logger := p.logger.WithField("func", "handlePipelineEvents")
	logger.Debugf("Received event %s", event.String())

	// Ignore pipeline events due to model status change to stop pointless processing
	// If models are ready or not has no bearing on whether we need to update pipeline
	if !event.ModelStatusChange {
		go func() {
			err := p.addPipeline(event.PipelineName)
			if err != nil {
				logger.WithError(err).Errorf("Failed to add pipeline %s", event.PipelineName)
			}
		}()
	}
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
					err := p.updateExperiment(exp)
					if err != nil {
						logger.WithError(err).Errorf("Failed to process sync for experiment %s", event.String())
						p.setExperimentStatus(event.ExperimentName, err.Error(), false)
					} else {
						p.setExperimentStatus(event.ExperimentName, "experiment active", true)
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
		err := p.updateModel(event.ModelName)
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

// This function does not call `updateEnvoy` directly and therefore callers should make sure
// that this is done (either in a batched update or directly)
func (p *IncrementalProcessor) removeRouteForServerInEnvoyCache(routeName string) error {
	logger := p.logger.WithField("func", "removeRouteForServerInEnvoyCache")
	err := p.xdsCache.RemoveRoute(routeName)
	if err != nil {
		logger.Errorf("Failed to remove route for %s", routeName)
		return err
	}
	return nil
}

func (p *IncrementalProcessor) addEnvoyClustersForModelVersion(modelRouteName string, modelVersion *store.ModelVersion, server *store.ServerSnapshot) {
	logger := p.logger.WithField("func", "updateEnvoyForModelVersion")

	assignment := modelVersion.GetAssignment()
	if len(assignment) == 0 {
		logger.Debugf("No assigned replicas so returning for %s", modelRouteName)
		return
	}

	httpClusterName, grpcClusterName := getClusterNames(modelVersion)
	p.xdsCache.AddCluster(httpClusterName, false)
	for _, replicaIdx := range assignment {
		replica, ok := server.Replicas[replicaIdx]
		if !ok {
			logger.Warnf("Invalid replica index %d for server %s", replicaIdx, server.Name)
		} else {
			p.xdsCache.AddEndpoint(httpClusterName, replica.GetInferenceSvc(), uint32(replica.GetInferenceHttpPort()))
		}
	}
	p.xdsCache.AddCluster(grpcClusterName, true)
	for _, replicaIdx := range assignment {
		replica, ok := server.Replicas[replicaIdx]
		if !ok {
			logger.Warnf("Invalid replica index %d for server %s", replicaIdx, server.Name)
		} else {
			p.xdsCache.AddEndpoint(grpcClusterName, replica.GetInferenceSvc(), uint32(replica.GetInferenceGrpcPort()))
		}
	}
}

func (p *IncrementalProcessor) sendTrafficToModelVersion(modelRouteName string, modelVersion *store.ModelVersion, server *store.ServerSnapshot, trafficPercent uint32, isMirror bool) {
	logger := p.logger.WithField("func", "sendTrafficToModelVersion")

	assignment := modelVersion.GetAssignment()
	if len(assignment) == 0 {
		logger.Debugf("No assigned replicas so returning for %s", modelRouteName)
		return
	}

	logPayloads := false
	if modelVersion.GetDeploymentSpec() != nil {
		logPayloads = modelVersion.GetDeploymentSpec().LogPayloads
	} else {
		logger.Warnf("model %s has not deployment spec", modelVersion.GetModel().GetMeta().GetName())
	}

	httpClusterName, grpcClusterName := getClusterNames(modelVersion)
	p.xdsCache.AddRouteClusterTraffic(modelRouteName, modelVersion.GetModel().GetMeta().GetName(), modelVersion.GetVersion(), trafficPercent, httpClusterName, grpcClusterName, logPayloads, isMirror)
}

func getClusterNames(modelVersion *store.ModelVersion) (httpClusterName, grpcClusterName string) {
	clusterNameBase := modelVersion.GetMeta().GetName() + "_" + strconv.FormatInt(int64(modelVersion.GetVersion()), 10)
	httpClusterName = clusterNameBase + "_http"
	grpcClusterName = clusterNameBase + "_grpc"
	return httpClusterName, grpcClusterName
}

func getTrafficShare(latestModel *store.ModelVersion, lastAvailableModelVersion *store.ModelVersion, weight uint32) (uint32, uint32) {
	lastAvailableReplicas := len(lastAvailableModelVersion.GetAssignment())
	latestReplicas := len(latestModel.GetAssignment())
	totalReplicas := lastAvailableReplicas + latestReplicas
	trafficLastAvailableModel := uint32((lastAvailableReplicas * int(weight)) / totalReplicas)
	trafficLatestModel := weight - trafficLastAvailableModel
	return trafficLatestModel, trafficLastAvailableModel
}

func (p *IncrementalProcessor) addModelTraffic(routeName string, model *store.ModelSnapshot, weight uint32, isMirror, isRouteStale bool) error {
	logger := p.logger.WithField("func", "addModelTraffic")

	modelName := model.Name
	latestModel := model.GetLatest()
	if latestModel == nil || !model.CanReceiveTraffic() {
		if latestModel == nil {
			logger.Infof("latest model is nil for model %s route %s", model.Name, routeName)
		}
		return fmt.Errorf("no live replica for model %s for model route %s", model.Name, routeName)
	}

	server, err := p.modelStore.GetServer(latestModel.Server(), false, false)
	if err != nil {
		return err
	}

	lastAvailableModel := model.GetLastAvailableModel()

	if lastAvailableModel != nil && latestModel.GetVersion() != lastAvailableModel.GetVersion() {
		trafficLatestModel, trafficLastAvailableModel := getTrafficShare(latestModel, lastAvailableModel, weight)
		lastAvailableServer, err := p.modelStore.GetServer(lastAvailableModel.Server(), false, false)
		if err != nil {
			logger.WithError(err).Errorf("Failed to find server %s for last available model %s", lastAvailableModel.Server(), modelName)
			return err
		}
		logger.Debugf("Splitting traffic between latest %s:%d %d percent and %s:%d %d percent",
			modelName,
			latestModel.GetVersion(),
			trafficLatestModel,
			modelName,
			lastAvailableModel.GetVersion(),
			trafficLastAvailableModel)

		if isRouteStale {
			p.addEnvoyClustersForModelVersion(routeName, lastAvailableModel, lastAvailableServer)
			p.addEnvoyClustersForModelVersion(routeName, latestModel, server)
			err = p.updateEnvoy()
			if err != nil {
				return err
			}

			err = p.removeRouteForServerInEnvoyCache(routeName)
			if err != nil {
				return err
			}
		}

		p.addEnvoyClustersForModelVersion(routeName, lastAvailableModel, lastAvailableServer)
		p.sendTrafficToModelVersion(routeName, lastAvailableModel, lastAvailableServer, trafficLastAvailableModel, isMirror)

		p.addEnvoyClustersForModelVersion(routeName, latestModel, server)
		p.sendTrafficToModelVersion(routeName, latestModel, server, trafficLatestModel, isMirror)

	} else {
		if isRouteStale {
			p.addEnvoyClustersForModelVersion(routeName, latestModel, server)
			err = p.updateEnvoy()
			if err != nil {
				return err
			}

			err = p.removeRouteForServerInEnvoyCache(routeName)
			if err != nil {
				return err
			}
		}

		p.addEnvoyClustersForModelVersion(routeName, latestModel, server)
		p.sendTrafficToModelVersion(routeName, latestModel, server, weight, isMirror)
	}
	return nil
}

func (p *IncrementalProcessor) addBaselineModelTraffic(baselineModel *store.ModelSnapshot, exp *experiment.Experiment) error {
	logger := p.logger.WithField("func", "addBaselineModelTraffic")
	logger.Infof("Trying to setup experiment for %s", baselineModel.Name)
	if exp.Default == nil {
		return fmt.Errorf("didn't find baseline in experiment for model %s", baselineModel.Name)
	}
	if *exp.Default != baselineModel.Name {
		return fmt.Errorf("didn't find expected model name baseline in experiment for model found %s but expected %s", *exp.Default, baselineModel.Name)
	}
	if exp.Deleted {
		return fmt.Errorf("experiment on model %s, but %s is deleted", baselineModel.Name, *exp.Default)
	}
	for _, candidate := range exp.Candidates {
		candidateModel, err := p.modelStore.GetModel(candidate.Name)
		if err != nil {
			return err
		}
		err = p.addModelTraffic(baselineModel.Name, candidateModel, candidate.Weight, false, true)
		if err != nil {
			return err
		}
	}
	if exp.Mirror != nil {
		mirrorModel, err := p.modelStore.GetModel(exp.Mirror.Name)
		if err != nil {
			return err
		}
		logger.Infof("Getting mirror model %s to add to model %s", mirrorModel.Name, baselineModel.Name)
		err = p.addModelTraffic(baselineModel.Name, mirrorModel, exp.Mirror.Percent, true, true)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *IncrementalProcessor) addExperimentTraffic(exp *experiment.Experiment) error {
	routeName := getExperimentRouteName(exp.Name)
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
			err = p.addModelTraffic(routeName, candidateModel, candidate.Weight, false, false)
			if err != nil {
				return err
			}
		}
		if exp.Mirror != nil {
			mirrorModel, err := p.modelStore.GetModel(exp.Mirror.Name)
			if err != nil {
				return err
			}
			err = p.addModelTraffic(routeName, mirrorModel, exp.Mirror.Percent, true, false)
			if err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("unknown resource type %v", exp.ResourceType)
	}
	return nil
}

// TODO make envoy updates for experiments batched
func (p *IncrementalProcessor) addExperiment(exp *experiment.Experiment) error {
	logger := p.logger.WithField("func", "addExperiment")
	p.mu.Lock()
	defer p.mu.Unlock()
	routeName := getExperimentRouteName(exp.Name)
	// first clear any existing routes
	if err := p.removeRouteForServerInEnvoyCache(routeName); err != nil {
		logger.WithError(err).Errorf("Failed to remove traffic for experiment %s", routeName)
		return err
	}

	if err := p.addExperimentTraffic(exp); err != nil {
		logger.WithError(err).Errorf("Failed to add traffic for experiment %s", routeName)
		return err
	}
	return p.updateEnvoy()
}

func (p *IncrementalProcessor) removeExperiment(exp *experiment.Experiment) error {
	logger := p.logger.WithField("func", "removeExperiment")
	p.mu.Lock()
	defer p.mu.Unlock()
	routeName := getExperimentRouteName(exp.Name)
	logger.Debugf("Remove experiment route %s", routeName)
	if err := p.removeRouteForServerInEnvoyCache(routeName); err != nil {
		logger.WithError(err).Errorf("Failed to remove traffic for experiment %s", routeName)
		return err
	}
	return p.updateEnvoy()
}

func getExperimentRouteName(experimentName string) string {
	return fmt.Sprintf("%s.%s", experimentName, "experiment")
}

func getPipelineRouteName(pipelineName string) string {
	return fmt.Sprintf("%s.%s", pipelineName, resources.SeldonPipelineHeaderSuffix)
}

// TODO make envoy updates for pipelines batched
func (p *IncrementalProcessor) addPipeline(pipelineName string) error {
	logger := p.logger.WithField("func", "addPipeline")
	p.mu.Lock()
	defer p.mu.Unlock()
	pip, err := p.pipelineHandler.GetPipeline(pipelineName)
	if err != nil {
		logger.WithError(err).Errorf("Failed to get pipeline %s", pipelineName)
		return err
	}
	logger.Debugf("Handling pipeline %s deleted %v", pip.Name, pip.Deleted)
	if pip.Deleted {
		p.xdsCache.RemovePipelineRoute(getPipelineRouteName(pip.Name))
		return p.updateEnvoy()
	}

	routeName := getPipelineRouteName(pip.Name)
	p.xdsCache.RemovePipelineRoute(routeName)
	exp := p.experimentServer.GetExperimentForBaselinePipeline(pip.Name)
	logger.Debugf("getting experiment for baseline %s returned %v", pip.Name, exp)
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

func (p *IncrementalProcessor) updateExperiment(exp *experiment.Experiment) error {
	logger := p.logger.WithField("func", "updateExperiment")
	if exp.Default != nil {
		switch exp.ResourceType {
		case experiment.PipelineResourceType:
			logger.Infof("Experiment %s sync - calling for pipeline %s", exp.Name, *exp.Default)
			err := p.addPipeline(*exp.Default)
			if err != nil {
				return err
			}
		case experiment.ModelResourceType:
			logger.Infof("Experiment %s sync - calling for model %s", exp.Name, *exp.Default)
			err := p.updateModel(*exp.Default)
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("unknown resource type %v", exp.ResourceType)
		}
	}
	return p.addExperiment(exp)
}

func (p *IncrementalProcessor) updateModel(modelName string) error {
	logger := p.logger.WithField("func", "updateModel")
	p.mu.Lock()
	defer p.mu.Unlock()
	p.modelStore.LockModel(modelName)

	logger.Debugf("Calling model update for %s", modelName)

	removeRouteAndUnlockModel := func(modelName string) error {
		if err := p.removeRouteForServerInEnvoyCache(modelName); err != nil {
			logger.WithError(err).Errorf("Failed to remove route from envoy for model: %s", modelName)
			p.modelStore.UnlockModel(modelName)
			return err
		}
		return nil
	}

	model, err := p.modelStore.GetModel(modelName)
	if err != nil {
		logger.WithError(err).Warnf("Failed to sync model %s", modelName)
		if err := removeRouteAndUnlockModel(modelName); err != nil {
			return err
		}
	}

	if model == nil {
		if err := removeRouteAndUnlockModel(modelName); err == nil {
			p.modelStore.UnlockModel(modelName)
		}
		return p.updateEnvoy() // in practice we should not be here
	}

	latestModel := model.GetLatest()
	if latestModel == nil {
		if err := removeRouteAndUnlockModel(modelName); err == nil {
			p.modelStore.UnlockModel(modelName)
		}
		return p.updateEnvoy() // in practice we should not be here
	}

	// Add a modelRemoved boolean so we can continue to batch update but skip steps if we have
	// decided this model can't be added for some reason. This allows batch deletion of routes
	// to take place for errors as well as the successful path through the methods
	modelRemoved := false
	if !model.CanReceiveTraffic() {
		logger.Debugf("sync: Model can't receive traffic - removing route for %s", modelName)
		if err := removeRouteAndUnlockModel(modelName); err != nil {
			return err
		}
		modelRemoved = true
	}

	if !modelRemoved {
		_, err = p.modelStore.GetServer(latestModel.Server(), false, false)
		if err != nil {
			logger.Debugf("sync: No server - removing route for %s", modelName)
			if err := removeRouteAndUnlockModel(modelName); err != nil {
				return err
			}
			modelRemoved = true
		}
	}

	if !modelRemoved {
		var err error
		exp := p.experimentServer.GetExperimentForBaselineModel(model.Name)
		if exp != nil {
			err = p.addBaselineModelTraffic(model, exp)
			if err != nil {
				logger.WithError(err).Debugf("Revert experiment traffic to just model %s", model.Name)
				err = p.removeRouteForServerInEnvoyCache(model.Name)
				if err == nil {
					err = p.addModelTraffic(model.Name, model, 100, false, true)
				}
			}
		} else {
			logger.Infof("Handle vanilla no experiment traffic for %s", model.Name)
			err = p.addModelTraffic(model.Name, model, 100, false, true)
		}

		if err != nil {
			logger.WithError(err).Errorf("Failed to add traffic for model %s", modelName)
			if err := removeRouteAndUnlockModel(modelName); err != nil {
				return err
			}
		}
	}

	// Add to batch pending Envoy sync
	p.pendingModelVersions = append(
		p.pendingModelVersions,
		&pendingModelVersion{
			name:    modelName,
			version: latestModel.GetVersion(),
		},
	)
	p.modelStore.UnlockModel(modelName)
	triggered := p.triggerModelSyncIfNeeded()

	if !triggered {
		// we still need to enable the cron timer as there is no guarantee that the manual trigger will be called
		if p.batchTrigger == nil && p.runEnvoyBatchUpdates {
			p.batchTrigger = time.AfterFunc(p.batchWait, p.modelSyncWithLock)
		}
	}

	return nil
}

func (p *IncrementalProcessor) triggerModelSyncIfNeeded() bool {
	// the first time we trigger the batch update we need to set the time
	if p.batchTriggerManual == nil {
		p.batchTriggerManual = new(time.Time)
		*p.batchTriggerManual = time.Now()
	}
	if time.Since(*p.batchTriggerManual) > p.batchWait {
		// we have waited long enough so we can trigger the batch update
		// we do this inline so that we do not require to release and reacquire the lock
		// which under heavy load there is no guarantee of order and therefore could lead
		// to starvation of the batch update
		p.modelSync()
		*p.batchTriggerManual = time.Now()
		return true
	}
	return false
}

func (p *IncrementalProcessor) modelSyncWithLock() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.modelSync()
}

func (p *IncrementalProcessor) modelSync() {
	logger := p.logger.WithField("func", "modelSync")
	logger.Debugf("Calling model sync")

	envoyErr := p.updateEnvoy()
	serverReplicaState := store.Available
	reason := ""
	if envoyErr != nil {
		serverReplicaState = store.LoadedUnavailable
		reason = envoyErr.Error()
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
		if err != nil {
			logger.Debugf("Failed to get server for model %s server %s", mv.name, v.Server())
			p.modelStore.UnlockModel(mv.name)
			continue
		}

		vs := v.ReplicaState()
		// Go through all replicas that can receive traffic
		for _, replicaIdx := range v.GetAssignment() {
			serverReplicaExpectedState := vs[replicaIdx].State
			// Ignore draining nodes to be changed to Available/Failed state
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
					logger.WithError(err2).Warnf("Failed to update replica state for model %s to %s from %s",
						mv.name, serverReplicaState.String(), serverReplicaExpectedState.String())
				}
			} else {
				logger.Debugf(
					"Skipping replica for model %s in state %s server replica %s%d as no longer in Loaded state",
					mv.name, serverReplicaExpectedState.String(), v.Server(), replicaIdx)
			}
		}
		// Go through all replicas that we have set to UnloadEnvoyRequested and mark them as UnloadRequested
		// to resume the unload process from servers
		for _, replicaIdx := range v.GetReplicaForState(store.UnloadEnvoyRequested) {
			serverReplicaExpectedState := vs[replicaIdx].State
			if err := p.modelStore.UpdateModelState(
				mv.name,
				v.GetVersion(),
				s.Name,
				replicaIdx,
				nil,
				serverReplicaExpectedState,
				store.UnloadRequested,
				"",
			); err != nil {
				logger.WithError(err).Warnf("Failed to update replica state for model %s to %s from %s",
					mv.name, store.UnloadRequested.String(), serverReplicaExpectedState.String())
			}
		}
		p.modelStore.UnlockModel(mv.name)
		p.callVersionCleanupIfNeeded(m.Name)
	}

	// Reset
	p.batchTrigger = nil
	p.pendingModelVersions = nil
	logger.Debugf("Done modelSync")
}

func (p *IncrementalProcessor) callVersionCleanupIfNeeded(modelName string) {
	logger := p.logger.WithField("func", "callVersionCleanupIfNeeded")
	if routes, ok := p.xdsCache.Routes[modelName]; ok {
		logger.Debugf("routes for model %s %v", modelName, routes)
		if p.versionCleaner != nil {
			logger.Debugf("Calling cleanup for model %s", modelName)
			p.versionCleaner.RunCleanup(modelName)
		}
	}
}

func (p *IncrementalProcessor) setExperimentStatus(name, reason string, active bool) {
	logger := p.logger.WithField("func", "setExperimentStatus")
	err := p.experimentServer.SetStatus(name, active, reason)
	if err != nil {
		logger.WithError(err).Errorf("Failed to set experiment status")
	}
}
