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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sort"
	"strconv"
	"strings"

	"encoding/json"
	"github.com/go-logr/logr"
	machinelearningv1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	"github.com/seldonio/seldon-core/operator/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"os"
)

const (
	ExplainerConfigMapKeyName = "explainer"
	EnvExplainerImageRelated  = "RELATED_IMAGE_EXPLAINER"
)

var (
	envExplainerImage = os.Getenv(EnvExplainerImageRelated)
)

type ExplainerInitialiser struct {
	clientset kubernetes.Interface
}

func NewExplainerInitializer(clientset kubernetes.Interface) *ExplainerInitialiser {
	return &ExplainerInitialiser{clientset: clientset}
}

type ExplainerConfig struct {
	Image string `json:"image"`
}

func (ei *ExplainerInitialiser) getExplainerConfigs() (*ExplainerConfig, error) {
	configMap, err := ei.clientset.CoreV1().ConfigMaps(ControllerNamespace).Get(ControllerConfigMapName, metav1.GetOptions{})
	if err != nil {
		//log.Error(err, "Failed to find config map", "name", ControllerConfigMapName)
		return nil, err
	}
	return getExplainerConfigsFromMap(configMap)
}

func getExplainerConfigsFromMap(configMap *corev1.ConfigMap) (*ExplainerConfig, error) {
	explainerConfig := &ExplainerConfig{}
	if initializerConfig, ok := configMap.Data[ExplainerConfigMapKeyName]; ok {
		err := json.Unmarshal([]byte(initializerConfig), &explainerConfig)
		if err != nil {
			panic(fmt.Errorf("Unable to unmarshall %v json string due to %v ", ExplainerConfigMapKeyName, err))
		}
	}
	return explainerConfig, nil
}

