/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package scheduler

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/coordinator"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/scheduler/filters"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/scheduler/sorters"
	store "github.com/seldonio/seldon-core/scheduler/v2/pkg/store"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/synchroniser"
)

const serverScaleupEventSource = "scheduler.server.scaleup"

type SimpleScheduler struct {
	muSortAndUpdate sync.Mutex
	store           store.ModelStore
	logger          log.FieldLogger
	synchroniser    synchroniser.Synchroniser
	eventHub        *coordinator.EventHub
	SchedulerConfig
}

type SchedulerConfig struct {
	serverFilters  []filters.ServerFilter
	serverSorts    []sorters.ServerSorter
	replicaFilters []filters.ReplicaFilter
	replicaSorts   []sorters.ReplicaSorter
}

func DefaultSchedulerConfig(store store.ModelStore) SchedulerConfig {
	return SchedulerConfig{
		serverFilters:  []filters.ServerFilter{filters.ServerReplicaFilter{}, filters.SharingServerFilter{}, filters.DeletedServerFilter{}, filters.ServerRequirementFilter{}},
		replicaFilters: []filters.ReplicaFilter{filters.AvailableMemoryReplicaFilter{}, filters.ExplainerFilter{}, filters.ReplicaDrainingFilter{}},
		serverSorts:    []sorters.ServerSorter{sorters.ModelAlreadyLoadedOnServerSorter{}},
		replicaSorts:   []sorters.ReplicaSorter{sorters.ReplicaIndexSorter{}, sorters.AvailableMemorySorter{}, sorters.ModelAlreadyLoadedSorter{}},
	}
}

func NewSimpleScheduler(logger log.FieldLogger,
	store store.ModelStore,
	schedulerConfig SchedulerConfig,
	synchroniser synchroniser.Synchroniser,
	eventHub *coordinator.EventHub,
) *SimpleScheduler {
	s := &SimpleScheduler{
		store:           store,
		logger:          logger.WithField("Name", "SimpleScheduler"),
		SchedulerConfig: schedulerConfig,
		synchroniser:    synchroniser,
		eventHub:        eventHub,
	}
	return s
}

func (s *SimpleScheduler) Schedule(modelKey string) error {
	s.synchroniser.WaitReady()
	serverEvent, err := s.scheduleToServer(modelKey)
	if serverEvent != nil {
		s.logger.Debugf("Sending server event for %s", serverEvent.ServerName)
		s.eventHub.PublishServerEvent(serverScaleupEventSource, *serverEvent)
	}
	return err
}

func (s *SimpleScheduler) ScheduleFailedModels() ([]string, error) {
	s.synchroniser.WaitReady()
	failedModels, err := s.getFailedModels()
	if err != nil {
		return nil, err
	}
	var updatedModels []string
	for _, modelName := range failedModels {
		_, err := s.scheduleToServer(modelName)
		if err != nil {
			s.logger.Debugf("Failed to schedule failed model %s", modelName)
		} else {
			updatedModels = append(updatedModels, modelName)
		}
	}
	return updatedModels, nil
}

// Get failed models
// Currently this includes:
// - models that have failed to schedule
// - models that have failed to load
// - models that have loaded but not all replicas are available (e.g. min replicas is met but not desired replicas)
func (s *SimpleScheduler) getFailedModels() ([]string, error) {
	models, err := s.store.GetModels()
	if err != nil {
		return nil, err
	}
	var failedModels []string
	for _, model := range models {
		version := model.GetLatest()
		if version != nil {
			versionState := version.ModelState()
			if versionState.State == store.ModelFailed || versionState.State == store.ScheduleFailed ||
				(versionState.State == store.ModelAvailable && versionState.AvailableReplicas < version.GetDeploymentSpec().GetReplicas()) {
				failedModels = append(failedModels, model.Name)
			}
		}
	}
	return failedModels, nil
}

