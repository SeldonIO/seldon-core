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

package dataflow

import (
	"testing"

	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"

	"github.com/seldonio/seldon-core/apis/go/v2/mlops/chainer"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/kafka"
)

func TestCreateTopicSources(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name         string
		server       *ChainerServer
		pipelineName string
		inputs       []string
		sources      []*chainer.PipelineTopic
	}

	getPtrStr := func(val string) *string { return &val }
	createTopicNamer := func(namespace string, topicPrefix string) *kafka.TopicNamer {
		tn, err := kafka.NewTopicNamer(namespace, topicPrefix)
		g.Expect(err).To(BeNil())
		return tn
	}
	tests := []test{
		{
			name: "misc inputs",
			server: &ChainerServer{
				logger:     log.New(),
				topicNamer: createTopicNamer("default", "seldon"),
			},
			pipelineName: "p1",
			inputs: []string{
				"a",
				"b.inputs",
				"c.inputs.t1",
			},
			sources: []*chainer.PipelineTopic{
				{PipelineName: "p1", TopicName: "seldon.default.model.a", Tensor: nil},
				{PipelineName: "p1", TopicName: "seldon.default.model.b.inputs", Tensor: nil},
				{PipelineName: "p1", TopicName: "seldon.default.model.c.inputs", Tensor: getPtrStr("t1")},
			},
		},
		{
			name: "default inputs",
			server: &ChainerServer{
				logger:     log.New(),
				topicNamer: createTopicNamer("ns1", "seldon"),
			},
			pipelineName: "p1",
			inputs:       []string{},
			sources: []*chainer.PipelineTopic{
				{PipelineName: "p1", TopicName: "seldon.ns1.pipeline.p1.inputs", Tensor: nil},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sources := test.server.createTopicSources(test.inputs, test.pipelineName)
			g.Expect(sources).To(Equal(test.sources))
		})
	}
}

func TestCreatePipelineTopicSources(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name    string
		server  *ChainerServer
		inputs  []string
		sources []*chainer.PipelineTopic
	}

	getPtrStr := func(val string) *string { return &val }
	createTopicNamer := func(namespace string, topicPrefix string) *kafka.TopicNamer {
		tn, err := kafka.NewTopicNamer(namespace, topicPrefix)
		g.Expect(err).To(BeNil())
		return tn
	}
	tests := []test{
		{
			name: "misc inputs",
			server: &ChainerServer{
				logger:     log.New(),
				topicNamer: createTopicNamer("default", "seldon"),
			},
			inputs: []string{
				"foo.inputs",
				"foo.outputs",
				"foo.step.bar.inputs",
				"foo.step.bar.outputs",
				"foo.step.bar.inputs.tensora",
			},
			sources: []*chainer.PipelineTopic{
				{PipelineName: "foo", TopicName: "seldon.default.pipeline.foo.inputs", Tensor: nil},
				{PipelineName: "foo", TopicName: "seldon.default.pipeline.foo.outputs", Tensor: nil},
				{PipelineName: "foo", TopicName: "seldon.default.model.bar.inputs", Tensor: nil},
				{PipelineName: "foo", TopicName: "seldon.default.model.bar.outputs", Tensor: nil},
				{PipelineName: "foo", TopicName: "seldon.default.model.bar.inputs", Tensor: getPtrStr("tensora")},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sources := test.server.createPipelineTopicSources(test.inputs)
			g.Expect(sources).To(Equal(test.sources))
		})
	}
}

func TestCreateTriggerSources(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name         string
		server       *ChainerServer
		pipelineName string
		inputs       []string
		sources      []*chainer.PipelineTopic
	}

	createTopicNamer := func(namespace string, topicPrefix string) *kafka.TopicNamer {
		tn, err := kafka.NewTopicNamer(namespace, topicPrefix)
		g.Expect(err).To(BeNil())
		return tn
	}
	getPtrStr := func(val string) *string { return &val }
	tests := []test{
		{
			name: "misc inputs",
			server: &ChainerServer{
				logger:     log.New(),
				topicNamer: createTopicNamer("default", "seldon"),
			},
			pipelineName: "p1",
			inputs: []string{
				"a",
				"b.inputs",
				"c.inputs.t1",
			},
			sources: []*chainer.PipelineTopic{
				{PipelineName: "p1", TopicName: "seldon.default.model.a", Tensor: nil},
				{PipelineName: "p1", TopicName: "seldon.default.model.b.inputs", Tensor: nil},
				{PipelineName: "p1", TopicName: "seldon.default.model.c.inputs", Tensor: getPtrStr("t1")},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sources := test.server.createTriggerSources(test.inputs, test.pipelineName)
			g.Expect(sources).To(Equal(test.sources))
		})
	}
}
