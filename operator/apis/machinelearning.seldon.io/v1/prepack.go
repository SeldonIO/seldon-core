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
	EnvRelatedImageSklearnServer   = "RELATED_IMAGE_SKLEARNSERVER"
	EnvRelatedImageXGBoostServer   = "RELATED_IMAGE_XGBOOSTSERVER"
	EnvRelatedImageMLFlowServer    = "RELATED_IMAGE_MLFLOWSERVER"
	EnvRelatedImageSklearnServerV2 = "RELATED_IMAGE_SKLEARNSERVER_V2"
	EnvRelatedImageXGBoostServerV2 = "RELATED_IMAGE_XGBOOSTSERVER_V2"
	EnvRelatedImageMLFlowServerV2  = "RELATED_IMAGE_MLFLOWSERVER_V2"
	EnvRelatedImageTensorflow      = "RELATED_IMAGE_TENSORFLOW"
	EnvRelatedImageTFProxy         = "RELATED_IMAGE_TFPROXY"
)

const (
	PrepackTensorflowName  = "TENSORFLOW_SERVER"
	PrepackSklearnName     = "SKLEARN_SERVER"
	PrepackXGBoostName     = "XGBOOST_SERVER"
	PrepackMLFlowName      = "MLFLOW_SERVER"
	PrepackHuggingFaceName = "HUGGINGFACE_SERVER"
	PrepackTritonName      = "TRITON_SERVER"
	PrepackTempoName       = "TEMPO_SERVER"
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
		imageConfig = p.PrepackImageConfig(ProtocolKFServing)
	}

	if imageConfig == nil && protocol == ProtocolKFServing {
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
	ControllerConfigMapName = "seldon-config"
	envSklearnServer        = os.Getenv(EnvRelatedImageSklearnServer)
	envXGBoostServer        = os.Getenv(EnvRelatedImageXGBoostServer)
	envMLFlowServer         = os.Getenv(EnvRelatedImageMLFlowServer)
	envSklearnServerV2      = os.Getenv(EnvRelatedImageSklearnServerV2)
	envXGBoostServerV2      = os.Getenv(EnvRelatedImageXGBoostServerV2)
	envMLFlowServerV2       = os.Getenv(EnvRelatedImageMLFlowServerV2)
	envTFServerServer       = os.Getenv(EnvRelatedImageTensorflow)
	envTFProxyServer        = os.Getenv(EnvRelatedImageTFProxy)
	relatedImageConfig      = map[string]PredictorServerConfig{}
)

func init() {
	relatedImageConfig = getRelatedImageConfig()
}

func getRelatedImageConfig() map[string]PredictorServerConfig {
	return map[string]PredictorServerConfig{
		PrepackSklearnName: {
			Protocols: map[Protocol]PredictorImageConfig{
				ProtocolSeldon: {
					ContainerImage: envSklearnServer,
				},
				ProtocolV2: {
					ContainerImage: envSklearnServerV2,
				},
			},
		},
		PrepackXGBoostName: {
			Protocols: map[Protocol]PredictorImageConfig{
				ProtocolSeldon: {
					ContainerImage: envXGBoostServer,
				},
				ProtocolV2: {
					ContainerImage: envXGBoostServerV2,
				},
			},
		},
		PrepackMLFlowName: {
			Protocols: map[Protocol]PredictorImageConfig{
				ProtocolSeldon: {
					ContainerImage: envMLFlowServer,
				},
				ProtocolV2: {
					ContainerImage: envMLFlowServerV2,
				},
			},
		},
		envTFServerServer: {
			Protocols: map[Protocol]PredictorImageConfig{
				ProtocolTensorflow: {
					ContainerImage: envTFServerServer,
				},
				ProtocolSeldon: {
					ContainerImage: envTFProxyServer,
				},
			},
		},
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
	// Get Server Config
	ServersConfigs, err := getPredictorServerConfigs()
	if err != nil {
		seldondeploymentLog.Error(err, "Failed to read prepacked model servers from configmap")
		return nil
	}
	ServerConfig, ok := ServersConfigs[serverName]
	if !ok {
		seldondeploymentLog.Error(nil, "No entry in predictors map for "+serverName)
		return nil
	}

	// Set related images if present
	RelatedServerConfig, present := relatedImages[serverName]
	if !present {
		return &ServerConfig
	}

	// If PredictorImageConfig.ContainerImage is not an empty string overwrite whole
	// PredictorImageConfig as this is expected by callers of this function.
	for protocol, imageConfig := range RelatedServerConfig.Protocols {
		if imageConfig.ContainerImage != "" {
			ServerConfig.Protocols[protocol] = imageConfig
		}
	}

	return &ServerConfig
}

func GetPrepackServerConfig(serverName string) *PredictorServerConfig {
	return getPrepackServerConfigWithRelated(serverName, relatedImageConfig)
}
