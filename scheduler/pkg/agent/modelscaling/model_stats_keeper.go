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

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/interfaces"
)

type modelStatsKeeper struct {
	key         string
	stats       map[string]interfaces.ModelStats
	newStatsFunc func() interfaces.ModelStats
	mu          sync.RWMutex
}

func NewModelStatsKeeper(key string, newStatFunc func() interfaces.ModelStats) *modelStatsKeeper {
	return &modelStatsKeeper{
		key:         key,
		stats:       make(map[string]interfaces.ModelStats),
		newStatsFunc: newStatFunc,
		mu:          sync.RWMutex{},
	}
}

func (keeper *modelStatsKeeper) ModelInferEnter(modelName, requestId string) error {
	return keeper.applyWithModel(modelName, func(stat interfaces.ModelStats) error {
		return stat.Enter(requestId)
	})
}

func (keeper *modelStatsKeeper) ModelInferExit(modelName, requestId string) error {
	return keeper.applyWithModel(modelName, func(stat interfaces.ModelStats) error {
		return stat.Exit(requestId)
	})
}

func (keeper *modelStatsKeeper) Add(modelName string) error {
	return keeper.applyWithModel(modelName, func(stat interfaces.ModelStats) error {
		return nil
	})
}
func (keeper *modelStatsKeeper) Delete(modelName string) error {
	keeper.mu.Lock()
	delete(keeper.stats, modelName)
	keeper.mu.Unlock()
	return nil
}

func (keeper *modelStatsKeeper) Get(modelName string) (uint32, error) {
	keeper.mu.RLock()
	defer keeper.mu.RUnlock()
	stat, ok := keeper.stats[modelName]
	if !ok {
		return 0, fmt.Errorf("model replica %s is not found", modelName)
	}
	return stat.Get(), nil
}

func (keeper *modelStatsKeeper) GetAll(threshold uint32, op interfaces.LogicOperation, reset bool) ([]*interfaces.ModelStatsKV, error) {
	if op == interfaces.Gte {
		return keeper.getAllGte(threshold, reset)
	} else {
		return nil, fmt.Errorf("operation not supported %d", op)
	}
}

func (keeper *modelStatsKeeper) applyWithModel(modelName string, fn func(interfaces.ModelStats) error) error {
	keeper.mu.RLock()
	stat, ok := keeper.stats[modelName]
	if !ok {
		keeper.mu.RUnlock()

		keeper.mu.Lock()
		keeper.stats[modelName] = keeper.newStatsFunc()
		keeper.mu.Unlock()

		//try again after creating the item in map
		keeper.mu.RLock()
		stat = keeper.stats[modelName]
	}

	defer keeper.mu.RUnlock()
	return fn(stat)
}

func (keeper *modelStatsKeeper) getAllGte(threshold uint32, reset bool) ([]*interfaces.ModelStatsKV, error) {
	keeper.mu.RLock()
	defer keeper.mu.RUnlock()

	rets := []*interfaces.ModelStatsKV{}

	for k, v := range keeper.stats {
		statValue := uint32(v.Get())
		if statValue >= threshold {
			rets = append(rets, &interfaces.ModelStatsKV{
				ModelName: k,
				Value:     statValue,
				Key:       keeper.key,
			})
			if reset {
				v.Reset()
			}
		}
	}
	return rets, nil
}
