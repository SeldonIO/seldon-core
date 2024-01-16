/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package modelscaling

import (
	"fmt"
	"time"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/cache"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/interfaces"
)

const (
	ModelLastUsedKey = "last_used"
)

type ModelReplicaLastUsedKeeper struct {
	// note that this a min-priority queue impl (i.e pop will return lowest value / highest priority)
	// we use negative unix timestamp to pop least recently used item first
	pq *cache.LRUCacheManager
}

func NewModelReplicaLastUsedKeeper() *ModelReplicaLastUsedKeeper {
	return &ModelReplicaLastUsedKeeper{
		pq: cache.MakeLRU(map[string]int64{}),
	}
}

func (luKeeper *ModelReplicaLastUsedKeeper) IncDefault(modelName string) error {
	return luKeeper.Inc(modelName, 0) // 0 not used
}

func (luKeeper *ModelReplicaLastUsedKeeper) Set(modelName string, ts uint32) error {
	if err := luKeeper.pq.Update(modelName, -int64(ts)); err != nil {
		return luKeeper.pq.Add(modelName, -int64(ts))
	} else {
		return err
	}
}

func (luKeeper *ModelReplicaLastUsedKeeper) Inc(modelName string, _ uint32) error {
	ts := -time.Now().Unix() // this is in seconds resolution

	// if we fail update, we assume the model does not exist and we add it.
	// in case of race add might also fail if another add succeeded beforehand
	// this is fine still as by definition these transactions are happening at the "same" time
	if err := luKeeper.pq.Update(modelName, ts); err != nil {
		return luKeeper.pq.Add(modelName, ts)
	} else {
		return err
	}
}

func (luKeeper *ModelReplicaLastUsedKeeper) Info() string {
	return "Model last used"
}

func (luKeeper *ModelReplicaLastUsedKeeper) Delete(modelName string) error {
	return luKeeper.pq.Delete(modelName)
}

func (luKeeper *ModelReplicaLastUsedKeeper) Add(modelName string) error {
	return luKeeper.IncDefault(modelName)
}

func (luKeeper *ModelReplicaLastUsedKeeper) Get(modelName string) (uint32, error) {
	if ts, err := luKeeper.pq.Get(modelName); err != nil {
		return 0, err
	} else {
		return uint32(-ts), nil
	}
}

func (luKeeper *ModelReplicaLastUsedKeeper) GetAll(threshold uint32, op interfaces.LogicOperation, reset bool) ([]*interfaces.ModelStatsKV, error) {
	if op == interfaces.Gte {
		return luKeeper.getAllGte(threshold, reset)
	} else {
		return nil, fmt.Errorf("Operation not supported %d", op)
	}
}

func (luKeeper *ModelReplicaLastUsedKeeper) getAllGte(threshold uint32, reset bool) ([]*interfaces.ModelStatsKV, error) {
	type kv struct {
		k string
		v int64
	}

	rets := []*interfaces.ModelStatsKV{}
	cutOff := time.Now().Unix() - int64(threshold)

	coldModels := []kv{}

	for {
		// TODO: after evict, the model can be used and therefore invalidating that it needs to scale down
		// we assume for now that this is not common in practice.
		model, priority, err := luKeeper.pq.Evict()
		if err != nil {
			break
		}
		if -priority > cutOff {
			_ = luKeeper.pq.Add(model, priority)
			break
		}
		coldModels = append(coldModels, kv{
			k: model,
			v: priority,
		})
	}

	for _, kv := range coldModels {
		rets = append(rets, &interfaces.ModelStatsKV{
			ModelName: kv.k,
			Value:     uint32(-kv.v),
			Key:       ModelLastUsedKey,
		})
		if !reset {
			// this can fail if the model has been used in the meantime, which is fine
			// because it will be added with a newer, more representative ts
			_ = luKeeper.pq.Add(kv.k, kv.v)
		}
	}

	return rets, nil
}

func (luKeeper *ModelReplicaLastUsedKeeper) DecDefault(modelName string) error {
	return fmt.Errorf("Not implemented")
}

func (luKeeper *ModelReplicaLastUsedKeeper) Reset(modelName string) error {
	return fmt.Errorf("Not implemented")
}

func (luKeeper *ModelReplicaLastUsedKeeper) Dec(modelName string, _ uint32) error {
	return fmt.Errorf("Not implemented")
}
