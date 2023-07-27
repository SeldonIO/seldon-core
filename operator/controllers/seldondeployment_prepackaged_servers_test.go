package controllers

import (
	"context"
	"github.com/go-logr/logr/testr"
	"k8s.io/client-go/kubernetes/fake"
	"strconv"
	"strings"
	"testing"
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
	crfake "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("Create a prepacked sklearn server.", func() {
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

var _ = Describe("Create a prepacked tfserving server for Seldon protocol and REST with resource requests", func() {
	const interval = time.Second * 1
	const name = "pp2"
	const sdepName = "prepack2b"
	By("Creating a resource")
	It("should create a resource with defaults", func() {
		Expect(k8sClient).NotTo(BeNil())
		cpuValue := "2"
		cpuRequest, err := resource.ParseQuantity(cpuValue)
		Expect(err).To(BeNil())
		modelName := "classifier"
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
							Name:           modelName,
							Type:           &modelType,
							Implementation: &impl,
							Endpoint:       &machinelearningv1.Endpoint{Type: machinelearningv1.REST},
						},
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
										{
											Name: "tfserving",
											Resources: corev1.ResourceRequirements{
												Requests: corev1.ResourceList{corev1.ResourceCPU: cpuRequest},
											},
										},
									},
								},
							},
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
				Expect(c.Resources.Requests.Cpu().String()).To(Equal(cpuValue))
			} else if c.Name == modelName {
				Expect(c.Resources.Requests.Cpu().String()).To(Equal(cpuValue))
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
	const sdepName = "prepack4"
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
	const sdepName = "prepack5"
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
		sdepName = "prepack6"
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
		Expect(depFetched.Spec.Template.Spec.SecurityContext.RunAsUser).ToNot(BeNil())
		Expect(*depFetched.Spec.Template.Spec.SecurityContext.RunAsUser).To(Equal(int64(2)))

		Expect(k8sClient.Delete(context.Background(), instance)).Should(Succeed())
	})

	It("should use MLServer when choosing the V2 protocol", func() {
		sdepName = "prepack7"
		instance.Name = sdepName
		instance.Spec.Protocol = machinelearningv1.ProtocolV2
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

	It("should use MLServer when choosing the V2 protocol", func() {
		sdepName = "prepack8"
		instance.Name = sdepName
		instance.Spec.Protocol = machinelearningv1.ProtocolV2
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
	const sdepName = "prepack9"
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
				Protocol: machinelearningv1.ProtocolV2,
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

		predictor := instance.Spec.Predictors[0]
		sPodSpec, idx := utils.GetSeldonPodSpecForPredictiveUnit(&predictor, predictor.Graph.Name)
		depName := machinelearningv1.GetDeploymentName(instance, predictor, sPodSpec, idx)
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

		container := utils.GetContainerForDeployment(depFetched, predictor.Graph.Name)
		Expect(container.SecurityContext).To(BeNil())

		Expect(k8sClient.Delete(context.Background(), instance)).Should(Succeed())

		//j, _ := json.Marshal(depFetched)
		//fmt.Println(string(j))
	})

})

