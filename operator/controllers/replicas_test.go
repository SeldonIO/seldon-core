package controllers

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"
	machinelearningv1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

func createSeldonDeploymentWithReplicas(name string, namespace string, specReplicas *int32, predictorReplicas *int32, componentSpecReplicas *int32, svcOrchReplicas *int32) *machinelearningv1.SeldonDeployment {
	envExecutorImage = "seldonio/executor:0.1"
	modelType := machinelearningv1.MODEL
	instance := &machinelearningv1.SeldonDeployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: machinelearningv1.SeldonDeploymentSpec{
			Replicas: specReplicas,
			Predictors: []machinelearningv1.PredictorSpec{
				{
					Name:     "p1",
					Replicas: predictorReplicas,
					ComponentSpecs: []*machinelearningv1.SeldonPodSpec{
						{
							Metadata: machinelearningv1.ObjectMeta{},
							Replicas: componentSpecReplicas,
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
				},
			},
		},
	}

	if svcOrchReplicas != nil {
		instance.Spec.Annotations = map[string]string{machinelearningv1.ANNOTATION_SEPARATE_ENGINE: "true"}
		instance.Spec.Predictors[0].SvcOrchSpec = machinelearningv1.SvcOrchSpec{
			Replicas: svcOrchReplicas,
		}
	}
	return instance
}

func TestDeploymentReplicas(t *testing.T) {
	g := NewGomegaWithT(t)
	svcOrchReplicas := int32(2)
	specReplicas := int32(3)
	predictorReplicas := int32(4)
	componentSpecReplicas := int32(5)
	name := "dep"
	namespace := "default"

	logger := ctrl.Log.WithName("controllers").WithName("SeldonDeployment")
	reconciler := &SeldonDeploymentReconciler{
		Log: logger,
	}

	// Just Predictor Replicas
	instance := createSeldonDeploymentWithReplicas(name, namespace, nil, &predictorReplicas, nil, nil)
	instance.Spec.DefaultSeldonDeployment(name, namespace)
	c, err := reconciler.createComponents(context.TODO(), instance, nil, logger)
	g.Expect(err).To(BeNil())
	g.Expect(len(c.deployments)).To(Equal(1))
	g.Expect(*c.deployments[0].Spec.Replicas).To(Equal(predictorReplicas))

	// Just Predictor Replicas and default replicas
	instance = createSeldonDeploymentWithReplicas(name, namespace, &specReplicas, &predictorReplicas, nil, nil)
	instance.Spec.DefaultSeldonDeployment(name, namespace)
	c, err = reconciler.createComponents(context.TODO(), instance, nil, logger)
	g.Expect(err).To(BeNil())
	g.Expect(len(c.deployments)).To(Equal(1))
	g.Expect(*c.deployments[0].Spec.Replicas).To(Equal(predictorReplicas))

	// ComponentSpec replica override
	instance = createSeldonDeploymentWithReplicas(name, namespace, &specReplicas, &predictorReplicas, &componentSpecReplicas, nil)
	instance.Spec.DefaultSeldonDeployment(name, namespace)
	c, err = reconciler.createComponents(context.TODO(), instance, nil, logger)
	g.Expect(err).To(BeNil())
	g.Expect(len(c.deployments)).To(Equal(1))
	g.Expect(*c.deployments[0].Spec.Replicas).To(Equal(componentSpecReplicas))

	// Just specReplicas
	instance = createSeldonDeploymentWithReplicas(name, namespace, &specReplicas, nil, nil, nil)
	instance.Spec.DefaultSeldonDeployment(name, namespace)
	c, err = reconciler.createComponents(context.TODO(), instance, nil, logger)
	g.Expect(err).To(BeNil())
	g.Expect(len(c.deployments)).To(Equal(1))
	g.Expect(*c.deployments[0].Spec.Replicas).To(Equal(specReplicas))

	// All nil
	instance = createSeldonDeploymentWithReplicas(name, namespace, nil, nil, nil, nil)
	instance.Spec.DefaultSeldonDeployment(name, namespace)
	c, err = reconciler.createComponents(context.TODO(), instance, nil, logger)
	g.Expect(err).To(BeNil())
	g.Expect(len(c.deployments)).To(Equal(1))
	g.Expect(c.deployments[0].Spec.Replicas).To(BeNil())

	// SvcOrchReplicas
	instance = createSeldonDeploymentWithReplicas(name, namespace, nil, nil, nil, &svcOrchReplicas)
	instance.Spec.DefaultSeldonDeployment(name, namespace)
	c, err = reconciler.createComponents(context.TODO(), instance, nil, logger)
	g.Expect(err).To(BeNil())
	g.Expect(len(c.deployments)).To(Equal(2))
	g.Expect(*c.deployments[0].Spec.Replicas).To(Equal(svcOrchReplicas))
	g.Expect(c.deployments[1].Spec.Replicas).To(BeNil())
}
