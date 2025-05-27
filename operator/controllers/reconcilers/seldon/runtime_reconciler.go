/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package server

import (
	"github.com/banzaicloud/k8s-objectmatcher/patch"
	"knative.dev/pkg/apis"
	"sigs.k8s.io/controller-runtime/pkg/client"

	mlopsv1alpha1 "github.com/seldonio/seldon-core/operator/v2/apis/mlops/v1alpha1"
	"github.com/seldonio/seldon-core/operator/v2/controllers/reconcilers/common"
	"github.com/seldonio/seldon-core/operator/v2/pkg/constants"
)

type SeldonRuntimeReconciler struct {
	common.ReconcilerConfig
	componentReconcilers []common.Reconciler
	rbacReconciler       common.Reconciler
	serviceReconciler    common.Reconciler
	configMapReconciler  common.Reconciler
}

func NewSeldonRuntimeReconciler(
	runtime *mlopsv1alpha1.SeldonRuntime,
	commonConfig common.ReconcilerConfig,
) (common.Reconciler, error) {
	var err error

	seldonConfig, err := mlopsv1alpha1.GetSeldonConfigForSeldonRuntime(runtime.Spec.SeldonConfig, commonConfig.Client)
	if err != nil {
		return nil, err
	}

	overrides := make(map[string]*mlopsv1alpha1.OverrideSpec)
	for _, o := range runtime.Spec.Overrides {
		overrides[o.Name] = o
	}

	annotator := patch.NewAnnotator(constants.LastAppliedConfig)

	var componentReconcilers []common.Reconciler
	for _, c := range seldonConfig.Spec.Components {
		override := overrides[c.Name]
		commonConfig.Logger.Info("Creating reconciler", "name", c.Name, "has override", override != nil)
		if override == nil || !override.Disable {
			commonConfig.Logger.Info("Creating component", "name", c.Name)
			if len(c.VolumeClaimTemplates) > 0 {
				statefulSetReconcilor, err := NewComponentStatefulSetReconciler(
					c.Name,
					commonConfig,
					runtime.ObjectMeta,
					c.PodSpec,
					c.VolumeClaimTemplates,
					c.Labels,
					c.Annotations,
					overrides[c.Name],
					seldonConfig.ObjectMeta,
					annotator,
				)
				if err != nil {
					return nil, err
				}
				componentReconcilers = append(componentReconcilers, statefulSetReconcilor)

			} else {
				deploymentReconcilor, err := NewComponentDeploymentReconciler(
					c.Name,
					commonConfig,
					runtime.ObjectMeta,
					c.PodSpec,
					c.Labels,
					c.Annotations,
					overrides[c.Name],
					seldonConfig.ObjectMeta,
					annotator,
				)
				if err != nil {
					return nil, err
				}
				componentReconcilers = append(componentReconcilers, deploymentReconcilor)
			}
		} else {
			commonConfig.Logger.Info("Disabling component", "name", c.Name)
		}
	}
	// Set last applied annotation for update
	for _, cr := range componentReconcilers {
		for _, res := range cr.GetResources() {
			if err := annotator.SetLastAppliedAnnotation(res); err != nil {
				return nil, err
			}
		}
	}

	runtime.Spec.Config.AddDefaults(seldonConfig.Spec.Config)
	configMapReconciler, err := NewConfigMapReconciler(commonConfig, &runtime.Spec.Config, runtime.ObjectMeta)
	if err != nil {
		return nil, err
	}

	svcReconciler := NewComponentServiceReconciler(commonConfig, runtime.ObjectMeta, runtime.Spec.Config.ServiceConfig, overrides, annotator)
	for _, res := range svcReconciler.GetResources() {
		if err := annotator.SetLastAppliedAnnotation(res); err != nil {
			return nil, err
		}
	}

	return &SeldonRuntimeReconciler{
		ReconcilerConfig:     commonConfig,
		componentReconcilers: componentReconcilers,
		rbacReconciler:       NewComponentRBACReconciler(commonConfig, runtime.ObjectMeta),
		serviceReconciler:    svcReconciler,
		configMapReconciler:  configMapReconciler,
	}, nil
}

func (s *SeldonRuntimeReconciler) GetResources() []client.Object {
	var objs []client.Object
	for _, c := range s.componentReconcilers {
		objs = append(objs, c.GetResources()...)
	}
	objs = append(objs, s.rbacReconciler.GetResources()...)
	objs = append(objs, s.serviceReconciler.GetResources()...)
	objs = append(objs, s.configMapReconciler.GetResources()...)
	return objs
}

func (s *SeldonRuntimeReconciler) GetConditions() []*apis.Condition {
	var conditions []*apis.Condition
	for _, c := range s.componentReconcilers {
		conditions = append(conditions, c.GetConditions()...)
	}
	conditions = append(conditions, s.rbacReconciler.GetConditions()...)
	conditions = append(conditions, s.serviceReconciler.GetConditions()...)
	conditions = append(conditions, s.configMapReconciler.GetConditions()...)
	return conditions
}

func (s *SeldonRuntimeReconciler) Reconcile() error {
	err := s.rbacReconciler.Reconcile()
	if err != nil {
		return err
	}
	err = s.configMapReconciler.Reconcile()
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
