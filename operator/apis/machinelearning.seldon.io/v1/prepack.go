package v1

import (
	"context"
	"encoding/json"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	k8types "k8s.io/apimachinery/pkg/types"
	"os"
)

const (
	EnvSklearnServerImageRestRelated = "RELATED_IMAGE_SKLEARNSERVER_REST"
	EnvSklearnserverImageGrpcRelated = "RELATED_IMAGE_SKLEARNSERVER_GRPC"
	EnvXgboostserverImageRestRelated = "RELATED_IMAGE_XGBOOSTSERVER_REST"
	EnvXgboostserverImageGrpcRelated = "RELATED_IMAGE_XGBOOSTSERVER_GRPC"
	EnvMlflowserverImageRestRelated  = "RELATED_IMAGE_MLFLOWSERVER_REST"
	EnvMlflowserverImageGrpcRelated  = "RELATED_IMAGE_MLFLOWSERVER_GRPC"
	EnvTensorflowImageRelated        = "RELATED_IMAGE_TENSORFLOW"
	EnvTfproxyImageRestRelated       = "RELATED_IMAGE_TFPROXY_REST"
	EnvTfproxyImageGrpcRelated       = "RELATED_IMAGE_TFPROXY_GRPC"
	PrepackTensorflowName            = "TENSORFLOW_SERVER"
	PrepackSklearnName               = "SKLEARN_SERVER"
	PrepackXgboostName               = "XGBOOST_SERVER"
	PrepackMlflowName                = "MLFLOW_SERVER"
	PrepackTritonName                = "TRITON_SERVER"
)

const PredictorServerConfigMapKeyName = "predictor_servers"

type PredictorImageConfig struct {
	ContainerImage      string `json:"image"`
	DefaultImageVersion string `json:"defaultImageVersion"`
}

type PredictorServerConfig struct {
	Tensorflow      bool                     `json:"tensorflow,omitempty"`
	TensorflowImage string                   `json:"tfImage,omitempty"`
	RestConfig      PredictorImageConfig     `json:"rest,omitempty"`
	GrpcConfig      PredictorImageConfig     `json:"grpc,omitempty"`
	Protocols       PredictorProtocolsConfig `json:"protocols,omitempty"`
}

type PredictorProtocolsConfig struct {
	Seldon     *PredictorImageConfig `json:"seldon,omitempty"`
	KFServing  *PredictorImageConfig `json:"kfserving,omitempty"`
	Tensorflow *PredictorImageConfig `json:"tensorflow,omitempty"`
}

var (
	ControllerConfigMapName          = "seldon-config"
	envSklearnServerRestImageRelated = os.Getenv(EnvSklearnServerImageRestRelated)
	envSklearnServerGrpcImageRelated = os.Getenv(EnvSklearnserverImageGrpcRelated)
	envXgboostServerRestImageRelated = os.Getenv(EnvXgboostserverImageRestRelated)
	envXgboostServerGrpcImageRelated = os.Getenv(EnvXgboostserverImageGrpcRelated)
	envMlflowServerRestImageRelated  = os.Getenv(EnvMlflowserverImageRestRelated)
	envMlflowServerGrpcImageRelated  = os.Getenv(EnvMlflowserverImageGrpcRelated)
	envTfserverServerImageRelated    = os.Getenv(EnvTensorflowImageRelated)
	envTfproxyServerRestImageRelated = os.Getenv(EnvTfproxyImageRestRelated)
	envTfproxyServerGrpcImageRelated = os.Getenv(EnvTfproxyImageGrpcRelated)
	relatedImageConfig               = map[string]PredictorServerConfig{}
)

func init() {
	if envSklearnServerRestImageRelated != "" {
		relatedImageConfig[PrepackSklearnName] = PredictorServerConfig{
			RestConfig: PredictorImageConfig{
				ContainerImage: envSklearnServerRestImageRelated,
			},
			GrpcConfig: PredictorImageConfig{
				ContainerImage: envSklearnServerGrpcImageRelated,
			},
		}
	}
	if envXgboostServerRestImageRelated != "" {
		relatedImageConfig[PrepackXgboostName] = PredictorServerConfig{
			RestConfig: PredictorImageConfig{
				ContainerImage: envXgboostServerRestImageRelated,
			},
			GrpcConfig: PredictorImageConfig{
				ContainerImage: envXgboostServerGrpcImageRelated,
			},
		}
	}
	if envMlflowServerRestImageRelated != "" {
		relatedImageConfig[PrepackMlflowName] = PredictorServerConfig{
			RestConfig: PredictorImageConfig{
				ContainerImage: envMlflowServerRestImageRelated,
			},
			GrpcConfig: PredictorImageConfig{
				ContainerImage: envMlflowServerGrpcImageRelated,
			},
		}
	}
	if envTfserverServerImageRelated != "" {
		relatedImageConfig[PrepackTensorflowName] = PredictorServerConfig{
			Tensorflow:      true,
			TensorflowImage: envTfserverServerImageRelated,
			RestConfig: PredictorImageConfig{
				ContainerImage: envTfproxyServerRestImageRelated,
			},
			GrpcConfig: PredictorImageConfig{
				ContainerImage: envTfproxyServerGrpcImageRelated,
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

func SetImageNameForPrepackContainer(pu *PredictiveUnit, c *corev1.Container, serverConfig *PredictorServerConfig) {
	// Add image: ignore version if empty
	if c.Image == "" {
		if pu.Endpoint.Type == REST {
			c.Image = serverConfig.RestConfig.ContainerImage
			if serverConfig.RestConfig.DefaultImageVersion != "" {
				c.Image = c.Image + ":" + serverConfig.RestConfig.DefaultImageVersion
			}
		} else {
			c.Image = serverConfig.GrpcConfig.ContainerImage
			if serverConfig.GrpcConfig.DefaultImageVersion != "" {
				c.Image = c.Image + ":" + serverConfig.GrpcConfig.DefaultImageVersion
			}
		}
	}
}
