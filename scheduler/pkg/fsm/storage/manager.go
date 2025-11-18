package storage

import (
	"context"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/fsm/state_machine"
)

type ClusterManager interface {
	GetClusterState(ctx context.Context) (state_machine.ClusterState, error)
	SaveClusterState(ctx context.Context, state_machine state_machine.ClusterState) error
}
