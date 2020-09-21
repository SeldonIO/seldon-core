package k8s

import (
	"context"
	"github.com/ghodss/yaml"
	"github.com/go-logr/logr"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type CrdCreator struct {
	clientset apiextensionsclient.Interface
	logger    logr.Logger
	ctx       context.Context
}

func NewCrdCreator(ctx context.Context, clientset apiextensionsclient.Interface, logger logr.Logger) *CrdCreator {
	return &CrdCreator{
		clientset: clientset,
		logger:    logger,
		ctx:       ctx,
	}
}

func (cc *CrdCreator) findCRD() (*v1beta1.CustomResourceDefinition, error) {
	client := cc.clientset.ApiextensionsV1beta1().CustomResourceDefinitions()
	return client.Get(cc.ctx, CRDName, v1.GetOptions{})
}

func (cc *CrdCreator) createCRD(rawYaml []byte) (*v1beta1.CustomResourceDefinition, error) {
	crd := v1beta1.CustomResourceDefinition{}
	err := yaml.Unmarshal(rawYaml, &crd)
	if err != nil {
		return nil, err
	}
	client := cc.clientset.ApiextensionsV1beta1().CustomResourceDefinitions()
	return client.Create(cc.ctx, &crd, v1.CreateOptions{})
}
func (cc *CrdCreator) findOrCreateCRD(rawYaml []byte) (*v1beta1.CustomResourceDefinition, error) {
	//Find or create CRD
	crd, err := cc.findCRD()
	if err != nil {
		if errors.IsNotFound(err) {
			// create CRD
			cc.logger.Info("CRD not found - trying to create")
			crd, err = cc.createCRD(rawYaml)
			if err != nil {
				return nil, err
			}
			cc.logger.Info("CRD created")
		} else {
			return nil, err
		}
	}
	return crd, nil
}
