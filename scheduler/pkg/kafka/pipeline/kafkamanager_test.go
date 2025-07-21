/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package pipeline

import (
	"sync"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"

	kafka_config "github.com/seldonio/seldon-core/components/kafka/v2/pkg/config"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/tracing"
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

func TestStoreAndLoadPipeline(t *testing.T) {
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
			km, err := NewKafkaManager(logrus.New(), "default", &kafka_config.KafkaConfig{}, tracer, 10)
			g.Expect(err).To(BeNil())
			if test.pipeline != nil {
				km.pipelines.Store(getPipelineKey(test.resourceName, test.isModel), test.pipeline)
			}
			err = km.StorePipeline(test.resourceName, test.isModel)
			g.Expect(err).To(BeNil())
			pipeline, err := km.LoadPipeline(test.resourceName, test.isModel)
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
