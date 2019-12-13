/*
Copyright 2019 The Seldon Authors.

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
	"bytes"
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/seldonio/seldon-core/operator/constants"
	"github.com/seldonio/seldon-core/operator/utils"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"knative.dev/pkg/kmp"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/go-logr/logr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	machinelearningv1 "github.com/seldonio/seldon-core/operator/apis/machinelearning/v1"

	"encoding/json"

	appsv1 "k8s.io/api/apps/v1"
	autoscaling "k8s.io/api/autoscaling/v2beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis/istio/common/v1alpha1"
	istio "knative.dev/pkg/apis/istio/v1alpha3"
)

const (
	ENV_DEFAULT_ENGINE_SERVER_PORT      = "ENGINE_SERVER_PORT"
	ENV_DEFAULT_ENGINE_SERVER_GRPC_PORT = "ENGINE_SERVER_GRPC_PORT"
	ENV_CONTROLLER_ID                   = "CONTROLLER_ID"

	DEFAULT_ENGINE_CONTAINER_PORT = 8000
	DEFAULT_ENGINE_GRPC_PORT      = 5001

	AMBASSADOR_ANNOTATION = "getambassador.io/config"
	LABEL_CONTROLLER_ID   = "seldon.io/controller-id"
)

// SeldonDeploymentReconciler reconciles a SeldonDeployment object
type SeldonDeploymentReconciler struct {
	client.Client
	Log       logr.Logger
	Scheme    *runtime.Scheme
	Namespace string
}

//---------------- Old part

type components struct {
	serviceDetails   map[string]*machinelearningv1.ServiceStatus
	deployments      []*appsv1.Deployment
	services         []*corev1.Service
	hpas             []*autoscaling.HorizontalPodAutoscaler
	virtualServices  []*istio.VirtualService
	destinationRules []*istio.DestinationRule
}

type serviceDetails struct {
	svcName        string
	deploymentName string
	svcUrl         string
	ambassadorUrl  string
}

type httpGrpcPorts struct {
	httpPort int
	grpcPort int
}

func createHpa(podSpec *machinelearningv1.SeldonPodSpec, deploymentName string, seldonId string, namespace string) *autoscaling.HorizontalPodAutoscaler {
	hpa := autoscaling.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deploymentName,
			Namespace: namespace,
			Labels:    map[string]string{machinelearningv1.Label_seldon_id: seldonId},
		},
		Spec: autoscaling.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: autoscaling.CrossVersionObjectReference{
				Name:       deploymentName,
				APIVersion: "extensions/v1beta1",
				Kind:       "Deployment",
			},
			MaxReplicas: podSpec.HpaSpec.MaxReplicas,
			Metrics:     podSpec.HpaSpec.Metrics,
		},
	}
	if podSpec.HpaSpec.MinReplicas != nil {
		hpa.Spec.MinReplicas = podSpec.HpaSpec.MinReplicas
	}

	return &hpa
}

// Create istio virtual service and destination rule.
// Creates routes for each predictor with traffic weight split
func createIstioResources(mlDep *machinelearningv1.SeldonDeployment,
	seldonId string,
	namespace string,
	ports []httpGrpcPorts,
	httpAllowed bool,
	grpcAllowed bool) ([]*istio.VirtualService, []*istio.DestinationRule) {

	istio_gateway := GetEnv(ENV_ISTIO_GATEWAY, "seldon-gateway")
	httpVsvc := &istio.VirtualService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      seldonId + "-http",
			Namespace: namespace,
		},
		Spec: istio.VirtualServiceSpec{
			Hosts:    []string{"*"},
			Gateways: []string{getAnnotation(mlDep, ANNOTATION_ISTIO_GATEWAY, istio_gateway)},
			HTTP: []istio.HTTPRoute{
				{
					Match: []istio.HTTPMatchRequest{
						{
							URI: &v1alpha1.StringMatch{Prefix: "/seldon/" + namespace + "/" + mlDep.Name + "/"},
						},
					},
					Rewrite: &istio.HTTPRewrite{URI: "/"},
				},
			},
		},
	}

	grpcVsvc := &istio.VirtualService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      seldonId + "-grpc",
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
								"seldon":    v1alpha1.StringMatch{Exact: mlDep.Name},
								"namespace": v1alpha1.StringMatch{Exact: namespace},
							},
						},
					},
				},
			},
		},
	}

	// shadows don't get destinations in the vs as a shadow is a mirror instead
	var shadows int = 0
	for i := 0; i < len(mlDep.Spec.Predictors); i++ {
		p := mlDep.Spec.Predictors[i]
		if p.Shadow == true {
			shadows += 1
		}
	}

	routesHttp := make([]istio.HTTPRouteDestination, len(mlDep.Spec.Predictors)-shadows)
	routesGrpc := make([]istio.HTTPRouteDestination, len(mlDep.Spec.Predictors)-shadows)

	// the shdadow/mirror entry does need a DestinationRule though
	drules := make([]*istio.DestinationRule, len(mlDep.Spec.Predictors))
	for i := 0; i < len(mlDep.Spec.Predictors); i++ {

		p := mlDep.Spec.Predictors[i]
		pSvcName := machinelearningv1.GetPredictorKey(mlDep, &p)

		drule := &istio.DestinationRule{
			ObjectMeta: metav1.ObjectMeta{
				Name:      pSvcName,
				Namespace: namespace,
			},
			Spec: istio.DestinationRuleSpec{
				Host: pSvcName,
				TrafficPolicy: &istio.TrafficPolicy{
					TLS: &istio.TLSSettings{
						Mode: istio.TLSmodeIstioMutual,
					},
				},
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

		drules[i] = drule

		if p.Shadow == true {
			//if there's a shadow then add a mirror section to the VirtualService

			httpVsvc.Spec.HTTP[0].Mirror = &istio.Destination{
				Host:   pSvcName,
				Subset: p.Name,
				Port: istio.PortSelector{
					Number: uint32(ports[i].httpPort),
				},
			}

			grpcVsvc.Spec.HTTP[0].Mirror = &istio.Destination{
				Host:   pSvcName,
				Subset: p.Name,
				Port: istio.PortSelector{
					Number: uint32(ports[i].grpcPort),
				},
			}

			continue
		}

		//we split by adding different routes with their own Weights
		//so not by tag - different destinations (like https://istio.io/docs/tasks/traffic-management/traffic-shifting/) distinguished by host
		routesHttp[i] = istio.HTTPRouteDestination{
			Destination: istio.Destination{
				Host:   pSvcName,
				Subset: p.Name,
				Port: istio.PortSelector{
					Number: uint32(ports[i].httpPort),
				},
			},
			Weight: int(p.Traffic),
		}
		routesGrpc[i] = istio.HTTPRouteDestination{
			Destination: istio.Destination{
				Host:   pSvcName,
				Subset: p.Name,
				Port: istio.PortSelector{
					Number: uint32(ports[i].grpcPort),
				},
			},
			Weight: int(p.Traffic),
		}

	}
	httpVsvc.Spec.HTTP[0].Route = routesHttp
	grpcVsvc.Spec.HTTP[0].Route = routesGrpc

	if httpAllowed && grpcAllowed {
		vscs := make([]*istio.VirtualService, 2)
		vscs[0] = httpVsvc
		vscs[1] = grpcVsvc
		return vscs, drules
	} else if httpAllowed {
		vscs := make([]*istio.VirtualService, 1)
		vscs[0] = httpVsvc
		return vscs, drules
	} else {
		vscs := make([]*istio.VirtualService, 1)
		vscs[0] = grpcVsvc
		return vscs, drules
	}
}

func getEngineHttpPort() (engine_http_port int, err error) {
	// Get engine http port from environment or use default
	engine_http_port = DEFAULT_ENGINE_CONTAINER_PORT
	var env_engine_http_port = GetEnv(ENV_DEFAULT_ENGINE_SERVER_PORT, "")
	if env_engine_http_port != "" {
		engine_http_port, err = strconv.Atoi(env_engine_http_port)
		if err != nil {
			return 0, err
		}
	}
	return engine_http_port, nil
}

func getEngineGrpcPort() (engine_grpc_port int, err error) {
	// Get engine grpc port from environment or use default
	engine_grpc_port = DEFAULT_ENGINE_GRPC_PORT
	var env_engine_grpc_port = GetEnv(ENV_DEFAULT_ENGINE_SERVER_GRPC_PORT, "")
	if env_engine_grpc_port != "" {
		engine_grpc_port, err = strconv.Atoi(env_engine_grpc_port)
		if err != nil {
			return 0, err
		}
	}
	return engine_grpc_port, nil
}

// Create all the components (Deployments, Services etc)
func createComponents(r *SeldonDeploymentReconciler, mlDep *machinelearningv1.SeldonDeployment, log logr.Logger) (*components, error) {
	c := components{}
	c.serviceDetails = map[string]*machinelearningv1.ServiceStatus{}
	seldonId := machinelearningv1.GetSeldonDeploymentName(mlDep)
	namespace := getNamespace(mlDep)

	engine_http_port, err := getEngineHttpPort()
	if err != nil {
		return nil, err
	}

	engine_grpc_port, err := getEngineGrpcPort()
	if err != nil {
		return nil, err
	}

	// variables to collect what ports will be exposed and whether we should expose http and grpc externally
	// If one of the predictors has noEngine then only one of http or grpc should be allowed dependent on
	// the type of the noEngine model: whether it is http or grpc
	externalPorts := make([]httpGrpcPorts, len(mlDep.Spec.Predictors))
	grpcAllowed := true
	httpAllowed := true
	// Attempt to set httpAllowed and grpcAllowed to false if we have an noEngine predictor
	for i := 0; i < len(mlDep.Spec.Predictors); i++ {
		p := mlDep.Spec.Predictors[i]
		_, noEngine := p.Annotations[machinelearningv1.ANNOTATION_NO_ENGINE]
		if noEngine && len(p.ComponentSpecs) > 0 && len(p.ComponentSpecs[0].Spec.Containers) > 0 {
			pu := machinelearningv1.GetPredictiveUnit(p.Graph, p.ComponentSpecs[0].Spec.Containers[0].Name)
			if pu != nil {
				if pu.Endpoint != nil && pu.Endpoint.Type == machinelearningv1.GRPC {
					httpAllowed = false
				}
				if pu.Endpoint == nil || pu.Endpoint.Type == machinelearningv1.REST {
					grpcAllowed = false
				}
			}
		}
	}

	for i := 0; i < len(mlDep.Spec.Predictors); i++ {
		p := mlDep.Spec.Predictors[i]
		_, noEngine := p.Annotations[machinelearningv1.ANNOTATION_NO_ENGINE]
		pSvcName := machinelearningv1.GetPredictorKey(mlDep, &p)
		log.Info("pSvcName", "val", pSvcName)
		// Add engine deployment if separate
		_, hasSeparateEnginePod := mlDep.Spec.Annotations[machinelearningv1.ANNOTATION_SEPARATE_ENGINE]
		if hasSeparateEnginePod && !noEngine {
			deploy, err := createEngineDeployment(mlDep, &p, pSvcName, engine_http_port, engine_grpc_port)
			if err != nil {
				return nil, err
			}
			c.deployments = append(c.deployments, deploy)
		}

		for j := 0; j < len(p.ComponentSpecs); j++ {
			cSpec := mlDep.Spec.Predictors[i].ComponentSpecs[j]

			// if no container spec then nothing to create at this point - prepackaged model server cases handled later
			if len(cSpec.Spec.Containers) == 0 {
				continue
			}

			// create Deployment from podspec
			depName := machinelearningv1.GetDeploymentName(mlDep, p, cSpec)
			deploy := createDeploymentWithoutEngine(depName, seldonId, cSpec, &p, mlDep)

			// Add HPA if needed
			if cSpec.HpaSpec != nil {
				c.hpas = append(c.hpas, createHpa(cSpec, depName, seldonId, namespace))
			} else {
				deploy.Spec.Replicas = &p.Replicas
			}

			// create services for each container
			for k := 0; k < len(cSpec.Spec.Containers); k++ {
				var con *corev1.Container
				// get the container on the created deployment, as createDeploymentWithoutEngine will have created as a copy of the spec in the manifest and added defaults to it
				// we need the reference as we may have to modify the container when creating the Service (e.g. to add probes)
				con = utils.GetContainerForDeployment(deploy, cSpec.Spec.Containers[k].Name)

				// engine will later get a special predictor service as it is entrypoint for graph
				// and no need to expose tfserving container as it's accessed via proxy
				if con.Name != EngineContainerName && con.Name != constants.TFServingContainerName {

					// service for hitting a model directly, not via engine - also adds ports to container if needed
					svc := createContainerService(deploy, p, mlDep, con, c)
					if svc != nil {
						c.services = append(c.services, svc)
					} else {
						// a user-supplied container may not be a pu so we may not create service for that
						log.Info("Not creating container service for " + con.Name)
						continue
					}

					if noEngine {
						deploy.ObjectMeta.Labels[machinelearningv1.Label_seldon_app] = pSvcName
						deploy.Spec.Selector.MatchLabels[machinelearningv1.Label_seldon_app] = pSvcName
						deploy.Spec.Template.ObjectMeta.Labels[machinelearningv1.Label_seldon_app] = pSvcName

						port := int(svc.Spec.Ports[0].Port)

						if svc.Spec.Ports[0].Name == "grpc" {
							httpAllowed = false
							externalPorts[i] = httpGrpcPorts{httpPort: 0, grpcPort: port}
							psvc, err := createPredictorService(pSvcName, seldonId, &p, mlDep, 0, port, "", log)
							if err != nil {
								return nil, err
							}

							c.services = append(c.services, psvc)

							c.serviceDetails[pSvcName] = &machinelearningv1.ServiceStatus{
								SvcName:      pSvcName,
								GrpcEndpoint: pSvcName + "." + namespace + ":" + strconv.Itoa(port),
							}
						} else {
							externalPorts[i] = httpGrpcPorts{httpPort: port, grpcPort: 0}
							grpcAllowed = false
							psvc, err := createPredictorService(pSvcName, seldonId, &p, mlDep, port, 0, "", log)
							if err != nil {
								return nil, err
							}

							c.services = append(c.services, psvc)

							c.serviceDetails[pSvcName] = &machinelearningv1.ServiceStatus{
								SvcName:      pSvcName,
								HttpEndpoint: pSvcName + "." + namespace + ":" + strconv.Itoa(port),
							}
						}

					}
				}
			}
			c.deployments = append(c.deployments, deploy)
		}

		err = createStandaloneModelServers(r, mlDep, &p, &c, p.Graph)
		if err != nil {
			return nil, err
		}

		if !noEngine {

			// Add service orchestrator to engine deployment if needed
			if !hasSeparateEnginePod {
				var deploy *appsv1.Deployment
				found := false

				// find the pu that the webhook marked as localhost as its corresponding deployment should get the engine
				pu := machinelearningv1.GetEnginePredictiveUnit(p.Graph)
				if pu == nil {
					// below should never happen - if it did would suggest problem in webhook
					return nil, fmt.Errorf("Engine not separate and no pu with localhost service - not clear where to inject engine")
				}
				// find the deployment with a container for the pu marked for engine
				for i, _ := range c.deployments {
					dep := c.deployments[i]
					for _, con := range dep.Spec.Template.Spec.Containers {
						if strings.Compare(con.Name, pu.Name) == 0 {
							deploy = dep
							found = true
						}
					}
				}

				if !found {
					// by this point we should have created the Deployment corresponding to the pu marked localhost - if we haven't something has gone wrong
					return nil, fmt.Errorf("Engine not separate and no deployment for pu with localhost service - not clear where to inject engine")
				}
				err := addEngineToDeployment(mlDep, &p, engine_http_port, engine_grpc_port, pSvcName, deploy)
				if err != nil {
					return nil, err
				}

			}

			//Create Service for Predictor - exposed externally (ambassador or istio) and points at engine
			httpPort := engine_http_port
			if httpAllowed == false {
				httpPort = 0
			}
			grpcPort := engine_grpc_port
			if grpcAllowed == false {
				grpcPort = 0
			}
			psvc, err := createPredictorService(pSvcName, seldonId, &p, mlDep, httpPort, grpcPort, "", log)
			if err != nil {

				return nil, err
			}

			c.services = append(c.services, psvc)
			if httpAllowed && grpcAllowed {
				c.serviceDetails[pSvcName] = &machinelearningv1.ServiceStatus{
					SvcName:      pSvcName,
					HttpEndpoint: pSvcName + "." + namespace + ":" + strconv.Itoa(engine_http_port),
					GrpcEndpoint: pSvcName + "." + namespace + ":" + strconv.Itoa(engine_grpc_port),
				}
			} else if httpAllowed {
				c.serviceDetails[pSvcName] = &machinelearningv1.ServiceStatus{
					SvcName:      pSvcName,
					HttpEndpoint: pSvcName + "." + namespace + ":" + strconv.Itoa(engine_http_port),
				}
			} else if grpcAllowed {
				c.serviceDetails[pSvcName] = &machinelearningv1.ServiceStatus{
					SvcName:      pSvcName,
					GrpcEndpoint: pSvcName + "." + namespace + ":" + strconv.Itoa(engine_grpc_port),
				}
			}

			externalPorts[i] = httpGrpcPorts{httpPort: httpPort, grpcPort: grpcPort}
		}

		err = createExplainer(r, mlDep, &p, &c, pSvcName, log)
		if err != nil {
			return nil, err
		}
	}

	//TODO Fixme - not changed to handle per predictor scenario
	if GetEnv(ENV_ISTIO_ENABLED, "false") == "true" {
		vsvcs, dstRule := createIstioResources(mlDep, seldonId, namespace, externalPorts, httpAllowed, grpcAllowed)
		c.virtualServices = append(c.virtualServices, vsvcs...)
		c.destinationRules = append(c.destinationRules, dstRule...)
	}
	return &c, nil
}

//Creates Service for Predictor - exposed externally (ambassador or istio)
func createPredictorService(pSvcName string, seldonId string, p *machinelearningv1.PredictorSpec,
	mlDep *machinelearningv1.SeldonDeployment,
	engine_http_port int,
	engine_grpc_port int,
	ambassadorNameOverride string,
	log logr.Logger) (pSvc *corev1.Service, err error) {
	namespace := getNamespace(mlDep)

	psvc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pSvcName,
			Namespace: namespace,
			Labels: map[string]string{machinelearningv1.Label_seldon_app: pSvcName,
				machinelearningv1.Label_seldon_id: seldonId},
		},
		Spec: corev1.ServiceSpec{
			Selector:        map[string]string{machinelearningv1.Label_seldon_app: pSvcName},
			SessionAffinity: corev1.ServiceAffinityNone,
			Type:            corev1.ServiceTypeClusterIP,
		},
	}

	if engine_http_port != 0 && len(psvc.Spec.Ports) == 0 {
		psvc.Spec.Ports = append(psvc.Spec.Ports, corev1.ServicePort{Protocol: corev1.ProtocolTCP, Port: int32(engine_http_port), TargetPort: intstr.FromInt(engine_http_port), Name: "http"})
	}

	if engine_grpc_port != 0 && len(psvc.Spec.Ports) < 2 {
		psvc.Spec.Ports = append(psvc.Spec.Ports, corev1.ServicePort{Protocol: corev1.ProtocolTCP, Port: int32(engine_grpc_port), TargetPort: intstr.FromInt(engine_grpc_port), Name: "grpc"})
	}

	if GetEnv("AMBASSADOR_ENABLED", "false") == "true" {
		psvc.Annotations = make(map[string]string)
		//Create top level Service
		ambassadorConfig, err := getAmbassadorConfigs(mlDep, p, pSvcName, engine_http_port, engine_grpc_port, ambassadorNameOverride)
		if err != nil {
			return nil, err
		}
		psvc.Annotations[AMBASSADOR_ANNOTATION] = ambassadorConfig
	}

	if getAnnotation(mlDep, machinelearningv1.ANNOTATION_HEADLESS_SVC, "false") != "false" {
		log.Info("Creating Headless SVC")
		psvc.Spec.ClusterIP = "None"
	}

	return psvc, err
}

// service for hitting a model directly, not via engine - not exposed externally, also adds probes
func createContainerService(deploy *appsv1.Deployment, p machinelearningv1.PredictorSpec, mlDep *machinelearningv1.SeldonDeployment, con *corev1.Container, c components) *corev1.Service {
	containerServiceKey := machinelearningv1.GetPredictorServiceNameKey(con)
	containerServiceValue := machinelearningv1.GetContainerServiceName(mlDep, p, con)
	pu := machinelearningv1.GetPredictiveUnit(p.Graph, con.Name)

	// only create services for containers defined as pus in the graph
	if pu == nil {
		return nil
	}
	namespace := getNamespace(mlDep)
	portType := "http"
	var portNum int32
	portNum = 0
	existingPort := machinelearningv1.GetPort(portType, con.Ports)
	if existingPort != nil {
		portNum = existingPort.ContainerPort
	}

	if pu.Endpoint.Type == machinelearningv1.GRPC {
		portType = "grpc"
	}

	// pu should have a port set by seldondeployment_create_update_handler.go (if not by user)
	// that mutator modifies SeldonDeployment and fires before this controller
	if pu.Endpoint.ServicePort != 0 {
		portNum = pu.Endpoint.ServicePort
	}

	if portNum == 0 {
		// should have port by now
		// if we don't know what it would respond to so can't create a service for it
		return nil
	}

	if portType == "grpc" {
		c.serviceDetails[containerServiceValue] = &machinelearningv1.ServiceStatus{
			SvcName:      containerServiceValue,
			GrpcEndpoint: containerServiceValue + "." + namespace + ":" + strconv.Itoa(int(portNum))}
	} else {
		c.serviceDetails[containerServiceValue] = &machinelearningv1.ServiceStatus{
			SvcName:      containerServiceValue,
			HttpEndpoint: containerServiceValue + "." + namespace + ":" + strconv.Itoa(int(portNum))}
	}

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      containerServiceValue,
			Namespace: namespace,
			Labels:    map[string]string{containerServiceKey: containerServiceValue, machinelearningv1.Label_seldon_id: mlDep.Spec.Name},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Protocol:   corev1.ProtocolTCP,
					Port:       portNum,
					TargetPort: intstr.FromInt(int(portNum)),
					Name:       portType,
				},
			},
			Type:            corev1.ServiceTypeClusterIP,
			Selector:        map[string]string{containerServiceKey: containerServiceValue},
			SessionAffinity: corev1.ServiceAffinityNone,
		},
	}

	//Add labels for this service to deployment
	deploy.ObjectMeta.Labels[containerServiceKey] = containerServiceValue
	deploy.Spec.Selector.MatchLabels[containerServiceKey] = containerServiceValue
	deploy.Spec.Template.ObjectMeta.Labels[containerServiceKey] = containerServiceValue

	if existingPort == nil || con.Ports == nil {
		con.Ports = append(con.Ports, corev1.ContainerPort{Name: portType, ContainerPort: portNum, Protocol: corev1.ProtocolTCP})
	}

	if con.LivenessProbe == nil {
		con.LivenessProbe = &corev1.Probe{Handler: corev1.Handler{TCPSocket: &corev1.TCPSocketAction{Port: intstr.FromString(portType)}}, InitialDelaySeconds: 60, PeriodSeconds: 5, SuccessThreshold: 1, FailureThreshold: 3, TimeoutSeconds: 1}
	}
	if con.ReadinessProbe == nil {
		con.ReadinessProbe = &corev1.Probe{Handler: corev1.Handler{TCPSocket: &corev1.TCPSocketAction{Port: intstr.FromString(portType)}}, InitialDelaySeconds: 20, PeriodSeconds: 5, SuccessThreshold: 1, FailureThreshold: 3, TimeoutSeconds: 1}
	}

	// Add livecycle probe
	if con.Lifecycle == nil {
		con.Lifecycle = &corev1.Lifecycle{PreStop: &corev1.Handler{Exec: &corev1.ExecAction{Command: []string{"/bin/sh", "-c", "/bin/sleep 10"}}}}
	}

	// Add Environment Variables
	if !utils.HasEnvVar(con.Env, machinelearningv1.ENV_PREDICTIVE_UNIT_SERVICE_PORT) {
		con.Env = append(con.Env, []corev1.EnvVar{
			corev1.EnvVar{Name: machinelearningv1.ENV_PREDICTIVE_UNIT_SERVICE_PORT, Value: strconv.Itoa(int(portNum))},
			corev1.EnvVar{Name: machinelearningv1.ENV_PREDICTIVE_UNIT_ID, Value: con.Name},
			corev1.EnvVar{Name: machinelearningv1.ENV_PREDICTOR_ID, Value: p.Name},
			corev1.EnvVar{Name: machinelearningv1.ENV_SELDON_DEPLOYMENT_ID, Value: mlDep.ObjectMeta.Name},
		}...)
	}

	if pu != nil && len(pu.Parameters) > 0 {
		if !utils.HasEnvVar(con.Env, machinelearningv1.ENV_PREDICTIVE_UNIT_PARAMETERS) {
			con.Env = append(con.Env, corev1.EnvVar{Name: machinelearningv1.ENV_PREDICTIVE_UNIT_PARAMETERS, Value: utils.GetPredictiveUnitAsJson(pu.Parameters)})
		}
	}

	return svc
}

func createDeploymentWithoutEngine(depName string, seldonId string, seldonPodSpec *machinelearningv1.SeldonPodSpec, p *machinelearningv1.PredictorSpec, mlDep *machinelearningv1.SeldonDeployment) *appsv1.Deployment {
	deploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      depName,
			Namespace: getNamespace(mlDep),
			Labels:    map[string]string{machinelearningv1.Label_seldon_id: seldonId, "app": depName, "fluentd": "true"},
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{machinelearningv1.Label_seldon_id: seldonId},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      map[string]string{machinelearningv1.Label_seldon_id: seldonId, "app": depName, "fluentd": "true"},
					Annotations: mlDep.Spec.Annotations,
				},
			},
			Strategy: appsv1.DeploymentStrategy{RollingUpdate: &appsv1.RollingUpdateDeployment{MaxUnavailable: &intstr.IntOrString{StrVal: "10%"}}},
		},
	}

	if p.Shadow == true {
		deploy.Spec.Template.ObjectMeta.Labels["shadow"] = "true"
	}

	if seldonPodSpec != nil {
		deploy.Spec.Template.Spec = seldonPodSpec.Spec
		// add more annotations
		for k, v := range seldonPodSpec.Metadata.Annotations {
			deploy.Spec.Template.ObjectMeta.Annotations[k] = v
		}
	}

	// add predictor labels
	for k, v := range p.Labels {
		deploy.ObjectMeta.Labels[k] = v
		deploy.Spec.Template.ObjectMeta.Labels[k] = v
	}

	//Add some default to help with diffs in controller
	if deploy.Spec.Template.Spec.RestartPolicy == "" {
		deploy.Spec.Template.Spec.RestartPolicy = corev1.RestartPolicyAlways
	}
	if deploy.Spec.Template.Spec.DNSPolicy == "" {
		deploy.Spec.Template.Spec.DNSPolicy = corev1.DNSClusterFirst
	}
	if deploy.Spec.Template.Spec.SchedulerName == "" {
		deploy.Spec.Template.Spec.SchedulerName = "default-scheduler"
	}
	if deploy.Spec.Template.Spec.SecurityContext == nil {
		deploy.Spec.Template.Spec.SecurityContext = &corev1.PodSecurityContext{}
	}
	var terminationGracePeriod int64 = 20
	deploy.Spec.Template.Spec.TerminationGracePeriodSeconds = &terminationGracePeriod

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

	return deploy
}

func getPort(name string, ports []corev1.ContainerPort) *corev1.ContainerPort {
	for i := 0; i < len(ports); i++ {
		if ports[i].Name == name {
			return &ports[i]
		}
	}
	return nil
}

// Create Services specified in components.
func createIstioServices(r *SeldonDeploymentReconciler, components *components, instance *machinelearningv1.SeldonDeployment, log logr.Logger) (bool, error) {
	ready := true
	for _, svc := range components.virtualServices {
		if err := controllerutil.SetControllerReference(instance, svc, r.Scheme); err != nil {
			return ready, err
		}
		found := &istio.VirtualService{}
		err := r.Get(context.TODO(), types.NamespacedName{Name: svc.Name, Namespace: svc.Namespace}, found)
		if err != nil && errors.IsNotFound(err) {
			ready = false
			log.Info("Creating Virtual Service", "namespace", svc.Namespace, "name", svc.Name)
			err = r.Create(context.TODO(), svc)
			if err != nil {
				return ready, err
			}

		} else if err != nil {
			return ready, err
		} else {
			// Update the found object and write the result back if there are any changes
			if !equality.Semantic.DeepEqual(svc.Spec, found.Spec) {
				desiredSvc := found.DeepCopy()
				found.Spec = svc.Spec
				log.Info("Updating Virtual Service", "namespace", svc.Namespace, "name", svc.Name)
				err = r.Update(context.TODO(), found)
				if err != nil {
					return ready, err
				}

				// Check if what came back from server modulo the defaults applied by k8s is the same or not
				if !equality.Semantic.DeepEqual(desiredSvc.Spec, found.Spec) {
					ready = false

					//For debugging we will show the difference
					diff, err := kmp.SafeDiff(desiredSvc.Spec, found.Spec)
					if err != nil {
						log.Error(err, "Failed to diff")
					} else {
						log.Info(fmt.Sprintf("Difference in VSVC: %v", diff))
					}
				} else {
					log.Info("The VSVC are the same - api server defaults ignored")
				}
			} else {
				log.Info("Found identical Virtual Service", "namespace", found.Namespace, "name", found.Name)

				if instance.Status.ServiceStatus == nil {
					instance.Status.ServiceStatus = map[string]machinelearningv1.ServiceStatus{}
				}

				/*
					if _, ok := instance.Status.ServiceStatus[found.Name]; !ok {
						instance.Status.ServiceStatus[found.Name] = *components.serviceDetails[found.Spec.HTTP[0].Route[0].Destination.Host]
						err = r.Status().Update(context.Background(), instance)
						if err != nil {
							return ready, err
						}
					}
				*/
			}
		}

	}

	for _, drule := range components.destinationRules {

		if err := controllerutil.SetControllerReference(instance, drule, r.Scheme); err != nil {
			return ready, err
		}
		found := &istio.DestinationRule{}
		err := r.Get(context.TODO(), types.NamespacedName{Name: drule.Name, Namespace: drule.Namespace}, found)
		if err != nil && errors.IsNotFound(err) {
			ready = false
			log.Info("Creating Istio Destination Rule", "namespace", drule.Namespace, "name", drule.Name)
			err = r.Create(context.TODO(), drule)
			if err != nil {
				return ready, err
			}

		} else if err != nil {
			return ready, err
		} else {
			// Update the found object and write the result back if there are any changes
			if !equality.Semantic.DeepEqual(drule.Spec, found.Spec) {
				desiredDrule := found.DeepCopy()
				found.Spec = drule.Spec
				log.Info("Updating Istio Destination Rule", "namespace", drule.Namespace, "name", drule.Name)
				err = r.Update(context.TODO(), found)
				if err != nil {
					return ready, err
				}

				// Check if what came back from server modulo the defaults applied by k8s is the same or not
				if !equality.Semantic.DeepEqual(desiredDrule.Spec, found.Spec) {
					ready = false

					//For debugging we will show the difference
					diff, err := kmp.SafeDiff(desiredDrule.Spec, found.Spec)
					if err != nil {
						log.Error(err, "Failed to diff")
					} else {
						log.Info(fmt.Sprintf("Difference in Destination Rules: %v", diff))
					}
				} else {
					log.Info("The Destination Rules are the same - api server defaults ignored")
				}
			} else {
				log.Info("Found identical Istio Destination Rule", "namespace", found.Namespace, "name", found.Name)

				if instance.Status.ServiceStatus == nil {
					instance.Status.ServiceStatus = map[string]machinelearningv1.ServiceStatus{}
				}

				if _, ok := instance.Status.ServiceStatus[found.Name]; !ok {
					instance.Status.ServiceStatus[found.Name] = *components.serviceDetails[found.Name]
					err = r.Status().Update(context.Background(), instance)
					if err != nil {
						return ready, err
					}
				}
			}
		}

	}

	return ready, nil
}

