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
	"fmt"

	"github.com/seldonio/seldon-core/operatorv2/controllers/reconcilers"
	"github.com/seldonio/seldon-core/operatorv2/controllers/reconcilers/common"
	"github.com/seldonio/seldon-core/operatorv2/pkg/constants"
	"github.com/seldonio/seldon-core/operatorv2/pkg/utils"
	scheduler "github.com/seldonio/seldon-core/operatorv2/scheduler"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"knative.dev/pkg/apis"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	mlopsv1alpha1 "github.com/seldonio/seldon-core/operatorv2/apis/mlops/v1alpha1"
	apimachinary_errors "k8s.io/apimachinery/pkg/api/errors"
)

// ServerReconciler reconciles a Server object
type ServerReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	Scheduler *scheduler.SchedulerClient
	Recorder  record.EventRecorder
}

func (r *ServerReconciler) handleFinalizer(ctx context.Context, server *mlopsv1alpha1.Server) (bool, error) {

	// Check if we are being deleted or not
	if server.ObjectMeta.DeletionTimestamp.IsZero() { // Not being deleted

		// Add our finalizer
		if !utils.ContainsStr(server.ObjectMeta.Finalizers, constants.ServerFinalizerName) {
			server.ObjectMeta.Finalizers = append(server.ObjectMeta.Finalizers, constants.ServerFinalizerName)
			if err := r.Update(context.Background(), server); err != nil {
				return true, err
			}
		}
	} else { // model is being deleted
		if utils.ContainsStr(server.ObjectMeta.Finalizers, constants.ServerFinalizerName) {
			// Handle unload in scheduler
			if err := r.Scheduler.ServerNotify(ctx, server); err != nil {
				return true, err
			}
			if server.Status.LoadedModelReplicas == 0 { // Remove finalizer if no models loaded otherwise we wait
				server.ObjectMeta.Finalizers = utils.RemoveStr(server.ObjectMeta.Finalizers, constants.ServerFinalizerName)
				if err := r.Update(ctx, server); err != nil {
					return true, err
				}
			}
		}
		// Stop reconciliation as the item is being deleted
		return true, nil
	}
	return false, nil
}

//+kubebuilder:rbac:groups=mlops.seldon.io,resources=servers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=mlops.seldon.io,resources=servers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=mlops.seldon.io,resources=servers/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch
// +kubebuilder:rbac:groups=v1,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=v1,resources=services/status,verbs=get
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=statefulsets/status,verbs=get

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Server object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *ServerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx).WithName("Reconcile")

	server := &mlopsv1alpha1.Server{}
	if err := r.Get(ctx, req.NamespacedName, server); err != nil {
		if errors.IsNotFound(err) {
			// we'll ignore not-found errors, since they can't be fixed by an immediate
			// requeue (we'll need to wait for a new notification), and we can get them
			// on deleted requests.
			return reconcile.Result{}, nil
		}
		logger.Error(err, "unable to fetch Server", "name", req.Name, "namespace", req.Namespace)
		return reconcile.Result{}, err
	}

	stop, err := r.handleFinalizer(ctx, server)
	if stop {
		return reconcile.Result{}, err
	}

	err = r.Scheduler.ServerNotify(ctx, server)
	if err != nil {
		return reconcile.Result{}, err
	}

	sr, err := reconcilers.NewServerReconciler(server, common.ReconcilerConfig{
		Ctx:    ctx,
		Logger: logger,
		Client: r.Client,
	})
	if err != nil {
		return reconcile.Result{}, err
	}

	// Set Controller References
	err = setControllerReferences(server, sr.GetResources(), r.Scheme)
	if err != nil {
		return reconcile.Result{}, err
	}

	err = sr.Reconcile()
	if err != nil {
		return reconcile.Result{}, err
	}

	conditions := sr.GetConditions()
	for _, condition := range conditions {
		server.Status.SetCondition(condition)
	}

	err = r.updateStatus(server)
	if err != nil {
		return reconcile.Result{}, err
	}

	return ctrl.Result{}, nil
}

func serverReady(status mlopsv1alpha1.ServerStatus) bool {
	return status.Conditions != nil &&
		status.GetCondition(apis.ConditionReady) != nil &&
		status.GetCondition(apis.ConditionReady).Status == v1.ConditionTrue
}

func (r *ServerReconciler) updateStatus(server *mlopsv1alpha1.Server) error {
	existingServer := &mlopsv1alpha1.Server{}
	namespacedName := types.NamespacedName{Name: server.Name, Namespace: server.Namespace}
	if err := r.Get(context.TODO(), namespacedName, existingServer); err != nil {
		if apimachinary_errors.IsNotFound(err) { //Ignore NotFound errors
			return nil
		}
		return err
	}

	if equality.Semantic.DeepEqual(existingServer.Status, server.Status) {
		// Not updating as no difference
	} else {
		if err := r.Status().Update(context.TODO(), server); err != nil {
			r.Recorder.Eventf(server, v1.EventTypeWarning, "UpdateFailed",
				"Failed to update status for Model %q: %v", server.Name, err)
			return err
		} else {
			prevWasReady := serverReady(existingServer.Status)
			currentIsReady := serverReady(server.Status)
			if prevWasReady && !currentIsReady {
				r.Recorder.Eventf(server, v1.EventTypeWarning, "ServerNotReady",
					fmt.Sprintf("Server %v is no longer Ready", server.GetName()))
			} else if !prevWasReady && currentIsReady {
				r.Recorder.Eventf(server, v1.EventTypeNormal, "ServerReady",
					fmt.Sprintf("Server %v is Ready", server.GetName()))
			}
		}
	}
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ServerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mlopsv1alpha1.Server{}).
		Owns(&appsv1.StatefulSet{}).
		Owns(&v1.Service{}).
		Complete(r)
}
