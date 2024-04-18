/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package modelscaling

import (
	"sync/atomic"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/interfaces"
)

const (
	ModelLagKey = "lag"
)

type lagStats struct {
	lag uint32
}

func newLagStats() interfaces.ModelStats {
	return &lagStats{
		lag: 0,
	}
}

func (stats *lagStats) Enter(requestId string) error {
	atomic.AddUint32(&stats.lag, 1)
	return nil
}

func (stats *lagStats) Exit(requestId string) error {
	// we want to decrement by 1, however we do not want to go below zero,
	// therfore we cannot use `AddUint32(&x, ^uint32(0))`.
	// the for loop is essentially to make sure that no concurrent requests have
	// changed the value of the `lag` while we are decrementing it.
	for {
		old := stats.Get()
		if old > 0 {
			new := old - 1
			swapped := atomic.CompareAndSwapUint32(&stats.lag, old, new)
			if swapped {
				break
			}
		} else {
			break
		}
	}
	return nil
}

func (stats *lagStats) Get() uint32 {
	return atomic.LoadUint32(&stats.lag)
}

func (stats *lagStats) Reset() error {
	atomic.StoreUint32(&stats.lag, 0)
	return nil
}

func NewModelReplicaLagsKeeper() *modelStatsKeeper {
	return NewModelStatsKeeper(ModelLagKey, newLagStats)
}
