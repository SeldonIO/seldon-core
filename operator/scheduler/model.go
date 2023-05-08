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
	"fmt"
	"io"

	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"
	"github.com/seldonio/seldon-core/operator/v2/apis/mlops/v1alpha1"
	"github.com/seldonio/seldon-core/operator/v2/pkg/constants"
	"github.com/seldonio/seldon-core/operator/v2/pkg/utils"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"knative.dev/pkg/apis"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (s *SchedulerClient) LoadModel(ctx context.Context, model *v1alpha1.Model) (error, bool) {
	logger := s.logger.WithName("LoadModel")
	conn, err := s.getConnection(model.Namespace)
	if err != nil {
		return err, true
	}
	grcpClient := scheduler.NewSchedulerClient(conn)

	md, err := model.AsSchedulerModel()
	if err != nil {
		return err, false
	}
	loadModelRequest := scheduler.LoadModelRequest{
		Model: md,
	}

	logger.Info("Load", "model name", model.Name)
	_, err = grcpClient.LoadModel(ctx, &loadModelRequest, grpc_retry.WithMax(2))
	if err != nil {
		return err, s.checkErrorRetryable(model.Kind, model.Name, err)
	}
	return nil, false
}

func (s *SchedulerClient) UnloadModel(ctx context.Context, model *v1alpha1.Model) (error, bool) {
	logger := s.logger.WithName("UnloadModel")
	conn, err := s.getConnection(model.Namespace)
	if err != nil {
		return err, true
	}
	grcpClient := scheduler.NewSchedulerClient(conn)

	modelRef := &scheduler.UnloadModelRequest{
		Model: &scheduler.ModelReference{
			Name: model.Name,
		},
		KubernetesMeta: &scheduler.KubernetesMeta{
			Namespace:  model.Namespace,
			Generation: model.Generation,
		},
	}
	logger.Info("Unload", "model name", model.Name)
	_, err = grcpClient.UnloadModel(ctx, modelRef, grpc_retry.WithMax(2))
	if err != nil {
		return err, s.checkErrorRetryable(model.Kind, model.Name, err)
	}
	return nil, false
}

func (s *SchedulerClient) SubscribeModelEvents(ctx context.Context, namespace string) error {
	logger := s.logger.WithName("SubscribeModelEvents")
	conn, err := s.getConnection(namespace)
	if err != nil {
		return err
	}
	grcpClient := scheduler.NewSchedulerClient(conn)

	stream, err := grcpClient.SubscribeModelStatus(ctx, &scheduler.ModelSubscriptionRequest{SubscriberName: "seldon manager"}, grpc_retry.WithMax(1))
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
		// The expected contract is just the latest version will be sent to us
		if len(event.Versions) < 1 {
			logger.Info("Expected a single model version", "numVersions", len(event.Versions), "name", event.GetModelName())
			continue
		}
		latestVersionStatus := event.Versions[0]
		if latestVersionStatus.GetKubernetesMeta() == nil {
			logger.Info("Ignoring event with no Kubernetes metadata.", "model", event.ModelName)
			continue
		}
		logger.Info("Received event", "name", event.ModelName, "version", latestVersionStatus.Version, "generation", latestVersionStatus.GetKubernetesMeta().Generation, "state", latestVersionStatus.State.State.String(), "reason", latestVersionStatus.State.Reason)

		// Handle terminated event to remove finalizer
		if canRemoveFinalizer(latestVersionStatus.State.State) {
			retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
				latestModel := &v1alpha1.Model{}
				err = s.Get(ctx, client.ObjectKey{Name: event.ModelName, Namespace: latestVersionStatus.GetKubernetesMeta().Namespace}, latestModel)
				if err != nil {
					return err
				}
				if !latestModel.ObjectMeta.DeletionTimestamp.IsZero() { // Model is being deleted
					// remove finalizer now we have completed successfully
					latestModel.ObjectMeta.Finalizers = utils.RemoveStr(latestModel.ObjectMeta.Finalizers, constants.ModelFinalizerName)
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
			err = s.Get(ctx, client.ObjectKey{Name: event.ModelName, Namespace: latestVersionStatus.GetKubernetesMeta().Namespace}, latestModel)
			if err != nil {
				return err
			}
			if latestVersionStatus.GetKubernetesMeta().Generation != latestModel.Generation {
				logger.Info("Ignoring event for old generation", "currentGeneration", latestModel.Generation, "eventGeneration", latestVersionStatus.GetKubernetesMeta().Generation, "model", event.ModelName)
				return nil
			}
			if !latestModel.ObjectMeta.DeletionTimestamp.IsZero() { // Model is being deleted
				return nil
			}
			// Handle status update
			switch latestVersionStatus.State.State {
			case scheduler.ModelStatus_ModelAvailable:
				logger.Info("Setting model to ready", "name", event.ModelName, "state", latestVersionStatus.State.State.String())
				latestModel.Status.CreateAndSetCondition(v1alpha1.ModelReady, true, latestVersionStatus.State.Reason)
			default:
				logger.Info("Setting model to not ready", "name", event.ModelName, "state", latestVersionStatus.State.State.String())
				latestModel.Status.CreateAndSetCondition(v1alpha1.ModelReady, false, latestVersionStatus.State.Reason)
			}
			// Set the total number of replicas targeted by this model
			latestModel.Status.Replicas = int32(latestVersionStatus.State.GetAvailableReplicas() + latestVersionStatus.State.GetUnavailableReplicas())
			return s.updateModelStatus(latestModel)
		})
		if retryErr != nil {
			logger.Error(err, "Failed to update status", "model", event.ModelName)
		}

	}
	return nil
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
			s.recorder.Eventf(model, v1.EventTypeWarning, "UpdateFailed",
				"Failed to update status for Model %q: %v", model.Name, err)
			return err
		} else {
			currentIsReady := modelReady(model.Status)
			if prevWasReady && !currentIsReady {
				s.recorder.Eventf(model, v1.EventTypeWarning, "ModelNotReady",
					fmt.Sprintf("Model [%v] is no longer Ready", model.GetName()))
			} else if !prevWasReady && currentIsReady {
				s.recorder.Eventf(model, v1.EventTypeNormal, "ModelReady",
					fmt.Sprintf("Model [%v] is Ready", model.GetName()))
			}
		}
	}
	return nil
}
