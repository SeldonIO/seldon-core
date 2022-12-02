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
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/seldonio/seldon-core/components/tls/v2/pkg/tls"
)

const (
	EnvKafkaClientPrefix = "KAFKA_CLIENT"
	EnvKafkaBrokerPrefix = "KAFKA_BROKER"
)

func AddKafkaSSLOptions(config kafka.ConfigMap) error {
	var cs *tls.CertificateStore
	var err error
	protocol := tls.GetSecurityProtocolFromEnv(tls.EnvSecurityPrefixKafka)
	if protocol == tls.SecurityProtocolSSL {
		cs, err = tls.NewCertificateStore(tls.Prefix(EnvKafkaClientPrefix),
			tls.ValidationPrefix(EnvKafkaBrokerPrefix))
		if err != nil {
			return err
		}
	}
	if cs != nil {
		cert := cs.GetCertificate()
		caCert := cs.GetValidationCertificate()
		config["security.protocol"] = "ssl"
		// issue is that ca.pem does not work with multiple certificiates defined
		// see https://github.com/confluentinc/confluent-kafka-go/issues/827
		if caCert != nil {
			config["ssl.ca.location"] = caCert.CaPath
		} else {
			config["ssl.ca.location"] = cert.CaPath
		}
		config["ssl.key.location"] = cert.KeyPath
		config["ssl.certificate.location"] = cert.CrtPath
		config["ssl.endpoint.identification.algorithm"] = "none"
	}
	return nil
}
