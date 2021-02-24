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
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	machinelearningv1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	"github.com/seldonio/seldon-core/operator/constants"
	"github.com/seldonio/seldon-core/operator/utils"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
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
	ctx       context.Context
}

func NewPrePackedInitializer(ctx context.Context, clientset kubernetes.Interface) *PrePackedInitialiser {
	return &PrePackedInitialiser{clientset: clientset, ctx: ctx}
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

func createTensorflowServingContainer(mlDepSepc *machinelearningv1.SeldonDeploymentSpec, pu *machinelearningv1.PredictiveUnit, tensorflowProtocol bool) *v1.Container {
	ServerConfig := machinelearningv1.GetPrepackServerConfig(string(*pu.Implementation))

	tfImage := ServerConfig.PrepackImageName(machinelearningv1.ProtocolTensorflow, pu)

	grpcPort := int32(constants.TfServingGrpcPort)
	restPort := int32(constants.TfServingRestPort)
	name := constants.TFServingContainerName
	if tensorflowProtocol {
		grpcPort = pu.Endpoint.GrpcPort
		restPort = pu.Endpoint.HttpPort
		name = pu.Name
	}

	return &v1.Container{
		Name:  name,
		Image: tfImage,
		Args: []string{
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

func (pi *PrePackedInitialiser) addTFServerContainer(mlDepSpec *machinelearningv1.SeldonDeploymentSpec, pu *machinelearningv1.PredictiveUnit, deploy *appsv1.Deployment, serverConfig *machinelearningv1.PredictorServerConfig) error {
	ty := machinelearningv1.MODEL
	pu.Type = &ty

	c := utils.GetContainerForDeployment(deploy, pu.Name)

	var tfServingContainer *v1.Container
	if mlDepSpec.Protocol == machinelearningv1.ProtocolTensorflow {
		tfServingContainer = c
	} else {
		c.Image = serverConfig.PrepackImageName(mlDepSpec.Protocol, pu)
		SetUriParamsForTFServingProxyContainer(pu, c)
		tfServingContainer = utils.GetContainerForDeployment(deploy, constants.TFServingContainerName)
	}

	existing := tfServingContainer != nil
	if !existing {
		tfServingContainer = createTensorflowServingContainer(mlDepSpec, pu, mlDepSpec.Protocol == machinelearningv1.ProtocolTensorflow)
		deploy.Spec.Template.Spec.Containers = append(deploy.Spec.Template.Spec.Containers, *tfServingContainer)
	} else {
		// Update any missing fields
		protoType := createTensorflowServingContainer(mlDepSpec, pu, mlDepSpec.Protocol == machinelearningv1.ProtocolTensorflow)
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

	mi := NewModelInitializer(pi.ctx, pi.clientset)
	_, err := mi.InjectModelInitializer(deploy, tfServingContainer.Name, pu.ModelURI, pu.ServiceAccountName, envSecretRefName, pu.StorageInitializerImage)
	if err != nil {
		return err
	}
	return nil
}

func (pi *PrePackedInitialiser) addTritonServer(mlDepSpec *machinelearningv1.SeldonDeploymentSpec, pu *machinelearningv1.PredictiveUnit, deploy *appsv1.Deployment, serverConfig *machinelearningv1.PredictorServerConfig) error {

	c := utils.GetContainerForDeployment(deploy, pu.Name)
	existing := c != nil

	tritonUser := int64(1000)

	cServer := &v1.Container{
		Name: pu.Name,
		Args: []string{
			"/opt/tritonserver/bin/tritonserver",
			"--grpc-port=" + strconv.Itoa(int(pu.Endpoint.GrpcPort)),
			"--http-port=" + strconv.Itoa(int(pu.Endpoint.HttpPort)),
			"--model-repository=" + DefaultModelLocalMountPath,
			"--strict-model-config=false",
		},
		Ports: []v1.ContainerPort{
			{
				Name:          "grpc",
				ContainerPort: pu.Endpoint.GrpcPort,
				Protocol:      v1.ProtocolTCP,
			},
			{
				Name:          "http",
				ContainerPort: pu.Endpoint.HttpPort,
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
		SecurityContext: &v1.SecurityContext{
			RunAsUser: &tritonUser,
		},
		VolumeMounts: []v1.VolumeMount{
			{
				Name:      machinelearningv1.PODINFO_VOLUME_NAME,
				MountPath: machinelearningv1.PODINFO_VOLUME_PATH,
			},
		},
	}
	cServer.Image = serverConfig.PrepackImageName(mlDepSpec.Protocol, pu)

	if existing {
		// Overwrite core items if not existing or required
		if c.Image == "" {
			c.Image = cServer.Image
		}
		if c.Args == nil {
			c.Args = cServer.Args
		}
		if c.ReadinessProbe == nil {
			c.ReadinessProbe = cServer.ReadinessProbe
		}
		if c.LivenessProbe == nil {
			c.LivenessProbe = cServer.LivenessProbe
		}
		if c.SecurityContext == nil {
			c.SecurityContext = cServer.SecurityContext
		}
		// Ports always overwritten
		// Need to look as we seem to add metrics ports automatically which mean this needs to be done
		c.Ports = cServer.Ports
	} else {
		if len(deploy.Spec.Template.Spec.Containers) > 0 {
			deploy.Spec.Template.Spec.Containers = append(deploy.Spec.Template.Spec.Containers, *cServer)
		} else {
			deploy.Spec.Template.Spec.Containers = []v1.Container{*cServer}
		}
	}

	envSecretRefName := extractEnvSecretRefName(pu)
	mi := NewModelInitializer(pi.ctx, pi.clientset)
	_, err := mi.InjectModelInitializer(deploy, c.Name, pu.ModelURI, pu.ServiceAccountName, envSecretRefName, pu.StorageInitializerImage)
	if err != nil {
		return err
	}
	return nil
}

func (pi *PrePackedInitialiser) addMLServerDefault(pu *machinelearningv1.PredictiveUnit, deploy *appsv1.Deployment) error {
	c, err := getMLServerContainer(pu)
	if err != nil {
		return err
	}

	existingContainer := utils.GetContainerForDeployment(deploy, pu.Name)
	if existingContainer != nil {
		c = mergeMLServerContainer(existingContainer, c)
	} else {
		templateSpec := deploy.Spec.Template.Spec
		if len(templateSpec.Containers) == 0 {
			templateSpec.Containers = []v1.Container{}
		}

		templateSpec.Containers = append(templateSpec.Containers, *c)
	}

	envSecretRefName := extractEnvSecretRefName(pu)
	mi := NewModelInitializer(pi.ctx, pi.clientset)

	_, err = mi.InjectModelInitializer(deploy, c.Name, pu.ModelURI, pu.ServiceAccountName, envSecretRefName, pu.StorageInitializerImage)
	if err != nil {
		return err
	}

	return nil
}

func (pi *PrePackedInitialiser) addModelDefaultServers(mlDepSepc *machinelearningv1.SeldonDeploymentSpec, pu *machinelearningv1.PredictiveUnit, deploy *appsv1.Deployment, serverConfig *machinelearningv1.PredictorServerConfig) error {
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

	if c.Image == "" {
		c.Image = serverConfig.PrepackImageName(mlDepSepc.Protocol, pu)
	}

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

	mi := NewModelInitializer(pi.ctx, pi.clientset)
	_, err = mi.InjectModelInitializer(deploy, c.Name, pu.ModelURI, pu.ServiceAccountName, envSecretRefName, pu.StorageInitializerImage)
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
		uriParam := machinelearningv1.Parameter{
			Name:  "rest_endpoint",
			Type:  "STRING",
			Value: "http://0.0.0.0:2001",
		}
		parameters = append(parameters, uriParam)
		uriParam = machinelearningv1.Parameter{
			Name:  "grpc_endpoint",
			Type:  "STRING",
			Value: "0.0.0.0:2000",
		}
		parameters = append(parameters, uriParam)
	}

	modelNameParam := machinelearningv1.Parameter{
		Name:  "model_name",
		Type:  "STRING",
		Value: pu.Name,
	}

	parameters = append(parameters, modelNameParam)

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
			switch *pu.Implementation {
			case machinelearningv1.PrepackTensorflowName:
				if err := pi.addTFServerContainer(&mlDep.Spec, pu, deploy, serverConfig); err != nil {
					return err
				}
			case machinelearningv1.PrepackTritonName:
				if err := pi.addTritonServer(&mlDep.Spec, pu, deploy, serverConfig); err != nil {
					return err
				}
			default:
				// If protocol is KFServing, try to add container with MLServer
				if mlDep.Spec.Protocol == machinelearningv1.ProtocolKfserving {
					err := pi.addMLServerDefault(pu, deploy)
					if err != nil {
						return err
					}
				} else {
					if err := pi.addModelDefaultServers(&mlDep.Spec, pu, deploy, serverConfig); err != nil {
						return err
					}
				}
			}
		} else {
			return fmt.Errorf("Failed to get server config for %s", *pu.Implementation)
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
