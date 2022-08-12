/*
Copyright 2020 The Seldon Team.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	machinelearningv1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	"github.com/seldonio/seldon-core/operator/constants"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/client-go/kubernetes/fake"
	ctrl "sigs.k8s.io/controller-runtime"
)

func createTestSDepWithExplainer() *machinelearningv1.SeldonDeployment {
	var modelType = machinelearningv1.MODEL
	key := types.NamespacedName{
		Name:      "dep",
		Namespace: "default",
	}
	return &machinelearningv1.SeldonDeployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      key.Name,
			Namespace: key.Namespace,
		},
		Spec: machinelearningv1.SeldonDeploymentSpec{
			Name: "mydep",
			Predictors: []machinelearningv1.PredictorSpec{
				{
					Name: "p1",
					ComponentSpecs: []*machinelearningv1.SeldonPodSpec{
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
					Graph: machinelearningv1.PredictiveUnit{
						Name: "classifier",
						Type: &modelType,
					},
					Explainer: &machinelearningv1.Explainer{
						Type: machinelearningv1.AlibiAnchorsTabularExplainer,
					},
				},
			},
		},
	}
}

func TestExplainerImageRelated(t *testing.T) {
	g := NewGomegaWithT(t)
	scheme = createScheme()
	client := fake.NewSimpleClientset()
	_, err := client.CoreV1().ConfigMaps(ControllerNamespace).Create(context.TODO(), configMap, metav1.CreateOptions{})
	g.Expect(err).To(BeNil())
	ei := NewExplainerInitializer(context.TODO(), client)
	sdep := createTestSDepWithExplainer()
	svcName := "s"
	c := components{
		serviceDetails: map[string]*machinelearningv1.ServiceStatus{
			svcName: &machinelearningv1.ServiceStatus{
				HttpEndpoint: "a.svc.local",
			},
		},
	}
	envExplainerImage = "explainer:123"
	ei.createExplainer(sdep, &sdep.Spec.Predictors[0], &c, svcName, nil, ctrl.Log)
	g.Expect(len(c.deployments)).To(Equal(1))
	g.Expect(c.deployments[0].Spec.Template.Spec.Containers[0].Image).To(Equal(envExplainerImage))
}

var _ = Describe("createExplainer", func() {
	var r *SeldonDeploymentReconciler
	var mlDep *machinelearningv1.SeldonDeployment
	var p *machinelearningv1.PredictorSpec
	var c *components
	var pSvcName string

	BeforeEach(func() {

		p = &machinelearningv1.PredictorSpec{
			Name: "main",
		}

		mlDep = &machinelearningv1.SeldonDeployment{
			ObjectMeta: metav1.ObjectMeta{
				Name: "my-model",
			},
			Spec: machinelearningv1.SeldonDeploymentSpec{
				Predictors: []machinelearningv1.PredictorSpec{*p},
			},
		}

		c = &components{}

		pSvcName = machinelearningv1.GetPredictorKey(mlDep, p)

		r = &SeldonDeploymentReconciler{
			Client:    k8sManager.GetClient(),
			ClientSet: clientset,
			Log:       ctrl.Log.WithName("controllers").WithName("SeldonDeployment"),
			Scheme:    k8sManager.GetScheme(),
			Recorder:  k8sManager.GetEventRecorderFor(constants.ControllerName),
		}
	})

	DescribeTable(
		"Empty explainers should not create any component",
		func(explainer *machinelearningv1.Explainer) {
			scheme = createScheme()
			client := fake.NewSimpleClientset()
			_, err := client.CoreV1().ConfigMaps(ControllerNamespace).Create(context.TODO(), configMap, metav1.CreateOptions{})
			Expect(err).To(BeNil())
			p.Explainer = explainer
			ei := NewExplainerInitializer(context.TODO(), client)
			err = ei.createExplainer(mlDep, p, c, pSvcName, nil, r.Log)

			Expect(err).ToNot(HaveOccurred())
			Expect(c.deployments).To(BeEmpty())
		},
		Entry("nil", nil),
		Entry("empty type", &machinelearningv1.Explainer{}),
	)
})

var _ = Describe("Create a KFserving(V2) Seldon Deployment with explainer", func() {
	const timeout = time.Second * 30
	const interval = time.Second * 1
	namespaceName := rand.String(10)
	v2protocol := machinelearningv1.ProtocolV2
	explainerInitParameters := "{'a': 1, 'b': 's', 'c': {'c1': [1, 1]}}"
	By("Creating a resource")
	It("should create a resource with defaults", func() {
		Expect(k8sClient).NotTo(BeNil())

		modelType := machinelearningv1.MODEL
		modelImplementation := machinelearningv1.PredictiveUnitImplementation(
			machinelearningv1.PrepackSklearnName,
		)
		key := types.NamespacedName{
			Name:      "dep",
			Namespace: namespaceName,
		}
		instance := &machinelearningv1.SeldonDeployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      key.Name,
				Namespace: key.Namespace,
			},
			Spec: machinelearningv1.SeldonDeploymentSpec{
				Protocol: v2protocol,
				Name:     "mydep",
				Predictors: []machinelearningv1.PredictorSpec{
					{
						Name: "p1",
						ComponentSpecs: []*machinelearningv1.SeldonPodSpec{
							{
								Spec: v1.PodSpec{
									Containers: []v1.Container{
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
							Implementation: &modelImplementation,
						},
						Explainer: &machinelearningv1.Explainer{
							Type:           machinelearningv1.AlibiAnchorsImageExplainer,
							InitParameters: explainerInitParameters,
						},
					},
				},
			},
		}

		//Create namespace
		namespace := &v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: namespaceName,
			},
		}
		Expect(k8sClient.Create(context.Background(), namespace)).Should(Succeed())

		// Run Defaulter
		instance.Default()
		envUseExecutor = "true"
		envDefaultUser = "2"
		Expect(k8sClient.Create(context.Background(), instance)).Should(Succeed())

		depKey := types.NamespacedName{
			Name:      machinelearningv1.GetDeploymentName(instance, instance.Spec.Predictors[0], instance.Spec.Predictors[0].ComponentSpecs[0], 0),
			Namespace: namespaceName,
		}
		depFetched := &appsv1.Deployment{}
		Eventually(func() error {
			err := k8sClient.Get(context.Background(), depKey, depFetched)
			return err
		}, timeout, interval).Should(BeNil())

		//Check explainer deployment
		depKey = types.NamespacedName{
			Name:      machinelearningv1.GetExplainerDeploymentName(instance.Name, &instance.Spec.Predictors[0]),
			Namespace: namespaceName,
		}
		depFetched = &appsv1.Deployment{}
		Eventually(func() error {
			err := k8sClient.Get(context.Background(), depKey, depFetched)
			return err
		}, timeout, interval).Should(BeNil())
		explainerEnvs := depFetched.Spec.Template.Spec.Containers[0].Env
		explainerExpectedExtraEnvs, _ := getAlibiExplainExtraEnvVars(machinelearningv1.AlibiAnchorsImageExplainer, "dep-p1."+namespaceName+":8000", "classifier", explainerInitParameters)
		explainerExpectedEnvs := []v1.EnvVar{
			{Name: MLServerHTTPPortEnv, Value: "9000"},
			{Name: MLServerModelImplementationEnv, Value: MLServerAlibiExplainImplementation},
			{Name: MLServerModelNameEnv, Value: "dep-p1-explainer"},
			{Name: MLServerModelURIEnv, Value: DefaultModelLocalMountPath},
			{Name: MLServerModelExtraEnv, Value: explainerExpectedExtraEnvs},
		}
		Expect(explainerEnvs).Should(Equal(explainerExpectedEnvs))
		Expect(depFetched.Spec.Template.Spec.Containers[0].Image).Should(Equal("seldonio/mlserver:0.6.0"))
	})

})

var _ = Describe("Create a Seldon Deployment with explainer", func() {
	const timeout = time.Second * 30
	const interval = time.Second * 1
	namespaceName := rand.String(10)
	replicas := int32(3)
	replicasExplainer := int32(2)
	By("Creating a resource")
	It("should create a resource with defaults", func() {
		Expect(k8sClient).NotTo(BeNil())
		var modelType = machinelearningv1.MODEL
		key := types.NamespacedName{
			Name:      "dep",
			Namespace: namespaceName,
		}
		instance := &machinelearningv1.SeldonDeployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      key.Name,
				Namespace: key.Namespace,
			},
			Spec: machinelearningv1.SeldonDeploymentSpec{
				Replicas: &replicas,
				Name:     "mydep",
				Predictors: []machinelearningv1.PredictorSpec{
					{
						Name: "p1",
						ComponentSpecs: []*machinelearningv1.SeldonPodSpec{
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
						Graph: machinelearningv1.PredictiveUnit{
							Name: "classifier",
							Type: &modelType,
						},
						Explainer: &machinelearningv1.Explainer{
							Type:     machinelearningv1.AlibiAnchorsTabularExplainer,
							Replicas: &replicasExplainer,
						},
					},
				},
			},
		}

		//Create namespace
		namespace := &v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: namespaceName,
			},
		}
		Expect(k8sClient.Create(context.Background(), namespace)).Should(Succeed())

		// Run Defaulter
		instance.Default()
		envUseExecutor = "true"
		envDefaultUser = "2"
		Expect(k8sClient.Create(context.Background(), instance)).Should(Succeed())
		//time.Sleep(time.Second * 5)

		fetched := &machinelearningv1.SeldonDeployment{}
		Eventually(func() error {
			err := k8sClient.Get(context.Background(), key, fetched)
			return err
		}, timeout, interval).Should(BeNil())
		Expect(fetched.Name).Should(Equal("dep"))

		// Check deployment created
		depKey := types.NamespacedName{
			Name:      machinelearningv1.GetDeploymentName(instance, instance.Spec.Predictors[0], instance.Spec.Predictors[0].ComponentSpecs[0], 0),
			Namespace: namespaceName,
		}
		depFetched := &appsv1.Deployment{}
		Eventually(func() error {
			err := k8sClient.Get(context.Background(), depKey, depFetched)
			return err
		}, timeout, interval).Should(BeNil())
		Expect(len(depFetched.Spec.Template.Spec.Containers)).Should(Equal(2))
		Expect(*depFetched.Spec.Replicas).To(Equal(replicas))
		Expect(*depFetched.Spec.Template.Spec.SecurityContext.RunAsUser).To(Equal(int64(2)))

		//Check explainer deployment
		depKey = types.NamespacedName{
			Name:      machinelearningv1.GetExplainerDeploymentName(instance.Name, &instance.Spec.Predictors[0]),
			Namespace: namespaceName,
		}
		depFetched = &appsv1.Deployment{}
		Eventually(func() error {
			err := k8sClient.Get(context.Background(), depKey, depFetched)
			return err
		}, timeout, interval).Should(BeNil())
		Expect(len(depFetched.Spec.Template.Spec.Containers)).Should(Equal(1))
		Expect(*depFetched.Spec.Replicas).To(Equal(replicasExplainer))
		Expect(*depFetched.Spec.Template.Spec.SecurityContext.RunAsUser).To(Equal(int64(2)))
		Expect(depFetched.Spec.Template.Spec.Containers[0].Image).To(Equal("seldonio/alibiexplainer:1.2.0"))

		//Check svc created
		svcKey := types.NamespacedName{
			Name:      machinelearningv1.GetContainerServiceName("dep", instance.Spec.Predictors[0], &instance.Spec.Predictors[0].ComponentSpecs[0].Spec.Containers[0]),
			Namespace: namespaceName,
		}
		svcFetched := &v1.Service{}
		Eventually(func() error {
			err := k8sClient.Get(context.Background(), svcKey, svcFetched)
			return err
		}, timeout, interval).Should(BeNil())

		// Check events created
		serviceCreatedEvents := 0
		deploymentsCreatedEvents := 0
		evts, err := clientset.CoreV1().Events(namespaceName).Search(scheme, fetched)
		Expect(err).To(BeNil())
		for _, evt := range evts.Items {
			if evt.Reason == constants.EventsCreateService {
				serviceCreatedEvents = serviceCreatedEvents + 1
			} else if evt.Reason == constants.EventsCreateDeployment {
				deploymentsCreatedEvents = deploymentsCreatedEvents + 1
			}
		}

		Expect(serviceCreatedEvents).To(Equal(3))
		Expect(deploymentsCreatedEvents).To(Equal(2))

		Expect(k8sClient.Delete(context.Background(), instance)).Should(Succeed())

	})

})

var _ = Describe("Create a Seldon Deployment with zero replicas with explainer", func() {
	const timeout = time.Second * 30
	const interval = time.Second * 1
	namespaceName := rand.String(10)
	replicas := int32(0)
	replicasExplainer := int32(3)
	By("Creating a resource")
	It("should create a resource with defaults", func() {
		Expect(k8sClient).NotTo(BeNil())
		var modelType = machinelearningv1.MODEL
		key := types.NamespacedName{
			Name:      "dep",
			Namespace: namespaceName,
		}
		instance := &machinelearningv1.SeldonDeployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      key.Name,
				Namespace: key.Namespace,
			},
			Spec: machinelearningv1.SeldonDeploymentSpec{
				Name:     "mydep",
				Replicas: &replicas,
				Predictors: []machinelearningv1.PredictorSpec{
					{
						Name: "p1",
						ComponentSpecs: []*machinelearningv1.SeldonPodSpec{
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
						Graph: machinelearningv1.PredictiveUnit{
							Name: "classifier",
							Type: &modelType,
						},
						Explainer: &machinelearningv1.Explainer{
							Type:     machinelearningv1.AlibiAnchorsTabularExplainer,
							Replicas: &replicasExplainer,
						},
					},
				},
			},
		}

		//Create namespace
		namespace := &v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: namespaceName,
			},
		}
		Expect(k8sClient.Create(context.Background(), namespace)).Should(Succeed())

		// Run Defaulter
		instance.Default()
		envUseExecutor = "true"
		envDefaultUser = "2"
		Expect(k8sClient.Create(context.Background(), instance)).Should(Succeed())
		//time.Sleep(time.Second * 5)

		fetched := &machinelearningv1.SeldonDeployment{}
		Eventually(func() error {
			err := k8sClient.Get(context.Background(), key, fetched)
			return err
		}, timeout, interval).Should(BeNil())
		Expect(fetched.Name).Should(Equal("dep"))

		// Check deployment created
		depKey := types.NamespacedName{
			Name:      machinelearningv1.GetDeploymentName(instance, instance.Spec.Predictors[0], instance.Spec.Predictors[0].ComponentSpecs[0], 0),
			Namespace: namespaceName,
		}
		depFetched := &appsv1.Deployment{}
		Eventually(func() error {
			err := k8sClient.Get(context.Background(), depKey, depFetched)
			return err
		}, timeout, interval).Should(BeNil())
		Expect(*depFetched.Spec.Replicas).To(Equal(replicas))

		//Check explainer deployment
		depKey = types.NamespacedName{
			Name:      machinelearningv1.GetExplainerDeploymentName(instance.Name, &instance.Spec.Predictors[0]),
			Namespace: namespaceName,
		}
		depFetched = &appsv1.Deployment{}
		Eventually(func() error {
			err := k8sClient.Get(context.Background(), depKey, depFetched)
			return err
		}, timeout, interval).Should(BeNil())
		Expect(*depFetched.Spec.Replicas).To(Equal(replicas))

		//Check svc created
		svcKey := types.NamespacedName{
			Name:      machinelearningv1.GetContainerServiceName("dep", instance.Spec.Predictors[0], &instance.Spec.Predictors[0].ComponentSpecs[0].Spec.Containers[0]),
			Namespace: namespaceName,
		}
		svcFetched := &v1.Service{}
		Eventually(func() error {
			err := k8sClient.Get(context.Background(), svcKey, svcFetched)
			return err
		}, timeout, interval).Should(BeNil())

		// Check events created
		serviceCreatedEvents := 0
		deploymentsCreatedEvents := 0
		evts, err := clientset.CoreV1().Events(namespaceName).Search(scheme, fetched)
		Expect(err).To(BeNil())
		for _, evt := range evts.Items {
			if evt.Reason == constants.EventsCreateService {
				serviceCreatedEvents = serviceCreatedEvents + 1
			} else if evt.Reason == constants.EventsCreateDeployment {
				deploymentsCreatedEvents = deploymentsCreatedEvents + 1
			}
		}

		Expect(serviceCreatedEvents).To(Equal(3))
		Expect(deploymentsCreatedEvents).To(Equal(2))

		Expect(k8sClient.Delete(context.Background(), instance)).Should(Succeed())

	})

})

var _ = Describe("Test override of environment variable for explainers", func() {
	const blankName = ""
	const secretName = "SECRET_NAME"
	const overrideName = "OVERRIDE_NAME"
	By("Creating a predictive unit resource with an envSecretRefName and a default env var")
	It("Should override the default env var with envSecretRefName", func() {
		// Overriding environment variable
		PredictiveUnitDefaultEnvSecretRefName = secretName
		explainer := machinelearningv1.Explainer{EnvSecretRefName: overrideName}
		resultSecretName := extractExplainerEnvSecretRefName(&explainer)
		Expect(resultSecretName).To(Equal(overrideName))
	})
	By("Creating a predictive unit resource with an envSecretRefName and no default env var")
	It("Should override the default env var with envSecretRefName", func() {
		// Overriding environment variable
		PredictiveUnitDefaultEnvSecretRefName = blankName
		explainer := machinelearningv1.Explainer{EnvSecretRefName: overrideName}
		resultSecretName := extractExplainerEnvSecretRefName(&explainer)
		Expect(resultSecretName).To(Equal(overrideName))
	})
	By("Creating a predictive unit resource without an envSecretRefName and a default env var")
	It("Should set the value to the default env var", func() {
		// Overriding environment variable
		PredictiveUnitDefaultEnvSecretRefName = secretName
		explainer := machinelearningv1.Explainer{}
		resultSecretName := extractExplainerEnvSecretRefName(&explainer)
		Expect(resultSecretName).To(Equal(secretName))
	})
	By("Creating a predictive unit resource without an envSecretRefName and without default env var")
	It("Should set the value to empty string", func() {
		// Overriding environment variable
		PredictiveUnitDefaultEnvSecretRefName = blankName
		explainer := machinelearningv1.Explainer{}
		resultSecretName := extractExplainerEnvSecretRefName(&explainer)
		Expect(resultSecretName).To(Equal(blankName))
	})
})
