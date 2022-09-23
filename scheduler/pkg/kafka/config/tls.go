package config

import (
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/seldonio/seldon-core-v2/components/tls/pkg/tls"
)

const (
	EnvKafkaClientPrefix = "KAFKA_CLIENT"
	EnvKafkaBrokerPrefix = "KAFKA_BROKER"
)

func AddKafkaSSLOptions(config kafka.ConfigMap) error {
	var cs, ca *tls.CertificateStore
	var err error
	protocol := tls.GetSecurityProtocolFromEnv(tls.EnvSecurityPrefixKafka)
	if protocol == tls.SecurityProtocolSSL {
		cs, err = tls.NewCertificateStore(tls.Prefix(EnvKafkaClientPrefix))
		if err != nil {
			return err
		}
		// Allow CA of Broker to be found at separate location
		ca, err = tls.NewCertificateStore(tls.Prefix(EnvKafkaBrokerPrefix), tls.CaOnly(true))
		if err != nil {
			return err
		}
	}
	if cs != nil {
		cert := cs.GetCertificate()
		var caCert *tls.CertificateWrapper
		if ca != nil {
			caCert = ca.GetCertificate()
		}
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
