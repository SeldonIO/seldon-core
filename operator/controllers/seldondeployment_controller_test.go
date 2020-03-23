/*
Copyright 2019 The Seldon Team.

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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	machinelearningv1 "github.com/seldonio/seldon-core/operator/apis/machinelearning/v1"
	"github.com/seldonio/seldon-core/operator/constants"
	"io/ioutil"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2beta2 "k8s.io/api/autoscaling/v2beta1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/rand"
	"path/filepath"
	"time"
)

func helperLoadBytes(name string) []byte {
	path := filepath.Join("testdata", name) // relative path
	bytes, _ := ioutil.ReadFile(path)
	return bytes
}

var _ = Describe("Create a Seldon Deployment", func() {
	const timeout = time.Second * 30
	const interval = time.Second * 1
	namespaceName := rand.String(10)
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
						Graph: &machinelearningv1.PredictiveUnit{
							Name: "classifier",
							Type: &modelType,
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

		Expect(serviceCreatedEvents).To(Equal(2))
		Expect(deploymentsCreatedEvents).To(Equal(1))

		Expect(k8sClient.Delete(context.Background(), instance)).Should(Succeed())

	})

})

var _ = Describe("Create a Seldon Deployment", func() {
	const timeout = time.Second * 30
	const interval = time.Second * 1
	namespaceName := rand.String(10)
	By("Creating a broken resource")
	It("should fail to create resources", func() {
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
											VolumeMounts: []v1.VolumeMount{
												{
													MountPath: "/tmp",
													Name:      "myvol",
												},
											},
										},
									},
								},
							},
						},
						Graph: &machinelearningv1.PredictiveUnit{
							Name: "classifier",
							Type: &modelType,
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

		Expect(k8sClient.Create(context.Background(), instance)).Should(Succeed())
		//time.Sleep(time.Second * 5)

		fetched := &machinelearningv1.SeldonDeployment{}
		Eventually(func() error {
			err := k8sClient.Get(context.Background(), key, fetched)
			return err
		}, timeout, interval).Should(BeNil())
		Expect(fetched.Name).Should(Equal("dep"))

		Eventually(func() bool {
			evts, err := clientset.CoreV1().Events(namespaceName).Search(scheme, fetched)
			if err == nil {
				for _, evt := range evts.Items {
					if evt.Reason == constants.EventsInternalError {
						return true
					}
				}
			}
			return false
		}, timeout, interval).Should(BeTrue())
		Expect(fetched.Name).Should(Equal("dep"))

		// Check events created
		serviceCreatedEvents := 0
		deploymentsCreatedEvents := 0
		internalErrorEvents := 0
		evts, err := clientset.CoreV1().Events(namespaceName).Search(scheme, fetched)
		Expect(err).To(BeNil())
		for _, evt := range evts.Items {
			if evt.Reason == constants.EventsCreateService {
				serviceCreatedEvents = serviceCreatedEvents + 1
			} else if evt.Reason == constants.EventsCreateDeployment {
				deploymentsCreatedEvents = deploymentsCreatedEvents + 1
			} else if evt.Reason == constants.EventsInternalError {
				internalErrorEvents = internalErrorEvents + 1
			}
		}

		Expect(serviceCreatedEvents).To(Equal(2))
		Expect(deploymentsCreatedEvents).To(Equal(0))
		Expect(internalErrorEvents).To(Equal(1))

		Eventually(func() error {
			err := k8sClient.Get(context.Background(), key, fetched)
			return err
		}, timeout, interval).Should(BeNil())
		Expect(fetched.Status.State).Should(Equal(machinelearningv1.StatusStateFailed))

		Expect(k8sClient.Delete(context.Background(), instance)).Should(Succeed())

	})

})

var _ = Describe("Create a Seldon Deployment with two ComponentSpecs", func() {
	const timeout = time.Second * 30
	const interval = time.Second * 1
	namespaceName := rand.String(10)
	It("should create the correct resources", func() {
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
							{
								Spec: v1.PodSpec{
									Containers: []v1.Container{
										{
											Image: "seldonio/mock_classifier:1.0",
											Name:  "classifier2",
										},
									},
								},
							},
						},
						Graph: &machinelearningv1.PredictiveUnit{
							Name: "classifier",
							Type: &modelType,
							Children: []machinelearningv1.PredictiveUnit{
								{
									Name: "classifier2",
									Type: &modelType,
								},
							},
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

		Expect(k8sClient.Create(context.Background(), instance)).Should(Succeed())
		//time.Sleep(time.Second * 5)

		fetched := &machinelearningv1.SeldonDeployment{}
		Eventually(func() error {
			err := k8sClient.Get(context.Background(), key, fetched)
			return err
		}, timeout, interval).Should(BeNil())
		Expect(fetched.Name).Should(Equal("dep"))

		// Check deployment created for 1st componenSpec
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

		depKey = types.NamespacedName{
			Name:      machinelearningv1.GetDeploymentName(instance, instance.Spec.Predictors[0], instance.Spec.Predictors[0].ComponentSpecs[1], 1),
			Namespace: namespaceName,
		}
		depFetched = &appsv1.Deployment{}
		Eventually(func() error {
			err := k8sClient.Get(context.Background(), depKey, depFetched)
			return err
		}, timeout, interval).Should(BeNil())
		Expect(len(depFetched.Spec.Template.Spec.Containers)).Should(Equal(1))

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

var _ = Describe("Create a Seldon Deployment with hpa", func() {
	const timeout = time.Second * 30
	const interval = time.Second * 1
	namespaceName := rand.String(10)
	utilization := int32(10)
	It("should create a resources", func() {
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
								HpaSpec: &machinelearningv1.SeldonHpaSpec{
									MinReplicas: nil,
									MaxReplicas: 10,
									Metrics: []autoscalingv2beta2.MetricSpec{
										{
											Type: autoscalingv2beta2.ResourceMetricSourceType,
											Resource: &autoscalingv2beta2.ResourceMetricSource{
												Name:                     v1.ResourceCPU,
												TargetAverageUtilization: &utilization,
											},
										},
									},
								},
							},
						},
						Graph: &machinelearningv1.PredictiveUnit{
							Name: "classifier",
							Type: &modelType,
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

		//Check hpa created
		hpaKey := types.NamespacedName{
			Name:      machinelearningv1.GetContainerServiceName("dep", instance.Spec.Predictors[0], &instance.Spec.Predictors[0].ComponentSpecs[0].Spec.Containers[0]),
			Namespace: namespaceName,
		}
		hpaFetched := &v1.Service{}
		Eventually(func() error {
			err := k8sClient.Get(context.Background(), hpaKey, hpaFetched)
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

		Expect(serviceCreatedEvents).To(Equal(2))
		Expect(deploymentsCreatedEvents).To(Equal(1))

		Expect(k8sClient.Delete(context.Background(), instance)).Should(Succeed())

	})

})
