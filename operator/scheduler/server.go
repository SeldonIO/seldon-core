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

	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"

	"github.com/seldonio/seldon-core/operator/v2/apis/mlops/v1alpha1"
	"github.com/seldonio/seldon-core/operator/v2/pkg/constants"
)

func (s *SchedulerClient) ServerNotify(ctx context.Context, grpcClient scheduler.SchedulerClient, servers []v1alpha1.Server, isFirstSync bool) error {
	logger := s.logger.WithName("NotifyServer")
	if grpcClient == nil {
		// we assume that all servers are in the same namespace
		namespace := servers[0].Namespace
		conn, err := s.getConnection(namespace)
		if err != nil {
			return err
		}
		grpcClient = scheduler.NewSchedulerClient(conn)
	}

	var requests []*scheduler.ServerNotify
	for _, server := range servers {
		var scalingSpec *v1alpha1.ValidatedScalingSpec
		var err error

		if !server.ObjectMeta.DeletionTimestamp.IsZero() {
			scalingSpec = &v1alpha1.ValidatedScalingSpec{
				Replicas:    0,
				MinReplicas: 0,
				MaxReplicas: 0,
			}
		} else {
			scalingSpec, err = v1alpha1.GetValidatedScalingSpec(server.Spec.Replicas, server.Spec.MinReplicas, server.Spec.MaxReplicas)
			if err != nil {
				return err
			}
		}

		logger.Info(
			"Notify server", "name", server.GetName(), "namespace", server.GetNamespace(),
			"replicas", scalingSpec.Replicas,
			"minReplicas", scalingSpec.MinReplicas,
			"maxReplicas", scalingSpec.MaxReplicas,
		)

		requests = append(requests, &scheduler.ServerNotify{
			Name:             server.GetName(),
			ExpectedReplicas: scalingSpec.Replicas,
			MinReplicas:      scalingSpec.MinReplicas,
			MaxReplicas:      scalingSpec.MaxReplicas,
			KubernetesMeta: &scheduler.KubernetesMeta{
				Namespace:  server.GetNamespace(),
				Generation: server.GetGeneration(),
			},
		})
	}
	request := &scheduler.ServerNotifyRequest{
		Servers:     requests,
		IsFirstSync: isFirstSync,
	}
	_, err := grpcClient.ServerNotify(
		ctx,
		request,
		grpc_retry.WithMax(schedulerConnectMaxRetries),
		grpc_retry.WithBackoff(grpc_retry.BackoffExponential(schedulerConnectBackoffScalar)),
	)
	if err != nil {
		logger.Error(err, "Failed to send notify server to scheduler")
		return err
	}
	logger.V(1).Info("Sent notify server to scheduler", "servers", len(servers), "isFirstSync", isFirstSync)
	return nil
}

// note: namespace is not used in this function
func (s *SchedulerClient) SubscribeServerEvents(ctx context.Context, grpcClient scheduler.SchedulerClient, namespace string) error {
	logger := s.logger.WithName("SubscribeServerEvents")

	stream, err := grpcClient.SubscribeServerStatus(
		ctx,
		&scheduler.ServerSubscriptionRequest{SubscriberName: "seldon manager"},
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

		logger.Info("Received event", "server", event.ServerName)

		if event.GetKubernetesMeta() == nil {
			logger.Info("Received server event with no k8s metadata so ignoring", "server", event.ServerName)
			continue
		}
		server := &v1alpha1.Server{}
		err = s.Get(ctx, client.ObjectKey{Name: event.ServerName, Namespace: event.GetKubernetesMeta().GetNamespace()}, server)
		if err != nil {
			logger.Error(err, "Failed to get server", "name", event.ServerName, "namespace", event.GetKubernetesMeta().GetNamespace())
			continue
		}

		// Try to update status
		retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			contextWithTimeout, cancel := context.WithTimeout(ctx, constants.K8sAPICallsTxTimeout)
			defer cancel()

			server := &v1alpha1.Server{}
			err = s.Get(contextWithTimeout, client.ObjectKey{Name: event.ServerName, Namespace: event.GetKubernetesMeta().GetNamespace()}, server)
			if err != nil {
				return err
			}
			if event.GetKubernetesMeta().Generation != server.Generation {
				logger.Info("Ignoring event for old generation", "currentGeneration", server.Generation, "eventGeneration", event.GetKubernetesMeta().Generation, "server", event.ServerName)
				return nil
			}

			// The types of updates we may get from the scheduler are:
			// 1. Status updates
			// 2. Requests for changing the number of server replicas
			// 3. Updates containing non-authoritative replica info, because the scheduler is in a
			// discovery phase (just starting up, after a restart)
			//
			// At the moment, the scheduler doesn't send multiple types of updates in a single event;
			switch event.GetType() {
			case scheduler.ServerStatusResponse_StatusUpdate:
				return s.updateServerStatus(contextWithTimeout, server) // todo: implement replica info update
			case scheduler.ServerStatusResponse_ScalingRequest:
				return nil // TODO: implement scaling request
			case scheduler.ServerStatusResponse_NonAuthoritativeReplicaInfo:
				// skip updating replica info, only update status
				return s.updateServerStatus(contextWithTimeout, server)
			default: // we ignore unknown event types
				return nil
			}
		})
		if retryErr != nil {
			logger.Error(err, "Failed to update status", "model", event.ServerName)
		}

	}
	return nil
}

func (s *SchedulerClient) updateServerStatus(ctx context.Context, server *v1alpha1.Server) error {
	if err := s.Status().Update(ctx, server); err != nil {
		s.recorder.Eventf(server, v1.EventTypeWarning, "UpdateFailed",
			"Failed to update status for Server %q: %v", server.Name, err)
		return err
	}
	return nil
}
