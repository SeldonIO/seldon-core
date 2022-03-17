package scheduler

import (
	"context"
	"io"

	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"github.com/seldonio/seldon-core/operatorv2/apis/mlops/v1alpha1"
	"github.com/seldonio/seldon-core/operatorv2/scheduler/apis/mlops/scheduler"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (s *SchedulerClient) ServerNotify(ctx context.Context, server *v1alpha1.Server) error {
	logger := s.logger.WithName("NotifyServer")
	grcpClient := scheduler.NewSchedulerClient(s.conn)

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
	_, err := grcpClient.ServerNotify(ctx, request, grpc_retry.WithMax(2))
	if err != nil {
		return err
	}
	return nil
}

func (s *SchedulerClient) SubscribeServerEvents(ctx context.Context) error {
	logger := s.logger.WithName("SubscribeServerEvents")
	grcpClient := scheduler.NewSchedulerClient(s.conn)

	stream, err := grcpClient.SubscribeServerStatus(ctx, &scheduler.ServerSubscriptionRequest{SubscriberName: "seldon manager"}, grpc_retry.WithMax(1))
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
