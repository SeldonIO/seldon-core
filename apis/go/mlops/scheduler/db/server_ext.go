package db

import (
	"slices"
)

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

func (s *ServerReplica) AddModelVersion(modelName string, modelVersion uint32, replicaState ModelReplicaState) {
	mvID := &ModelVersionID{
		Name:    modelName,
		Version: modelVersion,
	}

	if replicaState == ModelReplicaState_MODEL_REPLICA_STATE_LOADING {
		s.addToList(&s.LoadingModels, mvID)
		return
	}

	if replicaState == ModelReplicaState_MODEL_REPLICA_STATE_LOADED {
		s.removeFromList(&s.LoadingModels, mvID)
		s.addToList(&s.LoadedModels, mvID)
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
