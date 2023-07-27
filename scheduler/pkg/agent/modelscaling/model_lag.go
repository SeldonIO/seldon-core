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
