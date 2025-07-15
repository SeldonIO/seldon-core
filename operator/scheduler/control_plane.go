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
	"io"
	"time"

	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"
	"github.com/seldonio/seldon-core/operator/v2/pkg/constants"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *schedulerClient) SubscribeControlPlaneEvents(ctx context.Context, grpcClient scheduler.SchedulerClient, namespace string) error {
	logger := s.logger.WithName("SubscribeControlPlaneEvents")

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
			logger.Error(err, "event recv failed")
			return err
		}
		logger.Info("Received event to handle state", "event", event)

		fn := func(ctx context.Context) error {
			return s.handleStateOnReconnect(ctx, grpcClient, namespace, event.GetEvent())
		}
		// in general we could have also handled timeout via a context with timeout
		// but we want to handle the timeout in a more controlled way and not depending on the other side
		_, err = execWithTimeout(ctx, fn, constants.ControlPlaneExecTimeOut)
		if err != nil {
			logger.Error(err, "Failed to handle state on reconnect")
			return err
		}

		logger.Info("Handled state on reconnect")

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
