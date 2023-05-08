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
	"github.com/seldonio/seldon-core/operator/v2/pkg/utils"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"knative.dev/pkg/apis"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	StatefulSetPodLabel  = "statefulset.kubernetes.io/pod-name"
	DefaultHttpPortName  = "http"
	DefaultHttpPort      = int32(9000)
	DefaultGrpcPortName  = "grpc"
	DefaultGrpcPort      = int32(9500)
	DefaultAgentPortName = "agent"
	DefaultAgentPort     = int32(9005)
)

type ServerServiceReconciler struct {
	common.ReconcilerConfig
	meta      metav1.ObjectMeta
	Services  []*v1.Service
	Namespace string
}

func NewServerServiceReconciler(
	common common.ReconcilerConfig,
	meta metav1.ObjectMeta,
	scaling *mlopsv1alpha1.ScalingSpec) *ServerServiceReconciler {
	return &ServerServiceReconciler{
		ReconcilerConfig: common,
		meta:             meta,
		Services:         toServices(meta, int(*scaling.Replicas)),
		Namespace:        meta.Namespace,
	}
}

func (s *ServerServiceReconciler) GetResources() []metav1.Object {
	var objs []metav1.Object
	for _, svc := range s.Services {
		objs = append(objs, svc)
	}
	return objs
}

func toServices(meta metav1.ObjectMeta, replicas int) []*v1.Service {
	var svcs []*v1.Service

	// Create StatefulSet Services for each Replica
	for i := 0; i < replicas; i++ {
		name := utils.GetStatefulSetReplicaName(meta.GetName(), i)
		svc := &v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: meta.GetNamespace(),
				Labels: map[string]string{
					constants.ServerReplicaLabelKey:     meta.GetName(), // Shared value for all replicas
					constants.ServerReplicaNameLabelKey: name,           //Unique for this replica
				},
			},
			Spec: v1.ServiceSpec{
				ClusterIP: "None",
				Selector: map[string]string{
					StatefulSetPodLabel: name,
				},
				Ports: []v1.ServicePort{
					{
						Port:       DefaultHttpPort,
						TargetPort: intstr.FromString(DefaultHttpPortName),
						Name:       DefaultHttpPortName,
					},
					{
						Port:       DefaultGrpcPort,
						TargetPort: intstr.FromString(DefaultGrpcPortName),
						Name:       DefaultGrpcPortName,
					},
					{
						Port: DefaultAgentPort,
						Name: DefaultAgentPortName,
					},
				},
			},
		}
		svcs = append(svcs, svc)
	}
	return svcs
}

func (s *ServerServiceReconciler) getReconcileOperation(idx int, svc *v1.Service) (constants.ReconcileOperation, error) {
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

// Get the expected number of replicas of server specific services that currently should exist
// This is found by extracting the annotation added to the main svc
func (s *ServerServiceReconciler) getCurrentSvcReplicas() (int, error) {
	founds := &v1.ServiceList{}
	matchingLabel := client.MatchingLabels{constants.ServerReplicaLabelKey: s.meta.Name}
	namespace := client.InNamespace(s.Namespace)
	err := s.Client.List(s.Ctx, founds, matchingLabel, namespace)
	if err != nil {
		if errors.IsNotFound(err) {
			return 0, nil
		}
		return 0, err
	}
	return len(founds.Items), nil
}

// Delete svc replicas not needed
func (s *ServerServiceReconciler) removeExtraSvcReplicas() error {
	existingReplicas, err := s.getCurrentSvcReplicas()
	if err != nil {
		return err
	}
	numReplicas := len(s.Services)
	if existingReplicas > numReplicas {
		svcsNow := toServices(s.meta, existingReplicas)
		for i := numReplicas; i < existingReplicas; i++ {
			err = s.Client.Delete(s.Ctx, svcsNow[i])
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *ServerServiceReconciler) Reconcile() error {
	logger := s.Logger.WithName("ServiceReconcile")
	err := s.removeExtraSvcReplicas()
	if err != nil {
		return err
	}
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

func (s *ServerServiceReconciler) GetConditions() []*apis.Condition {
	return nil
}
