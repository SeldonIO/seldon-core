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
	iIsLoading := i.Model.IsLoadingOrLoaded(i.Server.Name, i.Replica.GetReplicaIdx())
	jIsLoading := j.Model.IsLoadingOrLoaded(j.Server.Name, j.Replica.GetReplicaIdx())
	return iIsLoading && !jIsLoading
}
