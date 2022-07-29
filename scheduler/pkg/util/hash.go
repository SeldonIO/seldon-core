package util

import (
	"fmt"
	"hash/fnv"

	"github.com/OneOfOne/xxhash"
)

func Hash(s string) (uint32, error) {
	h := fnv.New32a()
	_, err := h.Write([]byte(s))
	if err != nil {
		return 0, err
	}
	return h.Sum32(), nil
}

func XXHash(key string) string {
	h := xxhash.New32()
	return fmt.Sprintf("%x", h.Sum([]byte(key)))
}

// Map a model name / id to a consumer bucket consistently.
// This requires that number of buckets does not change between calls.
// If it changes there is a potential redundant work that is being done as kafka
// will restart from earliest.
func modelIdToConsumerBucket(modelId string, numBuckets int) uint32 {
	hash, err := Hash(modelId)
	if err != nil {
		// is this ok to revert to bucket 0?
		return 0
	}
	return hash % uint32(numBuckets)
}

func GetKafkaConsumerName(modelName, prefix string, maxConsumers int) string {
	idx := modelIdToConsumerBucket(modelName, maxConsumers)
	return fmt.Sprintf("%s-%d", prefix, idx)
}
