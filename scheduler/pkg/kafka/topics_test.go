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

	. "github.com/onsi/gomega"

	"github.com/seldonio/seldon-core/apis/go/v2/mlops/chainer"
)

func TestGetModelTopic(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name           string
		pipelineName   string
		topicNamer     *TopicNamer
		stepReference  string
		expectedTopic  string
		expectedTensor *string
	}

	createTopicNamer := func(namespace string, topicPrefix string) *TopicNamer {
		tn, err := NewTopicNamer(namespace, topicPrefix)
		g.Expect(err).To(BeNil())
		return tn
	}
	getPtrStr := func(val string) *string { return &val }
	tests := []test{
		{
			name:           "model",
			pipelineName:   "test",
			topicNamer:     createTopicNamer("default", ""),
			stepReference:  "model1.inputs",
			expectedTopic:  "seldon.default.model.model1.inputs",
			expectedTensor: nil,
		},
		{
			name:           "model with custom prefix",
			pipelineName:   "test",
			topicNamer:     createTopicNamer("default", "foo.bar"),
			stepReference:  "model1.inputs",
			expectedTopic:  "foo.bar.default.model.model1.inputs",
			expectedTensor: nil,
		},
		{
			name:           "model with custom prefix with a dot",
			pipelineName:   "test",
			topicNamer:     createTopicNamer("default", "foo.bar."),
			stepReference:  "model1.inputs",
			expectedTopic:  "foo.bar..default.model.model1.inputs",
			expectedTensor: nil,
		},
		{
			name:           "model with tensor",
			pipelineName:   "test",
			topicNamer:     createTopicNamer("default", defaultSeldonTopicPrefix),
			stepReference:  "model1.outputs.t1",
			expectedTopic:  "seldon.default.model.model1.outputs",
			expectedTensor: getPtrStr("t1"),
		},
		{
			name:           "pipeline reference",
			pipelineName:   "test",
			topicNamer:     createTopicNamer("default", defaultSeldonTopicPrefix),
			stepReference:  "test.inputs",
			expectedTopic:  "seldon.default.pipeline.test.inputs",
			expectedTensor: nil,
		},
		{
			name:           "pipeline reference tensor",
			pipelineName:   "test",
			topicNamer:     createTopicNamer("default", defaultSeldonTopicPrefix),
			stepReference:  "test.inputs.t1",
			expectedTopic:  "seldon.default.pipeline.test.inputs",
			expectedTensor: getPtrStr("t1"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res, tensor := test.topicNamer.GetModelOrPipelineTopicAndTensor(test.pipelineName, test.stepReference)
			g.Expect(res).To(Equal(test.expectedTopic))
			if tensor == nil {
				g.Expect(test.expectedTensor).To(BeNil())
			} else {
				g.Expect(test.expectedTensor).ToNot(BeNil())
				g.Expect(*tensor).To(Equal(*test.expectedTensor))
			}
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

	createTopicNamer := func(namespace string, topicPrefix string) *TopicNamer {
		tn, err := NewTopicNamer(namespace, topicPrefix)
		g.Expect(err).To(BeNil())
		return tn
	}
	tests := []test{
		{
			name:         "basic",
			pipelineName: "test",
			topicNamer:   createTopicNamer("default", defaultSeldonTopicPrefix),
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
			name:         "basic with custom prefix",
			pipelineName: "test",
			topicNamer:   createTopicNamer("default", "foo.bar"),
			in:           map[string]string{"step.inputs.t1": "t1in", "step.inputs.t2": "t2in"},
			expected: []*chainer.PipelineTensorMapping{
				{
					PipelineName:   "test",
					TopicAndTensor: "foo.bar.default.model.step.inputs.t1",
					TensorName:     "t1in",
				},
				{
					PipelineName:   "test",
					TopicAndTensor: "foo.bar.default.model.step.inputs.t2",
					TensorName:     "t2in",
				},
			},
		},
		{
			name:         "pipeline references",
			pipelineName: "test",
			topicNamer:   createTopicNamer("default", defaultSeldonTopicPrefix),
			in:           map[string]string{"test.inputs.t1": "t1"},
			expected: []*chainer.PipelineTensorMapping{
				{
					PipelineName:   "test",
					TopicAndTensor: "seldon.default.pipeline.test.inputs.t1",
					TensorName:     "t1",
				},
			},
		},
		{
			name:         "pipeline references with custom prefix",
			pipelineName: "test",
			topicNamer:   createTopicNamer("default", "foo.bar"),
			in:           map[string]string{"test.inputs.t1": "t1"},
			expected: []*chainer.PipelineTensorMapping{
				{
					PipelineName:   "test",
					TopicAndTensor: "foo.bar.default.pipeline.test.inputs.t1",
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

	createTopicNamer := func(namespace string, topicPrefix string) *TopicNamer {
		tn, err := NewTopicNamer(namespace, topicPrefix)
		g.Expect(err).To(BeNil())
		return tn
	}
	tests := []test{
		{
			name:       "pipeline inputs",
			topicNamer: createTopicNamer("default", defaultSeldonTopicPrefix),
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
			name:       "pipeline inputs custom prefix",
			topicNamer: createTopicNamer("default", "foo.bar"),
			in: map[string]string{
				"pipeline1.inputs.input1": "t1in",
			},
			expected: &chainer.PipelineTensorMapping{
				PipelineName:   "pipeline1",
				TopicAndTensor: "foo.bar.default.pipeline.pipeline1.inputs.input1",
				TensorName:     "t1in",
			},
		},
		{
			name:       "pipeline outputs with custom prefix",
			topicNamer: createTopicNamer("default", "foo"),
			in: map[string]string{
				"pipeline2.outputs.output1": "output2",
			},
			expected: &chainer.PipelineTensorMapping{
				PipelineName:   "pipeline2",
				TopicAndTensor: "foo.default.pipeline.pipeline2.outputs.output1",
				TensorName:     "output2",
			},
		},
		{
			name:       "basic",
			topicNamer: createTopicNamer("default", defaultSeldonTopicPrefix),
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
		topicPrefix   string
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
			name:          "model from topic ok with dot in prefix",
			namespace:     "default",
			topicPrefix:   "seldon.",
			topic:         "seldon..default.model.mymodel.inputs",
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
			name:      "bad default prefix",
			namespace: "default",
			topic:     "seldons.default.model.mymodel.inputs",
			err:       true,
		},
		{
			name:      "bad model separator models",
			namespace: "default",
			topic:     "seldon.default.models.mymodel.inputs",
			err:       true,
		},
		{
			name:      "not input separator",
			namespace: "default",
			topic:     "seldon.default.models.mymodel.inputs",
			err:       true,
		},
		{
			name:          "custom prefix",
			topicPrefix:   "foo.bar",
			namespace:     "default",
			topic:         "foo.bar.default.model.mymodel.inputs",
			expectedModel: "mymodel",
		},
		{
			name:        "custom prefix bad separator models",
			topicPrefix: "foo.bar",
			namespace:   "default",
			topic:       "foo.bar.default.models.mymodel.inputs",
			err:         true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tn, err := NewTopicNamer(test.namespace, test.topicPrefix)
			g.Expect(err).To(BeNil())
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

func TestGetTopicAndTensor(t *testing.T) {
	g := NewGomegaWithT(t)

	getPtrStr := func(val string) *string { return &val }
	type test struct {
		name           string
		reference      string
		expectedTopic  string
		expectedTensor *string
	}

	tests := []test{
		{
			name:           "has tensor",
			reference:      "model.inputs.tensor",
			expectedTopic:  "model.inputs",
			expectedTensor: getPtrStr("tensor"),
		},
		{
			name:           "no tensor",
			reference:      "model.inputs",
			expectedTopic:  "model.inputs",
			expectedTensor: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tn, err := NewTopicNamer("foo", "seldon-mesh")
			g.Expect(err).To(BeNil())
			topic, tensor := tn.getTopicReferenceAndTensor(test.reference)
			g.Expect(topic).To(Equal(test.expectedTopic))
			if tensor == nil {
				g.Expect(test.expectedTensor).To(BeNil())
			} else {
				g.Expect(test.expectedTensor).ToNot(BeNil())
				g.Expect(*tensor).To(Equal(*test.expectedTensor))
			}
		})
	}
}

func TestNewTopicNamer(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name        string
		namespace   string
		topicPrefix string
		err         bool
	}

	tests := []test{
		{
			name:        "default empty",
			namespace:   "seldon-mesh",
			topicPrefix: "",
			err:         false,
		},
		{
			name:        "normal prefix",
			namespace:   "seldon-mesh",
			topicPrefix: "myprefix",
			err:         false,
		},
		{
			name:        "normal prefix with dash",
			namespace:   "seldon-mesh",
			topicPrefix: "my-prefix",
			err:         false,
		},
		{
			name:        "normal prefix with dot",
			namespace:   "seldon-mesh",
			topicPrefix: "my.prefix",
			err:         false,
		},
		{
			name:        "normal prefix with underscore",
			namespace:   "seldon-mesh",
			topicPrefix: "my_prefix",
			err:         false,
		},
		{
			name:        "bad prefix",
			namespace:   "seldon-mesh",
			topicPrefix: "my_prefix&",
			err:         true,
		},
		{
			name:        "prefix with digits",
			namespace:   "seldon-mesh",
			topicPrefix: "a51691-preprod-seldon",
			err:         false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewTopicNamer(test.namespace, test.topicPrefix)
			if test.err {
				g.Expect(err).ToNot(BeNil())
			} else {
				g.Expect(err).To(BeNil())
			}
		})
	}
}
