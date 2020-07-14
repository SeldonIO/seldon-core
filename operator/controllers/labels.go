package controllers

import (
	machinelearningv1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

func addLabelsToService(svc *corev1.Service, pu *machinelearningv1.PredictiveUnit, p *machinelearningv1.PredictorSpec) {
	if pu.Type != nil {
		switch *pu.Type {
		case machinelearningv1.ROUTER:
			svc.Labels[machinelearningv1.Label_component] = machinelearningv1.Label_router
		case machinelearningv1.COMBINER:
			svc.Labels[machinelearningv1.Label_component] = machinelearningv1.Label_combiner
		case machinelearningv1.MODEL:
			svc.Labels[machinelearningv1.Label_component] = machinelearningv1.Label_model
		case machinelearningv1.TRANSFORMER:
			svc.Labels[machinelearningv1.Label_component] = machinelearningv1.Label_transformer
		case machinelearningv1.OUTPUT_TRANSFORMER:
			svc.Labels[machinelearningv1.Label_component] = machinelearningv1.Label_output_transformer
		}
	}
	if p.Shadow != true && (p.Traffic >= 50 || p.Traffic == 0) {
		svc.Labels[machinelearningv1.Label_endpoint] = machinelearningv1.Label_default
	}
	if p.Shadow == true {
		svc.Labels[machinelearningv1.Label_endpoint] = machinelearningv1.Label_shadow
	}
	if p.Traffic < 50 && p.Traffic > 0 {
		svc.Labels[machinelearningv1.Label_endpoint] = machinelearningv1.Label_canary
	}
	if !isEmptyExplainer(p.Explainer) {
		svc.Labels[machinelearningv1.Label_endpoint] = machinelearningv1.Label_explainer
	}
	svc.Labels[machinelearningv1.Label_managed_by] = machinelearningv1.Label_value_seldon
}

func addLabelsToDeployment(deploy *appsv1.Deployment, pu *machinelearningv1.PredictiveUnit, p *machinelearningv1.PredictorSpec) {
	if pu.Type != nil {
		switch *pu.Type {
		case machinelearningv1.ROUTER:
			deploy.Labels[machinelearningv1.Label_component] = machinelearningv1.Label_router
			deploy.Spec.Template.ObjectMeta.Labels[machinelearningv1.Label_component] = machinelearningv1.Label_router
		case machinelearningv1.COMBINER:
			deploy.Labels[machinelearningv1.Label_component] = machinelearningv1.Label_combiner
			deploy.Spec.Template.ObjectMeta.Labels[machinelearningv1.Label_component] = machinelearningv1.Label_combiner
		case machinelearningv1.MODEL:
			deploy.Labels[machinelearningv1.Label_component] = machinelearningv1.Label_model
			deploy.Spec.Template.ObjectMeta.Labels[machinelearningv1.Label_component] = machinelearningv1.Label_model
		case machinelearningv1.TRANSFORMER:
			deploy.Labels[machinelearningv1.Label_component] = machinelearningv1.Label_transformer
			deploy.Spec.Template.ObjectMeta.Labels[machinelearningv1.Label_component] = machinelearningv1.Label_transformer
		case machinelearningv1.OUTPUT_TRANSFORMER:
			deploy.Labels[machinelearningv1.Label_component] = machinelearningv1.Label_output_transformer
			deploy.Spec.Template.ObjectMeta.Labels[machinelearningv1.Label_component] = machinelearningv1.Label_output_transformer
		}
	}
	if p.Shadow != true && (p.Traffic >= 50 || p.Traffic == 0) {
		deploy.Labels[machinelearningv1.Label_endpoint] = machinelearningv1.Label_default
		deploy.Spec.Template.ObjectMeta.Labels[machinelearningv1.Label_endpoint] = machinelearningv1.Label_default
	}
	if p.Shadow == true {
		deploy.Labels[machinelearningv1.Label_endpoint] = machinelearningv1.Label_shadow
		deploy.Spec.Template.ObjectMeta.Labels[machinelearningv1.Label_endpoint] = machinelearningv1.Label_shadow
	}
	if p.Traffic < 50 && p.Traffic > 0 {
		deploy.Labels[machinelearningv1.Label_endpoint] = machinelearningv1.Label_canary
		deploy.Spec.Template.ObjectMeta.Labels[machinelearningv1.Label_endpoint] = machinelearningv1.Label_canary
	}
	if !isEmptyExplainer(p.Explainer) {
		deploy.Labels[machinelearningv1.Label_endpoint] = machinelearningv1.Label_explainer
		deploy.Spec.Template.ObjectMeta.Labels[machinelearningv1.Label_endpoint] = machinelearningv1.Label_explainer
	}
	deploy.ObjectMeta.Labels[machinelearningv1.Label_managed_by] = machinelearningv1.Label_value_seldon
	deploy.Spec.Template.ObjectMeta.Labels[machinelearningv1.Label_managed_by] = machinelearningv1.Label_value_seldon
}
