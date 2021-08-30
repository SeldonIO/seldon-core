package logger

import (
	"github.com/go-logr/logr"
	"os"
)

const (
	ENV_LOGGER_KAFKA_BROKER = "LOGGER_KAFKA_BROKER"
	ENV_LOGGER_KAFKA_TOPIC  = "LOGGER_KAFKA_TOPIC"
)

var WorkerQueue chan chan LogRequest

func StartDispatcher(nworkers int, log logr.Logger, sdepName string, namespace string, predictorName string, kafkaBroker string, kafkaTopic string) error {
	// First, initialize the channel we are going to put the workers' work channels into.
	WorkerQueue = make(chan chan LogRequest, nworkers)

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

	// Now, create all of our workers.
	for i := 0; i < nworkers; i++ {
		log.Info("Starting", "worker", i+1)
		worker, err := NewWorker(i+1, WorkerQueue, log, sdepName, namespace, predictorName, kafkaBroker, kafkaTopic)
		if err != nil {
			return err
		}
		worker.Start()
	}

	go func() {
		for {
			select {
			case work := <-WorkQueue:
				go func() {
					worker := <-WorkerQueue

					worker <- work
				}()
			}
		}
	}()

	return nil
}
