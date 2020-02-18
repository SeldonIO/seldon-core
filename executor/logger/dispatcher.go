package logger

import (
	"github.com/go-logr/logr"
)

var WorkerQueue chan chan LogRequest

func StartDispatcher(nworkers int, log logr.Logger, sdepName string, namespace string, predictorName string) {
	// First, initialize the channel we are going to put the workers' work channels into.
	WorkerQueue = make(chan chan LogRequest, nworkers)

	// Now, create all of our workers.
	for i := 0; i < nworkers; i++ {
		log.Info("Starting", "worker", i+1)
		worker := NewWorker(i+1, WorkerQueue, log, sdepName, namespace, predictorName)
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
}
