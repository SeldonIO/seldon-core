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

package tls

import (
	"fmt"
	"os"
)

const (
	EnvSecurityProtocolSuffix = "_SECURITY_PROTOCOL"
	SecurityProtocolTLS       = "TLS"
	SecurityProtocolSSL       = "SSL"
	SecurityProtocolPlaintxt  = "PLAINTEXT"
	SecurityProtocolSASLSSL   = "SASL_SSL"

	EnvSASLMechanismSuffix   = "_SASL_MECHANISM"
	SASLMechanismSCRAMSHA512 = "SCRAM-SHA-512"
	SASLMechanismSCRAMSHA256 = "SCRAM-SHA-256"
	SASLMechanismPlain       = "PLAIN"

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
