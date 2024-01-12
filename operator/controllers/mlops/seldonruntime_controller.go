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
	"fmt"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	auth "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	apimachinary_errors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"knative.dev/pkg/apis"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	mlopsv1alpha1 "github.com/seldonio/seldon-core/operator/v2/apis/mlops/v1alpha1"
	"github.com/seldonio/seldon-core/operator/v2/controllers/reconcilers/common"
	seldonreconcile "github.com/seldonio/seldon-core/operator/v2/controllers/reconcilers/seldon"
	"github.com/seldonio/seldon-core/operator/v2/pkg/constants"
	"github.com/seldonio/seldon-core/operator/v2/pkg/utils"
	scheduler "github.com/seldonio/seldon-core/operator/v2/scheduler"
)

// SeldonRuntimeReconciler reconciles a SeldonRuntime object
type SeldonRuntimeReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	Scheduler *scheduler.SchedulerClient
	Recorder  record.EventRecorder
}

func (r *SeldonRuntimeReconciler) getNumberOfServers(namespace string) (int, error) {
	servers := mlopsv1alpha1.ServerList{}
	inNamespace := client.InNamespace(namespace)
	err := r.List(context.TODO(), &servers, inNamespace)
	if err != nil {
		if errors.IsNotFound(err) {
			return 0, nil
		}
		return 0, err
	}
	return len(servers.Items), nil
}

func (r *SeldonRuntimeReconciler) handleFinalizer(ctx context.Context, logger logr.Logger, runtime *mlopsv1alpha1.SeldonRuntime) (bool, error) {
	// Check if we are being deleted or not
	if runtime.ObjectMeta.DeletionTimestamp.IsZero() { // Not being deleted
		// Add our finalizer
		if !utils.ContainsStr(runtime.ObjectMeta.Finalizers, constants.RuntimeFinalizerName) {
			runtime.ObjectMeta.Finalizers = append(runtime.ObjectMeta.Finalizers, constants.RuntimeFinalizerName)
			if err := r.Update(context.Background(), runtime); err != nil {
				return true, err
			}
		}
	} else { // runtime is being deleted
		numServers, err := r.getNumberOfServers(runtime.Namespace)
		logger.Info("Runtime being deleted", "namespace", runtime.Namespace, "numServers", numServers)
		if err != nil {
			return true, err
		}
		if numServers == 0 {
			logger.Info("Removing finalizer for runtime", "namespace", runtime.Namespace)
			runtime.ObjectMeta.Finalizers = utils.RemoveStr(runtime.ObjectMeta.Finalizers, constants.RuntimeFinalizerName)
			if err := r.Update(ctx, runtime); err != nil {
				return true, err
			}
			r.Scheduler.RemoveConnection(runtime.Namespace)
			// Stop reconciliation as the item is being deleted
			return true, nil
		} else {
			return true, fmt.Errorf("Runtime is being deleted but servers still running in namespace %s", runtime.Namespace)
		}
	}
	return false, nil
}

