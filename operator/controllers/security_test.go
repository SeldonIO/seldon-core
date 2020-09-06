package controllers

import (
	"context"
	. "github.com/onsi/gomega"
	machinelearningv1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"testing"
)

func TestSecurityContextCreateComponents(t *testing.T) {
	g := NewGomegaWithT(t)

	name := "test"
	namespace := "default"

	instance := createSeldonDeploymentWithReplicas(name, namespace, nil, nil, nil, nil)
	instance.Spec.DefaultSeldonDeployment(name, namespace)
	logger := ctrl.Log.WithName("controllers").WithName("SeldonDeployment")
	reconciler := &SeldonDeploymentReconciler{
		Log: logger,
	}

	user := int64(2)
	podSecurityContext := &corev1.PodSecurityContext{
		RunAsUser: &user,
	}
	c, err := reconciler.createComponents(context.TODO(), instance, podSecurityContext, reconciler.Log)
	g.Expect(err).To(BeNil())
	g.Expect(len(c.deployments)).To(Equal(1))
	g.Expect(*c.deployments[0].Spec.Template.Spec.SecurityContext.RunAsUser).To(Equal(user))
}

func TestNoSecurityContextCreateComponents(t *testing.T) {
	g := NewGomegaWithT(t)

	name := "test"
	namespace := "default"

	instance := createSeldonDeploymentWithReplicas(name, namespace, nil, nil, nil, nil)
	instance.Spec.DefaultSeldonDeployment(name, namespace)
	logger := ctrl.Log.WithName("controllers").WithName("SeldonDeployment")
	reconciler := &SeldonDeploymentReconciler{
		Log: logger,
	}

	var podSecurityContext *corev1.PodSecurityContext
	c, err := reconciler.createComponents(context.TODO(), instance, podSecurityContext, reconciler.Log)
	g.Expect(err).To(BeNil())
	g.Expect(len(c.deployments)).To(Equal(1))
	g.Expect(c.deployments[0].Spec.Template.Spec.SecurityContext).To(BeNil())
}

func TestGetSecurityContextExecutorUser(t *testing.T) {
	g := NewGomegaWithT(t)
	mlDep := &machinelearningv1.SeldonDeployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test",
		},
		Spec: machinelearningv1.SeldonDeploymentSpec{Annotations: map[string]string{machinelearningv1.ANNOTATION_EXECUTOR: "true"}},
	}
	envDefaultUser = "2"

	sc, err := createSecurityContext(mlDep)
	g.Expect(err).To(BeNil())
	g.Expect(sc).ToNot(BeNil())
	g.Expect(*sc.RunAsUser).To(Equal(int64(2)))
}