// TODO - clarify non shared models should not be scheduled
func (s *SimpleScheduler) scheduleToServer(modelName string) (*coordinator.ServerEventMsg, error) {
	logger := s.logger.WithField("func", "scheduleToServer").WithField("model", modelName)
	logger.Debug("Schedule model")

	s.store.LockModel(modelName)
	defer s.store.UnlockModel(modelName)

	// Get Model
	model, err := s.store.GetModel(modelName)
	if err != nil {
		return nil, err
	}
	if model == nil {
		return nil, errors.New("Unable to find model")
	}

	latestModel := model.GetLatest()
	if latestModel == nil {
		return nil, errors.New("Unable to find latest version for model")
	}

	if model.Deleted {
		// we need to call UpdateLoadedModels anyway:
		// - in case where we are deleting a model that doesnt have a server (FailedSchedule), server is ""
		// - otherwise proceed a normal
		server := ""
		if latestModel.HasServer() {
			server = latestModel.Server()
		}

		logger.Debug("Ensuring deleted model is removed")

		// remove old versions of the model if they exist
		// this is because the latest version might not have been applied properly and therefore
		// the old version might still be dangling around
		for _, mv := range model.Versions {
			if mv.HasLiveReplicas() {
				_, err := s.store.UnloadVersionModels(modelName, mv.GetVersion())
				if err != nil {
					logger.WithError(err).Warnf("Failed to unload model %s version %d", modelName, mv.GetVersion())
				}
			}
		}

		err = s.store.UpdateLoadedModels(modelName, latestModel.GetVersion(), server, []*store.ServerReplica{})
		if err != nil {
			logger.WithError(err).WithField("server", server).Warn("Failed to unschedule model replicas from server")
		}

		return nil, nil
	}

	// Model needs to be (re)scheduled
	var filteredServers []*store.ServerSnapshot

	// Get all servers
	servers, err := s.store.GetServers(false, true)
	if err != nil {
		return nil, err
	}

	// Filter and sort servers
	filteredServers = s.filterServers(latestModel, servers)
	if len(filteredServers) == 0 {
		totalServers := len(servers)
		modelName := latestModel.GetMeta().GetName()
		
		msg := fmt.Sprintf("Failed to schedule model '%s' as no matching servers are available (checked %d servers)", 
			modelName, totalServers)
		
		logger.WithField("total_servers_checked", totalServers).
			WithField("model_requirements", getModelRequirementsStr(latestModel)).
			Info(msg)
			
		s.store.FailedScheduling(latestModel, msg, !latestModel.HasLiveReplicas())
		return nil, errors.New(msg)
	}

	desiredReplicas := latestModel.DesiredReplicas()
	minReplicas := latestModel.GetDeploymentSpec().GetMinReplicas()

	s.sortServers(latestModel, filteredServers)
	logger.
		WithField("candidate_servers", filteredServers).
		WithField("desired_replicas", desiredReplicas).
		WithField("min_replicas", minReplicas).
		Debug("Identified candidate servers for model")

	// The main logic of trying to find a server for the model is as follows:
	// 1. If there are enough replicas on a server, schedule the model
	// 2. If there are not enough replicas on a server, try to schedule with min replicas. In this case we actually should get
	// the models loaded on all the replicas of the servers (assuming min replicas is less than the number of replicas on the server)
	// we also mark the model in this case as failed to schedule so that if the infra changes in the future we can try to reschedule

	// For each server filter and sort replicas and attempt schedule if enough replicas
	ok := s.findAndUpdateToServers(filteredServers, latestModel, desiredReplicas, desiredReplicas)
	// Try to scheduler with min replicas if not enough replicas
	okWithMinReplicas := false
	if !ok && minReplicas > 0 {
		okWithMinReplicas = s.findAndUpdateToServers(filteredServers, latestModel, desiredReplicas, int(minReplicas))
		if okWithMinReplicas {
			msg := fmt.Sprintf("Failed to schedule model '%s' with desired replicas (%d), managed to schedule with min replicas (%d)", 
				latestModel.GetMeta().GetName(), desiredReplicas, minReplicas)
			logger.WithField("desired_replicas", desiredReplicas).
				WithField("min_replicas", minReplicas).
				WithField("available_servers", len(filteredServers)).
				Warn(msg)
		}
	}

	var serverEvent *coordinator.ServerEventMsg
	if !ok {
		serverEvent = s.serverScaleUp(latestModel)
		if !okWithMinReplicas {
			msg := fmt.Sprintf("Failed to schedule model '%s' as no matching server had enough suitable replicas (desired: %d, min: %d, servers checked: %d)", 
				latestModel.GetMeta().GetName(), desiredReplicas, minReplicas, len(filteredServers))
			logger.WithField("desired_replicas", desiredReplicas).
				WithField("min_replicas", minReplicas).
				WithField("servers_checked", len(filteredServers)).
				WithField("model_requirements", getModelRequirementsStr(latestModel)).
				Info(msg)
			// we do not want to reset the server if it has live replicas or loading replicas
			// in the case of loading replicas, we need to make sure that we can unload them later.
			// for example in the case that a model is just marked as loading on a particular server replica
			// then it gets a delete request (before it is marked as loaded or available) we need to make sure
			// that we can unload it from the server
			s.store.FailedScheduling(latestModel, msg, !latestModel.HasLiveReplicas() && !latestModel.IsLoadingOrLoadedOnServer())
			return serverEvent, errors.New(msg)
		}
	}

	// TODO Cleanup previous version if needed?
	return serverEvent, nil
}

