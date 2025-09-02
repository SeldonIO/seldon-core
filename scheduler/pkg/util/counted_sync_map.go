/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package util

import (
	"sync"
	"sync/atomic"
)

func NewCountedSyncMap[T any]() *CountedSyncMap[T] {
	return &CountedSyncMap[T]{
		m: &sync.Map{},
	}
}

type CountedSyncMap[T any] struct {
	m     *sync.Map
	count int32
}

func (c *CountedSyncMap[T]) Store(key string, value T) {
	_, exists := c.m.Swap(key, value)
	if !exists {
		atomic.AddInt32(&c.count, 1)
	}
}

func (c *CountedSyncMap[T]) Load(key string) (*T, bool) {
	val, ok := c.m.Load(key)
	if !ok {
		return nil, false
	}
	v := val.(T)
	return &v, true
}

func (c *CountedSyncMap[T]) Delete(key string) {
	if _, exists := c.m.LoadAndDelete(key); exists {
		atomic.AddInt32(&c.count, -1)
	}
}

func (c *CountedSyncMap[T]) Length() int {
	return int(atomic.LoadInt32(&c.count))
}

func (c *CountedSyncMap[T]) Range(f func(key string, value T) bool) {
	c.m.Range(func(k, v any) bool {
		return f(k.(string), v.(T))
	})
}
