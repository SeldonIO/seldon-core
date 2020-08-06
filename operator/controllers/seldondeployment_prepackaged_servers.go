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
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	machinelearningv1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	"github.com/seldonio/seldon-core/operator/constants"
	"github.com/seldonio/seldon-core/operator/utils"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	ENV_PREDICTIVE_UNIT_DEFAULT_ENV_SECRET_REF_NAME = "PREDICTIVE_UNIT_DEFAULT_ENV_SECRET_REF_NAME"
)

var (
	PredictiveUnitDefaultEnvSecretRefName = utils.GetEnv(ENV_PREDICTIVE_UNIT_DEFAULT_ENV_SECRET_REF_NAME, "")
)

type PrePackedInitialiser struct {
	clientset kubernetes.Interface
}

func NewPrePackedInitializer(clientset kubernetes.Interface) *PrePackedInitialiser {
	return &PrePackedInitialiser{clientset: clientset}
}

func extractEnvSecretRefName(pu *machinelearningv1.PredictiveUnit) string {
	envSecretRefName := ""
	if pu.EnvSecretRefName == "" {
		envSecretRefName = PredictiveUnitDefaultEnvSecretRefName
	} else {
		envSecretRefName = pu.EnvSecretRefName
	}
	return envSecretRefName
}

func createTensorflowServingContainer(pu *machinelearningv1.PredictiveUnit, tensorflowProtocol bool) *v1.Container {
	ServerConfig := machinelearningv1.GetPrepackServerConfig(string(*pu.Implementation))

	tfImage := "tensorflow/serving:latest"
	if ServerConfig.TensorflowImage != "" {
		tfImage = ServerConfig.TensorflowImage
	}

	grpcPort := int32(constants.TfServingGrpcPort)
	restPort := int32(constants.TfServingRestPort)
	name := constants.TFServingContainerName
	if tensorflowProtocol {
		if pu.Endpoint.Type == machinelearningv1.GRPC {
			grpcPort = pu.Endpoint.ServicePort
		} else {
			restPort = pu.Endpoint.ServicePort
		}
		name = pu.Name
	}

	return &v1.Container{
		Name:  name,
		Image: tfImage,
		Args: []string{
			"/usr/bin/tensorflow_model_server",
			constants.TfServingArgPort + strconv.Itoa(int(grpcPort)),
			constants.TfServingArgRestPort + strconv.Itoa(int(restPort)),
			"--model_name=" + pu.Name,
			"--model_base_path=" + DefaultModelLocalMountPath},
		ImagePullPolicy: v1.PullIfNotPresent,
		Ports: []v1.ContainerPort{
			{
				ContainerPort: grpcPort,
				Protocol:      v1.ProtocolTCP,
			},
			{
				ContainerPort: restPort,
				Protocol:      v1.ProtocolTCP,
			},
		},
	}
}

func (pi *PrePackedInitialiser) addTFServerContainer(mlDep *machinelearningv1.SeldonDeployment, pu *machinelearningv1.PredictiveUnit, deploy *appsv1.Deployment, serverConfig *machinelearningv1.PredictorServerConfig) error {
	ty := machinelearningv1.MODEL
	pu.Type = &ty

	c := utils.GetContainerForDeployment(deploy, pu.Name)

	var tfServingContainer *v1.Container
	if mlDep.Spec.Protocol == machinelearningv1.ProtocolTensorflow {
		tfServingContainer = c
	} else {
		machinelearningv1.SetImageNameForPrepackContainer(pu, c, serverConfig)
		SetUriParamsForTFServingProxyContainer(pu, c)
		tfServingContainer = utils.GetContainerForDeployment(deploy, constants.TFServingContainerName)
	}

	existing := tfServingContainer != nil
	if !existing {
		tfServingContainer = createTensorflowServingContainer(pu, mlDep.Spec.Protocol == machinelearningv1.ProtocolTensorflow)
		deploy.Spec.Template.Spec.Containers = append(deploy.Spec.Template.Spec.Containers, *tfServingContainer)
	} else {
		// Update any missing fields
		protoType := createTensorflowServingContainer(pu, mlDep.Spec.Protocol == machinelearningv1.ProtocolTensorflow)
		if tfServingContainer.Image == "" {
			tfServingContainer.Image = protoType.Image
		}
		if tfServingContainer.Args == nil || len(tfServingContainer.Args) == 0 {
			tfServingContainer.Args = protoType.Args
		}
		if tfServingContainer.Ports == nil || len(tfServingContainer.Ports) == 0 {
			tfServingContainer.Ports = protoType.Ports
		}
	}

	envSecretRefName := extractEnvSecretRefName(pu)

	mi := NewModelInitializer(pi.clientset)
	_, err := mi.InjectModelInitializer(deploy, tfServingContainer.Name, pu.ModelURI, pu.ServiceAccountName, envSecretRefName)
	if err != nil {
		return err
	}
	return nil
}