func (s *SimpleScheduler) findAndUpdateToServers(filteredServers []*store.ServerSnapshot, latestModel *store.ModelVersion, desiredReplicas, minReplicas int) bool {
	modelName := latestModel.GetMeta().GetName()
	logger := s.logger.WithField("func", "findAndUpdateToServers").WithField("model", modelName)
	ok := false
	for _, candidateServer := range filteredServers {
		logger.WithField("server", candidateServer.Name).Debug("Checking compatibility with candidate server")
		var candidateReplicas *sorters.CandidateServer

		// we need a lock here, we could have many goroutines at sorting
		// without the store being reflected and hence sorting on stale values
		s.muSortAndUpdate.Lock()
		candidateReplicas = s.filterReplicas(latestModel, candidateServer)
		numServerReplicas := len(candidateReplicas.ChosenReplicas)
		if numServerReplicas < minReplicas {
			logger.
				WithField("server", candidateServer.Name).
				WithField("available_replicas", numServerReplicas).
				WithField("desired_replicas", desiredReplicas).
				WithField("min_replicas", minReplicas).
				Debug("Skipping server due to insufficient suitable replicas")

			s.muSortAndUpdate.Unlock()
			continue
		}

		s.sortReplicas(candidateReplicas)
		numReplicas := minReplicas
		if minReplicas != desiredReplicas {
			numReplicas = min(numServerReplicas, desiredReplicas) // we have more replicas for the server than min, so we can use all of them
		}
		err := s.store.UpdateLoadedModels(
			modelName,
			latestModel.GetVersion(),
			candidateServer.Name,
			candidateReplicas.ChosenReplicas[0:numReplicas],
		)
		s.muSortAndUpdate.Unlock()

		if err != nil {
			logger.WithField("server", candidateServer.Name).Warn("Failed to update model replicas")
		} else {
			logger.WithField("server", candidateServer.Name).Debug("Scheduled model onto server")
			ok = true
			break
		}
	}
	return ok
}

func showServerSlice(servers []*store.ServerSnapshot) string {
	var sb strings.Builder
	for idx, server := range servers {
		if idx > 0 {
			sb.WriteString(",")
		}
		sb.WriteString(server.Name)
	}
	return sb.String()
}

