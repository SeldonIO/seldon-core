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
	"github.com/seldonio/seldon-core/operator/v2/pkg/utils"
)

func (s *SchedulerClient) LoadPipeline(ctx context.Context, pipeline *v1alpha1.Pipeline, grpcClient scheduler.SchedulerClient) (bool, error) {
	logger := s.logger.WithName("LoadPipeline")
	var err error
	if grpcClient == nil {
		conn, err := s.getConnection(pipeline.Namespace)
		if err != nil {
			return true, err
		}
		grpcClient = scheduler.NewSchedulerClient(conn)
	}
	req := scheduler.LoadPipelineRequest{
		Pipeline: pipeline.AsSchedulerPipeline(),
	}
	logger.Info("Load", "pipeline name", pipeline.Name)
	_, err = grpcClient.LoadPipeline(
		ctx,
		&req,
		grpc_retry.WithMax(schedulerConnectMaxRetries),
		grpc_retry.WithBackoff(grpc_retry.BackoffExponential(schedulerConnectBackoffScalar)),
	)
	return s.checkErrorRetryable(pipeline.Kind, pipeline.Name, err), err
}

func (s *SchedulerClient) UnloadPipeline(ctx context.Context, pipeline *v1alpha1.Pipeline) (error, bool) {
	logger := s.logger.WithName("UnloadPipeline")
	conn, err := s.getConnection(pipeline.Namespace)
	if err != nil {
		return err, true
	}
	grpcClient := scheduler.NewSchedulerClient(conn)
	req := scheduler.UnloadPipelineRequest{
		Name: pipeline.Name,
	}
	logger.Info("Unload", "pipeline name", pipeline.Name)
	_, err = grpcClient.UnloadPipeline(
		ctx,
		&req,
		grpc_retry.WithMax(schedulerConnectMaxRetries),
		grpc_retry.WithBackoff(grpc_retry.BackoffExponential(schedulerConnectBackoffScalar)),
	)
	if err != nil {
		return err, s.checkErrorRetryable(pipeline.Kind, pipeline.Name, err)
	}
	pipeline.Status.CreateAndSetCondition(
		v1alpha1.PipelineReady,
		false,
		scheduler.PipelineVersionState_PipelineTerminating.String(),
		"Pipeline unload requested",
	)
	_ = s.updatePipelineStatusImpl(ctx, pipeline)
	return nil, false
}

