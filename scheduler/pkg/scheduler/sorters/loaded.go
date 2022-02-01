package sorters

type ModelAlreadyLoadedSorter struct{}

func (m ModelAlreadyLoadedSorter) Name() string {
	return "ModelAlreadyLoadedSorter"
}

func (m ModelAlreadyLoadedSorter) IsLess(i *CandidateReplica, j *CandidateReplica) bool {
	iIsLoading := i.Model.IsLoadingOrLoaded(i.Server.Name, i.Replica.GetReplicaIdx())
	jIsLoading := j.Model.IsLoadingOrLoaded(j.Server.Name, j.Replica.GetReplicaIdx())
	return iIsLoading && !jIsLoading
}