// Create Services specified in components.
func createServices(r *SeldonDeploymentReconciler, components *components, instance *machinelearningv1.SeldonDeployment, all bool, log logr.Logger) (bool, error) {
	ready := true
	for _, svc := range components.services {
		if !all {
			if _, ok := svc.Annotations[AMBASSADOR_ANNOTATION]; ok {
				log.Info("Skipping Ambassador Svc", "all", all, "namespace", svc.Namespace, "name", svc.Name)
				continue
			}
		}
		if err := ctrl.SetControllerReference(instance, svc, r.Scheme); err != nil {
			return ready, err
		}
		found := &corev1.Service{}
		err := r.Get(context.TODO(), types.NamespacedName{Name: svc.Name, Namespace: svc.Namespace}, found)
		if err != nil && errors.IsNotFound(err) {
			ready = false
			log.Info("Creating Service", "all", all, "namespace", svc.Namespace, "name", svc.Name)
			err = r.Create(context.TODO(), svc)
			if err != nil {
				return ready, err
			}

		} else if err != nil {
			return ready, err
		} else {
			svc.Spec.ClusterIP = found.Spec.ClusterIP
			// Update the found object and write the result back if there are any changes
			if !equality.Semantic.DeepEqual(svc.Spec, found.Spec) || !equality.Semantic.DeepEqual(svc.Annotations, found.Annotations) {
				desiredSvc := found.DeepCopy()
				desiredSvc.Annotations = svc.Annotations
				found.Spec = svc.Spec
				found.Annotations = svc.Annotations
				log.Info("Updating Service", "all", all, "namespace", svc.Namespace, "name", svc.Name)
				err = r.Update(context.TODO(), found)
				if err != nil {
					return ready, err
				}

				// Check if what came back from server modulo the defaults applied by k8s is the same or not
				if !equality.Semantic.DeepEqual(desiredSvc.Spec, found.Spec) {
					ready = false

					//For debugging we will show the difference
					diff, err := kmp.SafeDiff(desiredSvc, found)
					if err != nil {
						log.Error(err, "Failed to diff")
					} else {
						log.Info(fmt.Sprintf("Difference in SVCs: %v", diff))
					}
				} else {
					log.Info("The SVCs are the same - api server defaults ignored")
				}
			} else {
				log.Info("Found identical Service", "all", all, "namespace", found.Namespace, "name", found.Name, "status", found.Status)

				if instance.Status.ServiceStatus == nil {
					instance.Status.ServiceStatus = map[string]machinelearningv1.ServiceStatus{}
				}

				if _, ok := instance.Status.ServiceStatus[found.Name]; !ok {
					instance.Status.ServiceStatus[found.Name] = *components.serviceDetails[found.Name]
					err = r.Status().Update(context.Background(), instance)
					if err != nil {
						return ready, err
					}
				}
			}
		}

	}

	return ready, nil
}

