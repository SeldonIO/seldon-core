/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package scheduler

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/cenkalti/backoff/v4"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"

	"github.com/seldonio/seldon-core/operator/v2/pkg/constants"
)

func (s *SchedulerClient) SubscribeControlPlaneEvents(ctx context.Context, grpcClient scheduler.SchedulerClient, namespace string) error {
	logger := s.logger.WithName("SubscribeControlPlaneEvents")

	var cancel context.CancelFunc
	ctx, cancel = context.WithCancel(ctx)
	defer cancel()
	stream, err := grpcClient.SubscribeControlPlane(
		ctx,
		&scheduler.ControlPlaneSubscriptionRequest{SubscriberName: "seldon manager"},
		grpc_retry.WithMax(schedulerConnectMaxRetries),
		grpc_retry.WithBackoff(grpc_retry.BackoffExponential(schedulerConnectBackoffScalar)),
	)
	if err != nil {
		return err
	}

	for {
		event, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("got stream recv error: %w", err)
		}
		logger.Info("Received control plane event", "event", event)

		fn := func(ctx context.Context) error {
			return s.handleControlPlaneEvent(ctx, grpcClient, namespace, event.GetEvent())
		}

		go func() {
			err := backoff.Retry(func() error {
				// in general, we could have also handled timeout via a context with timeout
				// but we want to handle the timeout in a more controlled way and not depending on the other side
				_, err = execWithTimeout(ctx, fn, constants.ControlPlaneExecTimeOut)
				if err != nil {
					logger.Error(err, "Failed to process control plane event", "event", event)
					return err
				}
				return nil
			}, backoff.NewExponentialBackOff(backoff.WithMaxElapsedTime(time.Minute*10)))
			if err != nil {
				logger.Error(err, "Failed to handle event", "namespace", namespace, "event", event)
				return
			}

			logger.Info("Handled control plane event", "event", event)
		}()
	}
	return nil
}

func execWithTimeout(baseContext context.Context, f func(ctx context.Context) error, d time.Duration) (bool, error) {
	// cancel the context after the timeout
	ctxWithCancel, cancel := context.WithCancel(baseContext)
	defer cancel()

	errChan := make(chan error, 1)
	go func() {
		errChan <- f(ctxWithCancel)
		close(errChan)
	}()
	t := time.NewTimer(d)
	select {
	case <-t.C:
		return true, status.Errorf(codes.DeadlineExceeded, "Failed to send event within timeout")
	case err := <-errChan:
		if !t.Stop() {
			<-t.C
		}
		return false, err
	}
}
