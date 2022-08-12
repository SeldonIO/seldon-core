package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	corev1 "k8s.io/api/core/v1"
	k8types "k8s.io/apimachinery/pkg/types"
)

const (
	EnvSklearnServerImageRelated = "RELATED_IMAGE_SKLEARNSERVER"
	EnvXgboostserverImageRelated = "RELATED_IMAGE_XGBOOSTSERVER"
	EnvMlflowserverImageRelated  = "RELATED_IMAGE_MLFLOWSERVER"
	EnvTensorflowImageRelated    = "RELATED_IMAGE_TENSORFLOW"
	EnvTfproxyImageRelated       = "RELATED_IMAGE_TFPROXY"
	PrepackTensorflowName        = "TENSORFLOW_SERVER"
	PrepackSklearnName           = "SKLEARN_SERVER"
	PrepackXgboostName           = "XGBOOST_SERVER"
	PrepackMlflowName            = "MLFLOW_SERVER"
	PrepackHuggingFaceName       = "HUGGINGFACE_SERVER"
	PrepackTritonName            = "TRITON_SERVER"
	PrepackTempoName             = "TEMPO_SERVER"
)

const PredictorServerConfigMapKeyName = "predictor_servers"

type PredictorImageConfig struct {
	ContainerImage      string `json:"image"`
	DefaultImageVersion string `json:"defaultImageVersion"`
}

type PredictorServerConfig struct {
	Protocols map[Protocol]PredictorImageConfig `json:"protocols"`
}

func (p *PredictorServerConfig) PrepackImageName(protocol Protocol, pu *PredictiveUnit) string {
	if string(protocol) == "" {
		protocol = ProtocolSeldon
	}
	imageConfig := p.PrepackImageConfig(protocol)

	if imageConfig == nil && protocol == ProtocolV2 {
		imageConfig = p.PrepackImageConfig(ProtocolKfserving)
	}

	if imageConfig == nil && protocol == ProtocolKfserving {
		imageConfig = p.PrepackImageConfig(ProtocolV2)
	}

	if imageConfig == nil {
		return ""
	}

	if imageConfig.DefaultImageVersion != "" {
		return fmt.Sprintf("%s:%s", imageConfig.ContainerImage, imageConfig.DefaultImageVersion)
	}

	return imageConfig.ContainerImage
}

func (p *PredictorServerConfig) PrepackImageConfig(protocol Protocol) *PredictorImageConfig {
	if im, ok := p.Protocols[protocol]; ok {
		return &im //do something here
	} else {
		return nil
	}
}

type PredictorProtocolsConfig struct {
	Seldon     *PredictorImageConfig `json:"seldon,omitempty"`
	KFServing  *PredictorImageConfig `json:"kfserving,omitempty"`
	Tensorflow *PredictorImageConfig `json:"tensorflow,omitempty"`
}

var (
	ControllerConfigMapName       = "seldon-config"
	envSklearnServerImageRelated  = os.Getenv(EnvSklearnServerImageRelated)
	envXgboostServerImageRelated  = os.Getenv(EnvXgboostserverImageRelated)
	envMlflowServerImageRelated   = os.Getenv(EnvMlflowserverImageRelated)
	envTfserverServerImageRelated = os.Getenv(EnvTensorflowImageRelated)
	envTfproxyServerImageRelated  = os.Getenv(EnvTfproxyImageRelated)
	relatedImageConfig            = map[string]PredictorServerConfig{}
)

func init() {
	if envSklearnServerImageRelated != "" {
		relatedImageConfig[PrepackSklearnName] = PredictorServerConfig{
			Protocols: map[Protocol]PredictorImageConfig{
				ProtocolSeldon: {
					ContainerImage: envSklearnServerImageRelated,
				},
			},
		}
	}
	if envXgboostServerImageRelated != "" {
		relatedImageConfig[PrepackXgboostName] = PredictorServerConfig{
			Protocols: map[Protocol]PredictorImageConfig{
				ProtocolSeldon: {
					ContainerImage: envXgboostServerImageRelated,
				},
			},
		}
	}
	if envMlflowServerImageRelated != "" {
		relatedImageConfig[PrepackMlflowName] = PredictorServerConfig{
			Protocols: map[Protocol]PredictorImageConfig{
				ProtocolSeldon: {
					ContainerImage: envMlflowServerImageRelated,
				},
			},
		}
	}
	if envTfserverServerImageRelated != "" {
		relatedImageConfig[PrepackTensorflowName] = PredictorServerConfig{
			Protocols: map[Protocol]PredictorImageConfig{
				ProtocolTensorflow: {
					ContainerImage: envTfserverServerImageRelated,
				},
				ProtocolSeldon: {
					ContainerImage: envTfproxyServerImageRelated,
				},
			},
		}
	}
}

func IsPrepack(pu *PredictiveUnit) bool {
	isPrepack := len(*pu.Implementation) > 0 && *pu.Implementation != SIMPLE_MODEL && *pu.Implementation != SIMPLE_ROUTER && *pu.Implementation != RANDOM_ABTEST && *pu.Implementation != AVERAGE_COMBINER && *pu.Implementation != UNKNOWN_IMPLEMENTATION
	return isPrepack
}

func getPredictorServerConfigs() (map[string]PredictorServerConfig, error) {
	configMap := &corev1.ConfigMap{}

	err := C.Get(context.TODO(), k8types.NamespacedName{Name: ControllerConfigMapName, Namespace: ControllerNamespace}, configMap)

	if err != nil {
		fmt.Println("Failed to find config map " + ControllerConfigMapName)
		fmt.Println(err)
		return map[string]PredictorServerConfig{}, err
	}
	return getPredictorServerConfigsFromMap(configMap)
}

func getPredictorServerConfigsFromMap(configMap *corev1.ConfigMap) (map[string]PredictorServerConfig, error) {
	predictorServerConfig := make(map[string]PredictorServerConfig)
	if predictorConfig, ok := configMap.Data[PredictorServerConfigMapKeyName]; ok {
		err := json.Unmarshal([]byte(predictorConfig), &predictorServerConfig)
		if err != nil {
			panic(fmt.Errorf("Unable to unmarshall %v json string due to %v ", PredictorServerConfigMapKeyName, err))
		}
	}

	return predictorServerConfig, nil
}

func getPrepackServerConfigWithRelated(serverName string, relatedImages map[string]PredictorServerConfig) *PredictorServerConfig {
	//Use related images if present
	if val, ok := relatedImages[serverName]; ok {
		return &val
	}

	ServersConfigs, err := getPredictorServerConfigs()
	if err != nil {
		seldondeploymentlog.Error(err, "Failed to read prepacked model servers from configmap")
		return nil
	}
	ServerConfig, ok := ServersConfigs[serverName]
	if !ok {
		seldondeploymentlog.Error(nil, "No entry in predictors map for "+serverName)
		return nil
	}
	return &ServerConfig
}

func GetPrepackServerConfig(serverName string) *PredictorServerConfig {
	return getPrepackServerConfigWithRelated(serverName, relatedImageConfig)
}
