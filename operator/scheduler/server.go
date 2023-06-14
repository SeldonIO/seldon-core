/*
Copyright 2022 Seldon Technologies Ltd.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package scheduler

import (
	"context"
	"io"

	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"google.golang.org/grpc"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"

	"github.com/seldonio/seldon-core/operator/v2/apis/mlops/v1alpha1"
)

func (s *SchedulerClient) ServerNotify(ctx context.Context, server *v1alpha1.Server) error {
	logger := s.logger.WithName("NotifyServer")
	conn, err := s.getConnection(server.Namespace)
	if err != nil {
		return err
	}
	grcpClient := scheduler.NewSchedulerClient(conn)

	var replicas int32
	if !server.ObjectMeta.DeletionTimestamp.IsZero() {
		replicas = 0
	} else if server.Spec.Replicas != nil {
		replicas = *server.Spec.Replicas
	} else {
		replicas = 1
	}

	request := &scheduler.ServerNotifyRequest{
		Name:             server.GetName(),
		ExpectedReplicas: replicas,
		KubernetesMeta: &scheduler.KubernetesMeta{
			Namespace:  server.GetNamespace(),
			Generation: server.GetGeneration(),
		},
	}
	logger.Info("Notify server", "name", server.GetName(), "namespace", server.GetNamespace(), "replicas", replicas)
	_, err = grcpClient.ServerNotify(ctx, request, grpc_retry.WithMax(SchedulerConnectMaxRetries))
	if err != nil {
		return err
	}
	return nil
}

func (s *SchedulerClient) SubscribeServerEvents(ctx context.Context, conn *grpc.ClientConn) error {
	logger := s.logger.WithName("SubscribeServerEvents")
	grcpClient := scheduler.NewSchedulerClient(conn)

	stream, err := grcpClient.SubscribeServerStatus(ctx, &scheduler.ServerSubscriptionRequest{SubscriberName: "seldon manager"}, grpc_retry.WithMax(SchedulerConnectMaxRetries))
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
			server := &v1alpha1.Server{}
			err = s.Get(ctx, client.ObjectKey{Name: event.ServerName, Namespace: event.GetKubernetesMeta().GetNamespace()}, server)
			if err != nil {
				return err
			}
			if event.GetKubernetesMeta().Generation != server.Generation {
				logger.Info("Ignoring event for old generation", "currentGeneration", server.Generation, "eventGeneration", event.GetKubernetesMeta().Generation, "server", event.ServerName)
				return nil
			}
			// Handle status update
			// This is key for finalizer to remove server when loaded models is zero
			server.Status.LoadedModelReplicas = event.NumLoadedModelReplicas
			return s.updateServerStatus(server)
		})
		if retryErr != nil {
			logger.Error(err, "Failed to update status", "model", event.ServerName)
		}

	}
	return nil
}

func (s *SchedulerClient) updateServerStatus(server *v1alpha1.Server) error {
	if err := s.Status().Update(context.TODO(), server); err != nil {
		s.recorder.Eventf(server, v1.EventTypeWarning, "UpdateFailed",
			"Failed to update status for Server %q: %v", server.Name, err)
		return err
	}
	return nil
}
