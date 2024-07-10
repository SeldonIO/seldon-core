package scheduler

import (
	"context"

	"github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"
	"github.com/seldonio/seldon-core/operator/v2/apis/mlops/v1alpha1"
	"github.com/seldonio/seldon-core/operator/v2/pkg/constants"
	"github.com/seldonio/seldon-core/operator/v2/pkg/utils"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// TODO: unify these helper functions as they do more or less the same thing

func handleLoadedExperiments(
	ctx context.Context, namespace string, s *SchedulerClient, grcpClient scheduler.SchedulerClient) {
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
			if _, err := s.StartExperiment(ctx, &experiment, grcpClient); err != nil {
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
