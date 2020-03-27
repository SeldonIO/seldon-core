package client

import (
	"fmt"
	"github.com/go-logr/logr"
	"github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	clientset "github.com/seldonio/seldon-core/operator/client/machinelearning/v1/clientset/versioned/typed/machinelearning/v1"
	v1meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"path/filepath"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

type SeldonDeploymentClient struct {
	client *clientset.MachinelearningV1Client
	Log    logr.Logger
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

func (sd *SeldonDeploymentClient) GetPredictor(sdepName string, namespace string, predictorName string) (*v1.PredictorSpec, error) {
	sdep, err := sd.client.SeldonDeployments(namespace).Get(sdepName, v1meta.GetOptions{})
	if err != nil {
		return nil, err
	}
	for _, predictor := range sdep.Spec.Predictors {
		if predictor.Name == predictorName {
			return &predictor, nil
		}
	}

	return nil, fmt.Errorf("Failed to find predictor with name %s", predictorName)
}
