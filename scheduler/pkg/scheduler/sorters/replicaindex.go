package sorters

// Rationale: Lower indexed replicas will exist for longer as they will be scaled down later than a higher index
type ReplicaIndexSorter struct{}

func (m ReplicaIndexSorter) Name() string {
	return "ReplicaIndexSorter"
}

func (r ReplicaIndexSorter) IsLess(i *CandidateReplica, j *CandidateReplica) bool {
	return i.Replica.GetReplicaIdx() < j.Replica.GetReplicaIdx()
}
