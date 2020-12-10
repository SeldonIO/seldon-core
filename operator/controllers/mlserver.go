package controllers

import (
	"fmt"
	"strconv"

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
	MLServerModelNameEnv           = "MLSERVER_MODEL_NAME"
	MLServerModelImplementationEnv = "MLSERVER_MODEL_IMPLEMENTATION"
	MLServerModelURIEnv            = "MLSERVER_MODEL_URI"
)

func mergeMLServerContainer(existing *v1.Container, mlServer *v1.Container) *v1.Container {
	// Overwrite core items if not existing or required
	if existing.Image == "" {
		existing.Image = mlServer.Image
	}

	if existing.Args == nil {
		existing.Args = mlServer.Args
	}

	if existing.Env == nil {
		existing.Env = []v1.EnvVar{}
	}

	// TODO: Allow overriding some of the env vars
	existing.Env = append(existing.Env, mlServer.Env...)

	if existing.ReadinessProbe == nil {
		existing.ReadinessProbe = mlServer.ReadinessProbe
	}

	if existing.LivenessProbe == nil {
		existing.LivenessProbe = mlServer.LivenessProbe
	}

	if existing.SecurityContext == nil {
		existing.SecurityContext = mlServer.SecurityContext
	}

	// Ports always overwritten
	// Need to look as we seem to add metrics ports automatically which mean this needs to be done
	existing.Ports = mlServer.Ports

	return existing
}

func getMLServerContainer(pu *machinelearningv1.PredictiveUnit) (*v1.Container, error) {
	image, err := getMLServerImage(pu)
	if err != nil {
		return nil, err
	}

	envVars, err := getMLServerEnvVars(pu)
	if err != nil {
		return nil, err
	}

	httpPort := pu.Endpoint.HttpPort
	grpcPort := pu.Endpoint.GrpcPort

	cServer := &v1.Container{
		Name:  pu.Name,
		Image: image,
		Args: []string{
			"mlserver",
			"start",
			DefaultModelLocalMountPath,
		},
		Env: envVars,
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

	return cServer, nil
}

func getMLServerImage(pu *machinelearningv1.PredictiveUnit) (string, error) {
	prepackConfig := machinelearningv1.GetPrepackServerConfig(string(*pu.Implementation))
	if prepackConfig == nil {
		return "", fmt.Errorf("failed to get server config for %s", *pu.Implementation)
	}

	if kfservingConfig, ok := prepackConfig.Protocols[machinelearningv1.ProtocolKfserving]; ok {
		// Ignore version if empty
		image := kfservingConfig.ContainerImage
		if kfservingConfig.DefaultImageVersion != "" {
			image = fmt.Sprintf("%s:%s", image, kfservingConfig.DefaultImageVersion)
		}

		return image, nil
	} else {
		err := fmt.Errorf("no image compatible with kfserving protocol for %s", *pu.Implementation)
		return "", err
	}
}

func getMLServerEnvVars(pu *machinelearningv1.PredictiveUnit) ([]v1.EnvVar, error) {
	httpPort := pu.Endpoint.HttpPort
	grpcPort := pu.Endpoint.GrpcPort

	mlServerModelImplementation, err := getMLServerModelImplementation(pu)
	if err != nil {
		return nil, err
	}

	return []v1.EnvVar{
		{
			Name:  MLServerHTTPPortEnv,
			Value: strconv.Itoa(int(httpPort)),
		},
		{
			Name:  MLServerGRPCPortEnv,
			Value: strconv.Itoa(int(grpcPort)),
		},
		{
			Name:  MLServerModelImplementationEnv,
			Value: mlServerModelImplementation,
		},
		{
			Name:  MLServerModelNameEnv,
			Value: pu.Name,
		},
		{
			// TODO: Should we make version optional in MLServer?
			Name:  "MLSERVER_MODEL_VERSION",
			Value: "v1",
		},
		{
			Name:  MLServerModelURIEnv,
			Value: DefaultModelLocalMountPath,
		},
	}, nil
}

func getMLServerModelImplementation(pu *machinelearningv1.PredictiveUnit) (string, error) {
	switch *pu.Implementation {
	case machinelearningv1.PrepackSklearnName:
		return MLServerSKLearnImplementation, nil
	case machinelearningv1.PrepackXgboostName:
		return MLServerXGBoostImplementation, nil
	}

	err := fmt.Errorf("invalid implementation: %s", *pu.Implementation)
	return "", err
}
