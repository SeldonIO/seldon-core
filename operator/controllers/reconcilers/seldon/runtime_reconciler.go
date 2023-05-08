package server

import (
	mlopsv1alpha1 "github.com/seldonio/seldon-core/operator/v2/apis/mlops/v1alpha1"
	"github.com/seldonio/seldon-core/operator/v2/controllers/reconcilers/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis"
)

type SeldonRuntimeReconciler struct {
	common.ReconcilerConfig
	componentReconcilers []common.Reconciler
	rbacReconciler       common.Reconciler
	serviceReconciler    common.Reconciler
}

func NewSeldonRuntimeReconciler(
	runtime *mlopsv1alpha1.SeldonRuntime,
	commonConfig common.ReconcilerConfig) (common.Reconciler, error) {

	var err error

	seldonConfig, err := mlopsv1alpha1.GetSeldonConfigForSeldonRuntime(runtime.Spec.SeldonConfig, commonConfig.Client)
	if err != nil {
		return nil, err
	}

	var overrides map[string]*mlopsv1alpha1.OverrideSpec
	for _, o := range runtime.Spec.Overrides {
		overrides[o.Name] = o
	}

	var componentReconcilers []common.Reconciler
	for _, c := range seldonConfig.Spec.Components {
		if c.Stateful {
			componentReconcilers = append(componentReconcilers,
				NewComponentStatefulSetReconciler(
					c.Name,
					commonConfig,
					runtime.ObjectMeta,
					c.PodSpec,
					c.VolumeClaimTemplates,
					overrides[c.Name],
					seldonConfig.ObjectMeta))
		} else {
			componentReconcilers = append(componentReconcilers,
				NewComponentDeploymentReconciler(
					c.Name,
					commonConfig,
					runtime.ObjectMeta,
					c.PodSpec,
					overrides[c.Name],
					seldonConfig.ObjectMeta))
		}
	}

	return &SeldonRuntimeReconciler{
		ReconcilerConfig:     commonConfig,
		componentReconcilers: componentReconcilers,
		rbacReconciler:       NewComponentRBACReconciler(commonConfig, runtime.ObjectMeta),
		serviceReconciler:    NewComponentServiceReconciler(commonConfig, runtime.ObjectMeta, overrides),
	}, nil
}

func (s *SeldonRuntimeReconciler) GetResources() []metav1.Object {
	var objs []metav1.Object
	for _, c := range s.componentReconcilers {
		objs = append(objs, c.GetResources()...)
	}
	objs = append(objs, s.rbacReconciler.GetResources()...)
	objs = append(objs, s.serviceReconciler.GetResources()...)
	return objs
}

func (s *SeldonRuntimeReconciler) GetConditions() []*apis.Condition {
	var conditions []*apis.Condition
	for _, c := range s.componentReconcilers {
		conditions = append(conditions, c.GetConditions()...)
	}
	conditions = append(conditions, s.rbacReconciler.GetConditions()...)
	conditions = append(conditions, s.serviceReconciler.GetConditions()...)
	return conditions
}

func (s *SeldonRuntimeReconciler) Reconcile() error {
	err := s.rbacReconciler.Reconcile()
	if err != nil {
		return err
	}
	err = s.serviceReconciler.Reconcile()
	if err != nil {
		return err
	}
	for _, c := range s.componentReconcilers {
		err := c.Reconcile()
		if err != nil {
			return err
		}
	}
	return nil
}
