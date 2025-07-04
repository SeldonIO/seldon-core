/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package cli

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	. "github.com/onsi/gomega"
)

func TestGetConsumerConfig(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name         string
		seldonConfig *SeldonCLIConfig
		brokerIsSet  bool
		broker       string
		kafkaConfig  string
		// expected
		expectedConsumerConfig kafka.ConfigMap
		expectedTopicPrefix    string
		expectedNamespace      string
	}

	// TODO: test auth settings from k8s secret
	tests := []test{
		{
			name:         "with config",
			seldonConfig: &SeldonCLIConfig{Kafka: &KafkaConfig{Bootstrap: "broker", Namespace: "namespace", TopicPrefix: "prefix"}},
			brokerIsSet:  false,
			broker:       "x",
			expectedConsumerConfig: kafka.ConfigMap{
				"bootstrap.servers": "broker",
				"auto.offset.reset": "earliest",
				"message.max.bytes": DefaultMaxMessageSize,
			},
			expectedTopicPrefix: "prefix",
			expectedNamespace:   "namespace",
		},
		{
			name:         "with broker set",
			seldonConfig: &SeldonCLIConfig{Kafka: &KafkaConfig{Bootstrap: "broker", Namespace: "namespace", TopicPrefix: "prefix"}},
			brokerIsSet:  true,
			broker:       "x",
			expectedConsumerConfig: kafka.ConfigMap{
				"bootstrap.servers": "x",
				"auto.offset.reset": "earliest",
				"message.max.bytes": DefaultMaxMessageSize,
			},
			expectedTopicPrefix: "prefix",
			expectedNamespace:   "namespace",
		},
		{
			name:         "no cli config",
			seldonConfig: &SeldonCLIConfig{},
			brokerIsSet:  true,
			broker:       "x",
			expectedConsumerConfig: kafka.ConfigMap{
				"bootstrap.servers": "x",
				"auto.offset.reset": "earliest",
				"message.max.bytes": DefaultMaxMessageSize,
			},
			expectedTopicPrefix: SeldonDefaultTopicPrefix,
			expectedNamespace:   DefaultNamespace,
		},
		{
			name:         "with config map",
			seldonConfig: &SeldonCLIConfig{},
			kafkaConfig:  `{"bootstrap.servers":"kafka:9092","consumer":{"message.max.bytes": "6000"}, "topicPrefix": "configPrefix"}`,
			expectedConsumerConfig: kafka.ConfigMap{
				"bootstrap.servers": "kafka:9092",
				"auto.offset.reset": "earliest",
				"message.max.bytes": 6000,
			},
			expectedTopicPrefix: "configPrefix",
			expectedNamespace:   DefaultNamespace,
		},
		{
			name:         "with cli and config map",
			seldonConfig: &SeldonCLIConfig{Kafka: &KafkaConfig{Bootstrap: "broker", Namespace: "namespace", TopicPrefix: "prefix"}},
			kafkaConfig:  `{"bootstrap.servers":"kafka:9092", "topicPrefix": "configPrefix"}`,
			expectedConsumerConfig: kafka.ConfigMap{
				"bootstrap.servers": "broker",
				"auto.offset.reset": "earliest",
				"message.max.bytes": DefaultMaxMessageSize,
			},
			expectedTopicPrefix: "prefix",
			expectedNamespace:   "namespace",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var configFilePath string
			if test.kafkaConfig != "" {
				configFilePath = fmt.Sprintf("%s/kafka.json", t.TempDir())
				_ = os.WriteFile(configFilePath, []byte(test.kafkaConfig), 0644)
			}
			actualConsumerConfig, ActualNamespace, ActualTopicPrefix, _ := getKafkaConsumerConfig(test.brokerIsSet, test.broker, test.seldonConfig, configFilePath)
			// check consumer group.id
			g.Expect(strings.Contains(fmt.Sprintf("%s", actualConsumerConfig["group.id"]), "seldon-cli")).To(BeTrue())
			for key, value := range test.expectedConsumerConfig {
				g.Expect(actualConsumerConfig[key]).To(Equal(value))
			}
			g.Expect(ActualNamespace).To(Equal(test.expectedNamespace))
			g.Expect(ActualTopicPrefix).To(Equal(test.expectedTopicPrefix))
		})
	}
}
