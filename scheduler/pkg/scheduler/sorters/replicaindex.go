package sorters

type ReplicaIndexSorter struct{}

func (m ReplicaIndexSorter) Name() string {
	return "ReplicaIndexSorter"
}

func (r ReplicaIndexSorter) IsLess(i *CandidateReplica, j *CandidateReplica) bool {
	return i.Replica.GetReplicaIdx() < j.Replica.GetReplicaIdx()
}
