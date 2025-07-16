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
	"testing"

	"github.com/banzaicloud/k8s-objectmatcher/patch"
	logrtest "github.com/go-logr/logr/testr"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	auth "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	client2 "sigs.k8s.io/controller-runtime/pkg/client"

	mlopsv1alpha1 "github.com/seldonio/seldon-core/operator/v2/apis/mlops/v1alpha1"
	"github.com/seldonio/seldon-core/operator/v2/controllers/reconcilers/common"
	"github.com/seldonio/seldon-core/operator/v2/pkg/constants"
	testing2 "github.com/seldonio/seldon-core/operator/v2/pkg/utils/testing"
)

func TestServiceReconcile(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name             string
		serviceConfig    mlopsv1alpha1.ServiceConfig
		runtime          *mlopsv1alpha1.OverrideSpec
		overrides        map[string]*mlopsv1alpha1.OverrideSpec
		expectedSvcNames []string
		expectedSvcType  map[string]v1.ServiceType
	}
	tests := []test{
		{
			name: "normal services",
			serviceConfig: mlopsv1alpha1.ServiceConfig{
				GrpcServicePrefix: "",
			},
			overrides:        map[string]*mlopsv1alpha1.OverrideSpec{},
			expectedSvcNames: []string{SeldonMeshSVCName, mlopsv1alpha1.SchedulerName, mlopsv1alpha1.PipelineGatewayName},
			expectedSvcType: map[string]v1.ServiceType{
				SeldonMeshSVCName:                 v1.ServiceTypeLoadBalancer,
				mlopsv1alpha1.PipelineGatewayName: "",
			},
		},
		{
			name: "prefix services",
			serviceConfig: mlopsv1alpha1.ServiceConfig{
				GrpcServicePrefix: "grpc-",
			},
			overrides:        map[string]*mlopsv1alpha1.OverrideSpec{},
			expectedSvcNames: []string{SeldonMeshSVCName, mlopsv1alpha1.SchedulerName, mlopsv1alpha1.PipelineGatewayName},
			expectedSvcType: map[string]v1.ServiceType{
				SeldonMeshSVCName:                 v1.ServiceTypeLoadBalancer,
				mlopsv1alpha1.PipelineGatewayName: "",
			},
		},
		{
			name: "scheduler disabled",
			serviceConfig: mlopsv1alpha1.ServiceConfig{
				GrpcServicePrefix: "",
			},
			overrides: map[string]*mlopsv1alpha1.OverrideSpec{
				mlopsv1alpha1.SchedulerName: {
					Name:    mlopsv1alpha1.SchedulerName,
					Disable: true,
				},
			},
			expectedSvcNames: []string{SeldonMeshSVCName, mlopsv1alpha1.PipelineGatewayName},
			expectedSvcType: map[string]v1.ServiceType{
				SeldonMeshSVCName:                 v1.ServiceTypeLoadBalancer,
				mlopsv1alpha1.PipelineGatewayName: "",
			},
		},
		{
			name: "envoy disabled",
			serviceConfig: mlopsv1alpha1.ServiceConfig{
				GrpcServicePrefix: "",
			},
			overrides: map[string]*mlopsv1alpha1.OverrideSpec{
				mlopsv1alpha1.EnvoyName: {
					Name:    mlopsv1alpha1.EnvoyName,
					Disable: true,
				},
			},
			expectedSvcNames: []string{mlopsv1alpha1.SchedulerName, mlopsv1alpha1.PipelineGatewayName},
			expectedSvcType: map[string]v1.ServiceType{
				SeldonMeshSVCName:                 v1.ServiceTypeLoadBalancer,
				mlopsv1alpha1.PipelineGatewayName: "",
			},
		},
		{
			name: "pipeline gateway disabled",
			serviceConfig: mlopsv1alpha1.ServiceConfig{
				GrpcServicePrefix: "",
			},
			overrides: map[string]*mlopsv1alpha1.OverrideSpec{
				mlopsv1alpha1.PipelineGatewayName: {
					Name:    mlopsv1alpha1.PipelineGatewayName,
					Disable: true,
				},
			},
			expectedSvcNames: []string{SeldonMeshSVCName, mlopsv1alpha1.SchedulerName},
			expectedSvcType: map[string]v1.ServiceType{
				SeldonMeshSVCName: v1.ServiceTypeLoadBalancer,
			},
		},
		{
			name: "all components disabled",
			serviceConfig: mlopsv1alpha1.ServiceConfig{
				GrpcServicePrefix: "",
			},
			overrides: map[string]*mlopsv1alpha1.OverrideSpec{
				mlopsv1alpha1.SchedulerName: {
					Name:    mlopsv1alpha1.SchedulerName,
					Disable: true,
				},
				mlopsv1alpha1.EnvoyName: {
					Name:    mlopsv1alpha1.EnvoyName,
					Disable: true,
				},
				mlopsv1alpha1.PipelineGatewayName: {
					Name:    mlopsv1alpha1.PipelineGatewayName,
					Disable: true,
				},
			},
			expectedSvcNames: []string{}, // No services should be created
			expectedSvcType:  map[string]v1.ServiceType{},
		},
		{
			name: "mixed disabled and enabled",
			serviceConfig: mlopsv1alpha1.ServiceConfig{
				GrpcServicePrefix: "",
			},
			overrides: map[string]*mlopsv1alpha1.OverrideSpec{
				mlopsv1alpha1.SchedulerName: {
					Name:    mlopsv1alpha1.SchedulerName,
					Disable: true, // Disabled
				},
				mlopsv1alpha1.EnvoyName: {
					Name:        mlopsv1alpha1.EnvoyName,
					Disable:     false, // Explicitly enabled
					ServiceType: v1.ServiceTypeClusterIP,
				},
				mlopsv1alpha1.PipelineGatewayName: {
					Name: mlopsv1alpha1.PipelineGatewayName,
					// Not disabled, should be created
				},
			},
			expectedSvcNames: []string{SeldonMeshSVCName, mlopsv1alpha1.PipelineGatewayName},
			expectedSvcType: map[string]v1.ServiceType{
				SeldonMeshSVCName:                 v1.ServiceTypeClusterIP, // Override applied
				mlopsv1alpha1.PipelineGatewayName: "",
			},
		},
		{
			name: "scheduler replicas 0 should not create service",
			serviceConfig: mlopsv1alpha1.ServiceConfig{
				GrpcServicePrefix: "",
			},
			overrides: map[string]*mlopsv1alpha1.OverrideSpec{
				mlopsv1alpha1.SchedulerName: {
					Name:     mlopsv1alpha1.SchedulerName,
					Replicas: int32Ptr(0), // Scaled to zero
				},
			},
			expectedSvcNames: []string{SeldonMeshSVCName, mlopsv1alpha1.PipelineGatewayName},
			expectedSvcType: map[string]v1.ServiceType{
				SeldonMeshSVCName:                 v1.ServiceTypeLoadBalancer,
				mlopsv1alpha1.PipelineGatewayName: "",
			},
		},
		{
			name: "mixed replicas 0 and disable true",
			serviceConfig: mlopsv1alpha1.ServiceConfig{
				GrpcServicePrefix: "",
			},
			overrides: map[string]*mlopsv1alpha1.OverrideSpec{
				mlopsv1alpha1.SchedulerName: {
					Name:     mlopsv1alpha1.SchedulerName,
					Replicas: int32Ptr(0), // Scaled to zero
				},
				mlopsv1alpha1.EnvoyName: {
					Name:    mlopsv1alpha1.EnvoyName,
					Disable: true, // Explicitly disabled
				},
				mlopsv1alpha1.PipelineGatewayName: {
					Name:     mlopsv1alpha1.PipelineGatewayName,
					Replicas: int32Ptr(1), // Explicitly enabled with replicas
				},
			},
			expectedSvcNames: []string{mlopsv1alpha1.PipelineGatewayName},
			expectedSvcType: map[string]v1.ServiceType{
				mlopsv1alpha1.PipelineGatewayName: "",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			logger := logrtest.New(t)
			var client client2.Client
			scheme := runtime.NewScheme()
			err := mlopsv1alpha1.AddToScheme(scheme)
			g.Expect(err).To(BeNil())
			err = v1.AddToScheme(scheme)
			g.Expect(err).To(BeNil())
			err = auth.AddToScheme(scheme)
			g.Expect(err).To(BeNil())
			err = appsv1.AddToScheme(scheme)
			g.Expect(err).To(BeNil())
			g.Expect(err).To(BeNil())
			annotator := patch.NewAnnotator(constants.LastAppliedConfig)
			meta := metav1.ObjectMeta{
				Name:      "test",
				Namespace: "default",
			}
			client = testing2.NewFakeClient(scheme)
			sr := NewComponentServiceReconciler(
				common.ReconcilerConfig{Ctx: context.TODO(), Logger: logger, Client: client},
				meta,
				test.serviceConfig,
				test.overrides,
				annotator)
			g.Expect(err).To(BeNil())
			err = sr.Reconcile()

			g.Expect(err).To(BeNil())

			// Check that expected services are created
			for _, svcName := range test.expectedSvcNames {
				svc := &v1.Service{}
				err := client.Get(context.TODO(), types.NamespacedName{
					Name:      svcName,
					Namespace: meta.GetNamespace(),
				}, svc)
				g.Expect(err).To(BeNil(), "Expected service %s should be created", svcName)
				if test.expectedSvcType != nil {
					if svcType, ok := test.expectedSvcType[svcName]; ok {
						g.Expect(svc.Spec.Type).To(Equal(svcType))
					}
				}
			}

			// Check that disabled services are NOT created
			allPossibleServices := []string{SeldonMeshSVCName, mlopsv1alpha1.SchedulerName, mlopsv1alpha1.PipelineGatewayName}
			for _, svcName := range allPossibleServices {
				found := false
				for _, expectedSvc := range test.expectedSvcNames {
					if svcName == expectedSvc {
						found = true
						break
					}
				}
				if !found {
					// This service should NOT exist
					svc := &v1.Service{}
					err := client.Get(context.TODO(), types.NamespacedName{
						Name:      svcName,
						Namespace: meta.GetNamespace(),
					}, svc)
					g.Expect(err).ToNot(BeNil(), "Service %s should NOT be created when disabled", svcName)
				}
			}
		})
	}
}

