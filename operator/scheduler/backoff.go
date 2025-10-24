package scheduler

import (
	"time"

	v4backoff "github.com/cenkalti/backoff/v4"
)

func backoff(fn func() error, log func(err error, duration time.Duration)) error {
	return v4backoff.RetryNotify(func() error {
		return fn()
	}, v4backoff.WithMaxRetries(v4backoff.NewConstantBackOff(schedulerConstantBackoff), schedulerConnectMaxRetries), log)
}
