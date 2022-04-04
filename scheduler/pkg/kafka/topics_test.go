package kafka

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestGetModelTopic(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name        string
		topicNamer  *TopicNamer
		modelStream string
		expected    string
	}

	tests := []test{
		{
			name:        "no model name - just step name",
			topicNamer:  NewTopicNamer("default"),
			modelStream: "step1",
			expected:    "seldon.default.model.step1",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res := test.topicNamer.GetModelTopic(test.modelStream)
			g.Expect(res).To(Equal(test.expected))
		})
	}
}

func TestGetFullyQualifiedTensorMap(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name       string
		topicNamer *TopicNamer
		in         map[string]string
		expected   map[string]string
	}

	tests := []test{
		{
			name:       "basic",
			topicNamer: NewTopicNamer("default"),
			in:         map[string]string{"step.inputs.t1": "t1in", "step.inputs.t2": "t2in"},
			expected:   map[string]string{"seldon.default.model.step.inputs.t1": "t1in", "seldon.default.model.step.inputs.t2": "t2in"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res := test.topicNamer.GetFullyQualifiedTensorMap(test.in)
			for k, v := range res {
				g.Expect(v).To(Equal(test.expected[k]))
			}

		})
	}
}
