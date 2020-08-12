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
	"net/url"
	"strconv"
	"strings"

	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	types2 "github.com/gogo/protobuf/types"
	"github.com/seldonio/seldon-core/operator/constants"
	"github.com/seldonio/seldon-core/operator/utils"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/tools/record"
	"knative.dev/pkg/kmp"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/go-logr/logr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	machinelearningv1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"

	"encoding/json"

	istio_networking "istio.io/api/networking/v1alpha3"
	istio "istio.io/client-go/pkg/apis/networking/v1alpha3"
	appsv1 "k8s.io/api/apps/v1"
	autoscaling "k8s.io/api/autoscaling/v2beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	Recorder  record.EventRecorder
	ClientSet kubernetes.Interface
}

//---------------- Old part

type components struct {
	serviceDetails        map[string]*machinelearningv1.ServiceStatus
	deployments           []*appsv1.Deployment
	services              []*corev1.Service
	hpas                  []*autoscaling.HorizontalPodAutoscaler
	virtualServices       []*istio.VirtualService
	destinationRules      []*istio.DestinationRule
	defaultDeploymentName string
	addressable           *machinelearningv1.SeldonAddressable
}

type httpGrpcPorts struct {
	httpPort int
	grpcPort int
}

func init() {
	// Allow unknown fields in Istio API client.  This is so that we are more resilience
	// in cases user clusers have malformed resources.
	istio_networking.VirtualServiceUnmarshaler.AllowUnknownFields = true
	istio_networking.GatewayUnmarshaler.AllowUnknownFields = true
}

