package k8s

import (
	"fmt"
	"github.com/go-logr/logr"
	"io/ioutil"
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
	MutatingWebhookFilename   = "mutate.yaml"
	ValidatingWebhookFilename = "validate.yaml"
	ConfigMapFilename         = "configmap.yaml"
	ServiceFilename           = "service.yaml"
)

func LoadBytesFromFile(path string, name string) ([]byte, error) {
	fullpath := filepath.Join(path, name)
	return ioutil.ReadFile(fullpath)
}

func InitializeOperator(config *rest.Config, namespace string, logger logr.Logger) error {

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}
	host1 := fmt.Sprintf("seldon-webhook-service.%s", namespace)
	host2 := fmt.Sprintf("seldon-webhook-service.%s.svc", namespace)
	certs, err := certSetup([]string{host1, host2})
	if err != nil {
		return err
	}
	wc, err := NewWebhookCreator(clientset, certs, logger)
	if err != nil {
		return err
	}

	//Create/Update Mutating Webhook
	bytes, err := LoadBytesFromFile(ResourceFolder, MutatingWebhookFilename)
	if err != nil {
		return err
	}
	err = wc.CreateMutatingWebhookConfigurationFromFile(bytes)
	if err != nil {
		return err
	}

	//Create/Update Validating Webhook
	bytes, err = LoadBytesFromFile(ResourceFolder, ValidatingWebhookFilename)
	if err != nil {
		return err
	}
	err = wc.CreateValidatingWebhookConfigurationFromFile(bytes)
	if err != nil {
		return err
	}

	//Create/Update Webhook Service
	bytes, err = LoadBytesFromFile(ResourceFolder, ServiceFilename)
	if err != nil {
		return err
	}
	err = wc.CreateWebhookServiceFromFile(bytes, namespace)
	if err != nil {
		return err
	}

	//Create Configmap
	cc := NewConfigmapCreator(clientset, logger)
	bytes, err = LoadBytesFromFile(ResourceFolder, ConfigMapFilename)
	if err != nil {
		return err
	}
	err = cc.CreateConfigmap(bytes, namespace)
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
