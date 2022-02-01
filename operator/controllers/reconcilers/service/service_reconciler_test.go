package service

import (
	"context"
	"strings"
	"testing"

	"github.com/seldonio/seldon-core/operatorv2/pkg/constants"

	logrtest "github.com/go-logr/logr/testing"
	"github.com/seldonio/seldon-core/operatorv2/controllers/reconcilers/common"
	testing2 "github.com/seldonio/seldon-core/operatorv2/pkg/utils/testing"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	client2 "sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/onsi/gomega"
	mlopsv1alpha1 "github.com/seldonio/seldon-core/operatorv2/apis/mlops/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestToServices(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name            string
		meta            metav1.ObjectMeta
		replicas        int
		numExpectedSvcs int
		expectedLabels  []map[string]string
	}

	tests := []test{
		{
			name: "Create",
			meta: metav1.ObjectMeta{
				Name:      "foo",
				Namespace: "default",
			},
			replicas:        2,
			numExpectedSvcs: 2,
			expectedLabels: []map[string]string{
				{
					constants.ServerReplicaLabelKey:     "foo",
					constants.ServerReplicaNameLabelKey: "foo-0",
				},
				{
					constants.ServerReplicaLabelKey:     "foo",
					constants.ServerReplicaNameLabelKey: "foo-1",
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			svcs := toServices(test.meta, test.replicas)
			g.Expect(len(svcs)).To(Equal(test.numExpectedSvcs))
			for idx, svc := range svcs {
				g.Expect(svc.Spec.ClusterIP).To(Equal("None"))
				g.Expect(strings.HasPrefix(svc.GetName(), test.meta.GetName())).To(BeTrue())
				for k, v := range test.expectedLabels[idx] {
					g.Expect(svc.Labels[k]).To(Equal(v))
				}
			}
		})
	}
}

func TestReconcile(t *testing.T) {
	g := NewGomegaWithT(t)
	logger := logrtest.TestLogger{T: t}
	type test struct {
		name            string
		reconcilerTime1 *ServiceReconciler
		reconcilerTime2 *ServiceReconciler
	}

	getIntPtr := func(i int32) *int32 {
		return &i
	}
	tests := []test{
		{
			name: "Create",
			reconcilerTime2: NewServiceReconciler(common.ReconcilerConfig{Ctx: context.Background(), Logger: logger},
				metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "default",
				},
				&mlopsv1alpha1.ScalingSpec{
					Replicas: getIntPtr(2),
				}),
		},
		{
			name: "Existing svcs from previous reconcile",
			reconcilerTime1: NewServiceReconciler(common.ReconcilerConfig{Ctx: context.Background(), Logger: logger},
				metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "default",
				},
				&mlopsv1alpha1.ScalingSpec{
					Replicas: getIntPtr(1),
				}),
			reconcilerTime2: NewServiceReconciler(common.ReconcilerConfig{Ctx: context.Background(), Logger: logger},
				metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "default",
				},
				&mlopsv1alpha1.ScalingSpec{
					Replicas: getIntPtr(2),
				}),
		},
		{
			name: "decrease in number of replicas",
			reconcilerTime1: NewServiceReconciler(common.ReconcilerConfig{Ctx: context.Background(), Logger: logger},
				metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "default",
				},
				&mlopsv1alpha1.ScalingSpec{
					Replicas: getIntPtr(3),
				}),
			reconcilerTime2: NewServiceReconciler(common.ReconcilerConfig{Ctx: context.Background(), Logger: logger},
				metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "default",
				},
				&mlopsv1alpha1.ScalingSpec{
					Replicas: getIntPtr(1),
				}),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var client client2.Client
			scheme := runtime.NewScheme()
			err := appsv1.AddToScheme(scheme)
			g.Expect(err).To(BeNil())
			err = v1.AddToScheme(scheme)
			g.Expect(err).To(BeNil())
			if test.reconcilerTime1 != nil {
				var objs []client2.Object
				for _, svc := range test.reconcilerTime1.Services {
					objs = append(objs, svc)
				}
				client = testing2.NewFakeClient(scheme, objs...)
			} else {
				client = testing2.NewFakeClient(scheme)
			}

			test.reconcilerTime2.ReconcilerConfig.Client = client
			err = test.reconcilerTime2.Reconcile()
			g.Expect(err).To(BeNil())
			for _, svc := range test.reconcilerTime2.Services {
				found := &v1.Service{}
				err := client.Get(context.TODO(), types.NamespacedName{
					Name:      svc.GetName(),
					Namespace: svc.GetNamespace(),
				}, found)
				g.Expect(err).To(BeNil())
				g.Expect(found.Spec).To(Equal(svc.Spec))
			}
			if test.reconcilerTime1 != nil && len(test.reconcilerTime2.Services) < len(test.reconcilerTime1.Services) {
				for i := len(test.reconcilerTime2.Services); i < len(test.reconcilerTime1.Services); i++ {
					found := &v1.Service{}
					svc := test.reconcilerTime1.Services[i]
					err := client.Get(context.TODO(), types.NamespacedName{
						Name:      svc.GetName(),
						Namespace: svc.GetNamespace(),
					}, found)
					g.Expect(err).ToNot(BeNil())
					g.Expect(errors.IsNotFound(err)).To(BeTrue())
				}
			}
		})
	}
}
