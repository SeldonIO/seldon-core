package kafka

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestGetModelTopic(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name          string
		pipelineName  string
		topicNamer    *TopicNamer
		stepReference string
		expected      string
	}

	tests := []test{
		{
			name:          "model",
			pipelineName:  "test",
			topicNamer:    NewTopicNamer("default"),
			stepReference: "model1.inputs",
			expected:      "seldon.default.model.model1.inputs",
		},
		{
			name:          "model with tensor",
			pipelineName:  "test",
			topicNamer:    NewTopicNamer("default"),
			stepReference: "model1.outputs.t1",
			expected:      "seldon.default.model.model1.outputs.t1",
		},
		{
			name:          "pipeline reference",
			pipelineName:  "test",
			topicNamer:    NewTopicNamer("default"),
			stepReference: "test.inputs",
			expected:      "seldon.default.pipeline.test.inputs",
		},
		{
			name:          "pipeline reference tensor",
			pipelineName:  "test",
			topicNamer:    NewTopicNamer("default"),
			stepReference: "test.inputs.t1",
			expected:      "seldon.default.pipeline.test.inputs.t1",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res := test.topicNamer.GetModelOrPipelineTopic(test.pipelineName, test.stepReference)
			g.Expect(res).To(Equal(test.expected))
		})
	}
}

func TestGetFullyQualifiedTensorMap(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name         string
		pipelineName string
		topicNamer   *TopicNamer
		in           map[string]string
		expected     map[string]string
	}

	tests := []test{
		{
			name:         "basic",
			pipelineName: "test",
			topicNamer:   NewTopicNamer("default"),
			in:           map[string]string{"step.inputs.t1": "t1in", "step.inputs.t2": "t2in"},
			expected:     map[string]string{"seldon.default.model.step.inputs.t1": "t1in", "seldon.default.model.step.inputs.t2": "t2in"},
		},
		{
			name:         "pipeline references",
			pipelineName: "test",
			topicNamer:   NewTopicNamer("default"),
			in:           map[string]string{"test.inputs.t1": "t1"},
			expected:     map[string]string{"seldon.default.pipeline.test.inputs.t1": "t1"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res := test.topicNamer.GetFullyQualifiedTensorMap(test.pipelineName, test.in)
			for k, v := range res {
				g.Expect(v).To(Equal(test.expected[k]))
			}

		})
	}
}
