package k8s

import (
	"context"
	"github.com/Masterminds/semver"
	"github.com/ghodss/yaml"
	"github.com/go-logr/logr"
	extensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/discovery"
)

type CrdCreator struct {
	apiExtensionsClient apiextensionsclient.Interface
	discoveryClient     discovery.DiscoveryInterface
	logger              logr.Logger
	ctx                 context.Context
}

func NewCrdCreator(ctx context.Context, apiExtensionsClient apiextensionsclient.Interface, discoveryClient discovery.DiscoveryInterface, logger logr.Logger) *CrdCreator {
	return &CrdCreator{
		apiExtensionsClient: apiExtensionsClient,
		discoveryClient:     discoveryClient,
		logger:              logger,
		ctx:                 ctx,
	}
}

func (cc *CrdCreator) findCRDv1beta1() (*v1beta1.CustomResourceDefinition, error) {
	client := cc.apiExtensionsClient.ApiextensionsV1beta1().CustomResourceDefinitions()
	return client.Get(cc.ctx, CRDName, v1.GetOptions{})
}

func (cc *CrdCreator) findCRDv1() (*extensionsv1.CustomResourceDefinition, error) {
	client := cc.apiExtensionsClient.ApiextensionsV1().CustomResourceDefinitions()
	return client.Get(cc.ctx, CRDName, v1.GetOptions{})
}

func (cc *CrdCreator) createCRDV1beta1(rawYaml []byte) (*v1beta1.CustomResourceDefinition, error) {
	crd := v1beta1.CustomResourceDefinition{}
	err := yaml.Unmarshal(rawYaml, &crd)
	if err != nil {
		return nil, err
	}
	client := cc.apiExtensionsClient.ApiextensionsV1beta1().CustomResourceDefinitions()
	return client.Create(cc.ctx, &crd, v1.CreateOptions{})
}

func (cc *CrdCreator) createCRDV1(rawYaml []byte) (*extensionsv1.CustomResourceDefinition, error) {
	crd := extensionsv1.CustomResourceDefinition{}
	err := yaml.Unmarshal(rawYaml, &crd)
	if err != nil {
		return nil, err
	}
	client := cc.apiExtensionsClient.ApiextensionsV1().CustomResourceDefinitions()
	return client.Create(cc.ctx, &crd, v1.CreateOptions{})
}

func (cc *CrdCreator) findOrCreateCRDV1beta1(rawYaml []byte) (v1.Object, error) {
	//Find or create CRD
	crd, err := cc.findCRDv1beta1()
	if err != nil {
		if errors.IsNotFound(err) {
			// create CRD
			cc.logger.Info("CRD v1beta1 not found - trying to create")
			crd, err = cc.createCRDV1beta1(rawYaml)
			if err != nil {
				return nil, err
			}
			cc.logger.Info("CRD v1beta1 created")
		} else {
			return nil, err
		}
	} else {
		cc.logger.Info("CRD v1beta1 already exists")
	}
	return crd, nil
}

func (cc *CrdCreator) findOrCreateCRDV1(rawYaml []byte) (*extensionsv1.CustomResourceDefinition, error) {
	//Find or create CRD
	crd, err := cc.findCRDv1()
	if err != nil {
		if errors.IsNotFound(err) {
			// create CRD
			cc.logger.Info("CRD V1 not found - trying to create")
			crd, err = cc.createCRDV1(rawYaml)
			if err != nil {
				return nil, err
			}
			cc.logger.Info("CRD V1 created")
		} else {
			return nil, err
		}
	} else {
		cc.logger.Info("CRD v1 already exists")
	}
	return crd, nil
}

func (cc *CrdCreator) findOrCreateCRD(rawYamlv1 []byte, rawYamlv1beta1 []byte) (v1.Object, error) {
	sv, err := cc.discoveryClient.ServerVersion()
	if err != nil {
		return nil, err
	}
	v, err := semver.NewVersion(sv.String())
	if err != nil {
		return nil, err
	}
	c, err := semver.NewConstraint(">= 1.18.0")
	if err != nil {
		return nil, err
	}
	check := c.Check(v)
	if check {
		cc.logger.Info("Creating V1 CRD for K8s", "version", sv.String())
		return cc.findOrCreateCRDV1(rawYamlv1)
	} else {
		cc.logger.Info("Creating V1Beta1 CRD for K8s", "version", sv.String())
		return cc.findOrCreateCRDV1beta1(rawYamlv1beta1)
	}
}
