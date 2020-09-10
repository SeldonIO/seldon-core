package controllers

import (
	machinelearningv1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	"github.com/seldonio/seldon-core/operator/constants"
	v1 "k8s.io/api/core/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("MLServer helpers", func() {
	var pu *machinelearningv1.PredictiveUnit

	BeforeEach(func() {
		puType := machinelearningv1.MODEL
		puImplementation := machinelearningv1.PredictiveUnitImplementation(
			machinelearningv1.PrepackSklearnName,
		)

		pu = &machinelearningv1.PredictiveUnit{
			Name:           "my-model",
			Type:           &puType,
			Implementation: &puImplementation,
			Endpoint: &machinelearningv1.Endpoint{
				Type:        machinelearningv1.REST,
				ServicePort: int32(5001),
			},
		}
	})

	Describe("getMLServerEnvVars", func() {
		var envs []v1.EnvVar

		BeforeEach(func() {
			envs = getMLServerEnvVars(pu)
		})

		It("adds the right ports", func() {
			var httpPort string
			var grpcPort string
			for _, env := range envs {
				switch env.Name {
				case MLServerHTTPPortEnv:
					httpPort = env.Value
				case MLServerGRPCPortEnv:
					grpcPort = env.Value
				}
			}

			Expect(httpPort).To(Equal(string(pu.Endpoint.ServicePort)))
			Expect(grpcPort).To(Equal(string(constants.MLServerDefaultGrpcPort)))
		})

		It("adds the right model implementation and uri", func() {
			var modelImplementation string
			var modelURI string
			for _, env := range envs {
				switch env.Name {
				case MLServerModelImplementationEnv:
					modelImplementation = env.Value
				case MLServerModelURIEnv:
					modelURI = env.Value
				}
			}

			Expect(modelImplementation).To(Equal(MLServerSKLearnImplementation))
			Expect(modelURI).To(Equal(DefaultModelLocalMountPath))
		})
	})

	Describe("getMLServerPort", func() {
		DescribeTable(
			"returns the right port",
			func(endpointType machinelearningv1.EndpointType, serviceEndpointType machinelearningv1.EndpointType, expected int32) {
				pu.Endpoint.Type = serviceEndpointType

				port := getMLServerPort(pu, endpointType)
				Expect(port).To(Equal(expected))
			},
			Entry(
				"default httpPort",
				machinelearningv1.REST,
				machinelearningv1.GRPC,
				constants.MLServerDefaultHttpPort,
			),
			Entry(
				"default grpcPort",
				machinelearningv1.GRPC,
				machinelearningv1.REST,
				constants.MLServerDefaultGrpcPort,
			),
			Entry(
				"service httpPort",
				machinelearningv1.REST,
				machinelearningv1.REST,
				int32(5001),
			),
			Entry(
				"service grpcPort",
				machinelearningv1.GRPC,
				machinelearningv1.GRPC,
				int32(5001),
			),
		)
	})

	Describe("getMLServerModelImplementation", func() {
		DescribeTable(
			"returns the right implementation",
			func(implementation string, expected string) {
				modelImp := machinelearningv1.PredictiveUnitImplementation(implementation)
				pu.Implementation = &modelImp

				mlServerImplementation := getMLServerModelImplementation(pu)

				Expect(mlServerImplementation).To(Equal(expected))
			},
			Entry("sklearn", machinelearningv1.PrepackSklearnName, MLServerSKLearnImplementation),
			Entry("xgboost", machinelearningv1.PrepackXgboostName, MLServerXGBoostImplementation),
			Entry("unknown", "foo", ""),
		)
	})
})
