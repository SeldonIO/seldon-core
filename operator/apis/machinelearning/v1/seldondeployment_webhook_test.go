package v1

import (
	. "github.com/onsi/gomega"
	"github.com/seldonio/seldon-core/operator/constants"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

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
				Graph: &PredictiveUnit{
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
	g.Expect(len(serr.Status().Details.Causes)).To(Equal(2))
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
				Graph: &PredictiveUnit{
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
	g.Expect(len(serr.Status().Details.Causes)).To(Equal(2))
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
				Graph: &PredictiveUnit{
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
				Graph: &PredictiveUnit{
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
				Graph: &PredictiveUnit{
					Name: "classifier",
				},
			},
		},
	}

	spec.DefaultSeldonDeployment("mydep", "default")

	// Test Metric Ports
	metricPort := GetPort(constants.MetricsPortName, spec.Predictors[0].ComponentSpecs[0].Spec.Containers[0].Ports)
	g.Expect(metricPort).NotTo(BeNil())
	g.Expect(metricPort.ContainerPort).To(Equal(constants.FirstMetricsPortNumber))
	g.Expect(metricPort.Name).To(Equal(constants.MetricsPortName))

	// Graph
	pu := GetPredictiveUnit(spec.Predictors[0].Graph, "classifier")
	g.Expect(pu).ToNot(BeNil())
	g.Expect(pu.Endpoint.ServicePort).To(Equal(constants.FirstPortNumber))
	g.Expect(pu.Endpoint.ServiceHost).To(Equal(constants.DNSLocalHost))
	g.Expect(pu.Endpoint.Type).To(Equal(REST))

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
				Graph: &PredictiveUnit{
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
	metricPort := GetPort(constants.MetricsPortName, spec.Predictors[0].ComponentSpecs[0].Spec.Containers[0].Ports)
	g.Expect(metricPort).NotTo(BeNil())
	g.Expect(metricPort.ContainerPort).To(Equal(constants.FirstMetricsPortNumber))
	g.Expect(metricPort.Name).To(Equal(constants.MetricsPortName))

	metricPort = GetPort(constants.MetricsPortName, spec.Predictors[0].ComponentSpecs[0].Spec.Containers[1].Ports)
	g.Expect(metricPort).NotTo(BeNil())
	g.Expect(metricPort.ContainerPort).To(Equal(constants.FirstMetricsPortNumber + 1))
	g.Expect(metricPort.Name).To(Equal(constants.MetricsPortName))

	//Graph
	pu := GetPredictiveUnit(spec.Predictors[0].Graph, "classifier")
	g.Expect(pu).ToNot(BeNil())
	g.Expect(pu.Endpoint.ServicePort).To(Equal(constants.FirstPortNumber))
	g.Expect(pu.Endpoint.ServiceHost).To(Equal(constants.DNSLocalHost))
	g.Expect(pu.Endpoint.Type).To(Equal(REST))

	pu = GetPredictiveUnit(spec.Predictors[0].Graph, "classifier2")
	g.Expect(pu).ToNot(BeNil())
	g.Expect(pu.Endpoint.ServicePort).To(Equal(constants.FirstPortNumber + 1))
	g.Expect(pu.Endpoint.ServiceHost).To(Equal(constants.DNSLocalHost))
	g.Expect(pu.Endpoint.Type).To(Equal(REST))

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
				Graph: &PredictiveUnit{
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
	metricPort := GetPort(constants.MetricsPortName, spec.Predictors[0].ComponentSpecs[0].Spec.Containers[0].Ports)
	g.Expect(metricPort).NotTo(BeNil())
	g.Expect(metricPort.ContainerPort).To(Equal(constants.FirstMetricsPortNumber))
	g.Expect(metricPort.Name).To(Equal(constants.MetricsPortName))
	metricPort = GetPort(constants.MetricsPortName, spec.Predictors[0].ComponentSpecs[1].Spec.Containers[0].Ports)
	g.Expect(metricPort).NotTo(BeNil())
	g.Expect(metricPort.ContainerPort).To(Equal(constants.FirstMetricsPortNumber + 1))
	g.Expect(metricPort.Name).To(Equal(constants.MetricsPortName))

	//Graph
	pu := GetPredictiveUnit(spec.Predictors[0].Graph, "classifier")
	g.Expect(pu).ToNot(BeNil())
	g.Expect(pu.Endpoint.ServicePort).To(Equal(constants.FirstPortNumber))
	g.Expect(pu.Endpoint.ServiceHost).To(Equal(constants.DNSLocalHost))
	g.Expect(pu.Endpoint.Type).To(Equal(REST))

	pu = GetPredictiveUnit(spec.Predictors[0].Graph, "classifier2")
	g.Expect(pu).ToNot(BeNil())
	g.Expect(pu.Endpoint.ServicePort).To(Equal(constants.FirstPortNumber + 1))
	containerServiceValue := GetContainerServiceName(name, spec.Predictors[0], &spec.Predictors[0].ComponentSpecs[1].Spec.Containers[0])
	dnsName := containerServiceValue + "." + namespace + constants.DNSClusterLocalSuffix
	g.Expect(pu.Endpoint.ServiceHost).To(Equal(dnsName))
	g.Expect(pu.Endpoint.Type).To(Equal(REST))

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
									Ports: []v1.ContainerPort{{Name: constants.MetricsPortName, ContainerPort: containerPortMetrics},
										{Name: constants.HttpPortName, ContainerPort: containerPortAPI}},
								},
							},
						},
					},
				},
				Graph: &PredictiveUnit{
					Name: "classifier",
				},
			},
		},
	}

	spec.DefaultSeldonDeployment("mydep", "default")
	metricPort := GetPort(constants.MetricsPortName, spec.Predictors[0].ComponentSpecs[0].Spec.Containers[0].Ports)
	g.Expect(metricPort).NotTo(BeNil())
	g.Expect(metricPort.ContainerPort).To(Equal(containerPortMetrics))
	g.Expect(metricPort.Name).To(Equal(constants.MetricsPortName))

	//Graph
	pu := GetPredictiveUnit(spec.Predictors[0].Graph, "classifier")
	g.Expect(pu).ToNot(BeNil())
	g.Expect(pu.Endpoint.ServicePort).To(Equal(containerPortAPI))
	g.Expect(pu.Endpoint.ServiceHost).To(Equal(constants.DNSLocalHost))
	g.Expect(pu.Endpoint.Type).To(Equal(REST))
}

