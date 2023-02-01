package logger

import (
	"errors"
	"time"
)

const (
	DefaultWorkQueueSize            = 10000
	DefaultWriteTimeoutMilliseconds = 2000
)

var (
	// Default values of these variables are declared here. StartDispatcher can overwrite them with user provided values.
	// workQueue is a buffered channel that we can send work requests on.
	workQueue = make(chan LogRequest, DefaultWorkQueueSize)
	// writeTimeoutMilliseconds is the timeout for waiting for work to be written to the queue. If 0, will not wait if buffer is full.
	writeTimeoutMilliseconds = DefaultWriteTimeoutMilliseconds
)

func QueueLogRequest(req LogRequest) error {
	timer := time.NewTimer(time.Duration(writeTimeoutMilliseconds) * time.Millisecond)
	defer timer.Stop()
	select {
	case workQueue <- req:
		return nil
	case <-timer.C:
		return errors.New("timed out waiting to queue log request: buffer is full")
	}
}
