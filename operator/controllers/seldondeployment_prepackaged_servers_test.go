package controllers

import (
	"context"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	machinelearningv1alpha2 "github.com/seldonio/seldon-core/operator/api/v1alpha2"
	"github.com/seldonio/seldon-core/operator/utils"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"time"
)

var _ = Describe("Create a prepacked sklearn server", func() {
	const timeout = time.Second * 30
	const interval = time.Second * 1
	By("Creating a resource")
	It("should create a resource with defaults", func() {
		Expect(k8sClient).NotTo(BeNil())
		var modelType = machinelearningv1alpha2.MODEL
		var impl = machinelearningv1alpha2.PredictiveUnitImplementation("SKLEARN_SERVER")
		key := types.NamespacedName{
			Name:      "prepack",
			Namespace: "default",
		}
		instance := &machinelearningv1alpha2.SeldonDeployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      key.Name,
				Namespace: key.Namespace,
			},
			Spec: machinelearningv1alpha2.SeldonDeploymentSpec{
				Name: "pp",
				Predictors: []machinelearningv1alpha2.PredictorSpec{
					{
						Name: "p1",
						Graph: &machinelearningv1alpha2.PredictiveUnit{
							Name:           "classifier",
							Type:           &modelType,
							Implementation: &impl,
							Endpoint:       &machinelearningv1alpha2.Endpoint{Type: machinelearningv1alpha2.REST},
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
		Expect(fetched.Spec.Name).Should(Equal("pp"))

		sPodSpec := utils.GetSeldonPodSpecForPredictiveUnit(&instance.Spec.Predictors[0], instance.Spec.Predictors[0].Graph.Name)
		depName := machinelearningv1alpha2.GetDeploymentName(instance, instance.Spec.Predictors[0], sPodSpec)
		depKey := types.NamespacedName{
			Name:      depName,
			Namespace: "default",
		}
		depFetched := &appsv1.Deployment{}
		Eventually(func() error {
			err := k8sClient.Get(context.Background(), depKey, depFetched)
			return err
		}, timeout, interval).Should(BeNil())
		Expect(len(depFetched.Spec.Template.Spec.Containers)).Should(Equal(2))

		Expect(k8sClient.Delete(context.Background(), instance)).Should(Succeed())
	})

})
