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

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store"
)

type ServerReplicaFilter struct{}

func (r ServerReplicaFilter) Name() string {
	return "ServerReplicaFilter"
}

func (r ServerReplicaFilter) Filter(model *store.ModelVersion, server *store.ServerSnapshot) bool {
	return len(server.Replicas) > 0
}

func (r ServerReplicaFilter) Description(model *store.ModelVersion, server *store.ServerSnapshot) string {
	return fmt.Sprintf("%d server replicas (waiting for server replicas to connect)", len(server.Replicas))
}
