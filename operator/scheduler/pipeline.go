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
	"github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"
	"github.com/seldonio/seldon-core/operator/v2/apis/mlops/v1alpha1"
	"github.com/seldonio/seldon-core/operator/v2/pkg/constants"
	"github.com/seldonio/seldon-core/operator/v2/pkg/utils"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (s *SchedulerClient) LoadPipeline(ctx context.Context, pipeline *v1alpha1.Pipeline) (error, bool) {
	logger := s.logger.WithName("LoadPipeline")
	grcpClient := scheduler.NewSchedulerClient(s.conn)
	req := scheduler.LoadPipelineRequest{
		Pipeline: pipeline.AsSchedulerPipeline(),
	}
	logger.Info("Load", "pipeline name", pipeline.Name)
	_, err := grcpClient.LoadPipeline(ctx, &req, grpc_retry.WithMax(2))
	return err, s.checkErrorRetryable(pipeline.Kind, pipeline.Name, err)
}

func (s *SchedulerClient) UnloadPipeline(ctx context.Context, pipeline *v1alpha1.Pipeline) (error, bool) {
	logger := s.logger.WithName("UnloadPipeline")
	grcpClient := scheduler.NewSchedulerClient(s.conn)
	req := scheduler.UnloadPipelineRequest{
		Name: pipeline.Name,
	}
	logger.Info("Unload", "pipeline name", pipeline.Name)
	_, err := grcpClient.UnloadPipeline(ctx, &req, grpc_retry.WithMax(2))
	if err != nil {
		return err, s.checkErrorRetryable(pipeline.Kind, pipeline.Name, err)
	}
	pipeline.Status.CreateAndSetCondition(
		v1alpha1.PipelineReady,
		false,
		scheduler.PipelineVersionState_PipelineTerminating.String(),
		"Pipeline unload requested",
	)
	_ = s.updatePipelineStatusImpl(pipeline)
	return nil, false
}

func (s *SchedulerClient) SubscribePipelineEvents(ctx context.Context) error {
	logger := s.logger.WithName("SubscribePipelineEvents")
	grcpClient := scheduler.NewSchedulerClient(s.conn)

	stream, err := grcpClient.SubscribePipelineStatus(
		ctx,
		&scheduler.PipelineSubscriptionRequest{SubscriberName: "seldon manager"},
		grpc_retry.WithMax(1),
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

		pipeline := &v1alpha1.Pipeline{}
		err = s.Get(
			ctx,
			client.ObjectKey{
				Name:      event.PipelineName,
				Namespace: pv.GetPipeline().GetKubernetesMeta().GetNamespace(),
			},
			pipeline,
		)
		if err != nil {
			logger.Error(
				err,
				"Failed to get pipeline",
				"name", event.PipelineName,
				"namespace", pv.GetPipeline().GetKubernetesMeta().GetNamespace(),
			)
			continue
		}

		if !pipeline.ObjectMeta.DeletionTimestamp.IsZero() {
			logger.Info("Pipeline is pending deletion", "pipeline", pipeline.Name)
			if canRemovePipelineFinalizer(pv.State.Status) {
				retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
					latestPipeline := &v1alpha1.Pipeline{}
					err = s.Get(
						ctx,
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
						if err := s.Update(ctx, latestPipeline); err != nil {
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
		}

		// Try to update status
		retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			pipeline := &v1alpha1.Pipeline{}
			err = s.Get(
				ctx,
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
					"pipeline", event.PipelineName,
					"generation", pipeline.Generation,
				)
				pipeline.Status.CreateAndSetCondition(
					v1alpha1.PipelineReady,
					true,
					pv.State.Status.String(),
					pv.State.Reason,
				)
			default:
				logger.Info(
					"Setting pipeline to not ready",
					"pipeline", event.PipelineName,
					"generation", pipeline.Generation,
				)
				pipeline.Status.CreateAndSetCondition(
					v1alpha1.PipelineReady,
					false,
					pv.State.Status.String(),
					pv.State.Reason,
				)
			}
			// Set models ready
			if pv.State.ModelsReady {
				pipeline.Status.CreateAndSetCondition(v1alpha1.ModelsReady, true, "Models all available", "")
			} else {
				pipeline.Status.CreateAndSetCondition(v1alpha1.ModelsReady, false, "Some models are not available", "")
			}

			return s.updatePipelineStatusImpl(pipeline)
		})
		if retryErr != nil {
			logger.Error(retryErr, "Failed to update status", "pipeline", event.PipelineName)
		}

	}
	return nil
}

func (s *SchedulerClient) updatePipelineStatusImpl(pipeline *v1alpha1.Pipeline) error {
	if err := s.Status().Update(context.TODO(), pipeline); err != nil {
		s.recorder.Eventf(pipeline, v1.EventTypeWarning, "UpdateFailed",
			"Failed to update status for pipeline %q: %v", pipeline.Name, err)
		return err
	}
	return nil
}

func canRemovePipelineFinalizer(state scheduler.PipelineVersionState_PipelineStatus) bool {
	switch state {
	case scheduler.PipelineVersionState_PipelineTerminating:
		return false
	default:
		return true
	}
}
