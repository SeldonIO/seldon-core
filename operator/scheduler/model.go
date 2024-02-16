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
	"fmt"
	"io"

	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"google.golang.org/grpc"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"knative.dev/pkg/apis"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"

	"github.com/seldonio/seldon-core/operator/v2/apis/mlops/v1alpha1"
	"github.com/seldonio/seldon-core/operator/v2/pkg/constants"
	"github.com/seldonio/seldon-core/operator/v2/pkg/utils"
)

func (s *SchedulerClient) LoadModel(ctx context.Context, model *v1alpha1.Model) (error, bool) {
	logger := s.logger.WithName("LoadModel")
	conn, err := s.getConnection(model.Namespace)
	if err != nil {
		return err, true
	}
	grcpClient := scheduler.NewSchedulerClient(conn)
	logger.Info("Load", "model name", model.Name)
	md, err := model.AsSchedulerModel()
	if err != nil {
		return err, false
	}
	loadModelRequest := scheduler.LoadModelRequest{
		Model: md,
	}

	_, err = grcpClient.LoadModel(
		ctx,
		&loadModelRequest,
		grpc_retry.WithMax(SchedulerConnectMaxRetries),
		grpc_retry.WithBackoff(grpc_retry.BackoffExponential(SchedulerConnectBackoffScalar)),
	)
	if err != nil {
		return err, s.checkErrorRetryable(model.Kind, model.Name, err)
	}

	return nil, false
}


// UnloadModel unloads a model from the scheduler
// If the connection is not provided, get a new one
// In the case of errors we check if the error is retryable and return a boolean to indicate if the error is retryable
// For the cases we think we should retry, check logic in `checkErrorRetryable`
func (s *SchedulerClient) UnloadModel(ctx context.Context, model *v1alpha1.Model, conn *grpc.ClientConn) (bool, error) {
	logger := s.logger.WithName("UnloadModel")

	// If the connection is not provided, get a new one
	var err error
	retryableError := false
	if conn == nil {
		conn, err = s.getConnection(model.Namespace)
		if err != nil {
			retryableError = true
			return retryableError, err
		}
	}
	logger.Info("Unload", "model name", model.Name)
	modelRef := &scheduler.UnloadModelRequest{
		Model: &scheduler.ModelReference{
			Name: model.Name,
		},
		KubernetesMeta: &scheduler.KubernetesMeta{
			Namespace:  model.Namespace,
			Generation: model.Generation,
		},
	}

	grcpClient := scheduler.NewSchedulerClient(conn)
	_, err = grcpClient.UnloadModel(
		ctx,
		modelRef,
		grpc_retry.WithMax(SchedulerConnectMaxRetries),
		grpc_retry.WithBackoff(grpc_retry.BackoffExponential(SchedulerConnectBackoffScalar)),
	)
	if err != nil {
		return s.checkErrorRetryable(model.Kind, model.Name, err), err
	}
	return retryableError, nil
}

