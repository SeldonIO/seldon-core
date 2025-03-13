/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package config

import (
	"fmt"
	"os"
	"testing"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
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
				Consumer:         kafka.ConfigMap{"bootstrap.servers": "kafka:9092", "session.timeout.ms": 6000, "someBool": true, "someString": "foo", "log_level": 7},
				Producer:         kafka.ConfigMap{"bootstrap.servers": "kafka:9092", "linger.ms": 0, "log_level": 7},
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
				Consumer:         kafka.ConfigMap{"bootstrap.servers": "foo", "session.timeout.ms": 6000, "someBool": true, "someString": "foo", "log_level": 7},
				Producer:         kafka.ConfigMap{"bootstrap.servers": "foo", "linger.ms": 0, "log_level": 7},
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
			kc, err := NewKafkaConfig(configFilePath, "debug")
			if test.err {
				g.Expect(err).ToNot(BeNil())
			} else {
				g.Expect(kc).To(Equal(test.expected))
			}
		})
	}
}

func TestGetKafkaConsumerName(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name                  string
		namespace             string
		consumerGroupIdPrefix string
		componentPrefix       string
		id                    string
		expected              string
	}
	tests := []test{
		{
			name:                  "all params no namespace",
			namespace:             "",
			consumerGroupIdPrefix: "foo",
			componentPrefix:       "pipeline",
			id:                    "id",
			expected:              "foo-pipeline-id",
		},
		{
			name:                  "no consumer group prefix no namespace",
			namespace:             "",
			consumerGroupIdPrefix: "",
			componentPrefix:       "pipeline",
			id:                    "id",
			expected:              "pipeline-id",
		},
		{
			name:                  "all params",
			namespace:             "default",
			consumerGroupIdPrefix: "foo",
			componentPrefix:       "pipeline",
			id:                    "id",
			expected:              "foo-default-pipeline-id",
		},
		{
			name:                  "no consumer group prefix",
			namespace:             "default",
			consumerGroupIdPrefix: "",
			componentPrefix:       "pipeline",
			id:                    "id",
			expected:              "default-pipeline-id",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g.Expect(
				GetKafkaConsumerName(
					test.namespace, test.consumerGroupIdPrefix, test.componentPrefix, test.id),
			).To(Equal(
				test.expected),
			)
		})
	}
}

func TestParseSysLogLevel(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name     string
		level    string
		expected int
		err      bool
	}
	tests := []test{
		{
			name:     "debug",
			level:    "debug",
			expected: 7,
			err:      false,
		},
		{
			name:     "info",
			level:    "info",
			expected: 6,
			err:      false,
		},
		{
			name:     "warn",
			level:    "warn",
			expected: 4,
			err:      false,
		},
		{
			name:     "error",
			level:    "error",
			expected: 3,
			err:      false,
		},
		{
			name:     "invalid",
			level:    "invalid",
			expected: 0,
			err:      true,
		},
		{
			name:     "panic",
			level:    "panic",
			expected: 0,
			err:      false,
		},
		{
			name:     "warning",
			level:    "warning",
			expected: 4,
			err:      false,
		},
		{
			name:     "warn",
			level:    "warn",
			expected: 4,
			err:      false,
		},
		{
			name:     "alert",
			level:    "alert",
			expected: 1,
			err:      false,
		},
		{
			name:     "crit",
			level:    "crit",
			expected: 2,
			err:      false,
		},
		{
			name:     "emerg",
			level:    "emerg",
			expected: 0,
			err:      false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := parseSysLogLevel(test.level)
			if test.err {
				g.Expect(err).To(HaveOccurred())
			} else {
				g.Expect(int(result)).To(Equal(test.expected))
			}
		})
	}
}
