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

	"github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler/db"
)

type AvailableMemoryReplicaFilter struct{}

func (r AvailableMemoryReplicaFilter) Name() string {
	return "AvailableMemoryReplicaFilter"
}

func isModelReplicaLoadedOnServerReplica(model *db.ModelVersion, replica *db.ServerReplica) bool {
	if model.HasServer() {
		return model.Server == replica.GetServerName() && model.GetModelReplicaState(int(replica.GetReplicaIdx())).AlreadyLoadingOrLoaded()
	}
	return false
}

func (r AvailableMemoryReplicaFilter) Filter(model *db.ModelVersion, replica *db.ServerReplica) bool {
	mem := math.Max(0, float64(replica.GetAvailableMemory())-float64(replica.GetReservedMemory()))
	return model.GetRequiredMemory() <= uint64(mem) || isModelReplicaLoadedOnServerReplica(model, replica)
}

func (r AvailableMemoryReplicaFilter) Description(model *db.ModelVersion, replica *db.ServerReplica) string {
	return fmt.Sprintf("model memory %d replica memory %d replica reserved memory %d",
		model.GetRequiredMemory(), replica.GetAvailableMemory(), replica.GetReservedMemory())
}