// namespace is not used in this function
func (s *SchedulerClient) SubscribePipelineEvents(ctx context.Context, grpcClient scheduler.SchedulerClient, namespace string) error {
	logger := s.logger.WithName("SubscribePipelineEvents")

	stream, err := grpcClient.SubscribePipelineStatus(
		ctx,
		&scheduler.PipelineSubscriptionRequest{SubscriberName: "seldon manager"},
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
			logger.Error(err, "failed to receive pipeline event")
			return err
		}

		if len(event.Versions) != 1 {
			logger.Info(
				"Unexpected number of pipeline versions",
				"numVersions", len(event.Versions),
				"pipeline", event.PipelineName,
			)
			continue
		}

		pv := event.Versions[0]
		if pv.GetPipeline().GetKubernetesMeta() == nil {
			logger.Info("Received pipeline event with no k8s metadata so ignoring", "pipeline", event.PipelineName)
			continue
		}

		logger.Info(
			"Received event",
			"pipeline", event.PipelineName,
			"generation", pv.GetPipeline().GetKubernetesMeta().Generation,
			"version", pv.GetPipeline().Version,
			"State", pv.GetState().String(),
		)

		if canRemovePipelineFinalizer(pv.State.Status) {
			retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
				ctxWithTimeout, cancel := context.WithTimeout(ctx, constants.K8sAPICallTimeout)
				defer cancel()

				latestPipeline := &v1alpha1.Pipeline{}
				err = s.Get(
					ctxWithTimeout,
					client.ObjectKey{
						Name:      event.PipelineName,
						Namespace: pv.GetPipeline().GetKubernetesMeta().GetNamespace(),
					},
					latestPipeline,
				)
				if err != nil {
					return err
				}
				if !latestPipeline.ObjectMeta.DeletionTimestamp.IsZero() { // Pipeline is being deleted
					// remove finalizer now we have completed successfully
					latestPipeline.ObjectMeta.Finalizers = utils.RemoveStr(
						latestPipeline.ObjectMeta.Finalizers,
						constants.PipelineFinalizerName,
					)
					if err := s.Update(ctxWithTimeout, latestPipeline); err != nil {
						logger.Error(err, "Failed to remove finalizer", "pipeline", latestPipeline.GetName())
						return err
					}
				}
				return nil
			})
			if retryErr != nil {
				logger.Error(err, "Failed to remove finalizer after retries")
			}
		}

		// Try to update status
		{
			retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
				ctxWithTimeout, cancel := context.WithTimeout(ctx, constants.K8sAPICallTimeout)
				defer cancel()

				pipeline := &v1alpha1.Pipeline{}
				err = s.Get(
					ctxWithTimeout,
					client.ObjectKey{
						Name:      event.PipelineName,
						Namespace: pv.GetPipeline().GetKubernetesMeta().GetNamespace(),
					},
					pipeline,
				)
				if err != nil {
					return err
				}

				if pv.GetPipeline().GetKubernetesMeta().GetGeneration() != pipeline.Generation {
					logger.Info(
						"Ignoring event for old generation",
						"currentGeneration", pipeline.Generation,
						"eventGeneration", pv.GetPipeline().GetKubernetesMeta().GetGeneration(),
						"server", event.PipelineName,
					)
					return nil
				}

				// Handle status update
				switch pv.State.Status {
				case scheduler.PipelineVersionState_PipelineReady:
					logger.Info(
						"Setting pipeline to ready",
						"pipeline", pipeline.Name,
						"generation", pipeline.Generation,
					)
					pipeline.Status.CreateAndSetCondition(
						v1alpha1.PipelineReady,
						true,
						pv.State.Reason,
						pv.State.Status.String(),
					)
				default:
					logger.Info(
						"Setting pipeline to not ready",
						"pipeline", pipeline.Name,
						"generation", pipeline.Generation,
					)
					pipeline.Status.CreateAndSetCondition(
						v1alpha1.PipelineReady,
						false,
						pv.State.Reason,
						pv.State.Status.String(),
					)
				}
				// Set models ready
				if pv.State.ModelsReady {
					pipeline.Status.CreateAndSetCondition(v1alpha1.ModelsReady, true, "Models all available", "")
				} else {
					pipeline.Status.CreateAndSetCondition(v1alpha1.ModelsReady, false, "Some models are not available", "")
				}

				return s.updatePipelineStatusImpl(ctxWithTimeout, pipeline)
			})
			if retryErr != nil {
				logger.Error(retryErr, "Failed to update status", "pipeline", event.PipelineName)
			}
		}
	}
	return nil
}

func (s *SchedulerClient) updatePipelineStatusImpl(ctx context.Context, pipeline *v1alpha1.Pipeline) error {
	if err := s.Status().Update(ctx, pipeline); err != nil {
		s.recorder.Eventf(pipeline, v1.EventTypeWarning, "UpdateFailed",
			"Failed to update status for pipeline %q: %v", pipeline.Name, err)
		return err
	}
	return nil
}

func canRemovePipelineFinalizer(state scheduler.PipelineVersionState_PipelineStatus) bool {
	switch state {
	// we should wait if the state is not terminal for deleting the finalizer, it should be Terminated in the case of delete
	case scheduler.PipelineVersionState_PipelineTerminating, scheduler.PipelineVersionState_PipelineTerminate:
		return false
	default:
		return true
	}
}
