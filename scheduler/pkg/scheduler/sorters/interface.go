/*
Copyright 2022 Seldon Technologies Ltd.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package sorters

import "github.com/seldonio/seldon-core/scheduler/pkg/store"

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
