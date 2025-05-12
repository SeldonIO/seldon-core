package server

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/banzaicloud/k8s-objectmatcher/patch"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"knative.dev/pkg/apis"
	"sigs.k8s.io/controller-runtime/pkg/client"

	mlopsv1alpha1 "github.com/seldonio/seldon-core/operator/v2/apis/mlops/v1alpha1"
	"github.com/seldonio/seldon-core/operator/v2/controllers/reconcilers/common"
	"github.com/seldonio/seldon-core/operator/v2/pkg/constants"
	"github.com/seldonio/seldon-core/operator/v2/pkg/utils"
)

type ServerDeploymentReconciler struct {
	common.ReconcilerConfig
	Deployment *appsv1.Deployment
	Annotator  *patch.Annotator
}

func NewServerDeploymentReconciler(
	common common.ReconcilerConfig,
	meta metav1.ObjectMeta,
	podSpec *v1.PodSpec,
	scaling *mlopsv1alpha1.ScalingSpec,
	serverConfigMeta metav1.ObjectMeta,
	annotator *patch.Annotator,
) *ServerDeploymentReconciler {
	labels := utils.MergeMaps(meta.Labels, serverConfigMeta.Labels)
	annotations := utils.MergeMaps(meta.Annotations, serverConfigMeta.Annotations)
	return &ServerDeploymentReconciler{
		ReconcilerConfig: common,
		Deployment:       toDeploymentTest(meta, podSpec, scaling, labels, annotations),
		Annotator:        annotator,
	}
}

func (s *ServerDeploymentReconciler) GetResources() []client.Object {
	return []client.Object{s.Deployment}
}

func (s *ServerDeploymentReconciler) GetLabelSelector() string {
	return fmt.Sprintf("%s=%s", constants.ServerLabelNameKey, s.Deployment.GetName())
}

func toDeploymentTest(
	meta metav1.ObjectMeta,
	podSpec *v1.PodSpec,
	scaling *mlopsv1alpha1.ScalingSpec,
	labels map[string]string,
	annotations map[string]string,
) *appsv1.Deployment {
	metaLabels := utils.MergeMaps(map[string]string{constants.KubernetesNameLabelKey: constants.ServerLabelValue}, labels)
	templateLabels := utils.MergeMaps(map[string]string{constants.ServerLabelNameKey: meta.Name, constants.KubernetesNameLabelKey: constants.ServerLabelValue}, labels)

	// Start with any PVC volumes
	var volumes []v1.Volume

	// Add required volumes explicitly
	volumes = append(volumes,
		v1.Volume{
			Name: "config-volume",
			VolumeSource: v1.VolumeSource{
				ConfigMap: &v1.ConfigMapVolumeSource{
					LocalObjectReference: v1.LocalObjectReference{
						Name: "seldon-agent", // adjust to your actual ConfigMap name
					},
				},
			},
		},
		v1.Volume{
			Name: "tracing-config-volume",
			VolumeSource: v1.VolumeSource{
				ConfigMap: &v1.ConfigMapVolumeSource{
					LocalObjectReference: v1.LocalObjectReference{
						Name: "seldon-tracing", // adjust to your actual ConfigMap name
					},
				},
			},
		},
	)

	podSpec.Volumes = volumes
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:        meta.Name,
			Namespace:   meta.Namespace,
			Labels:      metaLabels,
			Annotations: annotations,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: scaling.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{constants.ServerLabelNameKey: meta.Name},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      templateLabels,
					Annotations: common.CopyMap(annotations),
					Name:        meta.Name,
					Namespace:   meta.Namespace,
				},
				Spec: *podSpec,
			},
		},
	}
}

