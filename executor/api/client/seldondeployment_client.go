package client

import (
	"github.com/go-logr/logr"
	"github.com/seldonio/seldon-core/executor/api/machinelearning/v1alpha2"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "github.com/seldonio/seldon-core/executor/client/clientset/versioned/typed/machinelearning/v1alpha2"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/homedir"
	"path/filepath"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

type SeldonDeploymentClient struct {
	client *clientset.MachinelearningV1alpha2Client
	Log logr.Logger
}


func NewSeldonDeploymentClient(path *string) *SeldonDeploymentClient {

	var config *rest.Config
	var err error
	if path != nil && *path != "" {
		config, err = clientcmd.BuildConfigFromFlags("", *path)
		if err != nil {
			panic(err.Error())
		}
	} else {
		config, err = rest.InClusterConfig()
		if err != nil {
			if home := homedir.HomeDir(); home != "" {
				homepath := filepath.Join(home, ".kube", "config")
				config, err = clientcmd.BuildConfigFromFlags("", homepath)
				if err != nil {
					panic(err.Error())
				}
			}
		}
	}

	kubeClientset, err := clientset.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}


	return &SeldonDeploymentClient{
		kubeClientset,
		logf.Log.WithName("SeldonRestApi"),
	}
}


func (sd *SeldonDeploymentClient) Get(name string, namespace string) (*v1alpha2.SeldonDeployment, error) {
	return sd.client.SeldonDeployments(namespace).Get(name,v1.GetOptions{})
}