func (s *SchedulerClient) SubscribeModelEvents(ctx context.Context, conn *grpc.ClientConn, namespace string) error {
	logger := s.logger.WithName("SubscribeModelEvents")
	grcpClient := scheduler.NewSchedulerClient(conn)

	stream, err := grcpClient.SubscribeModelStatus(
		ctx,
		&scheduler.ModelSubscriptionRequest{SubscriberName: "seldon manager"},
		grpc_retry.WithMax(SchedulerConnectMaxRetries),
		grpc_retry.WithBackoff(grpc_retry.BackoffExponential(SchedulerConnectBackoffScalar)),
	)
	if err != nil {
		return err
	}

	// on new reconnects check if we have models that are stuck in deletion and therefore we need to reconcile their states
	go s.handlePendingDeleteModels(ctx, namespace, conn)

	for {
		event, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			logger.Error(err, "event recv failed")
			return err
		}

		// The expected contract is just the latest version will be sent to us
		if len(event.Versions) < 1 {
			logger.Info(
				"Expected a single model version",
				"numVersions", len(event.Versions),
				"name", event.GetModelName(),
			)
			continue
		}
		latestVersionStatus := event.Versions[0]
		if latestVersionStatus.GetKubernetesMeta() == nil {
			logger.Info("Ignoring event with no Kubernetes metadata.", "model", event.ModelName)
			continue
		}

		logger.Info(
			"Received event",
			"name", event.ModelName,
			"version", latestVersionStatus.Version,
			"generation", latestVersionStatus.GetKubernetesMeta().Generation,
			"state", latestVersionStatus.State.State.String(),
			"reason", latestVersionStatus.State.Reason,
		)

		// Handle terminated event to remove finalizer
		if canRemoveFinalizer(latestVersionStatus.State.State) {
			retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
				latestModel := &v1alpha1.Model{}

				err = s.Get(
					ctx,
					client.ObjectKey{
						Name:      event.ModelName,
						Namespace: latestVersionStatus.GetKubernetesMeta().Namespace,
					},
					latestModel,
				)
				if err != nil {
					return err
				}

				if !latestModel.ObjectMeta.DeletionTimestamp.IsZero() { // Model is being deleted
					// remove finalizer now we have completed successfully
					latestModel.ObjectMeta.Finalizers = utils.RemoveStr(
						latestModel.ObjectMeta.Finalizers,
						constants.ModelFinalizerName,
					)
					if err := s.Update(ctx, latestModel); err != nil {
						logger.Error(err, "Failed to remove finalizer", "model", latestModel.GetName())
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
		retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			latestModel := &v1alpha1.Model{}

			err = s.Get(
				ctx,
				client.ObjectKey{
					Name:      event.ModelName,
					Namespace: latestVersionStatus.GetKubernetesMeta().Namespace,
				},
				latestModel,
			)
			if err != nil {
				return err
			}

			if latestVersionStatus.GetKubernetesMeta().Generation != latestModel.Generation {
				logger.Info(
					"Ignoring event for old generation",
					"currentGeneration", latestModel.Generation,
					"eventGeneration", latestVersionStatus.GetKubernetesMeta().Generation,
					"model", event.ModelName,
				)
				return nil
			}

			// Handle status update
			modelStatus := latestVersionStatus.GetState()
			switch modelStatus.GetState() {
			case scheduler.ModelStatus_ModelAvailable:
				logger.Info(
					"Setting model to ready",
					"name", event.ModelName,
					"state", modelStatus.GetState().String(),
				)
				latestModel.Status.CreateAndSetCondition(
					v1alpha1.ModelReady,
					true,
					modelStatus.GetState().String(),
					modelStatus.GetReason(),
				)
			default:
				logger.Info(
					"Setting model to not ready",
					"name", event.ModelName,
					"state", modelStatus.GetState().String(),
				)
				latestModel.Status.CreateAndSetCondition(
					v1alpha1.ModelReady,
					false,
					modelStatus.GetState().String(),
					modelStatus.GetReason(),
				)
			}

			// Set the total number of replicas targeted by this model
			latestModel.Status.Replicas = int32(
				modelStatus.GetAvailableReplicas() +
					modelStatus.GetUnavailableReplicas(),
			)
			return s.updateModelStatus(latestModel)
		})
		if retryErr != nil {
			logger.Error(err, "Failed to update status", "model", event.ModelName)
		}

	}
	return nil
}

func (s *SchedulerClient) handlePendingDeleteModels(
	ctx context.Context, namespace string, conn *grpc.ClientConn) {
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
			if retryUnload, err := s.UnloadModel(ctx, &model, conn); err != nil {
				if retryUnload {
					s.logger.Info("Failed to call unload model", "model", model.Name)
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
			break
		}
	}
}

func canRemoveFinalizer(state scheduler.ModelStatus_ModelState) bool {
	switch state {
	case scheduler.ModelStatus_ModelTerminated,
		scheduler.ModelStatus_ModelTerminateFailed,
		scheduler.ModelStatus_ModelFailed,
		scheduler.ModelStatus_ModelStateUnknown,
		scheduler.ModelStatus_ScheduleFailed:
		return true
	default:
		return false
	}
}

func modelReady(status v1alpha1.ModelStatus) bool {
	return status.Conditions != nil &&
		status.GetCondition(apis.ConditionReady) != nil &&
		status.GetCondition(apis.ConditionReady).Status == v1.ConditionTrue
}

func (s *SchedulerClient) updateModelStatus(model *v1alpha1.Model) error {
	existingModel := &v1alpha1.Model{}
	namespacedName := types.NamespacedName{Name: model.Name, Namespace: model.Namespace}

	if err := s.Get(context.TODO(), namespacedName, existingModel); err != nil {
		if errors.IsNotFound(err) { //Ignore NotFound errors
			return nil
		}
		return err
	}

	prevWasReady := modelReady(existingModel.Status)
	if equality.Semantic.DeepEqual(existingModel.Status, model.Status) {
		// Not updating as no difference
	} else {
		if err := s.Status().Update(context.TODO(), model); err != nil {
			s.recorder.Eventf(
				model,
				v1.EventTypeWarning,
				"UpdateFailed",
				"Failed to update status for Model %q: %v",
				model.Name,
				err,
			)
			return err
		} else {
			currentIsReady := modelReady(model.Status)
			if prevWasReady && !currentIsReady {
				s.recorder.Eventf(
					model,
					v1.EventTypeWarning,
					"ModelNotReady",
					fmt.Sprintf("Model [%v] is no longer Ready", model.GetName()),
				)
			} else if !prevWasReady && currentIsReady {
				s.recorder.Eventf(
					model,
					v1.EventTypeNormal,
					"ModelReady",
					fmt.Sprintf("Model [%v] is Ready", model.GetName()),
				)
			}
		}
	}
	return nil
}
