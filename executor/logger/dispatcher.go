package logger

import (
	"os"

	"github.com/go-logr/logr"
	"github.com/seldonio/seldon-core/executor/api/util"
)

const (
	ENV_LOGGER_KAFKA_BROKER = "LOGGER_KAFKA_BROKER"
	ENV_LOGGER_KAFKA_TOPIC  = "LOGGER_KAFKA_TOPIC"
)

var WorkerQueue chan chan LogRequest

func getSslElements() *SslKakfa {
	sslElements := SslKakfa{
		kafkaSslClientCertFile: util.GetEnv("KAFKA_SSL_CLIENT_CERT_FILE", ""),
		kafkaSslClientKeyFile:  util.GetEnv("KAFKA_SSL_CLIENT_KEY_FILE", ""),
		kafkaSslCACertFile:     util.GetEnv("KAFKA_SSL_CA_CERT_FILE", ""),
		kafkaSecurityProtocol:  util.GetEnv("KAFKA_SECURITY_PROTOCOL", ""),
		kafkaSslClientKeyPass:  util.GetEnv("KAFKA_SSL_CLIENT_KEY_PASS", ""),
	}
	return &sslElements
}
func StartDispatcher(nworkers int, logBufferSize int, writeTimeoutMs int, log logr.Logger, sdepName string, namespace string, predictorName string, kafkaBroker string, kafkaTopic string) error {
	if kafkaBroker == "" {
		kafkaBroker = os.Getenv(ENV_LOGGER_KAFKA_BROKER)
	}
	if kafkaBroker != "" {
		if kafkaTopic == "" {
			kafkaTopic = os.Getenv(ENV_LOGGER_KAFKA_TOPIC)
		}
		if kafkaTopic == "" {
			kafkaTopic = "seldon"
		}
	}
	sslKakfa := getSslElements()
	log.Info("kafkaSslClientCertFile", "clientcertfile", sslKakfa.kafkaSslClientCertFile)
	if sslKakfa.kafkaSslClientCertFile != "" && sslKakfa.kafkaSslClientKeyFile != "" && sslKakfa.kafkaSslCACertFile != "" {
		if sslKakfa.kafkaSecurityProtocol == "" {
			sslKakfa.kafkaSecurityProtocol = "ssl"
		}

		// if kafkaSecurityProtocol != "ssl" && kafkaSecurityProtocol != "sasl_ssl" {
		// 	log.Error("invalid config: kafka security protocol is not ssl based but ssl config is provided")
		// }
	}
	workQueue = make(chan LogRequest, logBufferSize)
	writeTimeoutMilliseconds = writeTimeoutMs

	// Now, create all of our workers.
	for i := 0; i < nworkers; i++ {
		log.Info("Starting", "worker", i+1)
		worker, err := NewWorker(i+1, workQueue, log, sdepName, namespace, predictorName, kafkaBroker, kafkaTopic, *sslKakfa)
		if err != nil {
			return err
		}
		worker.Start()
	}

	return nil
}
