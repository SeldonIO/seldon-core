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

func (s *SchedulerClient) StartExperiment(ctx context.Context, experiment *v1alpha1.Experiment, grpcClient scheduler.SchedulerClient) (bool, error) {
	logger := s.logger.WithName("StartExperiment")
	var err error
	if grpcClient == nil {
		conn, err := s.getConnection(experiment.Namespace)
		if err != nil {
			return true, err
		}
		grpcClient = scheduler.NewSchedulerClient(conn)
	}

	req := &scheduler.StartExperimentRequest{
		Experiment: experiment.AsSchedulerExperimentRequest(),
	}
	logger.Info("Start", "experiment name", experiment.Name)
	_, err = grpcClient.StartExperiment(
		ctx,
		req,
		grpc_retry.WithMax(schedulerConnectMaxRetries),
		grpc_retry.WithBackoff(grpc_retry.BackoffExponential(schedulerConnectBackoffScalar)),
	)
	return s.checkErrorRetryable(experiment.Kind, experiment.Name, err), err
}

func (s *SchedulerClient) StopExperiment(ctx context.Context, experiment *v1alpha1.Experiment, grpcClient scheduler.SchedulerClient) (bool, error) {
	logger := s.logger.WithName("StopExperiment")
	var err error
	if grpcClient == nil {
		conn, err := s.getConnection(experiment.Namespace)
		if err != nil {
			return true, err
		}
		grpcClient = scheduler.NewSchedulerClient(conn)
	}
	req := &scheduler.StopExperimentRequest{
		Name: experiment.Name,
	}
	logger.Info("Stop", "experiment name", experiment.Name)
	_, err = grpcClient.StopExperiment(
		ctx,
		req,
		grpc_retry.WithMax(schedulerConnectMaxRetries),
		grpc_retry.WithBackoff(grpc_retry.BackoffExponential(schedulerConnectBackoffScalar)),
	)
	return s.checkErrorRetryable(experiment.Kind, experiment.Name, err), err
}

// namespace is not used in this function
func (s *SchedulerClient) SubscribeExperimentEvents(ctx context.Context, grpcClient scheduler.SchedulerClient, namespace string) error {
	logger := s.logger.WithName("SubscribeExperimentEvents")

	stream, err := grpcClient.SubscribeExperimentStatus(
		ctx,
		&scheduler.ExperimentSubscriptionRequest{SubscriberName: "seldon manager"},
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

		logger.Info("Received event", "experiment", event.ExperimentName)

		if event.GetKubernetesMeta() == nil {
			logger.Info("Received experiment event with no k8s metadata so ignoring", "Experiment", event.ExperimentName)
			continue
		}

		if !event.Active {
			retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
				ctxWithTimeout, cancel := context.WithTimeout(ctx, constants.K8sAPICallsTxTimeout)
				defer cancel()

				latestExperiment := &v1alpha1.Experiment{}
				err = s.Get(ctxWithTimeout, client.ObjectKey{Name: event.ExperimentName, Namespace: event.KubernetesMeta.Namespace}, latestExperiment)
				if err != nil {
					return err
				}
				if !latestExperiment.ObjectMeta.DeletionTimestamp.IsZero() { // Experiment is being deleted
					// remove finalizer now we have completed successfully
					latestExperiment.ObjectMeta.Finalizers = utils.RemoveStr(latestExperiment.ObjectMeta.Finalizers, constants.ExperimentFinalizerName)
					if err := s.Update(ctxWithTimeout, latestExperiment); err != nil {
						logger.Error(err, "Failed to remove finalizer", "experiment", latestExperiment.GetName())
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
				ctxWithTimeout, cancel := context.WithTimeout(ctx, constants.K8sAPICallsTxTimeout)
				defer cancel()

				experiment := &v1alpha1.Experiment{}
				err = s.Get(ctxWithTimeout, client.ObjectKey{Name: event.ExperimentName, Namespace: event.KubernetesMeta.Namespace}, experiment)
				if err != nil {
					return err
				}
				if event.KubernetesMeta.Generation != experiment.Generation {
					logger.Info("Ignoring event for old generation", "currentGeneration", experiment.Generation, "eventGeneration", event.KubernetesMeta.Generation, "server", event.ExperimentName)
					return nil
				}
				// Handle status update
				if event.Active {
					logger.Info("Setting experiment to ready", "experiment", event.ExperimentName)
					experiment.Status.CreateAndSetCondition(v1alpha1.ExperimentReady, true, event.StatusDescription)
				} else {
					logger.Info("Setting experiment to not ready", "experiment", event.ExperimentName)
					experiment.Status.CreateAndSetCondition(v1alpha1.ExperimentReady, false, event.StatusDescription)
				}
				if event.CandidatesReady {
					experiment.Status.CreateAndSetCondition(v1alpha1.CandidatesReady, true, "Candidates ready")
				} else {
					experiment.Status.CreateAndSetCondition(v1alpha1.CandidatesReady, false, "Candidates not ready")
				}
				if event.MirrorReady {
					experiment.Status.CreateAndSetCondition(v1alpha1.MirrorReady, true, "Mirror ready")
				} else {
					experiment.Status.CreateAndSetCondition(v1alpha1.MirrorReady, false, "Mirror not ready")
				}
				return s.updateExperimentStatus(ctxWithTimeout, experiment)
			})
			if retryErr != nil {
				logger.Error(err, "Failed to update status", "experiment", event.ExperimentName)
			}
		}

	}
	return nil
}

func (s *SchedulerClient) updateExperimentStatus(ctx context.Context, experiment *v1alpha1.Experiment) error {
	if err := s.Status().Update(ctx, experiment); err != nil {
		s.recorder.Eventf(experiment, v1.EventTypeWarning, "UpdateFailed",
			"Failed to update status for experiment %q: %v", experiment.Name, err)
		return err
	}
	return nil
}
