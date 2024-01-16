/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
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
