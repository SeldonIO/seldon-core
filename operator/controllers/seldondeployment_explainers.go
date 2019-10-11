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
	"github.com/go-logr/logr"
	machinelearningv1alpha2 "github.com/seldonio/seldon-core/operator/api/v1alpha2"
	"github.com/seldonio/seldon-core/operator/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"knative.dev/pkg/apis/istio/common/v1alpha1"
	istio "knative.dev/pkg/apis/istio/v1alpha3"
	"strconv"
	"strings"
)

func createExplainer(r *SeldonDeploymentReconciler, mlDep *machinelearningv1alpha2.SeldonDeployment, p *machinelearningv1alpha2.PredictorSpec, c *components, pSvcName string, log logr.Logger) error {

	if p.Explainer.Type != "" {

		seldonId := machinelearningv1alpha2.GetSeldonDeploymentName(mlDep)

		depName := machinelearningv1alpha2.GetExplainerDeploymentName(mlDep.ObjectMeta.Name, p)

		explainerContainer := p.Explainer.ContainerSpec

		if explainerContainer.Name == "" {
			explainerContainer.Name = depName
		}

		if explainerContainer.ImagePullPolicy == "" {
			explainerContainer.ImagePullPolicy = corev1.PullIfNotPresent
		}

		if p.Graph.Endpoint == nil {
			p.Graph.Endpoint = &machinelearningv1alpha2.Endpoint{Type: machinelearningv1alpha2.REST}
		}

		if explainerContainer.Image == "" {
			// TODO: should use explainer type but this is the only one available currently
			explainerContainer.Image = "seldonio/alibiexplainer_grpc:0.2.3"
		}

		var portType string

		// explainer can get port from spec or from containerSpec or fall back on default
		var httpPort = 0
		var grpcPort = 0
		var portNum int32 = 9000
		if p.Explainer.Endpoint != nil && p.Explainer.Endpoint.ServicePort != 0 {
			portNum = p.Explainer.Endpoint.ServicePort
		}
		var pSvcEndpoint = ""
		customPort := getPort(portType, explainerContainer.Ports)

		if p.Explainer.Endpoint != nil && p.Explainer.Endpoint.Type == machinelearningv1alpha2.GRPC {
			portType = "grpc"
			grpcPort = int(portNum)
			pSvcEndpoint = c.serviceDetails[pSvcName].GrpcEndpoint
		} else {
			portType = "http"
			httpPort = int(portNum)
			//pSvcEndpoint = c.serviceDetails[pSvcName].HttpEndpoint
			// Default to grpc endpoint
			pSvcEndpoint = c.serviceDetails[pSvcName].GrpcEndpoint

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
			"--protocol=" + "seldon." + portType,
			"--http_port=" + strconv.Itoa(int(portNum)),
			"--use_grpc"}

		if p.Explainer.ModelUri != "" {
			explainerContainer.Args = append(explainerContainer.Args, "--storage_uri="+DefaultModelLocalMountPath)
		}

		explainerContainer.Args = append(explainerContainer.Args, strings.ToLower(p.Explainer.Type))

		if p.Explainer.Type == "anchor_images" {
			explainerContainer.Args = append(explainerContainer.Args, "--tf_data_type=float32")
		}
		for k, v := range p.Explainer.Config {
			//remote files in model location should get downloaded by initializer
			if p.Explainer.ModelUri != "" {
				v = strings.Replace(v, p.Explainer.ModelUri, "/mnt/models", 1)
			}
			arg := "--" + k + "=" + v
			explainerContainer.Args = append(explainerContainer.Args, arg)
		}
		// see https://github.com/cliveseldon/kfserving/tree/explainer_update_jul/docs/samples/explanation/income for more

		// Add Environment Variables - TODO: are these needed
		if !utils.HasEnvVar(explainerContainer.Env, machinelearningv1alpha2.ENV_PREDICTIVE_UNIT_SERVICE_PORT) {
			explainerContainer.Env = append(explainerContainer.Env, []corev1.EnvVar{
				corev1.EnvVar{Name: machinelearningv1alpha2.ENV_PREDICTIVE_UNIT_SERVICE_PORT, Value: strconv.Itoa(int(portNum))},
				corev1.EnvVar{Name: machinelearningv1alpha2.ENV_PREDICTIVE_UNIT_ID, Value: explainerContainer.Name},
				corev1.EnvVar{Name: machinelearningv1alpha2.ENV_PREDICTOR_ID, Value: p.Name},
				corev1.EnvVar{Name: machinelearningv1alpha2.ENV_SELDON_DEPLOYMENT_ID, Value: mlDep.ObjectMeta.Name},
			}...)
		}

		seldonPodSpec := machinelearningv1alpha2.SeldonPodSpec{Spec: corev1.PodSpec{
			Containers: []corev1.Container{explainerContainer},
		}}

		deploy := createDeploymentWithoutEngine(depName, seldonId, &seldonPodSpec, p, mlDep)

		if p.Explainer.ModelUri != "" {
			var err error
			deploy, err = InjectModelInitializer(deploy, explainerContainer.Name, p.Explainer.ModelUri, p.Explainer.ServiceAccountName, p.Explainer.EnvSecretRefName, r.Client)
			if err != nil {
				return err
			}
		}

		// for explainer use same service name as its Deployment
		eSvcName := machinelearningv1alpha2.GetExplainerDeploymentName(mlDep.ObjectMeta.Name, p)

		deploy.ObjectMeta.Labels[machinelearningv1alpha2.Label_seldon_app] = eSvcName
		deploy.Spec.Template.ObjectMeta.Labels[machinelearningv1alpha2.Label_seldon_app] = eSvcName

		c.deployments = append(c.deployments, deploy)

		// Use seldondeployment name dash explainer as the external service name. This should allow canarying.
		eSvc, err := createPredictorService(eSvcName, seldonId, p, mlDep, httpPort, grpcPort, mlDep.ObjectMeta.Name+"-explainer", log)
		if err != nil {
			return err
		}
		c.services = append(c.services, eSvc)
		c.serviceDetails[eSvcName] = &machinelearningv1alpha2.ServiceStatus{
			SvcName:      eSvcName,
			HttpEndpoint: eSvcName + "." + eSvc.Namespace + ":" + strconv.Itoa(httpPort),
			ExplainerFor: machinelearningv1alpha2.GetPredictorKey(mlDep, p),
		}
		if grpcPort > 0 {
			c.serviceDetails[eSvcName].GrpcEndpoint = eSvcName + "." + eSvc.Namespace + ":" + strconv.Itoa(grpcPort)
		}
		if GetEnv(ENV_ISTIO_ENABLED, "false") == "true" {
			vsvcs, dstRule := createExplainerIstioResources(eSvcName, p, mlDep, seldonId, getNamespace(mlDep), httpPort, grpcPort)
			c.virtualServices = append(c.virtualServices, vsvcs...)
			c.destinationRules = append(c.destinationRules, dstRule...)
		}
	}

	return nil
}

