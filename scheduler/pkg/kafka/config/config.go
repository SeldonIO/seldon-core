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
	"bytes"
	"encoding/json"
	"os"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type KafkaConfig struct {
	BootstrapServers string          `json:"bootstrap.servers,omitempty"`
	Debug            string          `json:"debug,omitempty"`
	Consumer         kafka.ConfigMap `json:"consumer,omitempty"`
	Producer         kafka.ConfigMap `json:"producer,omitempty"`
	Streams          kafka.ConfigMap `json:"streams,omitempty"`
	TopicPrefix      string          `json:"topicPrefix,omitempty"`
}

const (
	KafkaBootstrapServers = "bootstrap.servers"
	KafkaDebug            = "debug"
)

// Based on config options defined for librdkafka:
// https://github.com/confluentinc/librdkafka/blob/master/CONFIGURATION.md
var secretConfigFields = map[string]struct{}{
	"ssl.key.password":               struct{}{},
	"ssl.key.pem":                    struct{}{},
	"ssl_key":                        struct{}{},
	"ssl.keystore.password":          struct{}{},
	"sasl.username":                  struct{}{},
	"sasl.password":                  struct{}{},
	"sasl.oauthbearer.client.secret": struct{}{},
}

func CloneKafkaConfigMap(m kafka.ConfigMap) kafka.ConfigMap {
	m2 := make(kafka.ConfigMap)
	for k, v := range m {
		m2[k] = v
	}
	return m2
}

func NewKafkaConfig(path string) (*KafkaConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	kc := KafkaConfig{}
	d := json.NewDecoder(bytes.NewReader(data))
	d.DisallowUnknownFields() // So we fail if not exactly as required in schema
	d.UseNumber()             // So numbers are not turned into float64
	err = d.Decode(&kc)
	if err != nil {
		return nil, err
	}

	kc.Consumer, err = convertConfigMap(kc.Consumer)
	if err != nil {
		return nil, err
	}
	kc.Producer, err = convertConfigMap(kc.Producer)
	if err != nil {
		return nil, err
	}
	kc.Streams, err = convertConfigMap(kc.Streams)
	if err != nil {
		return nil, err
	}

	// Add bootstrap servers to each config if not present
	if _, ok := kc.Consumer[KafkaBootstrapServers]; !ok {
		kc.Consumer[KafkaBootstrapServers] = kc.BootstrapServers
	}
	if _, ok := kc.Producer[KafkaBootstrapServers]; !ok {
		kc.Producer[KafkaBootstrapServers] = kc.BootstrapServers
	}
	if _, ok := kc.Streams[KafkaBootstrapServers]; !ok {
		kc.Streams[KafkaBootstrapServers] = kc.BootstrapServers
	}

	if kc.Debug != "" {
		kc.Consumer[KafkaDebug] = kc.Debug
		kc.Producer[KafkaDebug] = kc.Debug
	}
	return &kc, nil
}

// All number types are treated as ints
// https://github.com/confluentinc/confluent-kafka-go/blob/e01dd295220b5bf55f3fbfabdf8cc6d3f0ae185f/kafka/config.go#L80-L99
func convertConfigMap(cm kafka.ConfigMap) (kafka.ConfigMap, error) {
	r := make(kafka.ConfigMap)
	for k, v := range cm {
		switch x := v.(type) {
		case json.Number:
			i, err := x.Int64()
			if err != nil {
				return nil, err
			}
			r[k] = int(i) //Assumes will not be truncated
		default:
			r[k] = v
		}
	}
	return r, nil
}

// Allow us to test if we have a valid Kafka confguration. For unit tests we can have no bootstrap server
// See usages of this method.
// TODO in future allow testing to run without this check
func (kc KafkaConfig) HasKafkaBootstrapServer() bool {
	bs := kc.Consumer[KafkaBootstrapServers]
	return bs != nil && bs != ""
}

func WithoutSecrets(c kafka.ConfigMap) kafka.ConfigMap {
	safe := make(kafka.ConfigMap)

	for k, v := range c {
		_, isSecret := secretConfigFields[k]
		if isSecret {
			safe[k] = "***"
		} else {
			safe[k] = v
		}
	}

	return safe
}
