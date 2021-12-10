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
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/seldonio/seldon-core/operatorv2/pkg/utils"
	scheduler "github.com/seldonio/seldon-core/operatorv2/scheduler"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	mlopsv1alpha1 "github.com/seldonio/seldon-core/operatorv2/apis/mlops/v1alpha1"
)

// ModelReconciler reconciles a Model object
type ModelReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	Scheduler *scheduler.SchedulerClient
	Recorder  record.EventRecorder
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

	finalizerName := "seldon.model.finalizer"
	// Check if we are being deleted or not
	if model.ObjectMeta.DeletionTimestamp.IsZero() { // Not being deleted

		// Add our finalizer
		if !utils.ContainsStr(model.ObjectMeta.Finalizers, finalizerName) {
			model.ObjectMeta.Finalizers = append(model.ObjectMeta.Finalizers, finalizerName)
			if err := r.Update(context.Background(), model); err != nil {
				return ctrl.Result{}, err
			}
		}
	} else { // model is being deleted
		if utils.ContainsStr(model.ObjectMeta.Finalizers, finalizerName) {
			// Handel unloadin scheduler
			if err := r.Scheduler.UnloadModel(ctx, model); err != nil {
				return ctrl.Result{}, err
			}

			// remove finalizer now we have completed successfully
			model.ObjectMeta.Finalizers = utils.RemoveStr(model.ObjectMeta.Finalizers, finalizerName)
			if err := r.Update(context.Background(), model); err != nil {
				return ctrl.Result{}, err
			}
		}

		// Stop reconciliation as the item is being deleted
		return ctrl.Result{}, nil
	}

	err := r.Scheduler.LoadModel(ctx, model)
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *ModelReconciler) updateStatus(model *mlopsv1alpha1.Model) error {
	existingModel := &mlopsv1alpha1.Model{}
	namespacedName := types.NamespacedName{Name: model.Name, Namespace: model.Namespace}
	if err := r.Get(context.TODO(), namespacedName, existingModel); err != nil {
		return err
	}
	//ready := modelReady(existingModel.Status)
	if equality.Semantic.DeepEqual(existingModel.Status, model.Status) {
		// Do nothing
	} else if err := r.Status().Update(context.TODO(), model); err != nil {
		//r.l.Error(err, "Failed to update InferenceService status", "InferenceService", model.Name)
		//r.Recorder.Eventf(model, v1.EventTypeWarning, "UpdateFailed",
		//	"Failed to update status for InferenceService %q: %v", desiredService.Name, err)
		return err
	} else {
		// If there was a difference and there was no error.
		//isReady := modelReady(model.Status)
		//if ready && !isReady { // Moved to NotReady State
		//	r.Recorder.Eventf(desiredService, v1.EventTypeWarning, string(InferenceServiceNotReadyState),
		//		fmt.Sprintf("InferenceService [%v] is no longer Ready", desiredService.GetName()))
		//} else if !wasReady && isReady { // Moved to Ready State
		//	r.Recorder.Eventf(desiredService, v1.EventTypeNormal, string(InferenceServiceReadyState),
		//		fmt.Sprintf("InferenceService [%v] is Ready", desiredService.GetName()))
		//}
	}
	return nil
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
