/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package tls

import (
	"fmt"
	"os"
)

const (
	EnvSecurityProtocolSuffix = "_SECURITY_PROTOCOL"
	SecurityProtocolSSL       = "SSL"
	SecurityProtocolPlaintxt  = "PLAINTEXT"
	SecurityProtocolSASLSSL   = "SASL_SSL"

	EnvSASLMechanismSuffix   = "_SASL_MECHANISM"
	SASLMechanismSCRAMSHA512 = "SCRAM-SHA-512"
	SASLMechanismSCRAMSHA256 = "SCRAM-SHA-256"
	SASLMechanismOAUTHBEARER = "OAUTHBEARER"
	SASLMechanismPlain       = "PLAIN"

	EnvEndpointIdentificationMechanismSuffix = "_TLS_ENDPOINT_IDENTIFICATION_ALGORITHM"
	EndpointIdentificationMechanismNone      = "none"
	EndpointIdentificationMechanismHTTPS     = "https"

	EnvSecurityPrefixControlPlane       = "CONTROL_PLANE"
	EnvSecurityPrefixControlPlaneServer = "CONTROL_PLANE_SERVER"
	EnvSecurityPrefixControlPlaneClient = "CONTROL_PLANE_CLIENT"

	EnvSecurityPrefixKafka       = "KAFKA"
	EnvSecurityPrefixKafkaServer = "KAFKA_SERVER"
	EnvSecurityPrefixKafkaClient = "KAFKA_CLIENT"

	EnvSecurityPrefixEnvoy                 = "ENVOY"
	EnvSecurityPrefixEnvoyUpstreamServer   = "ENVOY_UPSTREAM_SERVER"
	EnvSecurityPrefixEnvoyUpstreamClient   = "ENVOY_UPSTREAM_CLIENT"
	EnvSecurityPrefixEnvoyDownstreamServer = "ENVOY_DOWNSTREAM_SERVER"
	EnvSecurityPrefixEnvoyDownstreamClient = "ENVOY_DOWNSTREAM_CLIENT"
	EnvSecurityDownstreamClientMTLS        = "ENVOY_DOWNSTREAM_CLIENT_MTLS"
)

func GetSecurityProtocolFromEnv(prefix string) string {
	val, ok := os.LookupEnv(fmt.Sprintf("%s%s", prefix, EnvSecurityProtocolSuffix))
	if !ok {
		val = SecurityProtocolPlaintxt
	}
	return val
}

func GetSASLMechanismFromEnv(prefix string) string {
	val, ok := os.LookupEnv(fmt.Sprintf("%s%s", prefix, EnvSASLMechanismSuffix))
	if !ok {
		val = SASLMechanismPlain
	}
	return val
}

func GetEndpointIdentificationMechanismFromEnv(prefix string) string {
	val, ok := os.LookupEnv(fmt.Sprintf("%s%s", prefix, EnvEndpointIdentificationMechanismSuffix))
	if !ok {
		val = EndpointIdentificationMechanismHTTPS
	} else if val == "" {
		val = EndpointIdentificationMechanismNone
	}
	return val
}
