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

	"github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler/db"
)

type ReplicaDrainingFilter struct{}

func (r ReplicaDrainingFilter) Name() string {
	return "ReplicaDrainingFilter"
}

func (r ReplicaDrainingFilter) Filter(model *db.ModelVersion, replica *db.ServerReplica) bool {
	return !replica.GetIsDraining()
}

func (r ReplicaDrainingFilter) Description(model *db.ModelVersion, replica *db.ServerReplica) string {
	return fmt.Sprintf(
		"Replica server %d is draining check %t",
		replica.GetReplicaIdx(), replica.GetIsDraining())
}
