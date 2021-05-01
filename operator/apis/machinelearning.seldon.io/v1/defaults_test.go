package v1

import (
	. "github.com/onsi/gomega"
	"github.com/seldonio/seldon-core/operator/constants"
	"strconv"
	"testing"
)

func TestPredictiveUnitHttpPort(t *testing.T) {
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

	firstHttpPort := int32(9700)
	firstGrpcPort := int32(9800)
	envPredictiveUnitHttpServicePort = strconv.Itoa(int(firstHttpPort))
	envPredictiveUnitGrpcServicePort = strconv.Itoa(int(firstGrpcPort))
	spec.DefaultSeldonDeployment("mydep", "default")

	//Graph
	pu := GetPredictiveUnit(&spec.Predictors[0].Graph, "classifier")
	g.Expect(pu).ToNot(BeNil())
	g.Expect(pu.Endpoint.ServicePort).To(Equal(firstHttpPort))
	g.Expect(pu.Endpoint.HttpPort).To(Equal(firstHttpPort))
	g.Expect(pu.Endpoint.GrpcPort).To(Equal(firstGrpcPort))

	pu = GetPredictiveUnit(&spec.Predictors[0].Graph, "classifier2")
	g.Expect(pu).ToNot(BeNil())
	g.Expect(pu.Endpoint.ServicePort).To(Equal(firstHttpPort + 1))
	g.Expect(pu.Endpoint.HttpPort).To(Equal(firstHttpPort + 1))
	g.Expect(pu.Endpoint.GrpcPort).To(Equal(firstGrpcPort + 1))

	envPredictiveUnitHttpServicePort = ""
	envPredictiveUnitGrpcServicePort = ""

}
