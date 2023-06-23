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

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
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
	serverreconcile "github.com/seldonio/seldon-core/operator/v2/controllers/reconcilers/server"
	"github.com/seldonio/seldon-core/operator/v2/pkg/constants"
	"github.com/seldonio/seldon-core/operator/v2/pkg/utils"
	scheduler "github.com/seldonio/seldon-core/operator/v2/scheduler"
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
	} else { // server is being deleted
		if utils.ContainsStr(server.ObjectMeta.Finalizers, constants.ServerFinalizerName) {
			// Handle unload in scheduler
			if err := r.Scheduler.ServerNotify(ctx, server); err != nil {
				// Remove finalizer as we can't reach scheduler
				server.ObjectMeta.Finalizers = utils.RemoveStr(server.ObjectMeta.Finalizers, constants.ServerFinalizerName)
				if err2 := r.Update(ctx, server); err != nil {
					return true, err2
				}
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

	logger.Info("Received reconcile for Server", "name", req.Name, "namespace", req.NamespacedName.Namespace)

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

	logger.Info("Found server", "name", server.Name, "namespace", server.Namespace)

	stop, err := r.handleFinalizer(ctx, server)
	if stop {
		if err != nil {
			r.updateStatusFromError(ctx, logger, server, err)
		}
		return reconcile.Result{}, err
	}

	err = r.Scheduler.ServerNotify(ctx, server)
	if err != nil {
		r.updateStatusFromError(ctx, logger, server, err)
		return reconcile.Result{}, err
	}

	sr, err := serverreconcile.NewServerReconciler(server, common.ReconcilerConfig{
		Ctx:    ctx,
		Logger: logger,
		Client: r.Client,
	})
	if err != nil {
		r.updateStatusFromError(ctx, logger, server, err)
		return reconcile.Result{}, err
	}

	// Set Controller References
	err = setControllerReferences(server, common.ToMetaObjects(sr.GetResources()), r.Scheme)
	if err != nil {
		r.updateStatusFromError(ctx, logger, server, err)
		return reconcile.Result{}, err
	}

	err = sr.Reconcile()
	if err != nil {
		r.updateStatusFromError(ctx, logger, server, err)
		return reconcile.Result{}, err
	}

	conditions := sr.GetConditions()
	for _, condition := range conditions {
		server.Status.SetCondition(condition)
	}

	// Update status fields
	selector := sr.(common.LabelHandler).GetLabelSelector()
	replicas := *server.Spec.Replicas
	server.Status.Selector = selector
	server.Status.Replicas = replicas

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

func (r *ServerReconciler) updateStatusFromError(ctx context.Context, logger logr.Logger, server *mlopsv1alpha1.Server, err error) {
	server.Status.CreateAndSetCondition(mlopsv1alpha1.StatefulSetReady, false, err.Error())
	if errSet := r.Status().Update(ctx, server); errSet != nil {
		logger.Error(errSet, "Failed to set status for server on error", "server", server.Name, "error", err.Error())
	}
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

// Find Servers that need reconcilliation from a change to a given ServerConfig
func (r *ServerReconciler) mapServerFromServerConfig(obj client.Object) []reconcile.Request {
	logger := log.FromContext(context.Background()).WithName("mapServerFromServerConfig")
	var servers mlopsv1alpha1.ServerList
	if err := r.Client.List(context.Background(), &servers); err != nil {
		logger.Error(err, "error listing servers")
		return nil
	}

	serverConfig := obj.(*mlopsv1alpha1.ServerConfig)
	var req []reconcile.Request
	for _, server := range servers.Items {
		if server.Spec.ServerConfig == serverConfig.Name {
			req = append(req, reconcile.Request{
				NamespacedName: types.NamespacedName{
					Namespace: server.Namespace,
					Name:      server.Name,
				},
			})
		}
	}
	return req
}

// SetupWithManager sets up the controller with the Manager.
func (r *ServerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mlopsv1alpha1.Server{}).
		Owns(&appsv1.StatefulSet{}).
		Owns(&v1.Service{}).
		Watches(
			&source.Kind{Type: &mlopsv1alpha1.ServerConfig{}},
			handler.EnqueueRequestsFromMapFunc(r.mapServerFromServerConfig),
		).
		Complete(r)
}
