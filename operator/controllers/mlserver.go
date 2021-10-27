package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/seldonio/seldon-core/operator/utils"

	machinelearningv1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	"github.com/seldonio/seldon-core/operator/constants"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	MLServerSKLearnImplementation      = "mlserver_sklearn.SKLearnModel"
	MLServerXGBoostImplementation      = "mlserver_xgboost.XGBoostModel"
	MLServerTempoImplementation        = "tempo.mlserver.InferenceRuntime"
	MLServerMLFlowImplementation       = "mlserver_mlflow.MLflowRuntime"
	MLServerAlibiExplainImplementation = "mlserver_alibi_explain.AlibiExplainRuntime"

	MLServerHTTPPortEnv            = "MLSERVER_HTTP_PORT"
	MLServerGRPCPortEnv            = "MLSERVER_GRPC_PORT"
	MLServerModelNameEnv           = "MLSERVER_MODEL_NAME"
	MLServerModelImplementationEnv = "MLSERVER_MODEL_IMPLEMENTATION"
	MLServerModelURIEnv            = "MLSERVER_MODEL_URI"
	MLServerTempoRuntimeEnv        = "TEMPO_RUNTIME_OPTIONS"
	MLServerModelExtraEnv          = "MLSERVER_MODEL_EXTRA"
)

var (
	ExplainerTypeToMLServerExplainerType = map[machinelearningv1.AlibiExplainerType]string{
		machinelearningv1.AlibiAnchorsTabularExplainer:      "anchor_tabular",
		machinelearningv1.AlibiAnchorsImageExplainer:        "anchor_image",
		machinelearningv1.AlibiAnchorsTextExplainer:         "anchor_text",
		machinelearningv1.AlibiCounterfactualsExplainer:     "counterfactuals",
		machinelearningv1.AlibiContrastiveExplainer:         "contrastive",
		machinelearningv1.AlibiKernelShapExplainer:          "kernel_shap",
		machinelearningv1.AlibiIntegratedGradientsExplainer: "integrated_gradients",
		machinelearningv1.AlibiALEExplainer:                 "ALE",
		machinelearningv1.AlibiTreeShap:                     "tree_shap",
	}
)

func mergeMLServerContainer(existing *v1.Container, mlServer *v1.Container) *v1.Container {
	if mlServer == nil {
		// Nothing to merge.
		return existing
	}
	if existing == nil {
		existing = &v1.Container{}
	}
	// Overwrite core items if not existing or required
	if existing.Image == "" {
		existing.Image = mlServer.Image
	}

	if existing.Env == nil {
		existing.Env = []v1.EnvVar{}
	}

	for _, envVar := range existing.Env {
		if utils.HasEnvVar(mlServer.Env, envVar.Name) {
			mlServer.Env = utils.SetEnvVar(mlServer.Env, envVar)
		} else {
			mlServer.Env = append(mlServer.Env, envVar)
		}
	}
	existing.Env = mlServer.Env

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

func getMLServerContainer(pu *machinelearningv1.PredictiveUnit, namespace string) (*v1.Container, error) {
	if pu == nil {
		return nil, errors.New("received nil predictive unit")
	}
	image, err := getMLServerImage(pu)
	if err != nil {
		return nil, err
	}

	envVars, err := getMLServerEnvVars(pu, namespace)
	if err != nil {
		return nil, err
	}

	httpPort := pu.Endpoint.HttpPort
	grpcPort := pu.Endpoint.GrpcPort

	cServer := &v1.Container{
		Name:  pu.Name,
		Image: image,
		Env:   envVars,
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
	if pu == nil {
		return "", errors.New("received nil predictive unit")
	}
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

func getMLServerEnvVars(pu *machinelearningv1.PredictiveUnit, namespace string) ([]v1.EnvVar, error) {
	if pu == nil {
		return nil, errors.New("received nil predictive unit")
	}
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
		{
			Name:  MLServerTempoRuntimeEnv,
			Value: fmt.Sprintf("{\"k8s_options\": {\"defaultRuntime\": \"tempo.seldon.SeldonKubernetesRuntime\", \"namespace\": \"%s\"}}", namespace),
		},
	}, nil
}

func getMLServerModelImplementation(pu *machinelearningv1.PredictiveUnit) (string, error) {
	if pu == nil {
		return "", errors.New("received nil predictive unit")
	}
	switch *pu.Implementation {
	case machinelearningv1.PrepackSklearnName:
		return MLServerSKLearnImplementation, nil
	case machinelearningv1.PrepackXgboostName:
		return MLServerXGBoostImplementation, nil
	case machinelearningv1.PrepackTempoName:
		return MLServerTempoImplementation, nil
	case machinelearningv1.PrepackMlflowName:
		return MLServerMLFlowImplementation, nil
	default:
		return "", nil
	}
}

func getAlibiExplainExplainerTypeTag(explainerType machinelearningv1.AlibiExplainerType) (string, error) {
	return ExplainerTypeToMLServerExplainerType[explainerType], nil
}

func getAlibiExplainExtraEnvVars(explainerType machinelearningv1.AlibiExplainerType, pSvcEndpoint string, graphName string) (string, error) {
	explainerTypeTag, err := getAlibiExplainExplainerTypeTag(explainerType)
	if err != nil {
		return "", err
	}
	explain_env_map := map[string]string{
		"explainer_type": explainerTypeTag,
		"infer_uri":      "http://" + pSvcEndpoint + "/v2/models/" + graphName + "/infer",
	}
	explain_env_json, err := json.Marshal(explain_env_map)
	if err != nil {
		return "", err
	}
	return string(explain_env_json), nil
}

func getAlibiExplainEnvVars(httpPortNum int, explainerModelName string, explainerType machinelearningv1.AlibiExplainerType, pSvcEndpoint string, graphName string) ([]v1.EnvVar, error) {
	explain_extra_env, err := getAlibiExplainExtraEnvVars(explainerType, pSvcEndpoint, graphName)
	if err != nil {
		return nil, err
	}
	alibiEnvs := []v1.EnvVar{
		{
			Name:  MLServerHTTPPortEnv,
			Value: strconv.Itoa(httpPortNum),
		},
		// note: we skip grpc port settings, relying on mlserver default
		// TODO: add gprc port
		{
			Name:  MLServerModelImplementationEnv,
			Value: MLServerAlibiExplainImplementation,
		},
		{
			Name:  MLServerModelNameEnv,
			Value: explainerModelName,
		},
		{
			Name:  MLServerModelURIEnv,
			Value: DefaultModelLocalMountPath,
		},
		{
			// TODO2: nested dict for explain_init settings?
			Name:  MLServerModelExtraEnv,
			Value: explain_extra_env,
		},
	}
	return alibiEnvs, nil

}
