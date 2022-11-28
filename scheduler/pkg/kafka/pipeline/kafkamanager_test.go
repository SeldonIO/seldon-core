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

package pipeline

import (
	"sync"
	"testing"

	"github.com/seldonio/seldon-core/scheduler/pkg/kafka/config"

	"github.com/seldonio/seldon-core/scheduler/pkg/tracing"

	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
)

func TestGetPipelineKey(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name         string
		resourceName string
		isModel      bool
		expected     string
	}

	tests := []test{
		{
			name:         "pipeline",
			resourceName: "foo",
			isModel:      false,
			expected:     "foo.pipeline",
		},
		{
			name:         "model",
			resourceName: "foo",
			isModel:      true,
			expected:     "foo.model",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			key := getPipelineKey(test.resourceName, test.isModel)
			g.Expect(key).To(Equal(test.expected))
		})
	}
}

func TestLoadOrStorePipeline(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name              string
		pipeline          *Pipeline
		resourceName      string
		isModel           bool
		expectedPipelines int
	}
	logger := logrus.New()
	tracer, err := tracing.NewTraceProvider("test", nil, logger)
	g.Expect(err).To(BeNil())
	tests := []test{
		{
			name:              "model",
			resourceName:      "foo",
			isModel:           true,
			expectedPipelines: 1,
		},
		{
			name:              "model - existing in map",
			pipeline:          &Pipeline{wg: &sync.WaitGroup{}},
			resourceName:      "foo",
			isModel:           true,
			expectedPipelines: 1,
		},
		{
			name:              "pipeline",
			resourceName:      "foo",
			isModel:           false,
			expectedPipelines: 1,
		},
		{
			name:              "pipeline - existing in map",
			pipeline:          &Pipeline{wg: &sync.WaitGroup{}},
			resourceName:      "foo",
			isModel:           false,
			expectedPipelines: 1,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			km, err := NewKafkaManager(logrus.New(), "default", &config.KafkaConfig{}, tracer, 10, 10)
			g.Expect(err).To(BeNil())
			if test.pipeline != nil {
				km.pipelines.Store(getPipelineKey(test.resourceName, test.isModel), test.pipeline)
			}
			pipeline, err := km.loadOrStorePipeline(test.resourceName, test.isModel)
			g.Expect(err).To(BeNil())
			g.Expect(pipeline).ToNot(BeNil())
			count := 0
			km.pipelines.Range(func(key interface{}, val interface{}) bool {
				count = count + 1
				return true
			})
			g.Expect(count).To(Equal(test.expectedPipelines))
			if test.pipeline != nil {
				g.Expect(pipeline).To(Equal(test.pipeline))
			}
		})
	}
}