func TestMetricsPortAddedToPrepacked(t *testing.T) {
	g := NewGomegaWithT(t)
	scheme := runtime.NewScheme()
	C = fake.NewFakeClientWithScheme(scheme)
	impl := PredictiveUnitImplementation(constants.PrePackedServerTensorflow)
	spec := &SeldonDeploymentSpec{
		Predictors: []PredictorSpec{
			{
				Name: "p1",
				Graph: &PredictiveUnit{
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
	metricPort := GetPort(constants.MetricsPortName, spec.Predictors[0].ComponentSpecs[0].Spec.Containers[0].Ports)
	g.Expect(metricPort).NotTo(BeNil())
	g.Expect(metricPort.ContainerPort).To(Equal(constants.FirstMetricsPortNumber))
	g.Expect(metricPort.Name).To(Equal(constants.MetricsPortName))

	//Graph
	pu := GetPredictiveUnit(spec.Predictors[0].Graph, "classifier")
	g.Expect(pu).ToNot(BeNil())
	g.Expect(pu.Endpoint.ServicePort).To(Equal(constants.FirstPortNumber))
	g.Expect(pu.Endpoint.ServiceHost).To(Equal(constants.DNSLocalHost))
	g.Expect(pu.Endpoint.Type).To(Equal(REST))
}

func TestPredictorProtocolGrpc(t *testing.T) {
	g := NewGomegaWithT(t)
	scheme := runtime.NewScheme()
	C = fake.NewFakeClientWithScheme(scheme)
	impl := PredictiveUnitImplementation(constants.PrePackedServerTensorflow)
	spec := &SeldonDeploymentSpec{
		Predictors: []PredictorSpec{
			{
				Transport: TransportGrpc,
				Name:      "p1",
				Graph: &PredictiveUnit{
					Name:           "classifier",
					Implementation: &impl,
				},
			},
		},
	}

	spec.DefaultSeldonDeployment("mydep", "default")
	metricPort := GetPort(constants.MetricsPortName, spec.Predictors[0].ComponentSpecs[0].Spec.Containers[0].Ports)
	g.Expect(metricPort).NotTo(BeNil())
	g.Expect(metricPort.ContainerPort).To(Equal(constants.FirstMetricsPortNumber))
	g.Expect(metricPort.Name).To(Equal(constants.MetricsPortName))

	//Graph
	pu := GetPredictiveUnit(spec.Predictors[0].Graph, "classifier")
	g.Expect(pu).ToNot(BeNil())
	g.Expect(pu.Endpoint.ServicePort).To(Equal(constants.FirstPortNumber))
	g.Expect(pu.Endpoint.ServiceHost).To(Equal(constants.DNSLocalHost))
	g.Expect(pu.Endpoint.Type).To(Equal(GRPC))
}

func TestPrepackedWithExistingContainer(t *testing.T) {
	g := NewGomegaWithT(t)
	scheme := runtime.NewScheme()
	C = fake.NewFakeClientWithScheme(scheme)
	impl := PredictiveUnitImplementation(constants.PrePackedServerTensorflow)
	spec := &SeldonDeploymentSpec{
		Predictors: []PredictorSpec{
			{
				Transport: TransportGrpc,
				Name:      "p1",
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
				Graph: &PredictiveUnit{
					Name:           "classifier",
					Implementation: &impl,
				},
			},
		},
	}

	spec.DefaultSeldonDeployment("mydep", "default")
	metricPort := GetPort(constants.MetricsPortName, spec.Predictors[0].ComponentSpecs[0].Spec.Containers[0].Ports)
	g.Expect(metricPort).NotTo(BeNil())
	g.Expect(metricPort.ContainerPort).To(Equal(constants.FirstMetricsPortNumber))
	g.Expect(metricPort.Name).To(Equal(constants.MetricsPortName))

	// empty image name as no configmap - but is set
	g.Expect(spec.Predictors[0].ComponentSpecs[0].Spec.Containers[0].Image).To(Equal(":"))

	//Graph
	pu := GetPredictiveUnit(spec.Predictors[0].Graph, "classifier")
	g.Expect(pu).ToNot(BeNil())
	g.Expect(pu.Endpoint.ServicePort).To(Equal(constants.FirstPortNumber))
	g.Expect(pu.Endpoint.ServiceHost).To(Equal(constants.DNSLocalHost))
	g.Expect(pu.Endpoint.Type).To(Equal(GRPC))
}

func TestMetricsPortAddedToTwoPrepacked(t *testing.T) {
	g := NewGomegaWithT(t)
	scheme := runtime.NewScheme()
	C = fake.NewFakeClientWithScheme(scheme)
	impl := PredictiveUnitImplementation(constants.PrePackedServerTensorflow)
	spec := &SeldonDeploymentSpec{
		Predictors: []PredictorSpec{
			{
				Name: "p1",
				Graph: &PredictiveUnit{
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
	metricPort := GetPort(constants.MetricsPortName, spec.Predictors[0].ComponentSpecs[0].Spec.Containers[0].Ports)
	g.Expect(metricPort).NotTo(BeNil())
	g.Expect(metricPort.ContainerPort).To(Equal(constants.FirstMetricsPortNumber))
	g.Expect(metricPort.Name).To(Equal(constants.MetricsPortName))

	metricPort = GetPort(constants.MetricsPortName, spec.Predictors[0].ComponentSpecs[0].Spec.Containers[1].Ports)
	g.Expect(metricPort).NotTo(BeNil())
	g.Expect(metricPort.ContainerPort).To(Equal(constants.FirstMetricsPortNumber + 1))
	g.Expect(metricPort.Name).To(Equal(constants.MetricsPortName))

	//Graph
	pu := GetPredictiveUnit(spec.Predictors[0].Graph, "classifier")
	g.Expect(pu).ToNot(BeNil())
	g.Expect(pu.Endpoint.ServicePort).To(Equal(constants.FirstPortNumber))
	g.Expect(pu.Endpoint.ServiceHost).To(Equal(constants.DNSLocalHost))
	g.Expect(pu.Endpoint.Type).To(Equal(REST))

	pu = GetPredictiveUnit(spec.Predictors[0].Graph, "classifier2")
	g.Expect(pu).ToNot(BeNil())
	g.Expect(pu.Endpoint.ServicePort).To(Equal(constants.FirstPortNumber + 1))
	g.Expect(pu.Endpoint.ServiceHost).To(Equal(constants.DNSLocalHost))
	g.Expect(pu.Endpoint.Type).To(Equal(REST))
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
				Graph: &PredictiveUnit{
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
				Graph: &PredictiveUnit{},
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
				Graph: &PredictiveUnit{
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
				Graph: &PredictiveUnit{
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
				Graph: &PredictiveUnit{
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
				Graph: &PredictiveUnit{
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
				Graph: &PredictiveUnit{
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
