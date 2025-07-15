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
	mlopsv1alpha1 "github.com/seldonio/seldon-core/operator/v2/apis/mlops/v1alpha1"
	"github.com/seldonio/seldon-core/operator/v2/pkg/constants"
	"github.com/seldonio/seldon-core/operator/v2/pkg/utils"
	scheduler "github.com/seldonio/seldon-core/operator/v2/scheduler"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// ExperimentReconciler reconciles a Experiment object
type ExperimentReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	Scheduler scheduler.Client
	Recorder  record.EventRecorder
}

func (r *ExperimentReconciler) handleFinalizer(ctx context.Context, logger logr.Logger, experiment *mlopsv1alpha1.Experiment) (bool, error) {

	// Check if we are being deleted or not
	if experiment.ObjectMeta.DeletionTimestamp.IsZero() { // Not being deleted

		// Add our finalizer
		if !utils.ContainsStr(experiment.ObjectMeta.Finalizers, constants.ExperimentFinalizerName) {
			experiment.ObjectMeta.Finalizers = append(experiment.ObjectMeta.Finalizers, constants.ExperimentFinalizerName)
			if err := r.Update(ctx, experiment); err != nil {
				return true, err
			}
		}
	} else { // experiment is being deleted
		if utils.ContainsStr(experiment.ObjectMeta.Finalizers, constants.ExperimentFinalizerName) {
			// Handle unload in scheduler
			if retry, err := r.Scheduler.StopExperiment(ctx, experiment, nil); err != nil {
				if retry {
					return true, err
				} else {
					experiment.ObjectMeta.Finalizers = utils.RemoveStr(experiment.ObjectMeta.Finalizers, constants.ExperimentFinalizerName)
					if errUpdate := r.Update(ctx, experiment); errUpdate != nil {
						logger.Error(err, "Failed to remove finalizer", "experiment", experiment.Name)
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
	ctx, cancel := context.WithTimeout(ctx, constants.ReconcileTimeout)
	defer cancel()

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

	stop, err := r.handleFinalizer(ctx, logger, experiment)
	if stop {
		return reconcile.Result{}, err
	}

	retry, err := r.Scheduler.StartExperiment(ctx, experiment, nil)
	if err != nil {
		r.updateStatusFromError(ctx, logger, experiment, err)
		if retry {
			return ctrl.Result{}, err
		} else {
			return ctrl.Result{}, nil
		}

	}

	return ctrl.Result{}, nil
}

func (r *ExperimentReconciler) updateStatusFromError(ctx context.Context, logger logr.Logger, experiment *mlopsv1alpha1.Experiment, err error) {
	experiment.Status.CreateAndSetCondition(mlopsv1alpha1.ModelReady, false, err.Error())
	if errSet := r.Status().Update(ctx, experiment); errSet != nil {
		logger.Error(errSet, "Failed to set status for experiment on error", "model", experiment.Name, "error", err.Error())
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *ExperimentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	pred := predicate.GenerationChangedPredicate{}
	return ctrl.NewControllerManagedBy(mgr).
		For(&mlopsv1alpha1.Experiment{}).
		WithEventFilter(pred).
		Complete(r)
}