// Create Services specified in components.
func createHpas(r *SeldonDeploymentReconciler, components *components, instance *machinelearningv1.SeldonDeployment, log logr.Logger) (bool, error) {
	ready := true
	for _, hpa := range components.hpas {
		if err := ctrl.SetControllerReference(instance, hpa, r.Scheme); err != nil {
			return ready, err
		}
		found := &autoscaling.HorizontalPodAutoscaler{}
		err := r.Get(context.TODO(), types.NamespacedName{Name: hpa.Name, Namespace: hpa.Namespace}, found)
		if err != nil && errors.IsNotFound(err) {
			ready = false
			log.Info("Creating HPA", "namespace", hpa.Namespace, "name", hpa.Name)
			err = r.Create(context.TODO(), hpa)
			if err != nil {
				return ready, err
			}

		} else if err != nil {
			return ready, err
		} else {
			// Update the found object and write the result back if there are any changes
			if !equality.Semantic.DeepEqual(hpa.Spec, found.Spec) {

				desiredHpa := found.DeepCopy()
				found.Spec = hpa.Spec

				log.Info("Updating HPA", "namespace", hpa.Namespace, "name", hpa.Name)
				err = r.Update(context.TODO(), found)
				if err != nil {
					return ready, err
				}

				// Check if what came back from server modulo the defaults applied by k8s is the same or not
				if !equality.Semantic.DeepEqual(desiredHpa.Spec, found.Spec) {
					ready = false

					//For debugging we will show the difference
					diff, err := kmp.SafeDiff(desiredHpa.Spec, found.Spec)
					if err != nil {
						log.Error(err, "Failed to diff")
					} else {
						log.Info(fmt.Sprintf("Difference in HPAs: %v", diff))
					}
				} else {
					log.Info("The HPAs are the same - api server defaults ignored")
				}

			} else {
				log.Info("Found identical HPA", "namespace", found.Namespace, "name", found.Name, "status", found.Status)
			}
		}

	}
	return ready, nil
}

