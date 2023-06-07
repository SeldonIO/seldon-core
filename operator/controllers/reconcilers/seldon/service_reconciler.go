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
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"knative.dev/pkg/apis"
	"sigs.k8s.io/controller-runtime/pkg/client"

	mlopsv1alpha1 "github.com/seldonio/seldon-core/operator/v2/apis/mlops/v1alpha1"
	"github.com/seldonio/seldon-core/operator/v2/controllers/reconcilers/common"
	"github.com/seldonio/seldon-core/operator/v2/pkg/constants"
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

const (
	SeldonMeshSVCName = "seldon-mesh"
)

type ComponentServiceReconciler struct {
	common.ReconcilerConfig
	meta      metav1.ObjectMeta
	Services  []*v1.Service
	Annotator *patch.Annotator
}

func NewComponentServiceReconciler(
	common common.ReconcilerConfig,
	meta metav1.ObjectMeta,
	serviceConfig mlopsv1alpha1.ServiceConfig,
	overrides map[string]*mlopsv1alpha1.OverrideSpec,
	annotator *patch.Annotator,
) *ComponentServiceReconciler {
	return &ComponentServiceReconciler{
		ReconcilerConfig: common,
		meta:             meta,
		Services:         toServices(meta, serviceConfig, overrides),
		Annotator:        annotator,
	}
}

func (s *ComponentServiceReconciler) GetResources() []client.Object {
	var objs []client.Object
	for _, svc := range s.Services {
		objs = append(objs, svc)
	}
	return objs
}

func toServices(meta metav1.ObjectMeta, serviceConfig mlopsv1alpha1.ServiceConfig, overrides map[string]*mlopsv1alpha1.OverrideSpec) []*v1.Service {
	var svcs []*v1.Service
	svcs = append(svcs, getSchedulerService(meta, serviceConfig, overrides[mlopsv1alpha1.SchedulerName]))
	svcs = append(svcs, getSeldonMeshService(meta, serviceConfig, overrides[mlopsv1alpha1.EnvoyName]))
	svcs = append(svcs, getPipelinegatewayService(meta, overrides[mlopsv1alpha1.PipelineGatewayName]))
	return svcs
}

func getServiceType(serviceConfig mlopsv1alpha1.ServiceConfig, overrides *mlopsv1alpha1.OverrideSpec) v1.ServiceType {
	serviceType := v1.ServiceTypeLoadBalancer
	if serviceConfig.ServiceType != "" {
		serviceType = serviceConfig.ServiceType
	}
	if overrides != nil {
		serviceType = overrides.ServiceType
	}
	return serviceType
}

func getSeldonMeshService(meta metav1.ObjectMeta, serviceConfig mlopsv1alpha1.ServiceConfig, overrides *mlopsv1alpha1.OverrideSpec) *v1.Service {
	serviceType := getServiceType(serviceConfig, overrides)
	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      SeldonMeshSVCName,
			Namespace: meta.GetNamespace(),
			Labels: map[string]string{
				constants.KubernetesNameLabelKey: mlopsv1alpha1.EnvoyName,
			},
		},
		Spec: v1.ServiceSpec{
			Selector: map[string]string{
				constants.KubernetesNameLabelKey: mlopsv1alpha1.EnvoyName,
			},
			Type: serviceType,
			Ports: []v1.ServicePort{
				{
					Port:       80,
					TargetPort: intstr.FromString("http"),
					Name:       fmt.Sprintf("%sdata", serviceConfig.GrpcServicePrefix),
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
			Name:      mlopsv1alpha1.PipelineGatewayName,
			Namespace: meta.GetNamespace(),
			Labels: map[string]string{
				constants.KubernetesNameLabelKey: mlopsv1alpha1.PipelineGatewayName,
			},
		},
		Spec: v1.ServiceSpec{
			ClusterIP: v1.ClusterIPNone,
			Selector: map[string]string{
				constants.KubernetesNameLabelKey: mlopsv1alpha1.PipelineGatewayName,
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

func getSchedulerService(meta metav1.ObjectMeta, serviceConfig mlopsv1alpha1.ServiceConfig, overrides *mlopsv1alpha1.OverrideSpec) *v1.Service {
	serviceType := getServiceType(serviceConfig, overrides)
	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      mlopsv1alpha1.SchedulerName,
			Namespace: meta.GetNamespace(),
			Labels: map[string]string{
				constants.KubernetesNameLabelKey: mlopsv1alpha1.SchedulerName,
			},
		},
		Spec: v1.ServiceSpec{
			Selector: map[string]string{
				constants.KubernetesNameLabelKey: mlopsv1alpha1.SchedulerName,
			},
			Type: serviceType,
			Ports: []v1.ServicePort{
				{
					Port:       DefaultXdsPort,
					TargetPort: intstr.FromString(DefaultXdsPortName),
					Name:       fmt.Sprintf("%s%s", serviceConfig.GrpcServicePrefix, DefaultXdsPortName),
					Protocol:   v1.ProtocolTCP,
				},
				{
					Port:       DefaultSchedulerPort,
					TargetPort: intstr.FromString(DefaultSchedulerPortName),
					Name:       fmt.Sprintf("%s%s", serviceConfig.GrpcServicePrefix, DefaultSchedulerPortName),
					Protocol:   v1.ProtocolTCP,
				},
				{
					Port:       DefaultSchedulerMtlsPort,
					TargetPort: intstr.FromString(DefaultSchedulerMtlsPortName),
					Name:       fmt.Sprintf("%s%s", serviceConfig.GrpcServicePrefix, DefaultSchedulerMtlsPortName),
					Protocol:   v1.ProtocolTCP,
				},
				{
					Port:       DefaultAgentPort,
					TargetPort: intstr.FromString(DefaultAgentPortName),
					Name:       fmt.Sprintf("%s%s", serviceConfig.GrpcServicePrefix, DefaultAgentPortName),
					Protocol:   v1.ProtocolTCP,
				},
				{
					Port:       DefaultAgentMtlsPort,
					TargetPort: intstr.FromString(DefaultAgentMtlsPortName),
					Name:       fmt.Sprintf("%s%s", serviceConfig.GrpcServicePrefix, DefaultAgentMtlsPortName),
					Protocol:   v1.ProtocolTCP,
				},
				{
					Port:       DefaultDataflowPort,
					TargetPort: intstr.FromString(DefaultDataflowPortName),
					Name:       fmt.Sprintf("%s%s", serviceConfig.GrpcServicePrefix, DefaultDataflowPortName),
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
	opts := []patch.CalculateOption{
		patch.IgnoreStatusFields(),
		patch.IgnoreField("kind"),
		patch.IgnoreField("apiVersion"),
		patch.IgnoreField("metadata"),
	}
	patcherMaker := patch.NewPatchMaker(s.Annotator, &patch.K8sStrategicMergePatcher{}, &patch.BaseJSONMergePatcher{})
	patchResult, err := patcherMaker.Calculate(found, svc, opts...)
	if err != nil {
		return constants.ReconcileUnknown, err
	}
	if patchResult.IsEmpty() {
		// Update our version so we have Status if needed
		s.Services[idx] = found
		return constants.ReconcileNoChange, nil
	}
	err = s.Annotator.SetLastAppliedAnnotation(svc)
	if err != nil {
		return constants.ReconcileUnknown, err
	}
	// Update resource version so we can do a client Update successfully
	// This needs to be done after we annotate to also avoid false differences
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
