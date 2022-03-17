package v1

import (
	"context"
	. "github.com/onsi/gomega"
	"github.com/seldonio/seldon-core/operator/constants"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

func TestValidProtocolTransportServerType(t *testing.T) {
	g := NewGomegaWithT(t)
	spec := &SeldonDeploymentSpec{
		ServerType: ServerRPC,
		Protocol:   ProtocolTensorflow,
		Transport:  TransportGrpc,
		Predictors: []PredictorSpec{
			{
				Name: "p1",
				ComponentSpecs: []*SeldonPodSpec{
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
				Graph: PredictiveUnit{
					Name: "classifier",
				},
			},
		},
	}

	spec.DefaultSeldonDeployment("mydep", "default")
	err := spec.ValidateSeldonDeployment()
	g.Expect(err).To(BeNil())
}

func TestNoGraphType(t *testing.T) {
	g := NewGomegaWithT(t)
	spec := &SeldonDeploymentSpec{
		ServerType: ServerRPC,
		Protocol:   ProtocolTensorflow,
		Transport:  TransportGrpc,
		Predictors: []PredictorSpec{
			{
				Name: "p1",
				ComponentSpecs: []*SeldonPodSpec{
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
				Graph: PredictiveUnit{
					Name: "classifier",
				},
			},
		},
	}

	err := spec.ValidateSeldonDeployment()
	g.Expect(err).To(BeNil())
}

func createScheme() *runtime.Scheme {
	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)
	_ = appsv1.AddToScheme(scheme)
	_ = v1.AddToScheme(scheme)
	_ = v1beta1.AddToScheme(scheme)
	_ = serializer.NewCodecFactory(scheme).UniversalDeserializer().Decode
	return scheme
}

func setupTestConfigMap() error {
	scheme := createScheme()
	C = fake.NewFakeClientWithScheme(scheme)
	testConfigMap1 := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ControllerConfigMapName,
			Namespace: ControllerNamespace,
		},
		Data: configs,
	}
	return C.Create(context.TODO(), testConfigMap1)
}

var configs = map[string]string{
	"predictor_servers": `{
                          "TENSORFLOW_SERVER": {
          "protocols" : {
            "tensorflow": {
              "image": "tensorflow/serving",
              "defaultImageVersion": "2.1.0"
              },
            "seldon": {
              "image": "seldonio/tfserving-proxy",
              "defaultImageVersion": "1.3.0-dev"
              }
            }
        },
        "SKLEARN_SERVER": {
          "protocols" : {
            "seldon": {
              "image": "seldonio/sklearnserver",
              "defaultImageVersion": "1.3.0-dev"
              },
            "kfserving": {
              "image": "seldonio/mlserver",
              "defaultImageVersion": "0.1.0"
              }
            }
        },
        "XGBOOST_SERVER": {
          "protocols" : {
            "seldon": {
              "image": "seldonio/xgboostserver",
              "defaultImageVersion": "1.3.0-dev"
              },
            "kfserving": {
              "image": "seldonio/mlserver",
              "defaultImageVersion": "0.1.0"
              }
            }
        },
        "MLFLOW_SERVER": {
          "protocols" : {
            "seldon": {
              "image": "seldonio/mlflowserver",
              "defaultImageVersion": "1.3.0-dev"
              },
            "kfserving": {
              "image": "seldonio/mlserver",
              "defaultImageVersion": "0.1.0"
              }
            }
        },
        "TRITON_SERVER": {
          "protocols" : {
            "kfserving": {
              "image": "nvcr.io/nvidia/tritonserver",
              "defaultImageVersion": "21.08-py3"
              }
            }
        },
        "CUSTOM_SERVER": {
          "protocols" : {
            "seldon": {
              "image": "custom",
              "defaultImageVersion": "0.1"
              }
            }
         }
         }`,
}

