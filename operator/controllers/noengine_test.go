package controllers

import (
	"context"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	machinelearningv1alpha2 "github.com/seldonio/seldon-core/operator/api/v1alpha2"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"time"
)

var _ = Describe("Create a Seldon Deployment without engine", func() {
	const timeout = time.Second * 30
	const interval = time.Second * 1
	By("Creating a resource")
	It("should create a resource with defaults", func() {
		Expect(k8sClient).NotTo(BeNil())
		var modelType = machinelearningv1alpha2.MODEL
		key := types.NamespacedName{
			Name:      "dep2",
			Namespace: "default",
		}
		instance := &machinelearningv1alpha2.SeldonDeployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      key.Name,
				Namespace: key.Namespace,
			},
			Spec: machinelearningv1alpha2.SeldonDeploymentSpec{
				Name: "mydep2",
				Predictors: []machinelearningv1alpha2.PredictorSpec{
					{
						Annotations: map[string]string{
							"seldon.io/no-engine": "true",
						},
						Name: "p1",
						ComponentSpecs: []*machinelearningv1alpha2.SeldonPodSpec{
							{
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
						Graph: &machinelearningv1alpha2.PredictiveUnit{
							Name: "classifier",
							Type: &modelType,
						},
					},
				},
			},
		}

		// Run Defaulter
		instance.Default()

		Expect(k8sClient.Create(context.Background(), instance)).Should(Succeed())
		//time.Sleep(time.Second * 5)

		fetched := &machinelearningv1alpha2.SeldonDeployment{}
		Eventually(func() error {
			err := k8sClient.Get(context.Background(), key, fetched)
			return err
		}, timeout, interval).Should(BeNil())
		Expect(fetched.Spec.Name).Should(Equal("mydep2"))

		depKey := types.NamespacedName{
			Name:      machinelearningv1alpha2.GetDeploymentName(instance, instance.Spec.Predictors[0], instance.Spec.Predictors[0].ComponentSpecs[0]),
			Namespace: "default",
		}
		depFetched := &appsv1.Deployment{}
		Eventually(func() error {
			err := k8sClient.Get(context.Background(), depKey, depFetched)
			return err
		}, timeout, interval).Should(BeNil())
		Expect(len(depFetched.Spec.Template.Spec.Containers)).Should(Equal(1))

		Expect(k8sClient.Delete(context.Background(), instance)).Should(Succeed())
	})

})
