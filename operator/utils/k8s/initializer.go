package k8s

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"io/ioutil"
	appsv1 "k8s.io/api/apps/v1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/discovery"
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
	CRDFilenameV1Beta1        = "crd-v1beta1.yaml"
	CRDFilenameV1             = "crd-v1.yaml"
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

func findMyDeployment(ctx context.Context, clientset kubernetes.Interface, namespace string) (*appsv1.Deployment, error) {
	client := clientset.AppsV1().Deployments(namespace)
	return client.Get(ctx, ManagerDeploymentName, v1.GetOptions{})
}

func InitializeOperator(ctx context.Context, config *rest.Config, namespace string, logger logr.Logger, scheme *runtime.Scheme, watchNamespace bool) error {

	apiExtensionClient, err := apiextensionsclient.NewForConfig(config)
	if err != nil {
		return err
	}

	discoveryClient, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return err
	}

	crdCreator := NewCrdCreator(ctx, apiExtensionClient, discoveryClient, logger)
	bytesV1Beta1, err := LoadBytesFromFile(ResourceFolder, CRDFilenameV1Beta1)
	if err != nil {
		return err
	}
	bytesV1, err := LoadBytesFromFile(ResourceFolder, CRDFilenameV1)
	if err != nil {
		return err
	}
	crd, err := crdCreator.findOrCreateCRD(bytesV1, bytesV1Beta1)
	if err != nil {
		return err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	dep, err := findMyDeployment(ctx, clientset, namespace)
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

	//Delete mutating webhook if existing
	err = wc.DeleteMutatingWebhook(ctx)
	if err != nil {
		return err
	}

	//Create/Update Validating Webhook
	bytes, err := LoadBytesFromFile(ResourceFolder, ValidatingWebhookFilename)
	if err != nil {
		return err
	}
	err = wc.CreateValidatingWebhookConfigurationFromFile(ctx, bytes, namespace, crd, watchNamespace)
	if err != nil {
		return err
	}

	//Create/Update Webhook Service
	bytes, err = LoadBytesFromFile(ResourceFolder, ServiceFilename)
	if err != nil {
		return err
	}
	err = wc.CreateWebhookServiceFromFile(ctx, bytes, namespace, dep)
	if err != nil {
		return err
	}

	//Create Configmap
	cc := NewConfigmapCreator(clientset, logger, scheme)
	bytes, err = LoadBytesFromFile(ResourceFolder, ConfigMapFilename)
	if err != nil {
		return err
	}
	err = cc.CreateConfigmap(ctx, bytes, namespace, dep)
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
