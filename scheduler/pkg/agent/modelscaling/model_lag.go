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
	"sync"
	"sync/atomic"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/interfaces"
)

type operation int

const (
	inc operation = iota
	dec
	reset
	set
)

const (
	ModelLagKey = "lag"
)

type lagKeeper struct {
	lag uint32
}

func newLagKeeper() *lagKeeper {
	return &lagKeeper{
		lag: 0,
	}
}

func (lagKeeper *lagKeeper) inc() {
	atomic.AddUint32(&lagKeeper.lag, 1)
}

func (lagKeeper *lagKeeper) dec() {
	// we want to decrement by 1, however we do not want to go below zero,
	// therfore we cannot use `AddUint32(&x, ^uint32(0))`.
	// the for loop is essentially to make sure that no concurrent requests have
	// changed the value of the `lag` while we are decrementing it.
	for {
		old := lagKeeper.get()
		if old > 0 {
			new := old - 1
			swapped := atomic.CompareAndSwapUint32(&lagKeeper.lag, old, new)
			if swapped {
				break
			}
		} else {
			break
		}
	}
}

func (lagKeeper *lagKeeper) reset() {
	lagKeeper.set(uint32(0))
}

func (lagKeeper *lagKeeper) get() uint32 {
	return atomic.LoadUint32(&lagKeeper.lag)
}

func (lagKeeper *lagKeeper) set(v uint32) {
	atomic.StoreUint32(&lagKeeper.lag, v)
}

type ModelReplicaLagsKeeper struct {
	lags map[string]*lagKeeper
	mu   sync.RWMutex
}

func NewModelReplicaLagsKeeper() *ModelReplicaLagsKeeper {
	return &ModelReplicaLagsKeeper{
		lags: make(map[string]*lagKeeper),
		mu:   sync.RWMutex{},
	}
}

func (lagsKeeper *ModelReplicaLagsKeeper) IncDefault(modelName string) error {
	// note: value (0) not used
	return lagsKeeper.apply(modelName, inc, 0)
}

func (lagsKeeper *ModelReplicaLagsKeeper) DecDefault(modelName string) error {
	// note: value (0) not used
	return lagsKeeper.apply(modelName, dec, 0)
}

func (lagsKeeper *ModelReplicaLagsKeeper) Reset(modelName string) error {
	// note: value (0) not used
	return lagsKeeper.apply(modelName, reset, 0)
}

func (lagsKeeper *ModelReplicaLagsKeeper) Set(modelName string, value uint32) error {
	return lagsKeeper.apply(modelName, set, value)
}

func (lagsKeeper *ModelReplicaLagsKeeper) Inc(modelName string, _ uint32) error {
	return lagsKeeper.IncDefault(modelName)
}

func (lagsKeeper *ModelReplicaLagsKeeper) Dec(modelName string, _ uint32) error {
	return lagsKeeper.DecDefault(modelName)
}

func (lagsKeeper *ModelReplicaLagsKeeper) Info() string {
	return "Model Replica Lag: incoming - outgoing requests"
}

func (lagsKeeper *ModelReplicaLagsKeeper) Delete(modelName string) error {
	lagsKeeper.mu.Lock()
	delete(lagsKeeper.lags, modelName)
	lagsKeeper.mu.Unlock()
	return nil
}

func (luKeeper *ModelReplicaLagsKeeper) Add(modelName string) error {
	return luKeeper.IncDefault(modelName)
}

func (lagsKeeper *ModelReplicaLagsKeeper) Get(modelName string) (uint32, error) {
	lagsKeeper.mu.RLock()
	defer lagsKeeper.mu.RUnlock()
	lagKeeper, ok := lagsKeeper.lags[modelName]
	if !ok {
		return 0, fmt.Errorf("Model replica %s is not found", modelName)
	}
	return lagKeeper.get(), nil
}

func (lagsKeeper *ModelReplicaLagsKeeper) GetAll(threshold uint32, op interfaces.LogicOperation, reset bool) ([]*interfaces.ModelStatsKV, error) {
	if op == interfaces.Gte {
		return lagsKeeper.getAllGte(threshold, reset)
	} else {
		return nil, fmt.Errorf("Operation not supported %d", op)
	}
}

func (lagsKeeper *ModelReplicaLagsKeeper) getAllGte(threshold uint32, reset bool) ([]*interfaces.ModelStatsKV, error) {
	lagsKeeper.mu.RLock()
	defer lagsKeeper.mu.RUnlock()

	rets := []*interfaces.ModelStatsKV{}

	for k, v := range lagsKeeper.lags {
		lag := v.get()
		if lag >= threshold {
			rets = append(rets, &interfaces.ModelStatsKV{
				ModelName: k,
				Value:     lag,
				Key:       ModelLagKey,
			})
			if reset {
				v.reset()
			}
		}
	}
	return rets, nil
}

func (lagsKeeper *ModelReplicaLagsKeeper) new(modelName string) {
	lagsKeeper.mu.Lock()
	lagsKeeper.lags[modelName] = newLagKeeper()
	lagsKeeper.mu.Unlock()
}

func (lagsKeeper *ModelReplicaLagsKeeper) apply(modelName string, op operation, value uint32) error {
	lagsKeeper.mu.RLock()
	lagKeeper, ok := lagsKeeper.lags[modelName]
	if !ok {
		lagsKeeper.mu.RUnlock()
		lagsKeeper.new(modelName)
		//try again after creating the item in map
		lagsKeeper.mu.RLock()
		lagKeeper = lagsKeeper.lags[modelName]
	}

	defer lagsKeeper.mu.RUnlock()

	switch op {
	case inc:
		lagKeeper.inc()
	case dec:
		lagKeeper.dec()
	case reset:
		lagKeeper.reset()
	case set:
		lagKeeper.set(value)
	default:
		return fmt.Errorf("operation not supported")
	}

	return nil
}
