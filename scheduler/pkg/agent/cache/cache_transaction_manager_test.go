package cache

import (
	"fmt"
	"testing"

	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
)

func setupTxManager() *CacheTransactionManager {
	logger := log.New()
	logger.SetLevel(log.DebugLevel)

	return NewLRUCacheTransactionManager(logger)
}

func TestLRUCacheReloadTransaction(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name string
	}
	tests := []test{
		{name: "evicted_ok"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			id := "dummy_1"

			txManager := setupTxManager()

			result := make(chan bool, 1)

			endReloadFn, exists := txManager.StartReloadIfNotExists(id)
			g.Expect(exists).To(Equal(false))
			go func(r chan<- bool) {
				// this should be waiting
				_, result := txManager.StartReloadIfNotExists(id)
				r <- result
			}(result)
			_ = txManager.AddDefault(id)
			endReloadFn()

			actualResult := <-result
			g.Expect(actualResult).To(Equal(true))

		})
	}
}

func TestLRUCacheEvictTransaction(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name string
	}
	tests := []test{
		{name: "evicted_ok"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			id := "dummy_1"

			txManager := setupTxManager()
			_ = txManager.AddDefault(id)

			result := make(chan bool, 1)

			itemFromCache, _, _ := txManager.Peek()
			endEvictFn, _ := txManager.StartEvict(itemFromCache) // only one model really
			go func(r chan<- bool) {
				r <- txManager.Exists(itemFromCache, true)
			}(result)

			endEvictFn()

			actualResult := <-result
			g.Expect(actualResult).To(Equal(false))

		})
	}
}

func TestLRUCacheEvictTransactionEdgeCases(t *testing.T) {
	g := NewGomegaWithT(t)

	id := "model_1"

	txManager := setupTxManager()

	_ = txManager.AddDefault(id)

	endEvictFn, err := txManager.StartEvict(id)
	g.Expect(err).To(BeNil()) // no error here
	endEvictFn()

	endEvictFn, err = txManager.StartEvict(id)
	g.Expect(err).To(Equal(fmt.Errorf("could not find item model_1")))
	endEvictFn()

	g.Expect(txManager.Exists(id, true)).To(Equal(false))
}
