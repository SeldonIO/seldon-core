package logger

import (
	"github.com/pkg/errors"
	"time"
)

// TODO(ivan): Make configurable
const LoggerWorkerQueueSize = 10000

// A buffered channel that we can send work requests on.
var WorkQueue = make(chan LogRequest, LoggerWorkerQueueSize)

func QueueLogRequest(req LogRequest) error {
	select {
	case WorkQueue <- req:
		return nil
	case <- time.After(2 * time.Second): // TODO(ivan): make timeout configurable? The timeout is basically, maxLogWaitOnFullBuffer
		return errors.New("timed out waiting to queue log request: buffer is full")
	}
}
