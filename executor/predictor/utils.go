package predictor

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/ghodss/yaml"
	v1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	"io/ioutil"
	"os"
	"strings"
)

func GetPredictor(predictorName, filename, sdepName, namespace string, configPath *string) (*v1.PredictorSpec, error) {
	if filename != "" {
		predictor, err := getPredictorFromFile(predictorName, filename)
		if err != nil {
			return nil, err
		} else {
			return predictor, nil
		}
	} else {
		return getPredictorFromEnv()
	}
}

func getPredictorFromEnv() (*v1.PredictorSpec, error) {
	b64Predictor := os.Getenv("ENGINE_PREDICTOR")
	if b64Predictor != "" {
		bytes, err := base64.StdEncoding.DecodeString(b64Predictor)
		if err != nil {
			return nil, err
		}
		predictor := v1.PredictorSpec{}
		if err := json.Unmarshal(bytes, &predictor); err != nil {
			return nil, err
		} else {
			return &predictor, nil
		}
	} else {
		return nil, nil
	}
}

func getPredictorFromFile(predictorName string, filename string) (*v1.PredictorSpec, error) {
	dat, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	if strings.HasSuffix(filename, "yaml") {
		var sdep v1.SeldonDeployment
		err = yaml.Unmarshal(dat, &sdep)
		if err != nil {
			return nil, err
		}
		for _, predictor := range sdep.Spec.Predictors {
			if predictor.Name == predictorName {
				return &predictor, nil
			}
		}
		return nil, fmt.Errorf("Predictor not found %s", predictorName)
	} else {
		return nil, fmt.Errorf("Unsupported file type %s", filename)
	}
}
