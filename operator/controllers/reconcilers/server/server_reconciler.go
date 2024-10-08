/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package server

import (
	"fmt"
	"strings"

	"github.com/banzaicloud/k8s-objectmatcher/patch"
	v1 "k8s.io/api/core/v1"
	"knative.dev/pkg/apis"
	"sigs.k8s.io/controller-runtime/pkg/client"

	mlopsv1alpha1 "github.com/seldonio/seldon-core/operator/v2/apis/mlops/v1alpha1"
	"github.com/seldonio/seldon-core/operator/v2/controllers/reconcilers/common"
	"github.com/seldonio/seldon-core/operator/v2/pkg/constants"
)

const (
	EnvVarNameCapabilities = "SELDON_SERVER_CAPABILITIES"
)

type ServerReconciler struct {
	common.ReconcilerConfig
	StatefulSetReconciler common.Reconciler
	ServiceReconciler     common.Reconciler
}

func NewServerReconciler(server *mlopsv1alpha1.Server,
	common common.ReconcilerConfig) (common.Reconciler, error) {
	// Ensure defaults added to server
	server.Default()

	var err error
	sr := &ServerReconciler{
		ReconcilerConfig: common,
	}

	annotator := patch.NewAnnotator(constants.LastAppliedConfig)

	sr.StatefulSetReconciler, err = sr.createStatefulSetReconciler(server, annotator)
	if err != nil {
		return nil, err
	}

	// Add last applied annotation to all resources
	for _, res := range sr.StatefulSetReconciler.GetResources() {
		if err := annotator.SetLastAppliedAnnotation(res); err != nil {
			return nil, err
		}
	}

	sr.ServiceReconciler = NewServerServiceReconciler(common, server.ObjectMeta, &server.Spec.ScalingSpec)
	return sr, nil
}

func (s *ServerReconciler) GetLabelSelector() string {
	return s.StatefulSetReconciler.(common.LabelHandler).GetLabelSelector()
}

func (s *ServerReconciler) GetReplicas() (int32, error) {
	return s.StatefulSetReconciler.(common.ReplicaHandler).GetReplicas()
}

func (s *ServerReconciler) GetResources() []client.Object {
	objs := s.StatefulSetReconciler.GetResources()
	objs = append(objs, s.ServiceReconciler.GetResources()...)
	return objs
}

func (s *ServerReconciler) GetConditions() []*apis.Condition {
	conditions := s.StatefulSetReconciler.GetConditions()
	conditions = append(conditions, s.ServiceReconciler.GetConditions()...)
	return conditions
}

func (s *ServerReconciler) Reconcile() error {
	// Reconcile Services
	err := s.ServiceReconciler.Reconcile()
	if err != nil {
		return err
	}
	// Reconcile StatefulSet
	err = s.StatefulSetReconciler.Reconcile()
	if err != nil {
		return err
	}

	return nil
}

func updateCapabilities(capabilities []string, extraCapabilities []string, podSpec *v1.PodSpec) {
	if len(extraCapabilities) > 0 || len(capabilities) > 0 {
		for _, container := range podSpec.Containers {
			for idx, envVar := range container.Env {
				if envVar.Name == EnvVarNameCapabilities {
					if len(capabilities) > 0 {
						capabilitiesStr := strings.Join(capabilities, ",")
						container.Env[idx] = v1.EnvVar{Name: EnvVarNameCapabilities, Value: capabilitiesStr}
					} else { // Deprecated
						capabilitiesStr := strings.Join(extraCapabilities, ",")
						val := fmt.Sprintf("%s,%s", envVar.Value, capabilitiesStr)
						container.Env[idx] = v1.EnvVar{Name: EnvVarNameCapabilities, Value: val}
					}
				}
			}
		}
	}
}

func (s *ServerReconciler) createStatefulSetReconciler(server *mlopsv1alpha1.Server, annotator *patch.Annotator) (*ServerStatefulSetReconciler, error) {
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
	statefulSetReconciler := NewServerStatefulSetReconciler(s.ReconcilerConfig,
		server.ObjectMeta,
		podSpec,
		serverConfig.Spec.VolumeClaimTemplates,
		&server.Spec.ScalingSpec,
		server.Spec.StatefulSetPersistentVolumeClaimRetentionPolicy,
		serverConfig.ObjectMeta,
		annotator)
	return statefulSetReconciler, nil
}