func createAddressableResource(mlDep *machinelearningv1.SeldonDeployment, namespace string) (*machinelearningv1.SeldonAddressable, error) {
	// It was an explicit design decision to expose the service name instead of the ingress
	// Currently there will only be a URL for the first predictor, and assumes always REST
	firstPredictor := &mlDep.Spec.Predictors[0]
	sdepSvcName := machinelearningv1.GetPredictorKey(mlDep, firstPredictor)
	addressablePort, err := getEngineHttpPort()
	if err != nil {
		return nil, err
	}
	addressableHost := sdepSvcName + "." + namespace + ".svc.cluster.local" + ":" + strconv.Itoa(addressablePort)
	addressablePath := utils.GetPredictionPath(mlDep)
	addressableUrl := url.URL{Scheme: "http", Host: addressableHost, Path: addressablePath}

	return &machinelearningv1.SeldonAddressable{URL: addressableUrl.String()}, nil
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
				APIVersion: "apps/v1",
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
	grpcAllowed bool) ([]*istio.VirtualService, []*istio.DestinationRule, error) {

	istio_gateway := utils.GetEnv(ENV_ISTIO_GATEWAY, "seldon-gateway")
	istioTLSMode := utils.GetEnv(ENV_ISTIO_TLS_MODE, "")
	istioRetriesAnnotation := getAnnotation(mlDep, ANNOTATION_ISTIO_RETRIES, "")
	istioRetriesTimeoutAnnotation := getAnnotation(mlDep, ANNOTATION_ISTIO_RETRIES_TIMEOUT, "1")
	istioRetries := 0
	istioRetriesTimeout := 1
	var err error

	if istioRetriesAnnotation != "" {
		istioRetries, err = strconv.Atoi(istioRetriesAnnotation)
		if err != nil {
			return nil, nil, err
		}
		istioRetriesTimeout, err = strconv.Atoi(istioRetriesTimeoutAnnotation)
		if err != nil {
			return nil, nil, err
		}
	}
	httpVsvc := &istio.VirtualService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      seldonId + "-http",
			Namespace: namespace,
		},
		Spec: istio_networking.VirtualService{
			Hosts:    []string{"*"},
			Gateways: []string{getAnnotation(mlDep, ANNOTATION_ISTIO_GATEWAY, istio_gateway)},
			Http: []*istio_networking.HTTPRoute{
				{
					Match: []*istio_networking.HTTPMatchRequest{
						{
							Uri: &istio_networking.StringMatch{MatchType: &istio_networking.StringMatch_Prefix{Prefix: "/seldon/" + namespace + "/" + mlDep.Name + "/"}},
						},
					},
					Rewrite: &istio_networking.HTTPRewrite{Uri: "/"},
				},
			},
		},
	}

	grpcVsvc := &istio.VirtualService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      seldonId + "-grpc",
			Namespace: namespace,
		},
		Spec: istio_networking.VirtualService{
			Hosts:    []string{"*"},
			Gateways: []string{getAnnotation(mlDep, ANNOTATION_ISTIO_GATEWAY, istio_gateway)},
			Http: []*istio_networking.HTTPRoute{
				{
					Match: []*istio_networking.HTTPMatchRequest{
						{
							Uri: &istio_networking.StringMatch{MatchType: &istio_networking.StringMatch_Regex{Regex: constants.GRPCRegExMatchIstio}},
							Headers: map[string]*istio_networking.StringMatch{
								"seldon":    &istio_networking.StringMatch{MatchType: &istio_networking.StringMatch_Exact{Exact: mlDep.Name}},
								"namespace": &istio_networking.StringMatch{MatchType: &istio_networking.StringMatch_Exact{Exact: namespace}},
							},
						},
					},
				},
			},
		},
	}
	// Add retries
	if istioRetries > 0 {
		httpVsvc.Spec.Http[0].Retries = &istio_networking.HTTPRetry{Attempts: int32(istioRetries), PerTryTimeout: &types2.Duration{Seconds: int64(istioRetriesTimeout)}, RetryOn: "gateway-error,connect-failure,refused-stream"}
		grpcVsvc.Spec.Http[0].Retries = &istio_networking.HTTPRetry{Attempts: int32(istioRetries), PerTryTimeout: &types2.Duration{Seconds: int64(istioRetriesTimeout)}, RetryOn: "gateway-error,connect-failure,refused-stream"}
	}

	// shadows don't get destinations in the vs as a shadow is a mirror instead
	var shadows int = 0
	for i := 0; i < len(mlDep.Spec.Predictors); i++ {
		p := mlDep.Spec.Predictors[i]
		if p.Shadow == true {
			shadows += 1
		}
	}

	routesHttp := make([]*istio_networking.HTTPRouteDestination, len(mlDep.Spec.Predictors)-shadows)
	routesGrpc := make([]*istio_networking.HTTPRouteDestination, len(mlDep.Spec.Predictors)-shadows)

	// the shdadow/mirror entry does need a DestinationRule though
	drules := make([]*istio.DestinationRule, len(mlDep.Spec.Predictors))
	routesIdx := 0
	for i := 0; i < len(mlDep.Spec.Predictors); i++ {

		p := mlDep.Spec.Predictors[i]
		pSvcName := machinelearningv1.GetPredictorKey(mlDep, &p)

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
				TrafficPolicy: &istio_networking.TrafficPolicy{ConnectionPool: &istio_networking.ConnectionPoolSettings{Http: &istio_networking.ConnectionPoolSettings_HTTPSettings{IdleTimeout: &types2.Duration{Seconds: 60}}}},
			},
		}

		if istioTLSMode != "" {
			drule.Spec.TrafficPolicy = &istio_networking.TrafficPolicy{
				Tls: &istio_networking.TLSSettings{
					Mode: istio_networking.TLSSettings_TLSmode(istio_networking.TLSSettings_TLSmode_value[istioTLSMode]),
				},
			}
		}
		drules[i] = drule

		if p.Shadow == true {
			//if there's a shadow then add a mirror section to the VirtualService

			httpVsvc.Spec.Http[0].Mirror = &istio_networking.Destination{
				Host:   pSvcName,
				Subset: p.Name,
				Port: &istio_networking.PortSelector{
					Number: uint32(ports[i].httpPort),
				},
			}

			grpcVsvc.Spec.Http[0].Mirror = &istio_networking.Destination{
				Host:   pSvcName,
				Subset: p.Name,
				Port: &istio_networking.PortSelector{
					Number: uint32(ports[i].grpcPort),
				},
			}

			continue
		}

		//we split by adding different routes with their own Weights
		//so not by tag - different destinations (like https://istio.io/docs/tasks/traffic-management/traffic-shifting/) distinguished by host
		routesHttp[routesIdx] = &istio_networking.HTTPRouteDestination{
			Destination: &istio_networking.Destination{
				Host:   pSvcName,
				Subset: p.Name,
				Port: &istio_networking.PortSelector{
					Number: uint32(ports[i].httpPort),
				},
			},
			Weight: p.Traffic,
		}
		routesGrpc[routesIdx] = &istio_networking.HTTPRouteDestination{
			Destination: &istio_networking.Destination{
				Host:   pSvcName,
				Subset: p.Name,
				Port: &istio_networking.PortSelector{
					Number: uint32(ports[i].grpcPort),
				},
			},
			Weight: p.Traffic,
		}
		routesIdx += 1

	}
	httpVsvc.Spec.Http[0].Route = routesHttp
	grpcVsvc.Spec.Http[0].Route = routesGrpc

	if httpAllowed && grpcAllowed {
		vscs := make([]*istio.VirtualService, 2)
		vscs[0] = httpVsvc
		vscs[1] = grpcVsvc
		return vscs, drules, nil
	} else if httpAllowed {
		vscs := make([]*istio.VirtualService, 1)
		vscs[0] = httpVsvc
		return vscs, drules, nil
	} else {
		vscs := make([]*istio.VirtualService, 1)
		vscs[0] = grpcVsvc
		return vscs, drules, nil
	}
}

