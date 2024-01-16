/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package sorters

// Rationale: Lower indexed replicas will exist for longer as they will be scaled down later than a higher index
type ReplicaIndexSorter struct{}

func (m ReplicaIndexSorter) Name() string {
	return "ReplicaIndexSorter"
}

func (r ReplicaIndexSorter) IsLess(i *CandidateReplica, j *CandidateReplica) bool {
	return i.Replica.GetReplicaIdx() < j.Replica.GetReplicaIdx()
}
