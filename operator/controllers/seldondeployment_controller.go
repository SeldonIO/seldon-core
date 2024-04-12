/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package controllers

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/banzaicloud/k8s-objectmatcher/patch"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"

	v2 "github.com/emissary-ingress/emissary/v3/pkg/api/getambassador.io/v2"
	"github.com/seldonio/seldon-core/operator/controllers/ambassador"
	utils2 "github.com/seldonio/seldon-core/operator/controllers/utils"

	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"knative.dev/pkg/apis"

	duration "github.com/golang/protobuf/ptypes/duration"
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

	"encoding/json"

	kedav1alpha1 "github.com/kedacore/keda/v2/apis/keda/v1alpha1"
	machinelearningv1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"

	istio_networking "istio.io/api/networking/v1alpha3"
	istio "istio.io/client-go/pkg/apis/networking/v1alpha3"
	appsv1 "k8s.io/api/apps/v1"
	autoscaling "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	policy "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	LastAppliedConfig                   = "seldon.io/last-applied"
	ENV_DEFAULT_ENGINE_SERVER_PORT      = "ENGINE_SERVER_PORT"
	ENV_DEFAULT_ENGINE_SERVER_GRPC_PORT = "ENGINE_SERVER_GRPC_PORT"
	ENV_CONTROLLER_ID                   = "CONTROLLER_ID"

	ENV_PREDICTIVE_UNIT_DEFAULT_ENV_SECRET_REF_NAME = "PREDICTIVE_UNIT_DEFAULT_ENV_SECRET_REF_NAME"

	// This env var in the operator allows you to change the default path
	// 		to mount the cert in the containers
	ENV_DEFAULT_CERT_MOUNT_PATH_NAME = "DEFAULT_CERT_MOUNT_PATH_NAME"
	// The ENV VAR NAME for containers to be able to find the path
	SELDON_MOUNT_PATH_ENV_NAME = "SELDON_CERT_MOUNT_PATH"

	DEFAULT_ENGINE_CONTAINER_PORT = 8000
	DEFAULT_ENGINE_GRPC_PORT      = 5001

	AMBASSADOR_ANNOTATION = "getambassador.io/config"
	LABEL_CONTROLLER_ID   = "seldon.io/controller-id"

	ENV_KEDA_ENABLED = "KEDA_ENABLED"
)

var (
	envDefaultCertMountPath               = utils.GetEnv(ENV_DEFAULT_CERT_MOUNT_PATH_NAME, "/cert/")
	PredictiveUnitDefaultEnvSecretRefName = utils.GetEnv(ENV_PREDICTIVE_UNIT_DEFAULT_ENV_SECRET_REF_NAME, "")
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
	kedaScaledObjects     []*kedav1alpha1.ScaledObject
	pdbs                  []*policy.PodDisruptionBudget
	virtualServices       []*istio.VirtualService
	destinationRules      []*istio.DestinationRule
	mappings              []*v2.Mapping
	tlsContexts           []*v2.TLSContext
	defaultDeploymentName string
	addressable           *machinelearningv1.SeldonAddressable
}

func (c *components) deploymentHasHpa(depname string) bool {
	for _, hpa := range c.hpas {
		if hpa.Name == depname {
			return true
		}
	}
	return false
}

func (c *components) deploymentHasKeda(depname string) bool {
	for _, scaledObject := range c.kedaScaledObjects {
		if scaledObject.Name == depname {
			return true
		}
	}
	return false
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

func createFqdn(svcName string, namespace string) string {
	return svcName + "." + namespace + constants.DNSClusterLocalSuffix
}

func createAddressableResource(mlDep *machinelearningv1.SeldonDeployment, namespace string, externalPorts []httpGrpcPorts) (*machinelearningv1.SeldonAddressable, error) {
	// It was an explicit design decision to expose the service name instead of the ingress
	// Currently there will only be a URL for the first predictor, and assumes always REST
	firstPredictor := &mlDep.Spec.Predictors[0]
	sdepSvcName := machinelearningv1.GetPredictorKey(mlDep, firstPredictor)
	addressableHost := createFqdn(sdepSvcName, namespace) + ":" + strconv.Itoa(externalPorts[0].httpPort)
	addressablePath := utils.GetPredictionPath(mlDep)
	addressableUrl := url.URL{Scheme: "http", Host: addressableHost, Path: addressablePath}

	return &machinelearningv1.SeldonAddressable{URL: addressableUrl.String()}, nil
}

func createKeda(podSpec *machinelearningv1.SeldonPodSpec, deploymentName string, seldonId string, namespace string) *kedav1alpha1.ScaledObject {
	kedaScaledObj := &kedav1alpha1.ScaledObject{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deploymentName,
			Namespace: namespace,
			Labels:    map[string]string{machinelearningv1.Label_seldon_id: seldonId},
		},
		Spec: kedav1alpha1.ScaledObjectSpec{
			PollingInterval:  podSpec.KedaSpec.PollingInterval,
			CooldownPeriod:   podSpec.KedaSpec.CooldownPeriod,
			IdleReplicaCount: podSpec.KedaSpec.IdleReplicaCount,
			MaxReplicaCount:  podSpec.KedaSpec.MaxReplicaCount,
			MinReplicaCount:  podSpec.KedaSpec.MinReplicaCount,
			Advanced:         podSpec.KedaSpec.Advanced,
			Triggers:         podSpec.KedaSpec.Triggers,
			ScaleTargetRef: &kedav1alpha1.ScaleTarget{
				APIVersion: "apps/v1",
				Kind:       "Deployment",
				Name:       deploymentName,
			},
			Fallback: podSpec.KedaSpec.Fallback,
		},
	}
	return kedaScaledObj
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
			Metrics:     podSpec.HpaSpec.Metricsv2,
		},
	}
	if podSpec.HpaSpec.Metrics != nil {
		hpa.Spec.Metrics = machinelearningv1.ConvertMetricSpecSlice(podSpec.HpaSpec.Metrics)
	}
	if podSpec.HpaSpec.MinReplicas != nil {
		hpa.Spec.MinReplicas = podSpec.HpaSpec.MinReplicas
	}

	return &hpa
}

func createPdb(podSpec *machinelearningv1.SeldonPodSpec, deploymentName string, seldonId string, namespace string) *policy.PodDisruptionBudget {
	pdb := policy.PodDisruptionBudget{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deploymentName,
			Namespace: namespace,
			Labels:    map[string]string{machinelearningv1.Label_seldon_id: seldonId},
		},
		Spec: policy.PodDisruptionBudgetSpec{
			MinAvailable:   podSpec.PdbSpec.MinAvailable,
			MaxUnavailable: podSpec.PdbSpec.MaxUnavailable,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{machinelearningv1.Label_seldon_id: seldonId},
			},
		},
	}
	return &pdb
}

