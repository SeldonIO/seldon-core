/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

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

type ServerStatefulSetReconciler struct {
	common.ReconcilerConfig
	StatefulSet *appsv1.StatefulSet
	Annotator   *patch.Annotator
}

func NewServerStatefulSetReconciler(
	common common.ReconcilerConfig,
	meta metav1.ObjectMeta,
	podSpec *v1.PodSpec,
	volumeClaimTemplates []mlopsv1alpha1.PersistentVolumeClaim,
	scaling *mlopsv1alpha1.ScalingSpec,
	volumeClaimRetentionPolicy *appsv1.StatefulSetPersistentVolumeClaimRetentionPolicy,
	serverConfigMeta metav1.ObjectMeta,
	annotator *patch.Annotator,
) *ServerStatefulSetReconciler {
	labels := utils.MergeMaps(meta.Labels, serverConfigMeta.Labels)
	annotations := utils.MergeMaps(meta.Annotations, serverConfigMeta.Annotations)
	return &ServerStatefulSetReconciler{
		ReconcilerConfig: common,
		StatefulSet:      toStatefulSet(meta, podSpec, volumeClaimTemplates, scaling, volumeClaimRetentionPolicy, labels, annotations),
		Annotator:        annotator,
	}
}

func (s *ServerStatefulSetReconciler) GetResources() []client.Object {
	return []client.Object{s.StatefulSet}
}

func toStatefulSet(meta metav1.ObjectMeta,
	podSpec *v1.PodSpec,
	volumeClaimTemplates []mlopsv1alpha1.PersistentVolumeClaim,
	scaling *mlopsv1alpha1.ScalingSpec,
	volumeClaimRetentionPolicy *appsv1.StatefulSetPersistentVolumeClaimRetentionPolicy,
	labels map[string]string,
	annotations map[string]string) *appsv1.StatefulSet {
	labels[constants.KubernetesNameLabelKey] = constants.ServerLabelValue
	metaLabels := utils.MergeMaps(map[string]string{constants.KubernetesNameLabelKey: constants.ServerLabelValue}, labels)
	templateLabels := utils.MergeMaps(map[string]string{constants.ServerLabelNameKey: meta.Name, constants.KubernetesNameLabelKey: constants.ServerLabelValue}, labels)
	ss := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:        meta.Name,
			Namespace:   meta.Namespace,
			Labels:      metaLabels,
			Annotations: annotations,
		},
		Spec: appsv1.StatefulSetSpec{
			ServiceName: meta.Name,
			Replicas:    scaling.Replicas,
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
			PodManagementPolicy:                  appsv1.ParallelPodManagement,
			PersistentVolumeClaimRetentionPolicy: volumeClaimRetentionPolicy,
		},
	}

	// Add volume claim templates from internal resource
	for _, vct := range volumeClaimTemplates {
		ss.Spec.VolumeClaimTemplates = append(ss.Spec.VolumeClaimTemplates, v1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name: vct.Name,
			},
			Spec: vct.Spec,
		})
	}
	return ss
}

func (s *ServerStatefulSetReconciler) GetLabelSelector() string {
	return fmt.Sprintf("%s=%s", constants.ServerLabelNameKey, s.StatefulSet.GetName())
}

func (s *ServerStatefulSetReconciler) GetReplicas() (int32, error) {
	found := &appsv1.StatefulSet{}
	err := s.Client.Get(context.TODO(), types.NamespacedName{
		Name:      s.StatefulSet.GetName(),
		Namespace: s.StatefulSet.GetNamespace(),
	}, found)
	if err != nil {
		if errors.IsNotFound(err) {
			return 0, nil
		}
		return 0, err
	}
	return found.Status.Replicas, nil
}

