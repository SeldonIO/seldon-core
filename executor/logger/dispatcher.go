package logger

import (
	"os"

	"github.com/go-logr/logr"
)

const (
	ENV_LOGGER_KAFKA_BROKER = "LOGGER_KAFKA_BROKER"
	ENV_LOGGER_KAFKA_TOPIC  = "LOGGER_KAFKA_TOPIC"
)

func StartDispatcher(nworkers int, logBufferSize int, writeTimeoutMs int, log logr.Logger, sdepName string, namespace string, predictorName string, kafkaBroker string, kafkaTopic string, protocol string) error {
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

	workQueue = make(chan LogRequest, logBufferSize)
	writeTimeoutMilliseconds = writeTimeoutMs
	// Now, create all of our workers.
	for i := 0; i < nworkers; i++ {
		log.Info("Starting", "worker", i+1)
		worker, err := NewWorker(i+1, workQueue, log, sdepName, namespace, predictorName, kafkaBroker, kafkaTopic, protocol)
		if err != nil {
			return err
		}
		worker.Start()
	}

	return nil
}
