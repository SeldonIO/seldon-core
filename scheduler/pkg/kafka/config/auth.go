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
	"fmt"

	"github.com/confluentinc/confluent-kafka-go/kafka"

	"github.com/seldonio/seldon-core/components/tls/v2/pkg/password"
	"github.com/seldonio/seldon-core/components/tls/v2/pkg/tls"
	"github.com/seldonio/seldon-core/components/tls/v2/pkg/util"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/kafka/config/oauth"
)

const (
	EnvKafkaClientPrefix         = "KAFKA_CLIENT"
	EnvKafkaBrokerPrefix         = "KAFKA_BROKER"
	EnvSASLUsernameSuffix        = "_SASL_USERNAME"
	EnvPasswordLocationSuffix    = "_SASL_PASSWORD_LOCATION"
	EnvOAUTHConfigLocationSuffix = "_OAUTH_CONFIG_LOCATION"
	DefaultSASLUsername          = "seldon"
)

func AddKafkaSSLOptions(config kafka.ConfigMap) error {
	protocol := tls.GetSecurityProtocolFromEnv(tls.EnvSecurityPrefixKafka)
	switch protocol {
	case tls.SecurityProtocolSSL:
		return setupTLSAuthentication(config)
	case tls.SecurityProtocolSASLSSL: // Note: we don't support SASL_PLAINTEXT
		return setupSASLSSLAuthentication(config)
	case tls.SecurityProtocolPlaintxt:
		return nil
	}
	return nil
}

func setupSASLSSLAuthentication(config kafka.ConfigMap) error {

	mechanism := tls.GetSASLMechanismFromEnv(tls.EnvSecurityPrefixKafka)

	// TODO: Remove before merge (overwrite for testing)
	_ = mechanism
	mechanism = tls.SASLMechanismOAUTHBEARER

	var err error
	switch mechanism {
	case tls.SASLMechanismPlain:
		err = configureSASLSSLSCRAM(mechanism, config)
	case tls.SASLMechanismSCRAMSHA256, tls.SASLMechanismSCRAMSHA512:
		err = configureSASLSSLSCRAM(mechanism, config)
	case tls.SASLMechanismOAUTHBEARER:
		err = configureSASLSSLOAUTHBEARER(mechanism, config)
	default:
		err = fmt.Errorf("Provided SASL mechanism %s is not supported", mechanism)
	}

	return err
}

func configureSASLSSLSCRAM(mechanism string, config kafka.ConfigMap) error {
	// Set the SASL mechanism
	config["security.protocol"] = "SASL_SSL"
	config["sasl.mechanism"] = mechanism

	// Set the SASL username and password
	ps, err := password.NewPasswordStore(
		password.Prefix(EnvKafkaClientPrefix),
		password.LocationSuffix(EnvPasswordLocationSuffix),
	)
	if err != nil {
		return err
	}
	username, found := util.GetEnv(EnvKafkaClientPrefix, EnvSASLUsernameSuffix)
	if !found {
		username = DefaultSASLUsername
	}
	config["sasl.username"] = username
	config["sasl.password"] = ps.GetPassword()

	// Set the TLS Certificate
	cs, err := tls.NewCertificateStore(
		tls.ValidationOnly(true),
		tls.ValidationPrefix(EnvKafkaBrokerPrefix),
	)
	if err != nil {
		return err
	}
	caCert := cs.GetValidationCertificate()
	// issue is that ca.pem does not work with multiple certificates defined
	// see https://github.com/confluentinc/confluent-kafka-go/issues/827 (Fixed needs updating and testing in our code)

	config["ssl.ca.location"] = caCert.CaPath

	endpointMechanism := tls.GetEndpointIdentificationMechanismFromEnv(tls.EnvSecurityPrefixKafkaClient)
	config["ssl.endpoint.identification.algorithm"] = endpointMechanism
	return nil
}

func configureSASLSSLOAUTHBEARER(mechanism string, config kafka.ConfigMap) error {
	// Set the SASL mechanism
	config["security.protocol"] = "SASL_SSL"
	config["sasl.mechanism"] = "OAUTHBEARER"

	// Set OAUTH Configuration
	oauthStore, err := oauth.NewOAUTHStore(
		oauth.Prefix(EnvKafkaClientPrefix),
		oauth.LocationSuffix(EnvOAUTHConfigLocationSuffix),
	)
	if err != nil {
		return err
	}

	oauthConfig := oauthStore.GetOAUTHConfig()

	config["sasl.oauthbearer.method"] = oauthConfig.Method
	config["sasl.oauthbearer.client.id"] = oauthConfig.ClientID
	config["sasl.oauthbearer.client.secret"] = oauthConfig.ClientSecret
	config["sasl.oauthbearer.token.endpoint.url"] = oauthConfig.TokenEndpointURL
	config["sasl.oauthbearer.extensions"] = oauthConfig.Extensions

	// Set the TLS Certificate
	cs, err := tls.NewCertificateStore(
		tls.ValidationOnly(true),
		tls.ValidationPrefix(EnvKafkaBrokerPrefix),
	)
	if err != nil {
		return err
	}
	caCert := cs.GetValidationCertificate()
	// issue is that ca.pem does not work with multiple certificates defined
	// see https://github.com/confluentinc/confluent-kafka-go/issues/827 (Fixed needs updating and testing in our code)

	config["ssl.ca.location"] = caCert.CaPath

	endpointMechanism := tls.GetEndpointIdentificationMechanismFromEnv(tls.EnvSecurityPrefixKafkaClient)
	config["ssl.endpoint.identification.algorithm"] = endpointMechanism

	return nil
}

func setupTLSAuthentication(config kafka.ConfigMap) error {
	var cs *tls.CertificateStore
	cs, err := tls.NewCertificateStore(
		tls.Prefix(EnvKafkaClientPrefix),
		tls.ValidationPrefix(EnvKafkaBrokerPrefix),
	)
	// In case of error (missing client certificates) we fallback to one-sided TLS
	// In case mTLS is required for auth brokers will reject the consumer connection
	if err != nil {
		cs, err = tls.NewCertificateStore(
			tls.ValidationOnly(true),
			tls.ValidationPrefix(EnvKafkaBrokerPrefix),
		)
		if err != nil {
			return err
		}
	}

	cert := cs.GetCertificate()
	caCert := cs.GetValidationCertificate()
	config["security.protocol"] = "SSL"

	// issue is that ca.pem does not work with multiple certificates defined
	// see https://github.com/confluentinc/confluent-kafka-go/issues/827 (Fixed needs updating and test in our code)
	if caCert != nil {
		config["ssl.ca.location"] = caCert.CaPath
	} else if cert != nil {
		config["ssl.ca.location"] = cert.CaPath
	}

	if cert != nil {
		config["ssl.key.location"] = cert.KeyPath
		config["ssl.certificate.location"] = cert.CrtPath
	}

	endpointMechanism := tls.GetEndpointIdentificationMechanismFromEnv(tls.EnvSecurityPrefixKafkaClient)
	config["ssl.endpoint.identification.algorithm"] = endpointMechanism
	return nil
}