func (ei *ExplainerInitialiser) createExplainer(mlDep *machinelearningv1.SeldonDeployment, p *machinelearningv1.PredictorSpec, c *components, pSvcName string, podSecurityContect *corev1.PodSecurityContext, ingresses []Ingress, log logr.Logger) error {

	if !isEmptyExplainer(p.Explainer) {

		seldonId := machinelearningv1.GetSeldonDeploymentName(mlDep)

		depName := machinelearningv1.GetExplainerDeploymentName(mlDep.GetName(), p)

		explainerContainer := p.Explainer.ContainerSpec

		if explainerContainer.Name == "" {
			explainerContainer.Name = depName
		}

		if explainerContainer.ImagePullPolicy == "" {
			explainerContainer.ImagePullPolicy = corev1.PullIfNotPresent
		}

		if p.Graph.Endpoint == nil {
			p.Graph.Endpoint = &machinelearningv1.Endpoint{Type: machinelearningv1.REST}
		}

		// Image from configMap or Relalated Image if its not set
		if explainerContainer.Image == "" {
			if envExplainerImage != "" {
				explainerContainer.Image = envExplainerImage
			} else {
				config, err := ei.getExplainerConfigs()
				if err != nil {
					return err
				}
				explainerContainer.Image = config.Image
			}
		}

		// explainer can get port from spec or from containerSpec or fall back on default
		var httpPort = 0
		var grpcPort = 0
		var portNum int32 = 9000
		var explainerProtocol string
		if p.Explainer.Endpoint != nil && p.Explainer.Endpoint.ServicePort != 0 {
			portNum = p.Explainer.Endpoint.ServicePort
		}
		var pSvcEndpoint = ""
		// Explainer only accepts http at present
		portType := "http"
		httpPort = int(portNum)
		customPort := getPort(portType, explainerContainer.Ports)

		if p.Explainer.Endpoint != nil && p.Explainer.Endpoint.Type == machinelearningv1.GRPC {
			explainerProtocol = "grpc"
			pSvcEndpoint = c.serviceDetails[pSvcName].GrpcEndpoint
		} else {
			explainerProtocol = "http"
			pSvcEndpoint = c.serviceDetails[pSvcName].HttpEndpoint
		}

		if customPort == nil {
			explainerContainer.Ports = append(explainerContainer.Ports, corev1.ContainerPort{Name: portType, ContainerPort: portNum, Protocol: corev1.ProtocolTCP})
		} else {
			portNum = customPort.ContainerPort
			portType = customPort.Name
		}

		if explainerContainer.LivenessProbe == nil {
			explainerContainer.LivenessProbe = &corev1.Probe{Handler: corev1.Handler{TCPSocket: &corev1.TCPSocketAction{Port: intstr.FromString(portType)}}, InitialDelaySeconds: 60, PeriodSeconds: 5, SuccessThreshold: 1, FailureThreshold: 5, TimeoutSeconds: 1}
		}
		if explainerContainer.ReadinessProbe == nil {
			explainerContainer.ReadinessProbe = &corev1.Probe{Handler: corev1.Handler{TCPSocket: &corev1.TCPSocketAction{Port: intstr.FromString(portType)}}, InitialDelaySeconds: 20, PeriodSeconds: 5, SuccessThreshold: 1, FailureThreshold: 7, TimeoutSeconds: 1}
		}

		// Add livecycle probe
		if explainerContainer.Lifecycle == nil {
			explainerContainer.Lifecycle = &corev1.Lifecycle{PreStop: &corev1.Handler{Exec: &corev1.ExecAction{Command: []string{"/bin/sh", "-c", "/bin/sleep 10"}}}}
		}

		explainerContainer.Args = []string{
			"--model_name=" + mlDep.Name,
			"--predictor_host=" + pSvcEndpoint,
			"--protocol=" + "seldon." + explainerProtocol,
			"--http_port=" + strconv.Itoa(int(portNum)),
		}

		if p.Explainer.ModelUri != "" {
			explainerContainer.Args = append(explainerContainer.Args, "--storage_uri="+DefaultModelLocalMountPath)
		}

		explainerContainer.Args = append(explainerContainer.Args, string(p.Explainer.Type))

		if p.Explainer.Type == machinelearningv1.AlibiAnchorsImageExplainer {
			explainerContainer.Args = append(explainerContainer.Args, "--tf_data_type=float32")
		}

		// Order explainer config map keys
		var keys []string
		for k, _ := range p.Explainer.Config {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			v := p.Explainer.Config[k]
			//remote files in model location should get downloaded by initializer
			if p.Explainer.ModelUri != "" {
				v = strings.Replace(v, p.Explainer.ModelUri, "/mnt/models", 1)
			}
			arg := "--" + k + "=" + v
			explainerContainer.Args = append(explainerContainer.Args, arg)
		}
		// see https://github.com/cliveseldon/kfserving/tree/explainer_update_jul/docs/samples/explanation/income for more

		// Add Environment Variables - TODO: are these needed
		if !utils.HasEnvVar(explainerContainer.Env, machinelearningv1.ENV_PREDICTIVE_UNIT_SERVICE_PORT) {
			explainerContainer.Env = append(explainerContainer.Env, []corev1.EnvVar{
				corev1.EnvVar{Name: machinelearningv1.ENV_PREDICTIVE_UNIT_SERVICE_PORT, Value: strconv.Itoa(int(portNum))},
				corev1.EnvVar{Name: machinelearningv1.ENV_PREDICTIVE_UNIT_ID, Value: explainerContainer.Name},
				corev1.EnvVar{Name: machinelearningv1.ENV_PREDICTOR_ID, Value: p.Name},
				corev1.EnvVar{Name: machinelearningv1.ENV_SELDON_DEPLOYMENT_ID, Value: mlDep.ObjectMeta.Name},
			}...)
		}

		seldonPodSpec := machinelearningv1.SeldonPodSpec{Spec: corev1.PodSpec{
			Containers: []corev1.Container{explainerContainer},
		}}

		deploy := createDeploymentWithoutEngine(depName, seldonId, &seldonPodSpec, p, mlDep, podSecurityContect)

		if p.Explainer.ModelUri != "" {
			var err error

			mi := NewModelInitializer(ei.clientset)
			deploy, err = mi.InjectModelInitializer(deploy, explainerContainer.Name, p.Explainer.ModelUri, p.Explainer.ServiceAccountName, p.Explainer.EnvSecretRefName)
			if err != nil {
				return err
			}
		}

		// for explainer use same service name as its Deployment
		eSvcName := machinelearningv1.GetExplainerDeploymentName(mlDep.GetName(), p)

		deploy.ObjectMeta.Labels[machinelearningv1.Label_seldon_app] = eSvcName
		deploy.Spec.Template.ObjectMeta.Labels[machinelearningv1.Label_seldon_app] = eSvcName

		c.deployments = append(c.deployments, deploy)

		// Use seldondeployment name dash explainer as the external service name. This should allow canarying.
		eSvc, err := createService(eSvcName, seldonId, p, mlDep, httpPort, grpcPort, true, ingresses, log)
		if err != nil {
			return err
		}
		c.services = append(c.services, eSvc)
		c.serviceDetails[eSvcName] = &machinelearningv1.ServiceStatus{
			SvcName:      eSvcName,
			HttpEndpoint: eSvcName + "." + eSvc.Namespace + ":" + strconv.Itoa(httpPort),
			ExplainerFor: machinelearningv1.GetPredictorKey(mlDep, p),
		}
		if grpcPort > 0 {
			c.serviceDetails[eSvcName].GrpcEndpoint = eSvcName + "." + eSvc.Namespace + ":" + strconv.Itoa(grpcPort)
		}
		for _, ingress := range ingresses {
			ingressResources, err := ingress.GenerateExplainerResources(eSvcName, p, mlDep, seldonId, getNamespace(mlDep), httpPort, grpcPort)
			if err != nil {
				return err
			}
			c.ingressResources = mergeIngressResourceMap(c.ingressResources, ingressResources)
		}
	}

	return nil
}
