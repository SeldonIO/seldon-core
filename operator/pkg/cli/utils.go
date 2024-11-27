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
	"strconv"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"golang.org/x/exp/rand"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	kafka_config "github.com/seldonio/seldon-core/components/kafka/v2/pkg/config"
	config_tls "github.com/seldonio/seldon-core/components/tls/v2/pkg/config"
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

func getKafkaConsumerConfig(kafkaBrokerIsSet bool, kafkaBroker string, config *SeldonCLIConfig, kafkaConfigPath string) (kafka.ConfigMap, string, string, error) {
	var kafkaConfigMap *kafka_config.KafkaConfig
	if kafkaConfigPath != "" {
		var err error
		kafkaConfigMap, err = kafka_config.NewKafkaConfig(kafkaConfigPath)
		if err != nil {
			fmt.Printf("Failed to load Kafka config with error: %s\n", err.Error())
			return nil, "", "", err
		}
	}

	maxMessageSize := DefaultMaxMessageSize
	namespace := DefaultNamespace
	topicPrefix := SeldonDefaultTopicPrefix

	// Overwrite broker if set in config
	if kafkaConfigMap != nil && kafkaConfigMap.BootstrapServers != "" {
		kafkaBroker = kafkaConfigMap.BootstrapServers
	}
	if !kafkaBrokerIsSet && config.Kafka != nil && config.Kafka.Bootstrap != "" {
		kafkaBroker = config.Kafka.Bootstrap
	}

	// topic prefix
	if kafkaConfigMap != nil && kafkaConfigMap.TopicPrefix != "" {
		topicPrefix = kafkaConfigMap.TopicPrefix
	}
	if config.Kafka != nil {
		if config.Kafka.Namespace != "" {
			namespace = config.Kafka.Namespace
		}
		if config.Kafka.TopicPrefix != "" {
			topicPrefix = config.Kafka.TopicPrefix
		}
	}

	// message size
	if kafkaConfigMap != nil && kafkaConfigMap.Consumer != nil {
		var err error
		maxMessageSizeValue, err := kafkaConfigMap.Consumer.Get("message.max.bytes", strconv.Itoa(DefaultMaxMessageSize))
		if err == nil {
			if maxMessageSizeStr, ok := maxMessageSizeValue.(string); ok {
				var errConv error
				maxMessageSize, errConv = strconv.Atoi(maxMessageSizeStr)
				if errConv != nil {
					fmt.Printf("Failed to convert max message size to int with error: %s\n", errConv.Error())
					return nil, "", "", errConv
				}
			} else {
				fmt.Printf("Failed to assert max message size to int\n")
				return nil, "", "", err
			}
		} else {
			fmt.Printf("Failed to get max message size from config with error: %s\n", err.Error())
			return nil, "", "", err
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
	err := config_tls.AddKafkaSSLOptions(consumerConfig)
	if err != nil {
		fmt.Printf("Failed to add Kafka SSL options with error: %s\n", err.Error())
		return nil, "", "", err
	}

	// todo: use max message size from configMap
	consumerConfig["message.max.bytes"] = maxMessageSize

	fmt.Printf("Using consumer config %v\n", consumerConfig)

	return consumerConfig, namespace, topicPrefix, nil
}
