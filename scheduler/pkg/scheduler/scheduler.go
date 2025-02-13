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
		msg := "Failed to schedule model as no matching servers are available"
		logger.Debug(msg)
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
			msg := "Failed to schedule model as no matching server had enough suitable replicas, managed to schedule with min replicas"
			logger.Warn(msg)
		}
	}

	var serverEvent *coordinator.ServerEventMsg
	if !ok {
		serverEvent = s.serverScaleUp(latestModel)
		if !okWithMinReplicas {
			msg := "Failed to schedule model as no matching server had enough suitable replicas"
			logger.Debug(msg)
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
		UpdateContext: coordinator.SERVER_SCALE,
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
	for _, server := range servers {
		ok := true
		for _, serverFilter := range s.serverFilters {
			if !serverFilter.Filter(model, server) {
				logger.
					WithField("filter", serverFilter.Name()).
					WithField("server", server.Name).
					WithField("reason", serverFilter.Description(model, server)).
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

	return filteredServers
}

func (s *SimpleScheduler) filterReplicas(model *store.ModelVersion, server *store.ServerSnapshot) *sorters.CandidateServer {
	logger := s.logger.
		WithField("func", "filterReplicas").
		WithField("model", model.GetMeta().GetName()).
		WithField("server", server.Name)
	logger.Debug("Filtering server replicas for model")

	candidateServer := sorters.CandidateServer{Model: model, Server: server}
	for _, replica := range server.Replicas {
		ok := true
		for _, replicaFilter := range s.replicaFilters {
			if !replicaFilter.Filter(model, replica) {
				logger.
					WithField("filter", replicaFilter.Name()).
					WithField("replica", replica.GetReplicaIdx()).
					WithField("reason", replicaFilter.Description(model, replica)).
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

	return &candidateServer
}
