/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package gateway

import (
	"testing"

	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"

	kafka_config "github.com/seldonio/seldon-core/components/kafka/v2/pkg/config"

	seldontracer "github.com/seldonio/seldon-core/scheduler/v2/pkg/tracing"
)

func TestAddRemoveModel(t *testing.T) {
	g := NewGomegaWithT(t)

	type expected struct {
		models    int
		consumers int
	}

	type test struct {
		name             string
		removeModels     []string
		addModels        []string
		numConsumers     int
		runningConsumers int
		expected         expected
	}
	tests := []test{
		{
			name:             "one model - one consumer",
			removeModels:     []string{"foo"},
			addModels:        []string{"foo"},
			numConsumers:     1,
			runningConsumers: 1,
			expected: expected{
				models:    0,
				consumers: 0,
			},
		},
		{
			name:             "one model - two consumers",
			removeModels:     []string{"foo"},
			addModels:        []string{"foo"},
			numConsumers:     2,
			runningConsumers: 1,
			expected: expected{
				models:    0,
				consumers: 0,
			},
		},
		{
			name:             "two models - one consumer",
			removeModels:     []string{"foo", "bar"},
			addModels:        []string{"foo", "bar"},
			numConsumers:     1,
			runningConsumers: 1,
			expected: expected{
				models:    0,
				consumers: 0,
			},
		},
		{
			name:             "two models - two consumers",
			removeModels:     []string{"foo", "bar"},
			addModels:        []string{"foo", "bar"},
			numConsumers:     2,
			runningConsumers: 2,
			expected: expected{
				models:    0,
				consumers: 0,
			},
		},
		{
			name:             "remove model which does not exist - idempotent",
			removeModels:     []string{"bar"},
			addModels:        []string{"foo"},
			numConsumers:     2,
			runningConsumers: 1,
			expected: expected{
				models:    1,
				consumers: 1,
			},
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
			c := &ManagerConfig{SeldonKafkaConfig: &kafka_config.KafkaConfig{}, Namespace: "default", InferenceServerConfig: &kafkaServerConfig, TraceProvider: tp, NumWorkers: 0}
			cm, err := NewConsumerManager(logger, c, test.numConsumers, nil)
			g.Expect(err).To(BeNil())
			for _, model := range test.addModels {
				err := cm.AddModel(model)
				g.Expect(err).To(BeNil())
			}
			g.Expect(len(cm.consumers)).To(Equal(test.runningConsumers))

			for _, model := range test.removeModels {
				err := cm.RemoveModel(model, false, false)
				g.Expect(err).To(BeNil())
			}
			g.Expect(cm.GetNumModels()).To(Equal(test.expected.models))
			g.Expect(len(cm.consumers)).To(Equal(test.expected.consumers))
		})
	}
}

func TestExists(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name           string
		modelsConsumed []string
		model          string
		exists         bool
	}
	tests := []test{
		{
			name:           "exists",
			modelsConsumed: []string{"foo", "bar"},
			model:          "foo",
			exists:         true,
		},
		{
			name:           "doesnt Exist",
			modelsConsumed: []string{"foo"},
			model:          "bar",
			exists:         false,
		},
		{
			name:           "empty",
			modelsConsumed: []string{},
			model:          "bar",
			exists:         false,
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
			c := &ManagerConfig{SeldonKafkaConfig: &kafka_config.KafkaConfig{}, Namespace: "default", InferenceServerConfig: &kafkaServerConfig, TraceProvider: tp, NumWorkers: 0}
			cm, err := NewConsumerManager(logger, c, 5, nil)
			g.Expect(err).To(BeNil())
			for _, model := range test.modelsConsumed {
				err := cm.AddModel(model)
				g.Expect(err).To(BeNil())
			}
			g.Expect(cm.Exists(test.model)).To(Equal(test.exists))
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
			c := &ManagerConfig{SeldonKafkaConfig: &kafka_config.KafkaConfig{}, Namespace: "default", InferenceServerConfig: &kafkaServerConfig, TraceProvider: tp, NumWorkers: 0}
			cm, err := NewConsumerManager(logger, c, 10, nil)
			g.Expect(err).To(BeNil())
			ic, _ := cm.getInferKafkaConsumer(test.model, true)
			ic2, _ := cm.getInferKafkaConsumer(test.model, false)
			g.Expect(ic).To(Equal(ic2))
		})
	}
}
