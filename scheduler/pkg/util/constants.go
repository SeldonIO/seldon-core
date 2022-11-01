package util

import "time"

const (
	GrpcRetryBackoffMillisecs         = 100
	EnvoyUpdateDefaultBatchWaitMillis = 250 * time.Millisecond
)
