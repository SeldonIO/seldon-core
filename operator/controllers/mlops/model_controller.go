/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package mlops

import (
	"context"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	schedulerAPI "github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"

	mlopsv1alpha1 "github.com/seldonio/seldon-core/operator/v2/apis/mlops/v1alpha1"
	"github.com/seldonio/seldon-core/operator/v2/pkg/constants"
	"github.com/seldonio/seldon-core/operator/v2/pkg/utils"
	scheduler "github.com/seldonio/seldon-core/operator/v2/scheduler"
)

// ModelReconciler reconciles a Model object
type ModelReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	Scheduler *scheduler.SchedulerClient
	Recorder  record.EventRecorder
}

func (r *ModelReconciler) handleFinalizer(ctx context.Context, logger logr.Logger, model *mlopsv1alpha1.Model) (bool, error) {

	// Check if we are being deleted or not
	if model.ObjectMeta.DeletionTimestamp.IsZero() { // Not being deleted

		// Add our finalizer
		if !utils.ContainsStr(model.ObjectMeta.Finalizers, constants.ModelFinalizerName) {
			model.ObjectMeta.Finalizers = append(model.ObjectMeta.Finalizers, constants.ModelFinalizerName)
			if err := r.Update(context.Background(), model); err != nil {
				return true, err
			}
		}
	} else { // model is being deleted
		if utils.ContainsStr(model.ObjectMeta.Finalizers, constants.ModelFinalizerName) {
			// Handle unload in scheduler
			if retryUnload, err := r.Scheduler.UnloadModel(ctx, model, nil); err != nil {
				if retryUnload {
					return true, err
				} else {
					model.ObjectMeta.Finalizers = utils.RemoveStr(model.ObjectMeta.Finalizers, constants.ModelFinalizerName)
					if errUpdate := r.Update(ctx, model); errUpdate != nil {
						logger.Error(err, "Failed to remove finalizer", "model", model.Name)
						return true, err
					}
				}
			}
		}
		// Stop reconciliation as the item is being deleted
		return true, nil
	}
	return false, nil
}

//+kubebuilder:rbac:groups=mlops.seldon.io,resources=models,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=mlops.seldon.io,resources=models/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=mlops.seldon.io,resources=models/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch

func (r *ModelReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx).WithName("Reconcile")

	model := &mlopsv1alpha1.Model{}
	if err := r.Get(ctx, req.NamespacedName, model); err != nil {
		if errors.IsNotFound(err) {
			// we'll ignore not-found errors, since they can't be fixed by an immediate
			// requeue (we'll need to wait for a new notification), and we can get them
			// on deleted requests.
			return reconcile.Result{}, nil
		}
		logger.Error(err, "unable to fetch Model")
		return reconcile.Result{}, err
	}

	stop, err := r.handleFinalizer(ctx, logger, model)
	if stop {
		return reconcile.Result{}, err
	}

	retry, err := r.Scheduler.LoadModel(ctx, model, nil)
	if err != nil {
		r.updateStatusFromError(ctx, logger, model, retry, err)
		if retry {
			return ctrl.Result{}, err
		} else {
			return ctrl.Result{}, nil
		}
	}

	return ctrl.Result{}, nil
}

func (r *ModelReconciler) updateStatusFromError(
	ctx context.Context,
	logger logr.Logger,
	model *mlopsv1alpha1.Model,
	canRetry bool,
	err error,
) {
	modelStatus := schedulerAPI.ModelStatus_ModelFailed.String()
	if canRetry {
		modelStatus = schedulerAPI.ModelStatus_ModelProgressing.String()
	}

	model.Status.CreateAndSetCondition(mlopsv1alpha1.ModelReady, false, modelStatus, err.Error())
	if errSet := r.Status().Update(ctx, model); errSet != nil {
		logger.Error(errSet, "Failed to set status for model on error", "model", model.Name, "error", err.Error())
	}
}

// SetupWithManager sets up the controller with the Manager.
// Uses https://github.com/kubernetes-sigs/kubebuilder/issues/618#issuecomment-698018831
// This ensures we don't reconcile when just the status is updated by checking if generation changed
func (r *ModelReconciler) SetupWithManager(mgr ctrl.Manager) error {
	pred := predicate.GenerationChangedPredicate{}
	return ctrl.NewControllerManagedBy(mgr).
		For(&mlopsv1alpha1.Model{}).
		WithEventFilter(pred).
		Complete(r)
}
