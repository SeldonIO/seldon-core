package server

import (
	"context"
	"testing"

	"github.com/seldonio/seldon-core/operator/v2/pkg/constants"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	logrtest "github.com/go-logr/logr/testr"
	mlopsv1alpha1 "github.com/seldonio/seldon-core/operator/v2/apis/mlops/v1alpha1"
	"github.com/seldonio/seldon-core/operator/v2/controllers/reconcilers/common"
	testing2 "github.com/seldonio/seldon-core/operator/v2/pkg/utils/testing"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	auth "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	client2 "sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/onsi/gomega"
)

func TestRuntimeReconcile(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name                    string
		seldonConfig            *mlopsv1alpha1.SeldonConfig
		runtime                 *mlopsv1alpha1.SeldonRuntime
		error                   bool
		expectedSvcNames        []string
		expectedDeployments     []string
		expectedStatefulSetName string
	}
	configName := "config"
	tests := []test{
		{
			name: "scheduler",
			seldonConfig: &mlopsv1alpha1.SeldonConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      configName,
					Namespace: constants.SeldonNamespace,
				},
				Spec: mlopsv1alpha1.SeldonConfigSpec{
					Components: []*mlopsv1alpha1.ComponentDefn{
						{
							Name: mlopsv1alpha1.SchedulerName,
							PodSpec: &v1.PodSpec{
								Containers: []v1.Container{
									{
										Name:  "scheduler",
										Image: "seldonio/scheduler:latest",
									},
								},
							},
							VolumeClaimTemplates: []mlopsv1alpha1.PersistentVolumeClaim{
								{
									Name: "vc",
									Spec: v1.PersistentVolumeClaimSpec{},
								},
							},
						},
					},
				},
			},
			runtime: &mlopsv1alpha1.SeldonRuntime{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "runtime",
					Namespace: constants.SeldonNamespace,
				},
				Spec: mlopsv1alpha1.SeldonRuntimeSpec{
					SeldonConfig: configName,
				},
			},
			expectedSvcNames:        []string{SeldonMeshSVCName, mlopsv1alpha1.SchedulerName, mlopsv1alpha1.PipelineGatewayName},
			expectedStatefulSetName: mlopsv1alpha1.SchedulerName,
		},
		{
			name: "deployments",
			seldonConfig: &mlopsv1alpha1.SeldonConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      configName,
					Namespace: constants.SeldonNamespace,
				},
				Spec: mlopsv1alpha1.SeldonConfigSpec{
					Components: []*mlopsv1alpha1.ComponentDefn{
						{
							Name: mlopsv1alpha1.PipelineGatewayName,
							PodSpec: &v1.PodSpec{
								Containers: []v1.Container{
									{
										Name:  "pipeline-gateway",
										Image: "seldonio/pipeline:latest",
									},
								},
							},
						},
						{
							Name: mlopsv1alpha1.ModelGatewayName,
							PodSpec: &v1.PodSpec{
								Containers: []v1.Container{
									{
										Name:  "model-gateway",
										Image: "seldonio/modelgateway:latest",
									},
								},
							},
						},
					},
				},
			},
			runtime: &mlopsv1alpha1.SeldonRuntime{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "runtime",
					Namespace: constants.SeldonNamespace,
				},
				Spec: mlopsv1alpha1.SeldonRuntimeSpec{
					SeldonConfig: configName,
				},
			},
			expectedSvcNames:    []string{SeldonMeshSVCName, mlopsv1alpha1.SchedulerName, mlopsv1alpha1.PipelineGatewayName},
			expectedDeployments: []string{mlopsv1alpha1.PipelineGatewayName, mlopsv1alpha1.ModelGatewayName},
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
			if test.seldonConfig != nil {
				client = testing2.NewFakeClient(scheme, test.seldonConfig)
			} else {
				client = testing2.NewFakeClient(scheme)
			}
			g.Expect(err).To(BeNil())
			sr, err := NewSeldonRuntimeReconciler(test.runtime, common.ReconcilerConfig{Ctx: context.TODO(), Logger: logger, Client: client})
			g.Expect(err).To(BeNil())
			err = sr.Reconcile()
			if test.error {
				g.Expect(err).ToNot(BeNil())
			} else {
				g.Expect(err).To(BeNil())
				for _, svcName := range test.expectedSvcNames {
					svc := &v1.Service{}
					err := client.Get(context.TODO(), types.NamespacedName{
						Name:      svcName,
						Namespace: test.runtime.GetNamespace(),
					}, svc)
					g.Expect(err).To(BeNil())
				}
				if test.expectedStatefulSetName != "" {
					ss := &appsv1.StatefulSet{}
					err := client.Get(context.TODO(), types.NamespacedName{
						Name:      test.expectedStatefulSetName,
						Namespace: test.runtime.GetNamespace(),
					}, ss)
					g.Expect(err).To(BeNil())
				}
				for _, depName := range test.expectedDeployments {
					dep := &appsv1.Deployment{}
					err := client.Get(context.TODO(), types.NamespacedName{
						Name:      depName,
						Namespace: test.runtime.GetNamespace(),
					}, dep)
					g.Expect(err).To(BeNil())
				}
			}
		})
	}
}
