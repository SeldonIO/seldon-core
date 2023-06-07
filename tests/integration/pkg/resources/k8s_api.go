package resources

import (
	"context"
	"fmt"
	"github.com/seldonio/seldon-core/operator/v2/apis/mlops/v1alpha1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

type SeldonK8sAPI struct {
	namespace   string
	client      client.Client
	inferClient *SeldonInferAPI
}

func NewSeldonK8sAPI() (*SeldonK8sAPI, error) {
	scheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(v1alpha1.AddToScheme(scheme))
	cl, err := client.New(config.GetConfigOrDie(), client.Options{Scheme: scheme})
	if err != nil {
		return nil, err
	}
	namespace := "seldon-mesh"
	inferIP, err := getSeldonMeshIP(cl, namespace)
	if err != nil {
		return nil, err
	}
	ic, err := NewSeldonInferAPI(fmt.Sprintf("%s:80", inferIP))
	if err != nil {
		return nil, err
	}
	return &SeldonK8sAPI{
		namespace:   namespace,
		client:      cl,
		inferClient: ic,
	}, nil
}

func getSeldonMeshIP(client client.Client, namespace string) (string, error) {
	svc := v1.Service{}
	err := client.Get(context.Background(), types.NamespacedName{Name: "seldon-mesh", Namespace: namespace}, &svc)
	if err != nil {
		return "", err
	}
	return svc.Status.LoadBalancer.Ingress[0].IP, nil
}

func (k *SeldonK8sAPI) Load(filename string) error {
	resource, err := getResource(filename)
	if err != nil {
		return err
	}
	resource.obj.SetNamespace(k.namespace)
	return k.client.Create(context.Background(), resource.obj)
}

func (k *SeldonK8sAPI) UnLoad(filename string) error {
	resource, err := getResource(filename)
	if err != nil {
		return err
	}
	resource.obj.SetNamespace(k.namespace)
	return k.client.Delete(context.Background(), resource.obj)
}

func (k *SeldonK8sAPI) IsLoaded(filename string) (bool, error) {
	resource, err := getResource(filename)
	if err != nil {
		return false, err
	}
	resource.obj.SetNamespace(k.namespace)
	err = k.client.Get(context.Background(), types.NamespacedName{Name: resource.name, Namespace: k.namespace}, resource.obj)
	if err != nil {
		return false, err
	}
	switch resource.gvk.Kind {
	case resourceModelKind:
		return resource.obj.(*v1alpha1.Model).Status.IsReady(), nil
	case resourcePipelineKind:
		return resource.obj.(*v1alpha1.Pipeline).Status.IsReady(), nil
	case resourceExperimentKind:
		return resource.obj.(*v1alpha1.Experiment).Status.IsReady(), nil
	case resourceServerKind:
		return resource.obj.(*v1alpha1.Server).Status.IsReady(), nil
	default:
		return false, fmt.Errorf("Unknown resource type in %s found %s", filename, resource.gvk.String())
	}
}

func (s *SeldonK8sAPI) Infer(filename string, request string) error {
	return s.inferClient.Infer(filename, request)
}
