package controllers

import (
	"context"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	machinelearningv1 "github.com/seldonio/seldon-core/operator/apis/machinelearning/v1"
	"github.com/seldonio/seldon-core/operator/constants"
	"github.com/seldonio/seldon-core/operator/utils"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"strconv"
	"strings"
	"time"
)

var _ = Describe("Create a prepacked sklearn server", func() {
	const timeout = time.Second * 30
	const interval = time.Second * 1
	const name = "pp1"
	const sdepName = "prepack1"
	By("Creating a resource")
	It("should create a resource with defaults", func() {
		Expect(k8sClient).NotTo(BeNil())
		var modelType = machinelearningv1.MODEL
		var impl = machinelearningv1.PredictiveUnitImplementation(constants.PrePackedServerSklearn)
		key := types.NamespacedName{
			Name:      sdepName,
			Namespace: "default",
		}
		instance := &machinelearningv1.SeldonDeployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      key.Name,
				Namespace: key.Namespace,
			},
			Spec: machinelearningv1.SeldonDeploymentSpec{
				Name: name,
				Predictors: []machinelearningv1.PredictorSpec{
					{
						Name: "p1",
						Graph: &machinelearningv1.PredictiveUnit{
							Name:           "classifier",
							Type:           &modelType,
							Implementation: &impl,
							Endpoint:       &machinelearningv1.Endpoint{Type: machinelearningv1.REST},
						},
					},
				},
			},
		}

		configMapName := types.NamespacedName{Name: "seldon-config",
			Namespace: "seldon-system"}

		configResult := &corev1.ConfigMap{}
		const timeout = time.Second * 30
		Eventually(func() error { return k8sClient.Get(context.TODO(), configMapName, configResult) }, timeout).
			Should(Succeed())

		// Run Defaulter
		instance.Default()

		Expect(k8sClient.Create(context.Background(), instance)).Should(Succeed())
		//time.Sleep(time.Second * 5)

		fetched := &machinelearningv1.SeldonDeployment{}
		Eventually(func() error {
			err := k8sClient.Get(context.Background(), key, fetched)
			return err
		}, timeout, interval).Should(BeNil())
		Expect(fetched.Spec.Name).Should(Equal(name))

		sPodSpec := utils.GetSeldonPodSpecForPredictiveUnit(&instance.Spec.Predictors[0], instance.Spec.Predictors[0].Graph.Name)
		depName := machinelearningv1.GetDeploymentName(instance, instance.Spec.Predictors[0], sPodSpec)
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

var _ = Describe("Create a prepacked tfserving server for Seldon protocol and REST", func() {
	const timeout = time.Second * 30
	const interval = time.Second * 1
	const name = "pp2"
	const sdepName = "prepack2"
	By("Creating a resource")
	It("should create a resource with defaults", func() {
		Expect(k8sClient).NotTo(BeNil())
		var modelType = machinelearningv1.MODEL
		var impl = machinelearningv1.PredictiveUnitImplementation(constants.PrePackedServerTensorflow)
		key := types.NamespacedName{
			Name:      sdepName,
			Namespace: "default",
		}
		instance := &machinelearningv1.SeldonDeployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      key.Name,
				Namespace: key.Namespace,
			},
			Spec: machinelearningv1.SeldonDeploymentSpec{
				Name: name,
				Predictors: []machinelearningv1.PredictorSpec{
					{
						Name: name,
						Graph: &machinelearningv1.PredictiveUnit{
							Name:           "classifier",
							Type:           &modelType,
							Implementation: &impl,
							Endpoint:       &machinelearningv1.Endpoint{Type: machinelearningv1.REST},
						},
					},
				},
			},
		}

		configMapName := types.NamespacedName{Name: "seldon-config",
			Namespace: "seldon-system"}

		configResult := &corev1.ConfigMap{}
		const timeout = time.Second * 30
		Eventually(func() error { return k8sClient.Get(context.TODO(), configMapName, configResult) }, timeout).
			Should(Succeed())

		// Run Defaulter
		instance.Default()

		Expect(k8sClient.Create(context.Background(), instance)).Should(Succeed())
		//time.Sleep(time.Second * 5)

		fetched := &machinelearningv1.SeldonDeployment{}
		Eventually(func() error {
			err := k8sClient.Get(context.Background(), key, fetched)
			return err
		}, timeout, interval).Should(BeNil())
		Expect(fetched.Spec.Name).Should(Equal(name))

		sPodSpec := utils.GetSeldonPodSpecForPredictiveUnit(&instance.Spec.Predictors[0], instance.Spec.Predictors[0].Graph.Name)
		depName := machinelearningv1.GetDeploymentName(instance, instance.Spec.Predictors[0], sPodSpec)
		depKey := types.NamespacedName{
			Name:      depName,
			Namespace: "default",
		}
		depFetched := &appsv1.Deployment{}
		Eventually(func() error {
			err := k8sClient.Get(context.Background(), depKey, depFetched)
			return err
		}, timeout, interval).Should(BeNil())
		Expect(len(depFetched.Spec.Template.Spec.Containers)).Should(Equal(3))
		for _, c := range depFetched.Spec.Template.Spec.Containers {
			if c.Name == constants.TFServingContainerName {
				for _, arg := range c.Args {
					if strings.Index(arg, constants.TfServingArgPort) == 0 {
						Expect(arg).To(Equal(constants.TfServingArgPort + strconv.Itoa(constants.TfServingGrpcPort)))
					}
					if strings.Index(arg, constants.TfServingArgRestPort) == 0 {
						Expect(arg).To(Equal(constants.TfServingArgRestPort + strconv.Itoa(constants.TfServingRestPort)))
					}
				}
			}
		}

		Expect(k8sClient.Delete(context.Background(), instance)).Should(Succeed())
	})

})