// Create istio virtual service and destination rule for explainer.
// Explainers need one each with no traffic-splitting
func createExplainerIstioResources(pSvcName string, p *machinelearningv1alpha2.PredictorSpec,
	mlDep *machinelearningv1alpha2.SeldonDeployment,
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

	istio_gateway := GetEnv(ENV_ISTIO_GATEWAY, "seldon-gateway")
	httpVsvc := &istio.VirtualService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      vsNameHttp,
			Namespace: namespace,
		},
		Spec: istio.VirtualServiceSpec{
			Hosts:    []string{"*"},
			Gateways: []string{getAnnotation(mlDep, ANNOTATION_ISTIO_GATEWAY, istio_gateway)},
			HTTP: []istio.HTTPRoute{
				{
					Match: []istio.HTTPMatchRequest{
						{
							URI: &v1alpha1.StringMatch{Prefix: "/seldon/" + namespace + "/" + mlDep.Name + "/" + p.Name + "/explainer/"},
						},
					},
					Rewrite: &istio.HTTPRewrite{URI: "/"},
				},
			},
		},
	}

	grpcVsvc := &istio.VirtualService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      vsNameGrpc,
			Namespace: namespace,
		},
		Spec: istio.VirtualServiceSpec{
			Hosts:    []string{"*"},
			Gateways: []string{getAnnotation(mlDep, ANNOTATION_ISTIO_GATEWAY, istio_gateway)},
			HTTP: []istio.HTTPRoute{
				{
					Match: []istio.HTTPMatchRequest{
						{
							URI: &v1alpha1.StringMatch{Prefix: "/seldon.protos.Seldon/"},
							Headers: map[string]v1alpha1.StringMatch{
								"seldon":    v1alpha1.StringMatch{Exact: mlDep.Name}, //TODO: change this?
								"namespace": v1alpha1.StringMatch{Exact: namespace},
							},
						},
					},
				},
			},
		},
	}

	routesHttp := make([]istio.HTTPRouteDestination, 1)
	routesGrpc := make([]istio.HTTPRouteDestination, 1)
	drules := make([]*istio.DestinationRule, 1)

	drule := &istio.DestinationRule{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pSvcName,
			Namespace: namespace,
		},
		Spec: istio.DestinationRuleSpec{
			Host: pSvcName,
			Subsets: []istio.Subset{
				{
					Name: p.Name,
					Labels: map[string]string{
						"version": p.Labels["version"],
					},
				},
			},
		},
	}

	routesHttp[0] = istio.HTTPRouteDestination{
		Destination: istio.Destination{
			Host:   pSvcName,
			Subset: p.Name,
			Port: istio.PortSelector{
				Number: uint32(engine_http_port),
			},
		},
		Weight: int(100),
	}
	routesGrpc[0] = istio.HTTPRouteDestination{
		Destination: istio.Destination{
			Host:   pSvcName,
			Subset: p.Name,
			Port: istio.PortSelector{
				Number: uint32(engine_grpc_port),
			},
		},
		Weight: int(100),
	}
	drules[0] = drule

	httpVsvc.Spec.HTTP[0].Route = routesHttp
	grpcVsvc.Spec.HTTP[0].Route = routesGrpc
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
