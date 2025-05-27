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
	"strconv"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	mlopsv1alpha1 "github.com/seldonio/seldon-core/operator/v2/apis/mlops/v1alpha1"
)

// SeldonConfigReconciler reconciles a SeldonConfig object
type SeldonConfigReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=mlops.seldon.io,resources=seldonconfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=mlops.seldon.io,resources=seldonconfigs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=mlops.seldon.io,resources=seldonconfigs/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the SeldonConfig object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *SeldonConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx).WithName("Reconcile")

	// Fetch the SeldonConfig instance
	seldonConfig := &mlopsv1alpha1.SeldonConfig{}
	if err := r.Client.Get(ctx, req.NamespacedName, seldonConfig); err != nil {
		if errors.IsNotFound(err) {
			logger.Error(err, "unable to fetch SeldonConfig", "name", req.Name, "namespace", req.Namespace)
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	// TODO(user): your logic here
	err := ValidateDataflowScaleSpec(ctx, r.Client, req.Namespace, seldonConfig, logr.FromContextOrDiscard(ctx))
	return ctrl.Result{}, err
}

// SetupWithManager sets up the controller with the Manager.
func (r *SeldonConfigReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mlopsv1alpha1.SeldonConfig{}).
		Complete(r)
}

func ParseInt32(s string) (int32, error) {
	i64, err := strconv.ParseInt(s, 10, 32)
	return int32(i64), err
}

func ValidateDataflowScaleSpec(
	ctx context.Context,
	clt client.Client,
	namespace string,
	seldonConfig *mlopsv1alpha1.SeldonConfig,
	logger logr.Logger,
) error {
	// Get the number of Pipeline resources in the namespace
	var pipelineList mlopsv1alpha1.PipelineList
	if err := clt.List(ctx, &pipelineList, client.InNamespace(namespace)); err != nil {
		return fmt.Errorf("failed to list Pipeline resources in namespace %s: %w", namespace, err)
	}

	pipelineCount := int32(len(pipelineList.Items))
	logger.Info("Number of Pipeline resources", "namespace", namespace, "count", pipelineCount)

	// Get the numbers of partitions per topic
	kafkaConfig := seldonConfig.Spec.Config.KafkaConfig
	logger.Info("kafkaConfig.Topics", "Topics", kafkaConfig.Topics)

	numPartitions, err := ParseInt32(kafkaConfig.Topics["numPartitions"].StrVal)
	if err != nil {
		return fmt.Errorf("failed to parse numPartitions from KafkaConfig: %w", err)
	}

	logger.Info("Using numPartitions from KafkaConfig", "numPartitions", numPartitions)

	maxReplicas := numPartitions
	if pipelineCount != 0 {
		maxReplicas = numPartitions * pipelineCount
	}

	logger.Info("Maximum replicas for dataflow engine", "max_replicas", maxReplicas)

	// Get the number of replicas for the dataflow engine
	for _, component := range seldonConfig.Spec.Components {
		logger.Info("Checking component", "component_name", component.Name)

		if component.Name == "seldon-dataflow-engine" {
			if component.Replicas == nil {
				return fmt.Errorf("seldon-dataflow-engine component must have replicas defined")
			}
			replicas := *component.Replicas
			logger.Info("Seldon dataflow engine replicas", "replicas", replicas)

			// Validate that the number of replicas is less than or equal to the number of partitions
			if replicas > maxReplicas {
				return fmt.Errorf("seldon-dataflow-engine replicas %d cannot be greater than %d", replicas, maxReplicas)
			}
			break
		}
	}

	return nil
}