// Create istio virtual service and destination rule.
// Creates routes for each predictor with traffic weight split
func createIstioResources(mlDep *machinelearningv1.SeldonDeployment,
	seldonId string,
	namespace string,
	ports []httpGrpcPorts) ([]*istio.VirtualService, []*istio.DestinationRule, error) {

	istio_gateway := utils.GetEnv(ENV_ISTIO_GATEWAY, "seldon-gateway")
	istioTLSMode := utils.GetEnv(ENV_ISTIO_TLS_MODE, "")
	istioRetriesAnnotation := utils2.GetAnnotation(mlDep, ANNOTATION_ISTIO_RETRIES, "")
	istioRetriesTimeoutAnnotation := utils2.GetAnnotation(mlDep, ANNOTATION_ISTIO_RETRIES_TIMEOUT, "1")
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
	vsvc := &istio.VirtualService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      seldonId,
			Namespace: namespace,
		},
		Spec: istio_networking.VirtualService{
			Hosts:    utils2.HostsFromAnnotation(mlDep, ANNOTATION_ISTIO_HOST, "*"),
			Gateways: []string{utils2.GetAnnotation(mlDep, ANNOTATION_ISTIO_GATEWAY, istio_gateway)},
			Http: []*istio_networking.HTTPRoute{
				{
					Match: []*istio_networking.HTTPMatchRequest{
						{
							Uri: &istio_networking.StringMatch{MatchType: &istio_networking.StringMatch_Prefix{Prefix: "/seldon/" + namespace + "/" + mlDep.Name + "/"}},
						},
					},
					Rewrite: &istio_networking.HTTPRewrite{Uri: "/"},
				},
				{
					Match: []*istio_networking.HTTPMatchRequest{
						{
							Uri: &istio_networking.StringMatch{MatchType: &istio_networking.StringMatch_Regex{Regex: constants.GRPCRegExMatchIstio}},
							Headers: map[string]*istio_networking.StringMatch{
								"seldon":    {MatchType: &istio_networking.StringMatch_Exact{Exact: mlDep.Name}},
								"namespace": {MatchType: &istio_networking.StringMatch_Exact{Exact: namespace}},
							},
						},
					},
				},
			},
		},
	}

	// Add retries
	if istioRetries > 0 {
		vsvc.Spec.Http[0].Retries = &istio_networking.HTTPRetry{Attempts: int32(istioRetries), PerTryTimeout: &duration.Duration{Seconds: int64(istioRetriesTimeout)}, RetryOn: "gateway-error,connect-failure,refused-stream"}
		vsvc.Spec.Http[1].Retries = &istio_networking.HTTPRetry{Attempts: int32(istioRetries), PerTryTimeout: &duration.Duration{Seconds: int64(istioRetriesTimeout)}, RetryOn: "gateway-error,connect-failure,refused-stream"}
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
		pSvcFqdn := createFqdn(pSvcName, namespace)

		drule := &istio.DestinationRule{
			ObjectMeta: metav1.ObjectMeta{
				Name:      pSvcName,
				Namespace: namespace,
			},
			Spec: istio_networking.DestinationRule{
				Host: pSvcFqdn,
				Subsets: []*istio_networking.Subset{
					{
						Name: p.Name,
						Labels: map[string]string{
							"version": p.Labels["version"],
						},
					},
				},
				TrafficPolicy: &istio_networking.TrafficPolicy{ConnectionPool: &istio_networking.ConnectionPoolSettings{Http: &istio_networking.ConnectionPoolSettings_HTTPSettings{IdleTimeout: &duration.Duration{Seconds: 60}}}},
			},
		}

		if istioTLSMode != "" {
			drule.Spec.TrafficPolicy = &istio_networking.TrafficPolicy{
				Tls: &istio_networking.ClientTLSSettings{
					Mode: istio_networking.ClientTLSSettings_TLSmode(istio_networking.ClientTLSSettings_TLSmode_value[istioTLSMode]),
				},
			}
		}
		drules[i] = drule

		if p.Shadow == true {
			//if there's a shadow then add a mirror section to the VirtualService

			vsvc.Spec.Http[0].Mirror = &istio_networking.Destination{
				Host:   pSvcFqdn,
				Subset: p.Name,
				Port: &istio_networking.PortSelector{
					Number: uint32(ports[i].httpPort),
				},
			}

			vsvc.Spec.Http[1].Mirror = &istio_networking.Destination{
				Host:   pSvcFqdn,
				Subset: p.Name,
				Port: &istio_networking.PortSelector{
					Number: uint32(ports[i].grpcPort),
				},
			}

			if p.Traffic > 0 {
				//if shadow predictor's traffic is greater than 0, set the mirror percentage (like https://istio.io/latest/docs/tasks/traffic-management/mirroring/#mirroring-traffic-to-v2) in VirtualService
				vsvc.Spec.Http[0].MirrorPercentage = &istio_networking.Percent{
					Value: float64(p.Traffic),
				}

				vsvc.Spec.Http[1].MirrorPercentage = &istio_networking.Percent{
					Value: float64(p.Traffic),
				}
			}

			continue
		}

		//we split by adding different routes with their own Weights
		//so not by tag - different destinations (like https://istio.io/docs/tasks/traffic-management/traffic-shifting/) distinguished by host
		routesHttp[routesIdx] = &istio_networking.HTTPRouteDestination{
			Destination: &istio_networking.Destination{
				Host:   pSvcFqdn,
				Subset: p.Name,
				Port: &istio_networking.PortSelector{
					Number: uint32(ports[i].httpPort),
				},
			},
			Weight: p.Traffic,
		}
		routesGrpc[routesIdx] = &istio_networking.HTTPRouteDestination{
			Destination: &istio_networking.Destination{
				Host:   pSvcFqdn,
				Subset: p.Name,
				Port: &istio_networking.PortSelector{
					Number: uint32(ports[i].grpcPort),
				},
			},
			Weight: p.Traffic,
		}
		routesIdx += 1

	}
	vsvc.Spec.Http[0].Route = routesHttp
	vsvc.Spec.Http[1].Route = routesGrpc

	vscs := make([]*istio.VirtualService, 1)
	vscs[0] = vsvc
	return vscs, drules, nil

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
func (r *SeldonDeploymentReconciler) createComponents(ctx context.Context, mlDep *machinelearningv1.SeldonDeployment, securityContext *corev1.PodSecurityContext, log logr.Logger) (*components, error) {
	c := components{}
	c.serviceDetails = map[string]*machinelearningv1.ServiceStatus{}
	seldonId := machinelearningv1.GetSeldonDeploymentName(mlDep)
	namespace := utils2.GetNamespace(mlDep)

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

	for i := 0; i < len(mlDep.Spec.Predictors); i++ {
		p := mlDep.Spec.Predictors[i]
		noEngine := strings.ToLower(p.Annotations[machinelearningv1.ANNOTATION_NO_ENGINE]) == "true"
		pSvcName := machinelearningv1.GetPredictorKey(mlDep, &p)
		log.Info("pSvcName", "val", pSvcName)

		// SSL config is used to set ssl on each container
		certSecretRefName := ""
		predictorCertConfig := p.SSL
		if predictorCertConfig != nil {
			certSecretRefName = predictorCertConfig.CertSecretName
		}
		// Add engine deployment if separate
		hasSeparateEnginePod := machinelearningv1.HasSeparateEnginePod(mlDep.Spec)
		if hasSeparateEnginePod && !noEngine {
			deploy, err := createEngineDeployment(mlDep, &p, pSvcName, engine_http_port, engine_grpc_port)
			if err != nil {
				return nil, err
			}
			if securityContext != nil {
				deploy.Spec.Template.Spec.SecurityContext = securityContext
			}

			// Add secret ref name to the container of the separate svcorch if created
			if len(certSecretRefName) > 0 {
				utils.MountSecretToDeploymentContainers(deploy, certSecretRefName, envDefaultCertMountPath)
				certEnvVar := &corev1.EnvVar{Name: SELDON_MOUNT_PATH_ENV_NAME, Value: envDefaultCertMountPath}
				utils.AddEnvVarToDeploymentContainers(deploy, certEnvVar)
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
			deploy := createDeploymentWithoutEngine(depName, seldonId, cSpec, &p, mlDep, securityContext, true)

			if cSpec.KedaSpec != nil { // Add KEDA if needed
				r.Log.Info("Creating keda scaled object", "deployment", depName)
				c.kedaScaledObjects = append(c.kedaScaledObjects, createKeda(cSpec, depName, seldonId, namespace))
			} else if cSpec.HpaSpec != nil { // Add HPA if needed
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

			// Add PDB if needed
			if cSpec.PdbSpec != nil {
				c.pdbs = append(c.pdbs, createPdb(cSpec, depName, seldonId, namespace))
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

						httpPort := int(svc.Spec.Ports[0].Port)
						grpcPort := int(svc.Spec.Ports[1].Port)

						externalPorts[i] = httpGrpcPorts{httpPort: httpPort, grpcPort: grpcPort}
						psvc, err := createPredictorService(pSvcName, seldonId, &p, mlDep, httpPort, grpcPort, false, log, &c)
						if err != nil {
							return nil, err
						}
						psvc = addLabelsToService(psvc, pu, &p)

						c.services = append(c.services, psvc)

						c.serviceDetails[pSvcName] = &machinelearningv1.ServiceStatus{
							SvcName:      pSvcName,
							HttpEndpoint: pSvcName + "." + namespace + ":" + strconv.Itoa(httpPort),
							GrpcEndpoint: pSvcName + "." + namespace + ":" + strconv.Itoa(grpcPort),
						}
					}
				}
			}
			c.deployments = append(c.deployments, deploy)
		}

		pi := NewPrePackedInitializer(ctx, r.ClientSet)
		err = pi.addModelServersAndInitContainers(mlDep, &p, &c, &p.Graph, securityContext, log)
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
					return nil, fmt.Errorf("engine not separate and no pu with localhost service - not clear where to inject engine")
				}
				// find the deployment with a container for the pu marked for engine
				for i := range c.deployments {
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
					return nil, fmt.Errorf("engine not separate and no deployment for pu with localhost service - not clear where to inject engine")
				}
				err := addEngineToDeployment(mlDep, &p, engine_http_port, engine_grpc_port, pSvcName, deploy)
				if err != nil {
					return nil, err
				}

			}

			// Find the current deployment and add the environment variables for the certificate
			if len(certSecretRefName) > 0 {
				sPodSpec, idx := utils.GetSeldonPodSpecForPredictiveUnit(&p, p.Graph.Name)
				currentDeployName := machinelearningv1.GetDeploymentName(mlDep, p, sPodSpec, idx)
				for i := 0; i < len(c.deployments); i++ {
					d := c.deployments[i]
					if strings.Compare(d.Name, currentDeployName) == 0 {
						utils.MountSecretToDeploymentContainers(d, certSecretRefName, envDefaultCertMountPath)
						certEnvVar := &corev1.EnvVar{Name: SELDON_MOUNT_PATH_ENV_NAME, Value: envDefaultCertMountPath}
						utils.AddEnvVarToDeploymentContainers(d, certEnvVar)
						break
					}
				}
			}

			psvc, err := createPredictorService(pSvcName, seldonId, &p, mlDep, engine_http_port, engine_grpc_port, false, log, &c)
			if err != nil {

				return nil, err
			}

			c.services = append(c.services, psvc)
			c.serviceDetails[pSvcName] = &machinelearningv1.ServiceStatus{
				SvcName:      pSvcName,
				HttpEndpoint: pSvcName + "." + namespace + ":" + strconv.Itoa(engine_http_port),
				GrpcEndpoint: pSvcName + "." + namespace + ":" + strconv.Itoa(engine_grpc_port),
			}

			externalPorts[i] = httpGrpcPorts{httpPort: engine_http_port, grpcPort: engine_grpc_port}
		}

		ei := NewExplainerInitializer(ctx, r.ClientSet)
		err = ei.createExplainer(mlDep, &p, &c, pSvcName, securityContext, log)
		if err != nil {
			return nil, err
		}
	}

	// Create the addressable as all services are created when SeldonDeployment is ready
	c.addressable, err = createAddressableResource(mlDep, namespace, externalPorts)
	if err != nil {
		return nil, err
	}

	//TODO Fixme - not changed to handle per predictor scenario
	if utils.GetEnv(ENV_ISTIO_ENABLED, "false") == "true" {
		vsvcs, dstRule, err := createIstioResources(mlDep, seldonId, namespace, externalPorts)
		if err != nil {
			return nil, err
		}
		c.virtualServices = append(c.virtualServices, vsvcs...)
		c.destinationRules = append(c.destinationRules, dstRule...)
	}
	return &c, nil
}

