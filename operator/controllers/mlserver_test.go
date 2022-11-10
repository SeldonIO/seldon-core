package controllers

import (
	"fmt"

	machinelearningv1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	"github.com/seldonio/seldon-core/operator/constants"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

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

		It("should set probes to default if not present", func() {
			merged := mergeMLServerContainer(existing, mlServer)

			Expect(merged).ToNot(BeNil())
			Expect(merged.ReadinessProbe).ToNot(BeNil())
			Expect(merged.ReadinessProbe.ProbeHandler.HTTPGet).ToNot(BeNil())
			Expect(merged.ReadinessProbe.ProbeHandler.HTTPGet.Path).To(Equal(constants.KFServingProbeReadyPath))
			Expect(merged.LivenessProbe).ToNot(BeNil())
			Expect(merged.LivenessProbe.ProbeHandler.HTTPGet).ToNot(BeNil())
			Expect(merged.LivenessProbe.ProbeHandler.HTTPGet.Path).To(Equal(constants.KFServingProbeLivePath))
		})

		It("should override only path if probes present", func() {
			existing.ReadinessProbe = &v1.Probe{
				ProbeHandler: v1.ProbeHandler{
					TCPSocket: &v1.TCPSocketAction{Port: intstr.FromString("http")},
				},
				InitialDelaySeconds: 66,
			}
			existing.LivenessProbe = &v1.Probe{
				ProbeHandler: v1.ProbeHandler{
					TCPSocket: &v1.TCPSocketAction{Port: intstr.FromString("http")},
				},
				InitialDelaySeconds: 67,
			}

			merged := mergeMLServerContainer(existing, mlServer)

			Expect(merged).ToNot(BeNil())
			Expect(merged.ReadinessProbe).ToNot(BeNil())
			Expect(merged.ReadinessProbe.ProbeHandler.HTTPGet).ToNot(BeNil())
			Expect(merged.ReadinessProbe.ProbeHandler.HTTPGet.Path).To(Equal(constants.KFServingProbeReadyPath))
			Expect(merged.ReadinessProbe.InitialDelaySeconds).To(Equal(int32(66)))
			Expect(merged.LivenessProbe).ToNot(BeNil())
			Expect(merged.LivenessProbe.ProbeHandler.HTTPGet).ToNot(BeNil())
			Expect(merged.LivenessProbe.ProbeHandler.HTTPGet.Path).To(Equal(constants.KFServingProbeLivePath))
			Expect(merged.LivenessProbe.InitialDelaySeconds).To(Equal(int32(67)))
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
			Entry("xgboost", machinelearningv1.PrepackXGBoostName, MLServerXGBoostImplementation),
			Entry("tempo", machinelearningv1.PrepackTempoName, MLServerTempoImplementation),
			Entry("mlserver", machinelearningv1.PrepackMLFlowName, MLServerMLFlowImplementation),
			Entry("unknown", "foo", ""),
		)
	})
})

var _ = Describe("MLServer explain helpers", func() {
	Describe("getAlibiExplainExtraEnvVars", func() {
		DescribeTable(
			"returns the right extra envs",
			func(explainerType machinelearningv1.AlibiExplainerType, pSvcEndpoint string, graphName string, initParameters string, expected string) {

				extraEnvs, _ := getAlibiExplainExtraEnvVars(explainerType, pSvcEndpoint, graphName, initParameters)
				Expect(extraEnvs).To(Equal(expected))
			},
			Entry("anchor text", machinelearningv1.AlibiAnchorsTabularExplainer, "url", "p", "", "{\"explainer_type\":\"anchor_tabular\",\"infer_uri\":\"http://url/v2/models/p/infer\"}"),
			Entry("anchor image", machinelearningv1.AlibiAnchorsImageExplainer, "url", "p", "", "{\"explainer_type\":\"anchor_image\",\"infer_uri\":\"http://url/v2/models/p/infer\"}"),
			Entry("anchor text with empty init", machinelearningv1.AlibiAnchorsTabularExplainer, "url", "p", "{}", "{\"explainer_type\":\"anchor_tabular\",\"infer_uri\":\"http://url/v2/models/p/infer\",\"init_parameters\":{}}"),
			Entry("anchor text with init", machinelearningv1.AlibiAnchorsTabularExplainer, "url", "p", "{\"v\":2}", "{\"explainer_type\":\"anchor_tabular\",\"infer_uri\":\"http://url/v2/models/p/infer\",\"init_parameters\":{\"v\":2}}"),
		)
	})

	Describe("getAlibiExplainExplainerTypeTag", func() {
		DescribeTable(
			"returns the right explainer tag",
			func(explainerType machinelearningv1.AlibiExplainerType, expected string) {

				tag, err := getAlibiExplainExplainerTypeTag(explainerType)
				if err == nil {
					Expect(tag).To(Equal(expected))
				} else {
					// if there is an error, the tag should also be ""
					Expect(tag).To(Equal(""))
				}
			},
			Entry("anchor text", machinelearningv1.AlibiAnchorsTabularExplainer, "anchor_tabular"),
			Entry("anchor image", machinelearningv1.AlibiAnchorsImageExplainer, "anchor_image"),
			Entry("anchor image", machinelearningv1.AlibiAnchorsTextExplainer, "anchor_text"),
			Entry("anchor image", machinelearningv1.AlibiCounterfactualsExplainer, "counterfactuals"),
			Entry("anchor image", machinelearningv1.AlibiContrastiveExplainer, "contrastive"),
			Entry("anchor image", machinelearningv1.AlibiKernelShapExplainer, "kernel_shap"),
			Entry("anchor image", machinelearningv1.AlibiIntegratedGradientsExplainer, "integrated_gradients"),
			Entry("anchor image", machinelearningv1.AlibiALEExplainer, "ALE"),
			Entry("anchor image", machinelearningv1.AlibiTreeShap, "tree_shap"),
			Entry("unknown", machinelearningv1.AlibiExplainerType("unknown"), ""),
		)
	})
})