func getEngineHttpPort() (engine_http_port int, err error) {
	// Get engine http port from environment or use default
	engine_http_port = DEFAULT_ENGINE_CONTAINER_PORT
	var env_engine_http_port = utils.GetEnv(ENV_DEFAULT_ENGINE_SERVER_PORT, "")
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
	var env_engine_grpc_port = utils.GetEnv(ENV_DEFAULT_ENGINE_SERVER_GRPC_PORT, "")
	if env_engine_grpc_port != "" {
		engine_grpc_port, err = strconv.Atoi(env_engine_grpc_port)
		if err != nil {
			return 0, err
		}
	}
	return engine_grpc_port, nil
}

// Create all the components (Deployments, Services etc)
func (r *SeldonDeploymentReconciler) createComponents(mlDep *machinelearningv1.SeldonDeployment, securityContext *corev1.PodSecurityContext, log logr.Logger) (*components, error) {
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
		noEngine := strings.ToLower(p.Annotations[machinelearningv1.ANNOTATION_NO_ENGINE]) == "true"
		if noEngine && len(p.ComponentSpecs) > 0 && len(p.ComponentSpecs[0].Spec.Containers) > 0 {
			pu := machinelearningv1.GetPredictiveUnit(&p.Graph, p.ComponentSpecs[0].Spec.Containers[0].Name)
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
		noEngine := strings.ToLower(p.Annotations[machinelearningv1.ANNOTATION_NO_ENGINE]) == "true"
		pSvcName := machinelearningv1.GetPredictorKey(mlDep, &p)
		log.Info("pSvcName", "val", pSvcName)
		// Add engine deployment if separate
		hasSeparateEnginePod := strings.ToLower(mlDep.Spec.Annotations[machinelearningv1.ANNOTATION_SEPARATE_ENGINE]) == "true"
		if hasSeparateEnginePod && !noEngine {
			deploy, err := createEngineDeployment(mlDep, &p, pSvcName, engine_http_port, engine_grpc_port)
			if err != nil {
				return nil, err
			}
			if securityContext != nil {
				deploy.Spec.Template.Spec.SecurityContext = securityContext
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
			depName := machinelearningv1.GetDeploymentName(mlDep, p, cSpec, j)
			if i == 0 && j == 0 {
				c.defaultDeploymentName = depName
			}
			deploy := createDeploymentWithoutEngine(depName, seldonId, cSpec, &p, mlDep, securityContext)
			// Add HPA if needed
			if cSpec.HpaSpec != nil {
				c.hpas = append(c.hpas, createHpa(cSpec, depName, seldonId, namespace))
			} else { //set replicas from more specifc to more general replicas settings in spec
				if cSpec.Replicas != nil {
					deploy.Spec.Replicas = cSpec.Replicas
				} else if p.Replicas != nil {
					deploy.Spec.Replicas = p.Replicas
				} else {
					deploy.Spec.Replicas = mlDep.Spec.Replicas
				}
			}

			// create services for each container
			for k := 0; k < len(cSpec.Spec.Containers); k++ {
				var con *corev1.Container
				// get the container on the created deployment, as createDeploymentWithoutEngine will have created as a copy of the spec in the manifest and added defaults to it
				// we need the reference as we may have to modify the container when creating the Service (e.g. to add probes)
				con = utils.GetContainerForDeployment(deploy, cSpec.Spec.Containers[k].Name)
				pu := machinelearningv1.GetPredictiveUnit(&p.Graph, con.Name)
				deploy = addLabelsToDeployment(deploy, pu, &p)

				// engine will later get a special predictor service as it is entrypoint for graph
				// and no need to expose tfserving container as it's accessed via proxy
				if con.Name != EngineContainerName && con.Name != constants.TFServingContainerName {

					// service for hitting a model directly, not via engine - also adds ports to container if needed
					svc := createContainerService(deploy, p, mlDep, con, c, seldonId)
					if svc != nil {
						svc = addLabelsToService(svc, pu, &p)
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
							psvc, err := createPredictorService(pSvcName, seldonId, &p, mlDep, 0, port, false, log)
							if err != nil {
								return nil, err
							}
							psvc = addLabelsToService(psvc, pu, &p)

							c.services = append(c.services, psvc)

							c.serviceDetails[pSvcName] = &machinelearningv1.ServiceStatus{
								SvcName:      pSvcName,
								GrpcEndpoint: pSvcName + "." + namespace + ":" + strconv.Itoa(port),
							}
						} else {
							externalPorts[i] = httpGrpcPorts{httpPort: port, grpcPort: 0}
							grpcAllowed = false
							psvc, err := createPredictorService(pSvcName, seldonId, &p, mlDep, port, 0, false, log)
							if err != nil {
								return nil, err
							}
							psvc = addLabelsToService(psvc, pu, &p)

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

		pi := NewPrePackedInitializer(r.ClientSet)
		err = pi.createStandaloneModelServers(mlDep, &p, &c, &p.Graph, securityContext)
		if err != nil {
			return nil, err
		}

		if !noEngine {

			// Add service orchestrator to engine deployment if needed
			if !hasSeparateEnginePod {
				var deploy *appsv1.Deployment
				found := false

				// find the pu that the webhook marked as localhost as its corresponding deployment should get the engine
				pu := machinelearningv1.GetEnginePredictiveUnit(&p.Graph)
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
			psvc, err := createPredictorService(pSvcName, seldonId, &p, mlDep, httpPort, grpcPort, false, log)
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

		ei := NewExplainerInitializer(r.ClientSet)
		err = ei.createExplainer(mlDep, &p, &c, pSvcName, securityContext, log)
		if err != nil {
			return nil, err
		}
	}

	// Create the addressable as all services are created when SeldonDeployment is ready
	c.addressable, err = createAddressableResource(mlDep, namespace)
	if err != nil {
		return nil, err
	}

	//TODO Fixme - not changed to handle per predictor scenario
	if utils.GetEnv(ENV_ISTIO_ENABLED, "false") == "true" {
		vsvcs, dstRule, err := createIstioResources(mlDep, seldonId, namespace, externalPorts, httpAllowed, grpcAllowed)
		if err != nil {
			return nil, err
		}
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
	isExplainer bool,
	log logr.Logger) (pSvc *corev1.Service, err error) {
	namespace := getNamespace(mlDep)
	psvc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pSvcName,
			Namespace: namespace,
			Labels: map[string]string{
				machinelearningv1.Label_seldon_app: pSvcName,
				machinelearningv1.Label_seldon_id:  seldonId,
				machinelearningv1.Label_managed_by: machinelearningv1.Label_value_seldon,
			},
		},
		Spec: corev1.ServiceSpec{
			Selector:        map[string]string{machinelearningv1.Label_seldon_app: pSvcName},
			SessionAffinity: corev1.ServiceAffinityNone,
			Type:            corev1.ServiceTypeClusterIP,
		},
	}
	if isExecutorEnabled(mlDep) {
		if engine_http_port != 0 && len(psvc.Spec.Ports) == 0 {
			psvc.Spec.Ports = append(psvc.Spec.Ports, corev1.ServicePort{Protocol: corev1.ProtocolTCP, Port: int32(engine_http_port), TargetPort: intstr.FromInt(engine_http_port), Name: "http"})
		}

		if engine_grpc_port != 0 && len(psvc.Spec.Ports) < 2 {
			psvc.Spec.Ports = append(psvc.Spec.Ports, corev1.ServicePort{Protocol: corev1.ProtocolTCP, Port: int32(engine_grpc_port), TargetPort: intstr.FromInt(engine_http_port), Name: "http2"})
		}
	} else {
		if engine_http_port != 0 && len(psvc.Spec.Ports) == 0 {
			psvc.Spec.Ports = append(psvc.Spec.Ports, corev1.ServicePort{Protocol: corev1.ProtocolTCP, Port: int32(engine_http_port), TargetPort: intstr.FromInt(engine_http_port), Name: "http"})
		}

		if engine_grpc_port != 0 && len(psvc.Spec.Ports) < 2 {
			psvc.Spec.Ports = append(psvc.Spec.Ports, corev1.ServicePort{Protocol: corev1.ProtocolTCP, Port: int32(engine_grpc_port), TargetPort: intstr.FromInt(engine_grpc_port), Name: "http2"})
		}
	}

	if utils.GetEnv("AMBASSADOR_ENABLED", "false") == "true" {
		psvc.Annotations = make(map[string]string)
		//Create top level Service
		ambassadorConfig, err := getAmbassadorConfigs(mlDep, p, pSvcName, engine_http_port, engine_grpc_port, isExplainer)
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
func createContainerService(deploy *appsv1.Deployment,
	p machinelearningv1.PredictorSpec,
	mlDep *machinelearningv1.SeldonDeployment,
	con *corev1.Container,
	c components,
	seldonId string) *corev1.Service {
	containerServiceKey := machinelearningv1.Label_seldon_app_svc
	containerServiceValue := machinelearningv1.GetContainerServiceName(mlDep.Name, p, con)
	pu := machinelearningv1.GetPredictiveUnit(&p.Graph, con.Name)

	// only create services for containers defined as pus in the graph
	if pu == nil {
		return nil
	}
	namespace := getNamespace(mlDep)
	portType := "http"
	if pu.Endpoint.Type == machinelearningv1.GRPC {
		portType = "grpc"
	}
	var portNum int32
	portNum = 0
	existingPort := machinelearningv1.GetPort(portType, con.Ports)
	if existingPort != nil {
		portNum = existingPort.ContainerPort
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
			Labels:    map[string]string{containerServiceKey: containerServiceValue, machinelearningv1.Label_seldon_id: seldonId},
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
		con.Env = append(con.Env, corev1.EnvVar{Name: machinelearningv1.ENV_PREDICTIVE_UNIT_SERVICE_PORT, Value: strconv.Itoa(int(portNum))})
	}

	if pu != nil && len(pu.Parameters) > 0 {
		if !utils.HasEnvVar(con.Env, machinelearningv1.ENV_PREDICTIVE_UNIT_PARAMETERS) {
			con.Env = append(con.Env, corev1.EnvVar{Name: machinelearningv1.ENV_PREDICTIVE_UNIT_PARAMETERS, Value: utils.GetPredictiveUnitAsJson(pu.Parameters)})
		}
	}

	// Always set the predictive and deployment identifiers

	labels, err := json.Marshal(p.Labels)
	if err != nil {
		labels = []byte("{}")
	}

	con.Env = append(con.Env, []corev1.EnvVar{
		corev1.EnvVar{Name: machinelearningv1.ENV_PREDICTIVE_UNIT_ID, Value: con.Name},
		corev1.EnvVar{Name: machinelearningv1.ENV_PREDICTIVE_UNIT_IMAGE, Value: con.Image},
		corev1.EnvVar{Name: machinelearningv1.ENV_PREDICTOR_ID, Value: p.Name},
		corev1.EnvVar{Name: machinelearningv1.ENV_PREDICTOR_LABELS, Value: string(labels)},
		corev1.EnvVar{Name: machinelearningv1.ENV_SELDON_DEPLOYMENT_ID, Value: mlDep.ObjectMeta.Name},
		corev1.EnvVar{Name: machinelearningv1.ENV_SELDON_EXECUTOR_ENABLED, Value: strconv.FormatBool(isExecutorEnabled(mlDep))},
	}...)

	//Add Metric Env Var
	predictiveUnitMetricsPortName := utils.GetEnv(machinelearningv1.ENV_PREDICTIVE_UNIT_METRICS_PORT_NAME, constants.DefaultMetricsPortName)
	metricPort := getPort(predictiveUnitMetricsPortName, con.Ports)
	if metricPort != nil {
		con.Env = append(con.Env, []corev1.EnvVar{
			corev1.EnvVar{Name: machinelearningv1.ENV_PREDICTIVE_UNIT_SERVICE_PORT_METRICS, Value: strconv.Itoa(int(metricPort.ContainerPort))},
			corev1.EnvVar{Name: machinelearningv1.ENV_PREDICTIVE_UNIT_METRICS_ENDPOINT, Value: getPrometheusPath(mlDep)},
		}...)
	}

	return svc
}

func createDeploymentWithoutEngine(depName string, seldonId string, seldonPodSpec *machinelearningv1.SeldonPodSpec, p *machinelearningv1.PredictorSpec, mlDep *machinelearningv1.SeldonDeployment, podSecurityContext *corev1.PodSecurityContext) *appsv1.Deployment {
	deploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      depName,
			Namespace: getNamespace(mlDep),
			Labels: map[string]string{
				machinelearningv1.Label_seldon_id: seldonId,
				"app":                             depName,
				"fluentd":                         "true",
			},
			Annotations: map[string]string{},
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{machinelearningv1.Label_seldon_id: seldonId},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						machinelearningv1.Label_seldon_id: seldonId,
						machinelearningv1.Label_app:       depName,
						machinelearningv1.Label_fluentd:   "true",
					},
					Annotations: map[string]string{},
				},
			},
			Strategy: appsv1.DeploymentStrategy{RollingUpdate: &appsv1.RollingUpdateDeployment{MaxUnavailable: &intstr.IntOrString{StrVal: "10%"}}},
		},
	}

	if deploy.Spec.Template.Annotations == nil {
		deploy.Spec.Template.Annotations = map[string]string{}
	}
	// Add prometheus annotations
	deploy.Spec.Template.Annotations["prometheus.io/path"] = getPrometheusPath(mlDep)
	deploy.Spec.Template.Annotations["prometheus.io/scrape"] = "true"

	if p.Shadow == true {
		deploy.Spec.Template.ObjectMeta.Labels[machinelearningv1.Label_shadow] = "true"
	}

	//Add annotations from top level
	for k, v := range mlDep.Spec.Annotations {
		deploy.Annotations[k] = v
		deploy.Spec.Template.ObjectMeta.Annotations[k] = v
	}
	// Add annottaions from predictor
	for k, v := range p.Annotations {
		deploy.Annotations[k] = v
		deploy.Spec.Template.ObjectMeta.Annotations[k] = v
	}
	if seldonPodSpec != nil {
		deploy.Spec.Template.Spec = seldonPodSpec.Spec
		// add more annotations from metadata
		for k, v := range seldonPodSpec.Metadata.Annotations {
			deploy.Annotations[k] = v
			deploy.Spec.Template.ObjectMeta.Annotations[k] = v
		}
	}

	// Add Pod Security Context
	deploy.Spec.Template.Spec.SecurityContext = podSecurityContext

	// add predictor labels
	for k, v := range p.Labels {
		deploy.ObjectMeta.Labels[k] = v
		deploy.Spec.Template.ObjectMeta.Labels[k] = v
	}
	// add labels from podSpec metadata
	if seldonPodSpec != nil {
		for k, v := range seldonPodSpec.Metadata.Labels {
			deploy.ObjectMeta.Labels[k] = v
			deploy.Spec.Template.ObjectMeta.Labels[k] = v
		}
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
func (r *SeldonDeploymentReconciler) createIstioServices(components *components, instance *machinelearningv1.SeldonDeployment, log logr.Logger) (bool, error) {
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
			r.Recorder.Eventf(instance, corev1.EventTypeNormal, constants.EventsCreateVirtualService, "Created VirtualService %q", svc.GetName())
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
					r.Recorder.Eventf(instance, corev1.EventTypeNormal, constants.EventsUpdateVirtualService, "Updated VirtualService %q", svc.GetName())
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
			r.Recorder.Eventf(instance, corev1.EventTypeNormal, constants.EventsCreateDestinationRule, "Created DestinationRule %q", drule.GetName())
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
					r.Recorder.Eventf(instance, corev1.EventTypeNormal, constants.EventsUpdateDestinationRule, "Updated DestinationRule %q", drule.GetName())
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
			}
		}

	}

	//Cleanup unused VirtualService. This should usually only happen on Operator upgrades where there is a breaking change to the names of the VirtualServices created
	//Only run if we have virtualservices to create - implies we are running with istio active
	if len(components.virtualServices) > 0 && ready {
		cleaner := ResourceCleaner{instance: instance, client: r, virtualServices: components.virtualServices, logger: r.Log}
		deleted, err := cleaner.cleanUnusedVirtualServices()
		if err != nil {
			return ready, err
		}
		for _, vsvcDeleted := range deleted {
			r.Recorder.Eventf(instance, corev1.EventTypeNormal, constants.EventsDeleteVirtualService, "Delete VirtualService %q", vsvcDeleted.GetName())
		}
	}

	return ready, nil
}

// Create Services specified in components.
func (r *SeldonDeploymentReconciler) createServices(components *components, instance *machinelearningv1.SeldonDeployment, all bool, log logr.Logger) (bool, error) {
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
			r.Recorder.Eventf(instance, corev1.EventTypeNormal, constants.EventsCreateService, "Created Service %q", svc.GetName())
		} else if err != nil {
			return ready, err
		} else {
			svc.Spec.ClusterIP = found.Spec.ClusterIP
			// Configure addressable status so it can be reached through duck-typing
			instance.Status.Address = components.addressable
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
					r.Recorder.Eventf(instance, corev1.EventTypeNormal, constants.EventsUpdateService, "Updated Service %q", svc.GetName())
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
				}
			}
		}

	}

	return ready, nil
}

