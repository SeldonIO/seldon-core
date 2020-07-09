package logger

import "fmt"

const LoggerWorkerQueueSize = 100

// A buffered channel that we can send work requests on.
var WorkQueue = make(chan LogRequest, LoggerWorkerQueueSize)

func QueueLogRequest(req LogRequest) error {
	WorkQueue <- req
	fmt.Println("Work request queued")
	return nil
}
