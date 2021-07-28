package controllers

import (
	"fmt"
	machinelearningv1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
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
		customEnvValue := "{\"custom\":1}"

		BeforeEach(func() {
			existing = &v1.Container{
				Env: []v1.EnvVar{
					{Name: "FOO", Value: "BAR"},
					{Name: MLServerTempoRuntimeEnv, Value: customEnvValue},
				},
			}

			mlServer, _ = getMLServerContainer(pu, "default")
		})

		It("should merge containers adding extra env", func() {
			merged := mergeMLServerContainer(existing, mlServer)

			Expect(merged).ToNot(BeNil())
			Expect(merged.Env).To(ContainElement(v1.EnvVar{Name: "FOO", Value: "BAR"}))
			Expect(merged.Env).To(ContainElement(v1.EnvVar{Name: MLServerTempoRuntimeEnv, Value: customEnvValue}))
			Expect(merged.Env).To(ContainElements(mlServer.Env))
			Expect(merged.Image).To(Equal(mlServer.Image))
		})
	})

	Describe("getMLServerContainer", func() {
		var cServer *v1.Container

		BeforeEach(func() {
			cServer, _ = getMLServerContainer(pu, "default")
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
			envs, _ = getMLServerEnvVars(pu, "default")
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

			Expect(httpPort).To(Equal(fmt.Sprint(pu.Endpoint.HttpPort)))
			Expect(grpcPort).To(Equal(fmt.Sprint(pu.Endpoint.GrpcPort)))
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

	Describe("getMLServerModelImplementation", func() {
		DescribeTable(
			"returns the right implementation",
			func(implementation string, expected string) {
				modelImp := machinelearningv1.PredictiveUnitImplementation(implementation)
				pu.Implementation = &modelImp

				mlServerImplementation, err := getMLServerModelImplementation(pu)

				Expect(err).To(Not(HaveOccurred()))
				Expect(mlServerImplementation).To(Equal(expected))
			},
			Entry("sklearn", machinelearningv1.PrepackSklearnName, MLServerSKLearnImplementation),
			Entry("xgboost", machinelearningv1.PrepackXgboostName, MLServerXGBoostImplementation),
			Entry("tempo", machinelearningv1.PrepackTempoName, MLServerTempoImplementation),
			Entry("mlserver", machinelearningv1.PrepackMlflowName, MLServerMLFlowImplementation),
			Entry("unknown", "foo", ""),
		)
	})
})
