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
	"io"

	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"

	"github.com/seldonio/seldon-core/operator/v2/apis/mlops/v1alpha1"
	"github.com/seldonio/seldon-core/operator/v2/pkg/constants"
	"github.com/seldonio/seldon-core/operator/v2/pkg/utils"
)

// TODO: unify these helper functions as they do more or less the same thing

func handleLoadedExperiments(
	ctx context.Context, namespace string, s *SchedulerClient, grpcClient scheduler.SchedulerClient) {
	experimentList := &v1alpha1.ExperimentList{}
	// Get all experiments in the namespace
	err := s.List(
		ctx,
		experimentList,
		client.InNamespace(namespace),
	)
	if err != nil {
		return
	}

	for _, experiment := range experimentList.Items {
		// experiments that are not in the process of being deleted has DeletionTimestamp as zero
		if experiment.ObjectMeta.DeletionTimestamp.IsZero() {
			s.logger.V(1).Info("Calling start experiment (on reconnect)", "experiment", experiment.Name)
			if _, err := s.StartExperiment(ctx, &experiment, grpcClient); err != nil {
				// if this is a retryable error, we will retry on the next connection reconnect
				s.logger.Error(err, "Failed to call start experiment", "experiment", experiment.Name)
			} else {
				s.logger.V(1).Info("Start experiment called successfully", "experiment", experiment.Name)
			}
		}
	}
}

func handlePendingDeleteExperiments(
	ctx context.Context, namespace string, s *SchedulerClient) {
	experimentList := &v1alpha1.ExperimentList{}
	// Get all models in the namespace
	err := s.List(
		ctx,
		experimentList,
		client.InNamespace(namespace),
	)
	if err != nil {
		return
	}

	// Check if any experiments are being deleted
	for _, experiment := range experimentList.Items {
		if !experiment.ObjectMeta.DeletionTimestamp.IsZero() {
			s.logger.V(1).Info("Removing finalizer (on reconnect)", "experiment", experiment.Name)
			retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
				experiment.ObjectMeta.Finalizers = utils.RemoveStr(experiment.ObjectMeta.Finalizers, constants.ExperimentFinalizerName)
				if errUpdate := s.Update(ctx, &experiment); errUpdate != nil {
					s.logger.Error(err, "Failed to remove finalizer", "experiment", experiment.Name)
					return errUpdate
				}
				s.logger.Info("Removed finalizer", "experiment", experiment.Name)
				return nil
			})
			if retryErr != nil {
				s.logger.Error(err, "Failed to remove finalizer after retries", "experiment", experiment.Name)
			}
		}
	}
}

// when need to reload the models that are marked in k8s as loaded, this is because there could be a
// case where the scheduler has load the models state (if the scheduler and the model server restart at the same time)
func handleLoadedModels(
	ctx context.Context, namespace string, s *SchedulerClient, grpcClient scheduler.SchedulerClient) {
	modelList := &v1alpha1.ModelList{}
	// Get all models in the namespace
	err := s.List(
		ctx,
		modelList,
		client.InNamespace(namespace),
	)
	if err != nil {
		return
	}

	for _, model := range modelList.Items {
		// models that are not in the process of being deleted has DeletionTimestamp as zero
		if model.ObjectMeta.DeletionTimestamp.IsZero() {
			s.logger.V(1).Info("Calling Load model (on reconnect)", "model", model.Name)
			if _, err := s.LoadModel(ctx, &model, grpcClient); err != nil {
				// if this is a retryable error, we will retry on the next connection reconnect
				s.logger.Error(err, "Failed to call load model", "model", model.Name)
			} else {
				s.logger.V(1).Info("Load model called successfully", "model", model.Name)
			}
		} else {
			s.logger.V(1).Info("Model is being deleted, not loading", "model", model.Name)
		}
	}
}

func handleRegisteredServers(
	ctx context.Context, namespace string, s *SchedulerClient, grpcClient scheduler.SchedulerClient) {
	serverList := &v1alpha1.ServerList{}
	// Get all servers in the namespace
	err := s.List(
		ctx,
		serverList,
		client.InNamespace(namespace),
	)
	if err != nil {
		return
	}

	if err := s.ServerNotify(ctx, grpcClient, serverList.Items); err != nil {
		s.logger.Error(err, "Failed to notify servers", "servers", serverList.Items)
	}
}