var _ = Describe("Create a prepacked tfserving server for tensorflow protocol and REST", func() {
	const timeout = time.Second * 30
	const interval = time.Second * 1
	const name = "pp3"
	const sdepName = "prepack3"
	modelName := "classifier"
	By("Creating a resource")
	It("should create a resource with defaults", func() {
		Expect(k8sClient).NotTo(BeNil())
		var modelType = machinelearningv1.MODEL
		var impl = machinelearningv1.PredictiveUnitImplementation(constants.PrePackedServerTensorflow)
		key := types.NamespacedName{
			Name:      sdepName,
			Namespace: "default",
		}
		instance := &machinelearningv1.SeldonDeployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      key.Name,
				Namespace: key.Namespace,
			},
			Spec: machinelearningv1.SeldonDeploymentSpec{
				Name: name,
				Predictors: []machinelearningv1.PredictorSpec{
					{
						Name:     "p1",
						Protocol: machinelearningv1.ProtocolTensorflow,
						Graph: &machinelearningv1.PredictiveUnit{
							Name:           modelName,
							Type:           &modelType,
							Implementation: &impl,
							Endpoint:       &machinelearningv1.Endpoint{Type: machinelearningv1.REST},
						},
					},
				},
			},
		}

		configMapName := types.NamespacedName{Name: "seldon-config",
			Namespace: "seldon-system"}

		configResult := &corev1.ConfigMap{}
		const timeout = time.Second * 30
		Eventually(func() error { return k8sClient.Get(context.TODO(), configMapName, configResult) }, timeout).
			Should(Succeed())

		// Run Defaulter
		instance.Default()

		Expect(k8sClient.Create(context.Background(), instance)).Should(Succeed())
		//time.Sleep(time.Second * 5)

		fetched := &machinelearningv1.SeldonDeployment{}
		Eventually(func() error {
			err := k8sClient.Get(context.Background(), key, fetched)
			return err
		}, timeout, interval).Should(BeNil())
		Expect(fetched.Spec.Name).Should(Equal(name))

		sPodSpec := utils.GetSeldonPodSpecForPredictiveUnit(&instance.Spec.Predictors[0], instance.Spec.Predictors[0].Graph.Name)
		depName := machinelearningv1.GetDeploymentName(instance, instance.Spec.Predictors[0], sPodSpec)
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
		for _, c := range depFetched.Spec.Template.Spec.Containers {
			if c.Name == modelName {
				for _, arg := range c.Args {
					if strings.Index(arg, constants.TfServingArgPort) == 0 {
						Expect(arg).To(Equal(constants.TfServingArgPort + strconv.Itoa(constants.TfServingGrpcPort)))
					}
					if strings.Index(arg, constants.TfServingArgRestPort) == 0 {
						Expect(arg).To(Equal(constants.TfServingArgRestPort + strconv.Itoa(constants.FirstPortNumber)))
					}
				}
			}
		}

		Expect(k8sClient.Delete(context.Background(), instance)).Should(Succeed())
	})

})