//+kubebuilder:rbac:groups=mlops.seldon.io,resources=seldonruntimes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=mlops.seldon.io,resources=seldonruntimes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=mlops.seldon.io,resources=seldonruntimes/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch
// +kubebuilder:rbac:groups=v1,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=v1,resources=services/status,verbs=get
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=statefulsets/status,verbs=get
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=roles,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=rolebindings,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=v1,resources=serviceaccounts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the SeldonRuntime object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *SeldonRuntimeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx).WithName("Reconcile")

	seldonRuntime := &mlopsv1alpha1.SeldonRuntime{}
	if err := r.Get(ctx, req.NamespacedName, seldonRuntime); err != nil {
		if errors.IsNotFound(err) {
			// we'll ignore not-found errors, since they can't be fixed by an immediate
			// requeue (we'll need to wait for a new notification), and we can get them
			// on deleted requests.
			return reconcile.Result{}, nil
		}
		logger.Error(err, "unable to fetch SeldonRuntime", "name", req.Name, "namespace", req.Namespace)
		return reconcile.Result{}, err
	}

	stop, err := r.handleFinalizer(ctx, logger, seldonRuntime)
	if stop {
		return reconcile.Result{}, err
	}

	sr, err := seldonreconcile.NewSeldonRuntimeReconciler(seldonRuntime, common.ReconcilerConfig{
		Ctx:    ctx,
		Logger: logger,
		Client: r.Client,
	})
	if err != nil {
		return reconcile.Result{}, err
	}

	// Set Controller References
	err = setControllerReferences(seldonRuntime, common.ToMetaObjects(sr.GetResources()), r.Scheme)
	if err != nil {
		return reconcile.Result{}, err
	}

	err = sr.Reconcile()
	if err != nil {
		return reconcile.Result{}, err
	}

	conditions := sr.GetConditions()
	for _, condition := range conditions {
		seldonRuntime.Status.SetCondition(condition)
	}

	err = r.updateStatus(seldonRuntime, logger)
	if err != nil {
		return reconcile.Result{}, err
	}

	if seldonRuntime.Status.IsConditionReady(mlopsv1alpha1.SchedulerReady) {
		if err := r.Scheduler.KeepConnection(seldonRuntime.Namespace); err != nil {
			return reconcile.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func seldonRuntimeReady(status mlopsv1alpha1.SeldonRuntimeStatus) bool {
	return status.Conditions != nil &&
		status.GetCondition(apis.ConditionReady) != nil &&
		status.GetCondition(apis.ConditionReady).Status == v1.ConditionTrue
}

func (r *SeldonRuntimeReconciler) updateStatus(seldonRuntime *mlopsv1alpha1.SeldonRuntime, logger logr.Logger) error {
	existingRuntime := &mlopsv1alpha1.SeldonRuntime{}
	namespacedName := types.NamespacedName{Name: seldonRuntime.Name, Namespace: seldonRuntime.Namespace}
	if err := r.Get(context.TODO(), namespacedName, existingRuntime); err != nil {
		if apimachinary_errors.IsNotFound(err) { //Ignore NotFound errors
			return nil
		}
		return err
	}

	if equality.Semantic.DeepEqual(existingRuntime.Status, seldonRuntime.Status) {
		// Not updating as no difference
	} else {
		if err := r.Status().Update(context.TODO(), seldonRuntime); err != nil {
			logger.Info("Failed to update status", "name", seldonRuntime.Name, "namespace", seldonRuntime.Namespace)
			r.Recorder.Eventf(seldonRuntime, v1.EventTypeWarning, "UpdateFailed",
				"Failed to update status for SeldonRuntime %q: %v", seldonRuntime.Name, err)
			return err
		} else {
			logger.Info("Successfully updated status", "name", seldonRuntime.Name, "namespace", seldonRuntime.Namespace)
			prevWasReady := seldonRuntimeReady(existingRuntime.Status)
			currentIsReady := seldonRuntimeReady(seldonRuntime.Status)
			if prevWasReady && !currentIsReady {
				r.Recorder.Eventf(seldonRuntime, v1.EventTypeWarning, "SeldonRuntimeNotReady",
					fmt.Sprintf("SeldonRuntime %v is no longer Ready", seldonRuntime.GetName()))
			} else if !prevWasReady && currentIsReady {
				r.Recorder.Eventf(seldonRuntime, v1.EventTypeNormal, "RuntimeReady",
					fmt.Sprintf("SeldonRuntime %v is Ready", seldonRuntime.GetName()))
			}
		}
	}
	return nil
}

// Find SeldonRuntimes that reference the changes SeldonConfig
func (r *SeldonRuntimeReconciler) mapSeldonRuntimesFromSeldonConfig(obj client.Object) []reconcile.Request {
	logger := log.FromContext(context.Background()).WithName("mapSeldonRuntimesFromSeldonConfig")
	var seldonRuntimes mlopsv1alpha1.SeldonRuntimeList
	if err := r.Client.List(context.Background(), &seldonRuntimes); err != nil {
		logger.Error(err, "error listing seldonRuntimes")
		return nil
	}

	seldonConfig := obj.(*mlopsv1alpha1.SeldonConfig)
	var req []reconcile.Request
	for _, seldonRuntime := range seldonRuntimes.Items {
		if !seldonRuntime.Spec.DisableAutoUpdate && seldonRuntime.Spec.SeldonConfig == seldonConfig.Name {
			req = append(req, reconcile.Request{
				NamespacedName: types.NamespacedName{
					Namespace: seldonRuntime.Namespace,
					Name:      seldonRuntime.Name,
				},
			})
		}
	}
	return req
}

// SetupWithManager sets up the controller with the Manager.
func (r *SeldonRuntimeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	//pred := predicate.GenerationChangedPredicate{}
	return ctrl.NewControllerManagedBy(mgr).
		For(&mlopsv1alpha1.SeldonRuntime{}).
		//WithEventFilter(pred).
		Owns(&appsv1.StatefulSet{}).
		Owns(&appsv1.Deployment{}).
		Owns(&v1.Service{}).
		Owns(&auth.Role{}).
		Owns(&auth.RoleBinding{}).
		Owns(&v1.ServiceAccount{}).
		Owns(&v1.ConfigMap{}).
		Watches(
			&source.Kind{Type: &mlopsv1alpha1.SeldonConfig{}},
			handler.EnqueueRequestsFromMapFunc(r.mapSeldonRuntimesFromSeldonConfig),
		).
		Complete(r)
}
