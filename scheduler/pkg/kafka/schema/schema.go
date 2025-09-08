/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package schema

import (
	"os"

	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry"
	log "github.com/sirupsen/logrus"
)

const (
	EnvURL      = "SCHEMA_REGISTRY_URL"
	EnvUsername = "SCHEMA_REGISTRY_USERNAME"
	EnvPassword = "SCHEMA_REGISTRY_PASSWORD"
)

func NewSchemaRegistryClient(log *log.Logger) schemaregistry.Client {
	url := os.Getenv(EnvURL)
	username := os.Getenv(EnvUsername)
	password := os.Getenv(EnvPassword)
	logger := log.WithField("func", "setup")

	if url == "" {
		return nil
	}

	var conf *schemaregistry.Config

	conf = schemaregistry.NewConfigWithBasicAuthentication(url, username, password)

	srClient, err := schemaregistry.NewClient(conf)
	if err != nil {
		logger.Warnf("unable to create schema registry client: %v", err)
		return nil
	}

	_, err = srClient.GetAllSubjects()
	if err != nil {
		logger.Warnf("unable to get ping schema registry: %v", err)
	}

	srClient.Config()

	logger.Info("schema registry client created")
	return srClient
}

// TrimSchemaID trims the magic byte, schema id and message index of a kafka message that was sent with a schema ID
func TrimSchemaID(payload []byte) []byte {
	// If it's Schema Registry format (magic byte 0x0)
	if len(payload) < 6 {
		return payload
	}
	if payload[0] == 0x0 {
		// Skip magic byte (1) + schema ID (4) + message index (0) = 6 bytes

		payload = payload[6:]
	}
	return payload
}
