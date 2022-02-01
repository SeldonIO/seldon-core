package cache

import (
	"container/heap"
	"fmt"
	"sync"
	"time"
)

type LRUCacheManager struct {
	pq PriorityQueue
	mu sync.RWMutex
}

func (cache *LRUCacheManager) Evict() (string, int64, error) {
	cache.mu.Lock()
	defer cache.mu.Unlock()
	if cache.pq.Len() > 0 {
		item := heap.Pop(&(cache.pq)).(*Item)
		return item.id, item.priority, nil
	}
	return "", 0, fmt.Errorf("empty cache, cannot evict")
}

func (cache *LRUCacheManager) Add(id string, value int64) error {
	if id == "" {
		return fmt.Errorf("cannot use empty string")
	}
	if cache.Exists(id) {
		return fmt.Errorf("item already exists in cache %s", id)
	}
	cache.mu.Lock()
	defer cache.mu.Unlock()
	item := &Item{
		id:       id,
		priority: value,
	}
	heap.Push(&(cache.pq), item)
	return nil
}

func (cache *LRUCacheManager) AddDefault(id string) error {
	return cache.Add(id, -ts())
}

func (cache *LRUCacheManager) Update(id string, value int64) error {
	// find the item
	// TODO: make it efficient
	// do we really need write lock?
	cache.mu.Lock()
	defer cache.mu.Unlock()
	for _, item := range cache.pq {
		if item.id == id {
			cache.pq.update(item, item.id, value)
			return nil
		}
	}
	return fmt.Errorf("could not find item %s", id)
}

func (cache *LRUCacheManager) UpdateDefault(id string) error {
	return cache.Update(id, -ts())
}

func (cache *LRUCacheManager) Exists(id string) bool {
	// TODO: make it efficient?
	cache.mu.RLock()
	defer cache.mu.RUnlock()
	for _, item := range cache.pq {
		if item.id == id {
			return true
		}
	}
	return false
}

func (cache *LRUCacheManager) Get(id string) (int64, error) {
	// TODO: make it efficient?
	cache.mu.RLock()
	defer cache.mu.RUnlock()
	for _, item := range cache.pq {
		if item.id == id {
			return item.priority, nil
		}
	}
	return -1, fmt.Errorf("could not find item %s", id)
}

func (cache *LRUCacheManager) Delete(id string) error {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	for _, item := range cache.pq {
		if item.id == id {
			heap.Remove(&(cache.pq), item.index)
			return nil
		}
	}

	return fmt.Errorf("could not find item %s", id)
}

func (cache *LRUCacheManager) GetItems() ([]string, []int64) {
	// TODO: make it efficient?
	// this is not in priority order
	cache.mu.RLock()
	defer cache.mu.RUnlock()
	ids := make([]string, cache.pq.Len())
	priorities := make([]int64, cache.pq.Len())
	for i, item := range cache.pq {
		ids[i] = item.id
		priorities[i] = item.priority
	}
	return ids, priorities
}

func MakeLRU(initItems map[string]int64) *LRUCacheManager {
	pq := make(PriorityQueue, len(initItems))
	i := 0
	for id, priority := range initItems {
		pq[i] = &Item{
			id:       id,
			priority: priority,
			index:    i,
		}
		i++
	}
	heap.Init(&pq)
	return &LRUCacheManager{
		pq: pq,
		mu: sync.RWMutex{},
	}
}

func ts() int64 {
	now := time.Now().UnixNano()
	return now
}
