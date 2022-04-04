package pipeline

import (
	"testing"

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
		kafkaManager      *KafkaManager
		pipeline          *Pipeline
		resourceName      string
		isModel           bool
		expectedPipelines int
	}
	tests := []test{
		{
			name:              "model",
			kafkaManager:      NewKafkaManager(logrus.New(), "default"),
			resourceName:      "foo",
			isModel:           true,
			expectedPipelines: 1,
		},
		{
			name:              "model - existing in map",
			kafkaManager:      NewKafkaManager(logrus.New(), "default"),
			pipeline:          &Pipeline{},
			resourceName:      "foo",
			isModel:           true,
			expectedPipelines: 1,
		},
		{
			name:              "pipeline",
			kafkaManager:      NewKafkaManager(logrus.New(), "default"),
			resourceName:      "foo",
			isModel:           false,
			expectedPipelines: 1,
		},
		{
			name:              "pipeline - existing in map",
			kafkaManager:      NewKafkaManager(logrus.New(), "default"),
			pipeline:          &Pipeline{},
			resourceName:      "foo",
			isModel:           false,
			expectedPipelines: 1,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.pipeline != nil {
				test.kafkaManager.pipelines.Store(getPipelineKey(test.resourceName, test.isModel), test.pipeline)
			}
			pipeline, err := test.kafkaManager.loadOrStorePipeline(test.resourceName, test.isModel)
			g.Expect(err).To(BeNil())
			g.Expect(pipeline).ToNot(BeNil())
			count := 0
			test.kafkaManager.pipelines.Range(func(key interface{}, val interface{}) bool {
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