var _ = Describe("Create a prepacked triton server with deprecated kfserving protocol", func() {
	const timeout = time.Second * 30
	const interval = time.Second * 1
	const name = "pp1"
	const sdepName = "prepack10"
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
				Protocol: machinelearningv1.ProtocolKFServing,
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
	const sdepName = "prepack11"
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

var _ = Describe("Create a prepacked triton server with seldon.io/no-storage-initializer annotation", func() {
	const timeout = time.Second * 30
	const interval = time.Second * 1
	const name = "pp1"
	const sdepName = "prepack12"
	envExecutorUser = "2"
	By("Creating a resource")
	It("should create a resource with no storage initializer but triton args", func() {
		Expect(k8sClient).NotTo(BeNil())
		secretName := "s3-credentials"
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      secretName,
				Namespace: "default",
			},
			Type: corev1.SecretTypeOpaque,
			StringData: map[string]string{
				"AWS_DEFAULT_REGION":    "us-east-1",
				"AWS_ACCESS_KEY_ID":     "mykey",
				"AWS_SECRET_ACCESS_KEY": "mysecret",
			},
		}
		var modelType = machinelearningv1.MODEL
		modelName := "classifier"
		modelUri := "s3://mybucket/mymodel"
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
				Annotations: map[string]string{
					"seldon.io/no-storage-initializer": "true",
				},
				Name:     name,
				Protocol: machinelearningv1.ProtocolV2,
				Predictors: []machinelearningv1.PredictorSpec{
					{
						Annotations: map[string]string{
							"seldon.io/no-engine": "true",
						},
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
							Name:             modelName,
							ModelURI:         modelUri,
							Type:             &modelType,
							Implementation:   &impl,
							Endpoint:         &machinelearningv1.Endpoint{Type: machinelearningv1.REST},
							EnvSecretRefName: secretName,
							Parameters: []machinelearningv1.Parameter{
								{
									Name:  "model_control_mode",
									Type:  machinelearningv1.STRING,
									Value: "explicit",
								},
								{
									Name:  "load_model",
									Type:  machinelearningv1.STRING,
									Value: "model1",
								},
								{
									Name:  "strict_model_config",
									Type:  machinelearningv1.STRING,
									Value: "true",
								},
							},
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

		// Create secret
		Expect(k8sClient.Create(context.Background(), secret)).Should(Succeed())

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

		predictor := instance.Spec.Predictors[0]
		sPodSpec, idx := utils.GetSeldonPodSpecForPredictiveUnit(&predictor, predictor.Graph.Name)
		depName := machinelearningv1.GetDeploymentName(instance, predictor, sPodSpec, idx)
		depKey := types.NamespacedName{
			Name:      depName,
			Namespace: "default",
		}
		depFetched := &appsv1.Deployment{}
		Eventually(func() error {
			err := k8sClient.Get(context.Background(), depKey, depFetched)
			return err
		}, timeout, interval).Should(BeNil())
		Expect(len(depFetched.Spec.Template.Spec.InitContainers)).To(Equal(0)) // no storage
		Expect(len(depFetched.Spec.Template.Spec.Containers)).Should(Equal(1)) // no engine

		for _, c := range depFetched.Spec.Template.Spec.Containers {
			if c.Name == modelName {
				// Check we have 7 args total (server, http, grpc, model_uri, model_control_mode, load_model, strict_model_config)
				Expect(len(c.Args)).Should(Equal(7))
				for _, arg := range c.Args {
					if strings.Index(arg, constants.TritonArgModelRepository) == 0 {
						Expect(arg).To(Equal(constants.TritonArgModelRepository + modelUri))
					}
					if strings.Index(arg, constants.TritonArgModelControlMode) == 0 {
						Expect(arg).To(Equal(constants.TritonArgModelControlMode + "explicit"))
					}
					if strings.Index(arg, constants.TritonArgLoadModel) == 0 {
						Expect(arg).To(Equal(constants.TritonArgLoadModel + "model1"))
					}
					if strings.Index(arg, constants.TritonArgStrictModelConfig) == 0 {
						Expect(arg).To(Equal(constants.TritonArgStrictModelConfig + "true"))
					}
				}

				// Check env is set from secretName
				Expect(len(c.EnvFrom)).Should(Equal(1))
				Expect(c.EnvFrom[0].SecretRef.Name).Should(Equal(secretName))
			}
		}

		//j, _ := json.Marshal(depFetched)
		//fmt.Println(string(j))
	})

})

func createCustomModelWithUri() (*machinelearningv1.SeldonDeployment,
	*machinelearningv1.PredictorSpec,
	*components,
	*machinelearningv1.PredictiveUnit) {
	impl := machinelearningv1.SIMPLE_MODEL
	sdep := &machinelearningv1.SeldonDeployment{
		Spec: machinelearningv1.SeldonDeploymentSpec{
			Predictors: []machinelearningv1.PredictorSpec{
				{
					Name: "p1",
					ComponentSpecs: []*machinelearningv1.SeldonPodSpec{
						{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name: "model1",
									},
								},
							},
						},
					},
					Graph: machinelearningv1.PredictiveUnit{
						Name:           "model1",
						ModelURI:       "gs://abc",
						Implementation: &impl,
					},
				},
			},
		},
	}
	depName := machinelearningv1.GetDeploymentName(sdep, sdep.Spec.Predictors[0], sdep.Spec.Predictors[0].ComponentSpecs[0], 0)
	c := &components{
		deployments: []*appsv1.Deployment{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: depName,
				},
				Spec: appsv1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name: "model1",
								},
							},
						},
					},
				},
			},
		},
	}
	return sdep, &sdep.Spec.Predictors[0], c, &sdep.Spec.Predictors[0].Graph
}