func TestValidateBadProtocol(t *testing.T) {
	g := NewGomegaWithT(t)
	spec := &SeldonDeploymentSpec{
		Protocol: "abc",
		Predictors: []PredictorSpec{
			{
				Name: "p1",
				ComponentSpecs: []*SeldonPodSpec{
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
				Graph: PredictiveUnit{
					Name: "classifier",
				},
			},
		},
	}

	spec.DefaultSeldonDeployment("mydep", "default")
	err := spec.ValidateSeldonDeployment()
	g.Expect(err).ToNot(BeNil())
	serr := err.(*errors.StatusError)
	g.Expect(serr.Status().Code).To(Equal(int32(422)))
	g.Expect(len(serr.Status().Details.Causes)).To(Equal(1))
	g.Expect(serr.Status().Details.Causes[0].Type).To(Equal(v12.CauseTypeFieldValueInvalid))
	g.Expect(serr.Status().Details.Causes[0].Field).To(Equal("spec"))
}

func TestValidateBadTransport(t *testing.T) {
	g := NewGomegaWithT(t)
	impl := MODEL
	spec := &SeldonDeploymentSpec{
		Transport: "abc",
		Predictors: []PredictorSpec{
			{
				Name: "p1",
				ComponentSpecs: []*SeldonPodSpec{
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
				Graph: PredictiveUnit{
					Name: "classifier",
					Type: &impl,
				},
			},
		},
	}

	spec.DefaultSeldonDeployment("mydep", "default")
	err := spec.ValidateSeldonDeployment()
	g.Expect(err).ToNot(BeNil())
	serr := err.(*errors.StatusError)
	g.Expect(serr.Status().Code).To(Equal(int32(422)))
	g.Expect(len(serr.Status().Details.Causes)).To(Equal(1))
	g.Expect(serr.Status().Details.Causes[0].Type).To(Equal(v12.CauseTypeFieldValueInvalid))
	g.Expect(serr.Status().Details.Causes[0].Field).To(Equal("spec"))
}

func TestValidateBadServerType(t *testing.T) {
	g := NewGomegaWithT(t)
	spec := &SeldonDeploymentSpec{
		ServerType: "abc",
		Predictors: []PredictorSpec{
			{
				Name: "p1",
				ComponentSpecs: []*SeldonPodSpec{
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
				Graph: PredictiveUnit{
					Name: "classifier",
				},
			},
		},
	}

	spec.DefaultSeldonDeployment("mydep", "default")
	err := spec.ValidateSeldonDeployment()
	g.Expect(err).ToNot(BeNil())
	serr := err.(*errors.StatusError)
	g.Expect(serr.Status().Code).To(Equal(int32(422)))
	g.Expect(len(serr.Status().Details.Causes)).To(Equal(1))
	g.Expect(serr.Status().Details.Causes[0].Type).To(Equal(v12.CauseTypeFieldValueInvalid))
	g.Expect(serr.Status().Details.Causes[0].Field).To(Equal("spec"))
}

func TestValidateMixedTransport(t *testing.T) {
	g := NewGomegaWithT(t)
	impl := MODEL
	spec := &SeldonDeploymentSpec{
		Transport: TransportRest,
		Predictors: []PredictorSpec{
			{
				Name: "p1",
				ComponentSpecs: []*SeldonPodSpec{
					{
						Spec: v1.PodSpec{
							Containers: []v1.Container{
								{
									Image: "seldonio/mock_classifier:1.0",
									Name:  "classifier",
								},
								{
									Image: "seldonio/mock_classifier:1.0",
									Name:  "classifier2",
								},
							},
						},
					},
				},
				Graph: PredictiveUnit{
					Name: "classifier",
					Type: &impl,
					Endpoint: &Endpoint{
						Type: GRPC,
					},
					Children: []PredictiveUnit{
						{
							Name: "classifier2",
							Type: &impl,
							Endpoint: &Endpoint{
								Type: REST,
							},
						},
					},
				},
			},
		},
	}

	spec.DefaultSeldonDeployment("mydep", "default")
	err := spec.ValidateSeldonDeployment()
	g.Expect(err).ToNot(BeNil())
	serr := err.(*errors.StatusError)
	g.Expect(serr.Status().Code).To(Equal(int32(422)))
	g.Expect(len(serr.Status().Details.Causes)).To(Equal(1))
	g.Expect(serr.Status().Details.Causes[0].Type).To(Equal(v12.CauseTypeFieldValueInvalid))
	g.Expect(serr.Status().Details.Causes[0].Field).To(Equal("spec"))
}

func TestValidateMixedMultipleTransport(t *testing.T) {
	g := NewGomegaWithT(t)
	impl := MODEL
	spec := &SeldonDeploymentSpec{
		Transport: TransportRest,
		Predictors: []PredictorSpec{
			{
				Name: "p1",
				ComponentSpecs: []*SeldonPodSpec{
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
				Graph: PredictiveUnit{
					Name: "classifier",
					Type: &impl,
					Endpoint: &Endpoint{
						Type: GRPC,
					},
				},
			},
		},
	}

	spec.DefaultSeldonDeployment("mydep", "default")
	err := spec.ValidateSeldonDeployment()
	g.Expect(err).NotTo(BeNil())
	serr := err.(*errors.StatusError)
	g.Expect(serr.Status().Code).To(Equal(int32(422)))
	g.Expect(len(serr.Status().Details.Causes)).To(Equal(1))
	g.Expect(serr.Status().Details.Causes[0].Type).To(Equal(v12.CauseTypeFieldValueInvalid))
	g.Expect(serr.Status().Details.Causes[0].Field).To(Equal("spec"))
}

func TestDefaultSingleContainer(t *testing.T) {
	g := NewGomegaWithT(t)
	spec := &SeldonDeploymentSpec{
		Predictors: []PredictorSpec{
			{
				Name: "p1",
				ComponentSpecs: []*SeldonPodSpec{
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
				Graph: PredictiveUnit{
					Name: "classifier",
				},
			},
		},
	}

	spec.DefaultSeldonDeployment("mydep", "default")

	// Test Metric Ports
	metricPort := GetPort(envPredictiveUnitMetricsPortName, spec.Predictors[0].ComponentSpecs[0].Spec.Containers[0].Ports)
	g.Expect(metricPort).NotTo(BeNil())
	g.Expect(metricPort.ContainerPort).To(Equal(constants.FirstMetricsPortNumber))

	// Graph
	pu := GetPredictiveUnit(&spec.Predictors[0].Graph, "classifier")
	g.Expect(pu).ToNot(BeNil())
	g.Expect(pu.Endpoint.ServicePort).To(Equal(constants.FirstHttpPortNumber))
	g.Expect(pu.Endpoint.ServiceHost).To(Equal(constants.DNSLocalHost))

	// Volumes
	volFound := false
	for _, vol := range spec.Predictors[0].ComponentSpecs[0].Spec.Containers[0].VolumeMounts {
		if vol.Name == PODINFO_VOLUME_NAME {
			volFound = true
		}
	}
	g.Expect(volFound).To(BeTrue())
}

func TestMetricsPortAddedTwoContainers(t *testing.T) {
	g := NewGomegaWithT(t)
	spec := &SeldonDeploymentSpec{
		Predictors: []PredictorSpec{
			{
				Name: "p1",
				ComponentSpecs: []*SeldonPodSpec{
					{
						Spec: v1.PodSpec{
							Containers: []v1.Container{
								{
									Image: "seldonio/mock_classifier:1.0",
									Name:  "classifier",
								},
								{
									Image: "seldonio/mock_classifier:1.0",
									Name:  "classifier2",
								},
							},
						},
					},
				},
				Graph: PredictiveUnit{
					Name: "classifier",
					Children: []PredictiveUnit{
						{
							Name: "classifier2",
						},
					},
				},
			},
		},
	}

	//Metrics
	spec.DefaultSeldonDeployment("mydep", "default")
	metricPort := GetPort(envPredictiveUnitMetricsPortName, spec.Predictors[0].ComponentSpecs[0].Spec.Containers[0].Ports)
	g.Expect(metricPort).NotTo(BeNil())
	g.Expect(metricPort.ContainerPort).To(Equal(constants.FirstMetricsPortNumber))

	metricPort = GetPort(envPredictiveUnitMetricsPortName, spec.Predictors[0].ComponentSpecs[0].Spec.Containers[1].Ports)
	g.Expect(metricPort).NotTo(BeNil())
	g.Expect(metricPort.ContainerPort).To(Equal(constants.FirstMetricsPortNumber + 1))

	//Graph
	pu := GetPredictiveUnit(&spec.Predictors[0].Graph, "classifier")
	g.Expect(pu).ToNot(BeNil())
	g.Expect(pu.Endpoint.ServicePort).To(Equal(constants.FirstHttpPortNumber))
	g.Expect(pu.Endpoint.ServiceHost).To(Equal(constants.DNSLocalHost))

	pu = GetPredictiveUnit(&spec.Predictors[0].Graph, "classifier2")
	g.Expect(pu).ToNot(BeNil())
	g.Expect(pu.Endpoint.ServicePort).To(Equal(constants.FirstHttpPortNumber + 1))
	g.Expect(pu.Endpoint.ServiceHost).To(Equal(constants.DNSLocalHost))

	// Volumes
	volFound := false
	for _, vol := range spec.Predictors[0].ComponentSpecs[0].Spec.Containers[0].VolumeMounts {
		if vol.Name == PODINFO_VOLUME_NAME {
			volFound = true
		}
	}
	g.Expect(volFound).To(BeTrue())

	volFound = false
	for _, vol := range spec.Predictors[0].ComponentSpecs[0].Spec.Containers[1].VolumeMounts {
		if vol.Name == PODINFO_VOLUME_NAME {
			volFound = true
		}
	}
	g.Expect(volFound).To(BeTrue())
}

func TestMetricsPortAddedTwoComponentSpecsTwoContainers(t *testing.T) {
	g := NewGomegaWithT(t)
	spec := &SeldonDeploymentSpec{
		Predictors: []PredictorSpec{
			{
				Name: "p1",
				ComponentSpecs: []*SeldonPodSpec{
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
				Graph: PredictiveUnit{
					Name: "classifier",
					Children: []PredictiveUnit{
						{
							Name: "classifier2",
						},
					},
				},
			},
		},
	}

	name := "mydep"
	namespace := "default"
	spec.DefaultSeldonDeployment(name, namespace)

	// Metrics
	metricPort := GetPort(envPredictiveUnitMetricsPortName, spec.Predictors[0].ComponentSpecs[0].Spec.Containers[0].Ports)
	g.Expect(metricPort).NotTo(BeNil())
	g.Expect(metricPort.ContainerPort).To(Equal(constants.FirstMetricsPortNumber))
	metricPort = GetPort(envPredictiveUnitMetricsPortName, spec.Predictors[0].ComponentSpecs[1].Spec.Containers[0].Ports)
	g.Expect(metricPort).NotTo(BeNil())
	g.Expect(metricPort.ContainerPort).To(Equal(constants.FirstMetricsPortNumber + 1))

	//Graph
	pu := GetPredictiveUnit(&spec.Predictors[0].Graph, "classifier")
	g.Expect(pu).ToNot(BeNil())
	g.Expect(pu.Endpoint.ServicePort).To(Equal(constants.FirstHttpPortNumber))
	g.Expect(pu.Endpoint.ServiceHost).To(Equal(constants.DNSLocalHost))

	pu = GetPredictiveUnit(&spec.Predictors[0].Graph, "classifier2")
	g.Expect(pu).ToNot(BeNil())
	g.Expect(pu.Endpoint.ServicePort).To(Equal(constants.FirstHttpPortNumber + 1))
	containerServiceValue := GetContainerServiceName(name, spec.Predictors[0], &spec.Predictors[0].ComponentSpecs[1].Spec.Containers[0])
	dnsName := containerServiceValue + "." + namespace + constants.DNSClusterLocalSuffix
	g.Expect(pu.Endpoint.ServiceHost).To(Equal(dnsName))

	// Volumes
	volFound := false
	for _, vol := range spec.Predictors[0].ComponentSpecs[0].Spec.Containers[0].VolumeMounts {
		if vol.Name == PODINFO_VOLUME_NAME {
			volFound = true
		}
	}
	g.Expect(volFound).To(BeTrue())

	volFound = false
	for _, vol := range spec.Predictors[0].ComponentSpecs[1].Spec.Containers[0].VolumeMounts {
		if vol.Name == PODINFO_VOLUME_NAME {
			volFound = true
		}
	}
	g.Expect(volFound).To(BeTrue())
}

func TestOverrideMetricsPortName(t *testing.T) {
	g := NewGomegaWithT(t)
	scheme := runtime.NewScheme()
	C = fake.NewFakeClientWithScheme(scheme)
	impl := PredictiveUnitImplementation(constants.PrePackedServerTensorflow)
	spec := &SeldonDeploymentSpec{
		Predictors: []PredictorSpec{
			{
				Name: "p1",
				Graph: PredictiveUnit{
					Name:           "classifier",
					Implementation: &impl,
				},
			},
		},
	}

	envPredictiveUnitMetricsPortName = "myMetricsPort"
	spec.DefaultSeldonDeployment("mydep", "default")
	// Metrics
	customMetricsPort := GetPort("myMetricsPort", spec.Predictors[0].ComponentSpecs[0].Spec.Containers[0].Ports)
	g.Expect(customMetricsPort).NotTo(BeNil())
	g.Expect(customMetricsPort.ContainerPort).To(Equal(constants.FirstMetricsPortNumber))

	defaultMetricsPort := GetPort(constants.DefaultMetricsPortName, spec.Predictors[0].ComponentSpecs[0].Spec.Containers[0].Ports)
	g.Expect(defaultMetricsPort).To(BeNil())

	// Graph
	pu := GetPredictiveUnit(&spec.Predictors[0].Graph, "classifier")
	g.Expect(pu).ToNot(BeNil())
	g.Expect(*pu.Type).To(Equal(MODEL))
}

func TestPortUseExisting(t *testing.T) {
	g := NewGomegaWithT(t)
	containerPortMetrics := int32(1234)
	containerPortAPI := int32(5678)
	spec := &SeldonDeploymentSpec{
		Predictors: []PredictorSpec{
			{
				Name: "p1",
				ComponentSpecs: []*SeldonPodSpec{
					{
						Spec: v1.PodSpec{
							Containers: []v1.Container{
								{
									Image: "seldonio/mock_classifier:1.0",
									Name:  "classifier",
									Ports: []v1.ContainerPort{{Name: envPredictiveUnitMetricsPortName, ContainerPort: containerPortMetrics},
										{Name: constants.HttpPortName, ContainerPort: containerPortAPI}},
								},
							},
						},
					},
				},
				Graph: PredictiveUnit{
					Name: "classifier",
				},
			},
		},
	}

	spec.DefaultSeldonDeployment("mydep", "default")
	metricPort := GetPort(envPredictiveUnitMetricsPortName, spec.Predictors[0].ComponentSpecs[0].Spec.Containers[0].Ports)
	g.Expect(metricPort).NotTo(BeNil())
	g.Expect(metricPort.ContainerPort).To(Equal(containerPortMetrics))

	//Graph
	pu := GetPredictiveUnit(&spec.Predictors[0].Graph, "classifier")
	g.Expect(pu).ToNot(BeNil())
	g.Expect(pu.Endpoint.ServicePort).To(Equal(containerPortAPI))
	g.Expect(pu.Endpoint.ServiceHost).To(Equal(constants.DNSLocalHost))
}

func TestMetricsPortAddedToPrepacked(t *testing.T) {
	g := NewGomegaWithT(t)
	err := setupTestConfigMap()
	g.Expect(err).To(BeNil())
	impl := PredictiveUnitImplementation(constants.PrePackedServerTensorflow)
	spec := &SeldonDeploymentSpec{
		Predictors: []PredictorSpec{
			{
				Name: "p1",
				Graph: PredictiveUnit{
					Name:           "classifier",
					Implementation: &impl,
					Endpoint: &Endpoint{
						Type: REST,
					},
				},
			},
		},
	}

	spec.DefaultSeldonDeployment("mydep", "default")
	metricPort := GetPort(envPredictiveUnitMetricsPortName, spec.Predictors[0].ComponentSpecs[0].Spec.Containers[0].Ports)
	g.Expect(metricPort).NotTo(BeNil())
	g.Expect(metricPort.ContainerPort).To(Equal(constants.FirstMetricsPortNumber))

	//Graph
	pu := GetPredictiveUnit(&spec.Predictors[0].Graph, "classifier")
	g.Expect(pu).ToNot(BeNil())
	g.Expect(pu.Endpoint.ServicePort).To(Equal(constants.FirstHttpPortNumber))
	g.Expect(pu.Endpoint.ServiceHost).To(Equal(constants.DNSLocalHost))
	g.Expect(pu.Endpoint.Type).To(Equal(REST))
}

func TestPredictorProtocolGrpc(t *testing.T) {
	g := NewGomegaWithT(t)
	err := setupTestConfigMap()
	g.Expect(err).To(BeNil())
	impl := PredictiveUnitImplementation(constants.PrePackedServerTensorflow)
	spec := &SeldonDeploymentSpec{
		Transport: TransportGrpc,
		Predictors: []PredictorSpec{
			{
				Name: "p1",
				Graph: PredictiveUnit{
					Name:           "classifier",
					Implementation: &impl,
				},
			},
		},
	}

	spec.DefaultSeldonDeployment("mydep", "default")
	metricPort := GetPort(envPredictiveUnitMetricsPortName, spec.Predictors[0].ComponentSpecs[0].Spec.Containers[0].Ports)
	g.Expect(metricPort).NotTo(BeNil())
	g.Expect(metricPort.ContainerPort).To(Equal(constants.FirstMetricsPortNumber))

	//Graph
	pu := GetPredictiveUnit(&spec.Predictors[0].Graph, "classifier")
	g.Expect(pu).ToNot(BeNil())
	g.Expect(pu.Endpoint.ServicePort).To(Equal(constants.FirstGrpcPortNumber))
	g.Expect(pu.Endpoint.ServiceHost).To(Equal(constants.DNSLocalHost))
}

func TestPrepackedWithExistingContainer(t *testing.T) {
	g := NewGomegaWithT(t)
	err := setupTestConfigMap()
	g.Expect(err).To(BeNil())
	impl := PredictiveUnitImplementation(constants.PrePackedServerTensorflow)
	spec := &SeldonDeploymentSpec{
		Transport: TransportGrpc,
		Predictors: []PredictorSpec{
			{
				Name: "p1",
				ComponentSpecs: []*SeldonPodSpec{
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
				Graph: PredictiveUnit{
					Name:           "classifier",
					Implementation: &impl,
				},
			},
		},
	}

	spec.DefaultSeldonDeployment("mydep", "default")
	metricPort := GetPort(envPredictiveUnitMetricsPortName, spec.Predictors[0].ComponentSpecs[0].Spec.Containers[0].Ports)
	g.Expect(metricPort).NotTo(BeNil())
	g.Expect(metricPort.ContainerPort).To(Equal(constants.FirstMetricsPortNumber))

	// image set from configMap
	g.Expect(spec.Predictors[0].ComponentSpecs[0].Spec.Containers[0].Image).To(Equal("seldonio/tfserving-proxy:1.3.0-dev"))

	//Graph
	pu := GetPredictiveUnit(&spec.Predictors[0].Graph, "classifier")
	g.Expect(pu).ToNot(BeNil())
	g.Expect(pu.Endpoint.ServicePort).To(Equal(constants.FirstGrpcPortNumber))
	g.Expect(pu.Endpoint.ServiceHost).To(Equal(constants.DNSLocalHost))
}

func TestPrepackedWithCustom(t *testing.T) {
	g := NewGomegaWithT(t)
	err := setupTestConfigMap()
	g.Expect(err).To(BeNil())
	impl := PredictiveUnitImplementation("CUSTOM_SERVER")
	spec := &SeldonDeploymentSpec{
		Transport: TransportGrpc,
		Predictors: []PredictorSpec{
			{
				Name: "p1",
				Graph: PredictiveUnit{
					Name:           "classifier",
					Implementation: &impl,
				},
			},
		},
	}

	spec.DefaultSeldonDeployment("mydep", "default")
	metricPort := GetPort(envPredictiveUnitMetricsPortName, spec.Predictors[0].ComponentSpecs[0].Spec.Containers[0].Ports)
	g.Expect(metricPort).NotTo(BeNil())
	g.Expect(metricPort.ContainerPort).To(Equal(constants.FirstMetricsPortNumber))
	g.Expect(metricPort.Name).To(Equal(envPredictiveUnitMetricsPortName))

	// image set from configMap
	g.Expect(spec.Predictors[0].ComponentSpecs[0].Spec.Containers[0].Image).To(Equal("custom:0.1"))

	//Graph
	pu := GetPredictiveUnit(&spec.Predictors[0].Graph, "classifier")
	g.Expect(pu).ToNot(BeNil())
	g.Expect(pu.Endpoint.ServicePort).To(Equal(constants.FirstGrpcPortNumber))
	g.Expect(pu.Endpoint.ServiceHost).To(Equal(constants.DNSLocalHost))
}

func TestPrepackedWithExistingContainerAndImage(t *testing.T) {
	g := NewGomegaWithT(t)
	err := setupTestConfigMap()
	g.Expect(err).To(BeNil())
	impl := PredictiveUnitImplementation(constants.PrePackedServerTensorflow)
	image := "myimage:0.1"
	spec := &SeldonDeploymentSpec{
		Transport: TransportGrpc,
		Predictors: []PredictorSpec{
			{
				Name: "p1",
				ComponentSpecs: []*SeldonPodSpec{
					{
						Spec: v1.PodSpec{
							Containers: []v1.Container{
								{
									Name:  "classifier",
									Image: image,
								},
							},
						},
					},
				},
				Graph: PredictiveUnit{
					Name:           "classifier",
					Implementation: &impl,
				},
			},
		},
	}

	spec.DefaultSeldonDeployment("mydep", "default")
	metricPort := GetPort(envPredictiveUnitMetricsPortName, spec.Predictors[0].ComponentSpecs[0].Spec.Containers[0].Ports)
	g.Expect(metricPort).NotTo(BeNil())
	g.Expect(metricPort.ContainerPort).To(Equal(constants.FirstMetricsPortNumber))
	g.Expect(metricPort.Name).To(Equal(envPredictiveUnitMetricsPortName))

	// image set from configMap
	g.Expect(spec.Predictors[0].ComponentSpecs[0].Spec.Containers[0].Image).To(Equal(image))

	//Graph
	pu := GetPredictiveUnit(&spec.Predictors[0].Graph, "classifier")
	g.Expect(pu).ToNot(BeNil())
	g.Expect(pu.Endpoint.ServicePort).To(Equal(constants.FirstGrpcPortNumber))
	g.Expect(pu.Endpoint.ServiceHost).To(Equal(constants.DNSLocalHost))
}

func TestMetricsPortAddedToTwoPrepacked(t *testing.T) {
	g := NewGomegaWithT(t)
	err := setupTestConfigMap()
	g.Expect(err).To(BeNil())
	impl := PredictiveUnitImplementation(constants.PrePackedServerTensorflow)
	spec := &SeldonDeploymentSpec{
		Predictors: []PredictorSpec{
			{
				Name: "p1",
				Graph: PredictiveUnit{
					Name:           "classifier",
					Implementation: &impl,
					Children: []PredictiveUnit{
						{
							Name:           "classifier2",
							Implementation: &impl,
						},
					},
				},
			},
		},
	}

	spec.DefaultSeldonDeployment("mydep", "default")
	metricPort := GetPort(envPredictiveUnitMetricsPortName, spec.Predictors[0].ComponentSpecs[0].Spec.Containers[0].Ports)
	g.Expect(metricPort).NotTo(BeNil())
	g.Expect(metricPort.ContainerPort).To(Equal(constants.FirstMetricsPortNumber))

	metricPort = GetPort(envPredictiveUnitMetricsPortName, spec.Predictors[0].ComponentSpecs[0].Spec.Containers[1].Ports)
	g.Expect(metricPort).NotTo(BeNil())
	g.Expect(metricPort.ContainerPort).To(Equal(constants.FirstMetricsPortNumber + 1))

	//Graph
	pu := GetPredictiveUnit(&spec.Predictors[0].Graph, "classifier")
	g.Expect(pu).ToNot(BeNil())
	g.Expect(pu.Endpoint.ServicePort).To(Equal(constants.FirstHttpPortNumber))
	g.Expect(pu.Endpoint.ServiceHost).To(Equal(constants.DNSLocalHost))

	pu = GetPredictiveUnit(&spec.Predictors[0].Graph, "classifier2")
	g.Expect(pu).ToNot(BeNil())
	g.Expect(pu.Endpoint.ServicePort).To(Equal(constants.FirstHttpPortNumber + 1))
	g.Expect(pu.Endpoint.ServiceHost).To(Equal(constants.DNSLocalHost))
}

func TestDefaultPrepackagedServerType(t *testing.T) {
	g := NewGomegaWithT(t)
	err := setupTestConfigMap()
	g.Expect(err).To(BeNil())
	impl := PredictiveUnitImplementation(constants.PrePackedServerTensorflow)
	spec := &SeldonDeploymentSpec{
		Predictors: []PredictorSpec{
			{
				Name: "p1",
				Graph: PredictiveUnit{
					Name:           "classifier",
					Implementation: &impl,
				},
			},
		},
	}

	spec.DefaultSeldonDeployment("mydep", "default")

	// Graph
	pu := GetPredictiveUnit(&spec.Predictors[0].Graph, "classifier")
	g.Expect(pu).ToNot(BeNil())
	g.Expect(*pu.Type).To(Equal(MODEL))
}

func TestValidateSingleModel(t *testing.T) {
	g := NewGomegaWithT(t)
	spec := &SeldonDeploymentSpec{
		Predictors: []PredictorSpec{
			{
				Name: "p1",
				ComponentSpecs: []*SeldonPodSpec{
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
				Graph: PredictiveUnit{
					Name: "classifier",
				},
			},
		},
	}

	spec.DefaultSeldonDeployment("mydep", "default")
	err := spec.ValidateSeldonDeployment()
	g.Expect(err).To(BeNil())
}

func TestValidateSingleModelNoName(t *testing.T) {
	g := NewGomegaWithT(t)
	spec := &SeldonDeploymentSpec{
		Predictors: []PredictorSpec{
			{
				Name: "p1",
				ComponentSpecs: []*SeldonPodSpec{
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
				Graph: PredictiveUnit{},
			},
		},
	}

	spec.DefaultSeldonDeployment("mydep", "default")
	err := spec.ValidateSeldonDeployment()
	g.Expect(err).ToNot(BeNil())
	serr := err.(*errors.StatusError)
	g.Expect(serr.Status().Code).To(Equal(int32(422)))
	g.Expect(serr.Status().Code).To(Equal(int32(422)))
	g.Expect(serr.Status().Details.Causes[0].Type).To(Equal(v12.CauseTypeFieldValueInvalid))
	g.Expect(serr.Status().Details.Causes[0].Field).To(Equal("spec.predictors[0].graph"))
}

func TestValidateNoEngineMultiGraph(t *testing.T) {
	g := NewGomegaWithT(t)
	spec := &SeldonDeploymentSpec{
		Predictors: []PredictorSpec{
			{
				Name:        "p1",
				Annotations: map[string]string{ANNOTATION_NO_ENGINE: "true"},
				ComponentSpecs: []*SeldonPodSpec{
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
				Graph: PredictiveUnit{
					Name: "classifier",
					Children: []PredictiveUnit{
						{
							Name: "classifier2",
						},
					},
				},
			},
		},
	}

	spec.DefaultSeldonDeployment("mydep", "default")
	err := spec.ValidateSeldonDeployment()
	g.Expect(err).ToNot(BeNil())
	serr := err.(*errors.StatusError)
	g.Expect(serr.Status().Code).To(Equal(int32(422)))
	g.Expect(serr.Status().Code).To(Equal(int32(422)))
	g.Expect(serr.Status().Details.Causes[0].Type).To(Equal(v12.CauseTypeFieldValueInvalid))
	g.Expect(serr.Status().Details.Causes[0].Field).To(Equal("spec.predictors[0]"))
}

func TestValidateDuplPredictorName(t *testing.T) {
	g := NewGomegaWithT(t)
	spec := &SeldonDeploymentSpec{
		Predictors: []PredictorSpec{
			{
				Name: "p1",
				ComponentSpecs: []*SeldonPodSpec{
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
				Graph: PredictiveUnit{
					Name: "classifier",
				},
				Traffic: 50,
			},
			{
				Name: "p1",
				ComponentSpecs: []*SeldonPodSpec{
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
				Graph: PredictiveUnit{
					Name: "classifier",
				},
				Traffic: 50,
			},
		},
	}

	spec.DefaultSeldonDeployment("mydep", "default")
	err := spec.ValidateSeldonDeployment()
	g.Expect(err).ToNot(BeNil())
	serr := err.(*errors.StatusError)
	g.Expect(serr.Status().Code).To(Equal(int32(422)))
	g.Expect(len(serr.Status().Details.Causes)).To(Equal(1))
	g.Expect(serr.Status().Details.Causes[0].Type).To(Equal(v12.CauseTypeFieldValueInvalid))
	g.Expect(serr.Status().Details.Causes[0].Field).To(Equal("spec.predictors[1]"))
}

func TestValidateTrafficSum(t *testing.T) {
	g := NewGomegaWithT(t)
	spec := &SeldonDeploymentSpec{
		Predictors: []PredictorSpec{
			{
				Name: "p1",
				ComponentSpecs: []*SeldonPodSpec{
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
				Graph: PredictiveUnit{
					Name: "classifier",
				},
			},
			{
				Name: "p2",
				ComponentSpecs: []*SeldonPodSpec{
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
				Graph: PredictiveUnit{
					Name: "classifier",
				},
			},
		},
	}

	spec.DefaultSeldonDeployment("mydep", "default")
	err := spec.ValidateSeldonDeployment()
	g.Expect(err).ToNot(BeNil())
	serr := err.(*errors.StatusError)
	g.Expect(serr.Status().Code).To(Equal(int32(422)))
	g.Expect(len(serr.Status().Details.Causes)).To(Equal(1))
	g.Expect(serr.Status().Details.Causes[0].Type).To(Equal(v12.CauseTypeFieldValueInvalid))
	g.Expect(serr.Status().Details.Causes[0].Field).To(Equal("spec"))
}

func TestDefaultABTest(t *testing.T) {
	g := NewGomegaWithT(t)
	impl := RANDOM_ABTEST
	spec := &SeldonDeploymentSpec{
		Predictors: []PredictorSpec{
			{
				Name: "p1",
				ComponentSpecs: []*SeldonPodSpec{
					{
						Spec: v1.PodSpec{
							Containers: []v1.Container{
								{
									Image: "seldonio/mock_classifier:1.0",
									Name:  "classifier1",
								},
								{
									Image: "seldonio/mock_classifier:1.0",
									Name:  "classifier2",
								},
							},
						},
					},
				},
				Graph: PredictiveUnit{
					Implementation: &impl,
					Children: []PredictiveUnit{
						{
							Name: "classifier1",
						},
						{
							Name: "classifier2",
						},
					},
				},
			},
		},
	}

	spec.DefaultSeldonDeployment("mydep", "default")
	err := spec.ValidateSeldonDeployment()
	g.Expect(err).To(BeNil())
	graph := spec.Predictors[0].Graph
	g.Expect(graph.Type).To(BeNil())
	g.Expect(*graph.Implementation).To(Equal(RANDOM_ABTEST))
	g.Expect(*graph.Children[0].Type).To(Equal(MODEL))
	g.Expect(*graph.Children[1].Type).To(Equal(MODEL))
}

func TestValidateTensorflowProtocolNormalPrepackaged(t *testing.T) {
	g := NewGomegaWithT(t)
	err := setupTestConfigMap()
	g.Expect(err).To(BeNil())
	impl := PredictiveUnitImplementation(constants.PrePackedServerSklearn)
	spec := &SeldonDeploymentSpec{
		Protocol: ProtocolTensorflow,
		Predictors: []PredictorSpec{
			{
				Name: "p1",
				Graph: PredictiveUnit{
					Name:           "classifier",
					Implementation: &impl,
					ModelURI:       "s3://mybucket/model",
				},
			},
		},
	}

	spec.DefaultSeldonDeployment("mydep", "default")
	err = spec.ValidateSeldonDeployment()
	g.Expect(err).ToNot(BeNil())
	serr := err.(*errors.StatusError)
	g.Expect(serr.Status().Code).To(Equal(int32(422)))
	g.Expect(len(serr.Status().Details.Causes)).To(Equal(1))
	g.Expect(serr.Status().Details.Causes[0].Type).To(Equal(v12.CauseTypeFieldValueInvalid))
	g.Expect(serr.Status().Details.Causes[0].Field).To(Equal("spec.predictors[0].graph"))
}

func TestValidateTensorflowProtocolNormal(t *testing.T) {
	g := NewGomegaWithT(t)
	err := setupTestConfigMap()
	g.Expect(err).To(BeNil())
	impl := PredictiveUnitImplementation(constants.PrePackedServerTensorflow)
	spec := &SeldonDeploymentSpec{
		Protocol: ProtocolTensorflow,
		Predictors: []PredictorSpec{
			{
				Name: "p1",
				Graph: PredictiveUnit{
					Name:           "classifier",
					Implementation: &impl,
					ModelURI:       "s3://mybucket/model",
				},
			},
		},
	}

	spec.DefaultSeldonDeployment("mydep", "default")
	err = spec.ValidateSeldonDeployment()
	g.Expect(err).To(BeNil())
}

func TestValidateKFservingProtocolNormal(t *testing.T) {
	g := NewGomegaWithT(t)
	err := setupTestConfigMap()
	g.Expect(err).To(BeNil())
	impl := PredictiveUnitImplementation(constants.PrePackedServerTensorflow)
	spec := &SeldonDeploymentSpec{
		Protocol: ProtocolKfserving,
		Predictors: []PredictorSpec{
			{
				Name: "p1",
				Graph: PredictiveUnit{
					Name:           "classifier",
					Implementation: &impl,
					ModelURI:       "s3://mybucket/model",
				},
			},
		},
	}

	spec.DefaultSeldonDeployment("mydep", "default")
	err = spec.ValidateSeldonDeployment()
	g.Expect(err).To(BeNil())
}
func TestValidateV2ProtocolNormal(t *testing.T) {
	g := NewGomegaWithT(t)
	err := setupTestConfigMap()
	g.Expect(err).To(BeNil())
	impl := PredictiveUnitImplementation(constants.PrePackedServerTensorflow)
	spec := &SeldonDeploymentSpec{
		Protocol: ProtocolV2,
		Predictors: []PredictorSpec{
			{
				Name: "p1",
				Graph: PredictiveUnit{
					Name:           "classifier",
					Implementation: &impl,
					ModelURI:       "s3://mybucket/model",
				},
			},
		},
	}

	spec.DefaultSeldonDeployment("mydep", "default")
	err = spec.ValidateSeldonDeployment()
	g.Expect(err).To(BeNil())
}

func TestPredictorNoGraph(t *testing.T) {
	g := NewGomegaWithT(t)
	scheme := runtime.NewScheme()
	C = fake.NewFakeClientWithScheme(scheme)
	spec := &SeldonDeploymentSpec{
		Transport: TransportGrpc,
		Predictors: []PredictorSpec{
			{
				Name: "p1",
			},
		},
	}

	spec.DefaultSeldonDeployment("mydep", "default")
	err := spec.ValidateSeldonDeployment()
	g.Expect(err).ToNot(BeNil())
}

func TestShadowPredictor(t *testing.T) {
	g := NewGomegaWithT(t)
	scheme := runtime.NewScheme()
	C = fake.NewFakeClientWithScheme(scheme)
	err := setupTestConfigMap()
	g.Expect(err).To(BeNil())
	impl := PredictiveUnitImplementation(constants.PrePackedServerTensorflow)
	spec := &SeldonDeploymentSpec{
		Transport: TransportGrpc,
		Predictors: []PredictorSpec{
			{
				Name:   "p1",
				Shadow: true,
				Graph: PredictiveUnit{
					Name:           "classifier",
					Implementation: &impl,
					ModelURI:       "s3://mybucket/model",
				},
			},
		},
	}

	spec.DefaultSeldonDeployment("mydep", "default")
	err = spec.ValidateSeldonDeployment()
	g.Expect(err).ToNot(BeNil())
}

func TestNoPredictors(t *testing.T) {
	g := NewGomegaWithT(t)
	scheme := runtime.NewScheme()
	C = fake.NewFakeClientWithScheme(scheme)
	spec := &SeldonDeploymentSpec{
		Transport:  TransportGrpc,
		Predictors: []PredictorSpec{},
	}

	spec.DefaultSeldonDeployment("mydep", "default")
	err := spec.ValidateSeldonDeployment()
	g.Expect(err).ToNot(BeNil())
}

func TestValidateTwoShadows(t *testing.T) {
	g := NewGomegaWithT(t)
	err := setupTestConfigMap()
	g.Expect(err).To(BeNil())
	impl := PredictiveUnitImplementation(constants.PrePackedServerTensorflow)
	spec := &SeldonDeploymentSpec{
		Protocol: ProtocolTensorflow,
		Predictors: []PredictorSpec{
			{
				Name: "p1",
				Graph: PredictiveUnit{
					Name:           "classifier",
					Implementation: &impl,
					ModelURI:       "s3://mybucket/model",
				},
			},
			{
				Name: "p1",
				Graph: PredictiveUnit{
					Name:           "classifier",
					Implementation: &impl,
					ModelURI:       "s3://mybucket/model",
				},
				Shadow: true,
			},
			{
				Name: "p1",
				Graph: PredictiveUnit{
					Name:           "classifier",
					Implementation: &impl,
					ModelURI:       "s3://mybucket/model",
				},
				Shadow: true,
			},
		},
	}

	spec.DefaultSeldonDeployment("mydep", "default")
	err = spec.ValidateSeldonDeployment()
	g.Expect(err).ToNot(BeNil())
}

func TestValidateShadowTraffic(t *testing.T) {
	g := NewGomegaWithT(t)
	err := setupTestConfigMap()
	g.Expect(err).To(BeNil())
	impl := PredictiveUnitImplementation(constants.PrePackedServerTensorflow)
	spec := &SeldonDeploymentSpec{
		Protocol: ProtocolTensorflow,
		Predictors: []PredictorSpec{
			{
				Name: "p1",
				Graph: PredictiveUnit{
					Name:           "classifier",
					Implementation: &impl,
					ModelURI:       "s3://mybucket/model",
				},
				Traffic: 100,
			},
			{
				Name: "p2",
				Graph: PredictiveUnit{
					Name:           "classifier",
					Implementation: &impl,
					ModelURI:       "s3://mybucket/model",
				},
				Shadow:  true,
				Traffic: 101,
			},
		},
	}
	spec.DefaultSeldonDeployment("mydep", "default")
	err = spec.ValidateSeldonDeployment()
	g.Expect(err).ToNot(BeNil())
}
