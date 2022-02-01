package statefulset

import (
	"context"

	mlopsv1alpha1 "github.com/seldonio/seldon-core/operatorv2/apis/mlops/v1alpha1"
	"github.com/seldonio/seldon-core/operatorv2/controllers/reconcilers/common"
	"github.com/seldonio/seldon-core/operatorv2/pkg/constants"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"knative.dev/pkg/apis"
)

type StatefulSetReconciler struct {
	common.ReconcilerConfig
	StatefulSet *appsv1.StatefulSet
}

func NewStatefulSetReconciler(
	common common.ReconcilerConfig,
	meta metav1.ObjectMeta,
	podSpec *v1.PodSpec,
	volumeClaimTemplates []mlopsv1alpha1.PersistentVolumeClaim,
	scaling *mlopsv1alpha1.ScalingSpec,
) *StatefulSetReconciler {
	return &StatefulSetReconciler{
		ReconcilerConfig: common,
		StatefulSet:      toStatefulSet(meta, podSpec, volumeClaimTemplates, scaling),
	}
}

func (s *StatefulSetReconciler) GetResources() []metav1.Object {
	return []metav1.Object{s.StatefulSet}
}

func toStatefulSet(meta metav1.ObjectMeta, podSpec *v1.PodSpec, volumeClaimTemplates []mlopsv1alpha1.PersistentVolumeClaim, scaling *mlopsv1alpha1.ScalingSpec) *appsv1.StatefulSet {
	ss := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      meta.Name,
			Namespace: meta.Namespace,
			Labels:    map[string]string{constants.AppKey: constants.ServerLabelValue},
		},
		Spec: appsv1.StatefulSetSpec{
			ServiceName: meta.Name,
			Replicas:    scaling.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{constants.ServerLabelNameKey: meta.Name},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:    map[string]string{constants.ServerLabelNameKey: meta.Name, constants.AppKey: constants.ServerLabelValue},
					Name:      meta.Name,
					Namespace: meta.Namespace,
				},
				Spec: *podSpec,
			},
			PodManagementPolicy: appsv1.ParallelPodManagement,
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

func (s *StatefulSetReconciler) getReconcileOperation() (constants.ReconcileOperation, error) {
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
	s.StatefulSet.Status = found.Status
	if equality.Semantic.DeepEqual(s.StatefulSet.Spec, found.Spec) {
		// Update our version so we have Status which can be used
		s.StatefulSet = found
		return constants.ReconcileNoChange, nil
	}
	// Update resource version so we can do a client Update successfully
	s.StatefulSet.SetResourceVersion(found.ResourceVersion)
	return constants.ReconcileUpdateNeeded, nil
}

func (s *StatefulSetReconciler) Reconcile() error {
	logger := s.Logger.WithName("StatefulSetReconcile")
	op, err := s.getReconcileOperation()
	switch op {
	case constants.ReconcileCreateNeeded:
		logger.V(1).Info("StatefulSet Create", "Name", s.StatefulSet.GetName(), "Namespace", s.StatefulSet.GetNamespace())
		err = s.Client.Create(s.Ctx, s.StatefulSet)
		if err != nil {
			logger.Error(err, "Failed to create statefuleset", "Name", s.StatefulSet.GetName(), "Namespace", s.StatefulSet.GetNamespace())
			return err
		}
	case constants.ReconcileUpdateNeeded:
		logger.V(1).Info("StatefulSet Update", "Name", s.StatefulSet.GetName(), "Namespace", s.StatefulSet.GetNamespace())
		err = s.Client.Update(s.Ctx, s.StatefulSet)
		if err != nil {
			logger.Error(err, "Failed to update statefuleset", "Name", s.StatefulSet.GetName(), "Namespace", s.StatefulSet.GetNamespace())
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
)

func (s *StatefulSetReconciler) GetConditions() []*apis.Condition {
	ready := s.StatefulSet.Status.ReadyReplicas >= s.StatefulSet.Status.Replicas
	s.Logger.Info("Checking conditions for stateful set", "ready", ready, "replicas", s.StatefulSet.Status.Replicas, "availableReplicas", s.StatefulSet.Status.AvailableReplicas)
	if ready {
		return []*apis.Condition{mlopsv1alpha1.CreateCondition(mlopsv1alpha1.StatefulSetReady, ready, StatefulSetReadyReason)}
	} else {
		return []*apis.Condition{mlopsv1alpha1.CreateCondition(mlopsv1alpha1.StatefulSetReady, ready, StatefulSetNotReadyReason)}
	}
}
