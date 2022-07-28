/*
Copyright 2019 The Seldon Team.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	machinelearningv1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	"github.com/seldonio/seldon-core/operator/constants"
	"github.com/seldonio/seldon-core/operator/utils"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	ENV_DEFAULT_EXECUTOR_SERVER_PORT             = "EXECUTOR_SERVER_PORT"
	ENV_DEFAULT_EXECUTOR_SERVER_GRPC_PORT        = "EXECUTOR_SERVER_GRPC_PORT"
	ENV_DEFAULT_EXECUTOR_CPU_REQUEST             = "EXECUTOR_DEFAULT_CPU_REQUEST"
	ENV_DEFAULT_EXECUTOR_MEMORY_REQUEST          = "EXECUTOR_DEFAULT_MEMORY_REQUEST"
	ENV_DEFAULT_EXECUTOR_CPU_LIMIT               = "EXECUTOR_DEFAULT_CPU_LIMIT"
	ENV_DEFAULT_EXECUTOR_MEMORY_LIMIT            = "EXECUTOR_DEFAULT_MEMORY_LIMIT"
	ENV_EXECUTOR_METRICS_PORT_NAME               = "EXECUTOR_SERVER_METRICS_PORT_NAME"
	ENV_EXECUTOR_PROMETHEUS_PATH                 = "EXECUTOR_PROMETHEUS_PATH"
	ENV_EXECUTOR_REQUEST_LOGGER_WORK_QUEUE_SIZE  = "EXECUTOR_REQUEST_LOGGER_WORK_QUEUE_SIZE"
	ENV_EXECUTOR_REQUEST_LOGGER_WRITE_TIMEOUT_MS = "EXECUTOR_REQUEST_LOGGER_WRITE_TIMEOUT_MS"
	ENV_EXECUTOR_FULL_HEALTH_CHECKS              = "EXECUTOR_FULL_HEALTH_CHECKS"
	ENV_EXECUTOR_USER                            = "EXECUTOR_CONTAINER_USER"
	ENV_USE_EXECUTOR                             = "USE_EXECUTOR"

	DEFAULT_EXECUTOR_CONTAINER_PORT = 8000
	DEFAULT_EXECUTOR_GRPC_PORT      = 5001

	ENV_EXECUTOR_IMAGE         = "EXECUTOR_CONTAINER_IMAGE_AND_VERSION"
	ENV_EXECUTOR_IMAGE_RELATED = "RELATED_IMAGE_EXECUTOR" //RedHat specific
)

var (
	EngineContainerName     = "seldon-container-engine"
	envExecutorImage        = os.Getenv(ENV_EXECUTOR_IMAGE)
	envExecutorImageRelated = os.Getenv(ENV_EXECUTOR_IMAGE_RELATED)
	envExecutorUser         = os.Getenv(ENV_EXECUTOR_USER)
	envUseExecutor          = os.Getenv(ENV_USE_EXECUTOR)

	executorMetricsPortName = utils.GetEnv(ENV_EXECUTOR_METRICS_PORT_NAME, constants.DefaultMetricsPortName)

	executorDefaultCpuRequest       = utils.GetEnv(ENV_DEFAULT_EXECUTOR_CPU_REQUEST, constants.DefaultExecutorCpuRequest)
	executorDefaultCpuLimit         = utils.GetEnv(ENV_DEFAULT_EXECUTOR_CPU_LIMIT, constants.DefaultExecutorCpuLimit)
	executorDefaultMemoryRequest    = utils.GetEnv(ENV_DEFAULT_EXECUTOR_MEMORY_REQUEST, constants.DefaultExecutorMemoryRequest)
	executorDefaultMemoryLimit      = utils.GetEnv(ENV_DEFAULT_EXECUTOR_MEMORY_LIMIT, constants.DefaultExecutorMemoryLimit)
	executorReqLoggerWorkQueueSize  = utils.GetEnv(ENV_EXECUTOR_REQUEST_LOGGER_WORK_QUEUE_SIZE, constants.DefaultExecutorReqLoggerWorkQueueSize)
	executorReqLoggerWriteTimeoutMs = utils.GetEnv(ENV_EXECUTOR_REQUEST_LOGGER_WRITE_TIMEOUT_MS, constants.DefaultExecutorReqLoggerWriteTimeoutMs)
)

func addEngineToDeployment(mlDep *machinelearningv1.SeldonDeployment, p *machinelearningv1.PredictorSpec, engine_http_port int, engine_grpc_port int, pSvcName string, deploy *appsv1.Deployment) error {
	//check not already present
	for _, con := range deploy.Spec.Template.Spec.Containers {
		if strings.Compare(con.Name, EngineContainerName) == 0 {
			return nil
		}
	}
	engineContainer, err := createEngineContainer(mlDep, p, engine_http_port, engine_grpc_port)
	if err != nil {
		return err
	}
	deploy.Labels[machinelearningv1.Label_svc_orch] = "true"

	//downward api used to make pod info available to container
	volMount := false
	for _, vol := range engineContainer.VolumeMounts {
		if vol.Name == machinelearningv1.PODINFO_VOLUME_NAME {
			volMount = true
		}
	}
	if !volMount {
		engineContainer.VolumeMounts = append(engineContainer.VolumeMounts, corev1.VolumeMount{
			Name:      machinelearningv1.PODINFO_VOLUME_NAME,
			MountPath: machinelearningv1.PODINFO_VOLUME_PATH,
		})
	}

	deploy.Spec.Template.Spec.Containers = append(deploy.Spec.Template.Spec.Containers, *engineContainer)

	if deploy.Spec.Template.Annotations == nil {
		deploy.Spec.Template.Annotations = make(map[string]string)
	}
	//overwrite annotations with predictor annotations
	for _, ann := range p.Annotations {
		deploy.Spec.Template.Annotations[ann] = p.Annotations[ann]
	}

	deploy.ObjectMeta.Labels[machinelearningv1.Label_seldon_app] = pSvcName
	deploy.Spec.Selector.MatchLabels[machinelearningv1.Label_seldon_app] = pSvcName
	deploy.Spec.Template.ObjectMeta.Labels[machinelearningv1.Label_seldon_app] = pSvcName

	volFound := false
	for _, vol := range deploy.Spec.Template.Spec.Volumes {
		if vol.Name == machinelearningv1.PODINFO_VOLUME_NAME || vol.Name == machinelearningv1.OLD_PODINFO_VOLUME_NAME {
			volFound = true
		}
	}

	if !volFound {
		var defaultMode = corev1.DownwardAPIVolumeSourceDefaultMode
		//Add downwardAPI
		deploy.Spec.Template.Spec.Volumes = append(deploy.Spec.Template.Spec.Volumes, corev1.Volume{Name: machinelearningv1.PODINFO_VOLUME_NAME, VolumeSource: corev1.VolumeSource{
			DownwardAPI: &corev1.DownwardAPIVolumeSource{Items: []corev1.DownwardAPIVolumeFile{
				{Path: "annotations", FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.annotations", APIVersion: "v1"}}}, DefaultMode: &defaultMode}}})
	}

	return nil
}

func getExecutorHttpPort() (engine_http_port int, err error) {
	// Get engine http port from environment or use default
	engine_http_port = DEFAULT_EXECUTOR_CONTAINER_PORT
	var env_engine_http_port = utils.GetEnv(ENV_DEFAULT_EXECUTOR_SERVER_PORT, "")
	if env_engine_http_port != "" {
		engine_http_port, err = strconv.Atoi(env_engine_http_port)
		if err != nil {
			return 0, err
		}
	}
	return engine_http_port, nil
}

func getExecutorGrpcPort() (engine_grpc_port int, err error) {
	// Get engine grpc port from environment or use default
	engine_grpc_port = DEFAULT_EXECUTOR_GRPC_PORT
	var env_engine_grpc_port = utils.GetEnv(ENV_DEFAULT_EXECUTOR_SERVER_GRPC_PORT, "")
	if env_engine_grpc_port != "" {
		engine_grpc_port, err = strconv.Atoi(env_engine_grpc_port)
		if err != nil {
			return 0, err
		}
	}
	return engine_grpc_port, nil
}

func getPrometheusPath(mlDep *machinelearningv1.SeldonDeployment) string {
	prometheusPath := "/prometheus"
	prometheusPath = utils.GetEnv(ENV_EXECUTOR_PROMETHEUS_PATH, prometheusPath)
	return prometheusPath
}

func getSvcOrchSvcAccountName(mlDep *machinelearningv1.SeldonDeployment) string {
	svcAccount := "default"
	if svcAccountTmp, ok := os.LookupEnv("EXECUTOR_CONTAINER_SERVICE_ACCOUNT_NAME"); ok {
		svcAccount = svcAccountTmp
	}
	return svcAccount
}

func getSvcOrchUser(mlDep *machinelearningv1.SeldonDeployment) (*int64, error) {

	if envExecutorUser != "" {
		user, err := strconv.Atoi(envExecutorUser)
		if err != nil {
			return nil, err
		} else {
			engineUser := int64(user)
			return &engineUser, nil
		}
	}
	return nil, nil
}

func createExecutorContainer(mlDep *machinelearningv1.SeldonDeployment, p *machinelearningv1.PredictorSpec, predictorB64 string, http_port int, grpc_port int, resources *corev1.ResourceRequirements) (*corev1.Container, error) {
	protocol := mlDep.Spec.Protocol
	//Backwards compatibility for older resources
	if protocol == "" {
		protocol = machinelearningv1.ProtocolSeldon
	}

	transport := mlDep.Spec.Transport
	if transport == "" {
		transport = machinelearningv1.TransportRest
	}

	serverType := mlDep.Spec.ServerType
	if serverType == "" {
		serverType = machinelearningv1.ServerTypeRPC
	}

	// Get executor image from env vars in order of priority
	var executorImage string
	if executorImage = envExecutorImageRelated; executorImage == "" {
		if executorImage = envExecutorImage; executorImage == "" {
			return nil, fmt.Errorf("Failed to find executor image from environment. Check %s or %s are set.", ENV_EXECUTOR_IMAGE, ENV_EXECUTOR_IMAGE_RELATED)
		}
	}

	probeScheme := corev1.URISchemeHTTP
	if !utils.IsEmptyTLS(p) {
		probeScheme = corev1.URISchemeHTTPS
	}

	loggerQSize := getAnnotation(mlDep, machinelearningv1.ANNOTATION_LOGGER_WORK_QUEUE_SIZE, executorReqLoggerWorkQueueSize)
	_, err := strconv.Atoi(loggerQSize)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse %s as integer for %s. %w", loggerQSize, ENV_EXECUTOR_REQUEST_LOGGER_WORK_QUEUE_SIZE, err)
	}

	loggerWriteTimeout := getAnnotation(mlDep, machinelearningv1.ANNOTATION_LOGGER_WRITE_TIMEOUT_MS, executorReqLoggerWriteTimeoutMs)
	_, err = strconv.Atoi(loggerWriteTimeout)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse %s as integer for %s. %w", executorReqLoggerWriteTimeoutMs, ENV_EXECUTOR_REQUEST_LOGGER_WRITE_TIMEOUT_MS, err)
	}

	return &corev1.Container{
		Name:  EngineContainerName,
		Image: executorImage,
		Args: []string{
			"--sdep", mlDep.Name,
			"--namespace", mlDep.Namespace,
			"--predictor", p.Name,
			"--http_port", strconv.Itoa(http_port),
			"--grpc_port", strconv.Itoa(grpc_port),
			"--protocol", string(protocol),
			"--transport", string(transport),
			"--prometheus_path", getPrometheusPath(mlDep),
			"--server_type", string(serverType),
			"--log_work_buffer_size", loggerQSize,
			"--log_write_timeout_ms", loggerWriteTimeout,
			fmt.Sprintf("--full_health_checks=%s", utils.GetEnv(ENV_EXECUTOR_FULL_HEALTH_CHECKS, "false")),
		},
		ImagePullPolicy:          corev1.PullPolicy(utils.GetEnv("EXECUTOR_CONTAINER_IMAGE_PULL_POLICY", "IfNotPresent")),
		TerminationMessagePath:   "/dev/termination-log",
		TerminationMessagePolicy: corev1.TerminationMessageReadFile,
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      machinelearningv1.PODINFO_VOLUME_NAME,
				MountPath: machinelearningv1.PODINFO_VOLUME_PATH,
			},
		},
		Env: []corev1.EnvVar{
			{Name: "ENGINE_PREDICTOR", Value: predictorB64},
			{Name: "REQUEST_LOGGER_DEFAULT_ENDPOINT", Value: utils.GetEnv("EXECUTOR_REQUEST_LOGGER_DEFAULT_ENDPOINT", "http://default-broker")},
		},
		Ports: []corev1.ContainerPort{
			{ContainerPort: int32(http_port), Protocol: corev1.ProtocolTCP, Name: constants.HttpPortName},
			{ContainerPort: int32(http_port), Protocol: corev1.ProtocolTCP, Name: executorMetricsPortName},
			{ContainerPort: int32(grpc_port), Protocol: corev1.ProtocolTCP, Name: constants.GrpcPortName},
		},
		ReadinessProbe: &corev1.Probe{ProbeHandler: corev1.ProbeHandler{HTTPGet: &corev1.HTTPGetAction{Port: intstr.FromInt(http_port), Path: "/ready", Scheme: probeScheme}},
			InitialDelaySeconds: 20,
			PeriodSeconds:       5,
			FailureThreshold:    3,
			SuccessThreshold:    1,
			TimeoutSeconds:      60},
		LivenessProbe: &corev1.Probe{ProbeHandler: corev1.ProbeHandler{HTTPGet: &corev1.HTTPGetAction{Port: intstr.FromInt(http_port), Path: "/live", Scheme: probeScheme}},
			InitialDelaySeconds: 20,
			PeriodSeconds:       5,
			FailureThreshold:    3,
			SuccessThreshold:    1,
			TimeoutSeconds:      60},
		Resources: *resources,
	}, nil
}

// Create the Container for the service orchestrator.
func createEngineContainer(mlDep *machinelearningv1.SeldonDeployment, p *machinelearningv1.PredictorSpec, engine_http_port, engine_grpc_port int) (*corev1.Container, error) {
	// Get engine user
	engineUser, err := getSvcOrchUser(mlDep)
	if err != nil {
		return nil, err
	}
	// get predictor as base64 encoded json
	pCopy := p.DeepCopy()
	// Set traffic to zero to ensure this doesn't cause a diff in the resulting  deployment created
	pCopy.Traffic = 0
	// Set replicas to zero to ensure this doesn't cause a diff in the resulting deployment created
	var replicasCopy int32 = 0
	pCopy.Replicas = &replicasCopy
	predictorB64, err := getEngineVarJson(pCopy)
	if err != nil {
		return nil, err
	}

	//Engine resources
	var engineResources *corev1.ResourceRequirements = p.SvcOrchSpec.Resources
	if engineResources == nil {
		var cpu_request resource.Quantity
		var cpu_limit resource.Quantity
		var memory_request resource.Quantity
		var memory_limit resource.Quantity

		cpu_request = resource.MustParse(executorDefaultCpuRequest)
		cpu_limit = resource.MustParse(executorDefaultCpuLimit)
		memory_request = resource.MustParse(executorDefaultMemoryRequest)
		memory_limit = resource.MustParse(executorDefaultMemoryLimit)

		engineResources = &corev1.ResourceRequirements{
			Requests: map[corev1.ResourceName]resource.Quantity{
				corev1.ResourceCPU:    cpu_request,
				corev1.ResourceMemory: memory_request,
			},
			Limits: map[corev1.ResourceName]resource.Quantity{
				corev1.ResourceCPU:    cpu_limit,
				corev1.ResourceMemory: memory_limit,
			},
		}
	}

	var c *corev1.Container
	executor_http_port, err := getExecutorHttpPort()
	if err != nil {
		return nil, err
	}
	executor_grpc_port, err := getExecutorGrpcPort()
	if err != nil {
		return nil, err
	}
	c, err = createExecutorContainer(mlDep, p, predictorB64, executor_http_port, executor_grpc_port, engineResources)
	if err != nil {
		return nil, err
	}

	if engineUser != nil {
		c.SecurityContext = &corev1.SecurityContext{RunAsUser: engineUser}
	}

	// Environment vars if specified
	svcOrchEnvMap := make(map[string]string)
	if p.SvcOrchSpec.Env != nil {
		for _, env := range p.SvcOrchSpec.Env {
			c.Env = append(c.Env, *env)
			svcOrchEnvMap[env.Name] = env.Value
		}
	}

	engineEnvVarsFromAnnotations := getEngineEnvAnnotations(mlDep)
	for _, envVar := range engineEnvVarsFromAnnotations {
		//don't add env vars that are already present in svcOrchSpec
		if _, ok := svcOrchEnvMap[envVar.Name]; ok {
			//present so don't try to overwrite
		} else {
			c.Env = append(c.Env, envVar)
			// now put in map so we know it's there
			svcOrchEnvMap[envVar.Name] = envVar.Value
		}

	}

	return c, nil
}

// Create the service orchestrator.
func createEngineDeployment(mlDep *machinelearningv1.SeldonDeployment, p *machinelearningv1.PredictorSpec, seldonId string, engine_http_port, engine_grpc_port int) (*appsv1.Deployment, error) {

	var terminationGracePeriodSecs = int64(20)
	var defaultMode = corev1.DownwardAPIVolumeSourceDefaultMode

	depName := machinelearningv1.GetServiceOrchestratorName(mlDep, p)

	con, err := createEngineContainer(mlDep, p, engine_http_port, engine_grpc_port)
	if err != nil {
		return nil, err
	}
	deploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      depName,
			Namespace: getNamespace(mlDep),
			Labels: map[string]string{
				machinelearningv1.Label_svc_orch:   "true",
				machinelearningv1.Label_seldon_app: seldonId,
				machinelearningv1.Label_seldon_id:  seldonId,
				"app":                              depName,
				"version":                          "v1",
				"fluentd":                          "true",
			},
			Annotations: mlDep.Spec.Annotations,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{machinelearningv1.Label_seldon_app: seldonId, machinelearningv1.Label_seldon_id: seldonId},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{machinelearningv1.Label_seldon_app: seldonId, machinelearningv1.Label_seldon_id: seldonId, "app": depName},
					Annotations: map[string]string{
						"prometheus.io/path": getPrometheusPath(mlDep),
						// "prometheus.io/port":   strconv.Itoa(engine_http_port),
						"prometheus.io/scrape": "true",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						*con,
					},
					TerminationGracePeriodSeconds: &terminationGracePeriodSecs,
					DNSPolicy:                     corev1.DNSClusterFirst,
					SchedulerName:                 "default-scheduler",
					SecurityContext:               &corev1.PodSecurityContext{},
					Volumes: []corev1.Volume{
						{Name: machinelearningv1.PODINFO_VOLUME_NAME, VolumeSource: corev1.VolumeSource{DownwardAPI: &corev1.DownwardAPIVolumeSource{Items: []corev1.DownwardAPIVolumeFile{
							{Path: "annotations", FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.annotations", APIVersion: "v1"}},
						}, DefaultMode: &defaultMode}}},
					},
					RestartPolicy: corev1.RestartPolicyAlways,
				},
			},
			Strategy: appsv1.DeploymentStrategy{RollingUpdate: &appsv1.RollingUpdateDeployment{MaxUnavailable: &intstr.IntOrString{StrVal: "10%"}}},
		},
	}

	// Set replicas from more specific to more general settings in spec
	if p.SvcOrchSpec.Replicas != nil {
		deploy.Spec.Replicas = p.SvcOrchSpec.Replicas
	} else if p.Replicas != nil {
		deploy.Spec.Replicas = p.Replicas
	} else if mlDep.Spec.Replicas != nil {
		deploy.Spec.Replicas = mlDep.Spec.Replicas
	}

	// Add a particular service account rather than default for the engine
	svcAccountName := getSvcOrchSvcAccountName(mlDep)
	deploy.Spec.Template.Spec.ServiceAccountName = svcAccountName
	deploy.Spec.Template.Spec.DeprecatedServiceAccount = svcAccountName

	// add predictor labels
	for k, v := range p.Labels {
		deploy.ObjectMeta.Labels[k] = v
		deploy.Spec.Template.ObjectMeta.Labels[k] = v
	}
	return deploy, nil
}
