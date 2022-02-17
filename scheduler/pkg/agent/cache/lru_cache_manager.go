package cache

import (
	"container/heap"
	"fmt"
	"sync"
	"time"
)

type LRUCacheManager struct {
	pq        PriorityQueue
	mu        sync.RWMutex
	itemLocks sync.Map
}

func (cache *LRUCacheManager) itemLock(id string) error {
	var lock sync.RWMutex
	existingLock, loaded := cache.itemLocks.LoadOrStore(id, &lock)
	if loaded {
		return fmt.Errorf("Model is already dirty %s", id)
	}
	existingLock.(*sync.RWMutex).Lock()
	return nil
}

func (cache *LRUCacheManager) itemUnLock(id string) {
	existingLock, loaded := cache.itemLocks.LoadAndDelete(id)
	if loaded {
		existingLock.(*sync.RWMutex).Unlock()
	}
}

func (cache *LRUCacheManager) itemWait(id string) {
	existingLock, loaded := cache.itemLocks.Load(id)
	if loaded {
		existingLock.(*sync.RWMutex).RLock()
		defer existingLock.(*sync.RWMutex).RUnlock()
	}
}

func (cache *LRUCacheManager) StartEvict() (string, int64, error) {
	cache.mu.Lock()
	defer cache.mu.Unlock()
	if cache.pq.Len() > 0 {
		item := heap.Pop(&(cache.pq)).(*Item)
		if err := cache.itemLock(item.id); err != nil {
			// re-add
			cache.add(item.id, item.priority)
			cache.itemUnLock(item.id)
			return "", 0, fmt.Errorf("cannot evict")
		}
		return item.id, item.priority, nil
	}
	return "", 0, fmt.Errorf("empty cache, cannot evict")
}

func (cache *LRUCacheManager) EndEvict(id string, value int64, rollback bool) error {
	_, loaded := cache.itemLocks.Load(id)
	if !loaded {
		// item is not dirty, abort
		return fmt.Errorf("id %s is not dirty", id)
	}
	defer cache.itemUnLock(id)
	if rollback {
		// no locking here
		cache.add(id, value)
	}
	return nil
}

func (cache *LRUCacheManager) add(id string, value int64) {
	item := &Item{
		id:       id,
		priority: value,
	}
	heap.Push(&(cache.pq), item)
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
	cache.add(id, value)
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
	cache.itemWait(id)
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
		pq:        pq,
		mu:        sync.RWMutex{},
		itemLocks: sync.Map{},
	}
}

func ts() int64 {
	now := time.Now().UnixNano()
	return now
}
