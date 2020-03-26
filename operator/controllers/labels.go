package controllers

import (
	machinelearningv1 "github.com/seldonio/seldon-core/operator/apis/machinelearning/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

func addLabelsToService(svc *corev1.Service, pu *machinelearningv1.PredictiveUnit, p machinelearningv1.PredictorSpec) {
	if pu.Type != nil {
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
	}
	if p.Shadow != true && (p.Traffic >= 50 || p.Traffic == 0) {
		svc.Labels[machinelearningv1.Label_default] = "true"
	}
	if p.Shadow == true {
		svc.Labels[machinelearningv1.Label_shadow] = "true"
	}
	if p.Traffic < 50 && p.Traffic > 0 {
		svc.Labels[machinelearningv1.Label_canary] = "true"
	}
	if !isEmptyExplainer(p.Explainer) {
		svc.Labels[machinelearningv1.Label_explainer] = "true"
	}
}

func addLabelsToDeployment(deploy *appsv1.Deployment, containerServiceKey, containerServiceValue string) {
	deploy.ObjectMeta.Labels[containerServiceKey] = containerServiceValue
	deploy.Spec.Selector.MatchLabels[containerServiceKey] = containerServiceValue
	deploy.Spec.Template.ObjectMeta.Labels[containerServiceKey] = containerServiceValue
}