func jsonEquals(a, b interface{}) (bool, error) {
	b1, err := json.Marshal(a)
	if err != nil {
		return false, err
	}
	b2, err := json.Marshal(b)
	if err != nil {
		return false, err
	}
	return bytes.Equal(b1, b2), nil
}

// Create Deployments specified in components.
func createDeployments(r *SeldonDeploymentReconciler, components *components, instance *machinelearningv1.SeldonDeployment, log logr.Logger) (bool, error) {
	ready := true
	for _, deploy := range components.deployments {

		log.Info("Scheme", "r.scheme", r.Scheme)
		log.Info("createDeployments", "deploy", deploy)
		if err := ctrl.SetControllerReference(instance, deploy, r.Scheme); err != nil {
			return ready, err
		}

		// TODO(user): Change this for the object type created by your controller
		// Check if the Deployment already exists
		found := &appsv1.Deployment{}
		err := r.Get(context.TODO(), types.NamespacedName{Name: deploy.Name, Namespace: deploy.Namespace}, found)
		if err != nil && errors.IsNotFound(err) {
			ready = false
			log.Info("Creating Deployment", "namespace", deploy.Namespace, "name", deploy.Name)

			err = r.Create(context.TODO(), deploy)
			if err != nil {
				return ready, err
			}

		} else if err != nil {
			return ready, err
		} else {
			identical := true
			if !equality.Semantic.DeepEqual(deploy.Spec.Template.Spec, found.Spec.Template.Spec) {
				log.Info("Updating Deployment", "namespace", deploy.Namespace, "name", deploy.Name)

				desiredDeployment := found.DeepCopy()
				found.Spec = deploy.Spec

				err = r.Update(context.TODO(), found)
				if err != nil {
					return ready, err
				}

				// Check if what came back from server modulo the defaults applied by k8s is the same or not
				if !equality.Semantic.DeepEqual(desiredDeployment.Spec.Template.Spec, found.Spec.Template.Spec) {
					ready = false
					identical = false
					//For debugging we will show the difference
					diff, err := kmp.SafeDiff(desiredDeployment.Spec.Template.Spec, found.Spec.Template.Spec)
					if err != nil {
						log.Error(err, "Failed to diff")
					} else {
						log.Info(fmt.Sprintf("Difference in deployments: %v", diff))
					}
				} else {
					log.Info("The deployments are the same - api server defaults ignored")
				}

			}
			if identical {
				log.Info("Found identical deployment", "namespace", found.Namespace, "name", found.Name, "status", found.Status)
				deploymentStatus, present := instance.Status.DeploymentStatus[found.Name]

				if !present {
					deploymentStatus = machinelearningv1.DeploymentStatus{}
				}

				if deploymentStatus.Replicas != found.Status.Replicas || deploymentStatus.AvailableReplicas != found.Status.AvailableReplicas {
					deploymentStatus.Replicas = found.Status.Replicas
					deploymentStatus.AvailableReplicas = found.Status.AvailableReplicas
					if instance.Status.DeploymentStatus == nil {
						instance.Status.DeploymentStatus = map[string]machinelearningv1.DeploymentStatus{}
					}

					instance.Status.DeploymentStatus[found.Name] = deploymentStatus

					err = r.Status().Update(context.Background(), instance)
					if err != nil {
						return ready, err
					}
				}
				log.Info("Deployment status ", "name", found.Name, "status", found.Status)
				if found.Status.ReadyReplicas == 0 || found.Status.UnavailableReplicas > 0 {
					ready = false
				}
			}

		}
	}

	// Add new services
	// Clean up any old deployments and services
	// 1. Create any new services or virtual services
	// 2. Delete any svc-orchestroator deployments
	// 3. Delete any other deployments
	// 4. Delete any old services
	// Deletion is done in foreground so we wait for underlying pods to be removed
	if ready {

		//Create services
		ready, err := createServices(r, components, instance, true, log)
		if err != nil {
			return false, err
		}

		ready, err = createIstioServices(r, components, instance, log)
		if err != nil {
			return false, err
		}

		statusCopy := instance.Status.DeepCopy()
		//delete from copied status the current expected deployments by name
		for _, deploy := range components.deployments {
			delete(statusCopy.DeploymentStatus, deploy.Name)
		}
		for k := range components.serviceDetails {
			delete(statusCopy.ServiceStatus, k)
		}
		remaining := len(statusCopy.DeploymentStatus)
		// Any deployments left in status should be removed as they are not part of the current graph
		svcOrchExists := false
		for k := range statusCopy.DeploymentStatus {
			found := &appsv1.Deployment{}
			err := r.Get(context.TODO(), types.NamespacedName{Name: k, Namespace: instance.Namespace}, found)
			if err != nil && errors.IsNotFound(err) {

			} else {
				if _, ok := found.ObjectMeta.Labels[machinelearningv1.Label_svc_orch]; ok {
					log.Info("Found existing svc-orch")
					svcOrchExists = true
					break
				}
			}
		}
		for k := range statusCopy.DeploymentStatus {
			found := &appsv1.Deployment{}
			err := r.Get(context.TODO(), types.NamespacedName{Name: k, Namespace: instance.Namespace}, found)
			if err != nil && errors.IsNotFound(err) {
				log.Info("Failed to find old deployment - removing from status", "name", k)
				// clean up status
				delete(instance.Status.DeploymentStatus, k)
				err = r.Status().Update(context.Background(), instance)
				if err != nil {
					return ready, err
				}
				return ready, err
			} else {
				if svcOrchExists {
					if _, ok := found.ObjectMeta.Labels[machinelearningv1.Label_svc_orch]; ok {
						log.Info("Deleting old svc-orch deployment ", "name", k)

						err := r.Delete(context.TODO(), found, client.PropagationPolicy(metav1.DeletePropagationForeground))
						if err != nil {
							return ready, err
						}
					}
				} else {
					log.Info("Deleting old deployment (svc-orch does not exist)", "name", k)

					err := r.Delete(context.TODO(), found, client.PropagationPolicy(metav1.DeletePropagationForeground))
					if err != nil {
						return ready, err
					}
				}
			}
		}
		if remaining == 0 {
			log.Info("Removing unused services")
			for k := range statusCopy.ServiceStatus {
				found := &corev1.Service{}
				err := r.Get(context.TODO(), types.NamespacedName{Name: k, Namespace: instance.Namespace}, found)
				if err != nil && errors.IsNotFound(err) {
					log.Error(err, "Failed to find old service", "name", k)
					return ready, err
				} else {
					log.Info("Deleting old service ", "name", k)
					// clean up status
					delete(instance.Status.ServiceStatus, k)
					err = r.Status().Update(context.Background(), instance)
					if err != nil {
						return ready, err
					}
					err := r.Delete(context.TODO(), found)
					if err != nil {
						return ready, err
					}
				}
			}
		}
	}
	return ready, nil
}

