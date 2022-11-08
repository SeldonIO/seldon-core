package util

import "strings"

type KafkaSSL struct {
	ClientCert     string
	ClientKey      string
	CACert         string
	ClientCertFile string
	ClientKeyFile  string
	CACertFile     string
	ClientKeyPass  string
}

type KafkaSASL struct {
	UserName  string
	Password  string
	Mechanism string
}

func GetKafkaSecurityProtocol() string {
	return strings.ToUpper(GetEnv("KAFKA_SECURITY_PROTOCOL", ""))
}

func GetKafkaSSLConfig() *KafkaSSL {
	sslElements := KafkaSSL{
		ClientCert: GetEnv("KAFKA_SSL_CLIENT_CERT", ""),
		ClientKey:  GetEnv("KAFKA_SSL_CLIENT_KEY", ""),
		CACert:     GetEnv("KAFKA_SSL_CA_CERT", ""),
		// If we use path to files instead of string
		ClientCertFile: GetEnv("KAFKA_SSL_CLIENT_CERT_FILE", ""),
		ClientKeyFile:  GetEnv("KAFKA_SSL_CLIENT_KEY_FILE", ""),
		CACertFile:     GetEnv("KAFKA_SSL_CA_CERT_FILE", ""),
		// Optional password
		ClientKeyPass: GetEnv("KAFKA_SSL_CLIENT_KEY_PASS", ""),
	}
	return &sslElements

}

func GetKafkaSASLConfig() *KafkaSASL {
	saslElements := KafkaSASL{
		UserName:  GetEnv("KAFKA_SASL_USERNAME", ""),
		Password:  GetEnv("KAFKA_SASL_PASSWORD", ""),
		Mechanism: GetEnv("KAFKA_SASL_MECHANISM", ""),
	}
	return &saslElements
}
