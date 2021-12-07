package sorters

type ModelAlreadyLoadedSorter struct{}

func (m ModelAlreadyLoadedSorter) IsLess(i *CandidateReplica, j *CandidateReplica) bool {
	iIsLoading := i.Model.IsLoadingOrLoaded(i.Replica.GetReplicaIdx())
	jIsLoading := j.Model.IsLoadingOrLoaded(j.Replica.GetReplicaIdx())
	return iIsLoading && !jIsLoading
}