func TestToServicesWithDisableFlag(t *testing.T) {
	g := NewGomegaWithT(t)

	meta := metav1.ObjectMeta{
		Name:      "test-runtime",
		Namespace: "test-namespace",
	}

	serviceConfig := mlopsv1alpha1.ServiceConfig{
		GrpcServicePrefix: "",
	}

	type test struct {
		name               string
		overrides          map[string]*mlopsv1alpha1.OverrideSpec
		expectedServices   []string
		unexpectedServices []string
	}

	tests := []test{
		{
			name:               "no overrides - all services created",
			overrides:          map[string]*mlopsv1alpha1.OverrideSpec{},
			expectedServices:   []string{mlopsv1alpha1.SchedulerName, SeldonMeshSVCName, mlopsv1alpha1.PipelineGatewayName},
			unexpectedServices: []string{},
		},
		{
			name: "scheduler disabled",
			overrides: map[string]*mlopsv1alpha1.OverrideSpec{
				mlopsv1alpha1.SchedulerName: {Disable: true},
			},
			expectedServices:   []string{SeldonMeshSVCName, mlopsv1alpha1.PipelineGatewayName},
			unexpectedServices: []string{mlopsv1alpha1.SchedulerName},
		},
		{
			name: "envoy disabled",
			overrides: map[string]*mlopsv1alpha1.OverrideSpec{
				mlopsv1alpha1.EnvoyName: {Disable: true},
			},
			expectedServices:   []string{mlopsv1alpha1.SchedulerName, mlopsv1alpha1.PipelineGatewayName},
			unexpectedServices: []string{SeldonMeshSVCName},
		},
		{
			name: "pipeline gateway disabled",
			overrides: map[string]*mlopsv1alpha1.OverrideSpec{
				mlopsv1alpha1.PipelineGatewayName: {Disable: true},
			},
			expectedServices:   []string{mlopsv1alpha1.SchedulerName, SeldonMeshSVCName},
			unexpectedServices: []string{mlopsv1alpha1.PipelineGatewayName},
		},
		{
			name: "all components disabled",
			overrides: map[string]*mlopsv1alpha1.OverrideSpec{
				mlopsv1alpha1.SchedulerName:       {Disable: true},
				mlopsv1alpha1.EnvoyName:           {Disable: true},
				mlopsv1alpha1.PipelineGatewayName: {Disable: true},
			},
			expectedServices:   []string{},
			unexpectedServices: []string{mlopsv1alpha1.SchedulerName, SeldonMeshSVCName, mlopsv1alpha1.PipelineGatewayName},
		},
		{
			name: "scheduler disabled explicitly false should create service",
			overrides: map[string]*mlopsv1alpha1.OverrideSpec{
				mlopsv1alpha1.SchedulerName: {Disable: false},
			},
			expectedServices:   []string{mlopsv1alpha1.SchedulerName, SeldonMeshSVCName, mlopsv1alpha1.PipelineGatewayName},
			unexpectedServices: []string{},
		},
		{
			name: "scheduler replicas 0 should not create service",
			overrides: map[string]*mlopsv1alpha1.OverrideSpec{
				mlopsv1alpha1.SchedulerName: {Replicas: int32Ptr(0)},
			},
			expectedServices:   []string{SeldonMeshSVCName, mlopsv1alpha1.PipelineGatewayName},
			unexpectedServices: []string{mlopsv1alpha1.SchedulerName},
		},
		{
			name: "envoy replicas 0 should not create service",
			overrides: map[string]*mlopsv1alpha1.OverrideSpec{
				mlopsv1alpha1.EnvoyName: {Replicas: int32Ptr(0)},
			},
			expectedServices:   []string{mlopsv1alpha1.SchedulerName, mlopsv1alpha1.PipelineGatewayName},
			unexpectedServices: []string{SeldonMeshSVCName},
		},
		{
			name: "pipeline gateway replicas 0 should not create service",
			overrides: map[string]*mlopsv1alpha1.OverrideSpec{
				mlopsv1alpha1.PipelineGatewayName: {Replicas: int32Ptr(0)},
			},
			expectedServices:   []string{mlopsv1alpha1.SchedulerName, SeldonMeshSVCName},
			unexpectedServices: []string{mlopsv1alpha1.PipelineGatewayName},
		},
		{
			name: "mixed disable true and replicas 0",
			overrides: map[string]*mlopsv1alpha1.OverrideSpec{
				mlopsv1alpha1.SchedulerName:       {Disable: true},
				mlopsv1alpha1.EnvoyName:           {Replicas: int32Ptr(0)},
				mlopsv1alpha1.PipelineGatewayName: {Replicas: int32Ptr(1)},
			},
			expectedServices:   []string{mlopsv1alpha1.PipelineGatewayName},
			unexpectedServices: []string{mlopsv1alpha1.SchedulerName, SeldonMeshSVCName},
		},
		{
			name: "replicas 1 should create service",
			overrides: map[string]*mlopsv1alpha1.OverrideSpec{
				mlopsv1alpha1.SchedulerName: {Replicas: int32Ptr(1)},
			},
			expectedServices:   []string{mlopsv1alpha1.SchedulerName, SeldonMeshSVCName, mlopsv1alpha1.PipelineGatewayName},
			unexpectedServices: []string{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			services := toServices(meta, serviceConfig, test.overrides)

			// Check that expected services are present
			serviceNames := make([]string, len(services))
			for i, svc := range services {
				serviceNames[i] = svc.Name
			}

			for _, expectedSvc := range test.expectedServices {
				found := false
				for _, actualSvc := range serviceNames {
					if expectedSvc == actualSvc {
						found = true
						break
					}
				}
				g.Expect(found).To(BeTrue(), "Expected service %s to be created but it was not found", expectedSvc)
			}

			// Check that unexpected services are NOT present
			for _, unexpectedSvc := range test.unexpectedServices {
				found := false
				for _, actualSvc := range serviceNames {
					if unexpectedSvc == actualSvc {
						found = true
						break
					}
				}
				g.Expect(found).To(BeFalse(), "Service %s should NOT be created when disabled but it was found", unexpectedSvc)
			}

			// Verify exact count
			expectedCount := len(test.expectedServices)
			actualCount := len(services)
			g.Expect(actualCount).To(Equal(expectedCount), "Expected %d services but got %d", expectedCount, actualCount)
		})
	}
}

// Helper function to create int32 pointer
func int32Ptr(i int32) *int32 {
	return &i
}
