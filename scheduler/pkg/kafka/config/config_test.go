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

package config

import (
	"fmt"
	"os"
	"testing"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	. "github.com/onsi/gomega"
)

func TestNewKafkaConfig(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name     string
		data     string
		expected *KafkaConfig
		err      bool
	}

	tests := []test{
		{
			name: "with bootstrap servers",
			data: `
{
  "bootstrap.servers":"kafka:9092",
  "consumer":{"session.timeout.ms": 6000, "someBool": true, "someString":"foo"},
  "producer": {"linger.ms":0},
  "streams": {"replication.factor": 1}
}
`,
			expected: &KafkaConfig{
				BootstrapServers: "kafka:9092",
				Consumer:         kafka.ConfigMap{"bootstrap.servers": "kafka:9092", "session.timeout.ms": 6000, "someBool": true, "someString": "foo"},
				Producer:         kafka.ConfigMap{"bootstrap.servers": "kafka:9092", "linger.ms": 0},
				Streams:          kafka.ConfigMap{"bootstrap.servers": "kafka:9092", "replication.factor": 1},
			},
		},
		{
			name: "without bootstrap servers override",
			data: `
{
  "bootstrap.servers":"kafka:9092",
  "consumer":{"bootstrap.servers":"foo","session.timeout.ms": 6000, "someBool": true, "someString":"foo"},
  "producer": {"bootstrap.servers":"foo","linger.ms":0},
  "streams": {"bootstrap.servers":"foo","replication.factor": 1}
}
`,
			expected: &KafkaConfig{
				BootstrapServers: "kafka:9092",
				Consumer:         kafka.ConfigMap{"bootstrap.servers": "foo", "session.timeout.ms": 6000, "someBool": true, "someString": "foo"},
				Producer:         kafka.ConfigMap{"bootstrap.servers": "foo", "linger.ms": 0},
				Streams:          kafka.ConfigMap{"bootstrap.servers": "foo", "replication.factor": 1},
			},
		},
		{
			name: "error",
			data: `{"foo":"bar"}`,
			err:  true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			configFilePath := fmt.Sprintf("%s/kafka.json", t.TempDir())
			err := os.WriteFile(configFilePath, []byte(test.data), 0644)
			g.Expect(err).To(BeNil())
			kc, err := NewKafkaConfig(configFilePath)
			if test.err {
				g.Expect(err).ToNot(BeNil())
			} else {
				g.Expect(kc).To(Equal(test.expected))
			}
		})
	}
}