// Reconcile reads that state of the cluster for a SeldonDeployment object and makes changes based on the state read
// and what is in the SeldonDeployment.Spec
// Automatically generate RBAC rules to allow the Controller to read and write Deployments

// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=v1,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=v1,resources=services/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=networking.istio.io,resources=virtualservices,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=networking.istio.io,resources=virtualservices/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=networking.istio.io,resources=destinationrules,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=networking.istio.io,resources=destinationrules/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=autoscaling,resources=horizontalpodautoscalers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=autoscaling,resources=horizontalpodautoscalers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=machinelearning.seldon.io,resources=seldondeployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=machinelearning.seldon.io,resources=seldondeployments/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=machinelearning.seldon.io,resources=seldondeployments/finalizers,verbs=get;update;patch
// +kubebuilder:rbac:groups=v1,resources=namespaces,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch

func (r *SeldonDeploymentReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("SeldonDeployment", req.NamespacedName)
	log.Info("Reconcile called")
	// your logic here
	// Fetch the SeldonDeployment instance
	instance := &machinelearningv1.SeldonDeployment{}
	err := r.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "unable to fetch SeldonDeployment")
		return ctrl.Result{}, err
	}

	// Check if we are not namespaced and should ignore this as its in a namespace managed by another operator
	if r.Namespace == "" {
		ns := &corev1.Namespace{}
		err := r.Get(ctx, types.NamespacedName{Name: instance.Namespace}, ns)
		if err != nil {
			log.Error(err, "unable to fetch SeldonDeployment namespace", "namespace", instance.Namespace)
			return ctrl.Result{}, err
		} else {
			if ns.Labels[LABEL_CONTROLLER_ID] != "" {
				log.Info("Skipping reconcile of namespaced deployment")
				return ctrl.Result{}, nil
			}
		}
	}

	// Check we should reconcile this by matching controller-id
	controllerId := GetEnv(ENV_CONTROLLER_ID, "")
	desiredControllerId := instance.Labels[LABEL_CONTROLLER_ID]
	if desiredControllerId != controllerId {
		log.Info("Skipping reconcile of deployment.", "Our controller ID form Env", controllerId, " desired controller ID from label", desiredControllerId)
		return ctrl.Result{}, nil
	}

	components, err := createComponents(r, instance, log)
	if err != nil {
		return ctrl.Result{}, err
	}

	deploymentsReady, err := createDeployments(r, components, instance, log)
	if err != nil {
		return ctrl.Result{}, err
	}

	servicesReady, err := createServices(r, components, instance, false, log)
	if err != nil {
		return ctrl.Result{}, err
	}

	virtualServicesReady, err := createIstioServices(r, components, instance, log)
	if err != nil {
		return ctrl.Result{}, err
	}

	hpasReady, err := createHpas(r, components, instance, log)
	if err != nil {
		return ctrl.Result{}, err
	}

	if deploymentsReady && servicesReady && hpasReady && virtualServicesReady {
		instance.Status.State = "Available"
	} else {
		instance.Status.State = "Creating"
	}
	err = r.Status().Update(context.Background(), instance)
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