func handlePendingDeleteModels(
	ctx context.Context, namespace string, s *SchedulerClient, grpcClient scheduler.SchedulerClient) {
	modelList := &v1alpha1.ModelList{}
	// Get all models in the namespace
	err := s.List(
		ctx,
		modelList,
		client.InNamespace(namespace),
	)
	if err != nil {
		return
	}

	// Check if any models are being deleted
	for _, model := range modelList.Items {
		if !model.ObjectMeta.DeletionTimestamp.IsZero() {
			s.logger.V(1).Info("Calling Unload model (on reconnect)", "model", model.Name)
			if retryUnload, err := s.UnloadModel(ctx, &model, grpcClient); err != nil {
				if retryUnload {
					// caller will retry as this method is called on connection reconnect
					s.logger.Error(err, "Failed to call unload model", "model", model.Name)
					continue
				} else {
					// this is essentially a failed pre-condition (model does not exist in scheduler)
					// we can remove
					// note that there is still the chance the model is not updated from the different model servers
					// upon reconnection of the scheduler
					retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
						model.ObjectMeta.Finalizers = utils.RemoveStr(model.ObjectMeta.Finalizers, constants.ModelFinalizerName)
						if errUpdate := s.Update(ctx, &model); errUpdate != nil {
							s.logger.Error(err, "Failed to remove finalizer", "model", model.Name)
							return errUpdate
						}
						s.logger.Info("Removed finalizer", "model", model.Name)
						return nil
					})
					if retryErr != nil {
						s.logger.Error(err, "Failed to remove finalizer after retries", "model", model.Name)
					}
				}
			} else {
				// if the model exists in the scheduler so we wait until we get the event from the subscription stream
				s.logger.Info("Unload model called successfully, not removing finalizer", "model", model.Name)
			}
		} else {
			s.logger.V(1).Info("Model is not being deleted, not unloading", "model", model.Name)
		}
	}
}

func handleLoadedPipelines(
	ctx context.Context, namespace string, s *SchedulerClient, grpcClient scheduler.SchedulerClient) {
	pipelineList := &v1alpha1.PipelineList{}
	// Get all pipelines in the namespace
	err := s.List(
		ctx,
		pipelineList,
		client.InNamespace(namespace),
	)
	if err != nil {
		return
	}

	for _, pipeline := range pipelineList.Items {
		// pipelines that are not in the process of being deleted has DeletionTimestamp as zero
		if pipeline.ObjectMeta.DeletionTimestamp.IsZero() {
			s.logger.V(1).Info("Calling load pipeline (on reconnect)", "pipeline", pipeline.Name)
			if _, err := s.LoadPipeline(ctx, &pipeline, grpcClient); err != nil {
				// if this is a retryable error, we will retry on the next connection reconnect
				s.logger.Error(err, "Failed to call load pipeline", "pipeline", pipeline.Name)
			} else {
				s.logger.V(1).Info("Load pipeline called successfully", "pipeline", pipeline.Name)
			}
		}
	}
}

func handlePendingDeletePipelines(
	ctx context.Context, namespace string, s *SchedulerClient) {
	pipelineList := &v1alpha1.PipelineList{}
	// Get all models in the namespace
	err := s.List(
		ctx,
		pipelineList,
		client.InNamespace(namespace),
	)
	if err != nil {
		return
	}

	// Check if any pipelines are being deleted
	for _, pipeline := range pipelineList.Items {
		if !pipeline.ObjectMeta.DeletionTimestamp.IsZero() {
			s.logger.V(1).Info("Removing finalizer (on reconnect)", "pipeline", pipeline.Name)
			retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
				pipeline.ObjectMeta.Finalizers = utils.RemoveStr(pipeline.ObjectMeta.Finalizers, constants.PipelineFinalizerName)
				if errUpdate := s.Update(ctx, &pipeline); errUpdate != nil {
					s.logger.Error(err, "Failed to remove finalizer", "pipeline", pipeline.Name)
					return errUpdate
				}
				s.logger.Info("Removed finalizer", "pipeline", pipeline.Name)
				return nil
			})
			if retryErr != nil {
				s.logger.Error(err, "Failed to remove finalizer after retries", "pipeline", pipeline.Name)
			}
		}
	}
}

func getNumExperimentsFromScheduler(ctx context.Context, grpcClient scheduler.SchedulerClient) (int, error) {
	req := &scheduler.ExperimentStatusRequest{
		SubscriberName: "seldon manager",
	}
	streamForExperimentStatuses, err := grpcClient.ExperimentStatus(ctx, req)
	numExperimentsFromScheduler := 0
	if err != nil {
		return 0, err
	}
	for {
		data, err := streamForExperimentStatuses.Recv()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return 0, err
			}
		}
		print(data)
		numExperimentsFromScheduler++
	}
	return numExperimentsFromScheduler, nil
}

func getNumPipelinesFromScheduler(ctx context.Context, grpcClient scheduler.SchedulerClient) (int, error) {
	req := &scheduler.PipelineStatusRequest{
		SubscriberName: "seldon manager",
	}
	streamForPipelineStatuses, err := grpcClient.PipelineStatus(ctx, req)
	numPipelinesFromScheduler := 0
	if err != nil {
		return 0, err
	}
	for {
		data, err := streamForPipelineStatuses.Recv()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return 0, err
			}
		}
		print(data)
		numPipelinesFromScheduler++
	}
	return numPipelinesFromScheduler, nil
}
