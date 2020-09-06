package k8s

import (
	"context"
	"github.com/ghodss/yaml"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
)

type ConfigmapCreator struct {
	clientset kubernetes.Interface
	logger    logr.Logger
	scheme    *runtime.Scheme
}

func NewConfigmapCreator(client kubernetes.Interface, logger logr.Logger, scheme *runtime.Scheme) *ConfigmapCreator {
	return &ConfigmapCreator{
		clientset: client,
		logger:    logger,
		scheme:    scheme,
	}
}

func (cc *ConfigmapCreator) CreateConfigmap(ctx context.Context, rawYaml []byte, namespace string, owner *appsv1.Deployment) error {
	cc.logger.Info("Initialise ConfigMap")
	cm := corev1.ConfigMap{}

	err := yaml.Unmarshal(rawYaml, &cm)
	if err != nil {
		cc.logger.Info("Failed to unmarshall configmap")
		return err
	}

	//Set namespace
	cm.Namespace = namespace

	// add ownership
	err = ctrl.SetControllerReference(owner, &cm, cc.scheme)
	if err != nil {
		return err
	}

	client := cc.clientset.CoreV1().ConfigMaps(namespace)
	_, err = client.Get(ctx, cm.Name, v1.GetOptions{})
	if err != nil && errors.IsNotFound(err) {
		cc.logger.Info("Creating configmap")
		_, err = client.Create(ctx, &cm, v1.CreateOptions{})
	} else if err == nil {
		cc.logger.Info("Configmap exists will not overwrite")
	} else {
		cc.logger.Error(err, "Failed to get configmap")
	}
	return err
}
