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
