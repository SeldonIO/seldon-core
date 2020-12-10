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
				Protocols: map[Protocol]PredictorImageConfig{
					ProtocolSeldon: PredictorImageConfig{ContainerImage: "a", DefaultImageVersion: "1"},
				},
			},
			desiredImageName: "a:1",
		},
		"GRPC image with no version": {
			pu: &PredictiveUnit{Endpoint: &Endpoint{Type: GRPC}},
			config: &PredictorServerConfig{
				Protocols: map[Protocol]PredictorImageConfig{
					ProtocolSeldon: PredictorImageConfig{ContainerImage: "a", DefaultImageVersion: "1"},
				},
			},
			desiredImageName: "a:1",
		},
		"REST image with version": {
			pu: &PredictiveUnit{Endpoint: &Endpoint{Type: REST}},
			config: &PredictorServerConfig{
				Protocols: map[Protocol]PredictorImageConfig{
					ProtocolSeldon: PredictorImageConfig{ContainerImage: "a", DefaultImageVersion: "1"},
				},
			},
			desiredImageName: "a:1",
		},
		"REST image with no version": {
			pu: &PredictiveUnit{Endpoint: &Endpoint{Type: REST}},
			config: &PredictorServerConfig{
				Protocols: map[Protocol]PredictorImageConfig{
					ProtocolSeldon: PredictorImageConfig{ContainerImage: "a", DefaultImageVersion: "1"},
				},
			},
			desiredImageName: "a:1",
		},
	}

	for name, scenario := range scenarios {
		t.Logf("Scenario: %s", name)
		con := &corev1.Container{}
		con.Image = scenario.config.PrepackImageName(ProtocolSeldon, scenario.pu)
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
			serverName: PrepackSklearnName,
			relatedImageMap: map[string]PredictorServerConfig{PrepackSklearnName: {Protocols: map[Protocol]PredictorImageConfig{
				ProtocolSeldon: {ContainerImage: "a"}}},
			},
			desiredConfig: PredictorServerConfig{Protocols: map[Protocol]PredictorImageConfig{ProtocolSeldon: {ContainerImage: "a"}}},
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
			p.Protocols = map[Protocol]PredictorImageConfig{ProtocolSeldon: *test.imageConfig}
		}
		pu := &PredictiveUnit{Endpoint: &Endpoint{Type: REST}}

		image := p.PrepackImageName(ProtocolSeldon, pu)

		g.Expect(image).To(Equal(test.expected))
	}
}
