/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

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
