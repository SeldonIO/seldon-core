/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package server

import (
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store"
)

func sendWithTimeout(f func() error, d time.Duration) (bool, error) {
	errChan := make(chan error, 1)
	go func() {
		errChan <- f()
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

func shouldScaleUp(server *store.ServerSnapshot) bool {
	if server.Stats != nil {
		stats := server.Stats

		if server.ExpectedReplicas < 0 { // it's not set
			return false
		}

		return stats.MaxNumReplicaHostedModels > uint32(server.ExpectedReplicas)
	}
	return false
}
