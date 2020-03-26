package controllers

import (
	"testing"

	. "github.com/onsi/gomega"
	machinelearningv1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
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

func TestAddLabelsToService(t *testing.T) {
	g := NewGomegaWithT(t)
	cases1 := []struct {
		puType machinelearningv1.PredictiveUnitType
		result string
	}{
		{
			puType: machinelearningv1.ROUTER,
			result: machinelearningv1.Label_router,
		},
		{
			puType: machinelearningv1.COMBINER,
			result: machinelearningv1.Label_combiner,
		},
		{
			puType: machinelearningv1.MODEL,
			result: machinelearningv1.Label_model,
		},
		{
			puType: machinelearningv1.TRANSFORMER,
			result: machinelearningv1.Label_transformer,
		},
		{
			puType: machinelearningv1.OUTPUT_TRANSFORMER,
			result: machinelearningv1.Label_output_transformer,
		},
	}
	cases2 := []struct {
		shadow    bool
		explainer *machinelearningv1.Explainer
		traffic   int32
		result    string
	}{
		{
			shadow:    false,
			explainer: nil,
			traffic:   0,
			result:    machinelearningv1.Label_default,
		},
		{
			shadow:    false,
			explainer: nil,
			traffic:   50,
			result:    machinelearningv1.Label_default,
		},
		{
			shadow:    false,
			explainer: nil,
			traffic:   75,
			result:    machinelearningv1.Label_default,
		},
		{
			shadow:    true,
			explainer: nil,
			traffic:   0,
			result:    machinelearningv1.Label_shadow,
		},
		{
			shadow:    false,
			explainer: nil,
			traffic:   49,
			result:    machinelearningv1.Label_canary,
		},
		{
			shadow:    false,
			explainer: &machinelearningv1.Explainer{},
			traffic:   0,
			result:    machinelearningv1.Label_explainer,
		},
	}
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{},
		},
	}
	p := machinelearningv1.PredictorSpec{}
	pu := &machinelearningv1.PredictiveUnit{}
	t.Run("adds correct label to Service from Predictive Unit Type", func(t *testing.T) {
		for _, c := range cases1 {
			pu.Type = &c.puType
			addLabelsToService(svc, pu, p)
			g.Expect(svc.Labels[c.result]).To(Equal("true"))
		}
	})
	t.Run("adds correct label to Service for Default, Shadow, Canary & Explainer from Predictor Spec", func(t *testing.T) {
		for _, c := range cases2 {
			p.Shadow = c.shadow
			p.Explainer = c.explainer
			p.Traffic = c.traffic
			addLabelsToService(svc, pu, p)
			g.Expect(svc.Labels[c.result]).To(Equal("true"))
		}
	})
}