func (pi *PrePackedInitialiser) addModelDefaultServers(pu *machinelearningv1.PredictiveUnit, deploy *appsv1.Deployment, serverConfig *machinelearningv1.PredictorServerConfig) error {
	ty := machinelearningv1.MODEL
	pu.Type = &ty

	if pu.Endpoint == nil {
		pu.Endpoint = &machinelearningv1.Endpoint{Type: machinelearningv1.REST}
	}
	c := utils.GetContainerForDeployment(deploy, pu.Name)
	existing := c != nil
	if !existing {
		c = &v1.Container{
			Name: pu.Name,
			VolumeMounts: []v1.VolumeMount{
				{
					Name:      machinelearningv1.PODINFO_VOLUME_NAME,
					MountPath: machinelearningv1.PODINFO_VOLUME_PATH,
				},
			},
		}
	}

	machinelearningv1.SetImageNameForPrepackContainer(pu, c, serverConfig)

	// Add parameters envvar - point at mount path because initContainer will download
	params := pu.Parameters
	uriParam := machinelearningv1.Parameter{
		Name:  "model_uri",
		Type:  "STRING",
		Value: DefaultModelLocalMountPath,
	}
	params = append(params, uriParam)
	paramStr, err := json.Marshal(params)
	if err != nil {
		return err
	}

	if len(params) > 0 {
		if !utils.HasEnvVar(c.Env, machinelearningv1.ENV_PREDICTIVE_UNIT_PARAMETERS) {
			c.Env = append(c.Env, v1.EnvVar{Name: machinelearningv1.ENV_PREDICTIVE_UNIT_PARAMETERS, Value: string(paramStr)})
		} else {
			c.Env = utils.SetEnvVar(c.Env, v1.EnvVar{Name: machinelearningv1.ENV_PREDICTIVE_UNIT_PARAMETERS, Value: string(paramStr)})
		}

	}

	// Add container to deployment
	if !existing {
		if len(deploy.Spec.Template.Spec.Containers) > 0 {
			deploy.Spec.Template.Spec.Containers = append(deploy.Spec.Template.Spec.Containers, *c)
		} else {
			deploy.Spec.Template.Spec.Containers = []v1.Container{*c}
		}
	}

	envSecretRefName := extractEnvSecretRefName(pu)

	mi := NewModelInitializer(pi.clientset)
	_, err = mi.InjectModelInitializer(deploy, c.Name, pu.ModelURI, pu.ServiceAccountName, envSecretRefName)
	if err != nil {
		return err
	}
	return nil
}

