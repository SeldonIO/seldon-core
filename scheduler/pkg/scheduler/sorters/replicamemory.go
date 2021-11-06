package sorters

type AvailableMemorySorter struct{}

func (m AvailableMemorySorter) IsLess(i *CandidateReplica, j *CandidateReplica) bool {
	return i.Replica.GetAvailableMemory() > j.Replica.GetAvailableMemory()
}