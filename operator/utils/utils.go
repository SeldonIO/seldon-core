package utils

import (
	"encoding/json"
	"os"
	"strconv"
	"strings"

	machinelearningv1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
)

// Attempts to add environment variable but does NOT override if already existing
func AddEnvVarToDeploymentContainers(deploy *appsv1.Deployment, envVar *v1.EnvVar) {
	for containerIdx := 0; containerIdx < len(deploy.Spec.Template.Spec.Containers); containerIdx++ {
		deployContainer := &deploy.Spec.Template.Spec.Containers[containerIdx]
		// Before adding it ensures it doesn't exist as it will NOT override
		for _, e := range deployContainer.Env {
			if e.Name == envVar.Name {
				return
			}
		}
		deployContainer.Env = append(deployContainer.Env, *envVar)
	}
}

// Create a volume from the secret provided and mounts it to all the containers of deployment
func MountSecretToDeploymentContainers(deploy *appsv1.Deployment, secretRefName string, containerMountPath string) {
	volumeName := "seldon-cert-volume"
	volume := v1.Volume{
		Name: volumeName,
		VolumeSource: v1.VolumeSource{
			Secret: &v1.SecretVolumeSource{
				SecretName: secretRefName,
			},
		},
	}
	deploy.Spec.Template.Spec.Volumes = append(deploy.Spec.Template.Spec.Volumes, volume)
	// Mount the volume secret to all the containers in deployment
	for containerIdx := 0; containerIdx < len(deploy.Spec.Template.Spec.Containers); containerIdx++ {
		deployContainer := &deploy.Spec.Template.Spec.Containers[containerIdx]
		volumeMount := v1.VolumeMount{
			Name:      volumeName,
			MountPath: containerMountPath,
			ReadOnly:  true,
		}
		deployContainer.VolumeMounts = append(deployContainer.VolumeMounts, volumeMount)
	}
}

func GetPredictionPath(mlDep *machinelearningv1.SeldonDeployment) string {
	protocol := mlDep.Spec.Protocol

	if protocol == machinelearningv1.ProtocolTensorflow {
		// This will be updated as part of https://github.com/SeldonIO/seldon-core/issues/1611
		return "/v1/models/" + mlDep.Spec.Predictors[0].Graph.Name + ":predict"
	} else if protocol == machinelearningv1.ProtocolKfserving || protocol == machinelearningv1.ProtocolV2 {
		return "/v2/models/" + mlDep.Spec.Predictors[0].Graph.Name + "/infer"
	} else {
		return "/api/v1.0/predictions"
	}
}

func GetPredictiveUnitAsJson(params []machinelearningv1.Parameter) string {
	str, err := json.Marshal(params)
	if err != nil {
		return ""
	} else {
		return string(str)
	}
}

func GetSeldonPodSpecForPredictiveUnit(p *machinelearningv1.PredictorSpec, name string) (*machinelearningv1.SeldonPodSpec, int) {
	for j := 0; j < len(p.ComponentSpecs); j++ {
		cSpec := p.ComponentSpecs[j]
		for k := 0; k < len(cSpec.Spec.Containers); k++ {
			c := &cSpec.Spec.Containers[k]
			//the podSpec will have a container matching the PU name
			if c.Name == name {
				return cSpec, j
			}
		}
	}
	return nil, 0
}

func GetContainerForDeployment(deploy *appsv1.Deployment, name string) *v1.Container {
	var userContainer *v1.Container
	for idx, container := range deploy.Spec.Template.Spec.Containers {
		if strings.Compare(container.Name, name) == 0 {
			userContainer = &deploy.Spec.Template.Spec.Containers[idx]
			return userContainer
		}
	}
	return nil
}

func HasEnvVar(envVars []v1.EnvVar, name string) bool {
	for _, envVar := range envVars {
		if envVar.Name == name {
			return true
		}
	}
	return false
}

func SetEnvVar(envVars []v1.EnvVar, newVar v1.EnvVar, override bool) (newEnvVars []v1.EnvVar) {
	found := false
	index := 0
	for i, envVar := range envVars {
		if envVar.Name == newVar.Name {
			found = true
			index = i
		}
	}

	if found && override {
		newEnvVars = append(envVars[:index])
		newEnvVars = append(newEnvVars, newVar)
		newEnvVars = append(newEnvVars, envVars[index+1:]...)
	}

	if !found {
		newEnvVars = envVars
		newEnvVars = append(newEnvVars, newVar)
	}
	return newEnvVars
}

// Get an environment variable given by key or return the fallback.
func GetEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// Get an environment variable given by key or return the fallback.
func GetEnvAsBool(key string, fallback bool) bool {
	if raw, ok := os.LookupEnv(key); ok {
		val, err := strconv.ParseBool(raw)
		if err == nil {
			return val
		}
	}

	return fallback
}

func IsEmptyTLS(p *machinelearningv1.PredictorSpec) bool {
	return p.SSL == nil || len(p.SSL.CertSecretName) == 0
}
