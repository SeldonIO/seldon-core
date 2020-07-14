package controllers

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	machinelearningv1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("addLabelsToDeployment", func() {

	var d *appsv1.Deployment
	var p *machinelearningv1.PredictorSpec
	var pu *machinelearningv1.PredictiveUnit

	BeforeEach(func() {
		d = &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{},
			},
			Spec: appsv1.DeploymentSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{},
				},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{},
					},
				},
			},
		}
		p = &machinelearningv1.PredictorSpec{}
		pu = &machinelearningv1.PredictiveUnit{}
	})

	DescribeTable(
		"Adds correct Component label to Deployment from Predictive Unit Type",
		func(puType machinelearningv1.PredictiveUnitType, resultComponent, resultEndpoint string) {
			pu.Type = &puType
			addLabelsToDeployment(d, pu, p)

			Expect(d.Labels[machinelearningv1.Label_component]).To(Equal(resultComponent))
			Expect(d.Labels[machinelearningv1.Label_endpoint]).To(Equal(resultEndpoint))
			Expect(d.Labels[machinelearningv1.Label_managed_by]).To(Equal(machinelearningv1.Label_value_seldon))
		},
		Entry("router", machinelearningv1.ROUTER, machinelearningv1.Label_router, machinelearningv1.Label_default),
		Entry("combiner", machinelearningv1.COMBINER, machinelearningv1.Label_combiner, machinelearningv1.Label_default),
		Entry("model", machinelearningv1.MODEL, machinelearningv1.Label_model, machinelearningv1.Label_default),
		Entry("transformer", machinelearningv1.TRANSFORMER, machinelearningv1.Label_transformer, machinelearningv1.Label_default),
		Entry("output transformer", machinelearningv1.OUTPUT_TRANSFORMER, machinelearningv1.Label_output_transformer, machinelearningv1.Label_default),
	)

	DescribeTable(
		"Adds correct Endpoint label to Deployment from Predictor Spec",
		func(shadow bool, explainer *machinelearningv1.Explainer, traffic int, resultComponent, resultEndpoint string) {
			p.Shadow = shadow
			p.Explainer = explainer
			p.Traffic = int32(traffic)

			// Required until https://github.com/SeldonIO/seldon-core/pull/1600 is
			// merged. We should remove afterwards.
			puType := machinelearningv1.MODEL
			pu.Type = &puType

			addLabelsToDeployment(d, pu, p)

			Expect(d.Labels[machinelearningv1.Label_component]).To(Equal(resultComponent))
			Expect(d.Labels[machinelearningv1.Label_endpoint]).To(Equal(resultEndpoint))

			// Check template labels for pod
			Expect(d.Spec.Template.Labels[machinelearningv1.Label_component]).To(Equal(resultComponent))
			Expect(d.Spec.Template.Labels[machinelearningv1.Label_endpoint]).To(Equal(resultEndpoint))
		},
		Entry(
			"default",
			false,
			nil,
			0,
			machinelearningv1.Label_model,
			machinelearningv1.Label_default,
		),
		Entry(
			"default with 50%% traffic",
			false,
			nil,
			50,
			machinelearningv1.Label_model,
			machinelearningv1.Label_default,
		),
		Entry(
			"default with 75%% traffic",
			false,
			nil,
			50,
			machinelearningv1.Label_model,
			machinelearningv1.Label_default,
		),
		Entry(
			"default with empty explainer",
			false,
			&machinelearningv1.Explainer{},
			0,
			machinelearningv1.Label_model,
			machinelearningv1.Label_default,
		),
		Entry(
			"shadow",
			true,
			nil,
			0,
			machinelearningv1.Label_model,
			machinelearningv1.Label_shadow,
		),
		Entry(
			"canary",
			false,
			nil,
			48,
			machinelearningv1.Label_model,
			machinelearningv1.Label_canary,
		),
		Entry(
			"explainer",
			false,
			&machinelearningv1.Explainer{
				Type: machinelearningv1.AlibiAnchorsImageExplainer,
			},
			0,
			machinelearningv1.Label_model,
			machinelearningv1.Label_explainer,
		),
	)
})

