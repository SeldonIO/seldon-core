/*
Copyright 2022 Seldon Technologies Ltd.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
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
