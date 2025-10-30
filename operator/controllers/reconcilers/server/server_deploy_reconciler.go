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

	"github.com/banzaicloud/k8s-objectmatcher/patch"
	"knative.dev/pkg/apis"
	"sigs.k8s.io/controller-runtime/pkg/client"

	mlopsv1alpha1 "github.com/seldonio/seldon-core/operator/v2/apis/mlops/v1alpha1"
	"github.com/seldonio/seldon-core/operator/v2/controllers/reconcilers/common"
	"github.com/seldonio/seldon-core/operator/v2/pkg/constants"
)

type ServerReconcilerWithDeployment struct {
	common.ReconcilerConfig
	DeploymentReconciler common.Reconciler
}

func NewServerReconcilerWithDeployment(
	server *mlopsv1alpha1.Server,
	common common.ReconcilerConfig,
) (common.Reconciler, error) {
	// Ensure defaults added to server
	server.Default()

	var err error
	sr := &ServerReconcilerWithDeployment{ReconcilerConfig: common}
	annotator := patch.NewAnnotator(constants.LastAppliedConfig)

	sr.DeploymentReconciler, err = sr.createDeploymentReconciler(server, annotator)
	if err != nil {
		return nil, err
	}

	// Add last applied annotation to all resources
	for _, res := range sr.DeploymentReconciler.GetResources() {
		if err := annotator.SetLastAppliedAnnotation(res); err != nil {
			return nil, err
		}
	}

	return sr, nil
}

func (s *ServerReconcilerWithDeployment) GetLabelSelector() string {
	return s.DeploymentReconciler.(common.LabelHandler).GetLabelSelector()
}

func (s *ServerReconcilerWithDeployment) GetReplicas(ctx context.Context) (int32, error) {
	return s.DeploymentReconciler.(common.ReplicaHandler).GetReplicas(ctx)
}

func (s *ServerReconcilerWithDeployment) GetResources() []client.Object {
	objs := s.DeploymentReconciler.GetResources()
	return objs
}

func (s *ServerReconcilerWithDeployment) GetConditions() []*apis.Condition {
	conditions := s.DeploymentReconciler.GetConditions()
	return conditions
}

func (s *ServerReconcilerWithDeployment) Reconcile(ctx context.Context) error {
	err := s.DeploymentReconciler.Reconcile(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (s *ServerReconcilerWithDeployment) createDeploymentReconciler(server *mlopsv1alpha1.Server, annotator *patch.Annotator) (*ServerDeploymentReconciler, error) {
	//Get ServerConfig
	serverConfig, err := mlopsv1alpha1.GetServerConfigForServer(server.Spec.ServerConfig, s.Client)
	if err != nil {
		return nil, err
	}

	//Merge specs
	podSpec, err := common.MergePodSpecs(&serverConfig.Spec.PodSpec, server.Spec.PodSpec)
	if err != nil {
		return nil, err
	}

	// Update capabilities
	updateCapabilities(server.Spec.Capabilities, server.Spec.ExtraCapabilities, podSpec)

	// Reconcile ReplicaSet
	deploymentReconciler := NewServerDeploymentReconciler(s.ReconcilerConfig,
		server.ObjectMeta,
		podSpec,
		&server.Spec.ScalingSpec,
		&server.Spec.DeploymentStrategy,
		serverConfig.ObjectMeta,
		annotator)
	return deploymentReconciler, nil
}