func (s *SimpleScheduler) sortServers(model *store.ModelVersion, server []*store.ServerSnapshot) {
	logger := s.logger.WithField("func", "sortServers")
	for _, sorter := range s.serverSorts {
		logger.Debugf("About to sort servers for %s:%d with %s: %s", model.Key(), model.GetVersion(), sorter.Name(), showServerSlice(server))
		sort.SliceStable(server, func(i, j int) bool {
			return sorter.IsLess(&sorters.CandidateServer{Model: model, Server: server[i]}, &sorters.CandidateServer{Model: model, Server: server[j]})
		})
		logger.Debugf("Sorted servers for %s:%d with %s: %s", model.Key(), model.GetVersion(), sorter.Name(), showServerSlice(server))
	}
}

func (s *SimpleScheduler) serverScaleUp(modelVersion *store.ModelVersion) *coordinator.ServerEventMsg {
	logger := s.logger.WithField("func", "serverScaleUp")

	if modelVersion.Server() == "" {
		logger.Warnf("Empty server for %s so ignoring scale up request", modelVersion.GetMeta().Name)
		return nil
	}

	return &coordinator.ServerEventMsg{
		ServerName:    modelVersion.Server(),
		UpdateContext: coordinator.SERVER_SCALE_UP,
	}
}

func showReplicaSlice(candidateServer *sorters.CandidateServer) string {
	var sb strings.Builder
	for idx, replica := range candidateServer.ChosenReplicas {
		if idx > 0 {
			sb.WriteString(",")
		}
		sb.WriteString(strconv.Itoa(replica.GetReplicaIdx()))
		sb.WriteString(":")
		sb.WriteString(replica.GetInferenceSvc())
	}
	return sb.String()
}

func (s *SimpleScheduler) sortReplicas(candidateServer *sorters.CandidateServer) {
	logger := s.logger.WithField("func", "sortReplicas")
	for _, sorter := range s.replicaSorts {
		logger.Debugf("About to sort replicas for %s:%d with %s: %s", candidateServer.Model.Key(), candidateServer.Model.GetVersion(), sorter.Name(), showReplicaSlice(candidateServer))
		sort.SliceStable(candidateServer.ChosenReplicas, func(i, j int) bool {
			return sorter.IsLess(&sorters.CandidateReplica{Model: candidateServer.Model, Server: candidateServer.Server, Replica: candidateServer.ChosenReplicas[i]},
				&sorters.CandidateReplica{Model: candidateServer.Model, Server: candidateServer.Server, Replica: candidateServer.ChosenReplicas[j]})
		})
		logger.Debugf("Sorted replicas for %s:%d with %s: %s", candidateServer.Model.Key(), candidateServer.Model.GetVersion(), sorter.Name(), showReplicaSlice(candidateServer))
	}
}

// Filter servers for this model
func (s *SimpleScheduler) filterServers(model *store.ModelVersion, servers []*store.ServerSnapshot) []*store.ServerSnapshot {
	logger := s.logger.WithField("func", "filterServer").WithField("model", model.GetMeta().GetName())
	logger.WithField("num_servers", len(servers)).Debug("Filtering servers for model")

	var filteredServers []*store.ServerSnapshot
	var rejectionReasons []string
	
	for _, server := range servers {
		ok := true
		for _, serverFilter := range s.serverFilters {
			if !serverFilter.Filter(model, server) {
				reason := serverFilter.Description(model, server)
				rejectionReasons = append(rejectionReasons, fmt.Sprintf("Server '%s' rejected by %s: %s", server.Name, serverFilter.Name(), reason))
				
				logger.
					WithField("filter", serverFilter.Name()).
					WithField("server", server.Name).
					WithField("reason", reason).
					Debug("Rejecting server for model")

				ok = false
				break
			}
		}

		if ok {
			logger.WithField("server", server.Name).Debug("Accepting server for model")
			filteredServers = append(filteredServers, server)
		}
	}

	// Store rejection reasons for verbose error reporting
	if len(filteredServers) == 0 && len(rejectionReasons) > 0 {
		logger.WithField("rejection_details", rejectionReasons).Info("All servers rejected for model scheduling")
		// Store rejection reasons in model metadata for CLI access
		s.storeRejectionReasons(model, rejectionReasons)
	}

	return filteredServers
}

