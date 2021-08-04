package controllers

import (
	"context"
	"strconv"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	machinelearningv1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	"github.com/seldonio/seldon-core/operator/constants"
	"github.com/seldonio/seldon-core/operator/utils"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("Create a prepacked sklearn server", func() {
	const timeout = time.Second * 30
	const interval = time.Second * 1
	const name = "pp1"
	const sdepName = "prepack1"
	envExecutorUser = "2"
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
						Graph: machinelearningv1.PredictiveUnit{
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
		Expect(fetched.Name).Should(Equal(sdepName))

		sPodSpec, idx := utils.GetSeldonPodSpecForPredictiveUnit(&instance.Spec.Predictors[0], instance.Spec.Predictors[0].Graph.Name)
		depName := machinelearningv1.GetDeploymentName(instance, instance.Spec.Predictors[0], sPodSpec, idx)
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
						Graph: machinelearningv1.PredictiveUnit{
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
		Expect(fetched.Name).Should(Equal(sdepName))

		sPodSpec, idx := utils.GetSeldonPodSpecForPredictiveUnit(&instance.Spec.Predictors[0], instance.Spec.Predictors[0].Graph.Name)
		depName := machinelearningv1.GetDeploymentName(instance, instance.Spec.Predictors[0], sPodSpec, idx)
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
				Name:     name,
				Protocol: machinelearningv1.ProtocolTensorflow,
				Predictors: []machinelearningv1.PredictorSpec{
					{
						Name: "p1",
						Graph: machinelearningv1.PredictiveUnit{
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
		Expect(fetched.Name).Should(Equal(sdepName))

		sPodSpec, idx := utils.GetSeldonPodSpecForPredictiveUnit(&instance.Spec.Predictors[0], instance.Spec.Predictors[0].Graph.Name)
		depName := machinelearningv1.GetDeploymentName(instance, instance.Spec.Predictors[0], sPodSpec, idx)
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
						Expect(arg).To(Equal(constants.TfServingArgPort + strconv.Itoa(int(constants.FirstGrpcPortNumber))))
					}
					if strings.Index(arg, constants.TfServingArgRestPort) == 0 {
						Expect(arg).To(Equal(constants.TfServingArgRestPort + strconv.Itoa(int(constants.FirstHttpPortNumber))))
					}
				}
			}
		}

		Expect(k8sClient.Delete(context.Background(), instance)).Should(Succeed())
	})

})

var _ = Describe("Create a prepacked tfserving server for tensorflow protocol and REST with existing container", func() {
	const interval = time.Second * 1
	const name = "pp3"
	const sdepName = "prepack3b"
	modelName := "classifier"
	By("Creating a resource")
	It("should create a resource with defaults", func() {
		Expect(k8sClient).NotTo(BeNil())
		cpuRequest, err := resource.ParseQuantity("2")
		Expect(err).To(BeNil())
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
				Name:     name,
				Protocol: machinelearningv1.ProtocolTensorflow,
				Predictors: []machinelearningv1.PredictorSpec{
					{
						Name: "p1",
						ComponentSpecs: []*machinelearningv1.SeldonPodSpec{
							{
								Spec: corev1.PodSpec{
									Containers: []corev1.Container{
										{
											Name: modelName,
											Resources: corev1.ResourceRequirements{
												Requests: corev1.ResourceList{corev1.ResourceCPU: cpuRequest},
											},
										},
									},
								},
							},
						},
						Graph: machinelearningv1.PredictiveUnit{
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
		const timeout = time.Second * 300
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
		Expect(fetched.Name).Should(Equal(sdepName))

		sPodSpec, idx := utils.GetSeldonPodSpecForPredictiveUnit(&instance.Spec.Predictors[0], instance.Spec.Predictors[0].Graph.Name)
		depName := machinelearningv1.GetDeploymentName(instance, instance.Spec.Predictors[0], sPodSpec, idx)
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
				Expect(c.Image).ToNot(BeNil())
				Expect(c.Resources.Requests.Cpu()).ToNot(BeNil())
				Expect(*c.Resources.Requests.Cpu()).To(Equal(cpuRequest))
				for _, arg := range c.Args {
					if strings.Index(arg, constants.TfServingArgPort) == 0 {
						Expect(arg).To(Equal(constants.TfServingArgPort + strconv.Itoa(int(constants.FirstGrpcPortNumber))))
					}
					if strings.Index(arg, constants.TfServingArgRestPort) == 0 {
						Expect(arg).To(Equal(constants.TfServingArgRestPort + strconv.Itoa(int(constants.FirstHttpPortNumber))))
					}
				}
			}
		}

		Expect(k8sClient.Delete(context.Background(), instance)).Should(Succeed())
	})

})

var _ = Describe("Test override of environment variable", func() {
	const blankName = ""
	const secretName = "SECRET_NAME"
	const overrideName = "OVERRIDE_NAME"
	By("Creating a predictive unit resource with an envSecretRefName and a default env var")
	It("Should override the default env var with envSecretRefName", func() {
		// Overriding environment variable
		PredictiveUnitDefaultEnvSecretRefName = secretName
		predictiveUnit := machinelearningv1.PredictiveUnit{EnvSecretRefName: overrideName}
		resultSecretName := extractEnvSecretRefName(&predictiveUnit)
		Expect(resultSecretName).To(Equal(overrideName))
	})
	By("Creating a predictive unit resource with an envSecretRefName and no default env var")
	It("Should override the default env var with envSecretRefName", func() {
		// Overriding environment variable
		PredictiveUnitDefaultEnvSecretRefName = blankName
		predictiveUnit := machinelearningv1.PredictiveUnit{EnvSecretRefName: overrideName}
		resultSecretName := extractEnvSecretRefName(&predictiveUnit)
		Expect(resultSecretName).To(Equal(overrideName))
	})
	By("Creating a predictive unit resource without an envSecretRefName and a default env var")
	It("Should set the value to the default env var", func() {
		// Overriding environment variable
		PredictiveUnitDefaultEnvSecretRefName = secretName
		predictiveUnit := machinelearningv1.PredictiveUnit{}
		resultSecretName := extractEnvSecretRefName(&predictiveUnit)
		Expect(resultSecretName).To(Equal(secretName))
	})
	By("Creating a predictive unit resource without an envSecretRefName and without default env var")
	It("Should set the value to empty string", func() {
		// Overriding environment variable
		PredictiveUnitDefaultEnvSecretRefName = blankName
		predictiveUnit := machinelearningv1.PredictiveUnit{}
		resultSecretName := extractEnvSecretRefName(&predictiveUnit)
		Expect(resultSecretName).To(Equal(blankName))
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
				Name:      name,
				Protocol:  machinelearningv1.ProtocolTensorflow,
				Transport: machinelearningv1.TransportGrpc,
				Predictors: []machinelearningv1.PredictorSpec{
					{
						Name: "p1",
						Graph: machinelearningv1.PredictiveUnit{
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
		Expect(fetched.Name).Should(Equal(sdepName))

		sPodSpec, idx := utils.GetSeldonPodSpecForPredictiveUnit(&instance.Spec.Predictors[0], instance.Spec.Predictors[0].Graph.Name)
		depName := machinelearningv1.GetDeploymentName(instance, instance.Spec.Predictors[0], sPodSpec, idx)
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
						Expect(arg).To(Equal(constants.TfServingArgPort + strconv.Itoa(int(constants.FirstGrpcPortNumber))))
					}
					if strings.Index(arg, constants.TfServingArgRestPort) == 0 {
						Expect(arg).To(Equal(constants.TfServingArgRestPort + strconv.Itoa(int(constants.FirstHttpPortNumber))))
					}
				}
			}
		}

		Expect(k8sClient.Delete(context.Background(), instance)).Should(Succeed())
	})

})

