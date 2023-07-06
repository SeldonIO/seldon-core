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

package gateway

import (
	"testing"

	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/kafka/config"
	seldontracer "github.com/seldonio/seldon-core/scheduler/v2/pkg/tracing"
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
			c := &ManagerConfig{SeldonKafkaConfig: &config.KafkaConfig{}, Namespace: "default", InferenceServerConfig: &kafkaServerConfig, TraceProvider: tp, NumWorkers: 0}
			cm, err := NewConsumerManager(logger, c, test.numConsumers)
			g.Expect(err).To(BeNil())
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
			c := &ManagerConfig{SeldonKafkaConfig: &config.KafkaConfig{}, Namespace: "default", InferenceServerConfig: &kafkaServerConfig, TraceProvider: tp, NumWorkers: 0}
			cm, err := NewConsumerManager(logger, c, 10)
			g.Expect(err).To(BeNil())
			ic, _ := cm.getInferKafkaConsumer(test.model, true)
			ic2, _ := cm.getInferKafkaConsumer(test.model, false)
			g.Expect(ic).To(Equal(ic2))
		})
	}
}