func createPrePackedServer() (*machinelearningv1.SeldonDeployment,
	*machinelearningv1.PredictorSpec,
	*components,
	*machinelearningv1.PredictiveUnit) {
	impl := machinelearningv1.PredictiveUnitImplementation("SKLEARN_SERVER")
	sdep := &machinelearningv1.SeldonDeployment{
		Spec: machinelearningv1.SeldonDeploymentSpec{
			Predictors: []machinelearningv1.PredictorSpec{
				{
					Name: "p1",
					ComponentSpecs: []*machinelearningv1.SeldonPodSpec{
						{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name: "model1",
									},
								},
							},
						},
					},
					Graph: machinelearningv1.PredictiveUnit{
						Name:           "model1",
						ModelURI:       "gs://abc",
						Implementation: &impl,
					},
				},
			},
		},
	}
	depName := machinelearningv1.GetDeploymentName(sdep, sdep.Spec.Predictors[0], sdep.Spec.Predictors[0].ComponentSpecs[0], 0)
	c := &components{
		deployments: []*appsv1.Deployment{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: depName,
				},
				Spec: appsv1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name: "model1",
								},
							},
						},
					},
				},
			},
		},
	}
	return sdep, &sdep.Spec.Predictors[0], c, &sdep.Spec.Predictors[0].Graph
}

func getNumInitContainers(c *components) int {
	tot := 0
	for _, d := range c.deployments {
		tot = tot + len(d.Spec.Template.Spec.InitContainers)
	}
	return tot
}

func setupTestConfigMap() error {
	scheme := createScheme()
	machinelearningv1.C = crfake.NewFakeClientWithScheme(scheme)
	testConfigMap1 := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ControllerConfigMapName,
			Namespace: ControllerNamespace,
		},
		Data: configs,
	}
	return machinelearningv1.C.Create(context.TODO(), testConfigMap1)
}

func TestAddModelServersAndInitContainers(t *testing.T) {
	g := NewGomegaWithT(t)
	err := setupTestConfigMap()
	g.Expect(err).To(BeNil())
	type test struct {
		name      string
		error     bool
		generator func() (
			*machinelearningv1.SeldonDeployment,
			*machinelearningv1.PredictorSpec,
			*components,
			*machinelearningv1.PredictiveUnit)
		expectedNumInitContainers int
	}

	tests := []test{
		{
			name:                      "model uri in custom model",
			generator:                 createCustomModelWithUri,
			expectedNumInitContainers: 1,
		},
		{
			name:                      "model uri in prepackaged server",
			generator:                 createPrePackedServer,
			expectedNumInitContainers: 1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sdep, p, c, pu := test.generator()
			cs := fake.NewSimpleClientset(configMap)
			pi := NewPrePackedInitializer(context.TODO(), cs)
			logger := testr.New(t)
			err := pi.addModelServersAndInitContainers(sdep, p, c, pu, nil, logger)
			if test.error {
				g.Expect(err).ToNot(BeNil())
			} else {
				g.Expect(err).To(BeNil())
				g.Expect(getNumInitContainers(c)).To(Equal(test.expectedNumInitContainers))
			}
		})
	}
}