var _ = Describe("Create a prepacked sklearn server", func() {

	const (
		timeout         = time.Second * 30
		interval        = time.Second * 1
		name            = "pp1"
		envExecutorUser = "2"
	)

	var sdepName string
	var instance *machinelearningv1.SeldonDeployment
	var key types.NamespacedName

	BeforeEach(func() {
		sdepName = "prepack5"
		modelType := machinelearningv1.MODEL
		impl := machinelearningv1.PredictiveUnitImplementation(constants.PrePackedServerSklearn)

		key = types.NamespacedName{
			Name:      sdepName,
			Namespace: "default",
		}

		instance = &machinelearningv1.SeldonDeployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      key.Name,
				Namespace: key.Namespace,
			},
			Spec: machinelearningv1.SeldonDeploymentSpec{
				Name: name,
				Predictors: []machinelearningv1.PredictorSpec{
					{
						Name: "p1",
						Graph: machinelearningv1.PredictiveUnit{
							Name:           "classifier",
							Type:           &modelType,
							Implementation: &impl,
							Endpoint:       &machinelearningv1.Endpoint{Type: machinelearningv1.REST},
						},
					},
				},
			},
		}

	})

	BeforeEach(func() {
		configMapName := types.NamespacedName{Name: "seldon-config",
			Namespace: "seldon-system"}

		configResult := &corev1.ConfigMap{}
		Eventually(func() error {
			return k8sClient.Get(context.TODO(), configMapName, configResult)
		}, time.Second*30).Should(Succeed())
	})

	It("should create a resource with defaults and security context", func() {
		// Run Defaulter
		instance.Default()

		//set security user
		envUseExecutor = "true"
		envDefaultUser = "2"

		Expect(k8sClient.Create(context.Background(), instance)).Should(Succeed())

		fetched := &machinelearningv1.SeldonDeployment{}
		Eventually(func() error {
			err := k8sClient.Get(context.Background(), key, fetched)
			return err
		}, timeout, interval).Should(BeNil())
		Expect(fetched.Name).Should(Equal(sdepName))

		sPodSpec, idx := utils.GetSeldonPodSpecForPredictiveUnit(&instance.Spec.Predictors[0], instance.Spec.Predictors[0].Graph.Name)
		depName := machinelearningv1.GetDeploymentName(instance, instance.Spec.Predictors[0], sPodSpec, idx)
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
		Expect(depFetched.Spec.Template.Spec.SecurityContext).ToNot(BeNil())
		Expect(*depFetched.Spec.Template.Spec.SecurityContext.RunAsUser).To(Equal(int64(2)))

		Expect(k8sClient.Delete(context.Background(), instance)).Should(Succeed())
	})

	It("should use MLServer when choosing the KFServing protocol", func() {
		sdepName = "prepack6"
		instance.Name = sdepName
		instance.Spec.Protocol = machinelearningv1.ProtocolKfserving
		key.Name = sdepName

		instance.Default()
		Expect(k8sClient.Create(context.Background(), instance)).Should(Succeed())

		fetched := &machinelearningv1.SeldonDeployment{}
		Eventually(func() error {
			err := k8sClient.Get(context.Background(), key, fetched)
			return err
		}, timeout, interval).Should(BeNil())
		Expect(fetched.Name).Should(Equal(sdepName))

		predictor := instance.Spec.Predictors[0]
		sPodSpec, idx := utils.GetSeldonPodSpecForPredictiveUnit(&predictor, predictor.Graph.Name)

		depName := machinelearningv1.GetDeploymentName(instance, predictor, sPodSpec, idx)
		depKey := types.NamespacedName{Name: depName, Namespace: "default"}
		depFetched := &appsv1.Deployment{}
		Eventually(func() error {
			err := k8sClient.Get(context.Background(), depKey, depFetched)
			return err
		}, timeout, interval).Should(BeNil())

		container := utils.GetContainerForDeployment(depFetched, predictor.Graph.Name)
		expectedImage, _ := getMLServerImage(&predictor.Graph)
		Expect(container.Image).To(Equal(expectedImage))

		Expect(k8sClient.Delete(context.Background(), instance)).Should(Succeed())
	})
})

