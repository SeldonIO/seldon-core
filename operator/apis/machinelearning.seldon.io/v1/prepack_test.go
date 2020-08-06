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
