package cache

import (
	"sync"

	log "github.com/sirupsen/logrus"
)

type CacheTransactionManager struct {
	cache     CacheManager
	itemLocks sync.Map
	logger    log.FieldLogger
}

func (tx *CacheTransactionManager) itemLock(id string) {
	var lock sync.RWMutex
	existingLock, _ := tx.itemLocks.LoadOrStore(id, &lock)
	existingLock.(*sync.RWMutex).Lock()
}

func (tx *CacheTransactionManager) itemRLock(id string) {
	var lock sync.RWMutex
	existingLock, _ := tx.itemLocks.LoadOrStore(id, &lock)
	existingLock.(*sync.RWMutex).RLock()
}

func (tx *CacheTransactionManager) itemRUnLock(id string) {
	existingLock, loaded := tx.itemLocks.Load(id)
	if loaded {
		existingLock.(*sync.RWMutex).RUnlock()
	}
}

func (tx *CacheTransactionManager) itemUnLock(id string) {
	existingLock, loaded := tx.itemLocks.Load(id)
	if loaded {
		existingLock.(*sync.RWMutex).Unlock()
	}
}

func (tx *CacheTransactionManager) itemWait(id string) {
	existingLock, loaded := tx.itemLocks.Load(id)
	if loaded {
		existingLock.(*sync.RWMutex).RLock()
		defer existingLock.(*sync.RWMutex).RUnlock()
	}
}

func (tx *CacheTransactionManager) Lock(id string) {
	tx.itemLock(id)
}

func (tx *CacheTransactionManager) Unlock(id string) {
	tx.itemUnLock(id)
}

func (tx *CacheTransactionManager) RLock(id string) {
	tx.itemRLock(id)
}

func (tx *CacheTransactionManager) RUnlock(id string) {
	tx.itemRUnLock(id)
}

func (tx *CacheTransactionManager) StartEvict(id string) (func(), error) {
	// we also return a function to call at the end of the transaction

	tx.itemLock(id)
	// TODO: this can be made efficient as top of queue could still be id?
	return func() {
		tx.itemUnLock(id)
	}, tx.cache.Delete(id)
}

func (tx *CacheTransactionManager) StartReloadIfNotExists(id string) (func(), bool) {
	// TODO: how can we simplify the logic in this function, perhaps by introducing another lock?
	// we also return a function to call at the end of the transaction

	exists := tx.Exists(id, true)
	if !exists {
		tx.itemLock(id)
		// check again here if item is still not in cache
		exists = tx.Exists(id, false)

		if exists {
			// TODO: what will happen if something happens between unlock and rlock?
			// 404?
			tx.itemUnLock(id)
			tx.itemRLock(id)
		}
	} else {
		// TODO: it is possible because of race conditions that the item is not anymore
		// in the cache. this will probably fail downstream somewhere
		// perhaps it is fine because of have retries anyway
		tx.itemRLock(id)
	}
	if exists {
		return func() {
			tx.itemRUnLock(id)
		}, true
	} else {
		return func() {
			tx.itemUnLock(id)
		}, false
	}

}

func (tx *CacheTransactionManager) Exists(id string, waitOnItem bool) bool {
	if waitOnItem {
		tx.itemWait(id)
	}
	return tx.cache.Exists(id)
}

func (tx *CacheTransactionManager) UpdateDefault(id string) error {
	return tx.cache.UpdateDefault(id)
}

func (tx *CacheTransactionManager) AddDefault(id string) error {
	return tx.cache.AddDefault(id)
}

func (tx *CacheTransactionManager) Delete(id string) error {
	return tx.cache.Delete(id)
}

func (tx *CacheTransactionManager) GetItems() ([]string, []int64) {
	return tx.cache.GetItems()
}

func (tx *CacheTransactionManager) Get(id string) (int64, error) {
	return tx.cache.Get(id)
}

func (tx *CacheTransactionManager) Peek() (string, int64, error) {
	return tx.cache.Peek()
}

func newCacheTransactionManager(cache CacheManager, logger log.FieldLogger) *CacheTransactionManager {
	return &CacheTransactionManager{
		cache:     cache,
		itemLocks: sync.Map{},
		logger:    logger.WithField("Source", "CacheTransactionManager"),
	}
}

func NewLRUCacheTransactionManager(logger log.FieldLogger) *CacheTransactionManager {
	lru := MakeLRU(map[string]int64{})
	return newCacheTransactionManager(lru, logger)
}
