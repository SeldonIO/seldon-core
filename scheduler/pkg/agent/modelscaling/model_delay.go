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
	"fmt"
	"sync"
	"time"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/interfaces"
)

const (
	ModelDelayKey = "delay"
)

type delayStats struct {
	delays     []int64
	startTimes map[string]time.Time
	mu         sync.RWMutex
}

func newDelayStats() interfaces.ModelStats {
	return &delayStats{
		delays:     []int64{},
		startTimes: make(map[string]time.Time),
		mu:         sync.RWMutex{},
	}
}

func (stats *delayStats) Enter(requestId string) error {
	stats.mu.Lock()
	defer stats.mu.Unlock()
	stats.startTimes[requestId] = time.Now()

	return nil
}

func (stats *delayStats) Exit(requestId string) error {
	stats.mu.Lock()
	defer stats.mu.Unlock()
	if startTime, ok := stats.startTimes[requestId]; ok {
		delete(stats.startTimes, requestId)
		delay := time.Since(startTime).Milliseconds()
		stats.delays = append(stats.delays, delay)
	} else {
		return fmt.Errorf("failed to find the start time of the request")
	}
	return nil
}

func (stats *delayStats) Reset() error {
	stats.mu.Lock()
	defer stats.mu.Unlock()
	stats.delays = []int64{}
	stats.startTimes = make(map[string]time.Time)
	return nil
}

// Get average delay in milliseconds
func (stats *delayStats) Get() uint32 {
	stats.mu.RLock()
	defer stats.mu.RUnlock()

	if len(stats.delays) == 0 {
		return uint32(0)
	}
	sum := int64(0)
	for _, delay := range stats.delays {
		sum += delay
	}
	average := sum / int64(len(stats.delays))
	return uint32(average)
}

func NewModelReplicaDelaysKeeper() *modelStatsKeeper {
	return NewModelStatsKeeper(ModelDelayKey, newDelayStats)
}
