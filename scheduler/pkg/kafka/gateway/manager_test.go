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
		name                 string
		models               []string
		numModelsPerConsumer int
	}
	tests := []test{
		{
			name:                 "one model - many consumers",
			models:               []string{"foo"},
			numModelsPerConsumer: 100,
		},
		{
			name:                 "two models - many consumers",
			models:               []string{"foo", "bar"},
			numModelsPerConsumer: 100,
		},
		{
			name:                 "two models - 1 model per consumer",
			models:               []string{"foo", "bar"},
			numModelsPerConsumer: 1,
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
			cm := NewConsumerManager(logger, c, test.numModelsPerConsumer)
			for _, model := range test.models {
				err := cm.AddModel(model)
				g.Expect(err).To(BeNil())
			}
			for _, model := range test.models {
				err := cm.RemoveModel(model)
				g.Expect(err).To(BeNil())
			}
			g.Expect(cm.GetNumModels()).To(Equal(0))
		})
	}
}
