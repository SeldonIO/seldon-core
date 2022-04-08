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

	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/seldonio/seldon-core/operatorv2/pkg/constants"
	"github.com/seldonio/seldon-core/operatorv2/pkg/utils"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	mlopsv1alpha1 "github.com/seldonio/seldon-core/operatorv2/apis/mlops/v1alpha1"
	scheduler "github.com/seldonio/seldon-core/operatorv2/scheduler"
)

// ExperimentReconciler reconciles a Experiment object
type ExperimentReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	Scheduler *scheduler.SchedulerClient
	Recorder  record.EventRecorder
}

func (r *ExperimentReconciler) handleFinalizer(ctx context.Context, experiment *mlopsv1alpha1.Experiment) (bool, error) {

	// Check if we are being deleted or not
	if experiment.ObjectMeta.DeletionTimestamp.IsZero() { // Not being deleted

		// Add our finalizer
		if !utils.ContainsStr(experiment.ObjectMeta.Finalizers, constants.ExperimentFinalizerName) {
			experiment.ObjectMeta.Finalizers = append(experiment.ObjectMeta.Finalizers, constants.ExperimentFinalizerName)
			if err := r.Update(context.Background(), experiment); err != nil {
				return true, err
			}
		}
	} else { // experiment is being deleted
		if utils.ContainsStr(experiment.ObjectMeta.Finalizers, constants.ExperimentFinalizerName) {
			// Handle unload in scheduler
			if err := r.Scheduler.StopExperiment(ctx, experiment); err != nil {
				return true, err
			}
		}
		// Stop reconciliation as the item is being deleted
		return true, nil
	}
	return false, nil
}

//+kubebuilder:rbac:groups=mlops.seldon.io,resources=experiments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=mlops.seldon.io,resources=experiments/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=mlops.seldon.io,resources=experiments/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Experiment object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *ExperimentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx).WithName("Reconcile")

	experiment := &mlopsv1alpha1.Experiment{}
	if err := r.Get(ctx, req.NamespacedName, experiment); err != nil {
		if errors.IsNotFound(err) {
			// we'll ignore not-found errors, since they can't be fixed by an immediate
			// requeue (we'll need to wait for a new notification), and we can get them
			// on deleted requests.
			return reconcile.Result{}, nil
		}
		logger.Error(err, "unable to fetch Experiment", "name", req.Name, "namespace", req.Namespace)
		return reconcile.Result{}, err
	}

	stop, err := r.handleFinalizer(ctx, experiment)
	if stop {
		return reconcile.Result{}, err
	}

	err = r.Scheduler.StartExperiment(ctx, experiment)
	if err != nil {
		return reconcile.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ExperimentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	pred := predicate.GenerationChangedPredicate{}
	return ctrl.NewControllerManagedBy(mgr).
		For(&mlopsv1alpha1.Experiment{}).
		WithEventFilter(pred).
		Complete(r)
}
