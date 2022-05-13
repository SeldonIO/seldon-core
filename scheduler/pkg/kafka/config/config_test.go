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
