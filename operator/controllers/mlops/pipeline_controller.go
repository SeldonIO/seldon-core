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

// PipelineReconciler reconciles a Pipeline object
type PipelineReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	Scheduler *scheduler.SchedulerClient
	Recorder  record.EventRecorder
}

func (r *PipelineReconciler) handleFinalizer(
	ctx context.Context,
	logger logr.Logger,
	pipeline *mlopsv1alpha1.Pipeline,
) (bool, error) {
	// Check if we are being deleted or not
	if pipeline.ObjectMeta.DeletionTimestamp.IsZero() { // Not being deleted

		// Add our finalizer
		if !utils.ContainsStr(pipeline.ObjectMeta.Finalizers, constants.PipelineFinalizerName) {
			pipeline.ObjectMeta.Finalizers = append(pipeline.ObjectMeta.Finalizers, constants.PipelineFinalizerName)
			if err := r.Update(context.Background(), pipeline); err != nil {
				return true, err
			}
		}
	} else { // pipeline is being deleted
		if utils.ContainsStr(pipeline.ObjectMeta.Finalizers, constants.PipelineFinalizerName) {
			// Handle unload in scheduler
			if err, retry := r.Scheduler.UnloadPipeline(ctx, pipeline); err != nil {
				if retry {
					return true, err
				} else {
					// Remove pipeline anyway on error as we assume errors from scheduler are fatal here
					pipeline.ObjectMeta.Finalizers = utils.RemoveStr(
						pipeline.ObjectMeta.Finalizers,
						constants.PipelineFinalizerName,
					)
					if errUpdate := r.Update(ctx, pipeline); errUpdate != nil {
						logger.Error(err, "Failed to remove finalizer", "pipeline", pipeline.Name)
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

//+kubebuilder:rbac:groups=mlops.seldon.io,resources=pipelines,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=mlops.seldon.io,resources=pipelines/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=mlops.seldon.io,resources=pipelines/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Pipeline object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *PipelineReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx).WithName("Reconcile")

	pipeline := &mlopsv1alpha1.Pipeline{}
	if err := r.Get(ctx, req.NamespacedName, pipeline); err != nil {
		if errors.IsNotFound(err) {
			// we'll ignore not-found errors, since they can't be fixed by an immediate
			// requeue (we'll need to wait for a new notification), and we can get them
			// on deleted requests.
			return reconcile.Result{}, nil
		}
		logger.Error(err, "unable to fetch Pipeline", "name", req.Name, "namespace", req.Namespace)
		return reconcile.Result{}, err
	}

	stop, err := r.handleFinalizer(ctx, logger, pipeline)
	if stop {
		return reconcile.Result{}, err
	}

	err, retry := r.Scheduler.LoadPipeline(ctx, pipeline)
	if err != nil {
		r.updateStatusFromError(ctx, logger, pipeline, retry, err)
		if retry {
			return ctrl.Result{}, err
		} else {
			return ctrl.Result{}, nil
		}
	}
	return ctrl.Result{}, nil
}

func (r *PipelineReconciler) updateStatusFromError(
	ctx context.Context,
	logger logr.Logger,
	pipeline *mlopsv1alpha1.Pipeline,
	canRetry bool,
	err error,
) {
	pipelineStatus := schedulerAPI.PipelineVersionState_PipelineFailed.String()
	if canRetry {
		pipelineStatus = schedulerAPI.PipelineVersionState_PipelineCreating.String()
	}

	pipeline.Status.CreateAndSetCondition(
		mlopsv1alpha1.PipelineReady,
		false,
		pipelineStatus,
		err.Error(),
	)
	if errSet := r.Status().Update(ctx, pipeline); errSet != nil {
		logger.Error(
			errSet,
			"Failed to set status on pipeline on error",
			"pipeline", pipeline.Name,
			"error", err.Error(),
		)
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *PipelineReconciler) SetupWithManager(mgr ctrl.Manager) error {
	pred := predicate.GenerationChangedPredicate{}
	return ctrl.NewControllerManagedBy(mgr).
		For(&mlopsv1alpha1.Pipeline{}).
		WithEventFilter(pred).
		Complete(r)
}
