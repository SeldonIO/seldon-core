package controllers

import (
	"github.com/go-logr/logr"
	machinelearningv1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	v1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
)

type IngressResourceType int

const (
	ContourHTTPProxies IngressResourceType = iota
	IstioVirtualServices
	IstioDestinationRules
)

func (t IngressResourceType) String() string {
	return [...]string{"ContourHTTPProxies", "IstioVirtualServices", "IstioDestinationRules"}[t]
}

type Ingress interface {
	AddToScheme(scheme *runtime.Scheme)
	// Takes the manager, does any initial setup and returns a list of objects that should be owned by the constructed controller
	SetupWithManager(mgr ctrl.Manager) ([]runtime.Object, error)
	// Generate predictor resources
	GeneratePredictorResources(mlDep *v1.SeldonDeployment, seldonId string, namespace string, ports []httpGrpcPorts, httpAllowed bool, grpcAllowed bool) (map[IngressResourceType][]runtime.Object, error)
	// Generate explainer resources
	GenerateExplainerResources(pSvcName string, p *v1.PredictorSpec, mlDep *v1.SeldonDeployment, seldonId string, namespace string, engineHttpPort int, engineGrpcPort int) (map[IngressResourceType][]runtime.Object, error)
	// Create ingress resources previously generated, returns boolean to indicate readiness
	CreateResources(resources map[IngressResourceType][]runtime.Object, instance *v1.SeldonDeployment, log logr.Logger) (bool, error)
	// Supplies annotations to service objects if required for ingress plugin
	GenerateServiceAnnotations(mlDep *machinelearningv1.SeldonDeployment, p *machinelearningv1.PredictorSpec, serviceName string, engineHttpPort, engineGrpcPort int, isExplainer bool) (map[string]string, error)
}

type DefaultIngress struct {
}

func NewDefaultIngress() Ingress {
	return &DefaultIngress{}
}

func (d DefaultIngress) SetupWithManager(mgr ctrl.Manager) ([]runtime.Object, error) {
	return nil, nil
}

func (d DefaultIngress) GenerateServiceAnnotations(mlDep *machinelearningv1.SeldonDeployment, p *machinelearningv1.PredictorSpec, serviceName string, engineHttpPort, engineGrpcPort int, isExplainer bool) (map[string]string, error) {
	return nil, nil
}

func (d DefaultIngress) GeneratePredictorResources(mlDep *v1.SeldonDeployment, seldonId string, namespace string, ports []httpGrpcPorts, httpAllowed bool, grpcAllowed bool) (map[IngressResourceType][]runtime.Object, error) {
	return nil, nil
}

func (d DefaultIngress) GenerateExplainerResources(pSvcName string, p *v1.PredictorSpec, mlDep *v1.SeldonDeployment, seldonId string, namespace string, engineHttpPort int, engineGrpcPort int) (map[IngressResourceType][]runtime.Object, error) {
	return nil, nil
}

func (d DefaultIngress) CreateResources(resources map[IngressResourceType][]runtime.Object, instance *v1.SeldonDeployment, log logr.Logger) (bool, error) {
	return true, nil
}

func (d DefaultIngress) AddToScheme(_ *runtime.Scheme) {}
