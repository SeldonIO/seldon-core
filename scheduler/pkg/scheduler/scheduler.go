/*
Copyright 2022 Seldon Technologies Ltd.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package scheduler

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/scheduler/filters"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/scheduler/sorters"
	store "github.com/seldonio/seldon-core/scheduler/v2/pkg/store"
)

type SimpleScheduler struct {
	muSortAndUpdate sync.Mutex
	store           store.ModelStore
	logger          log.FieldLogger
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
		serverSorts:    []sorters.ServerSorter{},
		replicaSorts:   []sorters.ReplicaSorter{sorters.ReplicaIndexSorter{}, sorters.AvailableMemorySorter{}, sorters.ModelAlreadyLoadedSorter{}},
	}
}

func NewSimpleScheduler(logger log.FieldLogger,
	store store.ModelStore,
	schedulerConfig SchedulerConfig) *SimpleScheduler {
	s := &SimpleScheduler{
		store:           store,
		logger:          logger.WithField("Name", "SimpleScheduler"),
		SchedulerConfig: schedulerConfig,
	}
	return s
}

func (s *SimpleScheduler) Schedule(modelKey string) error {
	return s.scheduleToServer(modelKey)
}

func (s *SimpleScheduler) ScheduleFailedModels() ([]string, error) {
	failedModels, err := s.getFailedModels()
	if err != nil {
		return nil, err
	}
	var updatedModels []string
	for _, modelName := range failedModels {
		err := s.scheduleToServer(modelName)
		if err != nil {
			s.logger.Debugf("Failed to schedule failed model %s", modelName)
		} else {
			updatedModels = append(updatedModels, modelName)
		}
	}
	return updatedModels, nil
}

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
			if versionState.State == store.ModelFailed || versionState.State == store.ScheduleFailed {
				failedModels = append(failedModels, model.Name)
			}
		}
	}
	return failedModels, nil
}

// TODO - clarify non shared models should not be scheduled
func (s *SimpleScheduler) scheduleToServer(modelName string) error {
	logger := s.logger.WithField("func", "scheduleToServer").WithField("model", modelName)
	logger.Debug("Schedule model")

	s.store.LockModel(modelName)
	defer s.store.UnlockModel(modelName)

	// Get Model
	model, err := s.store.GetModel(modelName)
	if err != nil {
		return err
	}
	if model == nil {
		return fmt.Errorf("Can't find model with key %s", modelName)
	}

	latestModel := model.GetLatest()
	if latestModel == nil {
		return fmt.Errorf("No latest model for %s", modelName)
	}

	if model.Deleted {
		// we need to LoadedModels anyway:
		// - in case where we are deleting a model that doesnt have a server (FailedSchedule), server is ""
		// - otherwise proceed a normal
		server := ""
		if latestModel.HasServer() {
			server = latestModel.Server()
		}

		logger.Debug("Ensuring deleted model is removed", modelName)
		err = s.store.UpdateLoadedModels(modelName, latestModel.GetVersion(), server, []*store.ServerReplica{})
		if err != nil {
			logger.WithError(err).WithField("server", server).Warn("Failed to unschedule model replicas from server")
		}
	} else {
		var debugTrail []string
		var filteredServers []*store.ServerSnapshot

		// Get all servers
		servers, err := s.store.GetServers(false, true)
		if err != nil {
			return err
		}

		// Filter and sort servers
		filteredServers, debugTrail = s.filterServers(latestModel, servers, debugTrail)
		s.sortServers(latestModel, filteredServers)
		ok := false
		logger.
			WithField("candidate_servers", filteredServers).
			WithField("desired_replicas", latestModel.DesiredReplicas()).
			Debug("Identified candidate servers for model")

		// For each server filter and sort replicas and attempt schedule if enough replicas
		for _, candidateServer := range filteredServers {
			logger.WithField("server", candidateServer.Name).Debug("Checking compatibility with candidate server")
			var candidateReplicas *sorters.CandidateServer

			// we need a lock here, we could have many goroutines at sorting
			// without the store being reflected and hence sorting on stale values
			s.muSortAndUpdate.Lock()
			candidateReplicas, debugTrail = s.filterReplicas(latestModel, candidateServer, debugTrail)
			if len(candidateReplicas.ChosenReplicas) < latestModel.DesiredReplicas() {
				logger.
					WithField("server", candidateServer.Name).
					WithField("available_replicas", len(candidateReplicas.ChosenReplicas)).
					WithField("desired_replicas", latestModel.DesiredReplicas()).
					Debug("Skipping server due to insufficient available replicas")

				s.muSortAndUpdate.Unlock()
				continue
			}

			s.sortReplicas(candidateReplicas)
			err = s.store.UpdateLoadedModels(modelName, latestModel.GetVersion(), candidateServer.Name, candidateReplicas.ChosenReplicas[0:latestModel.DesiredReplicas()])
			s.muSortAndUpdate.Unlock()

			if err != nil {
				logger.WithField("server", candidateServer.Name).Warn("Failed to update model replicas")
			} else {
				ok = true
				break
			}
		}

		if !ok {
			failureErrMsg := fmt.Sprintf("failed to schedule model %s. %v", modelName, debugTrail)
			// we do not want to reset the server if it has live replicas
			s.store.FailedScheduling(latestModel, failureErrMsg, !latestModel.HasLiveReplicas())
			return fmt.Errorf(failureErrMsg)
		}
	}

	//TODO Cleanup previous version if needed?
	return nil
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
func (s *SimpleScheduler) filterServers(model *store.ModelVersion, servers []*store.ServerSnapshot, debugTrail []string) ([]*store.ServerSnapshot, []string) {
	logger := s.logger.WithField("func", "filterServer")

	var filteredServers []*store.ServerSnapshot
	for _, server := range servers {
		ok := true
		for _, serverFilter := range s.serverFilters {
			if !serverFilter.Filter(model, server) {
				msg := fmt.Sprintf(
					"failed server filter %s for server replica %s : %s",
					serverFilter.Name(),
					server.Name,
					serverFilter.Description(model, server),
				)
				logger.Debugf(msg)
				debugTrail = append(debugTrail, msg)

				ok = false
				break
			}
		}

		if ok {
			filteredServers = append(filteredServers, server)
		}
	}

	return filteredServers, debugTrail
}

func (s *SimpleScheduler) filterReplicas(model *store.ModelVersion, server *store.ServerSnapshot, debugTrail []string) (*sorters.CandidateServer, []string) {
	logger := s.logger.WithField("func", "filterReplicas")
	candidateServer := sorters.CandidateServer{Model: model, Server: server}
	for _, replica := range server.Replicas {
		ok := true
		for _, replicaFilter := range s.replicaFilters {
			if !replicaFilter.Filter(model, replica) {
				msg := fmt.Sprintf("failed replica filter %s for server replica %s:%d : %s",
					replicaFilter.Name(),
					server.Name,
					replica.GetReplicaIdx(),
					replicaFilter.Description(model, replica))
				logger.Debugf(msg)
				debugTrail = append(debugTrail, msg)
				ok = false
				break
			}
		}
		if ok {
			candidateServer.ChosenReplicas = append(candidateServer.ChosenReplicas, replica)
		}
	}
	return &candidateServer, debugTrail
}
