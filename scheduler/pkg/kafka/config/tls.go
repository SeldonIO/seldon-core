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
	"github.com/seldonio/seldon-core/components/tls/v2/pkg/password"
	"github.com/seldonio/seldon-core/components/tls/v2/pkg/tls"
	"github.com/seldonio/seldon-core/components/tls/v2/pkg/util"
)

const (
	EnvKafkaClientPrefix      = "KAFKA_CLIENT"
	EnvKafkaBrokerPrefix      = "KAFKA_BROKER"
	EnvSASLUsernameSuffix     = "_SASL_USERNAME"
	EnvPasswordLocationSuffix = "_SASL_PASSWORD_LOCATION"
	DefaultSASLUsername       = "seldon"
)

func AddKafkaSSLOptions(config kafka.ConfigMap) error {
	protocol := tls.GetSecurityProtocolFromEnv(tls.EnvSecurityPrefixKafka)
	switch protocol {
	case tls.SecurityProtocolSSL:
		return setupTLSAuthentication(config)
	case tls.SecurityProtocolSASLSSL:
		return setupSASLSSLAuthentication(config)
	case tls.SecurityProtocolPlaintxt:
		return nil
	}
	return nil
}

func setupSASLSSLAuthentication(config kafka.ConfigMap) error {
	cs, err := tls.NewCertificateStore(tls.ValidationOnly(true), tls.ValidationPrefix(EnvKafkaBrokerPrefix))
	if err != nil {
		return err
	}
	caCert := cs.GetValidationCertificate()
	config["security.protocol"] = "SASL_SSL"
	config["sasl.mechanism"] = "SCRAM-SHA-512"
	ps, err := password.NewPasswordStore(password.Prefix(EnvKafkaClientPrefix),
		password.LocationSuffix(EnvPasswordLocationSuffix))
	if err != nil {
		return err
	}
	username, found := util.GetEnv(EnvKafkaClientPrefix, EnvSASLUsernameSuffix)
	if !found {
		username = DefaultSASLUsername
	}
	config["sasl.username"] = username
	config["sasl.password"] = ps.GetPassword()
	// issue is that ca.pem does not work with multiple certificiates defined
	// see https://github.com/confluentinc/confluent-kafka-go/issues/827 (Fixed but not yet released)
	config["ssl.ca.location"] = caCert.CaPath
	return nil
}

func setupTLSAuthentication(config kafka.ConfigMap) error {
	cs, err := tls.NewCertificateStore(tls.Prefix(EnvKafkaClientPrefix),
		tls.ValidationPrefix(EnvKafkaBrokerPrefix))
	if err != nil {
		return err
	}
	cert := cs.GetCertificate()
	caCert := cs.GetValidationCertificate()
	config["security.protocol"] = "ssl"
	// issue is that ca.pem does not work with multiple certificiates defined
	// see https://github.com/confluentinc/confluent-kafka-go/issues/827 (Fixed but not yet released)
	if caCert != nil {
		config["ssl.ca.location"] = caCert.CaPath
	} else {
		config["ssl.ca.location"] = cert.CaPath
	}
	config["ssl.key.location"] = cert.KeyPath
	config["ssl.certificate.location"] = cert.CrtPath
	config["ssl.endpoint.identification.algorithm"] = "none"
	return nil
}