var _ = Describe("Create a prepacked tfserving server for tensorflow protocol and grpc", func() {
	const timeout = time.Second * 30
	const interval = time.Second * 1
	const name = "pp4"
	const sdepName = "prepack4"
	modelName := "classifier"
	By("Creating a resource")
	It("should create a resource with defaults", func() {
		Expect(k8sClient).NotTo(BeNil())
		var modelType = machinelearningv1.MODEL
		var impl = machinelearningv1.PredictiveUnitImplementation(constants.PrePackedServerTensorflow)
		key := types.NamespacedName{
			Name:      sdepName,
			Namespace: "default",
		}
		instance := &machinelearningv1.SeldonDeployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      key.Name,
				Namespace: key.Namespace,
			},
			Spec: machinelearningv1.SeldonDeploymentSpec{
				Name: name,
				Predictors: []machinelearningv1.PredictorSpec{
					{
						Name:      "p1",
						Protocol:  machinelearningv1.ProtocolTensorflow,
						Transport: machinelearningv1.TransportGrpc,
						Graph: &machinelearningv1.PredictiveUnit{
							Name:           modelName,
							Type:           &modelType,
							Implementation: &impl,
							Endpoint:       &machinelearningv1.Endpoint{Type: machinelearningv1.GRPC},
						},
					},
				},
			},
		}

		configMapName := types.NamespacedName{Name: "seldon-config",
			Namespace: "seldon-system"}

		configResult := &corev1.ConfigMap{}
		const timeout = time.Second * 30
		Eventually(func() error { return k8sClient.Get(context.TODO(), configMapName, configResult) }, timeout).
			Should(Succeed())

		// Run Defaulter
		instance.Default()

		Expect(k8sClient.Create(context.Background(), instance)).Should(Succeed())
		//time.Sleep(time.Second * 5)

		fetched := &machinelearningv1.SeldonDeployment{}
		Eventually(func() error {
			err := k8sClient.Get(context.Background(), key, fetched)
			return err
		}, timeout, interval).Should(BeNil())
		Expect(fetched.Spec.Name).Should(Equal(name))

		sPodSpec := utils.GetSeldonPodSpecForPredictiveUnit(&instance.Spec.Predictors[0], instance.Spec.Predictors[0].Graph.Name)
		depName := machinelearningv1.GetDeploymentName(instance, instance.Spec.Predictors[0], sPodSpec)
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
		for _, c := range depFetched.Spec.Template.Spec.Containers {
			if c.Name == modelName {
				for _, arg := range c.Args {
					if strings.Index(arg, constants.TfServingArgPort) == 0 {
						Expect(arg).To(Equal(constants.TfServingArgPort + strconv.Itoa(constants.FirstPortNumber)))
					}
					if strings.Index(arg, constants.TfServingArgRestPort) == 0 {
						Expect(arg).To(Equal(constants.TfServingArgRestPort + strconv.Itoa(constants.TfServingRestPort)))
					}
				}
			}
		}

		Expect(k8sClient.Delete(context.Background(), instance)).Should(Succeed())
	})

})
