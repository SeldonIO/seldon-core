/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package v1alpha3

import (
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

var (
	// log is for logging in this package.
	seldondeploymentLog = logf.Log.WithName("seldondeployment")
	C                   client.Client
)

func (r *SeldonDeployment) SetupWebhookWithManager(mgr ctrl.Manager) error {
	C = mgr.GetClient()
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// REMOVE+kubebuilder:webhook:webhookVersions=v1,verbs=create;update,path=/validate-machinelearning-seldon-io-v1-seldondeployment,mutating=false,matchPolicy=exact,failurePolicy=fail,sideEffects=None,admissionReviewVersions=v1;v1beta1,groups=machinelearning.seldon.io,resources=seldondeployments,versions=v1alpha3,name=v1.vseldondeployment.kb.io

var _ webhook.Validator = &SeldonDeployment{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *SeldonDeployment) ValidateCreate() (warnings admission.Warnings, err error) {
	seldondeploymentLog.Info("Validating v1alpha3 Webhook called for CREATE", "name", r.Name)
	return []string{}, r.Spec.ValidateSeldonDeployment()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *SeldonDeployment) ValidateUpdate(_ runtime.Object) (warnings admission.Warnings, err error) {
	seldondeploymentLog.Info("Validating v1alpha3 webhook called for UPDATE", "name", r.Name)
	return []string{}, r.Spec.ValidateSeldonDeployment()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *SeldonDeployment) ValidateDelete() (warnings admission.Warnings, err error) {
	seldondeploymentLog.Info("Validating v1alpha3 webhook called for DELETE", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return []string{}, nil
}
