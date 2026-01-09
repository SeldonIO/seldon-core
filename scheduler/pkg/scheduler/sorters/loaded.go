/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package sorters

type ModelAlreadyLoadedSorter struct{}

func (m ModelAlreadyLoadedSorter) Name() string {
	return "ModelAlreadyLoadedSorter"
}

func (m ModelAlreadyLoadedSorter) IsLess(i *CandidateReplica, j *CandidateReplica) bool {
	iIsLoading := i.Model.IsLoadingOrLoaded(i.Server.Name, int(i.Replica.GetReplicaIdx()))
	jIsLoading := j.Model.IsLoadingOrLoaded(j.Server.Name, int(j.Replica.GetReplicaIdx()))
	return iIsLoading && !jIsLoading
}

// This sorter favours servers that have the models already loaded on them, this is useful to minimise ping-pong of models between servers
// which can be expensive in terms of model loading time.
type ModelAlreadyLoadedOnServerSorter struct{}

func (m ModelAlreadyLoadedOnServerSorter) Name() string {
	return "ModelAlreadyLoadedOnServerSorter"
}

func (m ModelAlreadyLoadedOnServerSorter) IsLess(i *CandidateServer, j *CandidateServer) bool {
	return i.Model.Server == i.Server.Name
}
