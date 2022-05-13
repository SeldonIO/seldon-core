package pipeline

import (
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
	tracer, err := tracing.NewTracer("test")
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
			pipeline:          &Pipeline{},
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
			pipeline:          &Pipeline{},
			resourceName:      "foo",
			isModel:           false,
			expectedPipelines: 1,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			km, err := NewKafkaManager(logrus.New(), "default", &config.KafkaConfig{}, tracer)
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
