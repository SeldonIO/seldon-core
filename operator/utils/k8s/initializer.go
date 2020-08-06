package k8s

import (
	"fmt"
	"github.com/go-logr/logr"
	"io/ioutil"
	appsv1 "k8s.io/api/apps/v1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"os"
	"path/filepath"
)

const (
	CertsFolder = "/tmp/k8s-webhook-server/serving-certs"
	CertsTLSKey = "tls.key"
	CertsTLSCrt = "tls.crt"
	CertsTLSCa  = "ca.crt"

	ResourceFolder            = "/tmp/operator-resources"
	CRDFilename               = "crd.yaml"
	MutatingWebhookFilename   = "mutate.yaml"
	ValidatingWebhookFilename = "validate.yaml"
	ConfigMapFilename         = "configmap.yaml"
	ServiceFilename           = "service.yaml"

	ManagerDeploymentName = "seldon-controller-manager"
	CRDName               = "seldondeployments.machinelearning.seldon.io"
)

func LoadBytesFromFile(path string, name string) ([]byte, error) {
	fullpath := filepath.Join(path, name)
	return ioutil.ReadFile(fullpath)
}

func findMyDeployment(clientset kubernetes.Interface, namespace string) (*appsv1.Deployment, error) {
	client := clientset.AppsV1().Deployments(namespace)
	return client.Get(ManagerDeploymentName, v1.GetOptions{})
}

func InitializeOperator(config *rest.Config, namespace string, logger logr.Logger, scheme *runtime.Scheme, watchNamespace bool) error {

	apiExtensionClient, err := apiextensionsclient.NewForConfig(config)
	if err != nil {
		return err
	}

	crdCreator := NewCrdCreator(apiExtensionClient, logger)
	bytes, err := LoadBytesFromFile(ResourceFolder, CRDFilename)
	if err != nil {
		return err
	}
	crd, err := crdCreator.findOrCreateCRD(bytes)
	if err != nil {
		return err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	dep, err := findMyDeployment(clientset, namespace)
	if err != nil {
		return err
	}

	// Create certs
	host1 := fmt.Sprintf("seldon-webhook-service.%s", namespace)
	host2 := fmt.Sprintf("seldon-webhook-service.%s.svc", namespace)
	certs, err := certSetup([]string{host1, host2})
	if err != nil {
		return err
	}

	// Create webhooks
	wc, err := NewWebhookCreator(clientset, certs, logger, scheme)
	if err != nil {
		return err
	}

	//Create/Update Mutating Webhook
	bytes, err = LoadBytesFromFile(ResourceFolder, MutatingWebhookFilename)
	if err != nil {
		return err
	}
	err = wc.CreateMutatingWebhookConfigurationFromFile(bytes, namespace, crd, watchNamespace)
	if err != nil {
		return err
	}

	//Create/Update Validating Webhook
	bytes, err = LoadBytesFromFile(ResourceFolder, ValidatingWebhookFilename)
	if err != nil {
		return err
	}
	err = wc.CreateValidatingWebhookConfigurationFromFile(bytes, namespace, crd, watchNamespace)
	if err != nil {
		return err
	}

	//Create/Update Webhook Service
	bytes, err = LoadBytesFromFile(ResourceFolder, ServiceFilename)
	if err != nil {
		return err
	}
	err = wc.CreateWebhookServiceFromFile(bytes, namespace, dep)
	if err != nil {
		return err
	}

	//Create Configmap
	cc := NewConfigmapCreator(clientset, logger, scheme)
	bytes, err = LoadBytesFromFile(ResourceFolder, ConfigMapFilename)
	if err != nil {
		return err
	}
	err = cc.CreateConfigmap(bytes, namespace, dep)
	if err != nil {
		return err
	}

	// Create cert files
	createCertFiles(certs, logger)

	return nil
}

func createCertFiles(certs *Cert, logger logr.Logger) error {
	//Save certs to filesystem
	os.MkdirAll(CertsFolder, os.ModePerm)

	filename := fmt.Sprintf("%s/%s", CertsFolder, CertsTLSCa)
	logger.Info("Creating ", "filename", filename)
	err := ioutil.WriteFile(filename, []byte(certs.caPEM), 0600)
	if err != nil {
		return err
	}
	filename = fmt.Sprintf("%s/%s", CertsFolder, CertsTLSKey)
	logger.Info("Creating ", "filename", filename)
	err = ioutil.WriteFile(filename, []byte(certs.privKeyPEM), 0600)
	if err != nil {
		return err
	}
	filename = fmt.Sprintf("%s/%s", CertsFolder, CertsTLSCrt)
	logger.Info("Creating ", "filename", filename)
	err = ioutil.WriteFile(filename, []byte(certs.certificatePEM), 0600)
	if err != nil {
		return err
	}

	return nil
}
