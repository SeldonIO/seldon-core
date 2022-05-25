package sorters

import (
	"math"

	"github.com/seldonio/seldon-core/scheduler/pkg/store"
)

// Rationale - put models on replicas with more memory, including the models that are currently loading
// note that we can double count here as available memory (returned by the agent) could include memory
// that is allocated while the model is being loaded.
type AvailableMemoryWhileLoadingSorter struct {
	Store store.ModelStore
}

func (m AvailableMemoryWhileLoadingSorter) Name() string {
	return "AvailableMemorySorter"
}

func (m AvailableMemoryWhileLoadingSorter) IsLess(i *CandidateReplica, j *CandidateReplica) bool {
	iMem := math.Max(0, float64(i.Replica.GetAvailableMemory()-getTotalMemoryBytesForLoadingModels(i, m.Store)))
	jMem := math.Max(0, float64(j.Replica.GetAvailableMemory()-getTotalMemoryBytesForLoadingModels(j, m.Store)))
	return iMem > jMem
}

func getTotalMemoryBytesForLoadingModels(r *CandidateReplica, s store.ModelStore) uint64 {
	mem := uint64(0)
	if s == nil {
		return mem
	}
	models, err := s.GetModels()
	if err != nil {
		return mem
	}
	for _, model := range models {
		for _, version := range model.Versions {
			if version.Server() == r.Server.Name {
				replicaState := version.GetModelReplicaState(r.Replica.GetReplicaIdx())
				if replicaState.IsLoading() {
					mem += version.GetRequiredMemory()
				}
			}
		}
	}
	return mem
}