func SetUriParamsForTFServingProxyContainer(pu *machinelearningv1.PredictiveUnit, c *v1.Container) {

	parameters := pu.Parameters

	hasUriParams := false
	if len(pu.Parameters) > 0 {
		for _, paramElement := range pu.Parameters {
			if paramElement.Name == "rest_endpoint" || paramElement.Name == "grpc_endpoint" {

				hasUriParams = true
			}
		}
	}
	if !hasUriParams {
		var uriParam machinelearningv1.Parameter

		if pu.Endpoint.Type == machinelearningv1.REST {
			uriParam = machinelearningv1.Parameter{
				Name:  "rest_endpoint",
				Type:  "STRING",
				Value: "http://0.0.0.0:2001",
			}
		} else {
			uriParam = machinelearningv1.Parameter{
				Name:  "grpc_endpoint",
				Type:  "STRING",
				Value: "0.0.0.0:2000",
			}

		}

		parameters = append(pu.Parameters, uriParam)

		modelNameParam := machinelearningv1.Parameter{
			Name:  "model_name",
			Type:  "STRING",
			Value: pu.Name,
		}

		parameters = append(parameters, modelNameParam)

	}

	if len(parameters) > 0 {
		if !utils.HasEnvVar(c.Env, machinelearningv1.ENV_PREDICTIVE_UNIT_PARAMETERS) {
			c.Env = append(c.Env, v1.EnvVar{Name: machinelearningv1.ENV_PREDICTIVE_UNIT_PARAMETERS, Value: utils.GetPredictiveUnitAsJson(parameters)})
		} else {
			c.Env = utils.SetEnvVar(c.Env, v1.EnvVar{Name: machinelearningv1.ENV_PREDICTIVE_UNIT_PARAMETERS, Value: utils.GetPredictiveUnitAsJson(parameters)})
		}

	}
}

func (pi *PrePackedInitialiser) createStandaloneModelServers(mlDep *machinelearningv1.SeldonDeployment, p *machinelearningv1.PredictorSpec, c *components, pu *machinelearningv1.PredictiveUnit, podSecurityContext *v1.PodSecurityContext) error {

	if machinelearningv1.IsPrepack(pu) {
		sPodSpec, idx := utils.GetSeldonPodSpecForPredictiveUnit(p, pu.Name)
		if sPodSpec == nil {
			return fmt.Errorf("Failed to find PodSpec for Prepackaged server PreditiveUnit named %s", pu.Name)
		}
		depName := machinelearningv1.GetDeploymentName(mlDep, *p, sPodSpec, idx)
		seldonId := machinelearningv1.GetSeldonDeploymentName(mlDep)

		var deploy *appsv1.Deployment
		existing := false
		for i := 0; i < len(c.deployments); i++ {
			d := c.deployments[i]
			if strings.Compare(d.Name, depName) == 0 {
				deploy = d
				existing = true
				break
			}
		}

		// might not be a Deployment yet - if so we have to create one
		if deploy == nil {
			seldonId := machinelearningv1.GetSeldonDeploymentName(mlDep)
			deploy = createDeploymentWithoutEngine(depName, seldonId, sPodSpec, p, mlDep, podSecurityContext)
		}

		// apply serviceAccountName to pod to enable EKS fine-grained IAM roles
		if pu.ServiceAccountName != "" {
			deploy.Spec.Template.Spec.ServiceAccountName = pu.ServiceAccountName
		}

		serverConfig := machinelearningv1.GetPrepackServerConfig(string(*pu.Implementation))
		if serverConfig != nil {
			if *pu.Implementation != machinelearningv1.PrepackTensorflowName {
				if err := pi.addModelDefaultServers(pu, deploy, serverConfig); err != nil {
					return err
				}
			} else {
				if err := pi.addTFServerContainer(mlDep, pu, deploy, serverConfig); err != nil {
					return err
				}
			}
		}

		if !existing {

			// this is a new deployment so its containers won't have a containerService
			for k := 0; k < len(deploy.Spec.Template.Spec.Containers); k++ {
				con := &deploy.Spec.Template.Spec.Containers[k]

				//checking for con.Name != "" is a fallback check that we haven't got an empty/nil container as name is required
				if con.Name != EngineContainerName && con.Name != constants.TFServingContainerName && con.Name != "" {
					svc := createContainerService(deploy, *p, mlDep, con, *c, seldonId)
					c.services = append(c.services, svc)
				}
			}
			if len(deploy.Spec.Template.Spec.Containers) > 0 && deploy.Spec.Template.Spec.Containers[0].Name != "" {
				// Add deployment, provided we have a non-empty spec
				c.deployments = append(c.deployments, deploy)
			}
		}
	}

	for i := 0; i < len(pu.Children); i++ {
		if err := pi.createStandaloneModelServers(mlDep, p, c, &pu.Children[i], podSecurityContext); err != nil {
			return err
		}
	}
	return nil
}