var (
	ownerKey = ".metadata.controller"
	apiGVStr = machinelearningv1.GroupVersion.String()
)

func (r *SeldonDeploymentReconciler) SetupWithManager(mgr ctrl.Manager) error {

	if err := mgr.GetFieldIndexer().IndexField(&appsv1.Deployment{}, ownerKey, func(rawObj runtime.Object) []string {
		// grab the deployment object, extract the owner...
		dep := rawObj.(*appsv1.Deployment)
		owner := metav1.GetControllerOf(dep)
		if owner == nil {
			return nil
		}
		// ...make sure it's a SeldonDeployment...
		if owner.APIVersion != apiGVStr || owner.Kind != "SeldonDeployment" {
			return nil
		}

		// ...and if so, return it
		return []string{owner.Name}
	}); err != nil {
		return err
	}

	if err := mgr.GetFieldIndexer().IndexField(&corev1.Service{}, ownerKey, func(rawObj runtime.Object) []string {
		// grab the deployment object, extract the owner...
		svc := rawObj.(*corev1.Service)
		owner := metav1.GetControllerOf(svc)
		if owner == nil {
			return nil
		}
		// ...make sure it's a SeldonDeployment...
		if owner.APIVersion != apiGVStr || owner.Kind != "SeldonDeployment" {
			return nil
		}

		// ...and if so, return it
		return []string{owner.Name}
	}); err != nil {
		return err
	}

	if GetEnv(ENV_ISTIO_ENABLED, "false") == "true" {
		if err := mgr.GetFieldIndexer().IndexField(&istio.VirtualService{}, ownerKey, func(rawObj runtime.Object) []string {
			// grab the deployment object, extract the owner...
			vsvc := rawObj.(*istio.VirtualService)
			owner := metav1.GetControllerOf(vsvc)
			if owner == nil {
				return nil
			}
			// ...make sure it's a SeldonDeployment...
			if owner.APIVersion != apiGVStr || owner.Kind != "SeldonDeployment" {
				return nil
			}

			// ...and if so, return it
			return []string{owner.Name}
		}); err != nil {
			return err
		}
		return ctrl.NewControllerManagedBy(mgr).
			For(&machinelearningv1.SeldonDeployment{}).
			Owns(&appsv1.Deployment{}).
			Owns(&corev1.Service{}).
			Owns(&istio.VirtualService{}).
			Complete(r)
	} else {
		return ctrl.NewControllerManagedBy(mgr).
			For(&machinelearningv1.SeldonDeployment{}).
			Owns(&appsv1.Deployment{}).
			Owns(&corev1.Service{}).
			Complete(r)
	}

}
