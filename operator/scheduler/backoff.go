/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package scheduler

import (
	"context"
	"time"

	v4backoff "github.com/cenkalti/backoff/v4"
	"github.com/go-logr/logr"
	"google.golang.org/grpc"

	"github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"
)

func retryFnConstBackoff(fn func() error, log func(err error, duration time.Duration)) error {
	return v4backoff.RetryNotify(func() error {
		return fn()
	}, v4backoff.WithMaxRetries(v4backoff.NewConstantBackOff(schedulerConstantBackoff), schedulerConnectMaxRetries), log)
}

func retryFnExpBackoff(
	fn func(context context.Context, grpcClient scheduler.SchedulerClient, namespace string) error,
	conn *grpc.ClientConn, namespace string, logger logr.Logger,
) error {
	logger.Info("Retrying to connect", "namespace", namespace)
	logFailure := func(err error, delay time.Duration) {
		logger.Error(err, "Scheduler not ready")
	}
	backOffExp := getClientExponentialBackoff()
	fnWithArgs := func() error {
		grpcClient := scheduler.NewSchedulerClient(conn)
		return fn(context.Background(), grpcClient, namespace)
	}
	err := v4backoff.RetryNotify(fnWithArgs, backOffExp, logFailure)
	if err != nil {
		logger.Error(err, "Failed to connect to scheduler", "namespace", namespace)
		return err
	}
	return nil
}

func getClientExponentialBackoff() *v4backoff.ExponentialBackOff {
	backOffExp := v4backoff.NewExponentialBackOff()
	backOffExp.MaxElapsedTime = backoffMaxElapsedTime
	backOffExp.MaxInterval = backOffMaxInterval
	backOffExp.InitialInterval = backOffInitialInterval
	return backOffExp
}
