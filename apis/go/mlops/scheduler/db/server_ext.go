/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package db

import (
	"slices"
)

func (s *Server) initReplicas() {
	if s.Replicas == nil {
		s.Replicas = map[int32]*ServerReplica{}
	}
}

func (s *Server) AddReplica(replicaID int32, replica *ServerReplica) {
	s.initReplicas()
	s.Replicas[replicaID] = replica
}

func (s *ServerReplica) GetLoadedOrLoadingModelVersions() []*ModelVersionID {
	var models []*ModelVersionID
	models = append(models, s.LoadedModels...)
	models = append(models, s.LoadingModels...)
	return models
}

func (s *ServerReplica) UpdateReservedMemory(memBytes uint64, isAdd bool) {
	if isAdd {
		s.ReservedMemory += memBytes
		return
	}
	if memBytes > s.ReservedMemory {
		s.ReservedMemory = 0
		return
	}
	s.ReservedMemory -= memBytes
}

func (s *ServerReplica) GetNumLoadedModels() int {
	return len(s.UniqueLoadedModels)
}

func (s *ServerReplica) initUniqueLoadedModels() {
	if s.UniqueLoadedModels == nil {
		s.UniqueLoadedModels = map[string]bool{}
	}
}

func (s *ServerReplica) AddModelVersion(modelName string, modelVersion uint32, replicaState ModelReplicaState) {
	mvID := &ModelVersionID{
		Name:    modelName,
		Version: modelVersion,
	}

	if replicaState == ModelReplicaState_Loading {
		s.addToList(&s.LoadingModels, mvID)
		return
	}

	if replicaState == ModelReplicaState_Loaded {
		s.removeFromList(&s.LoadingModels, mvID)
		s.addToList(&s.LoadedModels, mvID)
		s.initUniqueLoadedModels()
		s.UniqueLoadedModels[modelName] = true
	}
}

func (s *ServerReplica) DeleteModelVersion(modelName string, modelVersion uint32) {
	mvID := &ModelVersionID{
		Name:    modelName,
		Version: modelVersion,
	}

	s.removeFromList(&s.LoadingModels, mvID)
	s.removeFromList(&s.LoadedModels, mvID)

	if !s.modelExistsInList(s.LoadedModels, modelName) {
		if s.UniqueLoadedModels == nil {
			return
		}
		delete(s.UniqueLoadedModels, modelName)
	}
}

func (s *ServerReplica) addToList(list *[]*ModelVersionID, mvID *ModelVersionID) {
	for _, model := range *list {
		if model.Version == mvID.Version && model.Name == mvID.Name {
			return
		}
	}
	*list = append(*list, &ModelVersionID{
		Name:    mvID.Name,
		Version: mvID.Version,
	})
}

func (s *ServerReplica) removeFromList(list *[]*ModelVersionID, mvID *ModelVersionID) {
	*list = slices.DeleteFunc(*list, func(m *ModelVersionID) bool {
		return m.Version == mvID.Version && m.Name == mvID.Name
	})
}

func (s *ServerReplica) modelExistsInList(list []*ModelVersionID, modelName string) bool {
	for _, model := range list {
		if model.Name == modelName {
			return true
		}
	}
	return false
}
