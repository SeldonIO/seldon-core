package util

import (
	"strings"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

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

func GetKafkaProducerConfig(broker string) *kafka.ConfigMap {
	producerConfig := kafka.ConfigMap{
		"bootstrap.servers":   broker,
		"go.delivery.reports": false, // Need this otherwise will get memory leak
	}

	kafkaSecurityProtocol := GetKafkaSecurityProtocol()

	if kafkaSecurityProtocol == "SSL" || kafkaSecurityProtocol == "SASL_SSL" {
		sslConfig := GetKafkaSSLConfig()
		producerConfig["security.protocol"] = kafkaSecurityProtocol
		if sslConfig.CACertFile != "" && sslConfig.ClientCertFile != "" {
			producerConfig["ssl.ca.location"] = sslConfig.CACertFile
			producerConfig["ssl.key.location"] = sslConfig.ClientKeyFile
			producerConfig["ssl.certificate.location"] = sslConfig.ClientCertFile
		}
		if sslConfig.CACert != "" && sslConfig.ClientCert != "" {
			producerConfig["ssl.ca.pem"] = sslConfig.CACert
			producerConfig["ssl.key.pem"] = sslConfig.ClientKey
			producerConfig["ssl.certificate.pem"] = sslConfig.ClientCert
		}
		producerConfig["ssl.key.password"] = sslConfig.ClientKeyPass // Key password, if any
	}

	if kafkaSecurityProtocol == "SASL_PLAINTEXT" || kafkaSecurityProtocol == "SASL_SSL" {
		saslConfig := GetKafkaSASLConfig()
		producerConfig["sasl.mechanisms"] = saslConfig.Mechanism
		if saslConfig.UserName != "" && saslConfig.Password != "" {
			producerConfig["sasl.username"] = saslConfig.UserName
			producerConfig["sasl.password"] = saslConfig.Password
		}
	}

	return &producerConfig
}
