/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	config_tls "github.com/seldonio/seldon-core/components/tls/v2/pkg/config"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"golang.org/x/exp/rand"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func printPrettyJson(data []byte) {
	prettyJson, err := prettyJson(data)
	if err == nil {
		fmt.Printf("%s\n", prettyJson)
	}
}

func prettyJson(data []byte) (string, error) {
	var prettyJSON bytes.Buffer
	err := json.Indent(&prettyJSON, data, "", "\t")
	if err != nil {
		return "", err
	}
	return prettyJSON.String(), nil
}

func PrintProto(msg proto.Message) {
	resJson, err := protojson.Marshal(msg)
	if err != nil {
		fmt.Printf("Failed to print proto: %s", err.Error())
	} else {
		fmt.Printf("%s\n", string(resJson))
	}
}

func getKafkaConsumerConfig(kafkaBrokerIsSet bool, kafkaBroker string, config *SeldonCLIConfig) (kafka.ConfigMap, string, string) {
	maxMessageSize := DefaultMaxMessageSize
	namespace := DefaultNamespace
	topicPrefix := SeldonDefaultTopicPrefix

	// Overwrite broker if set in config
	if !kafkaBrokerIsSet && config.Kafka != nil && config.Kafka.Bootstrap != "" {
		kafkaBroker = config.Kafka.Bootstrap
	}

	if config.Kafka != nil {
		if config.Kafka.Namespace != "" {
			namespace = config.Kafka.Namespace
		}
		if config.Kafka.TopicPrefix != "" {
			topicPrefix = config.Kafka.TopicPrefix
		}
	}

	s1 := rand.NewSource(uint64(time.Now().UnixNano()))
	r1 := rand.New(s1)

	consumerConfig := kafka.ConfigMap{
		"bootstrap.servers": kafkaBroker,
		// TODO: use ConsumerGroupIdPrefix from configMap
		"group.id":          fmt.Sprintf("seldon-cli-%d", r1.Int()),
		"auto.offset.reset": "earliest",
	}
	config_tls.AddKafkaSSLOptions(consumerConfig)

	// todo: use max message size from configMap
	consumerConfig["message.max.bytes"] = maxMessageSize

	fmt.Printf("Using consumer config %v\n", consumerConfig)

	return consumerConfig, namespace, topicPrefix
}