// Creates Service for Predictor - exposed externally (ambassador or istio)
func createPredictorService(pSvcName string, seldonId string, p *machinelearningv1.PredictorSpec,
	mlDep *machinelearningv1.SeldonDeployment,
	engine_http_port int,
	engine_grpc_port int,
	isExplainer bool,
	log logr.Logger,
	c *components) (pSvc *corev1.Service, err error) {
	namespace := utils2.GetNamespace(mlDep)
	psvc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pSvcName,
			Namespace: namespace,
			Labels: map[string]string{
				machinelearningv1.Label_seldon_app: pSvcName,
				machinelearningv1.Label_seldon_id:  seldonId,
				machinelearningv1.Label_managed_by: machinelearningv1.Label_value_seldon,
			},
			Annotations: map[string]string{},
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

	if utils.GetEnv(ENV_AMBASSADOR_ENABLED, "false") == "true" {
		if utils.GetEnv(ENV_AMBASSADOR_VERSION, "v1") == "v1" {
			//Create top level Service
			ambassadorConfig, err := ambassador.GetAmbassadorConfigs(mlDep, p, pSvcName, engine_http_port, engine_grpc_port, isExplainer)
			if err != nil {
				return nil, err
			}
			psvc.Annotations[AMBASSADOR_ANNOTATION] = ambassadorConfig
		} else {
			mappings, tlsContexts, err := ambassador.GetV2AmbassadorConfigs(mlDep, p, pSvcName, engine_http_port, engine_grpc_port, isExplainer)
			if err != nil {
				return nil, err
			}
			c.mappings = append(c.mappings, mappings...)
			c.tlsContexts = append(c.tlsContexts, tlsContexts...)
		}

	}
	if utils2.GetAnnotation(mlDep, machinelearningv1.ANNOTATION_HEADLESS_SVC, "false") != "false" {
		log.Info("Creating Headless SVC")
		psvc.Spec.ClusterIP = "None"
	}
	// Add annotations from predictorspec
	for k, v := range p.Annotations {
		psvc.Annotations[k] = v
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
	containerServiceKey := fmt.Sprintf("%s-%s", machinelearningv1.Label_seldon_app_svc, con.Name)
	containerServiceValue := machinelearningv1.GetContainerServiceName(mlDep.Name, p, con)
	pSvcName := machinelearningv1.GetPredictorKey(mlDep, &p)
	pu := machinelearningv1.GetPredictiveUnit(&p.Graph, con.Name)

	// only create services for containers defined as pus in the graph
	if pu == nil {
		return nil
	}
	namespace := utils2.GetNamespace(mlDep)

	c.serviceDetails[containerServiceValue] = &machinelearningv1.ServiceStatus{
		SvcName:      containerServiceValue,
		HttpEndpoint: containerServiceValue + "." + namespace + ":" + strconv.Itoa(int(pu.Endpoint.HttpPort)),
		GrpcEndpoint: containerServiceValue + "." + namespace + ":" + strconv.Itoa(int(pu.Endpoint.GrpcPort))}

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      containerServiceValue,
			Namespace: namespace,
			Labels: map[string]string{
				containerServiceKey:                containerServiceValue,
				machinelearningv1.Label_seldon_id:  seldonId,
				machinelearningv1.Label_seldon_app: pSvcName},
			Annotations: map[string]string{},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Protocol:   corev1.ProtocolTCP,
					Port:       pu.Endpoint.HttpPort,
					TargetPort: intstr.FromInt(int(pu.Endpoint.HttpPort)),
					Name:       "http",
				},
				{
					Protocol:   corev1.ProtocolTCP,
					Port:       pu.Endpoint.GrpcPort,
					TargetPort: intstr.FromInt(int(pu.Endpoint.GrpcPort)),
					Name:       "grpc",
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

	existingHttpPort := machinelearningv1.GetPort("http", con.Ports)
	if existingHttpPort == nil || con.Ports == nil {
		con.Ports = append(con.Ports, corev1.ContainerPort{Name: "http", ContainerPort: pu.Endpoint.HttpPort, Protocol: corev1.ProtocolTCP})
	}
	existingGrpcPort := machinelearningv1.GetPort("grpc", con.Ports)
	if existingGrpcPort == nil || con.Ports == nil {
		con.Ports = append(con.Ports, corev1.ContainerPort{Name: "grpc", ContainerPort: pu.Endpoint.GrpcPort, Protocol: corev1.ProtocolTCP})
	}

	// Add annotations from predictorspec
	for k, v := range p.Annotations {
		svc.Annotations[k] = v
	}

	// Backwards compatible additions. From 1.5.0 onwards could always call httpPort as both should be available but for
	// previously wrapped components need to look at transport.
	// TODO: deprecate and just call httpPort
	if con.LivenessProbe == nil {
		if mlDep.Spec.Transport == machinelearningv1.TransportGrpc || pu.Endpoint.Type == machinelearningv1.GRPC {
			con.LivenessProbe = &corev1.Probe{ProbeHandler: corev1.ProbeHandler{TCPSocket: &corev1.TCPSocketAction{Port: intstr.FromInt(int(pu.Endpoint.GrpcPort))}}, InitialDelaySeconds: 60, PeriodSeconds: 5, SuccessThreshold: 1, FailureThreshold: 3, TimeoutSeconds: 1}
		} else {
			con.LivenessProbe = &corev1.Probe{ProbeHandler: corev1.ProbeHandler{TCPSocket: &corev1.TCPSocketAction{Port: intstr.FromInt(int(pu.Endpoint.HttpPort))}}, InitialDelaySeconds: 60, PeriodSeconds: 5, SuccessThreshold: 1, FailureThreshold: 3, TimeoutSeconds: 1}
		}
	}
	if con.ReadinessProbe == nil {
		if mlDep.Spec.Transport == machinelearningv1.TransportGrpc || pu.Endpoint.Type == machinelearningv1.GRPC {
			con.ReadinessProbe = &corev1.Probe{ProbeHandler: corev1.ProbeHandler{TCPSocket: &corev1.TCPSocketAction{Port: intstr.FromInt(int(pu.Endpoint.GrpcPort))}}, InitialDelaySeconds: 20, PeriodSeconds: 5, SuccessThreshold: 1, FailureThreshold: 3, TimeoutSeconds: 1}
		} else {
			con.ReadinessProbe = &corev1.Probe{ProbeHandler: corev1.ProbeHandler{TCPSocket: &corev1.TCPSocketAction{Port: intstr.FromInt(int(pu.Endpoint.HttpPort))}}, InitialDelaySeconds: 20, PeriodSeconds: 5, SuccessThreshold: 1, FailureThreshold: 3, TimeoutSeconds: 1}
		}
	}

	// Add lifecycle probe
	if con.Lifecycle == nil {
		con.Lifecycle = &corev1.Lifecycle{PreStop: &corev1.LifecycleHandler{Exec: &corev1.ExecAction{Command: []string{"/bin/sh", "-c", "/bin/sleep 10"}}}}
	}

	//
	// Backwards compatibility - set to either Http or Grpc
	//
	// TODO: deprecate and remove
	if !utils.HasEnvVar(con.Env, machinelearningv1.ENV_PREDICTIVE_UNIT_SERVICE_PORT) {
		if pu.Endpoint.Type == machinelearningv1.GRPC || mlDep.Spec.Transport == machinelearningv1.TransportGrpc {
			con.Env = append(con.Env, corev1.EnvVar{Name: machinelearningv1.ENV_PREDICTIVE_UNIT_SERVICE_PORT, Value: strconv.Itoa(int(pu.Endpoint.GrpcPort))})
		} else {
			con.Env = append(con.Env, corev1.EnvVar{Name: machinelearningv1.ENV_PREDICTIVE_UNIT_SERVICE_PORT, Value: strconv.Itoa(int(pu.Endpoint.HttpPort))})
		}
	}

	addPortEnvs(pu, con)
	addModelEnvs(pu, con)

	// Always set the predictive and deployment identifiers

	labels, err := json.Marshal(p.Labels)
	if err != nil {
		labels = []byte("{}")
	}

	con.Env = append(con.Env, []corev1.EnvVar{
		{Name: machinelearningv1.ENV_PREDICTIVE_UNIT_ID, Value: con.Name},
		{Name: MLServerModelNameEnv, Value: con.Name},
		{Name: machinelearningv1.ENV_PREDICTIVE_UNIT_IMAGE, Value: con.Image},
		{Name: machinelearningv1.ENV_PREDICTOR_ID, Value: p.Name},
		{Name: machinelearningv1.ENV_PREDICTOR_LABELS, Value: string(labels)},
		{Name: machinelearningv1.ENV_SELDON_DEPLOYMENT_ID, Value: mlDep.ObjectMeta.Name},
		{Name: machinelearningv1.ENV_SELDON_EXECUTOR_ENABLED, Value: strconv.FormatBool(true)},
	}...)

	//Add Metric Env Var
	predictiveUnitMetricsPortName := utils.GetEnv(machinelearningv1.ENV_PREDICTIVE_UNIT_METRICS_PORT_NAME, constants.DefaultMetricsPortName)
	metricPort := getPort(predictiveUnitMetricsPortName, con.Ports)
	if metricPort != nil {
		con.Env = append(con.Env, []corev1.EnvVar{
			{Name: machinelearningv1.ENV_PREDICTIVE_UNIT_SERVICE_PORT_METRICS, Value: strconv.Itoa(int(metricPort.ContainerPort))},
			{Name: machinelearningv1.ENV_PREDICTIVE_UNIT_METRICS_ENDPOINT, Value: getPrometheusPath(mlDep)},
			{Name: MLServerMetricsPortEnv, Value: strconv.Itoa(int(metricPort.ContainerPort))},
			{Name: MLServerMetricsEndpointEnv, Value: getPrometheusPath(mlDep)},
		}...)
	}

	return svc
}

// compareServices ignores those fields set by k8s by default.
// Specifically, the fields to be compared are:
//   - Annotations
//   - Labels
//   - Spec.Ports
//   - Spec.Type
//   - Spec.Selector
//   - Spec.SessionAffinity
func compareServices(desiredSvc, found *corev1.Service) bool {
	return equality.Semantic.DeepEqual(desiredSvc.Annotations, found.Annotations) &&
		equality.Semantic.DeepEqual(desiredSvc.Labels, found.Labels) &&
		equality.Semantic.DeepEqual(desiredSvc.Spec.Ports, found.Spec.Ports) &&
		equality.Semantic.DeepEqual(desiredSvc.Spec.Selector, found.Spec.Selector) &&
		equality.Semantic.DeepEqual(desiredSvc.Spec.Type, found.Spec.Type) &&
		equality.Semantic.DeepEqual(desiredSvc.Spec.SessionAffinity, found.Spec.SessionAffinity)
}

// updateService ignores those immutable fields and fields set by k8s by default.
// Specifically, the updated fields are:
//   - Annotations
//   - Labels
//   - Spec.Ports
//   - Spec.Type
//   - Spec.SessionAffinity
func updateService(foundSvc, desiredSvc *corev1.Service) *corev1.Service {
	foundSvc.Annotations = desiredSvc.Annotations
	foundSvc.Labels = desiredSvc.Labels
	foundSvc.Spec.Ports = desiredSvc.Spec.Ports
	foundSvc.Spec.Type = desiredSvc.Spec.Type
	foundSvc.Spec.SessionAffinity = desiredSvc.Spec.SessionAffinity
	return foundSvc
}

func addModelEnvs(pu *machinelearningv1.PredictiveUnit, con *corev1.Container) {
	if len(pu.Parameters) > 0 {
		// Set V1 env vars
		paramsEnvVar := corev1.EnvVar{
			Name:  machinelearningv1.ENV_PREDICTIVE_UNIT_PARAMETERS,
			Value: utils.GetPredictiveUnitAsJson(pu.Parameters),
		}
		con.Env = utils.SetEnvVar(con.Env, paramsEnvVar, false)
	}

	// If storageUri is present, set model URI for V2
	if len(pu.ModelURI) != 0 {
		modelURIEnv := corev1.EnvVar{
			Name:  MLServerModelURIEnv,
			Value: DefaultModelLocalMountPath,
		}
		con.Env = utils.SetEnvVar(con.Env, modelURIEnv, false)
	}
}

func addPortEnvs(pu *machinelearningv1.PredictiveUnit, con *corev1.Container) {

	// HTTP Ports
	httpPort := strconv.Itoa(int(pu.Endpoint.HttpPort))
	httpEnvVarNames := []string{
		machinelearningv1.ENV_PREDICTIVE_UNIT_HTTP_SERVICE_PORT,
		MLServerHTTPPortEnv,
	}
	for _, envVarName := range httpEnvVarNames {
		httpEnvVar := corev1.EnvVar{
			Name:  envVarName,
			Value: httpPort,
		}
		con.Env = utils.SetEnvVar(con.Env, httpEnvVar, false)
	}

	// gRPC Ports
	grpcPort := strconv.Itoa(int(pu.Endpoint.GrpcPort))
	grpcEnvVarNames := []string{
		machinelearningv1.ENV_PREDICTIVE_UNIT_GRPC_SERVICE_PORT,
		MLServerGRPCPortEnv,
	}
	for _, envVarName := range grpcEnvVarNames {
		grpcEnvVar := corev1.EnvVar{
			Name:  envVarName,
			Value: grpcPort,
		}
		con.Env = utils.SetEnvVar(con.Env, grpcEnvVar, false)
	}
}

func createDeploymentWithoutEngine(depName string, seldonId string, seldonPodSpec *machinelearningv1.SeldonPodSpec, p *machinelearningv1.PredictorSpec, mlDep *machinelearningv1.SeldonDeployment, podSecurityContext *corev1.PodSecurityContext, metricsEnabled bool) *appsv1.Deployment {
	deploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      depName,
			Namespace: utils2.GetNamespace(mlDep),
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
			Strategy:                appsv1.DeploymentStrategy{Type: appsv1.RollingUpdateDeploymentStrategyType, RollingUpdate: &appsv1.RollingUpdateDeployment{MaxUnavailable: &intstr.IntOrString{StrVal: "10%"}}},
			ProgressDeadlineSeconds: p.ProgressDeadlineSeconds,
		},
	}

	if deploy.Spec.Template.Annotations == nil {
		deploy.Spec.Template.Annotations = map[string]string{}
	}

	if metricsEnabled {
		// Add prometheus annotations
		deploy.Spec.Template.Annotations["prometheus.io/path"] = getPrometheusPath(mlDep)
		deploy.Spec.Template.Annotations["prometheus.io/scrape"] = "true"
	}

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

	if deploy.Spec.Template.Spec.SecurityContext == nil {
		// Add Pod Security Context
		deploy.Spec.Template.Spec.SecurityContext = podSecurityContext
	}

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

	// Set TerminationGracePeriodSeconds
	var terminationGracePeriod int64 = 20
	deploy.Spec.Template.Spec.TerminationGracePeriodSeconds = &terminationGracePeriod
	if seldonPodSpec != nil && seldonPodSpec.Spec.TerminationGracePeriodSeconds != nil {
		deploy.Spec.Template.Spec.TerminationGracePeriodSeconds = seldonPodSpec.Spec.TerminationGracePeriodSeconds
	}

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

