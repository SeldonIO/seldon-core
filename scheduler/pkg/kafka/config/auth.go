/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package config

import (
	"fmt"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"

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
	EnvOAuthConfigLocationSuffix = "_OAUTH_CONFIG_LOCATION"
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

	var err error
	switch mechanism {
	// PLAIN and SCRAM are password-based mechanisms handled in the same way
	case tls.SASLMechanismPlain, tls.SASLMechanismSCRAMSHA256, tls.SASLMechanismSCRAMSHA512:
		err = withPasswordAuth(mechanism, config)
	case tls.SASLMechanismOAUTHBEARER:
		err = withOAuth(config)
	default:
		err = fmt.Errorf("Provided SASL mechanism %s is not supported", mechanism)
	}

	return err
}

func withPasswordAuth(mechanism string, config kafka.ConfigMap) error {
	// Set the SASL mechanism
	config["security.protocol"] = tls.SecurityProtocolSASLSSL
	config["sasl.mechanism"] = mechanism

	// Set the SASL username and password
	passwordStore, err := password.NewPasswordStore(
		password.PasswordStoreOptions{
			Prefix:         EnvKafkaClientPrefix,
			LocationSuffix: EnvPasswordLocationSuffix,
		},
	)
	if err != nil {
		return err
	}

	username, found := util.GetNonEmptyEnv(EnvKafkaClientPrefix, EnvSASLUsernameSuffix)
	if !found {
		username = DefaultSASLUsername
	}

	config["sasl.username"] = username
	config["sasl.password"] = passwordStore.GetPassword()

	// Set the TLS Certificate
	cs, err := tls.NewCertificateStore(
		tls.ValidationOnly(true),
		tls.ValidationPrefix(EnvKafkaBrokerPrefix),
	)
	if err != nil {
		return err
	}

	// issue is that ca.pem does not work with multiple certificates defined
	// see https://github.com/confluentinc/confluent-kafka-go/issues/827 (Fixed needs updating and testing in our code)
	caCert := cs.GetValidationCertificate()
	config["ssl.ca.location"] = caCert.CaPath

	endpointMechanism := tls.GetEndpointIdentificationMechanismFromEnv(tls.EnvSecurityPrefixKafkaClient)
	config["ssl.endpoint.identification.algorithm"] = endpointMechanism

	return nil
}

func withOAuth(config kafka.ConfigMap) error {
	// Set the SASL mechanism
	config["security.protocol"] = tls.SecurityProtocolSASLSSL
	config["sasl.mechanism"] = tls.SASLMechanismOAUTHBEARER

	// Set OAuth Configuration
	oauthStore, err := oauth.NewOAuthStore(
		oauth.OAuthStoreOptions{
			Prefix:         EnvKafkaClientPrefix,
			LocationSuffix: EnvOAuthConfigLocationSuffix,
		},
	)
	if err != nil {
		return err
	}

	oauthConfig := oauthStore.GetOAuthConfig()

	config["sasl.oauthbearer.method"] = oauthConfig.Method
	config["sasl.oauthbearer.client.id"] = oauthConfig.ClientID
	config["sasl.oauthbearer.client.secret"] = oauthConfig.ClientSecret
	config["sasl.oauthbearer.scope"] = oauthConfig.Scope
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
	config["security.protocol"] = tls.SecurityProtocolSSL

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
