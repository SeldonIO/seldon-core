/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package util

import (
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
			name:                  "all params except namespace",
			consumerGroupIdPrefix: "foo",
			modelName:             "model",
			prefix:                "p",
			maxConsumers:          1,
			expected:              "foo-p-0",
		},
		{
			name:                  "no consumer prefix or namespace",
			consumerGroupIdPrefix: "",
			modelName:             "model",
			prefix:                "p",
			maxConsumers:          1,
			expected:              "p-0",
		},
		{
			name:                  "all params",
			namespace:             "default",
			consumerGroupIdPrefix: "foo",
			modelName:             "model",
			prefix:                "p",
			maxConsumers:          1,
			expected:              "foo-default-p-0",
		},
		{
			name:                  "no consumer prefix",
			namespace:             "default",
			consumerGroupIdPrefix: "",
			modelName:             "model",
			prefix:                "p",
			maxConsumers:          1,
			expected:              "default-p-0",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g.Expect(
				GetKafkaConsumerName(test.namespace, test.consumerGroupIdPrefix, test.modelName, test.prefix, test.maxConsumers),
			).To(Equal(
				test.expected,
			))

		})
	}
}
