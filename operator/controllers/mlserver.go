package controllers

import (
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
	MLServerHuggingFaceImplementation  = "mlserver_huggingface.HuggingFaceRuntime"
	MLServerSKLearnImplementation      = "mlserver_sklearn.SKLearnModel"
	MLServerXGBoostImplementation      = "mlserver_xgboost.XGBoostModel"
	MLServerTempoImplementation        = "tempo.mlserver.InferenceRuntime"
	MLServerMLFlowImplementation       = "mlserver_mlflow.MLflowRuntime"
	MLServerAlibiExplainImplementation = "mlserver_alibi_explain.AlibiExplainRuntime"

	MLServerParallelWorkersEnv         = "MLSERVER_PARALLEL_WORKERS"
	MLServerParallelWorkersEnvDefault  = "0"
	MLServerHuggingFaceCacheEnv        = "XDG_CACHE_HOME"
	MLServerHuggingFaceCacheEnvDefault = "/opt/mlserver"
	MLServerHTTPPortEnv                = "MLSERVER_HTTP_PORT"
	MLServerGRPCPortEnv                = "MLSERVER_GRPC_PORT"
	MLServerMetricsPortEnv             = "MLSERVER_METRICS_PORT"
	MLServerMetricsEndpointEnv         = "MLSERVER_METRICS_ENDPOINT"
	MLServerModelNameEnv               = "MLSERVER_MODEL_NAME"
	MLServerModelImplementationEnv     = "MLSERVER_MODEL_IMPLEMENTATION"
	MLServerModelURIEnv                = "MLSERVER_MODEL_URI"
	MLServerTempoRuntimeEnv            = "TEMPO_RUNTIME_OPTIONS"
	MLServerModelExtraEnv              = "MLSERVER_MODEL_EXTRA"
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
		mlServer.Env = utils.SetEnvVar(mlServer.Env, envVar, true)
	}
	existing.Env = mlServer.Env

	// If the readiness or liveness probe already exist, ensure the handler is
	// V2-compatible (otherwise, set all to default)
	if existing.ReadinessProbe == nil {
		existing.ReadinessProbe = mlServer.ReadinessProbe
	} else {
		existing.ReadinessProbe.ProbeHandler = mlServer.ReadinessProbe.ProbeHandler
	}

	if existing.LivenessProbe == nil {
		existing.LivenessProbe = mlServer.LivenessProbe
	} else {
		existing.LivenessProbe.ProbeHandler = mlServer.LivenessProbe.ProbeHandler
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
			ProbeHandler: v1.ProbeHandler{HTTPGet: &v1.HTTPGetAction{
				Path:   constants.KFServingProbeReadyPath,
				Port:   intstr.FromString("http"),
				Scheme: v1.URISchemeHTTP,
			}},
			InitialDelaySeconds: 20,
			PeriodSeconds:       5,
			FailureThreshold:    3,
			SuccessThreshold:    1,
			TimeoutSeconds:      60,
		},
		LivenessProbe: &v1.Probe{
			ProbeHandler: v1.ProbeHandler{HTTPGet: &v1.HTTPGetAction{
				Path:   constants.KFServingProbeLivePath,
				Port:   intstr.FromString("http"),
				Scheme: v1.URISchemeHTTP,
			}},
			InitialDelaySeconds: 20,
			PeriodSeconds:       5,
			FailureThreshold:    3,
			SuccessThreshold:    1,
			TimeoutSeconds:      60,
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

	if kfservingConfig, ok := prepackConfig.Protocols[machinelearningv1.ProtocolKFServing]; ok {
		// Ignore version if empty
		image := kfservingConfig.ContainerImage
		if kfservingConfig.DefaultImageVersion != "" {
			image = fmt.Sprintf("%s:%s", image, kfservingConfig.DefaultImageVersion)
		}

		return image, nil
	} else if v2Config, ok := prepackConfig.Protocols[machinelearningv1.ProtocolV2]; ok {
		// Ignore version if empty
		image := v2Config.ContainerImage
		if v2Config.DefaultImageVersion != "" {
			image = fmt.Sprintf("%s:%s", image, v2Config.DefaultImageVersion)
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

	envVars := []v1.EnvVar{
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
	}

	if *pu.Implementation == machinelearningv1.PrepackHuggingFaceName {

		huggingFaceDefaultEnvs := []v1.EnvVar{
			// Disable parallel workers by default until transformers working correctly in parallel inference
			{
				Name:  MLServerParallelWorkersEnv,
				Value: MLServerParallelWorkersEnvDefault,
			},
			// Ensure the cache folder is set to have write permissions for pretrained models
			{
				Name:  MLServerHuggingFaceCacheEnv,
				Value: MLServerHuggingFaceCacheEnvDefault,
			},
		}

		envVars = append(envVars, huggingFaceDefaultEnvs...)
	}

	return envVars, nil
}

func getMLServerModelImplementation(pu *machinelearningv1.PredictiveUnit) (string, error) {
	if pu == nil {
		return "", errors.New("received nil predictive unit")
	}
	switch *pu.Implementation {
	case machinelearningv1.PrepackSklearnName:
		return MLServerSKLearnImplementation, nil
	case machinelearningv1.PrepackXGBoostName:
		return MLServerXGBoostImplementation, nil
	case machinelearningv1.PrepackTempoName:
		return MLServerTempoImplementation, nil
	case machinelearningv1.PrepackMLFlowName:
		return MLServerMLFlowImplementation, nil
	case machinelearningv1.PrepackHuggingFaceName:
		return MLServerHuggingFaceImplementation, nil
	default:
		return "", nil
	}
}

func getAlibiExplainExplainerTypeTag(explainerType machinelearningv1.AlibiExplainerType) (string, error) {
	tag, ok := ExplainerTypeToMLServerExplainerType[explainerType]
	if ok {
		return tag, nil
	} else {
		return "", errors.New(string(explainerType) + " not supported")
	}
}

func wrapDoubleQuotes(str string) string {
	const escQuotes string = "\""
	return escQuotes + str + escQuotes
}
func getAlibiExplainExtraEnvVars(explainerType machinelearningv1.AlibiExplainerType, pSvcEndpoint string, graphName string, initParameters string) (string, error) {
	// we need to pack one big envVar for MLSERVER_MODEL_EXTRA that can contain nested json / dict
	explainerTypeTag, err := getAlibiExplainExplainerTypeTag(explainerType)
	if err != nil {
		return "", err
	}

	v2URI := "http://" + pSvcEndpoint + "/v2/models/" + graphName + "/infer"
	explainExtraEnv := "{" + wrapDoubleQuotes("explainer_type") + ":" + wrapDoubleQuotes(explainerTypeTag)
	explainExtraEnv = explainExtraEnv + "," + wrapDoubleQuotes("infer_uri") + ":" + wrapDoubleQuotes(v2URI)

	if initParameters != "" {
		//init parameters is passed as json string so we need to reconstruct the dictionary
		explainExtraEnv = explainExtraEnv + "," + wrapDoubleQuotes("init_parameters") + ":" + initParameters
	}

	// end
	explainExtraEnv = explainExtraEnv + "}"

	return explainExtraEnv, nil
}

func getAlibiExplainEnvVars(httpPortNum int, explainerModelName string, explainerType machinelearningv1.AlibiExplainerType, pSvcEndpoint string, graphName string, initParameters string) ([]v1.EnvVar, error) {
	explain_extra_env, err := getAlibiExplainExtraEnvVars(explainerType, pSvcEndpoint, graphName, initParameters)
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
			Name:  MLServerModelExtraEnv,
			Value: explain_extra_env,
		},
	}
	return alibiEnvs, nil

}
