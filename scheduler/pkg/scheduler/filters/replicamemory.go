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

package filters

import (
	"fmt"
	"math"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store"
)

type AvailableMemoryReplicaFilter struct{}

func (r AvailableMemoryReplicaFilter) Name() string {
	return "AvailableMemoryReplicaFilter"
}

func isModelReplicaLoadedOnServerReplica(model *store.ModelVersion, replica *store.ServerReplica) bool {
	if model.HasServer() {
		return model.Server() == replica.GetServerName() && model.GetModelReplicaState(replica.GetReplicaIdx()).AlreadyLoadingOrLoaded()
	}
	return false
}

func (r AvailableMemoryReplicaFilter) Filter(model *store.ModelVersion, replica *store.ServerReplica) bool {
	mem := math.Max(0, float64(replica.GetAvailableMemory())-float64(replica.GetReservedMemory()))
	return model.GetRequiredMemory() <= uint64(mem) || isModelReplicaLoadedOnServerReplica(model, replica)
}

func (r AvailableMemoryReplicaFilter) Description(model *store.ModelVersion, replica *store.ServerReplica) string {
	return fmt.Sprintf("model memory %d replica memory %d", model.GetRequiredMemory(), replica.GetAvailableMemory())
}