func (r *SeldonDeploymentReconciler) createAmbassadorMappings(components *components, instance *machinelearningv1.SeldonDeployment, log logr.Logger) (bool, error) {
	ready := true
	for _, mapping := range components.mappings {
		if err := controllerutil.SetControllerReference(instance, mapping, r.Scheme); err != nil {
			return ready, err
		}
		found := &v2.Mapping{}
		err := r.Get(context.TODO(), types.NamespacedName{Name: mapping.Name, Namespace: mapping.Namespace}, found)
		if err != nil && errors.IsNotFound(err) {
			ready = false
			log.Info("Creating Ambassador Mapping", "namespace", mapping.Namespace, "name", mapping.Name)
			err = r.Create(context.TODO(), mapping)
			if err != nil {
				return ready, err
			}
			r.Recorder.Eventf(instance, corev1.EventTypeNormal, constants.EventsCreateAmbassadorMapping, "Created AmbassadorMapping %q", mapping.GetName())
		} else if err != nil {
			return ready, err
		} else {
			// Update the found object and write the result back if there are any changes
			if !equality.Semantic.DeepEqual(mapping.Spec, found.Spec) {
				desiredSvc := found.DeepCopy()
				found.Spec = mapping.Spec
				log.Info("Updating Ambassador Mapping", "namespace", mapping.Namespace, "name", mapping.Name)
				err = r.Update(context.TODO(), found)
				if err != nil {
					return ready, err
				}

				// Check if what came back from server modulo the defaults applied by k8s is the same or not
				if !equality.Semantic.DeepEqual(desiredSvc.Spec, found.Spec) {
					ready = false
					r.Recorder.Eventf(instance, corev1.EventTypeNormal, constants.EventsUpdateAmbassadorMapping, "Updated Ambassador Mapping %q", mapping.GetName())
					//For debugging we will show the difference
					diff, err := kmp.SafeDiff(desiredSvc.Spec, found.Spec)
					if err != nil {
						log.Error(err, "Failed to diff")
					} else {
						log.Info(fmt.Sprintf("Difference in Mapping: %v", diff))
					}
				} else {
					log.Info("The Ambassador Mappings are the same - api server defaults ignored")
				}
			} else {
				log.Info("Found identical Ambassador Mapping", "namespace", found.Namespace, "name", found.Name)
			}
		}
	}

	//Cleanup unused Ambassador mappings. This should usually only happen on Operator upgrades where there is a breaking change to the names of the VirtualServices created
	//Only run if we have virtualservices to create - implies we are running with istio active
	if len(components.mappings) > 0 && ready {
		cleaner := AmbassadoroResourceCleaner{instance: instance, client: r.Client, mappings: components.mappings, logger: r.Log}
		deleted, err := cleaner.cleanUnusedAmbassadorMappings()
		if err != nil {
			return ready, err
		}
		for _, mappingDeleted := range deleted {
			r.Recorder.Eventf(instance, corev1.EventTypeNormal, constants.EventsDeleteAmbassadorMapping, "Delete Ambassador mapping %q", mappingDeleted.GetName())
		}
	}

	if ready {
		var reason string
		if len(components.mappings) > 0 {
			reason = machinelearningv1.AmbassadorMappingReady
		} else {
			reason = machinelearningv1.AmbassadorMappingNotDefined
		}
		instance.Status.CreateCondition(machinelearningv1.AmbassadorMappingsReady, true, reason)
	} else {
		instance.Status.CreateCondition(machinelearningv1.AmbassadorMappingsReady, false, machinelearningv1.AmbassadorMappingNotReady)
	}

	return ready, nil
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
			if !cmp.Equal(&svc.Spec, &found.Spec, protocmp.Transform()) {

				desiredSvc := found.DeepCopy()
				found.Spec = *svc.Spec.DeepCopy()
				log.Info("Updating Virtual Service", "namespace", svc.Namespace, "name", svc.Name)
				err = r.Update(context.TODO(), found)
				if err != nil {
					return ready, err
				}

				// Check if what came back from server modulo the defaults applied by k8s is the same or not
				if !cmp.Equal(&desiredSvc.Spec, &found.Spec, protocmp.Transform()) {
					ready = false
					if diff := cmp.Diff(&desiredSvc.Spec, &found.Spec, protocmp.Transform()); diff != "" {
						log.Info("Found virtual service differences", "diff", diff)
					}
					r.Recorder.Eventf(instance, corev1.EventTypeNormal, constants.EventsUpdateVirtualService, "Updated VirtualService %q", svc.GetName())
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
			if !cmp.Equal(&drule.Spec, &found.Spec, protocmp.Transform()) {
				desiredDrule := found.DeepCopy()
				found.Spec = *drule.Spec.DeepCopy()
				log.Info("Updating Istio Destination Rule", "namespace", drule.Namespace, "name", drule.Name)
				err = r.Update(context.TODO(), found)
				if err != nil {
					return ready, err
				}

				// Check if what came back from server modulo the defaults applied by k8s is the same or not
				if !cmp.Equal(&desiredDrule.Spec, &found.Spec, protocmp.Transform()) {
					ready = false
					if diff := cmp.Diff(&desiredDrule.Spec, &found.Spec, protocmp.Transform()); diff != "" {
						log.Info("Found virtual service differences", "diff", diff)
					}
					r.Recorder.Eventf(instance, corev1.EventTypeNormal, constants.EventsUpdateDestinationRule, "Updated DestinationRule %q", drule.GetName())
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
		cleaner := IstioResourceCleaner{instance: instance, client: r.Client, virtualServices: components.virtualServices, logger: r.Log}
		deleted, err := cleaner.cleanUnusedVirtualServices()
		if err != nil {
			return ready, err
		}
		for _, vsvcDeleted := range deleted {
			r.Recorder.Eventf(instance, corev1.EventTypeNormal, constants.EventsDeleteVirtualService, "Delete VirtualService %q", vsvcDeleted.GetName())
		}
	}

	if ready {
		var reason string
		if len(components.virtualServices) > 0 {
			reason = machinelearningv1.VirtualServiceReady
		} else {
			reason = machinelearningv1.VirtualServiceNotDefined
		}
		instance.Status.CreateCondition(machinelearningv1.VirtualServicesReady, true, reason)
	} else {
		instance.Status.CreateCondition(machinelearningv1.VirtualServicesReady, false, machinelearningv1.VirtualServiceNotReady)
	}

	return ready, nil
}

// Create Services specified in components.
func (r *SeldonDeploymentReconciler) createServices(components *components, instance *machinelearningv1.SeldonDeployment, all bool, log logr.Logger) (bool, error) {
	ready := true
	for _, svc := range components.services {
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
			// Configure addressable status so it can be reached through duck-typing
			instance.Status.Address = components.addressable
			// Check if the found one modulo the defaults applied by k8s has changed
			// Update the found object and write the result back if there are any changes
			if !compareServices(svc, found) {
				// Check if svc selectors have changed - if so then we need to recreate svc
				if !equality.Semantic.DeepEqual(svc.Spec.Selector, found.Spec.Selector) {
					ready = false
					// Delete the svc as selectors are immutable
					err = r.Delete(context.TODO(), found, client.PropagationPolicy(metav1.DeletePropagationForeground))
					if err != nil {
						return ready, err
					}
					r.Recorder.Eventf(instance, corev1.EventTypeNormal, constants.EventsDeleteService, "Deleted Service %q", svc.GetName())
					// Re-create svc
					err = r.Create(context.TODO(), svc)
					if err != nil {
						return ready, err
					}
					r.Recorder.Eventf(instance, corev1.EventTypeNormal, constants.EventsCreateService, "Recreated Service %q", svc.GetName())
				} else {
					// Just update those required and mutable fields of the found one and leave other defaults applied by k8s unchanged
					updateService(found, svc)
					// desiredSvc is a temporary deep copy of found object just before it is re updated by the result back from the api server
					desiredSvc := found.DeepCopy()
					log.Info("Updating Service", "all", all, "namespace", svc.Namespace, "name", svc.Name)
					err = r.Update(context.TODO(), found)
					if err != nil {
						return ready, err
					}

					// Check if what came back from server modulo the defaults applied by k8s is the same or not
					if !compareServices(desiredSvc, found) {
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

	if all && ready {
		instance.Status.CreateCondition(machinelearningv1.ServicesReady, true, machinelearningv1.SvcReadyReason)
	} else {
		instance.Status.CreateCondition(machinelearningv1.ServicesReady, false, machinelearningv1.SvcNotReadyReason)
	}

	return ready, nil
}

func (r *SeldonDeploymentReconciler) createKedaScaledObjects(components *components, instance *machinelearningv1.SeldonDeployment, log logr.Logger) (bool, error) {
	ready := true
	scaledObjSet := make(map[string]bool)
	for _, scaledObj := range components.kedaScaledObjects {
		if err := ctrl.SetControllerReference(instance, scaledObj, r.Scheme); err != nil {
			return ready, err
		}
		scaledObjSet[scaledObj.Name] = true
		found := &kedav1alpha1.ScaledObject{}
		err := r.Get(context.TODO(), types.NamespacedName{Name: scaledObj.Name, Namespace: scaledObj.Namespace}, found)
		if err != nil && errors.IsNotFound(err) {
			ready = false
			log.Info("Creating KEDA ScaledObject", "namespace", scaledObj.Namespace, "name", scaledObj.Name)
			err = r.Create(context.TODO(), scaledObj)
			if err != nil {
				return ready, err
			}
			r.Recorder.Eventf(instance, corev1.EventTypeNormal, constants.EventsCreateScaledObject, "Created KEDA ScaledObject %q", scaledObj.GetName())
		} else if err != nil {
			return ready, err
		} else {
			// Update the found object and write the result back if there are any changes
			if !equality.Semantic.DeepEqual(scaledObj.Spec, found.Spec) {

				desiredScaledObj := found.DeepCopy()
				found.Spec = scaledObj.Spec

				log.Info("Updating KEDA ScaledObject", "namespace", scaledObj.Namespace, "name", scaledObj.Name)
				err = r.Update(context.TODO(), found)
				if err != nil {
					return ready, err
				}

				// Check if what came back from server modulo the defaults applied by k8s is the same or not
				if !equality.Semantic.DeepEqual(desiredScaledObj.Spec, found.Spec) {
					ready = false
					r.Recorder.Eventf(instance, corev1.EventTypeNormal, constants.EventsUpdateHPA, "Updated KEDA ScaledObject %q", scaledObj.GetName())
					//For debugging we will show the difference
					diff, err := kmp.SafeDiff(desiredScaledObj.Spec, found.Spec)
					if err != nil {
						log.Error(err, "Failed to diff")
					} else {
						log.Info(fmt.Sprintf("Difference in KEDA ScaledObjects: %v", diff))
					}
				} else {
					log.Info("The KEDA ScaledObjects are the same - api server defaults ignored")
				}

			} else {
				log.Info("Found identical KEDA ScaledObject", "namespace", found.Namespace, "name", found.Name, "status", found.Status)
			}
		}

	}

	// For all Deployments check if any ScaledObjects exist and they are not required
	for _, deploy := range components.deployments {
		if _, ok := scaledObjSet[deploy.Name]; !ok {
			found := &kedav1alpha1.ScaledObject{}
			err := r.Get(context.TODO(), types.NamespacedName{Name: deploy.Name, Namespace: deploy.Namespace}, found)
			if err != nil {
				if !errors.IsNotFound(err) {
					return false, err
				}
				// Do nothing
			} else {
				// Delete ScaledObject
				log.Info("Deleting KEDA ScaledObject", "name", deploy.Name)
				err := r.Delete(context.TODO(), found, client.PropagationPolicy(metav1.DeletePropagationForeground))
				if err != nil {
					return ready, err
				}
			}
		}
	}

	if ready {
		var reason string
		if len(components.kedaScaledObjects) > 0 {
			reason = machinelearningv1.KedaReadyReason
		} else {
			reason = machinelearningv1.KedaNotDefinedReason
		}
		instance.Status.CreateCondition(machinelearningv1.KedaReady, true, reason)
	} else {
		instance.Status.CreateCondition(machinelearningv1.KedaReady, false, machinelearningv1.KedaNotReadyReason)
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

	if ready {
		var reason string
		if len(components.hpas) > 0 {
			reason = machinelearningv1.HpaReadyReason
		} else {
			reason = machinelearningv1.HpaNotDefinedReason
		}
		instance.Status.CreateCondition(machinelearningv1.HpasReady, true, reason)
	} else {
		instance.Status.CreateCondition(machinelearningv1.HpasReady, false, machinelearningv1.HpaNotReadyReason)
	}

	return ready, nil
}

// Create Services specified in components.
func (r *SeldonDeploymentReconciler) createPdbs(components *components, instance *machinelearningv1.SeldonDeployment, log logr.Logger) (bool, error) {
	ready := true
	pdbSet := make(map[string]bool)
	for _, pdb := range components.pdbs {
		if err := ctrl.SetControllerReference(instance, pdb, r.Scheme); err != nil {
			return ready, err
		}
		pdbSet[pdb.Name] = true
		found := &policy.PodDisruptionBudget{}
		err := r.Get(context.TODO(), types.NamespacedName{Name: pdb.Name, Namespace: pdb.Namespace}, found)
		if err != nil && errors.IsNotFound(err) {
			ready = false
			log.Info("Creating PDB", "namespace", pdb.Namespace, "name", pdb.Name)
			err = r.Create(context.TODO(), pdb)
			if err != nil {
				return ready, err
			}
			r.Recorder.Eventf(instance, corev1.EventTypeNormal, constants.EventsCreatePDB, "Created PodDisruptionBudget %q", pdb.GetName())
		} else if err != nil {
			return ready, err
		} else {
			// Update the found object and write the result back if there are any changes
			if !equality.Semantic.DeepEqual(pdb.Spec, found.Spec) {

				desiredPdb := found.DeepCopy()
				found.Spec = pdb.Spec

				log.Info("Updating PDB", "namespace", pdb.Namespace, "name", pdb.Name)
				err = r.Update(context.TODO(), found)
				if err != nil {
					return ready, err
				}

				// Check if what came back from server modulo the defaults applied by k8s is the same or not
				if !equality.Semantic.DeepEqual(desiredPdb.Spec, found.Spec) {
					ready = false
					r.Recorder.Eventf(instance, corev1.EventTypeNormal, constants.EventsUpdatePDB, "Updated HorizontalPodAutoscaler %q", pdb.GetName())
					//For debugging we will show the difference
					diff, err := kmp.SafeDiff(desiredPdb.Spec, found.Spec)
					if err != nil {
						log.Error(err, "Failed to diff")
					} else {
						log.Info(fmt.Sprintf("Difference in PDBs: %v", diff))
					}
				} else {
					log.Info("The PDBs are the same - api server defaults ignored")
				}

			} else {
				log.Info("Found identical PDB", "namespace", found.Namespace, "name", found.Name, "status", found.Status)
			}
		}

	}

	// For all Deployments check if any PDBs exist and they are not required
	for _, deploy := range components.deployments {
		if _, ok := pdbSet[deploy.Name]; !ok {
			found := &policy.PodDisruptionBudget{}
			err := r.Get(context.TODO(), types.NamespacedName{Name: deploy.Name, Namespace: deploy.Namespace}, found)
			if err != nil {
				if !errors.IsNotFound(err) {
					return false, err
				}
				// Do nothing
			} else {
				// Delete PDB
				log.Info("Deleting pdb", "name", deploy.Name)
				err := r.Delete(context.TODO(), found, client.PropagationPolicy(metav1.DeletePropagationForeground))
				if err != nil {
					return ready, err
				}
			}
		}
	}

	if ready {
		var reason string
		if len(components.pdbs) > 0 {
			reason = machinelearningv1.PdbReadyReason
		} else {
			reason = machinelearningv1.PdbNotDefinedReason
		}
		instance.Status.CreateCondition(machinelearningv1.PdbsReady, true, reason)
	} else {
		instance.Status.CreateCondition(machinelearningv1.PdbsReady, false, machinelearningv1.PdbNotReadyReason)
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

// Create Deployments specified in components, returns ready, progressing, error.
func (r *SeldonDeploymentReconciler) createDeployments(components *components, instance *machinelearningv1.SeldonDeployment, log logr.Logger) (bool, bool, error) {
	ready := true
	progressing := true
	var lastSuccessfulCondition *apis.Condition
	annotator := patch.NewAnnotator(LastAppliedConfig)
	for _, deploy := range components.deployments {
		log.Info("Scheme", "r.scheme", r.Scheme)
		log.Info("createDeployments", "deploy", deploy)
		if err := ctrl.SetControllerReference(instance, deploy, r.Scheme); err != nil {
			return ready, progressing, err
		}

		// TODO(user): Change this for the object type created by your controller
		// Check if the Deployment already exists
		found := &appsv1.Deployment{}
		err := r.Get(context.TODO(), types.NamespacedName{Name: deploy.Name, Namespace: deploy.Namespace}, found)
		if err != nil && errors.IsNotFound(err) {
			ready = false
			if err := annotator.SetLastAppliedAnnotation(deploy); err != nil {
				return ready, progressing, err
			}
			log.Info("Creating Deployment", "namespace", deploy.Namespace, "name", deploy.Name)
			err = r.Create(context.TODO(), deploy)
			if err != nil {
				return ready, progressing, err
			}
			r.Recorder.Eventf(instance, corev1.EventTypeNormal, constants.EventsCreateDeployment, "Created Deployment %q", deploy.GetName())
		} else if err != nil {
			return ready, progressing, err
		} else if !equality.Semantic.DeepEqual(deploy.Spec.Selector, found.Spec.Selector) {
			// Need to recreate deployment as selector has changed
			err = r.Delete(context.TODO(), found, client.PropagationPolicy(metav1.DeletePropagationForeground))
			if err != nil {
				return ready, progressing, err
			}
			r.Recorder.Eventf(instance, corev1.EventTypeNormal, constants.EventsDeleteDeployment, "Deleted Deployment %q", deploy.GetName())
			err = r.Create(context.TODO(), deploy)
			if err != nil {
				return ready, progressing, err
			}
			r.Recorder.Eventf(instance, corev1.EventTypeNormal, constants.EventsCreateDeployment, "Recreated Deployment (selector changed) %q", deploy.GetName())
		} else {
			identical := true
			opts := []patch.CalculateOption{
				patch.IgnoreStatusFields(),
				patch.IgnoreField("kind"),
				patch.IgnoreField("apiVersion"),
				patch.IgnoreField("metadata"),
			}
			// Ignore replica count disparities as that is controlled by HPA/KEDA
			if components.deploymentHasHpa(deploy.Name) || components.deploymentHasKeda(deploy.Name) {
				opts = append(opts, utils2.IgnoreReplicas())
			}
			patcherMaker := patch.NewPatchMaker(annotator, &patch.K8sStrategicMergePatcher{}, &patch.BaseJSONMergePatcher{})
			patchResult, err := patcherMaker.Calculate(found, deploy, opts...)
			if err != nil {
				return ready, progressing, err
			}
			if !patchResult.IsEmpty() {
				log.Info("Updating Deployment", "namespace", deploy.Namespace, "name", deploy.Name)
				log.V(1).Info("Deployment differs", "patch result", patchResult.String())
				b, err := json.Marshal(deploy.Spec.Template.Spec)
				if err == nil {
					log.V(1).Info("Deployment differs", "existing", string(b))
				}
				b2, err := json.Marshal(found.Spec.Template.Spec)
				if err == nil {
					log.V(1).Info("Deployment differs", "found", string(b2))
				}

				desiredDeployment := found.DeepCopy()
				found.Spec = deploy.Spec
				// Add annotations and labels to main metadata
				found.Annotations = deploy.Annotations
				found.Labels = deploy.Labels

				if deploy.Spec.Replicas == nil {
					found.Spec.Replicas = desiredDeployment.Spec.Replicas
				}

				if err := annotator.SetLastAppliedAnnotation(found); err != nil {
					return ready, progressing, err
				}

				err = r.Update(context.TODO(), found)
				if err != nil {
					return ready, progressing, err
				}

				// Check if what came back from server modulo the defaults applied by k8s is the same or not
				if !equality.Semantic.DeepDerivative(desiredDeployment.Spec.Template.Spec, found.Spec.Template.Spec) {
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
				log.Info("Deployment status", "name", found.Name, "status", found.Status)
				if found.Status.ReadyReplicas == 0 || found.Status.UnavailableReplicas > 0 {
					if ready {
						availableCondition := getDeploymentCondition(found, appsv1.DeploymentAvailable)
						log.Info("Updating availableCondition for deployment", "name", found.Name, "availableCondition", availableCondition)
						instance.Status.SetCondition(machinelearningv1.DeploymentsReady, availableCondition)
						log.Info("Inference status", "status", instance.Status)

						progressingCondition := getDeploymentCondition(found, appsv1.DeploymentProgressing)
						if progressingCondition.IsFalse() && availableCondition.IsFalse() {
							log.Info("Deployment is not progressing, returning failed status", "name", found.Name)
							progressing = false
						}
					}
					ready = false
				}

				if ready {
					condition := getDeploymentCondition(found, appsv1.DeploymentAvailable)
					if lastSuccessfulCondition == nil || lastSuccessfulCondition.LastTransitionTime.Inner.Before(&condition.LastTransitionTime.Inner) {
						lastSuccessfulCondition = condition
					}
				}
			}
		}
	}

	if ready {
		instance.Status.SetCondition(machinelearningv1.DeploymentsReady, lastSuccessfulCondition)
	}
	return ready, progressing, nil
}

func getDeploymentCondition(deployment *appsv1.Deployment, conditionType appsv1.DeploymentConditionType) *apis.Condition {
	condition := apis.Condition{}
	for _, con := range deployment.Status.Conditions {
		if con.Type == conditionType {
			condition.Type = apis.ConditionType(conditionType)
			condition.Status = con.Status
			condition.Message = con.Message
			condition.LastTransitionTime = apis.VolatileTime{
				Inner: con.LastTransitionTime,
			}
			condition.Reason = con.Reason
			break
		}
	}
	return &condition
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

	_, err = r.createAmbassadorMappings(components, instance, log)
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

			// Delete any dangling PDBs
			foundPdb := &policy.PodDisruptionBudget{}
			err = r.Get(context.TODO(), types.NamespacedName{Name: found.Name, Namespace: found.Namespace}, foundPdb)
			if err != nil {
				if !errors.IsNotFound(err) {
					return err
				}
				// Do nothing
			} else {
				// Delete PDB that should not exist
				log.Info("Deleting pdb for removed predictor", "name", foundPdb.Name)
				err := r.Delete(context.TODO(), foundPdb, client.PropagationPolicy(metav1.DeletePropagationForeground))
				if err != nil {
					return err
				}
				r.Recorder.Eventf(instance, corev1.EventTypeNormal, constants.EventsDeletePDB, "Deleted PodDisruptionBudget %q", foundPdb.GetName())
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

		// cleanup any orphan services
		seldonId := machinelearningv1.GetSeldonDeploymentName(instance)
		var svcs corev1.ServiceList
		err := r.List(context.TODO(), &svcs, client.MatchingLabels{machinelearningv1.Label_seldon_id: seldonId}, client.InNamespace(instance.Namespace))
		if err != nil {
			return err
		}
		for _, svc := range svcs.Items {
			if _, ok := components.serviceDetails[svc.Name]; !ok {
				log.Info("SVC check: Found orphaned svc so will delete it", "svc", svc.Name)
				err := r.Delete(context.TODO(), &svc)
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
// +kubebuilder:rbac:groups=keda.sh,resources=scaledobjects,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=keda.sh,resources=scaledobjects/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=keda.sh,resources=scaledobjects/finalizers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=policy,resources=poddisruptionbudgets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=policy,resources=poddisruptionbudgets/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=machinelearning.seldon.io,resources=seldondeployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=machinelearning.seldon.io,resources=seldondeployments/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=machinelearning.seldon.io,resources=seldondeployments/finalizers,verbs=get;update;patch
// +kubebuilder:rbac:groups=v1,resources=namespaces,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch
// +kubebuilder:rbac:groups=getambassador.io,resources=mappings,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=getambassador.io,resources=mappings/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=getambassador.io,resources=tlscontexts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=getambassador.io,resources=tlscontexts/status,verbs=get;update;patch

func (r *SeldonDeploymentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	//ctx := context.Background()
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

	// Required for foreground deletion (e.g. ArgoCD does it)
	if !instance.ObjectMeta.DeletionTimestamp.IsZero() {
		// If Deletion Timestamp is set it means object is being deleted.
		// We should take no action in this situation.
		log.Info("Deletion timestamp is set. Doing nothing.")
		return ctrl.Result{}, nil
	}

	// Check if we are not namespaced and should ignore this as it's in a namespace managed by another operator
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

	//run defaulting
	instance.Default()

	components, err := r.createComponents(ctx, instance, podSecurityContext, log)
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

	kedaScaledObjectsReady := false
	withKedaSupport := utils.GetEnv(ENV_KEDA_ENABLED, "false") == "true"
	if withKedaSupport {
		kedaScaledObjectsReady, err = r.createKedaScaledObjects(components, instance, log)
		if err != nil {
			r.Recorder.Eventf(instance, corev1.EventTypeWarning, constants.EventsInternalError, err.Error())
			r.updateStatusForError(instance, err, log)
			return ctrl.Result{}, err
		}
	} else {
		instance.Status.CreateCondition(machinelearningv1.KedaReady, true, machinelearningv1.KedaNotDefinedReason)
	}

	pdbsReady, err := r.createPdbs(components, instance, log)
	if err != nil {
		r.Recorder.Eventf(instance, corev1.EventTypeWarning, constants.EventsInternalError, err.Error())
		r.updateStatusForError(instance, err, log)
		return ctrl.Result{}, err
	}

	deploymentsReady, deploymentsProgressing, err := r.createDeployments(components, instance, log)
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

	switch {
	// Everything is available - happy case.
	case deploymentsReady && servicesReady && hpasReady && pdbsReady && (!withKedaSupport || kedaScaledObjectsReady):
		instance.Status.State = machinelearningv1.StatusStateAvailable
		instance.Status.Description = ""
	// Deployment is not ready and no longer progressing - set status to failed.
	case !deploymentsProgressing && !deploymentsReady:
		instance.Status.State = machinelearningv1.StatusStateFailed
		instance.Status.Description = "Deployment is no longer progressing and not available."
	// Everything else is still creating.
	default:
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
	//Ignore conflict errors
	switch se := err.(type) {
	case *errors.StatusError:
		if se.Status().Reason == metav1.StatusReasonConflict {
			return
		}
	}

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

func (r *SeldonDeploymentReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager, name string) error {

	if err := mgr.GetFieldIndexer().IndexField(ctx, &appsv1.Deployment{}, ownerKey, func(rawObj client.Object) []string {
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

	if err := mgr.GetFieldIndexer().IndexField(ctx, &corev1.Service{}, ownerKey, func(rawObj client.Object) []string {
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

	// Base builder
	builder := ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&machinelearningv1.SeldonDeployment{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{})

	// Add istio CRs
	if utils.GetEnv(ENV_ISTIO_ENABLED, "false") == "true" {
		if err := mgr.GetFieldIndexer().IndexField(ctx, &istio.VirtualService{}, ownerKey, func(rawObj client.Object) []string {
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
		builder.Owns(&istio.VirtualService{})
	}
	// Add Ambassador CRs
	if utils.GetEnv(ENV_AMBASSADOR_ENABLED, "false") == "true" &&
		utils.GetEnv(ENV_AMBASSADOR_VERSION, "v2") == "v2" {
		if err := mgr.GetFieldIndexer().IndexField(ctx, &v2.Mapping{}, ownerKey, func(rawObj client.Object) []string {
			// grab the deployment object, extract the owner...
			mapping := rawObj.(*v2.Mapping)
			owner := metav1.GetControllerOf(mapping)
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
		if err := mgr.GetFieldIndexer().IndexField(ctx, &v2.TLSContext{}, ownerKey, func(rawObj client.Object) []string {
			// grab the deployment object, extract the owner...
			tlsContext := rawObj.(*v2.TLSContext)
			owner := metav1.GetControllerOf(tlsContext)
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
		builder.
			Owns(&v2.Mapping{}).
			Owns(&v2.TLSContext{})
	}
	return builder.Complete(r)
}
