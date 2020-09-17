package v1

import (
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"testing"
)

func TestImageSetNormal(t *testing.T) {
	g := NewGomegaWithT(t)
	err := setupTestConfigMap()
	g.Expect(err).To(BeNil())

	scenarios := map[string]struct {
		pu               *PredictiveUnit
		config           *PredictorServerConfig
		desiredImageName string
	}{
		"GRPC image with version": {
			pu: &PredictiveUnit{Endpoint: &Endpoint{Type: GRPC}},
			config: &PredictorServerConfig{
				GrpcConfig: PredictorImageConfig{ContainerImage: "a", DefaultImageVersion: "1"},
			},
			desiredImageName: "a:1",
		},
		"GRPC image with no version": {
			pu: &PredictiveUnit{Endpoint: &Endpoint{Type: GRPC}},
			config: &PredictorServerConfig{
				GrpcConfig: PredictorImageConfig{ContainerImage: "a"},
			},
			desiredImageName: "a",
		},
		"REST image with version": {
			pu: &PredictiveUnit{Endpoint: &Endpoint{Type: REST}},
			config: &PredictorServerConfig{
				RestConfig: PredictorImageConfig{ContainerImage: "a", DefaultImageVersion: "1"},
			},
			desiredImageName: "a:1",
		},
		"REST image with no version": {
			pu: &PredictiveUnit{Endpoint: &Endpoint{Type: REST}},
			config: &PredictorServerConfig{
				RestConfig: PredictorImageConfig{ContainerImage: "a"},
			},
			desiredImageName: "a",
		},
	}

	for name, scenario := range scenarios {
		t.Logf("Scenario: %s", name)
		con := &corev1.Container{}
		SetImageNameForPrepackContainer(scenario.pu, con, scenario.config)
		g.Expect(con.Image).To(Equal(scenario.desiredImageName))
	}
}

func TestGetPredictorConfig(t *testing.T) {
	g := NewGomegaWithT(t)
	err := setupTestConfigMap()
	g.Expect(err).To(BeNil())
	serverConfigs, err := getPredictorServerConfigs()
	g.Expect(err).To(BeNil())
	scenarios := map[string]struct {
		serverName      string
		relatedImageMap map[string]PredictorServerConfig
		desiredConfig   PredictorServerConfig
	}{

		"related image sklearn": {
			serverName:      PrepackSklearnName,
			relatedImageMap: map[string]PredictorServerConfig{PrepackSklearnName: {RestConfig: PredictorImageConfig{ContainerImage: "a"}}},
			desiredConfig:   PredictorServerConfig{RestConfig: PredictorImageConfig{ContainerImage: "a"}},
		},
		"default image sklearn": {
			serverName:      PrepackSklearnName,
			relatedImageMap: map[string]PredictorServerConfig{},
			desiredConfig:   serverConfigs[PrepackSklearnName],
		},
	}
	for name, scenario := range scenarios {
		t.Logf("Scenario: %s", name)
		config := getPrepackServerConfigWithRelated(scenario.serverName, scenario.relatedImageMap)
		g.Expect(*config).To(Equal(scenario.desiredConfig))
	}
}

func TestPredictorServerConfigPrepackImageConfig(t *testing.T) {
	g := NewGomegaWithT(t)

	tests := []struct {
		config       *PredictorServerConfig
		protocol     Protocol
		endpointType EndpointType
		expected     *PredictorImageConfig
	}{
		{
			config: &PredictorServerConfig{
				RestConfig: PredictorImageConfig{ContainerImage: "rest"},
				GrpcConfig: PredictorImageConfig{ContainerImage: "grpc"},
			},
			protocol:     ProtocolSeldon,
			endpointType: REST,
			expected:     &PredictorImageConfig{ContainerImage: "rest"},
		},
		{
			config: &PredictorServerConfig{
				RestConfig: PredictorImageConfig{ContainerImage: "rest"},
				GrpcConfig: PredictorImageConfig{ContainerImage: "grpc"},
			},
			protocol:     ProtocolSeldon,
			endpointType: GRPC,
			expected:     &PredictorImageConfig{ContainerImage: "grpc"},
		},
		{
			config: &PredictorServerConfig{
				RestConfig: PredictorImageConfig{ContainerImage: "rest"},
				GrpcConfig: PredictorImageConfig{ContainerImage: "grpc"},
			},
			protocol:     ProtocolKfserving,
			endpointType: GRPC,
			expected:     &PredictorImageConfig{ContainerImage: "grpc"},
		},
		{
			config: &PredictorServerConfig{
				RestConfig: PredictorImageConfig{ContainerImage: "rest"},
				GrpcConfig: PredictorImageConfig{ContainerImage: "grpc"},
				Protocols: PredictorProtocolsConfig{
					KFServing: &PredictorImageConfig{ContainerImage: "kfserving"},
				},
			},
			protocol:     ProtocolKfserving,
			endpointType: GRPC,
			expected:     &PredictorImageConfig{ContainerImage: "kfserving"},
		},
	}

	for _, test := range tests {
		mlDep := &SeldonDeploymentSpec{Protocol: test.protocol}
		pu := &PredictiveUnit{Endpoint: &Endpoint{Type: test.endpointType}}

		imageConfig := test.config.PrepackImageConfig(mlDep, pu)

		g.Expect(imageConfig).To(Equal(test.expected))
	}
}

func TestPredictorServerConfigPrepackImageName(t *testing.T) {
	g := NewGomegaWithT(t)

	tests := []struct {
		imageConfig *PredictorImageConfig
		expected    string
	}{
		{
			imageConfig: &PredictorImageConfig{
				ContainerImage:      "my-image",
				DefaultImageVersion: "v0.1.0",
			},
			expected: "my-image:v0.1.0",
		},
		{
			imageConfig: &PredictorImageConfig{
				ContainerImage: "my-image",
			},
			expected: "my-image",
		},
		{
			imageConfig: &PredictorImageConfig{
				ContainerImage: "my-image:0.2.0",
			},
			expected: "my-image:0.2.0",
		},
		{
			imageConfig: nil,
			expected:    "",
		},
	}

	for _, test := range tests {
		p := &PredictorServerConfig{}
		if test.imageConfig != nil {
			p.RestConfig = *test.imageConfig
		}
		mlDep := &SeldonDeploymentSpec{}
		pu := &PredictiveUnit{Endpoint: &Endpoint{Type: REST}}

		image := p.PrepackImageName(mlDep, pu)

		g.Expect(image).To(Equal(test.expected))
	}
}
