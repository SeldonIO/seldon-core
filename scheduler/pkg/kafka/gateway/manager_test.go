package gateway

import (
	"testing"

	. "github.com/onsi/gomega"
	"github.com/seldonio/seldon-core/scheduler/pkg/kafka/config"
	seldontracer "github.com/seldonio/seldon-core/scheduler/pkg/tracing"
	log "github.com/sirupsen/logrus"
)

func TestAddRemoveModel(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name             string
		models           []string
		numConsumers     int
		runningConsumers int
	}
	tests := []test{
		{
			name:             "one model - one consumer",
			models:           []string{"foo"},
			numConsumers:     1,
			runningConsumers: 1,
		},
		{
			name:             "one model - two consumers",
			models:           []string{"foo"},
			numConsumers:     2,
			runningConsumers: 1,
		},
		{
			name:             "two models - one consumer",
			models:           []string{"foo", "bar"},
			numConsumers:     1,
			runningConsumers: 1,
		},
		{
			name:             "two models - two consumers",
			models:           []string{"foo", "bar"},
			numConsumers:     2,
			runningConsumers: 2,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			logger := log.New()
			t.Log("Start test", test.name)
			kafkaServerConfig := InferenceServerConfig{
				Host:     "0.0.0.0",
				HttpPort: 1234,
				GrpcPort: 1235,
			}
			tp, err := seldontracer.NewTraceProvider("test", nil, logger)
			g.Expect(err).To(BeNil())
			c := &ConsumerConfig{KafkaConfig: &config.KafkaConfig{}, Namespace: "default", InferenceServerConfig: &kafkaServerConfig, TraceProvider: tp, NumWorkers: 0}
			cm := NewConsumerManager(logger, c, test.numConsumers)
			for _, model := range test.models {
				err := cm.AddModel(model)
				g.Expect(err).To(BeNil())
			}
			g.Expect(len(cm.consumers)).To(Equal(test.runningConsumers))

			for _, model := range test.models {
				err := cm.RemoveModel(model)
				g.Expect(err).To(BeNil())
			}
			g.Expect(cm.GetNumModels()).To(Equal(0))
			g.Expect(len(cm.consumers)).To(Equal(0))
		})
	}
}

func TestConsistentModelToConsumer(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name  string
		model string
	}
	tests := []test{
		{
			name:  "smoke",
			model: "foo",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			logger := log.New()
			t.Log("Start test", test.name)
			kafkaServerConfig := InferenceServerConfig{
				Host:     "0.0.0.0",
				HttpPort: 1234,
				GrpcPort: 1235,
			}
			tp, err := seldontracer.NewTraceProvider("test", nil, logger)
			g.Expect(err).To(BeNil())
			c := &ConsumerConfig{KafkaConfig: &config.KafkaConfig{}, Namespace: "default", InferenceServerConfig: &kafkaServerConfig, TraceProvider: tp, NumWorkers: 0}
			cm := NewConsumerManager(logger, c, 10)
			ic, _ := cm.getInferKafkaConsumer(test.model, true)
			ic2, _ := cm.getInferKafkaConsumer(test.model, false)
			g.Expect(ic).To(Equal(ic2))
		})
	}
}
