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
	"github.com/seldonio/seldon-core/operator/constants"
	"os"
	"strconv"
	"strings"

	machinelearningv1 "github.com/seldonio/seldon-core/operator/apis/machinelearning/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	ENV_DEFAULT_EXECUTOR_SERVER_PORT      = "EXECUTOR_SERVER_PORT"
	ENV_DEFAULT_EXECUTOR_SERVER_GRPC_PORT = "EXECUTOR_SERVER_GRPC_PORT"
	ENV_EXECUTOR_PROMETHEUS_PATH          = "EXECUTOR_PROMETHEUS_PATH"
	ENV_ENGINE_PROMETHEUS_PATH            = "ENGINE_PROMETHEUS_PATH"

	DEFAULT_EXECUTOR_CONTAINER_PORT = 8000
	DEFAULT_EXECUTOR_GRPC_PORT      = 5001
)

var (
	EngineContainerName = "seldon-container-engine"
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
		if vol.Name == machinelearningv1.PODINFO_VOLUME_NAME {
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
	var env_engine_http_port = GetEnv(ENV_DEFAULT_EXECUTOR_SERVER_PORT, "")
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
	var env_engine_grpc_port = GetEnv(ENV_DEFAULT_EXECUTOR_SERVER_GRPC_PORT, "")
	if env_engine_grpc_port != "" {
		engine_grpc_port, err = strconv.Atoi(env_engine_grpc_port)
		if err != nil {
			return 0, err
		}
	}
	return engine_grpc_port, nil
}

func isExecutorEnabled(mlDep *machinelearningv1.SeldonDeployment) bool {
	useExecutor := getAnnotation(mlDep, machinelearningv1.ANNOTATION_EXECUTOR, "false")
	useExecutorEnv := GetEnv("USE_EXECUTOR", "false")
	return useExecutor == "true" || useExecutorEnv == "true"
}

func getPrometheusPath(mlDep *machinelearningv1.SeldonDeployment) string {
	prometheusPath := "/prometheus"
	if isExecutorEnabled(mlDep) {
		prometheusPath = GetEnv(ENV_EXECUTOR_PROMETHEUS_PATH, prometheusPath)
	} else {
		prometheusPath = GetEnv(ENV_ENGINE_PROMETHEUS_PATH, prometheusPath)
	}
	return prometheusPath
}

func getSvcOrchSvcAccountName(mlDep *machinelearningv1.SeldonDeployment) string {
	svcAccount := "default"
	if isExecutorEnabled(mlDep) {
		if svcAccountTmp, ok := os.LookupEnv("EXECUTOR_CONTAINER_SERVICE_ACCOUNT_NAME"); ok {
			svcAccount = svcAccountTmp
		}
	} else {
		if svcAccountTmp, ok := os.LookupEnv("ENGINE_CONTAINER_SERVICE_ACCOUNT_NAME"); ok {
			svcAccount = svcAccountTmp
		}
	}
	return svcAccount
}

func getSvcOrchUser(mlDep *machinelearningv1.SeldonDeployment) (int64, error) {
	var engineUser int64 = -1
	if isExecutorEnabled(mlDep) {
		if engineUserEnv, ok := os.LookupEnv("EXECUTOR_CONTAINER_USER"); ok {
			user, err := strconv.Atoi(engineUserEnv)
			if err != nil {
				return -1, err
			} else {
				engineUser = int64(user)
			}
		}

	} else {
		if engineUserEnv, ok := os.LookupEnv("ENGINE_CONTAINER_USER"); ok {
			user, err := strconv.Atoi(engineUserEnv)
			if err != nil {
				return -1, err
			} else {
				engineUser = int64(user)
			}
		}
	}
	return engineUser, nil
}

func createExecutorContainer(mlDep *machinelearningv1.SeldonDeployment, p *machinelearningv1.PredictorSpec, predictorB64 string, http_port int, grpc_port int, resources *corev1.ResourceRequirements) corev1.Container {
	transport := mlDep.Spec.Transport
	//Backwards compatible with older resources
	if transport == "" {
		if p.Graph.Endpoint.Type == machinelearningv1.GRPC {
			transport = machinelearningv1.TransportGrpc
		} else {
			transport = machinelearningv1.TransportRest
		}
	}
	protocol := mlDep.Spec.Protocol
	//Backwards compatibility for older resources
	if protocol == "" {
		protocol = machinelearningv1.ProtocolSeldon
	}
	return corev1.Container{
		Name:  EngineContainerName,
		Image: GetEnv("EXECUTOR_CONTAINER_IMAGE_AND_VERSION", "seldonio/seldon-core-executor:1.0.1-SNAPSHOT"),
		Args: []string{
			"--sdep", mlDep.Name,
			"--namespace", mlDep.Namespace,
			"--predictor", p.Name,
			"--http_port", strconv.Itoa(http_port),
			"--grpc_port", strconv.Itoa(grpc_port),
			"--transport", string(transport),
			"--protocol", string(protocol),
			"--prometheus_path", getPrometheusPath(mlDep),
		},
		ImagePullPolicy:          corev1.PullPolicy(GetEnv("EXECUTOR_CONTAINER_IMAGE_PULL_POLICY", "IfNotPresent")),
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
			{Name: "REQUEST_LOGGER_DEFAULT_ENDPOINT_PREFIX", Value: GetEnv("EXECUTOR_REQUEST_LOGGER_DEFAULT_ENDPOINT_PREFIX", "")},
		},
		Ports: []corev1.ContainerPort{
			{ContainerPort: int32(http_port), Protocol: corev1.ProtocolTCP},
			{ContainerPort: int32(http_port), Protocol: corev1.ProtocolTCP, Name: constants.MetricsPortName},
			{ContainerPort: int32(grpc_port), Protocol: corev1.ProtocolTCP},
		},
		ReadinessProbe: &corev1.Probe{Handler: corev1.Handler{HTTPGet: &corev1.HTTPGetAction{Port: intstr.FromInt(http_port), Path: "/ready", Scheme: corev1.URISchemeHTTP}},
			InitialDelaySeconds: 20,
			PeriodSeconds:       5,
			FailureThreshold:    3,
			SuccessThreshold:    1,
			TimeoutSeconds:      60},
		LivenessProbe: &corev1.Probe{Handler: corev1.Handler{HTTPGet: &corev1.HTTPGetAction{Port: intstr.FromInt(http_port), Path: "/live", Scheme: corev1.URISchemeHTTP}},
			InitialDelaySeconds: 20,
			PeriodSeconds:       5,
			FailureThreshold:    3,
			SuccessThreshold:    1,
			TimeoutSeconds:      60},
		Resources: *resources,
	}
}

