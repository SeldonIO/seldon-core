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
	"strings"
	"testing"

	logrtest "github.com/go-logr/logr/testr"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	client2 "sigs.k8s.io/controller-runtime/pkg/client"

	mlopsv1alpha1 "github.com/seldonio/seldon-core/operator/v2/apis/mlops/v1alpha1"
	"github.com/seldonio/seldon-core/operator/v2/controllers/reconcilers/common"
	"github.com/seldonio/seldon-core/operator/v2/pkg/constants"
	testing2 "github.com/seldonio/seldon-core/operator/v2/pkg/utils/testing"
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

func TestServiceReconcile(t *testing.T) {
	g := NewGomegaWithT(t)
	logger := logrtest.New(t)
	type test struct {
		name            string
		reconcilerTime1 *ServerServiceReconciler
		reconcilerTime2 *ServerServiceReconciler
	}

	getIntPtr := func(i int32) *int32 {
		return &i
	}
	tests := []test{
		{
			name: "Create",
			reconcilerTime2: NewServerServiceReconciler(common.ReconcilerConfig{Ctx: context.Background(), Logger: logger},
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
			reconcilerTime1: NewServerServiceReconciler(common.ReconcilerConfig{Ctx: context.Background(), Logger: logger},
				metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "default",
				},
				&mlopsv1alpha1.ScalingSpec{
					Replicas: getIntPtr(1),
				}),
			reconcilerTime2: NewServerServiceReconciler(common.ReconcilerConfig{Ctx: context.Background(), Logger: logger},
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
			reconcilerTime1: NewServerServiceReconciler(common.ReconcilerConfig{Ctx: context.Background(), Logger: logger},
				metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "default",
				},
				&mlopsv1alpha1.ScalingSpec{
					Replicas: getIntPtr(3),
				}),
			reconcilerTime2: NewServerServiceReconciler(common.ReconcilerConfig{Ctx: context.Background(), Logger: logger},
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
