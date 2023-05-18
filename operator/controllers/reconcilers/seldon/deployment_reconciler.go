/*
Copyright 2022 Seldon Technologies Ltd.

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

package server

import (
	"context"
	"fmt"

	"github.com/banzaicloud/k8s-objectmatcher/patch"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/seldonio/seldon-core/operator/v2/pkg/utils"

	mlopsv1alpha1 "github.com/seldonio/seldon-core/operator/v2/apis/mlops/v1alpha1"
	"github.com/seldonio/seldon-core/operator/v2/controllers/reconcilers/common"
	"github.com/seldonio/seldon-core/operator/v2/pkg/constants"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"knative.dev/pkg/apis"
)

type ComponentDeploymentReconciler struct {
	common.ReconcilerConfig
	Name       string
	Deployment *appsv1.Deployment
	Annotator  *patch.Annotator
}

func NewComponentDeploymentReconciler(
	name string,
	common common.ReconcilerConfig,
	meta metav1.ObjectMeta,
	podSpec *v1.PodSpec,
	override *mlopsv1alpha1.OverrideSpec,
	seldonConfigMeta metav1.ObjectMeta,
	annotator *patch.Annotator,
) *ComponentDeploymentReconciler {
	labels := utils.MergeMaps(meta.Labels, seldonConfigMeta.Labels)
	annotations := utils.MergeMaps(meta.Annotations, seldonConfigMeta.Annotations)

	return &ComponentDeploymentReconciler{
		ReconcilerConfig: common,
		Name:             name,
		Deployment:       toDeployment(name, meta, podSpec, override, labels, annotations),
		Annotator:        annotator,
	}
}

func (s *ComponentDeploymentReconciler) GetResources() []client.Object {
	return []client.Object{s.Deployment}
}

func addEnvoyAnnotations(annotations map[string]string) map[string]string {
	annotations["prometheus.io/path"] = "/stats/prometheus"
	annotations["prometheus.io/scrape"] = "true"
	annotations["prometheus.io/port"] = "9003"
	return annotations
}

func toDeployment(
	name string,
	meta metav1.ObjectMeta,
	podSpec *v1.PodSpec,
	override *mlopsv1alpha1.OverrideSpec,
	labels map[string]string,
	annotations map[string]string) *appsv1.Deployment {
	var replicas int32
	if override != nil && override.Replicas != nil {
		replicas = *override.Replicas
	} else {
		replicas = 1
	}
	// Envoy annotations
	if name == mlopsv1alpha1.EnvoyName {
		annotations = addEnvoyAnnotations(annotations)
	}
	metaLabels := utils.MergeMaps(map[string]string{constants.KubernetesNameLabelKey: name}, labels)
	templateLabels := utils.MergeMaps(map[string]string{constants.KubernetesNameLabelKey: name}, labels)
	d := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   meta.Namespace,
			Labels:      metaLabels,
			Annotations: annotations,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{constants.KubernetesNameLabelKey: name},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      templateLabels,
					Annotations: common.CopyMap(annotations),
					Name:        name,
					Namespace:   meta.Namespace,
				},
				Spec: *podSpec,
			},
		},
	}

	return d
}

func (s *ComponentDeploymentReconciler) GetLabelSelector() string {
	return fmt.Sprintf("%s=%s", constants.KubernetesNameLabelKey, s.Name)
}

func (s *ComponentDeploymentReconciler) getReconcileOperation() (constants.ReconcileOperation, error) {
	found := &appsv1.Deployment{}
	err := s.Client.Get(context.TODO(), types.NamespacedName{
		Name:      s.Deployment.GetName(),
		Namespace: s.Deployment.GetNamespace(),
	}, found)
	if err != nil {
		if errors.IsNotFound(err) {
			return constants.ReconcileCreateNeeded, nil
		}
		return constants.ReconcileUnknown, err
	}
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
		// Update our version so we have Status which can be used
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

func (s *ComponentDeploymentReconciler) Reconcile() error {
	logger := s.Logger.WithName("DeploymentReconcile")
	op, err := s.getReconcileOperation()
	switch op {
	case constants.ReconcileCreateNeeded:
		logger.V(1).Info("Deployment Create", "Name", s.Deployment.GetName(), "Namespace", s.Deployment.GetNamespace())
		err = s.Client.Create(s.Ctx, s.Deployment)
		if err != nil {
			logger.Error(err, "Failed to create Deployment", "Name", s.Deployment.GetName(), "Namespace", s.Deployment.GetNamespace())
			return err
		}
	case constants.ReconcileUpdateNeeded:
		logger.V(1).Info("Deployment Update", "Name", s.Deployment.GetName(), "Namespace", s.Deployment.GetNamespace())
		err = s.Client.Update(s.Ctx, s.Deployment)
		if err != nil {
			logger.Error(err, "Failed to update statefuleset", "Name", s.Deployment.GetName(), "Namespace", s.Deployment.GetNamespace())
			return err
		}
	case constants.ReconcileNoChange:
		err = nil
		logger.V(1).Info("Deployment No Change", "Name", s.Deployment.GetName(), "Namespace", s.Deployment.GetNamespace())
	case constants.ReconcileUnknown:
		if err != nil {
			logger.Error(err, "Failed to get reconcile operation for Deployment", "Name", s.Deployment.GetName(), "Namespace", s.Deployment.GetNamespace())
			return err
		}
		return err
	}
	return nil
}

const (
	DeloymentReadyReason     = "Deployment replicas matches desired replicas"
	DeploymentNotReadyReason = "Deployment replicas does not match desired replicas"
)

func (s *ComponentDeploymentReconciler) GetConditions() []*apis.Condition {
	ready := s.Deployment.Status.ReadyReplicas >= s.Deployment.Status.Replicas
	s.Logger.Info("Checking conditions for Deployment",
		"name", s.Name,
		"ready", ready,
		"namespace", s.Deployment.Namespace,
		"readyReplicas", s.Deployment.Status.ReadyReplicas,
		"replicas", s.Deployment.Status.Replicas)
	if conditionType, ok := mlopsv1alpha1.ConditionNameMap[s.Name]; ok {
		if ready {
			return []*apis.Condition{mlopsv1alpha1.CreateCondition(conditionType, ready, DeloymentReadyReason)}
		} else {
			return []*apis.Condition{mlopsv1alpha1.CreateCondition(conditionType, ready, DeploymentNotReadyReason)}
		}
	} else {
		return []*apis.Condition{}
	}
}
