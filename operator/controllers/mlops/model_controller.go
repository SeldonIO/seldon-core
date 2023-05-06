/*
Copyright 2021.

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

package mlops

import (
	"context"

	"github.com/go-logr/logr"

	"github.com/seldonio/seldon-core/operator/v2/pkg/constants"

	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/seldonio/seldon-core/operator/v2/pkg/utils"
	scheduler "github.com/seldonio/seldon-core/operator/v2/scheduler"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	mlopsv1alpha1 "github.com/seldonio/seldon-core/operator/v2/apis/mlops/v1alpha1"
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
			if err, retry := r.Scheduler.UnloadModel(ctx, model); err != nil {
				if retry {
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

//+kubebuilder:rbac:groups=mlops.seldon.io,namespace=seldon-mesh,resources=models,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=mlops.seldon.io,namespace=seldon-mesh,resources=models/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=mlops.seldon.io,namespace=seldon-mesh,resources=models/finalizers,verbs=update
//+kubebuilder:rbac:groups="",namespace=seldon-mesh,resources=events,verbs=create;patch

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

	err, retry := r.Scheduler.LoadModel(ctx, model)
	if err != nil {
		r.updateStatusFromError(ctx, logger, model, err)
		if retry {
			return ctrl.Result{}, err
		} else {
			return ctrl.Result{}, nil
		}
	}

	return ctrl.Result{}, nil
}

func (r *ModelReconciler) updateStatusFromError(ctx context.Context, logger logr.Logger, model *mlopsv1alpha1.Model, err error) {
	model.Status.CreateAndSetCondition(mlopsv1alpha1.ModelReady, false, err.Error())
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
