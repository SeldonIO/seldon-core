package sorters

// This sorter favours servers that have the models already loaded on them, this is useful to minimise ping-pong of models between servers
// which can be expensive in terms of model loading time.
type ModelAlreadyLoadedOnServerSorter struct{}

func (m ModelAlreadyLoadedOnServerSorter) Name() string {
	return "ModelAlreadyLoadedOnServerSorter"
}

func (m ModelAlreadyLoadedOnServerSorter) IsLess(i *CandidateServer, j *CandidateServer) bool {
	return i.Model.ServerName == i.Server.Name
}