func createEngineContainerSpec(mlDep *machinelearningv1.SeldonDeployment, p *machinelearningv1.PredictorSpec, predictorB64 string,
	engine_http_port int, engine_grpc_port int, engineResources *corev1.ResourceRequirements) corev1.Container {
	return corev1.Container{
		Name:                     EngineContainerName,
		Image:                    GetEnv("ENGINE_CONTAINER_IMAGE_AND_VERSION", "seldonio/engine:0.4.0"),
		ImagePullPolicy:          corev1.PullPolicy(GetEnv("ENGINE_CONTAINER_IMAGE_PULL_POLICY", "IfNotPresent")),
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
			{Name: "DEPLOYMENT_NAME", Value: mlDep.Name},
			{Name: "DEPLOYMENT_NAMESPACE", Value: mlDep.ObjectMeta.Namespace},
			{Name: "ENGINE_SERVER_PORT", Value: strconv.Itoa(engine_http_port)},
			{Name: "ENGINE_SERVER_GRPC_PORT", Value: strconv.Itoa(engine_grpc_port)},
			{Name: "JAVA_OPTS", Value: getAnnotation(mlDep, machinelearningv1.ANNOTATION_JAVA_OPTS, "-server -Dcom.sun.management.jmxremote.rmi.port=9090 -Dcom.sun.management.jmxremote -Dcom.sun.management.jmxremote.port=9090 -Dcom.sun.management.jmxremote.ssl=false -Dcom.sun.management.jmxremote.authenticate=false -Dcom.sun.management.jmxremote.local.only=false -Djava.rmi.server.hostname=127.0.0.1")},
		},
		Ports: []corev1.ContainerPort{
			{ContainerPort: int32(engine_http_port), Protocol: corev1.ProtocolTCP},
			{ContainerPort: int32(engine_http_port), Protocol: corev1.ProtocolTCP, Name: constants.MetricsPortName},
			{ContainerPort: int32(engine_grpc_port), Protocol: corev1.ProtocolTCP},
			{ContainerPort: 8082, Name: "admin", Protocol: corev1.ProtocolTCP},
			{ContainerPort: 9090, Name: "jmx", Protocol: corev1.ProtocolTCP},
		},
		ReadinessProbe: &corev1.Probe{Handler: corev1.Handler{HTTPGet: &corev1.HTTPGetAction{Port: intstr.FromString("admin"), Path: "/ready", Scheme: corev1.URISchemeHTTP}},
			InitialDelaySeconds: 20,
			PeriodSeconds:       5,
			FailureThreshold:    3,
			SuccessThreshold:    1,
			TimeoutSeconds:      60},
		LivenessProbe: &corev1.Probe{Handler: corev1.Handler{HTTPGet: &corev1.HTTPGetAction{Port: intstr.FromString("admin"), Path: "/live", Scheme: corev1.URISchemeHTTP}},
			InitialDelaySeconds: 20,
			PeriodSeconds:       5,
			FailureThreshold:    3,
			SuccessThreshold:    1,
			TimeoutSeconds:      60},
		Lifecycle: &corev1.Lifecycle{
			PreStop: &corev1.Handler{
				Exec: &corev1.ExecAction{Command: []string{"/bin/sh", "-c", "curl 127.0.0.1:" + strconv.Itoa(engine_http_port) + "/pause; /bin/sleep 10"}},
			},
		},
		Resources: *engineResources,
	}
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
	predictorB64, err := getEngineVarJson(pCopy)
	if err != nil {
		return nil, err
	}

	//Engine resources
	engineResources := p.SvcOrchSpec.Resources
	if engineResources == nil {
		cpuQuantity, _ := resource.ParseQuantity("0.1")
		engineResources = &corev1.ResourceRequirements{
			Requests: map[corev1.ResourceName]resource.Quantity{
				corev1.ResourceCPU: cpuQuantity,
			},
		}
	}

	var c corev1.Container
	if isExecutorEnabled(mlDep) {
		executor_http_port, err := getExecutorHttpPort()
		if err != nil {
			return nil, err
		}
		executor_grpc_port, err := getExecutorGrpcPort()
		if err != nil {
			return nil, err
		}
		c = createExecutorContainer(mlDep, p, predictorB64, executor_http_port, executor_grpc_port, engineResources)
	} else {
		c = createEngineContainerSpec(mlDep, p, predictorB64, engine_http_port, engine_grpc_port, engineResources)
	}

	if engineUser != -1 {
		c.SecurityContext = &corev1.SecurityContext{RunAsUser: &engineUser}
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

	if _, ok := svcOrchEnvMap["SELDON_LOG_MESSAGES_EXTERNALLY"]; ok {
		//this env var is set already so no need to set a default
	} else {
		c.Env = append(c.Env, corev1.EnvVar{Name: "SELDON_LOG_MESSAGES_EXTERNALLY", Value: GetEnv("ENGINE_LOG_MESSAGES_EXTERNALLY", "false")})
	}
	return &c, nil
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
			Name:        depName,
			Namespace:   getNamespace(mlDep),
			Labels:      map[string]string{machinelearningv1.Label_svc_orch: "true", machinelearningv1.Label_seldon_app: seldonId, machinelearningv1.Label_seldon_id: seldonId, "app": depName, "version": "v1", "fluentd": "true"},
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
			Replicas: &p.Replicas,
		},
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
