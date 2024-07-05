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

func (s *SchedulerClient) StartExperiment(ctx context.Context, experiment *v1alpha1.Experiment, conn *grpc.ClientConn) (bool, error) {
	logger := s.logger.WithName("StartExperiment")
	var err error
	if conn == nil {
		conn, err = s.getConnection(experiment.Namespace)
		if err != nil {
			return true, err
		}
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
	return s.checkErrorRetryable(experiment.Kind, experiment.Name, err), err
}

func (s *SchedulerClient) StopExperiment(ctx context.Context, experiment *v1alpha1.Experiment, conn *grpc.ClientConn) (bool, error) {
	logger := s.logger.WithName("StopExperiment")
	var err error
	if conn == nil {
		conn, err = s.getConnection(experiment.Namespace)
		if err != nil {
			return true, err
		}
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
	return s.checkErrorRetryable(experiment.Kind, experiment.Name, err), err
}

// namespace is not used in this function
func (s *SchedulerClient) SubscribeExperimentEvents(ctx context.Context, conn *grpc.ClientConn, namespace string) error {
	logger := s.logger.WithName("SubscribeExperimentEvents")
	grcpClient := scheduler.NewSchedulerClient(conn)

	stream, err := grcpClient.SubscribeExperimentStatus(ctx, &scheduler.ExperimentSubscriptionRequest{SubscriberName: "seldon manager"}, grpc_retry.WithMax(1))
	if err != nil {
		return err
	}

	// get experiments from the scheduler
	// if there are no experiments in the scheduler state then we need to create them
	// this is likely because of a restart of the scheduler that mnigrated the state
	// to v2 (where we delete the experiments from the scheduler state)
	numExperimentsFromScheduler, err := getNumExperimentsFromScheduler(ctx, grcpClient)
	if err != nil {
		return err
	}
	// if there are no experiments in the scheduler state then we need to create them if they exist in k8s
	// also remove finalizers from experiments that are being deleted
	if numExperimentsFromScheduler == 0 {
		s.handleLoadedExperiments(ctx, namespace, conn)
		s.handlePendingDeleteExperiments(ctx, namespace, conn)
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

func getNumExperimentsFromScheduler(ctx context.Context, grcpClient scheduler.SchedulerClient) (int, error) {
	req := &scheduler.ExperimentStatusRequest{
		SubscriberName: "seldon manager",
	}
	streamForExperimentStatuses, err := grcpClient.ExperimentStatus(ctx, req)
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

func (s *SchedulerClient) updateExperimentStatus(experiment *v1alpha1.Experiment) error {
	if err := s.Status().Update(context.TODO(), experiment); err != nil {
		s.recorder.Eventf(experiment, v1.EventTypeWarning, "UpdateFailed",
			"Failed to update status for experiment %q: %v", experiment.Name, err)
		return err
	}
	return nil
}

func (s *SchedulerClient) handleLoadedExperiments(
	ctx context.Context, namespace string, conn *grpc.ClientConn) {
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
			if _, err := s.StartExperiment(ctx, &experiment, conn); err != nil {
				// if this is a retryable error, we will retry on the next connection reconnect
				s.logger.Error(err, "Failed to call start experiment", "experiment", experiment.Name)
			} else {
				s.logger.V(1).Info("Start experiment called successfully", "experiment", experiment.Name)
			}
		} else {
			s.logger.V(1).Info("Experiment is being deleted, not starting", "experiment", experiment.Name)
		}
	}
}

func (s *SchedulerClient) handlePendingDeleteExperiments(
	ctx context.Context, namespace string, conn *grpc.ClientConn) {
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
			s.logger.V(1).Info("Calling Stop experiment (on reconnect)", "experiment", experiment.Name)
			if retryUnload, err := s.StopExperiment(ctx, &experiment, conn); err != nil {
				if retryUnload {
					// caller will retry as this method is called on connection reconnect
					s.logger.Error(err, "Failed to call stop experiment", "experiment", experiment.Name)
					continue
				} else {
					// this is essentially a failed pre-condition (experiment does not exist in scheduler)
					// we can remove
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
			} else {
				// if the experiment exists in the scheduler so we wait until we get the event from the subscription stream
				s.logger.Info("Stop experiment called successfully, not removing finalizer", "experiment", experiment.Name)
			}
		} else {
			s.logger.V(1).Info("Experiment is not being deleted, not unloading", "experiment", experiment.Name)
		}
	}
}
