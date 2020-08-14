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

	var dep *appsv1.Deployment
	var p *machinelearningv1.PredictorSpec
	var pu *machinelearningv1.PredictiveUnit

	BeforeEach(func() {
		dep = &appsv1.Deployment{
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
		"Adds correct label to Deployment from Predictive Unit Type",
		func(puType machinelearningv1.PredictiveUnitType, result string) {
			pu.Type = &puType
			addLabelsToDeployment(dep, pu, p)

			Expect(dep.Labels[result]).To(Equal("true"))
			Expect(dep.Spec.Template.ObjectMeta.Labels[result]).To(Equal("true"))
			Expect(dep.Labels[machinelearningv1.Label_managed_by]).To(Equal(machinelearningv1.Label_value_seldon))
			Expect(dep.Spec.Template.ObjectMeta.Labels[machinelearningv1.Label_managed_by]).To(Equal(machinelearningv1.Label_value_seldon))
		},
		Entry("router", machinelearningv1.ROUTER, machinelearningv1.Label_router),
		Entry("combiner", machinelearningv1.COMBINER, machinelearningv1.Label_combiner),
		Entry("model", machinelearningv1.MODEL, machinelearningv1.Label_model),
		Entry("transformer", machinelearningv1.TRANSFORMER, machinelearningv1.Label_transformer),
		Entry("output transformer", machinelearningv1.OUTPUT_TRANSFORMER, machinelearningv1.Label_output_transformer),
	)

	DescribeTable(
		"Adds correct label to Deployment from Predictor Spec",
		func(shadow bool, explainer *machinelearningv1.Explainer, traffic int, result string, missing []string) {
			p.Shadow = shadow
			p.Explainer = explainer
			p.Traffic = int32(traffic)

			// Required until https://github.com/SeldonIO/seldon-core/pull/1600 is
			// merged. We should remove afterwards.
			if explainer != nil {
				pu.Type = nil
			} else {
				puType := machinelearningv1.MODEL
				pu.Type = &puType
			}

			addLabelsToDeployment(dep, pu, p)

			Expect(dep.Labels[result]).To(Equal("true"))
			for _, m := range missing {
				Expect(dep.Labels).ToNot(HaveKey(m))
			}
		},
		Entry(
			"default",
			false,
			nil,
			0,
			machinelearningv1.Label_default,
			[]string{machinelearningv1.Label_explainer},
		),
		Entry(
			"default with 50%% traffic",
			false,
			nil,
			50,
			machinelearningv1.Label_default,
			[]string{machinelearningv1.Label_explainer},
		),
		Entry(
			"default with 75%% traffic",
			false,
			nil,
			50,
			machinelearningv1.Label_default,
			[]string{machinelearningv1.Label_explainer},
		),
		Entry(
			"default with empty explainer",
			false,
			&machinelearningv1.Explainer{},
			0,
			machinelearningv1.Label_default,
			[]string{machinelearningv1.Label_explainer},
		),
		Entry(
			"shadow",
			true,
			nil,
			0,
			machinelearningv1.Label_shadow,
			[]string{
				machinelearningv1.Label_explainer,
				machinelearningv1.Label_default,
			},
		),
		Entry(
			"canary",
			false,
			nil,
			48,
			machinelearningv1.Label_canary,
			[]string{
				machinelearningv1.Label_explainer,
				machinelearningv1.Label_default,
			},
		),
		Entry(
			"explainer",
			false,
			&machinelearningv1.Explainer{
				Type: machinelearningv1.AlibiAnchorsImageExplainer,
			},
			0,
			machinelearningv1.Label_explainer,
			[]string{},
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
		"Adds correct label to Service from Predictive Unit Type",
		func(puType machinelearningv1.PredictiveUnitType, result string) {
			pu.Type = &puType
			addLabelsToService(svc, pu, p)

			Expect(svc.Labels[result]).To(Equal("true"))
			Expect(svc.Labels[machinelearningv1.Label_managed_by]).To(Equal(machinelearningv1.Label_value_seldon))
		},
		Entry("router", machinelearningv1.ROUTER, machinelearningv1.Label_router),
		Entry("combiner", machinelearningv1.COMBINER, machinelearningv1.Label_combiner),
		Entry("model", machinelearningv1.MODEL, machinelearningv1.Label_model),
		Entry("transformer", machinelearningv1.TRANSFORMER, machinelearningv1.Label_transformer),
		Entry("output transformer", machinelearningv1.OUTPUT_TRANSFORMER, machinelearningv1.Label_output_transformer),
	)

	DescribeTable(
		"Adds correct label to Service from Predictor Spec",
		func(shadow bool, explainer *machinelearningv1.Explainer, traffic int, result string, missing []string) {
			p.Shadow = shadow
			p.Explainer = explainer
			p.Traffic = int32(traffic)

			// Required until https://github.com/SeldonIO/seldon-core/pull/1600 is
			// merged. We should remove afterwards.
			if explainer != nil {
				pu.Type = nil
			} else {
				puType := machinelearningv1.MODEL
				pu.Type = &puType
			}

			addLabelsToService(svc, pu, p)

			Expect(svc.Labels[result]).To(Equal("true"))
			for _, m := range missing {
				Expect(svc.Labels).ToNot(HaveKey(m))
			}
		},
		Entry(
			"default",
			false,
			nil,
			0,
			machinelearningv1.Label_default,
			[]string{machinelearningv1.Label_explainer},
		),
		Entry(
			"default with 50%% traffic",
			false,
			nil,
			50,
			machinelearningv1.Label_default,
			[]string{machinelearningv1.Label_explainer},
		),
		Entry(
			"default with 75%% traffic",
			false,
			nil,
			50,
			machinelearningv1.Label_default,
			[]string{machinelearningv1.Label_explainer},
		),
		Entry(
			"default with empty explainer",
			false,
			&machinelearningv1.Explainer{},
			0,
			machinelearningv1.Label_default,
			[]string{machinelearningv1.Label_explainer},
		),
		Entry(
			"shadow",
			true,
			nil,
			0,
			machinelearningv1.Label_shadow,
			[]string{
				machinelearningv1.Label_explainer,
				machinelearningv1.Label_default,
			},
		),
		Entry(
			"canary",
			false,
			nil,
			48,
			machinelearningv1.Label_canary,
			[]string{
				machinelearningv1.Label_explainer,
				machinelearningv1.Label_default,
			},
		),
		Entry(
			"explainer",
			false,
			&machinelearningv1.Explainer{
				Type: machinelearningv1.AlibiAnchorsImageExplainer,
			},
			0,
			machinelearningv1.Label_explainer,
			[]string{},
		),
	)
})
