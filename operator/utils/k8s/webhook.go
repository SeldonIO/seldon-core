package k8s

import (
	"github.com/ghodss/yaml"
	"github.com/go-logr/logr"
	"k8s.io/api/admissionregistration/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
	"strconv"
	"strings"
)

type WebhookCreator struct {
	clientset    kubernetes.Interface
	certs        *Cert
	logger       logr.Logger
	majorVersion int
	minorVersion int
	scheme       *runtime.Scheme
}

func NewWebhookCreator(client kubernetes.Interface, certs *Cert, logger logr.Logger, scheme *runtime.Scheme) (*WebhookCreator, error) {
	serverVersion, err := client.Discovery().ServerVersion()
	if err != nil {
		return nil, err
	}
	logger.Info("Server version", "Major", serverVersion.Major, "Minor", serverVersion.Minor)
	majorVersion, err := strconv.Atoi(serverVersion.Major)
	if err != nil {
		logger.Error(err, "Failed to parse majorVersion defaulting to 1")
		majorVersion = 1
	}
	minorVersion, err := strconv.Atoi(serverVersion.Minor)
	if err != nil {
		if strings.HasSuffix(serverVersion.Minor, "+") {
			minorVersion, err = strconv.Atoi(serverVersion.Minor[0 : len(serverVersion.Minor)-1])
			if err != nil {
				logger.Error(err, "Failed to parse minorVersion defaulting to 12")
				minorVersion = 12
			}
		} else {
			logger.Error(err, "Failed to parse minorVersion defaulting to 12")
			minorVersion = 12
		}
	}
	return &WebhookCreator{
		clientset:    client,
		certs:        certs,
		logger:       logger,
		majorVersion: majorVersion,
		minorVersion: minorVersion,
		scheme:       scheme,
	}, nil
}

func (wc *WebhookCreator) CreateMutatingWebhookConfigurationFromFile(rawYaml []byte, namespace string, owner *apiextensionsv1beta1.CustomResourceDefinition, watchNamespace bool) error {
	mwc := v1beta1.MutatingWebhookConfiguration{}
	err := yaml.Unmarshal(rawYaml, &mwc)
	if err != nil {
		return err
	}

	// create or update via client
	client := wc.clientset.AdmissionregistrationV1beta1().MutatingWebhookConfigurations()

	if watchNamespace {
		// Try to delete clusterwide webhook config if available
		_, err := client.Get(mwc.Name, v1.GetOptions{})
		if err != nil && errors.IsNotFound(err) {
			wc.logger.Info("existing clusterwide mwc not found", "name", mwc.Name)
		} else {
			client.Delete(mwc.Name, &v1.DeleteOptions{})
			if err != nil {
				return err
			}
			wc.logger.Info("Deleted clusterwide mwc", "name", mwc.Name)
		}
		mwc.Name = mwc.Name + "-" + namespace
	}

	for idx, _ := range mwc.Webhooks {
		// add caBundle
		mwc.Webhooks[idx].ClientConfig.CABundle = []byte(wc.certs.caPEM)
		// set namespace
		mwc.Webhooks[idx].ClientConfig.Service.Namespace = namespace
		//Remove selector if version too low
		if wc.majorVersion == 1 && wc.minorVersion < 15 {
			//mwc.Webhooks[idx].ObjectSelector = nil
		}

		//Remove namespace exclusion if watchNamespace
		//TODO change to add a namespace selector if the initalizer adds a label to namespace
		if watchNamespace {
			mwc.Webhooks[idx].NamespaceSelector = nil
		}

	}

	// add ownership
	err = ctrl.SetControllerReference(owner, &mwc, wc.scheme)
	if err != nil {
		return err
	}

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

func (wc *WebhookCreator) CreateValidatingWebhookConfigurationFromFile(rawYaml []byte, namespace string, owner *apiextensionsv1beta1.CustomResourceDefinition, watchNamespace bool) error {
	vwc := v1beta1.ValidatingWebhookConfiguration{}
	err := yaml.Unmarshal(rawYaml, &vwc)
	if err != nil {
		return err
	}

	// create or update via client
	client := wc.clientset.AdmissionregistrationV1beta1().ValidatingWebhookConfigurations()

	if watchNamespace {
		// Try to delete clusterwide webhook config if available
		_, err := client.Get(vwc.Name, v1.GetOptions{})
		if err != nil && errors.IsNotFound(err) {
			wc.logger.Info("existing clusterwide vwc not found", "name", vwc.Name)
		} else {
			client.Delete(vwc.Name, &v1.DeleteOptions{})
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
		return err
	}

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

func (wc *WebhookCreator) CreateWebhookServiceFromFile(rawYaml []byte, namespace string, owner *appsv1.Deployment) error {
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
