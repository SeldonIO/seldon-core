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
