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
	"testing"

	. "github.com/onsi/gomega"
)

func TestConsistentHash(t *testing.T) {
	g := NewGomegaWithT(t)

	numBuckets := 10

	type test struct {
		name    string
		modelId string
		hash    uint32
	}
	tests := []test{
		{
			name:    "smoke test",
			modelId: "dumm1",
			hash:    modelIdToConsumerBucket("dumm1", numBuckets),
		},
		{
			name:    "smoke empty test",
			modelId: "",
			hash:    modelIdToConsumerBucket("", numBuckets),
		},
		{
			name:    "smoke escape chars test",
			modelId: "x%$\\",
			hash:    modelIdToConsumerBucket("x%$\\", numBuckets),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g.Expect(modelIdToConsumerBucket(test.modelId, numBuckets)).To(Equal(test.hash))

		})
	}
}

func TestGetKafkaConsumerName(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name                  string
		namespace             string
		consumerGroupIdPrefix string
		modelName             string
		prefix                string
		maxConsumers          int
		expected              string
	}
	tests := []test{
		{
			name:                  "all params",
			consumerGroupIdPrefix: "foo",
			modelName:             "model",
			prefix:                "p",
			maxConsumers:          100,
			expected:              fmt.Sprintf("%s-%s-%d", "foo", "p", modelIdToConsumerBucket("model", 100)),
		},
		{
			name:                  "no consumer prefix",
			consumerGroupIdPrefix: "",
			modelName:             "model",
			prefix:                "p",
			maxConsumers:          100,
			expected:              fmt.Sprintf("%s-%d", "p", modelIdToConsumerBucket("model", 100)),
		},
		{
			name:                  "all params",
			namespace:             "default",
			consumerGroupIdPrefix: "foo",
			modelName:             "model",
			prefix:                "p",
			maxConsumers:          100,
			expected:              fmt.Sprintf("%s-%s-%s-%d", "foo", "default", "p", modelIdToConsumerBucket("model", 100)),
		},
		{
			name:                  "no consumer prefix",
			namespace:             "default",
			consumerGroupIdPrefix: "",
			modelName:             "model",
			prefix:                "p",
			maxConsumers:          100,
			expected:              fmt.Sprintf("%s-%s-%d", "default", "p", modelIdToConsumerBucket("model", 100)),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g.Expect(GetKafkaConsumerName(test.namespace, test.consumerGroupIdPrefix, test.modelName, test.prefix, test.maxConsumers)).To(Equal(test.expected))

		})
	}
}
