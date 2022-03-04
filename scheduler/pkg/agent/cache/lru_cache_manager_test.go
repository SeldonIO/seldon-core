package cache

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"

	. "github.com/onsi/gomega"
)

type fn func(jobs <-chan int, wg *sync.WaitGroup, cache *LRUCacheManager)

func doEvict(cache *LRUCacheManager) (string, int64, error) {
	return cache.Evict()
}

func adderDefault(jobs <-chan int, wg *sync.WaitGroup, cache *LRUCacheManager) {
	for i := range jobs {
		modelId := fmt.Sprintf("model_%d", i)
		_ = cache.AddDefault(modelId)
		_ = cache.UpdateDefault(modelId)
		wg.Done()
	}
}

func adder(jobs <-chan int, wg *sync.WaitGroup, cache *LRUCacheManager) {
	for i := range jobs {
		modelId := fmt.Sprintf("model_%d", i)
		_ = cache.Add(modelId, int64(i))
		wg.Done()
	}
}

func updater(jobs <-chan int, wg *sync.WaitGroup, cache *LRUCacheManager) {
	for i := range jobs {
		modelId := fmt.Sprintf("model_%d", i)
		_ = cache.Update(modelId, int64(i))
		wg.Done()
	}
}

func deleter(jobs <-chan int, wg *sync.WaitGroup, cache *LRUCacheManager) {
	for i := range jobs {
		modelId := fmt.Sprintf("model_%d", i)
		_ = cache.AddDefault(modelId)
		_ = cache.Delete(modelId)
		wg.Done()
	}
}

func evicter(jobs <-chan int, wg *sync.WaitGroup, cache *LRUCacheManager) {
	for range jobs {
		_, _, _ = doEvict(cache)
		wg.Done()
	}
}

func randomer(jobs <-chan int, wg *sync.WaitGroup, cache *LRUCacheManager) {
	for i := range jobs {
		modelId := fmt.Sprintf("model_%d", i)
		switch rand.Intn(6) {
		case 0:
			_, _, _ = doEvict(cache)
		case 1:
			_ = cache.AddDefault(modelId)
		case 2:
			_ = cache.UpdateDefault(modelId)
		case 3:
			_ = cache.Add(modelId, int64(i))
		case 4:
			_ = cache.Update(modelId, int64(i))
		case 5:
			_ = cache.Delete(modelId)
		}

		wg.Done()
	}
}

func createJobs(f fn, numJobs int, numWorkers int, lruCache *LRUCacheManager) {
	var wg sync.WaitGroup
	jobs := make(chan int, numJobs)
	defer close(jobs)

	for i := 0; i < numWorkers; i++ {
		go f(jobs, &wg, lruCache)
	}
	for i := 0; i < numJobs; i++ {
		wg.Add(1)
		jobs <- i
	}
	wg.Wait()
}

func checkEvictOrder(numJobs int, lruCache *LRUCacheManager, g *WithT) {
	counter := numJobs - 1
	for {
		id, value, err := doEvict(lruCache)
		if err != nil {
			break
		}
		g.Expect(id).To(Equal(fmt.Sprintf("model_%d", counter)))
		g.Expect(value).To(Equal(int64(counter)))
		counter--
	}

}

func TestLRUCacheConcurrentLoadwithUpdate(t *testing.T) {
	g := NewGomegaWithT(t)

	numJobs := 20000
	numWorkers := 1000

	//update
	//1. create models for prepoluation
	models := map[string]int64{}
	for i := 0; i < numJobs; i++ {
		models[fmt.Sprintf("model_%d", i)] = 0
	}

	//2. do actual load
	lruCache := MakeLRU(models)
	createJobs(updater, numJobs, numWorkers, lruCache)
	ids, priorities := lruCache.GetItems()
	g.Expect(len(ids)).To(Equal(numJobs))
	g.Expect(len(priorities)).To(Equal(numJobs))

	// check we get evicts in descending count (0 based priority)
	checkEvictOrder(numJobs, lruCache, g)

}

func TestLRUCacheConcurrentLoadwithAdd(t *testing.T) {
	g := NewGomegaWithT(t)

	numJobs := 20000
	numWorkers := 100

	//add
	lruCache := MakeLRU(map[string]int64{})
	createJobs(adder, numJobs, numWorkers, lruCache)
	ids, priorities := lruCache.GetItems()
	g.Expect(len(ids)).To(Equal(numJobs))
	g.Expect(len(priorities)).To(Equal(numJobs))

	// check we get evicts in descending count (0 based priority)
	checkEvictOrder(numJobs, lruCache, g)
}

