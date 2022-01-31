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
	"fmt"
	"sort"
	"strconv"
	"strings"

	"k8s.io/client-go/kubernetes"

	"encoding/json"
	"os"

	"github.com/go-logr/logr"
	machinelearningv1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	"github.com/seldonio/seldon-core/operator/constants"
	"github.com/seldonio/seldon-core/operator/utils"
	istio_networking "istio.io/api/networking/v1alpha3"
	istio "istio.io/client-go/pkg/apis/networking/v1alpha3"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
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
	ctx       context.Context
}

func NewExplainerInitializer(ctx context.Context, clientset kubernetes.Interface) *ExplainerInitialiser {
	return &ExplainerInitialiser{clientset: clientset, ctx: ctx}
}

type ExplainerConfig struct {
	Image    string `json:"image"`
	Image_v2 string `json:"image_v2"`
}

func (ei *ExplainerInitialiser) getExplainerConfigs() (*ExplainerConfig, error) {
	configMap, err := ei.clientset.CoreV1().ConfigMaps(ControllerNamespace).Get(ei.ctx, ControllerConfigMapName, metav1.GetOptions{})
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

func (ei *ExplainerInitialiser) createExplainer(mlDep *machinelearningv1.SeldonDeployment, p *machinelearningv1.PredictorSpec, c *components, pSvcName string, podSecurityContect *corev1.PodSecurityContext, log logr.Logger) error {

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

		explainerProtocol := string(machinelearningv1.ProtocolSeldon)
		if mlDep.Spec.Protocol == machinelearningv1.ProtocolTensorflow {
			explainerProtocol = string(machinelearningv1.ProtocolTensorflow)
		}
		if mlDep.Spec.Protocol == machinelearningv1.ProtocolKfserving || mlDep.Spec.Protocol == machinelearningv1.ProtocolV2 {
			explainerProtocol = string(machinelearningv1.ProtocolV2)
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
				if explainerProtocol == string(machinelearningv1.ProtocolV2) {
					explainerContainer.Image = config.Image_v2
				} else {
					explainerContainer.Image = config.Image
				}
			}
		}

		// explainer can get port from spec or from containerSpec or fall back on default
		var httpPort = 0
		var grpcPort = 0
		var portNum int32 = 9000
		var explainerTransport string
		if p.Explainer.Endpoint != nil && p.Explainer.Endpoint.ServicePort != 0 {
			portNum = p.Explainer.Endpoint.ServicePort
		}
		var pSvcEndpoint = ""
		//Explainer only accepts http at present
		portType := "http"
		httpPort = int(portNum)
		customPort := getPort(portType, explainerContainer.Ports)

		if mlDep.Spec.Transport == machinelearningv1.TransportGrpc || (p.Explainer.Endpoint != nil && p.Explainer.Endpoint.Type == machinelearningv1.GRPC) {
			explainerTransport = "grpc"
			pSvcEndpoint = c.serviceDetails[pSvcName].GrpcEndpoint
		} else {
			explainerTransport = "http"
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

		//TODO need to change python explainers to accept v2 as protocol name
		if explainerProtocol == string(machinelearningv1.ProtocolV2) {
			// add mlserver alibi runtime env vars
			// alibi-specific json
			explainEnvs, err := getAlibiExplainEnvVars(int(portNum), explainerContainer.Name, p.Explainer.Type, pSvcEndpoint, p.Graph.Name, p.Explainer.InitParameters)
			if err != nil {
				return err
			}

			explainerContainer.Env = explainEnvs
		} else {
			explainerContainer.Args = []string{
				"--model_name=" + p.Graph.Name,
				"--predictor_host=" + pSvcEndpoint,
				"--protocol=" + explainerProtocol + "." + explainerTransport,
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
		}

		seldonPodSpec := machinelearningv1.SeldonPodSpec{Spec: corev1.PodSpec{
			Containers: []corev1.Container{explainerContainer},
		}}

		deploy := createDeploymentWithoutEngine(depName, seldonId, &seldonPodSpec, p, mlDep, podSecurityContect, false)

		// Set replicas to zero if main predictor or graph has zero replicas otherwise set to explainer replicas
		if p.Replicas != nil && *p.Replicas == 0 {
			deploy.Spec.Replicas = p.Replicas
		} else if p.Replicas == nil && mlDep.Spec.Replicas != nil && *mlDep.Spec.Replicas == 0 {
			deploy.Spec.Replicas = mlDep.Spec.Replicas
		} else {
			deploy.Spec.Replicas = p.Explainer.Replicas
		}

		if p.Explainer.ModelUri != "" {
			var err error

			mi := NewModelInitializer(ei.ctx, ei.clientset)
			deploy, err = mi.InjectModelInitializer(deploy, explainerContainer.Name, p.Explainer.ModelUri, p.Explainer.ServiceAccountName, p.Explainer.EnvSecretRefName, p.Explainer.StorageInitializerImage)
			if err != nil {
				return err
			}
		}

		if p.Explainer.ServiceAccountName != "" {
			deploy.Spec.Template.Spec.ServiceAccountName = p.Explainer.ServiceAccountName
		}

		// for explainer use same service name as its Deployment
		eSvcName := machinelearningv1.GetExplainerDeploymentName(mlDep.GetName(), p)

		deploy = addLabelsToDeployment(deploy, nil, p)
		deploy.ObjectMeta.Labels[machinelearningv1.Label_seldon_app] = eSvcName
		deploy.Spec.Template.ObjectMeta.Labels[machinelearningv1.Label_seldon_app] = eSvcName

		c.deployments = append(c.deployments, deploy)

		// Use seldondeployment name dash explainer as the external service name. This should allow canarying.
		eSvc, err := createPredictorService(eSvcName, seldonId, p, mlDep, httpPort, grpcPort, true, log)
		if err != nil {
			return err
		}

		// Overwrite main Seldon App label onto SVC
		pSvcName := machinelearningv1.GetPredictorKey(mlDep, p)
		eSvc.Labels[machinelearningv1.Label_seldon_app] = pSvcName

		eSvc = addLabelsToService(eSvc, nil, p)
		c.services = append(c.services, eSvc)
		c.serviceDetails[eSvcName] = &machinelearningv1.ServiceStatus{
			SvcName:      eSvcName,
			HttpEndpoint: eSvcName + "." + eSvc.Namespace + ":" + strconv.Itoa(httpPort),
			ExplainerFor: machinelearningv1.GetPredictorKey(mlDep, p),
		}
		if grpcPort > 0 {
			c.serviceDetails[eSvcName].GrpcEndpoint = eSvcName + "." + eSvc.Namespace + ":" + strconv.Itoa(grpcPort)
		}
		if utils.GetEnv(ENV_ISTIO_ENABLED, "false") == "true" {
			vsvcs, dstRule := createExplainerIstioResources(eSvcName, p, mlDep, seldonId, getNamespace(mlDep), httpPort, grpcPort)
			c.virtualServices = append(c.virtualServices, vsvcs...)
			c.destinationRules = append(c.destinationRules, dstRule...)
		}
	}

	return nil
}

// Create istio virtual service and destination rule for explainer.
// Explainers need one each with no traffic-splitting
func createExplainerIstioResources(pSvcName string, p *machinelearningv1.PredictorSpec,
	mlDep *machinelearningv1.SeldonDeployment,
	seldonId string,
	namespace string,
	engine_http_port int,
	engine_grpc_port int) ([]*istio.VirtualService, []*istio.DestinationRule) {

	vsNameHttp := pSvcName + "-http"
	if len(vsNameHttp) > 63 {
		vsNameHttp = vsNameHttp[0:63]
		vsNameHttp = strings.TrimSuffix(vsNameHttp, "-")
	}

	vsNameGrpc := pSvcName + "-grpc"
	if len(vsNameGrpc) > 63 {
		vsNameGrpc = vsNameGrpc[0:63]
		vsNameGrpc = strings.TrimSuffix(vsNameGrpc, "-")
	}

	istio_gateway := utils.GetEnv(ENV_ISTIO_GATEWAY, "seldon-gateway")
	httpVsvc := &istio.VirtualService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      vsNameHttp,
			Namespace: namespace,
		},
		Spec: istio_networking.VirtualService{
			Hosts:    []string{"*"},
			Gateways: []string{getAnnotation(mlDep, ANNOTATION_ISTIO_GATEWAY, istio_gateway)},
			Http: []*istio_networking.HTTPRoute{
				{
					Match: []*istio_networking.HTTPMatchRequest{
						{
							Uri: &istio_networking.StringMatch{MatchType: &istio_networking.StringMatch_Prefix{Prefix: "/seldon/" + namespace + "/" + mlDep.GetName() + constants.ExplainerPathSuffix + "/" + p.Name + "/"}},
						},
					},
					Rewrite: &istio_networking.HTTPRewrite{Uri: "/"},
				},
			},
		},
	}

	grpcVsvc := &istio.VirtualService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      vsNameGrpc,
			Namespace: namespace,
		},
		Spec: istio_networking.VirtualService{
			Hosts:    []string{"*"},
			Gateways: []string{getAnnotation(mlDep, ANNOTATION_ISTIO_GATEWAY, istio_gateway)},
			Http: []*istio_networking.HTTPRoute{
				{
					Match: []*istio_networking.HTTPMatchRequest{
						{
							Uri: &istio_networking.StringMatch{MatchType: &istio_networking.StringMatch_Prefix{Prefix: "/seldon.protos.Seldon/"}},
							Headers: map[string]*istio_networking.StringMatch{
								"seldon":    &istio_networking.StringMatch{MatchType: &istio_networking.StringMatch_Exact{Exact: mlDep.GetName()}},
								"namespace": &istio_networking.StringMatch{MatchType: &istio_networking.StringMatch_Exact{Exact: namespace}},
							},
						},
					},
				},
			},
		},
	}

	routesHttp := make([]*istio_networking.HTTPRouteDestination, 1)
	routesGrpc := make([]*istio_networking.HTTPRouteDestination, 1)
	drules := make([]*istio.DestinationRule, 1)

	drule := &istio.DestinationRule{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pSvcName,
			Namespace: namespace,
		},
		Spec: istio_networking.DestinationRule{
			Host: pSvcName,
			Subsets: []*istio_networking.Subset{
				{
					Name: p.Name,
					Labels: map[string]string{
						"version": p.Labels["version"],
					},
				},
			},
		},
	}

	routesHttp[0] = &istio_networking.HTTPRouteDestination{
		Destination: &istio_networking.Destination{
			Host:   pSvcName,
			Subset: p.Name,
			Port: &istio_networking.PortSelector{
				Number: uint32(engine_http_port),
			},
		},
		Weight: int32(100),
	}
	routesGrpc[0] = &istio_networking.HTTPRouteDestination{
		Destination: &istio_networking.Destination{
			Host:   pSvcName,
			Subset: p.Name,
			Port: &istio_networking.PortSelector{
				Number: uint32(engine_grpc_port),
			},
		},
		Weight: int32(100),
	}
	drules[0] = drule

	httpVsvc.Spec.Http[0].Route = routesHttp
	grpcVsvc.Spec.Http[0].Route = routesGrpc
	vscs := make([]*istio.VirtualService, 0, 2)
	// explainer may not expose REST and grpc (presumably engine ensures predictors do?)
	if engine_http_port > 0 {
		vscs = append(vscs, httpVsvc)
	}
	if engine_grpc_port > 0 {
		vscs = append(vscs, grpcVsvc)
	}

	return vscs, drules
}
