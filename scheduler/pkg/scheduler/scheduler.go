package scheduler

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/seldonio/seldon-core/scheduler/pkg/scheduler/filters"

	"github.com/seldonio/seldon-core/scheduler/pkg/scheduler/sorters"
	store "github.com/seldonio/seldon-core/scheduler/pkg/store"
	log "github.com/sirupsen/logrus"
)

type SimpleScheduler struct {
	muFailedModels  sync.RWMutex
	muSortAndUpdate sync.Mutex
	store           store.ModelStore
	logger          log.FieldLogger
	SchedulerConfig
	failedModels map[string]bool
}

type SchedulerConfig struct {
	serverFilters  []ServerFilter
	serverSorts    []sorters.ServerSorter
	replicaFilters []ReplicaFilter
	replicaSorts   []sorters.ReplicaSorter
}

func DefaultSchedulerConfig(store store.ModelStore) SchedulerConfig {
	return SchedulerConfig{
		serverFilters:  []ServerFilter{filters.SharingServerFilter{}, filters.DeletedServerFilter{}},
		replicaFilters: []ReplicaFilter{filters.RequirementsReplicaFilter{}, filters.AvailableMemoryReplicaFilter{}, filters.ExplainerFilter{}},
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
		failedModels:    make(map[string]bool),
	}
	return s
}

func (s *SimpleScheduler) Schedule(modelKey string) error {
	err := s.scheduleToServer(modelKey)
	// Set model state using error
	if err != nil {
		s.muFailedModels.Lock()
		defer s.muFailedModels.Unlock()
		s.failedModels[modelKey] = true
		return err
	}
	return nil
}

func (s *SimpleScheduler) ScheduleFailedModels() ([]string, error) {
	s.muFailedModels.RLock()
	defer s.muFailedModels.RUnlock()
	var updatedModels []string
	for modelName := range s.failedModels {
		err := s.scheduleToServer(modelName)
		if err != nil {
			s.logger.Debugf("Failed to schedule failed model %s", modelName)
		} else {
			updatedModels = append(updatedModels, modelName)
		}
	}
	return updatedModels, nil
}

//TODO - clarify non shared models should not be scheduled
func (s *SimpleScheduler) scheduleToServer(modelName string) error {
	logger := s.logger.WithField("func", "scheduleToServer")
	logger.Debugf("Schedule model %s", modelName)
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
		if latestModel.HasServer() {
			logger.Debugf("Model %s is deleted ensuring removed", modelName)
			err = s.store.UpdateLoadedModels(modelName, latestModel.GetVersion(), latestModel.Server(), []*store.ServerReplica{})
			if err != nil {
				logger.Warnf("Failed to unschedule model replicas for model %s on server %s", modelName, latestModel.Server())
			}
		}
		delete(s.failedModels, modelName) // Ensure model removed from failed models if its there
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
		logger.Debugf("Model %s candidate servers %v", modelName, filteredServers)
		// For each server filter and sort replicas and attempt schedule if enough replicas
		for _, candidateServer := range filteredServers {
			var candidateReplicas *sorters.CandidateServer
			candidateReplicas, debugTrail = s.filterReplicas(latestModel, candidateServer, debugTrail)
			if len(candidateReplicas.ChosenReplicas) < latestModel.DesiredReplicas() {
				continue
			}

			// we need a lock here, we could have many goroutines at sorting
			// without the store being reflected and hence storing on stale values
			s.muSortAndUpdate.Lock()
			s.sortReplicas(candidateReplicas)
			err = s.store.UpdateLoadedModels(modelName, latestModel.GetVersion(), candidateServer.Name, candidateReplicas.ChosenReplicas[0:latestModel.DesiredReplicas()])
			s.muSortAndUpdate.Unlock()

			if err != nil {
				logger.Warnf("Failed to update model replicas for model %s on server %s", modelName, candidateServer.Name)
			} else {
				ok = true
				break
			}
		}
		if !ok {
			failureErrMsg := fmt.Sprintf("failed to schedule model %s. %v", modelName, debugTrail)
			s.store.FailedScheduling(latestModel, failureErrMsg)
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
				msg := fmt.Sprintf("failed server filter %s for server replica %s : %s",
					serverFilter.Name(),
					server.Name,
					serverFilter.Description(model, server))
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