var _ = Describe("Create a prepacked triton server", func() {
	const timeout = time.Second * 30
	const interval = time.Second * 1
	const name = "pp1"
	const sdepName = "prepack5"
	envExecutorUser = "2"
	By("Creating a resource")
	It("should create a resource with defaults and security context", func() {
		Expect(k8sClient).NotTo(BeNil())
		var modelType = machinelearningv1.MODEL
		modelName := "classifier"
		var impl = machinelearningv1.PredictiveUnitImplementation(constants.PrePackedServerTriton)
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
				Name:     name,
				Protocol: machinelearningv1.ProtocolKfserving,
				Predictors: []machinelearningv1.PredictorSpec{
					{
						Name: "p1",
						ComponentSpecs: []*machinelearningv1.SeldonPodSpec{
							{
								Spec: corev1.PodSpec{
									Containers: []corev1.Container{
										{
											Name: modelName,
										},
									},
								},
							},
						},
						Graph: machinelearningv1.PredictiveUnit{
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

		//set security user
		envUseExecutor = "true"
		envDefaultUser = "2"

		Expect(k8sClient.Create(context.Background(), instance)).Should(Succeed())
		//time.Sleep(time.Second * 5)

		fetched := &machinelearningv1.SeldonDeployment{}
		Eventually(func() error {
			err := k8sClient.Get(context.Background(), key, fetched)
			return err
		}, timeout, interval).Should(BeNil())
		Expect(fetched.Name).Should(Equal(sdepName))

		sPodSpec, idx := utils.GetSeldonPodSpecForPredictiveUnit(&instance.Spec.Predictors[0], instance.Spec.Predictors[0].Graph.Name)
		depName := machinelearningv1.GetDeploymentName(instance, instance.Spec.Predictors[0], sPodSpec, idx)
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
		Expect(depFetched.Spec.Template.Spec.SecurityContext).ToNot(BeNil())
		Expect(*depFetched.Spec.Template.Spec.SecurityContext.RunAsUser).To(Equal(int64(2)))

		Expect(k8sClient.Delete(context.Background(), instance)).Should(Succeed())

		//j, _ := json.Marshal(depFetched)
		//fmt.Println(string(j))
	})

})

var _ = Describe("Create a prepacked mlflow server with existing container", func() {
	const timeout = time.Second * 30
	const interval = time.Second * 1
	const name = "pp1"
	const sdepName = "prepack1"
	envExecutorUser = "2"
	By("Creating a resource")
	It("should create a resource with defaults", func() {
		Expect(k8sClient).NotTo(BeNil())
		var modelType = machinelearningv1.MODEL
		var impl = machinelearningv1.PredictiveUnitImplementation(constants.PrePackedMlflow)
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
						ComponentSpecs: []*machinelearningv1.SeldonPodSpec{
							{
								Spec: corev1.PodSpec{
									Containers: []corev1.Container{
										{
											Name: "classifier",
										},
									},
								},
							},
						},
						Graph: machinelearningv1.PredictiveUnit{
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
		Expect(fetched.Name).Should(Equal(sdepName))

		sPodSpec, idx := utils.GetSeldonPodSpecForPredictiveUnit(&instance.Spec.Predictors[0], instance.Spec.Predictors[0].Graph.Name)
		depName := machinelearningv1.GetDeploymentName(instance, instance.Spec.Predictors[0], sPodSpec, idx)
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
