package storage

import (
	"context"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/fsm/state_machine"
)

// todo: we could do our tradicional locking too but for this to probably work we would need a manager to do the updates
// todo: the manager would then lock the entire Handle of an event and do its processing, could become a little bit of a problem when having many status updates from Agents
// KVStore interface for state persistence with optimistic locking
type KVStore interface {
	// Get retrieves value and current version
	Get(ctx context.Context, key string) (value []byte, version int64, err error)
	// Set without version check (for initial writes)
	Set(ctx context.Context, key string, value []byte) (int64, error)
	Delete(ctx context.Context, key string) error
}

type ClusterManager interface {
	GetClusterState(ctx context.Context) (state_machine.ClusterState, error)
	SaveClusterState(ctx context.Context, state_machine state_machine.ClusterState) error
}
