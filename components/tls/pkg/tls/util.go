package tls

import (
	"fmt"
	"os"
)

const (
	EnvSecurityProtocolSuffix     = "_SECURITY_PROTOCOL"
	SecurityProtocolSSL           = "SSL"
	SecurityProtocolPlaintxt      = "PLAINTEXT"
	EnvSecurityPrefixControlPlane = "CONTROL_PLANE"
	EnvSecurityPrefixDataPlane    = "DATA_PLANE"
	EnvSecurityPrefixKafka        = "KAFKA"
)

func GetSecurityProtocolFromEnv(prefix string) string {
	val, ok := os.LookupEnv(fmt.Sprintf("%s%s", prefix, EnvSecurityProtocolSuffix))
	if !ok {
		val = SecurityProtocolPlaintxt
	}
	return val
}
