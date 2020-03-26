package controllers

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	machinelearningv1 "github.com/seldonio/seldon-core/operator/apis/machinelearning/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestAddLabelsToDeployment(t *testing.T) {
	g := NewGomegaWithT(t)
	d := &appsv1.Deployment{
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
	t.Run("adds correct label to Deployment", func(t *testing.T) {
		addLabelsToDeployment(d, "TestKey", "TestValue")
		g.Expect(d.ObjectMeta.Labels["TestKey"]).To(Equal("TestValue"))
		g.Expect(d.Spec.Selector.MatchLabels["TestKey"]).To(Equal("TestValue"))
		g.Expect(d.Spec.Template.ObjectMeta.Labels["TestKey"]).To(Equal("TestValue"))
	})
}

var _ = Describe("addLabelsToService", func() {

	var svc *corev1.Service
	var p machinelearningv1.PredictorSpec
	var pu *machinelearningv1.PredictiveUnit

	BeforeEach(func() {
		svc = &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{},
			},
		}
		p = machinelearningv1.PredictorSpec{}
		pu = &machinelearningv1.PredictiveUnit{}
	})

	DescribeTable(
		"Adds correct label to Service from Predictive Unit Type",
		func(puType machinelearningv1.PredictiveUnitType, result string) {
			pu.Type = &puType
			addLabelsToService(svc, pu, p)

			Expect(svc.Labels[result]).To(Equal("true"))
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
			puType := machinelearningv1.MODEL
			pu.Type = &puType

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
