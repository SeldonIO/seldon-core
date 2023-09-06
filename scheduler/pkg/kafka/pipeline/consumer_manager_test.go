/*
Copyright 2023 Seldon Technologies Ltd.

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

package pipeline

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestGetKafkaConsumerName(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name                  string
		namespace             string
		consumerGroupIdPrefix string
		componentPrefix       string
		id                    string
		expected              string
	}
	tests := []test{
		{
			name:                  "all params no namespace",
			namespace:             "",
			consumerGroupIdPrefix: "foo",
			componentPrefix:       "pipeline",
			id:                    "id",
			expected:              "foo-pipeline-id",
		},
		{
			name:                  "no consumer group prefix no namespace",
			namespace:             "",
			consumerGroupIdPrefix: "",
			componentPrefix:       "pipeline",
			id:                    "id",
			expected:              "pipeline-id",
		},
		{
			name:                  "all params",
			namespace:             "default",
			consumerGroupIdPrefix: "foo",
			componentPrefix:       "pipeline",
			id:                    "id",
			expected:              "foo-default-pipeline-id",
		},
		{
			name:                  "no consumer group prefix",
			namespace:             "default",
			consumerGroupIdPrefix: "",
			componentPrefix:       "pipeline",
			id:                    "id",
			expected:              "default-pipeline-id",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g.Expect(
				getKafkaConsumerName(
					test.namespace, test.consumerGroupIdPrefix, test.componentPrefix, test.id),
			).To(Equal(
				test.expected),
			)
		})
	}
}
