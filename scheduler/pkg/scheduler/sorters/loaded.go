package sorters

type ModelAlreadyLoadedSorter struct{}

func (m ModelAlreadyLoadedSorter) IsLess(i *CandidateReplica, j *CandidateReplica) bool {
	iIsLoading := i.Model.IsLoading(i.Replica.GetReplicaIdx())
	jIsLoading := j.Model.IsLoading(j.Replica.GetReplicaIdx())
	return iIsLoading && !jIsLoading
}



