package controllers

import (
	"fmt"

	machinelearningv1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	"github.com/seldonio/seldon-core/operator/constants"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	MLServerSKLearnImplementation = "mlserver.models.SKLearnModel"
	MLServerXGBoostImplementation = "mlserver.models.XGBoostModel"

	MLServerHTTPPortEnv            = "MLSERVER_HTTP_PORT"
	MLServerGRPCPortEnv            = "MLSERVER_GRPC_PORT"
	MLServerModelImplementationEnv = "MLSERVER_MODEL_IMPLEMENTATION"
	MLServerModelURIEnv            = "MLSERVER_MODEL_URI"
)

func getMLServerContainer(pu *machinelearningv1.PredictiveUnit, serverConfig *machinelearningv1.PredictorServerConfig) *v1.Container {
	image := getMLServerImage(pu)
	httpPort := getMLServerPort(pu, machinelearningv1.REST)
	grpcPort := getMLServerPort(pu, machinelearningv1.GRPC)

	cServer := &v1.Container{
		Name:  pu.Name,
		Image: image,
		Args: []string{
			"mlserver",
			"start",
			DefaultModelLocalMountPath,
		},
		Ports: []v1.ContainerPort{
			{
				Name:          "grpc",
				ContainerPort: grpcPort,
				Protocol:      v1.ProtocolTCP,
			},
			{
				Name:          "http",
				ContainerPort: httpPort,
				Protocol:      v1.ProtocolTCP,
			},
		},
		ReadinessProbe: &v1.Probe{
			Handler: v1.Handler{HTTPGet: &v1.HTTPGetAction{
				Path: constants.KFServingProbeReadyPath,
				Port: intstr.FromString("http"),
			}},
			InitialDelaySeconds: 20,
			TimeoutSeconds:      1,
			PeriodSeconds:       10,
			SuccessThreshold:    1,
			FailureThreshold:    3,
		},
		LivenessProbe: &v1.Probe{
			Handler: v1.Handler{HTTPGet: &v1.HTTPGetAction{
				Path: constants.KFServingProbeLivePath,
				Port: intstr.FromString("http"),
			}},
			InitialDelaySeconds: 60,
			TimeoutSeconds:      1,
			PeriodSeconds:       10,
			SuccessThreshold:    1,
			FailureThreshold:    3,
		},
		VolumeMounts: []v1.VolumeMount{
			{
				Name:      machinelearningv1.PODINFO_VOLUME_NAME,
				MountPath: machinelearningv1.PODINFO_VOLUME_PATH,
			},
		},
	}

	return cServer
}

func getMLServerImage(pu *machinelearningv1.PredictiveUnit) string {
	prepackConfig := machinelearningv1.GetPrepackServerConfig(string(*pu.Implementation))
	kfservingConfig := prepackConfig.Protocols.KFServing

	if kfservingConfig == nil {
		// TODO: Raise error if empty (i.e. pre-packaged server is incompatible with protocol)
		return ""
	}

	// Ignore version if empty
	image := kfservingConfig.ContainerImage
	if kfservingConfig.DefaultImageVersion != "" {
		image = fmt.Sprintf("%s:%s", image, kfservingConfig.DefaultImageVersion)
	}

	return image
}

func getMLServerEnvVars(pu *machinelearningv1.PredictiveUnit) []v1.EnvVar {
	httpPort := getMLServerPort(pu, machinelearningv1.REST)
	grpcPort := getMLServerPort(pu, machinelearningv1.GRPC)

	return []v1.EnvVar{
		{
			Name:  MLServerHTTPPortEnv,
			Value: string(httpPort),
		},
		{
			Name:  MLServerGRPCPortEnv,
			Value: string(grpcPort),
		},
		{
			Name:  MLServerModelImplementationEnv,
			Value: getMLServerModelImplementation(pu),
		},
		{
			Name:  MLServerModelURIEnv,
			Value: DefaultModelLocalMountPath,
		},
	}
}

func getMLServerPort(pu *machinelearningv1.PredictiveUnit, endpointType machinelearningv1.EndpointType) int32 {
	if pu.Endpoint.Type == endpointType {
		return pu.Endpoint.ServicePort
	}

	// TODO: Error if something else
	switch endpointType {
	case machinelearningv1.REST:
		return constants.MLServerDefaultHttpPort
	case machinelearningv1.GRPC:
		return constants.MLServerDefaultGrpcPort
	}

	return 0
}

func getMLServerModelImplementation(pu *machinelearningv1.PredictiveUnit) string {
	switch *pu.Implementation {
	case machinelearningv1.PrepackSklearnName:
		return MLServerSKLearnImplementation
	case machinelearningv1.PrepackXgboostName:
		return MLServerXGBoostImplementation
	}

	// TODO: Error if something else
	return ""
}
