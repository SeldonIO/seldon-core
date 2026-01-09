/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package server

import (
	"context"
	"errors"
	"math/rand/v2"
	"time"

	"github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler/db"
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
		if errors.Is(err, context.Canceled) {
			return true, status.Errorf(codes.Canceled, "Stream disconnected")
		}
		return false, err
	}
}

func shouldScaleUp(server *db.Server, stats *store.ServerStats) (bool, uint32) {
	if server.ExpectedReplicas < 0 || server.MaxReplicas < 1 {
		return false, 0
	}
	if stats != nil {
		maxNumReplicaHostedModels := stats.MaxNumReplicaHostedModels
		return maxNumReplicaHostedModels > uint32(server.ExpectedReplicas), min(maxNumReplicaHostedModels, uint32(server.MaxReplicas))
	}
	return false, 0
}

func shouldScaleDown(server *db.Server, stats *store.ServerStats, perc float32) (bool, uint32) {

	if stats != nil {
		currentReplicas := uint32(server.ExpectedReplicas)
		minReplicas := uint32(server.MinReplicas)
		if minReplicas == 0 {
			minReplicas = 1
		}
		// 25% chance of trying to pack replicas if models are not fully packed
		tryPack := false
		rand := rand.Float32()
		if rand > (1 - perc) {
			if stats.MaxNumReplicaHostedModels < currentReplicas {
				tryPack = true
			}
		}
		// we do scaling down if:
		// 1. we are trying to pack replicas: max number of replicas for any hosted model is less than the number of expected replicas (only 25% of the time)
		// 2. we have empty replicas and the server has more than one expected replicas

		targetReplicas := max(minReplicas, currentReplicas-stats.NumEmptyReplicas)
		if tryPack {
			toRemoveReplicasfromPack := currentReplicas - stats.MaxNumReplicaHostedModels
			remainingReplicasfromPack := currentReplicas - toRemoveReplicasfromPack
			if toRemoveReplicasfromPack > stats.NumEmptyReplicas && remainingReplicasfromPack > 0 {
				targetReplicas = max(minReplicas, remainingReplicasfromPack)
			}
		}

		return (tryPack || stats.NumEmptyReplicas > 0) && server.ExpectedReplicas > 1, targetReplicas
	}
	return false, 0
}
