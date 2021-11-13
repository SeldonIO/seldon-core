package k8s

import (
	"context"
	"github.com/ghodss/yaml"
	"github.com/go-logr/logr"
	adv1 "k8s.io/api/admissionregistration/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
)

const MutatingWebhookName = "seldon-mutating-webhook-configuration"

type WebhookCreator struct {
	clientset kubernetes.Interface
	certs     *Cert
	logger    logr.Logger
	scheme    *runtime.Scheme
}

func NewWebhookCreator(client kubernetes.Interface, certs *Cert, logger logr.Logger, scheme *runtime.Scheme) *WebhookCreator {
	return &WebhookCreator{
		clientset: client,
		certs:     certs,
		logger:    logger.WithName("WebhookCreator"),
		scheme:    scheme,
	}
}

func (wc *WebhookCreator) CreateValidatingWebhookConfigurationFromFile(ctx context.Context, rawYaml []byte, namespace string, owner v1.Object, watchNamespace bool) error {
	vwc := adv1.ValidatingWebhookConfiguration{}
	err := yaml.Unmarshal(rawYaml, &vwc)
	if err != nil {
		wc.logger.Error(err, "Failed to unmarshall validating webhook")
		return err
	}

	// create or update via client
	client := wc.clientset.AdmissionregistrationV1().ValidatingWebhookConfigurations()

	if watchNamespace {
		// Try to delete clusterwide webhook config if available
		_, err := client.Get(ctx, vwc.Name, v1.GetOptions{})
		if err != nil {
			if errors.IsNotFound(err) {
				wc.logger.Info("existing clusterwide vwc not found", "name", vwc.Name)
			} else {
				wc.logger.Error(err, "Ignoring error trying to get existing validating webhook")
			}
		} else {
			client.Delete(ctx, vwc.Name, v1.DeleteOptions{})
			if err != nil {
				return err
			}
			wc.logger.Info("Deleted clusterwide vwc", "name", vwc.Name)
		}
		vwc.Name = vwc.Name + "-" + namespace
	}

	// add caBundle
	for idx, _ := range vwc.Webhooks {
		vwc.Webhooks[idx].ClientConfig.CABundle = []byte(wc.certs.caPEM)
		// set namespace
		vwc.Webhooks[idx].ClientConfig.Service.Namespace = namespace

		//Remove namespace exclusion if watchNamespace
		//TODO change to add a namespace selector if the initalizer adds a label to namespace
		if watchNamespace {
			vwc.Webhooks[idx].NamespaceSelector = nil
		}
	}

	// add ownership
	err = ctrl.SetControllerReference(owner, &vwc, wc.scheme)
	if err != nil {
		wc.logger.Error(err, "Failed to set owner reference on validating webhook")
		return err
	}

	found, err := client.Get(ctx, vwc.Name, v1.GetOptions{})
	if err != nil && errors.IsNotFound(err) {
		wc.logger.Info("Creating validating webhook")
		_, err = client.Create(ctx, &vwc, v1.CreateOptions{})
		if err != nil {
			wc.logger.Error(err, "Failed to create validating webhook")
		}
	} else if err == nil {
		wc.logger.Info("Updating validating webhook")
		found.Webhooks = vwc.Webhooks
		_, err = client.Update(ctx, found, v1.UpdateOptions{})
		if err != nil {
			wc.logger.Error(err, "Failed to update validating webhook")
		}
		return err
	}
	return err
}

func (wc *WebhookCreator) CreateWebhookServiceFromFile(ctx context.Context, rawYaml []byte, namespace string, owner *appsv1.Deployment) error {
	svcRaw := corev1.Service{}
	err := yaml.Unmarshal(rawYaml, &svcRaw)
	if err != nil {
		return err
	}

	svc := corev1.Service{}
	svc.ObjectMeta = svcRaw.ObjectMeta
	svc.Spec.Ports = svcRaw.Spec.Ports
	svc.Spec.Selector = svcRaw.Spec.Selector

	// Ensure namespace matches required namespace
	svc.Namespace = namespace

	// add ownership
	err = ctrl.SetControllerReference(owner, &svc, wc.scheme)
	if err != nil {
		return err
	}

	// create or update via client
	client := wc.clientset.CoreV1().Services(namespace)

	_, err = client.Get(ctx, svc.Name, v1.GetOptions{})
	if err != nil && errors.IsNotFound(err) {
		wc.logger.Info("Creating webhook svc")
		_, err = client.Create(ctx, &svc, v1.CreateOptions{})
		if err != nil {
			wc.logger.Error(err, "Failed to update webhook svc")
		}
	} else if err == nil {
		wc.logger.Info("Webhook svc exists won't update - need a patch in future")
		return nil
	}
	return err
}
