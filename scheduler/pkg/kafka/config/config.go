package config

import (
	"bytes"
	"encoding/json"
	"io/ioutil"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type KafkaConfig struct {
	BootstrapServers string          `json:"bootstrap.servers,omitempty"`
	Consumer         kafka.ConfigMap `json:"consumer,omitempty"`
	Producer         kafka.ConfigMap `json:"producer,omitempty"`
	Streams          kafka.ConfigMap `json:"streams,omitempty"`
}

const (
	KafkaBootstrapServers = "bootstrap.servers"
)

func CloneKafkaConfigMap(m kafka.ConfigMap) kafka.ConfigMap {
	m2 := make(kafka.ConfigMap)
	for k, v := range m {
		m2[k] = v
	}
	return m2
}

func NewKafkaConfig(path string) (*KafkaConfig, error) {
	data, err := ioutil.ReadFile(path)
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
