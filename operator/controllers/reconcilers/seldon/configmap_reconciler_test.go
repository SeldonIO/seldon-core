package server

import (
	"context"
	logrtest "github.com/go-logr/logr/testr"
	. "github.com/onsi/gomega"
	mlopsv1alpha1 "github.com/seldonio/seldon-core/operator/v2/apis/mlops/v1alpha1"
	"github.com/seldonio/seldon-core/operator/v2/controllers/reconcilers/common"
	testing2 "github.com/seldonio/seldon-core/operator/v2/pkg/utils/testing"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	auth "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	client2 "sigs.k8s.io/controller-runtime/pkg/client"
	"testing"
)

func TestConfigMapReconcile(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name               string
		config             *mlopsv1alpha1.SeldonConfiguration
		error              bool
		expectedConfigMaps []string
	}
	tests := []test{
		{
			name:               "normal configmaps",
			config:             &mlopsv1alpha1.SeldonConfiguration{},
			expectedConfigMaps: []string{agentConfigMapName, kafkaConfigMapName, traceConfigMapName},
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
			meta := metav1.ObjectMeta{
				Name:      "test",
				Namespace: "default",
			}
			client = testing2.NewFakeClient(scheme)
			sr, err := NewConfigMapReconciler(
				common.ReconcilerConfig{Ctx: context.TODO(), Logger: logger, Client: client},
				test.config,
				meta)
			g.Expect(err).To(BeNil())
			err = sr.Reconcile()
			if test.error {
				g.Expect(err).ToNot(BeNil())
			} else {
				g.Expect(err).To(BeNil())
				for _, configMapName := range test.expectedConfigMaps {
					svc := &v1.ConfigMap{}
					err := client.Get(context.TODO(), types.NamespacedName{
						Name:      configMapName,
						Namespace: meta.GetNamespace(),
					}, svc)
					g.Expect(err).To(BeNil())
				}
			}
		})
	}
}
