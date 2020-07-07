package client

import (
	"fmt"

	"github.com/go-logr/logr"
	v1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	clientset "github.com/seldonio/seldon-core/operator/client/machinelearning.seldon.io/v1/clientset/versioned/typed/machinelearning.seldon.io/v1"
	v1meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type SeldonDeploymentClient struct {
	client *clientset.MachinelearningV1Client
	Log    logr.Logger
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
