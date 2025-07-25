// /*
// Copyright (c) 2024 Seldon Technologies Ltd.

// Use of this software is governed by
// (1) the license included in the LICENSE file or
// (2) if the license included in the LICENSE file is the Business Source License 1.1,
// the Change License after the Change Date as each is defined in accordance with the LICENSE file.
// */

package server

import (
	"context"
	"testing"

	logrtest "github.com/go-logr/logr/testr"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	client2 "sigs.k8s.io/controller-runtime/pkg/client"

	mlopsv1alpha1 "github.com/seldonio/seldon-core/operator/v2/apis/mlops/v1alpha1"
	"github.com/seldonio/seldon-core/operator/v2/controllers/reconcilers/common"
	"github.com/seldonio/seldon-core/operator/v2/pkg/constants"
	testing2 "github.com/seldonio/seldon-core/operator/v2/pkg/utils/testing"
)

func TestServerReconcileWithDeployment(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name                   string
		serverConfig           *mlopsv1alpha1.ServerConfig
		server                 *mlopsv1alpha1.Server
		error                  bool
		expectedDeploymentName string
	}
	mlserverConfigName := "mlserver-config"
	tests := []test{
		{
			name: "MLServer",
			serverConfig: &mlopsv1alpha1.ServerConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      mlserverConfigName,
					Namespace: constants.SeldonNamespace,
				},
				Spec: mlopsv1alpha1.ServerConfigSpec{
					PodSpec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Name:  "mlserver",
								Image: "seldonio/mlserver:0.5",
							},
						},
					},
				},
			},
			server: &mlopsv1alpha1.Server{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "myserver",
					Namespace: constants.SeldonNamespace,
				},
				Spec: mlopsv1alpha1.ServerSpec{
					ServerConfig: mlserverConfigName,
				},
			},
			expectedDeploymentName: "myserver",
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
			err = appsv1.AddToScheme(scheme)
			g.Expect(err).To(BeNil())
			if test.serverConfig != nil {
				client = testing2.NewFakeClient(scheme, test.serverConfig)
			} else {
				client = testing2.NewFakeClient(scheme)
			}
			g.Expect(err).To(BeNil())
			sr, err := NewServerReconcilerWithDeployment(test.server, common.ReconcilerConfig{Ctx: context.TODO(), Logger: logger, Client: client})
			g.Expect(err).To(BeNil())
			err = sr.Reconcile()
			if test.error {
				g.Expect(err).ToNot(BeNil())
			} else {
				g.Expect(err).To(BeNil())
				deployment := &appsv1.Deployment{}
				err := client.Get(context.TODO(), types.NamespacedName{
					Name:      test.expectedDeploymentName,
					Namespace: test.server.GetNamespace(),
				}, deployment)
				g.Expect(err).To(BeNil())
			}
		})
	}
}

func TestNewServerReconcilerWithDeployment(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name         string
		serverConfig *mlopsv1alpha1.ServerConfig
		server       *mlopsv1alpha1.Server
		error        bool
	}
	tests := []test{
		{
			name: "MLServer",
			serverConfig: &mlopsv1alpha1.ServerConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "mlserver",
					Namespace: constants.SeldonNamespace,
				},
				Spec: mlopsv1alpha1.ServerConfigSpec{
					PodSpec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Name:  "mlserver",
								Image: "seldonio/mlserver:0.5",
							},
						},
					},
				},
			},
			server: &mlopsv1alpha1.Server{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "myserver",
					Namespace: constants.SeldonNamespace,
				},
				Spec: mlopsv1alpha1.ServerSpec{
					ServerConfig: "mlserver",
				},
			},
		},
		{
			name: "MissingServer",
			server: &mlopsv1alpha1.Server{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "myserver",
					Namespace: constants.SeldonNamespace,
				},
				Spec: mlopsv1alpha1.ServerSpec{
					ServerConfig: "mlserver",
				},
			},
			error: true,
		},
		{
			name: "CustomPodSpec",
			serverConfig: &mlopsv1alpha1.ServerConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "mlserver",
					Namespace: constants.SeldonNamespace,
				},
				Spec: mlopsv1alpha1.ServerConfigSpec{
					PodSpec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Name:  "mlserver",
								Image: "seldonio/mlserver:0.5",
							},
						},
					},
				},
			},
			server: &mlopsv1alpha1.Server{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "myserver",
					Namespace: constants.SeldonNamespace,
				},
				Spec: mlopsv1alpha1.ServerSpec{
					ServerConfig: "mlserver",
					PodSpec: &mlopsv1alpha1.PodSpec{
						NodeName: "node",
					},
				},
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
			if test.serverConfig != nil {
				client = testing2.NewFakeClient(scheme, test.serverConfig)
			} else {
				client = testing2.NewFakeClient(scheme)
			}
			g.Expect(err).To(BeNil())
			_, err = NewServerReconcilerWithDeployment(test.server, common.ReconcilerConfig{Ctx: context.TODO(), Logger: logger, Client: client})
			if test.error {
				g.Expect(err).ToNot(BeNil())
			} else {
				g.Expect(err).To(BeNil())
			}
		})
	}
}
