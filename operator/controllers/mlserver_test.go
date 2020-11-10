package controllers

import (
	"fmt"
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

	Describe("mergeMLServerContainer", func() {
		var existing *v1.Container
		var mlServer *v1.Container
		var merged *v1.Container

		BeforeEach(func() {
			existing = &v1.Container{
				Env: []v1.EnvVar{
					{Name: "FOO", Value: "BAR"},
				},
				LivenessProbe:  &v1.Probe{},
				ReadinessProbe: &v1.Probe{},
			}

			mlServer, _ = getMLServerContainer(pu)

			merged = mergeMLServerContainer(existing, mlServer)
		})

		It("should merge containers adding extra env", func() {
			Expect(merged.Env).To(ContainElement(v1.EnvVar{Name: "FOO", Value: "BAR"}))
			Expect(merged.Env).To(ContainElements(mlServer.Env))
			Expect(merged.Image).To(Equal(mlServer.Image))
		})

		It("should override liveness and readiness probes", func() {
			Expect(merged.LivenessProbe.Handler.HTTPGet.Path).To(Equal(constants.KFServingProbeLivePath))
			Expect(merged.ReadinessProbe.Handler.HTTPGet.Path).To(Equal(constants.KFServingProbeReadyPath))
		})
	})

	Describe("getMLServerContainer", func() {
		var cServer *v1.Container

		BeforeEach(func() {
			cServer, _ = getMLServerContainer(pu)
		})

		It("creates container with image", func() {
			Expect(cServer.Image).To(Equal("seldonio/mlserver:0.1.0"))
		})
	})

	Describe("getMLServerImage", func() {
		It("returns error if no kfserving entry is set", func() {
			invalidImplementation := machinelearningv1.PredictiveUnitImplementation(
				machinelearningv1.PrepackTensorflowName,
			)
			pu.Implementation = &invalidImplementation

			image, err := getMLServerImage(pu)

			Expect(err).To(HaveOccurred())
			Expect(image).To(Equal(""))
		})

		It("returns image name for kfserving", func() {
			image, err := getMLServerImage(pu)

			Expect(err).To(Not(HaveOccurred()))
			Expect(image).To(Equal("seldonio/mlserver:0.1.0"))
		})
	})

	Describe("getMLServerEnvVars", func() {
		var envs []v1.EnvVar

		BeforeEach(func() {
			envs, _ = getMLServerEnvVars(pu)
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

			Expect(httpPort).To(Equal(fmt.Sprint(pu.Endpoint.ServicePort)))
			Expect(grpcPort).To(Equal(fmt.Sprint(constants.MLServerDefaultGrpcPort)))
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

		It("adds the right model name", func() {
			var modelName string
			for _, env := range envs {
				switch env.Name {
				case MLServerModelNameEnv:
					modelName = env.Value
				}
			}

			Expect(modelName).To(Equal(pu.Name))
		})
	})

	Describe("getMLServerPort", func() {
		DescribeTable(
			"returns the right port",
			func(endpointType machinelearningv1.EndpointType, serviceEndpointType machinelearningv1.EndpointType, expected int32) {
				pu.Endpoint.Type = serviceEndpointType

				port, err := getMLServerPort(pu, endpointType)

				Expect(err).NotTo(HaveOccurred())
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

				mlServerImplementation, err := getMLServerModelImplementation(pu)

				if expected == "" {
					Expect(err).To(HaveOccurred())
					Expect(mlServerImplementation).To(Equal(expected))
				} else {
					Expect(err).To(Not(HaveOccurred()))
					Expect(mlServerImplementation).To(Equal(expected))
				}
			},
			Entry("sklearn", machinelearningv1.PrepackSklearnName, MLServerSKLearnImplementation),
			Entry("xgboost", machinelearningv1.PrepackXgboostName, MLServerXGBoostImplementation),
			Entry("unknown", "foo", ""),
		)
	})
})
