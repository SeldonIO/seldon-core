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
	"google.golang.org/grpc"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"

	"github.com/seldonio/seldon-core/operator/v2/apis/mlops/v1alpha1"
	"github.com/seldonio/seldon-core/operator/v2/pkg/constants"
	"github.com/seldonio/seldon-core/operator/v2/pkg/utils"
)

func (s *SchedulerClient) StartExperiment(ctx context.Context, experiment *v1alpha1.Experiment) (error, bool) {
	logger := s.logger.WithName("StartExperiment")
	conn, err := s.getConnection(experiment.Namespace)
	if err != nil {
		return err, true
	}
	grcpClient := scheduler.NewSchedulerClient(conn)
	req := &scheduler.StartExperimentRequest{
		Experiment: experiment.AsSchedulerExperimentRequest(),
	}
	logger.Info("Start", "experiment name", experiment.Name)
	_, err = grcpClient.StartExperiment(
		ctx,
		req,
		grpc_retry.WithMax(SchedulerConnectMaxRetries),
		grpc_retry.WithBackoff(grpc_retry.BackoffExponential(SchedulerConnectBackoffScalar)),
	)
	return err, s.checkErrorRetryable(experiment.Kind, experiment.Name, err)
}

func (s *SchedulerClient) StopExperiment(ctx context.Context, experiment *v1alpha1.Experiment) (error, bool) {
	logger := s.logger.WithName("StopExperiment")
	conn, err := s.getConnection(experiment.Namespace)
	if err != nil {
		return err, true
	}
	grcpClient := scheduler.NewSchedulerClient(conn)
	req := &scheduler.StopExperimentRequest{
		Name: experiment.Name,
	}
	logger.Info("Stop", "experiment name", experiment.Name)
	_, err = grcpClient.StopExperiment(
		ctx,
		req,
		grpc_retry.WithMax(SchedulerConnectMaxRetries),
		grpc_retry.WithBackoff(grpc_retry.BackoffExponential(SchedulerConnectBackoffScalar)),
	)
	return err, s.checkErrorRetryable(experiment.Kind, experiment.Name, err)
}

// namespace is not used in this function
func (s *SchedulerClient) SubscribeExperimentEvents(ctx context.Context, conn *grpc.ClientConn, namespace string) error {
	logger := s.logger.WithName("SubscribeExperimentEvents")
	grcpClient := scheduler.NewSchedulerClient(conn)

	stream, err := grcpClient.SubscribeExperimentStatus(ctx, &scheduler.ExperimentSubscriptionRequest{SubscriberName: "seldon manager"}, grpc_retry.WithMax(1))
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
		experiment := &v1alpha1.Experiment{}
		err = s.Get(ctx, client.ObjectKey{Name: event.ExperimentName, Namespace: event.KubernetesMeta.Namespace}, experiment)
		if err != nil {
			logger.Error(err, "Failed to get experiment", "name", event.ExperimentName, "namespace", event.KubernetesMeta.Namespace)
			continue
		}

		if !experiment.ObjectMeta.DeletionTimestamp.IsZero() {
			logger.Info("Experiment is pending deletion", "experiment", experiment.Name)
			if !event.Active {
				retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
					latestExperiment := &v1alpha1.Experiment{}
					err = s.Get(ctx, client.ObjectKey{Name: event.ExperimentName, Namespace: event.KubernetesMeta.Namespace}, latestExperiment)
					if err != nil {
						return err
					}
					if !latestExperiment.ObjectMeta.DeletionTimestamp.IsZero() { // Experiment is being deleted
						// remove finalizer now we have completed successfully
						latestExperiment.ObjectMeta.Finalizers = utils.RemoveStr(latestExperiment.ObjectMeta.Finalizers, constants.ExperimentFinalizerName)
						if err := s.Update(ctx, latestExperiment); err != nil {
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
		}

		// Try to update status
		retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			experiment := &v1alpha1.Experiment{}
			err = s.Get(ctx, client.ObjectKey{Name: event.ExperimentName, Namespace: event.KubernetesMeta.Namespace}, experiment)
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
			return s.updateExperimentStatus(experiment)
		})
		if retryErr != nil {
			logger.Error(err, "Failed to update status", "experiment", event.ExperimentName)
		}

	}
	return nil
}

func (s *SchedulerClient) updateExperimentStatus(experiment *v1alpha1.Experiment) error {
	if err := s.Status().Update(context.TODO(), experiment); err != nil {
		s.recorder.Eventf(experiment, v1.EventTypeWarning, "UpdateFailed",
			"Failed to update status for experiment %q: %v", experiment.Name, err)
		return err
	}
	return nil
}