func (s *ServerDeploymentReconciler) getReconcileOperation() (constants.ReconcileOperation, error) {
	found := &appsv1.Deployment{}
	err := s.Client.Get(
		context.TODO(),
		types.NamespacedName{
			Name:      s.Deployment.GetName(),
			Namespace: s.Deployment.GetNamespace(),
		},
		found,
	)

	if err != nil {
		if errors.IsNotFound(err) {
			return constants.ReconcileCreateNeeded, nil
		}
		return constants.ReconcileUnknown, err
	}

	depJson, err := json.Marshal(s.Deployment)
	if err != nil {
		return constants.ReconcileUnknown, err
	}
	s.Logger.Info("Found Deployment", "Deployment", string(depJson))
	opts := []patch.CalculateOption{
		patch.IgnoreStatusFields(),
		patch.IgnoreField("kind"),
		patch.IgnoreField("apiVersion"),
		patch.IgnoreField("metadata"),
	}
	patcherMaker := patch.NewPatchMaker(s.Annotator, &patch.K8sStrategicMergePatcher{}, &patch.BaseJSONMergePatcher{})
	patchResult, err := patcherMaker.Calculate(found, s.Deployment, opts...)
	if err != nil {
		return constants.ReconcileUnknown, err
	}
	s.Deployment.Status = found.Status
	if patchResult.IsEmpty() {
		s.Deployment = found
		return constants.ReconcileNoChange, nil
	}
	err = s.Annotator.SetLastAppliedAnnotation(s.Deployment)
	if err != nil {
		return constants.ReconcileUnknown, err
	}
	// Update resource version so we can do a client Update successfully
	// This needs to be done after we annotate to also avoid false differences
	s.Deployment.SetResourceVersion(found.ResourceVersion)
	return constants.ReconcileUpdateNeeded, nil
}

func (s *ServerDeploymentReconciler) Reconcile() error {
	logger := s.Logger.WithName("DeploymentReconcile")
	op, err := s.getReconcileOperation()

	switch op {
	case constants.ReconcileCreateNeeded:
		logger.V(1).Info("Deployment Create", "Name", s.Deployment.GetName(), "Namespace", s.Deployment.GetNamespace())
		err = s.Client.Create(s.Ctx, s.Deployment)
		if err != nil {
			logger.Error(err, "Failed to create deployment", "Name", s.Deployment.GetName(), "Namespace", s.Deployment.GetNamespace())
			return err
		}
	case constants.ReconcileUpdateNeeded:
		logger.V(1).Info("Deployment Update", "Name", s.Deployment.GetName(), "Namespace", s.Deployment.GetNamespace())
		err = s.Client.Update(s.Ctx, s.Deployment)
		if err != nil {
			logger.Error(err, "Failed to update deployment", "Name", s.Deployment.GetName(), "Namespace", s.Deployment.GetNamespace())
			return err
		}
	case constants.ReconcileNoChange:
		err = nil
		logger.V(1).Info("Deployment No Change", "Name", s.Deployment.GetName(), "Namespace", s.Deployment.GetNamespace())
	case constants.ReconcileUnknown:
		if err != nil {
			logger.Error(err, "Failed to get reconcile operation for deployment", "Name", s.Deployment.GetName(), "Namespace", s.Deployment.GetNamespace())
			return err
		}
		return err
	}
	return nil
}

const (
	DeploymentReadyReason    = "Deployment replicas matches desired replicas"
	DeploymentNotReadyReason = "Deployment replicas does not match desired replicas"
	DeploymentReplicasNil    = "[BUG] Deployment replicas is nil"
)

func (s *ServerDeploymentReconciler) GetConditions() []*apis.Condition {
	// Replicas should never be nil as it is set to a default when not given explicitly
	// Check to defend against programmatic setting to nil (i.e a bug in the code)
	if s.Deployment.Spec.Replicas == nil {
		s.Logger.Info(DeploymentReplicasNil)
		return []*apis.Condition{mlopsv1alpha1.CreateCondition(mlopsv1alpha1.DeploymentReady, false, DeploymentReplicasNil)}
	}

	ready := s.Deployment.Status.ReadyReplicas >= *s.Deployment.Spec.Replicas
	s.Logger.Info("Checking conditions for deployment", "ready", ready, ".spec.replicas", *s.Deployment.Spec.Replicas, ".status.replicas", s.Deployment.Status.Replicas, "availableReplicas", s.Deployment.Status.AvailableReplicas)
	if ready {
		return []*apis.Condition{mlopsv1alpha1.CreateCondition(mlopsv1alpha1.DeploymentReady, ready, DeploymentReadyReason)}
	} else {
		return []*apis.Condition{mlopsv1alpha1.CreateCondition(mlopsv1alpha1.DeploymentReady, ready, DeploymentNotReadyReason)}
	}
}