func (s *ServerStatefulSetReconciler) getReconcileOperation() (constants.ReconcileOperation, error) {
	found := &appsv1.StatefulSet{}
	err := s.Client.Get(context.TODO(), types.NamespacedName{
		Name:      s.StatefulSet.GetName(),
		Namespace: s.StatefulSet.GetNamespace(),
	}, found)
	if err != nil {
		if errors.IsNotFound(err) {
			return constants.ReconcileCreateNeeded, nil
		}
		return constants.ReconcileUnknown, err
	}
	stsJson, err := json.Marshal(s.StatefulSet)
	if err != nil {
		return constants.ReconcileUnknown, err
	}
	s.Logger.Info("Found StatefulSet", "StatefulSet", string(stsJson))
	opts := []patch.CalculateOption{
		patch.IgnoreStatusFields(),
		patch.IgnoreField("kind"),
		patch.IgnoreField("apiVersion"),
		patch.IgnoreField("metadata"),
		patch.IgnoreVolumeClaimTemplateTypeMetaAndStatus(),
		common.IgnoreVolumeClaimTemplateVolumeModel(),
	}
	patcherMaker := patch.NewPatchMaker(s.Annotator, &patch.K8sStrategicMergePatcher{}, &patch.BaseJSONMergePatcher{})
	patchResult, err := patcherMaker.Calculate(found, s.StatefulSet, opts...)
	if err != nil {
		return constants.ReconcileUnknown, err
	}
	s.StatefulSet.Status = found.Status
	if patchResult.IsEmpty() {
		// Update our version so we have Status which can be used
		s.StatefulSet = found
		return constants.ReconcileNoChange, nil
	}
	err = s.Annotator.SetLastAppliedAnnotation(s.StatefulSet)
	if err != nil {
		return constants.ReconcileUnknown, err
	}
	// Update resource version so we can do a client Update successfully
	// This needs to be done after we annotate to also avoid false differences
	s.StatefulSet.SetResourceVersion(found.ResourceVersion)
	return constants.ReconcileUpdateNeeded, nil
}

func (s *ServerStatefulSetReconciler) Reconcile() error {
	logger := s.Logger.WithName("StatefulSetReconcile")
	op, err := s.getReconcileOperation()
	switch op {
	case constants.ReconcileCreateNeeded:
		logger.V(1).Info("StatefulSet Create", "Name", s.StatefulSet.GetName(), "Namespace", s.StatefulSet.GetNamespace())
		err = s.Client.Create(s.Ctx, s.StatefulSet)
		if err != nil {
			logger.Error(err, "Failed to create statefulset", "Name", s.StatefulSet.GetName(), "Namespace", s.StatefulSet.GetNamespace())
			return err
		}
	case constants.ReconcileUpdateNeeded:
		logger.V(1).Info("StatefulSet Update", "Name", s.StatefulSet.GetName(), "Namespace", s.StatefulSet.GetNamespace())
		err = s.Client.Update(s.Ctx, s.StatefulSet)
		if err != nil {
			logger.Error(err, "Failed to update statefulset", "Name", s.StatefulSet.GetName(), "Namespace", s.StatefulSet.GetNamespace())
			return err
		}
	case constants.ReconcileNoChange:
		err = nil
		logger.V(1).Info("StatefulSet No Change", "Name", s.StatefulSet.GetName(), "Namespace", s.StatefulSet.GetNamespace())
	case constants.ReconcileUnknown:
		if err != nil {
			logger.Error(err, "Failed to get reconcile operation for statefulset", "Name", s.StatefulSet.GetName(), "Namespace", s.StatefulSet.GetNamespace())
			return err
		}
		return err
	}
	return nil
}

const (
	StatefulSetReadyReason    = "StatefulSet replicas matches desired replicas"
	StatefulSetNotReadyReason = "StatefulSet replicas does not match desired replicas"
	StatefulSetReplicasNil    = "[BUG] StatefulSet replicas is nil"
)

func (s *ServerStatefulSetReconciler) GetConditions() []*apis.Condition {
	// Replicas should never be nil as it is set to a default when not given explicitly
	// Check to defend against programmatic setting to nil (i.e a bug in the code)
	if s.StatefulSet.Spec.Replicas == nil {
		s.Logger.Info(StatefulSetReplicasNil)
		return []*apis.Condition{mlopsv1alpha1.CreateCondition(mlopsv1alpha1.StatefulSetReady, false, StatefulSetReplicasNil)}
	}

	ready := s.StatefulSet.Status.ReadyReplicas >= *s.StatefulSet.Spec.Replicas
	s.Logger.Info("Checking conditions for stateful set", "ready", ready, ".spec.replicas", *s.StatefulSet.Spec.Replicas, ".status.replicas", s.StatefulSet.Status.Replicas, "availableReplicas", s.StatefulSet.Status.AvailableReplicas)
	if ready {
		return []*apis.Condition{mlopsv1alpha1.CreateCondition(mlopsv1alpha1.StatefulSetReady, ready, StatefulSetReadyReason)}
	} else {
		return []*apis.Condition{mlopsv1alpha1.CreateCondition(mlopsv1alpha1.StatefulSetReady, ready, StatefulSetNotReadyReason)}
	}
}
