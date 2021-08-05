package logger

const LoggerWorkerQueueSize = 100

// A buffered channel that we can send work requests on.
var WorkQueue = make(chan LogRequest, LoggerWorkerQueueSize)

func QueueLogRequest(req LogRequest) error {
	WorkQueue <- req
	return nil
}