func TestLRUCacheConcurrent(t *testing.T) {
	//TODO break this down in proper tests
	g := NewGomegaWithT(t)

	numJobs := 20000
	numWorkers := 100

	t.Logf("Start!")

	type test struct {
		name     string
		f        fn
		expected int
	}
	tests := []test{
		{
			name:     "add",
			f:        adder,
			expected: numJobs,
		},
		{
			name:     "addDefault",
			f:        adderDefault,
			expected: numJobs,
		},
		{
			name:     "Evict",
			f:        evicter,
			expected: 0,
		},
		{
			name:     "Delete",
			f:        deleter,
			expected: 0,
		},
		{
			name:     "Random",
			f:        randomer,
			expected: -1, // just to signal not to check value
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			lruCache := MakeLRU(map[string]int64{})
			createJobs(test.f, numJobs, numWorkers, lruCache)
			ids, priorities := lruCache.GetItems()
			if test.expected != -1 {
				g.Expect(len(ids)).To(Equal(test.expected))
				g.Expect(len(priorities)).To(Equal(test.expected))
			}
		})
	}

	t.Logf("Done!")
}

func TestPQCacheSmoke(t *testing.T) {
	//TODO break this down in proper tests
	g := NewGomegaWithT(t)

	models := map[string]int64{
		"model_1": 1, "model_2": 2, "model_3": 3,
	}

	lruCache := MakeLRU(models)

	//get items
	ids, priorities := lruCache.GetItems()
	g.Expect(len(ids)).To(Equal(lruCache.pq.Len()))
	g.Expect(len(priorities)).To(Equal(lruCache.pq.Len()))

	//add new item
	_ = lruCache.Add("model_6", 6)
	ids, priorities = lruCache.GetItems()
	g.Expect(len(ids)).To(Equal(4))
	g.Expect(len(priorities)).To(Equal(4))

	//exists
	exists := lruCache.Exists("model_6")
	g.Expect(exists).To(BeTrue())

	//delete
	err := lruCache.Delete("model_6")
	g.Expect(err).To(BeNil())
	exists = lruCache.Exists("model_6")
	g.Expect(exists).To(BeFalse())

	//not exists
	exists = lruCache.Exists("model_dummy")
	g.Expect(exists).To(BeFalse())

	//get
	priority, _ := lruCache.Get("model_1")
	g.Expect(priority).To(Equal(int64(1)))

	//update item priority
	_ = lruCache.Update("model_1", 7)
	priority, _ = lruCache.Get("model_1")
	g.Expect(priority).To(Equal(int64(7)))

	// evict
	deleted, deletedValue, _ := doEvict(lruCache)
	g.Expect(deleted).To(Equal("model_1"))
	g.Expect(deletedValue).To(Equal(int64(7)))

	//error with get
	_, err = lruCache.Get("model_1")
	g.Expect(err).NotTo(BeNil())

	t.Logf("Done!")
}

func TestLRUCacheSmoke(t *testing.T) {
	//TODO break this down in proper tests
	g := NewGomegaWithT(t)

	//add new/update item with default priority
	lruCache := MakeLRU(map[string]int64{})

	_ = lruCache.AddDefault("model_1")
	_ = lruCache.AddDefault("model_2")
	_ = lruCache.UpdateDefault("model_1")
	deleted, _, _ := doEvict(lruCache)
	g.Expect(deleted).To(Equal("model_2"))
	_ = lruCache.AddDefault("model_3")
	deleted, _, _ = doEvict(lruCache)
	g.Expect(deleted).To(Equal("model_1"))
	deleted, _, _ = doEvict(lruCache)
	g.Expect(deleted).To(Equal("model_3"))

	t.Logf("Done!")
}

func TestLRUCacheSmokeEdgeCases(t *testing.T) {
	g := NewGomegaWithT(t)

	//add new/update item with default priority
	lruCache := MakeLRU(map[string]int64{})

	_, _, err := doEvict(lruCache)
	g.Expect(err).ToNot(BeNil())

	err = lruCache.Delete("model_1")
	g.Expect(err).ToNot(BeNil()) // model_1 does not exist

	err = lruCache.UpdateDefault("model_1")
	g.Expect(err).ToNot(BeNil())

	err = lruCache.Delete("model_2")
	g.Expect(err).ToNot(BeNil()) // model_2 does not exist

	err = lruCache.AddDefault("model_1")
	g.Expect(err).To(BeNil()) // no error
	err = lruCache.AddDefault("model_1")
	g.Expect(err).ToNot(BeNil()) // error, item exists
	err = lruCache.Delete("model_1")
	g.Expect(err).To(BeNil()) // model_1 exist

	err = lruCache.AddDefault("")
	g.Expect(err).ToNot(BeNil()) // error, empty id

	t.Logf("Done!")
}
