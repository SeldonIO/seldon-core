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
	"k8s.io/apimachinery/pkg/util/intstr"

	mlopsv1alpha1 "github.com/seldonio/seldon-core/operator/v2/apis/mlops/v1alpha1"
	"github.com/seldonio/seldon-core/operator/v2/controllers/reconcilers/common"
	"github.com/seldonio/seldon-core/operator/v2/pkg/constants"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"knative.dev/pkg/apis"
)

const (
	DefaultXdsPortName           = "xds"
	DefaultXdsPort               = int32(9002)
	DefaultSchedulerPortName     = "scheduler"
	DefaultSchedulerPort         = int32(9004)
	DefaultSchedulerMtlsPortName = "scheduler-mtls"
	DefaultSchedulerMtlsPort     = int32(9044)
	DefaultAgentPortName         = "agent"
	DefaultAgentPort             = int32(9005)
	DefaultAgentMtlsPortName     = "agent-mtls"
	DefaultAgentMtlsPort         = int32(9055)
	DefaultDataflowPortName      = "dataflow"
	DefaultDataflowPort          = int32(9008)
)

type ComponentServiceReconciler struct {
	common.ReconcilerConfig
	meta     metav1.ObjectMeta
	Services []*v1.Service
}

func NewComponentServiceReconciler(
	common common.ReconcilerConfig,
	meta metav1.ObjectMeta,
	overrides map[string]*mlopsv1alpha1.OverrideSpec,
) *ComponentServiceReconciler {
	return &ComponentServiceReconciler{
		ReconcilerConfig: common,
		meta:             meta,
		Services:         toServices(meta, overrides),
	}
}

func (s *ComponentServiceReconciler) GetResources() []metav1.Object {
	var objs []metav1.Object
	for _, svc := range s.Services {
		objs = append(objs, svc)
	}
	return objs
}

func toServices(meta metav1.ObjectMeta, overrides map[string]*mlopsv1alpha1.OverrideSpec) []*v1.Service {
	var svcs []*v1.Service
	svcs = append(svcs, getSchedulerService(meta, overrides[mlopsv1alpha1.SchedulerName]))
	svcs = append(svcs, getSeldonMeshService(meta, overrides[mlopsv1alpha1.EnvoyName]))
	svcs = append(svcs, getPipelinegatewayService(meta, overrides[mlopsv1alpha1.PipelineGatewayName]))
	return svcs
}

func getSeldonMeshService(meta metav1.ObjectMeta, overrides *mlopsv1alpha1.OverrideSpec) *v1.Service {
	serviceType := v1.ServiceTypeLoadBalancer
	if overrides != nil {
		serviceType = overrides.ServiceType
	}
	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      overrides.Name,
			Namespace: meta.GetNamespace(),
			Labels: map[string]string{
				constants.KubernetesNameLabelKey: overrides.Name,
			},
		},
		Spec: v1.ServiceSpec{
			Selector: map[string]string{
				constants.KubernetesNameLabelKey: overrides.Name,
			},
			Type: serviceType,
			Ports: []v1.ServicePort{
				{
					Port:       80,
					TargetPort: intstr.FromString("http"),
					Name:       "http",
					Protocol:   v1.ProtocolTCP,
				},
				{
					Port:       9003,
					TargetPort: intstr.FromString("envoy-admin"),
					Name:       "admin",
					Protocol:   v1.ProtocolTCP,
				},
			},
		},
	}
	return svc
}

func getPipelinegatewayService(meta metav1.ObjectMeta, overrides *mlopsv1alpha1.OverrideSpec) *v1.Service {
	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      overrides.Name,
			Namespace: meta.GetNamespace(),
			Labels: map[string]string{
				constants.KubernetesNameLabelKey: overrides.Name,
			},
		},
		Spec: v1.ServiceSpec{
			ClusterIP: v1.ClusterIPNone,
			Selector: map[string]string{
				constants.KubernetesNameLabelKey: overrides.Name,
			},
			Ports: []v1.ServicePort{
				{
					Port:       9010,
					TargetPort: intstr.FromString("http"),
					Name:       "http",
					Protocol:   v1.ProtocolTCP,
				},
				{
					Port:       9011,
					TargetPort: intstr.FromString("grpc"),
					Name:       "gprc",
					Protocol:   v1.ProtocolTCP,
				},
			},
		},
	}
	return svc
}

