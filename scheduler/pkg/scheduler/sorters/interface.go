/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package sorters

import "github.com/seldonio/seldon-core/scheduler/v2/pkg/store"

type CandidateServer struct {
	Model          *store.ModelVersion
	Server         *store.ServerSnapshot
	ChosenReplicas []*store.ServerReplica
}

type CandidateReplica struct {
	Model   *store.ModelVersion
	Server  *store.ServerSnapshot
	Replica *store.ServerReplica
}

type ServerSorter interface {
	Name() string
	IsLess(i *CandidateServer, j *CandidateServer) bool
}

type ReplicaSorter interface {
	Name() string
	IsLess(i *CandidateReplica, j *CandidateReplica) bool
}