// Create Services specified in components.
func (r *SeldonDeploymentReconciler) createHpas(components *components, instance *machinelearningv1.SeldonDeployment, log logr.Logger) (bool, error) {
	ready := true
	hpaSet := make(map[string]bool)
	for _, hpa := range components.hpas {
		if err := ctrl.SetControllerReference(instance, hpa, r.Scheme); err != nil {
			return ready, err
		}
		hpaSet[hpa.Name] = true
		found := &autoscaling.HorizontalPodAutoscaler{}
		err := r.Get(context.TODO(), types.NamespacedName{Name: hpa.Name, Namespace: hpa.Namespace}, found)
		if err != nil && errors.IsNotFound(err) {
			ready = false
			log.Info("Creating HPA", "namespace", hpa.Namespace, "name", hpa.Name)
			err = r.Create(context.TODO(), hpa)
			if err != nil {
				return ready, err
			}
			r.Recorder.Eventf(instance, corev1.EventTypeNormal, constants.EventsCreateHPA, "Created HorizontalPodAutoscaler %q", hpa.GetName())
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
					r.Recorder.Eventf(instance, corev1.EventTypeNormal, constants.EventsUpdateHPA, "Updated HorizontalPodAutoscaler %q", hpa.GetName())
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

	// For all Deployments check if any Hpas exist and they are not required
	for _, deploy := range components.deployments {
		if _, ok := hpaSet[deploy.Name]; !ok {
			found := &autoscaling.HorizontalPodAutoscaler{}
			err := r.Get(context.TODO(), types.NamespacedName{Name: deploy.Name, Namespace: deploy.Namespace}, found)
			if err != nil {
				if !errors.IsNotFound(err) {
					return false, err
				}
				// Do nothing
			} else {
				// Delete HPA
				log.Info("Deleting hpa", "name", deploy.Name)
				err := r.Delete(context.TODO(), found, client.PropagationPolicy(metav1.DeletePropagationForeground))
				if err != nil {
					return ready, err
				}
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
func (r *SeldonDeploymentReconciler) createDeployments(components *components, instance *machinelearningv1.SeldonDeployment, log logr.Logger) (bool, error) {
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
			r.Recorder.Eventf(instance, corev1.EventTypeNormal, constants.EventsCreateDeployment, "Created Deployment %q", deploy.GetName())
		} else if err != nil {
			return ready, err
		} else {
			identical := true
			if !equality.Semantic.DeepEqual(deploy.Spec.Template.Spec, found.Spec.Template.Spec) {
				log.Info("Updating Deployment", "namespace", deploy.Namespace, "name", deploy.Name)

				desiredDeployment := found.DeepCopy()
				found.Spec = deploy.Spec

				if deploy.Spec.Replicas == nil {
					found.Spec.Replicas = desiredDeployment.Spec.Replicas
				}

				err = r.Update(context.TODO(), found)
				if err != nil {
					return ready, err
				}

				// Check if what came back from server modulo the defaults applied by k8s is the same or not
				if !equality.Semantic.DeepEqual(desiredDeployment.Spec.Template.Spec, found.Spec.Template.Spec) {
					ready = false
					identical = false
					r.Recorder.Eventf(instance, corev1.EventTypeNormal, constants.EventsUpdateDeployment, "Updated Deployment %q", deploy.GetName())
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
					if found.Name == components.defaultDeploymentName {
						instance.Status.Replicas = found.Status.Replicas
					}
				}
				log.Info("Deployment status ", "name", found.Name, "status", found.Status)
				if found.Status.ReadyReplicas == 0 || found.Status.UnavailableReplicas > 0 {
					ready = false
				}
			}

		}
	}
	return ready, nil
}

func (r *SeldonDeploymentReconciler) completeServiceCreation(instance *machinelearningv1.SeldonDeployment, components *components, log logr.Logger) error {
	//Create services
	_, err := r.createServices(components, instance, true, log)
	if err != nil {
		return err
	}

	_, err = r.createIstioServices(components, instance, log)
	if err != nil {
		return err
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
		} else {
			if svcOrchExists {
				if _, ok := found.ObjectMeta.Labels[machinelearningv1.Label_svc_orch]; ok {
					log.Info("Deleting old svc-orch deployment ", "name", k)

					err := r.Delete(context.TODO(), found, client.PropagationPolicy(metav1.DeletePropagationForeground))
					if err != nil {
						return err
					}
					r.Recorder.Eventf(instance, corev1.EventTypeNormal, constants.EventsDeleteDeployment, "Deleted Deployment %q", found.GetName())
				}
			} else {
				log.Info("Deleting old deployment (svc-orch does not exist)", "name", k)

				err := r.Delete(context.TODO(), found, client.PropagationPolicy(metav1.DeletePropagationForeground))
				if err != nil {
					return err
				}
				r.Recorder.Eventf(instance, corev1.EventTypeNormal, constants.EventsDeleteDeployment, "Deleted Deployment %q", found.GetName())
			}

			// Delete any dangling HPAs
			foundHpa := &autoscaling.HorizontalPodAutoscaler{}
			err := r.Get(context.TODO(), types.NamespacedName{Name: found.Name, Namespace: found.Namespace}, foundHpa)
			if err != nil {
				if !errors.IsNotFound(err) {
					return err
				}
				// Do nothing
			} else {
				// Delete HPA that should not exist
				log.Info("Deleting hpa for removed predictor", "name", foundHpa.Name)
				err := r.Delete(context.TODO(), foundHpa, client.PropagationPolicy(metav1.DeletePropagationForeground))
				if err != nil {
					return err
				}
				r.Recorder.Eventf(instance, corev1.EventTypeNormal, constants.EventsDeleteHPA, "Deleted HorizontalPodAutoscaler %q", foundHpa.GetName())
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
				return err
			} else {
				log.Info("Deleting old service ", "name", k)
				// clean up status
				delete(instance.Status.ServiceStatus, k)
				err := r.Delete(context.TODO(), found)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
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
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch

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
	controllerId := utils.GetEnv(ENV_CONTROLLER_ID, "")
	desiredControllerId := instance.Labels[LABEL_CONTROLLER_ID]
	if desiredControllerId != controllerId {
		log.Info("Skipping reconcile of deployment.", "Our controller ID form Env", controllerId, " desired controller ID from label", desiredControllerId)
		return ctrl.Result{}, nil
	}

	//Get Security Context
	podSecurityContext, err := createSecurityContext(instance)

	//rerun defaulting
	before := instance.DeepCopy()
	instance.Default()
	if !equality.Semantic.DeepEqual(instance.Spec, before.Spec) {
		diff, err := kmp.SafeDiff(instance.Spec, before.Spec)
		if err != nil {
			log.Error(err, "Failed to diff")
		} else {
			log.Info(fmt.Sprintf("Difference in seldon deployments: %v", diff))
		}
	}

	components, err := r.createComponents(instance, podSecurityContext, log)
	if err != nil {
		r.Recorder.Eventf(instance, corev1.EventTypeWarning, constants.EventsInternalError, err.Error())
		r.updateStatusForError(instance, err, log)
		return ctrl.Result{}, err
	}

	servicesReady, err := r.createServices(components, instance, false, log)
	if err != nil {
		r.Recorder.Eventf(instance, corev1.EventTypeWarning, constants.EventsInternalError, err.Error())
		r.updateStatusForError(instance, err, log)
		return ctrl.Result{}, err
	}

	hpasReady, err := r.createHpas(components, instance, log)
	if err != nil {
		r.Recorder.Eventf(instance, corev1.EventTypeWarning, constants.EventsInternalError, err.Error())
		r.updateStatusForError(instance, err, log)
		return ctrl.Result{}, err
	}

	deploymentsReady, err := r.createDeployments(components, instance, log)
	if err != nil {
		r.Recorder.Eventf(instance, corev1.EventTypeWarning, constants.EventsInternalError, err.Error())
		r.updateStatusForError(instance, err, log)
		return ctrl.Result{}, err
	}

	if deploymentsReady {
		err := r.completeServiceCreation(instance, components, log)
		if err != nil {
			r.Recorder.Eventf(instance, corev1.EventTypeWarning, constants.EventsInternalError, err.Error())
			r.updateStatusForError(instance, err, log)
			return ctrl.Result{}, err
		}
	}

	if deploymentsReady && servicesReady && hpasReady {
		instance.Status.State = machinelearningv1.StatusStateAvailable
		instance.Status.Description = ""
	} else {
		instance.Status.State = machinelearningv1.StatusStateCreating
		instance.Status.Description = ""
	}
	err = r.updateStatus(instance, log)
	if err != nil {
		return ctrl.Result{}, err
	}

	r.Recorder.Eventf(instance, corev1.EventTypeNormal, constants.EventsUpdated, "Updated SeldonDeployment %q", instance.GetName())
	return ctrl.Result{}, nil
}

func (r *SeldonDeploymentReconciler) updateStatusForError(desired *machinelearningv1.SeldonDeployment, err error, log logr.Logger) {
	desired.Status.State = machinelearningv1.StatusStateFailed
	desired.Status.Description = err.Error()

	existing := &machinelearningv1.SeldonDeployment{}
	namespacedName := types.NamespacedName{Name: desired.Name, Namespace: desired.Namespace}
	if err := r.Get(context.TODO(), namespacedName, existing); err != nil {
		log.Error(err, "Failed to get SeldonDeployment")
		return
	}
	if equality.Semantic.DeepEqual(existing.Status, desired.Status) {
		//Do nothing
	} else if err := r.Status().Update(context.Background(), desired); err != nil {
		log.Error(err, "Failed to update InferenceService status")
		r.Recorder.Eventf(desired, corev1.EventTypeWarning, constants.EventsUpdateFailed,
			"Failed to update status for SeldonDeployment %q: %v", desired.Name, err)
	}
}

func (r *SeldonDeploymentReconciler) updateStatus(desired *machinelearningv1.SeldonDeployment, log logr.Logger) error {
	existing := &machinelearningv1.SeldonDeployment{}
	namespacedName := types.NamespacedName{Name: desired.Name, Namespace: desired.Namespace}
	if err := r.Get(context.TODO(), namespacedName, existing); err != nil {
		return err
	}
	if equality.Semantic.DeepEqual(existing.Status, desired.Status) {
		//Do nothing
	} else if err := r.Status().Update(context.Background(), desired); err != nil {
		log.Error(err, "Failed to update InferenceService status")
		r.Recorder.Eventf(desired, corev1.EventTypeWarning, constants.EventsUpdateFailed,
			"Failed to update status for SeldonDeployment %q: %v", desired.Name, err)
		return err
	}
	return nil
}

var (
	ownerKey = ".metadata.controller"
	apiGVStr = machinelearningv1.GroupVersion.String()
)

func (r *SeldonDeploymentReconciler) SetupWithManager(mgr ctrl.Manager, name string) error {

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

	if utils.GetEnv(ENV_ISTIO_ENABLED, "false") == "true" {
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
			Named(name).
			For(&machinelearningv1.SeldonDeployment{}).
			Owns(&appsv1.Deployment{}).
			Owns(&corev1.Service{}).
			Owns(&istio.VirtualService{}).
			Complete(r)
	} else {
		return ctrl.NewControllerManagedBy(mgr).
			Named(name).
			For(&machinelearningv1.SeldonDeployment{}).
			Owns(&appsv1.Deployment{}).
			Owns(&corev1.Service{}).
			Complete(r)
	}

}
