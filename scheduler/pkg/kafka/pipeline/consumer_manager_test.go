package pipeline

import (
	"fmt"
	. "github.com/onsi/gomega"
	"testing"
)

func TestGetKafkaConsumerName(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name                  string
		consumerGroupIdPrefix string
		componentPrefix       string
		id                    string
		expected              string
	}
	tests := []test{
		{
			name:                  "all params",
			consumerGroupIdPrefix: "foo",
			componentPrefix:       "pipeline",
			id:                    "id",
			expected:              fmt.Sprintf("%s-%s-%s", "foo", "pipeline", "id"),
		},
		{
			name:                  "no consumer group prefix",
			consumerGroupIdPrefix: "",
			componentPrefix:       "pipeline",
			id:                    "id",
			expected:              fmt.Sprintf("%s-%s", "pipeline", "id"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g.Expect(getKafkaConsumerName(test.consumerGroupIdPrefix, test.componentPrefix, test.id)).To(Equal(test.expected))

		})
	}
}