func (s *SimpleScheduler) filterReplicas(model *store.ModelVersion, server *store.ServerSnapshot) *sorters.CandidateServer {
	logger := s.logger.
		WithField("func", "filterReplicas").
		WithField("model", model.GetMeta().GetName()).
		WithField("server", server.Name)
	logger.Debug("Filtering server replicas for model")

	candidateServer := sorters.CandidateServer{Model: model, Server: server}
	var rejectedReplicas []string
	
	for _, replica := range server.Replicas {
		ok := true
		for _, replicaFilter := range s.replicaFilters {
			if !replicaFilter.Filter(model, replica) {
				reason := replicaFilter.Description(model, replica)
				rejectedReplicas = append(rejectedReplicas, 
					fmt.Sprintf("Replica %d rejected by %s: %s", replica.GetReplicaIdx(), replicaFilter.Name(), reason))
				
				logger.
					WithField("filter", replicaFilter.Name()).
					WithField("replica", replica.GetReplicaIdx()).
					WithField("reason", reason).
					Debug("Rejecting server replica for model")

				ok = false
				break
			}
		}

		if ok {
			logger.WithField("replica", replica.GetReplicaIdx()).Debug("Accepting server replica for model")
			candidateServer.ChosenReplicas = append(candidateServer.ChosenReplicas, replica)
		}
	}

	// Log detailed replica rejection information when no replicas are available
	if len(candidateServer.ChosenReplicas) == 0 && len(rejectedReplicas) > 0 {
		logger.WithField("replica_rejections", rejectedReplicas).
			WithField("total_replicas_checked", len(server.Replicas)).
			Info("No suitable replicas found on server for model")
	}

	return &candidateServer
}

// storeRejectionReasons stores detailed rejection reasons in the model's metadata
// for later access by CLI tools and debugging
func (s *SimpleScheduler) storeRejectionReasons(model *store.ModelVersion, reasons []string) {
	// For now, we'll log the detailed reasons at Info level so they appear in logs
	// This could be enhanced to store in model metadata when that capability is available
	reasonStr := strings.Join(reasons, "; ")
	s.logger.WithField("model", model.GetMeta().GetName()).
		WithField("detailed_reasons", reasonStr).
		Info("Model scheduling failed - detailed reasons available")
}

// getModelRequirementsStr returns a human-readable string of model requirements for debugging
func getModelRequirementsStr(model *store.ModelVersion) string {
	deploymentSpec := model.GetDeploymentSpec()
	modelSpec := model.GetModel().GetModelSpec()
	requirements := []string{}
	
	// Add replica requirements
	if deploymentSpec.GetReplicas() > 0 {
		requirements = append(requirements, fmt.Sprintf("replicas=%d", deploymentSpec.GetReplicas()))
	}
	
	if deploymentSpec.GetMinReplicas() > 0 {
		requirements = append(requirements, fmt.Sprintf("min_replicas=%d", deploymentSpec.GetMinReplicas()))
	}
	
	// Add memory requirements if available
	if modelSpec.GetMemoryBytes() > 0 {
		memoryMB := modelSpec.GetMemoryBytes() / (1024 * 1024)
		requirements = append(requirements, fmt.Sprintf("memory=%dMB", memoryMB))
	}
	
	// Add server requirements if specified
	if modelSpec.GetServer() != "" {
		requirements = append(requirements, fmt.Sprintf("server=%s", modelSpec.GetServer()))
	}
	
	// Add capability requirements if specified
	if len(modelSpec.GetRequirements()) > 0 {
		requirements = append(requirements, fmt.Sprintf("capabilities=%v", modelSpec.GetRequirements()))
	}
	
	if len(requirements) == 0 {
		return "none specified"
	}
	
	return strings.Join(requirements, ", ")
}
