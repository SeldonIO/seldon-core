/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package storage

import (
	"context"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/fsm/state_machine"
)

type ClusterManager interface {
	GetClusterState(ctx context.Context) (state_machine.ClusterState, error)
	SaveClusterState(ctx context.Context, state_machine state_machine.ClusterState) error
}
