package controllers

import (
	"context"
	"fmt"
	"testing"

	. "github.com/onsi/gomega"
	machinelearningv1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

func createSeldonDeploymentWithPredictorsWithAnnotation(name string, namespace string, addAnnotation bool, engineSeparatePod bool) *machinelearningv1.SeldonDeployment {
	envExecutorImage = "seldonio/executor:0.1"
	modelType := machinelearningv1.MODEL
	instance := &machinelearningv1.SeldonDeployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: machinelearningv1.SeldonDeploymentSpec{
			Predictors: []machinelearningv1.PredictorSpec{
				{
					Name: "p1",
					ComponentSpecs: []*machinelearningv1.SeldonPodSpec{
						{
							Metadata: machinelearningv1.ObjectMeta{},
							Spec: v1.PodSpec{
								Containers: []v1.Container{
									{
										Image: "seldonio/mock_classifier:1.0",
										Name:  "classifier",
									},
								},
							},
						},
					},
					Graph: machinelearningv1.PredictiveUnit{
						Name: "classifier",
						Type: &modelType,
					},
					Traffic: 70,
				},
				{
					Name: "p2",
					ComponentSpecs: []*machinelearningv1.SeldonPodSpec{
						{
							Metadata: machinelearningv1.ObjectMeta{},
							Spec: v1.PodSpec{
								Containers: []v1.Container{
									{
										Image: "seldonio/mock_classifier:1.0",
										Name:  "classifier",
									},
								},
							},
						},
					},
					Graph: machinelearningv1.PredictiveUnit{
						Name: "classifier",
						Type: &modelType,
					},
					Traffic: 30,
				},
			},
		},
	}
	if addAnnotation {
		instance.Spec.Annotations = map[string]string{"seldon.io/engine-separate-pod": fmt.Sprintf("%v", engineSeparatePod)}
	}
	return instance
}

func TestEngineSeparatePodAnnotation(t *testing.T) {
	g := NewGomegaWithT(t)

	name := "test"
	namespace := "default"

	logger := ctrl.Log.WithName("controllers").WithName("SeldonDeployment")
	reconciler := &SeldonDeploymentReconciler{
		Log: logger,
	}

	// With separte engine annotation true
	instance := createSeldonDeploymentWithPredictorsWithAnnotation(name, namespace, true, true)
	instance.Spec.DefaultSeldonDeployment(name, namespace)
	_, err := reconciler.createComponents(context.TODO(), instance, nil, logger)
	g.Expect(err).To(BeNil())

	// With separte engine annotation false
	instance = createSeldonDeploymentWithPredictorsWithAnnotation(name, namespace, true, false)
	instance.Spec.DefaultSeldonDeployment(name, namespace)
	_, err = reconciler.createComponents(context.TODO(), instance, nil, logger)
	g.Expect(err).To(BeNil())

	// No annotations
	instance = createSeldonDeploymentWithPredictorsWithAnnotation(name, namespace, false, false)
	instance.Spec.DefaultSeldonDeployment(name, namespace)
	_, err = reconciler.createComponents(context.TODO(), instance, nil, logger)
	g.Expect(err).To(BeNil())

}