var _ = Describe("addLabelsToService", func() {

	var svc *corev1.Service
	var p *machinelearningv1.PredictorSpec
	var pu *machinelearningv1.PredictiveUnit

	BeforeEach(func() {
		svc = &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{},
			},
		}
		p = &machinelearningv1.PredictorSpec{}
		pu = &machinelearningv1.PredictiveUnit{}
	})

	DescribeTable(
		"Adds correct Component label to Service from Predictive Unit Type",
		func(puType machinelearningv1.PredictiveUnitType, resultComponent, resultEndpoint string) {
			pu.Type = &puType
			addLabelsToService(svc, pu, p)

			Expect(svc.Labels[machinelearningv1.Label_component]).To(Equal(resultComponent))
			Expect(svc.Labels[machinelearningv1.Label_endpoint]).To(Equal(resultEndpoint))
			Expect(svc.Labels[machinelearningv1.Label_managed_by]).To(Equal(machinelearningv1.Label_value_seldon))
		},
		Entry("router", machinelearningv1.ROUTER, machinelearningv1.Label_router, machinelearningv1.Label_default),
		Entry("combiner", machinelearningv1.COMBINER, machinelearningv1.Label_combiner, machinelearningv1.Label_default),
		Entry("model", machinelearningv1.MODEL, machinelearningv1.Label_model, machinelearningv1.Label_default),
		Entry("transformer", machinelearningv1.TRANSFORMER, machinelearningv1.Label_transformer, machinelearningv1.Label_default),
		Entry("output transformer", machinelearningv1.OUTPUT_TRANSFORMER, machinelearningv1.Label_output_transformer, machinelearningv1.Label_default),
	)

	DescribeTable(
		"Adds correct label to Service from Predictor Spec",
		func(shadow bool, explainer *machinelearningv1.Explainer, traffic int, resultComponent, resultEndpoint string) {
			p.Shadow = shadow
			p.Explainer = explainer
			p.Traffic = int32(traffic)

			// Required until https://github.com/SeldonIO/seldon-core/pull/1600 is
			// merged. We should remove afterwards.
			puType := machinelearningv1.MODEL
			pu.Type = &puType

			addLabelsToService(svc, pu, p)

			Expect(svc.Labels[machinelearningv1.Label_component]).To(Equal(resultComponent))
			Expect(svc.Labels[machinelearningv1.Label_endpoint]).To(Equal(resultEndpoint))
		},
		Entry(
			"default",
			false,
			nil,
			0,
			machinelearningv1.Label_model,
			machinelearningv1.Label_default,
		),
		Entry(
			"default with 50%% traffic",
			false,
			nil,
			50,
			machinelearningv1.Label_model,
			machinelearningv1.Label_default,
		),
		Entry(
			"default with 75%% traffic",
			false,
			nil,
			50,
			machinelearningv1.Label_model,
			machinelearningv1.Label_default,
		),
		Entry(
			"default with empty explainer",
			false,
			&machinelearningv1.Explainer{},
			0,
			machinelearningv1.Label_model,
			machinelearningv1.Label_default,
		),
		Entry(
			"shadow",
			true,
			nil,
			0,
			machinelearningv1.Label_model,
			machinelearningv1.Label_shadow,
		),
		Entry(
			"canary",
			false,
			nil,
			48,
			machinelearningv1.Label_model,
			machinelearningv1.Label_canary,
		),
		Entry(
			"explainer",
			false,
			&machinelearningv1.Explainer{
				Type: machinelearningv1.AlibiAnchorsImageExplainer,
			},
			0,
			machinelearningv1.Label_model,
			machinelearningv1.Label_explainer,
		),
	)
})
