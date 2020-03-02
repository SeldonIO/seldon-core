package k8s

import (
	"github.com/ghodss/yaml"
	"github.com/go-logr/logr"
	"k8s.io/api/admissionregistration/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"strconv"
)

type WebhookCreator struct {
	clientset    kubernetes.Interface
	certs        *Cert
	logger       logr.Logger
	majorVersion int
	minorVersion int
}

func NewWebhookCreator(client kubernetes.Interface, certs *Cert, logger logr.Logger) (*WebhookCreator, error) {
	serverVersion, err := client.Discovery().ServerVersion()
	if err != nil {
		return nil, err
	}
	logger.Info("Server version", "Major", serverVersion.Major, "Minor", serverVersion.Minor)
	majorVersion, err := strconv.Atoi(serverVersion.Major)
	if err != nil {
		logger.Error(err, "Failed to parse majorVersion defaulting to 0")
		majorVersion = 0
	}
	minorVersion, err := strconv.Atoi(serverVersion.Minor)
	if err != nil {
		logger.Error(err, "Failed to parse minorVersion defaulting to 0")
		minorVersion = 0
	}
	return &WebhookCreator{
		clientset:    client,
		certs:        certs,
		logger:       logger,
		majorVersion: majorVersion,
		minorVersion: minorVersion,
	}, nil
}

func (wc *WebhookCreator) CreateMutatingWebhookConfigurationFromFile(rawYaml []byte) error {
	mwc := v1beta1.MutatingWebhookConfiguration{}
	err := yaml.Unmarshal(rawYaml, &mwc)
	if err != nil {
		return err
	}

	for idx, _ := range mwc.Webhooks {
		// add caBundle
		mwc.Webhooks[idx].ClientConfig.CABundle = []byte(wc.certs.caPEM)
		//Remove selector if version too low
		if wc.majorVersion == 1 && wc.minorVersion < 15 {
			mwc.Webhooks[idx].NamespaceSelector = nil
		}
	}

	// add ownership

	// create or update via client
	client := wc.clientset.AdmissionregistrationV1beta1().MutatingWebhookConfigurations()

	found, err := client.Get(mwc.Name, v1.GetOptions{})
	if err != nil && errors.IsNotFound(err) {
		wc.logger.Info("Creating mutating webhook")
		_, err = client.Create(&mwc)
	} else if err == nil {
		wc.logger.Info("Updating mutating webhook")
		found.Webhooks = mwc.Webhooks
		_, err = client.Update(found)
		return err
	}
	return err
}

func (wc *WebhookCreator) CreateValidatingWebhookConfigurationFromFile(rawYaml []byte) error {
	vwc := v1beta1.ValidatingWebhookConfiguration{}
	err := yaml.Unmarshal(rawYaml, &vwc)
	if err != nil {
		return err
	}
	// add caBundle
	for idx, _ := range vwc.Webhooks {
		vwc.Webhooks[idx].ClientConfig.CABundle = []byte(wc.certs.caPEM)
	}

	// add ownership

	// create or update via client
	client := wc.clientset.AdmissionregistrationV1beta1().ValidatingWebhookConfigurations()

	found, err := client.Get(vwc.Name, v1.GetOptions{})
	if err != nil && errors.IsNotFound(err) {
		wc.logger.Info("Creating validating webhook")
		_, err = client.Create(&vwc)
	} else if err == nil {
		wc.logger.Info("Updating validating webhook")
		found.Webhooks = vwc.Webhooks
		_, err = client.Update(found)
		return err
	}
	return err
}

func (wc *WebhookCreator) CreateWebhookServiceFromFile(rawYaml []byte, namespace string) error {
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

	// create or update via client
	client := wc.clientset.CoreV1().Services(namespace)

	_, err = client.Get(svc.Name, v1.GetOptions{})
	if err != nil && errors.IsNotFound(err) {
		wc.logger.Info("Creating webhook svc")
		_, err = client.Create(&svc)
	} else if err == nil {
		wc.logger.Info("Webhook svc exists won't update - need a patch in future")
		return err
	}
	return err
}
