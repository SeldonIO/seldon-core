package server

import (
	"context"
	"testing"

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
	testing2 "github.com/seldonio/seldon-core/operator/v2/pkg/utils/testing"
)

func TestRBACReconcile(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name                    string
		error                   bool
		expectedRoles           []string
		expectedServiceAccounts []string
		expectedRoleBindings    []string
	}
	tests := []test{
		{
			name:                    "normal services",
			expectedRoles:           []string{AgentRoleName, HodometerRoleName, SchedulerRoleName},
			expectedServiceAccounts: []string{ServerServiceAccountName, HodometerServiceAccountName, SchedulerServiceAccountName},
			expectedRoleBindings:    []string{AgentRoleBindingName, HodometerRoleBindingName, SchedulerRoleBindingName},
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
			sr := NewComponentRBACReconciler(
				common.ReconcilerConfig{Ctx: context.TODO(), Logger: logger, Client: client},
				meta)
			g.Expect(err).To(BeNil())
			err = sr.Reconcile()
			if test.error {
				g.Expect(err).ToNot(BeNil())
			} else {
				g.Expect(err).To(BeNil())
				for _, roleName := range test.expectedRoles {
					svc := &auth.Role{}
					err := client.Get(context.TODO(), types.NamespacedName{
						Name:      roleName,
						Namespace: meta.GetNamespace(),
					}, svc)
					g.Expect(err).To(BeNil())
				}
				for _, svcAccountName := range test.expectedServiceAccounts {
					svc := &v1.ServiceAccount{}
					err := client.Get(context.TODO(), types.NamespacedName{
						Name:      svcAccountName,
						Namespace: meta.GetNamespace(),
					}, svc)
					g.Expect(err).To(BeNil())
				}
				for _, roleBindingName := range test.expectedRoleBindings {
					svc := &auth.RoleBinding{}
					err := client.Get(context.TODO(), types.NamespacedName{
						Name:      roleBindingName,
						Namespace: meta.GetNamespace(),
					}, svc)
					g.Expect(err).To(BeNil())
				}
			}
		})
	}
}
