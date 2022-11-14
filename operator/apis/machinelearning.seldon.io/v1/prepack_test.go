package v1

import (
	"testing"

	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
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
					ProtocolSeldon: {ContainerImage: "a", DefaultImageVersion: "1"},
				},
			},
			desiredImageName: "a:1",
		},
		"GRPC image with no version": {
			pu: &PredictiveUnit{Endpoint: &Endpoint{Type: GRPC}},
			config: &PredictorServerConfig{
				Protocols: map[Protocol]PredictorImageConfig{
					ProtocolSeldon: {ContainerImage: "a", DefaultImageVersion: "1"},
				},
			},
			desiredImageName: "a:1",
		},
		"REST image with version": {
			pu: &PredictiveUnit{Endpoint: &Endpoint{Type: REST}},
			config: &PredictorServerConfig{
				Protocols: map[Protocol]PredictorImageConfig{
					ProtocolSeldon: {ContainerImage: "a", DefaultImageVersion: "1"},
				},
			},
			desiredImageName: "a:1",
		},
		"REST image with no version": {
			pu: &PredictiveUnit{Endpoint: &Endpoint{Type: REST}},
			config: &PredictorServerConfig{
				Protocols: map[Protocol]PredictorImageConfig{
					ProtocolSeldon: {ContainerImage: "a", DefaultImageVersion: "1"},
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
		protocol        Protocol
		serverName      string
		relatedImageMap map[string]PredictorServerConfig
		desiredConfig   PredictorServerConfig
	}{
		"default image sklearn": {
			protocol:        "",
			serverName:      PrepackSklearnName,
			relatedImageMap: map[string]PredictorServerConfig{},
			desiredConfig:   serverConfigs[PrepackSklearnName],
		},
		"default image xgboost": {
			protocol:        "",
			serverName:      PrepackXGBoostName,
			relatedImageMap: map[string]PredictorServerConfig{},
			desiredConfig:   serverConfigs[PrepackXGBoostName],
		},
		"default image mlflow": {
			protocol:        "",
			serverName:      PrepackMLFlowName,
			relatedImageMap: map[string]PredictorServerConfig{},
			desiredConfig:   serverConfigs[PrepackMLFlowName],
		},
		"related image sklearn v1": {
			protocol:   ProtocolSeldon,
			serverName: PrepackSklearnName,
			relatedImageMap: map[string]PredictorServerConfig{PrepackSklearnName: {Protocols: map[Protocol]PredictorImageConfig{
				ProtocolSeldon: {ContainerImage: "related-sklearn:v1"}}},
			},
			desiredConfig: PredictorServerConfig{Protocols: map[Protocol]PredictorImageConfig{ProtocolSeldon: {ContainerImage: "related-sklearn:v1"}}},
		},
		"related image sklearn v2": {
			protocol:   ProtocolV2,
			serverName: PrepackSklearnName,
			relatedImageMap: map[string]PredictorServerConfig{PrepackSklearnName: {Protocols: map[Protocol]PredictorImageConfig{
				ProtocolV2: {ContainerImage: "related-sklearn:v2"}}},
			},
			desiredConfig: PredictorServerConfig{Protocols: map[Protocol]PredictorImageConfig{ProtocolV2: {ContainerImage: "related-sklearn:v2"}}},
		},
		"related image xgboost v1": {
			protocol:   ProtocolSeldon,
			serverName: PrepackXGBoostName,
			relatedImageMap: map[string]PredictorServerConfig{PrepackXGBoostName: {Protocols: map[Protocol]PredictorImageConfig{
				ProtocolSeldon: {ContainerImage: "related-xgboost:v1"}}},
			},
			desiredConfig: PredictorServerConfig{Protocols: map[Protocol]PredictorImageConfig{ProtocolSeldon: {ContainerImage: "related-xgboost:v1"}}},
		},
		"related image xgboost v2": {
			protocol:   ProtocolV2,
			serverName: PrepackXGBoostName,
			relatedImageMap: map[string]PredictorServerConfig{PrepackXGBoostName: {Protocols: map[Protocol]PredictorImageConfig{
				ProtocolV2: {ContainerImage: "related-xgboost:v2"}}},
			},
			desiredConfig: PredictorServerConfig{Protocols: map[Protocol]PredictorImageConfig{ProtocolV2: {ContainerImage: "related-xgboost:v2"}}},
		},
		"related image mlflow v1": {
			protocol:   ProtocolSeldon,
			serverName: PrepackMLFlowName,
			relatedImageMap: map[string]PredictorServerConfig{PrepackMLFlowName: {Protocols: map[Protocol]PredictorImageConfig{
				ProtocolSeldon: {ContainerImage: "related-mlflow:v1"}}},
			},
			desiredConfig: PredictorServerConfig{Protocols: map[Protocol]PredictorImageConfig{ProtocolSeldon: {ContainerImage: "related-mlflow:v1"}}},
		},
		"related image mlflow v2": {
			protocol:   ProtocolV2,
			serverName: PrepackMLFlowName,
			relatedImageMap: map[string]PredictorServerConfig{PrepackMLFlowName: {Protocols: map[Protocol]PredictorImageConfig{
				ProtocolV2: {ContainerImage: "related-mlflow:v2"}}},
			},
			desiredConfig: PredictorServerConfig{Protocols: map[Protocol]PredictorImageConfig{ProtocolV2: {ContainerImage: "related-mlflow:v2"}}},
		},
	}
	for name, scenario := range scenarios {
		t.Logf("Scenario: %s", name)
		config := getPrepackServerConfigWithRelated(scenario.serverName, scenario.relatedImageMap)
		if scenario.protocol != "" {
			g.Expect(config.Protocols[scenario.protocol]).To(Equal(scenario.desiredConfig.Protocols[scenario.protocol]))
		} else {
			g.Expect(*config).To(Equal(scenario.desiredConfig))
		}
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
