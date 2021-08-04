package controllers

import (
	machinelearningv1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

func addLabelsToService(svc *corev1.Service, pu *machinelearningv1.PredictiveUnit, p *machinelearningv1.PredictorSpec) *corev1.Service {
	if pu != nil && pu.Type != nil {
		switch *pu.Type {
		case machinelearningv1.ROUTER:
			svc.Labels[machinelearningv1.Label_router] = "true"
		case machinelearningv1.COMBINER:
			svc.Labels[machinelearningv1.Label_combiner] = "true"
		case machinelearningv1.MODEL:
			svc.Labels[machinelearningv1.Label_model] = "true"
		case machinelearningv1.TRANSFORMER:
			svc.Labels[machinelearningv1.Label_transformer] = "true"
		case machinelearningv1.OUTPUT_TRANSFORMER:
			svc.Labels[machinelearningv1.Label_output_transformer] = "true"
		}
	} else if !isEmptyExplainer(p.Explainer) {
		svc.Labels[machinelearningv1.Label_explainer] = "true"
	}
	if p.Shadow {
		svc.Labels[machinelearningv1.Label_shadow] = "true"
	}
	svc.Labels[machinelearningv1.Label_managed_by] = machinelearningv1.Label_value_seldon
	return svc
}

func addLabelsToDeployment(deploy *appsv1.Deployment, pu *machinelearningv1.PredictiveUnit, p *machinelearningv1.PredictorSpec) *appsv1.Deployment {
	if pu != nil && pu.Type != nil {
		switch *pu.Type {
		case machinelearningv1.ROUTER:
			deploy.Labels[machinelearningv1.Label_router] = "true"
			deploy.Spec.Template.ObjectMeta.Labels[machinelearningv1.Label_router] = "true"
		case machinelearningv1.COMBINER:
			deploy.Labels[machinelearningv1.Label_combiner] = "true"
			deploy.Spec.Template.ObjectMeta.Labels[machinelearningv1.Label_combiner] = "true"
		case machinelearningv1.MODEL:
			deploy.Labels[machinelearningv1.Label_model] = "true"
			deploy.Spec.Template.ObjectMeta.Labels[machinelearningv1.Label_model] = "true"
		case machinelearningv1.TRANSFORMER:
			deploy.Labels[machinelearningv1.Label_transformer] = "true"
			deploy.Spec.Template.ObjectMeta.Labels[machinelearningv1.Label_transformer] = "true"
		case machinelearningv1.OUTPUT_TRANSFORMER:
			deploy.Labels[machinelearningv1.Label_output_transformer] = "true"
			deploy.Spec.Template.ObjectMeta.Labels[machinelearningv1.Label_output_transformer] = "true"
		}
	} else if !isEmptyExplainer(p.Explainer) {
		deploy.Labels[machinelearningv1.Label_explainer] = "true"
		deploy.Spec.Template.ObjectMeta.Labels[machinelearningv1.Label_explainer] = "true"
	}
	if p.Shadow {
		deploy.Labels[machinelearningv1.Label_shadow] = "true"
		deploy.Spec.Template.ObjectMeta.Labels[machinelearningv1.Label_shadow] = "true"
	}
	deploy.ObjectMeta.Labels[machinelearningv1.Label_managed_by] = machinelearningv1.Label_value_seldon
	deploy.Spec.Template.ObjectMeta.Labels[machinelearningv1.Label_managed_by] = machinelearningv1.Label_value_seldon
	return deploy
}
