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

package kafka

import (
	"testing"

	"github.com/seldonio/seldon-core/apis/go/v2/mlops/chainer"

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
		expected     []*chainer.PipelineTensorMapping
	}

	tests := []test{
		{
			name:         "basic",
			pipelineName: "test",
			topicNamer:   NewTopicNamer("default"),
			in:           map[string]string{"step.inputs.t1": "t1in", "step.inputs.t2": "t2in"},
			expected: []*chainer.PipelineTensorMapping{
				{
					PipelineName:   "test",
					TopicAndTensor: "seldon.default.model.step.inputs.t1",
					TensorName:     "t1in",
				},
				{
					PipelineName:   "test",
					TopicAndTensor: "seldon.default.model.step.inputs.t2",
					TensorName:     "t2in",
				},
			},
		},
		{
			name:         "pipeline references",
			pipelineName: "test",
			topicNamer:   NewTopicNamer("default"),
			in:           map[string]string{"test.inputs.t1": "t1"},
			expected: []*chainer.PipelineTensorMapping{
				{
					PipelineName:   "test",
					TopicAndTensor: "seldon.default.pipeline.test.inputs.t1",
					TensorName:     "t1",
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			results := test.topicNamer.GetFullyQualifiedTensorMap(test.pipelineName, test.in)
			for _, tm := range results {
				found := false
				for _, exp := range test.expected {
					if tm.PipelineName == exp.PipelineName &&
						tm.TopicAndTensor == exp.TopicAndTensor &&
						tm.TensorName == exp.TensorName {
						found = true
						break
					}
				}
				g.Expect(found).To(BeTrue())
			}
		})
	}
}

func TestGetFullyQualifiedPipelineTensorMap(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name       string
		topicNamer *TopicNamer
		in         map[string]string
		expected   *chainer.PipelineTensorMapping
	}

	tests := []test{
		{
			name:       "pipeline inputs",
			topicNamer: NewTopicNamer("default"),
			in: map[string]string{
				"pipeline1.inputs.input1": "t1in",
			},
			expected: &chainer.PipelineTensorMapping{
				PipelineName:   "pipeline1",
				TopicAndTensor: "seldon.default.pipeline.pipeline1.inputs.input1",
				TensorName:     "t1in",
			},
		},
		{
			name:       "pipeline outputs",
			topicNamer: NewTopicNamer("default"),
			in: map[string]string{
				"pipeline2.outputs.output1": "output2",
			},
			expected: &chainer.PipelineTensorMapping{
				PipelineName:   "pipeline2",
				TopicAndTensor: "seldon.default.pipeline.pipeline2.outputs.output1",
				TensorName:     "output2",
			},
		},
		{
			name:       "basic",
			topicNamer: NewTopicNamer("default"),
			in: map[string]string{
				"pipeline3.steps.model1.outputs.output1": "output2",
			},
			expected: &chainer.PipelineTensorMapping{
				PipelineName:   "pipeline3",
				TopicAndTensor: "seldon.default.model.model1.outputs.output1",
				TensorName:     "output2",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			results := test.topicNamer.GetFullyQualifiedPipelineTensorMap(test.in)
			result := results[0]
			g.Expect(result.PipelineName).To(Equal(test.expected.PipelineName))
			g.Expect(result.TopicAndTensor).To(Equal(test.expected.TopicAndTensor))
			g.Expect(result.TensorName).To(Equal(test.expected.TensorName))
		})
	}
}

func TestGetModelNameFromModelInputTopic(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name          string
		namespace     string
		topic         string
		expectedModel string
		err           bool
	}

	tests := []test{
		{
			name:          "model from topic ok",
			namespace:     "default",
			topic:         "seldon.default.model.mymodel.inputs",
			expectedModel: "mymodel",
		},
		{
			name:      "topic wrong number of separators",
			namespace: "default",
			topic:     "seldon.default.default.model.mymodel.inputs",
			err:       true,
		},
		{
			name:      "bad namespace",
			namespace: "default",
			topic:     "seldon.foo.model.mymodel.inputs",
			err:       true,
		},
		{
			name:      "bad prefix",
			namespace: "default",
			topic:     "seldons.default.model.mymodel.inputs",
			err:       true,
		},
		{
			name:      "bad model separator",
			namespace: "default",
			topic:     "seldon.default.models.mymodel.inputs",
			err:       true,
		},
		{
			name:      "bad model separator",
			namespace: "default",
			topic:     "seldon.default.models.mymodel.outputs",
			err:       true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tn := NewTopicNamer(test.namespace)
			modelName, err := tn.GetModelNameFromModelInputTopic(test.topic)
			if test.err {
				g.Expect(err).ToNot(BeNil())
			} else {
				g.Expect(err).To(BeNil())
				g.Expect(modelName).To(Equal(test.expectedModel))
			}
		})
	}
}
