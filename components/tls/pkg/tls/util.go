package tls

import (
	"fmt"
	"os"
)

const (
	EnvSecurityProtocolSuffix = "_SECURITY_PROTOCOL"
	SecurityProtocolSSL       = "SSL"
	SecurityProtocolPlaintxt  = "PLAINTEXT"

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
)

func GetSecurityProtocolFromEnv(prefix string) string {
	val, ok := os.LookupEnv(fmt.Sprintf("%s%s", prefix, EnvSecurityProtocolSuffix))
	if !ok {
		val = SecurityProtocolPlaintxt
	}
	return val
}