func getSchedulerService(meta metav1.ObjectMeta, overrides *mlopsv1alpha1.OverrideSpec) *v1.Service {
	serviceType := v1.ServiceTypeLoadBalancer
	if overrides != nil {
		serviceType = overrides.ServiceType
	}
	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      overrides.Name,
			Namespace: meta.GetNamespace(),
			Labels: map[string]string{
				constants.KubernetesNameLabelKey: overrides.Name,
			},
		},
		Spec: v1.ServiceSpec{
			Selector: map[string]string{
				constants.KubernetesNameLabelKey: overrides.Name,
			},
			Type: serviceType,
			Ports: []v1.ServicePort{
				{
					Port:       DefaultXdsPort,
					TargetPort: intstr.FromString(DefaultXdsPortName),
					Name:       DefaultXdsPortName,
					Protocol:   v1.ProtocolTCP,
				},
				{
					Port:       DefaultSchedulerPort,
					TargetPort: intstr.FromString(DefaultSchedulerPortName),
					Name:       DefaultSchedulerPortName,
					Protocol:   v1.ProtocolTCP,
				},
				{
					Port:       DefaultSchedulerMtlsPort,
					TargetPort: intstr.FromString(DefaultSchedulerMtlsPortName),
					Name:       DefaultSchedulerMtlsPortName,
					Protocol:   v1.ProtocolTCP,
				},
				{
					Port:       DefaultAgentPort,
					TargetPort: intstr.FromString(DefaultAgentPortName),
					Name:       DefaultAgentPortName,
					Protocol:   v1.ProtocolTCP,
				},
				{
					Port:       DefaultAgentMtlsPort,
					TargetPort: intstr.FromString(DefaultAgentMtlsPortName),
					Name:       DefaultAgentMtlsPortName,
					Protocol:   v1.ProtocolTCP,
				},
				{
					Port:       DefaultDataflowPort,
					TargetPort: intstr.FromString(DefaultDataflowPortName),
					Name:       DefaultDataflowPortName,
					Protocol:   v1.ProtocolTCP,
				},
			},
		},
	}
	return svc
}

func (s *ComponentServiceReconciler) getReconcileOperation(idx int, svc *v1.Service) (constants.ReconcileOperation, error) {
	found := &v1.Service{}
	err := s.Client.Get(context.TODO(), types.NamespacedName{
		Name:      svc.GetName(),
		Namespace: svc.GetNamespace(),
	}, found)
	if err != nil {
		if errors.IsNotFound(err) {
			return constants.ReconcileCreateNeeded, nil
		}
		return constants.ReconcileUnknown, err
	}
	if equality.Semantic.DeepEqual(svc.Spec, found.Spec) {
		// Update our version so we have Status if needed
		s.Services[idx] = found
		return constants.ReconcileNoChange, nil
	}
	// Update resource vesion so the client Update will succeed
	s.Services[idx].SetResourceVersion(found.ResourceVersion)
	return constants.ReconcileUpdateNeeded, nil
}

func (s *ComponentServiceReconciler) Reconcile() error {
	logger := s.Logger.WithName("ServiceReconcile")
	for idx, svc := range s.Services {
		op, err := s.getReconcileOperation(idx, svc)
		switch op {
		case constants.ReconcileCreateNeeded:
			logger.V(1).Info("Service Create", "Name", svc.GetName(), "Namespace", svc.GetNamespace())
			err = s.Client.Create(s.Ctx, svc)
			if err != nil {
				logger.Error(err, "Failed to create service", "Name", svc.GetName(), "Namespace", svc.GetNamespace())
				return err
			}
		case constants.ReconcileUpdateNeeded:
			logger.V(1).Info("Service Update", "Name", svc.GetName(), "Namespace", svc.GetNamespace())
			err = s.Client.Update(s.Ctx, svc)
			if err != nil {
				logger.Error(err, "Failed to update service", "Name", svc.GetName(), "Namespace", svc.GetNamespace())
				return err
			}
		case constants.ReconcileNoChange:
			logger.V(1).Info("Service No Change", "Name", svc.GetName(), "Namespace", svc.GetNamespace())
		case constants.ReconcileUnknown:
			logger.Error(err, "Failed to get reconcile operation for service", "Name", svc.GetName(), "Namespace", svc.GetNamespace())
			return err
		}
	}
	return nil
}

func (s *ComponentServiceReconciler) GetConditions() []*apis.Condition {
	return nil
}
