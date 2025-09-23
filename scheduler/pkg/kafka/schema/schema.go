/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package schema

import (
	"fmt"
	"os"

	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	log "github.com/sirupsen/logrus"
)

const (
	EnvSchemaRegistryConfigPath = "SCHEMA_REGISTRY_CONFIG_PATH"
	FileName                    = "confluent-schema.yaml"
)

type config struct {
	URL      string `koanf:"schemaRegistry.client.URL"`
	Username string `koanf:"schemaRegistry.client.username"`
	Password string `koanf:"schemaRegistry.client.password"`
}

func NewSchemaRegistryClient(log *log.Logger, k *koanf.Koanf) (schemaregistry.Client, error) {
	logger := log.WithField("func", "NewSchemaRegistryClient")
	schemaConfigPath := os.Getenv(EnvSchemaRegistryConfigPath)
	if schemaConfigPath == "" {
		return nil, nil
	}

	if err := k.Load(file.Provider(schemaConfigPath+"/."+FileName), yaml.Parser()); err != nil {
		return nil, fmt.Errorf("error loading schema registry config: %v", err)
	}

	var cfg config

	err := k.UnmarshalWithConf("", &cfg, koanf.UnmarshalConf{Tag: "koanf", FlatPaths: true})
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling config: %v", err)
	}

	if cfg.URL == "" {
		return nil, fmt.Errorf("configuration url is required for schema registry client")
	}

	conf := schemaregistry.NewConfigWithBasicAuthentication(cfg.URL, cfg.Username, cfg.Password)

	srClient, err := schemaregistry.NewClient(conf)
	if err != nil {
		return nil, fmt.Errorf("error creating schema registry client: %v", err)
	}

	_, err = srClient.GetAllSubjects()
	if err != nil {
		logger.Warnf("unable to get ping schema registry: %v", err)
	}

	srClient.Config()

	return srClient, nil
}

// TrimSchemaID trims the magic byte, schema id and message index of a kafka message that was sent with a schema ID
func TrimSchemaID(payload []byte) []byte {
	// If it's Schema Registry format (magic byte 0x0)
	if len(payload) < 6 {
		return payload
	}
	if payload[0] == 0x0 {
		// Skip magic byte (1) + schema ID (4) + message index (1) = 6 bytes

		payload = payload[6:]
	}
	return payload
}
