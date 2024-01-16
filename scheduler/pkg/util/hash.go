/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package util

import (
	"fmt"
	"hash/fnv"
	"strings"

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

func GetKafkaConsumerName(namespace, consumerGroupIdPrefix, modelName, prefix string, maxConsumers int) string {
	idx := modelIdToConsumerBucket(modelName, maxConsumers)
	var sb strings.Builder
	if consumerGroupIdPrefix != "" {
		sb.WriteString(consumerGroupIdPrefix + "-")
	}
	if namespace != "" {
		sb.WriteString(namespace + "-")
	}
	sb.WriteString(prefix + "-")
	sb.WriteString(fmt.Sprintf("%d", idx))
	return sb.String()
}
