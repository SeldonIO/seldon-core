package scheduler

import (
	"fmt"
	"github.com/seldonio/seldon-core/scheduler/pkg/scheduler/sorters"
	store "github.com/seldonio/seldon-core/scheduler/pkg/store"
	log "github.com/sirupsen/logrus"
	"sort"
	"sync"
)

type SimpleScheduler struct {
	mu     sync.RWMutex
	store  store.SchedulerStore
	logger log.FieldLogger
	serverFilters []ServerFilter
	serverSorts []sorters.ServerSorter
	replicaFilters []ReplicaFilter
	replicaSorts []sorters.ReplicaSorter
	failedModels map[string]bool
}

func NewSimpleScheduler(logger log.FieldLogger,
	store store.SchedulerStore,
	serverFilters []ServerFilter,
	replicaFilters []ReplicaFilter,
	serverSorts []sorters.ServerSorter,
	replicaSorts []sorters.ReplicaSorter) *SimpleScheduler {
	return &SimpleScheduler{
		store: store,
		logger: logger,
		serverFilters: serverFilters,
		serverSorts: serverSorts,
		replicaFilters: replicaFilters,
		replicaSorts: replicaSorts,
		failedModels: make(map[string]bool),
	}
}

func (s *SimpleScheduler) Schedule(modelKey string) error {
	err := s.scheduleToServer(modelKey)
	// Set model state using error
	if err != nil {
		s.mu.Lock()
		defer s.mu.Unlock()
		s.failedModels[modelKey] = true
		return err
	}
	return nil
}

func (s *SimpleScheduler) ScheduleFailedModels() ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var updatedModels []string
	for modelName,_ := range s.failedModels {
		err := s.scheduleToServer(modelName)
		if err != nil {
			s.logger.Debugf("Failed to schedule failed model %s",modelName)
		} else {
			updatedModels = append(updatedModels, modelName)
		}
	}
	return updatedModels, nil
}

//TODO - clarify non shared models should not be scheduled
func (s *SimpleScheduler) scheduleToServer(modelKey string) error {
	s.logger.Debugf("Schedule model %s",modelKey)
	// Get Model
	model, err := s.store.GetModel(modelKey)
	if err != nil {
		return err
	}
	if model == nil {
		fmt.Errorf("Can't find model with key %s",modelKey)
	}
	latestModel := model.GetLatest()
	if latestModel == nil {
		return fmt.Errorf("No latest model for %s", modelKey)
	}

	if model.Deleted && latestModel.HasServer(){
		s.logger.Debugf("Model %s is deleted ensuring removed",modelKey)
		err = s.store.UpdateLoadedModels(modelKey, latestModel.GetVersion(), latestModel.Server(), []*store.ServerReplica{})
		if err != nil {
			s.logger.Warnf("Failed to unschedule model replicas for model %s on server %s",modelKey,latestModel.Server())
		}
	} else {
		// Get all servers
		servers, err := s.store.GetServers()
		if err != nil {
			return err
		}
		// Filter and sort servers
		filteredServers := s.filterServers(latestModel, servers)
		s.sortServers(latestModel, filteredServers)
		ok := false
		s.logger.Debugf("Model %s candidate servers %v",modelKey, filteredServers)
		// For each server filter and sort replicas and attempt schedule if enough replicas
		for _,candidateServer := range filteredServers {
			candidateReplicas := s.filterReplicas(latestModel, candidateServer)
			if len(candidateReplicas.ChosenReplicas) < latestModel.DesiredReplicas() {
				continue
			}
			s.sortReplicas(candidateReplicas)

			err = s.store.UpdateLoadedModels(modelKey, latestModel.GetVersion(), candidateServer.Name, candidateReplicas.ChosenReplicas[0:latestModel.DesiredReplicas()])
			if err != nil {
				s.logger.Warnf("Failed to update model replicas for model %s on server %s",modelKey,candidateServer.Name)
			} else {
				ok = true
				break
			}
		}
		if !ok {
			return fmt.Errorf("failed to schedule model %s", modelKey)
		}
	}

	//TODO Cleanup other model versions
	return nil
}

func (s *SimpleScheduler) sortServers(model *store.ModelVersion, server []*store.ServerSnapshot) {
	for _,sorter := range s.serverSorts {
		sort.SliceStable(server, func(i, j int) bool { return sorter.IsLess(&sorters.CandidateServer{Model: model, Server: server[i]}, &sorters.CandidateServer{Model: model, Server: server[j]})})
	}
}

func (s *SimpleScheduler) sortReplicas(candidateServer *sorters.CandidateServer) {
	for _,sorter := range s.replicaSorts {
		sort.SliceStable(candidateServer.ChosenReplicas, func(i, j int) bool {
			return sorter.IsLess(&sorters.CandidateReplica{Model: candidateServer.Model, Server: candidateServer.Server, Replica: candidateServer.ChosenReplicas[i]},
				&sorters.CandidateReplica{Model: candidateServer.Model, Server: candidateServer.Server, Replica: candidateServer.ChosenReplicas[j]})})
	}
}

// Filter servers for this model
func (s *SimpleScheduler) filterServers(model *store.ModelVersion, servers []*store.ServerSnapshot) []*store.ServerSnapshot {
	var filteredServers []*store.ServerSnapshot
	for _,server := range servers {
		ok := true
		for _, serverFilter := range s.serverFilters {
			if !serverFilter.Filter(model, server) {
				ok = false
				break
			}
		}
		if ok {
			filteredServers = append(filteredServers, server)
		}
	}
	return filteredServers
}


func (s *SimpleScheduler) filterReplicas(model *store.ModelVersion, server *store.ServerSnapshot) *sorters.CandidateServer {
	candidateServer := sorters.CandidateServer{Model: model, Server: server}
	for _,replica := range server.Replicas {
		ok := true
		for _,replicaFilter := range s.replicaFilters {
			if !replicaFilter.Filter(model, replica) {
				ok = false
				break
			}
		}
		if ok {
			candidateServer.ChosenReplicas = append(candidateServer.ChosenReplicas, replica)
		}
	}
	return &candidateServer
}


