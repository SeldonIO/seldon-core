package sorters

import (
	"math"
)

// Rationale - put models on replicas with more memory, including the models that are currently loading
// note that we can double count here as available memory (returned by the agent) could include memory
// that is allocated while the model is being loaded.
type AvailableMemorySorter struct{}

func (m AvailableMemorySorter) Name() string {
	return "AvailableMemorySorter"
}

func (m AvailableMemorySorter) IsLess(i *CandidateReplica, j *CandidateReplica) bool {
	iMem := math.Max(0, float64(i.Replica.GetAvailableMemory()-i.Replica.GetReservedMemory()))
	jMem := math.Max(0, float64(j.Replica.GetAvailableMemory()-j.Replica.GetReservedMemory()))
	return iMem > jMem
}
