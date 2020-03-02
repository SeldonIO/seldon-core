package k8s

import (
	"github.com/ghodss/yaml"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type ConfigmapCreator struct {
	clientset kubernetes.Interface
	logger    logr.Logger
}

func NewConfigmapCreator(client kubernetes.Interface, logger logr.Logger) *ConfigmapCreator {
	return &ConfigmapCreator{
		clientset: client,
		logger:    logger,
	}
}

func (cc *ConfigmapCreator) CreateConfigmap(rawYaml []byte, namespace string) error {
	cc.logger.Info("Initialise ConfigMap")
	cm := corev1.ConfigMap{}

	err := yaml.Unmarshal(rawYaml, &cm)
	if err != nil {
		cc.logger.Info("Failed to unmarshall configmap")
		return err
	}

	//Set namespace
	cm.Namespace = namespace

	client := cc.clientset.CoreV1().ConfigMaps(namespace)
	_, err = client.Get(cm.Name, v1.GetOptions{})
	if err != nil && errors.IsNotFound(err) {
		cc.logger.Info("Creating configmap")
		_, err = client.Create(&cm)
	} else if err == nil {
		cc.logger.Info("Configmap exists will not overwrite")
	} else {
		cc.logger.Error(err, "Failed to get configmap")
	}
	return err
}
