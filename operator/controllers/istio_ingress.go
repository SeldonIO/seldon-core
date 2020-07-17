package controllers

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/gogo/protobuf/types"
	"github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	"github.com/seldonio/seldon-core/operator/constants"
	"github.com/seldonio/seldon-core/operator/utils"
	v1alpha32 "istio.io/api/networking/v1alpha3"
	"istio.io/client-go/pkg/apis/networking/v1alpha3"
	v13 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	types2 "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"knative.dev/pkg/kmp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"strconv"
	"strings"

	machinelearningv1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	istio "istio.io/client-go/pkg/apis/networking/v1alpha3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ENV_ISTIO_ENABLED                = "ISTIO_ENABLED"
	ENV_ISTIO_GATEWAY                = "ISTIO_GATEWAY"
	ENV_ISTIO_TLS_MODE               = "ISTIO_TLS_MODE"
	ANNOTATION_ISTIO_GATEWAY         = "seldon.io/istio-gateway"
	ANNOTATION_ISTIO_RETRIES         = "seldon.io/istio-retries"
	ANNOTATION_ISTIO_RETRIES_TIMEOUT = "seldon.io/istio-retries-timeout"
)

// +kubebuilder:rbac:groups=networking.istio.io,resources=virtualservices,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=networking.istio.io,resources=virtualservices/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=networking.istio.io,resources=destinationrules,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=networking.istio.io,resources=destinationrules/status,verbs=get;update;patch

type IstioIngress struct {
	client   client.Client
	recorder record.EventRecorder
	scheme   *runtime.Scheme
}

func NewIstioIngress() Ingress {
	return &IstioIngress{}
}

func (i *IstioIngress) AddToScheme(scheme *runtime.Scheme) {
	_ = istio.AddToScheme(scheme)
}

func (i *IstioIngress) SetupWithManager(mgr ctrl.Manager, namespace string) ([]runtime.Object, error) {
	// Store the client, recorder and scheme for use later
	i.client = mgr.GetClient()
	i.recorder = mgr.GetEventRecorderFor(constants.ControllerName)
	i.scheme = mgr.GetScheme()

	// Index on VirtualService
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
		return nil, err
	}
	return []runtime.Object{&istio.VirtualService{}}, nil
}

func (i *IstioIngress) GeneratePredictorResources(mlDep *machinelearningv1.SeldonDeployment, seldonId string, namespace string, ports []httpGrpcPorts, httpAllowed bool, grpcAllowed bool) (map[IngressResourceType][]runtime.Object, error) {

	istioGateway := utils.GetEnv(ENV_ISTIO_GATEWAY, "seldon-gateway")
	istioTLSMode := utils.GetEnv(ENV_ISTIO_TLS_MODE, "")

	istioRetriesAnnotation := getAnnotation(mlDep, ANNOTATION_ISTIO_RETRIES, "")
	istioRetriesTimeoutAnnotation := getAnnotation(mlDep, ANNOTATION_ISTIO_RETRIES_TIMEOUT, "1")
	istioRetries := 0
	istioRetriesTimeout := 1
	var err error

	if istioRetriesAnnotation != "" {
		istioRetries, err = strconv.Atoi(istioRetriesAnnotation)
		if err != nil {
			return nil, err
		}
		istioRetriesTimeout, err = strconv.Atoi(istioRetriesTimeoutAnnotation)
		if err != nil {
			return nil, err
		}
	}
	httpVsvc := &v1alpha3.VirtualService{
		ObjectMeta: v12.ObjectMeta{
			Name:      seldonId + "-http",
			Namespace: namespace,
		},
		Spec: v1alpha32.VirtualService{
			Hosts:    []string{"*"},
			Gateways: []string{getAnnotation(mlDep, ANNOTATION_ISTIO_GATEWAY, istioGateway)},
			Http: []*v1alpha32.HTTPRoute{
				{
					Match: []*v1alpha32.HTTPMatchRequest{
						{
							Uri: &v1alpha32.StringMatch{MatchType: &v1alpha32.StringMatch_Prefix{Prefix: "/seldon/" + namespace + "/" + mlDep.Name + "/"}},
						},
					},
					Rewrite: &v1alpha32.HTTPRewrite{Uri: "/"},
				},
			},
		},
	}

	grpcVsvc := &v1alpha3.VirtualService{
		ObjectMeta: v12.ObjectMeta{
			Name:      seldonId + "-grpc",
			Namespace: namespace,
		},
		Spec: v1alpha32.VirtualService{
			Hosts:    []string{"*"},
			Gateways: []string{getAnnotation(mlDep, ANNOTATION_ISTIO_GATEWAY, istioGateway)},
			Http: []*v1alpha32.HTTPRoute{
				{
					Match: []*v1alpha32.HTTPMatchRequest{
						{
							Uri: &v1alpha32.StringMatch{MatchType: &v1alpha32.StringMatch_Regex{Regex: constants.GRPCRegExMatchIstio}},
							Headers: map[string]*v1alpha32.StringMatch{
								"seldon":    {MatchType: &v1alpha32.StringMatch_Exact{Exact: mlDep.Name}},
								"namespace": {MatchType: &v1alpha32.StringMatch_Exact{Exact: namespace}},
							},
						},
					},
				},
			},
		},
	}
	// Add retries
	if istioRetries > 0 {
		httpVsvc.Spec.Http[0].Retries = &v1alpha32.HTTPRetry{Attempts: int32(istioRetries), PerTryTimeout: &types.Duration{Seconds: int64(istioRetriesTimeout)}, RetryOn: "gateway-error,connect-failure,refused-stream"}
		grpcVsvc.Spec.Http[0].Retries = &v1alpha32.HTTPRetry{Attempts: int32(istioRetries), PerTryTimeout: &types.Duration{Seconds: int64(istioRetriesTimeout)}, RetryOn: "gateway-error,connect-failure,refused-stream"}
	}

	// shadows don't get destinations in the vs as a shadow is a mirror instead
	var shadows = 0
	for i := 0; i < len(mlDep.Spec.Predictors); i++ {
		p := mlDep.Spec.Predictors[i]
		if p.Shadow == true {
			shadows += 1
		}
	}

	routesHttp := make([]*v1alpha32.HTTPRouteDestination, len(mlDep.Spec.Predictors)-shadows)
	routesGrpc := make([]*v1alpha32.HTTPRouteDestination, len(mlDep.Spec.Predictors)-shadows)

	// the shadow/mirror entry does need a DestinationRule though
	drules := make([]runtime.Object, len(mlDep.Spec.Predictors))
	routesIdx := 0
	for i := 0; i < len(mlDep.Spec.Predictors); i++ {

		p := mlDep.Spec.Predictors[i]
		pSvcName := v1.GetPredictorKey(mlDep, &p)

		drule := &v1alpha3.DestinationRule{
			ObjectMeta: v12.ObjectMeta{
				Name:      pSvcName,
				Namespace: namespace,
			},
			Spec: v1alpha32.DestinationRule{
				Host: pSvcName,
				Subsets: []*v1alpha32.Subset{
					{
						Name: p.Name,
						Labels: map[string]string{
							"version": p.Labels["version"],
						},
					},
				},
			},
		}

		if istioTLSMode != "" {
			drule.Spec.TrafficPolicy = &v1alpha32.TrafficPolicy{
				Tls: &v1alpha32.TLSSettings{
					Mode: v1alpha32.TLSSettings_TLSmode(v1alpha32.TLSSettings_TLSmode_value[istioTLSMode]),
				},
			}
		}
		drules[i] = drule

		if p.Shadow == true {
			//if there's a shadow then add a mirror section to the VirtualService

			httpVsvc.Spec.Http[0].Mirror = &v1alpha32.Destination{
				Host:   pSvcName,
				Subset: p.Name,
				Port: &v1alpha32.PortSelector{
					Number: uint32(ports[i].httpPort),
				},
			}

			grpcVsvc.Spec.Http[0].Mirror = &v1alpha32.Destination{
				Host:   pSvcName,
				Subset: p.Name,
				Port: &v1alpha32.PortSelector{
					Number: uint32(ports[i].grpcPort),
				},
			}

			continue
		}

		// We split by adding different routes with their own Weights
		// So not by tag - different destinations (like https://istio.io/docs/tasks/traffic-management/traffic-shifting/) distinguished by host
		routesHttp[routesIdx] = &v1alpha32.HTTPRouteDestination{
			Destination: &v1alpha32.Destination{
				Host:   pSvcName,
				Subset: p.Name,
				Port: &v1alpha32.PortSelector{
					Number: uint32(ports[i].httpPort),
				},
			},
			Weight: p.Traffic,
		}
		routesGrpc[routesIdx] = &v1alpha32.HTTPRouteDestination{
			Destination: &v1alpha32.Destination{
				Host:   pSvcName,
				Subset: p.Name,
				Port: &v1alpha32.PortSelector{
					Number: uint32(ports[i].grpcPort),
				},
			},
			Weight: p.Traffic,
		}
		routesIdx += 1

	}
	httpVsvc.Spec.Http[0].Route = routesHttp
	grpcVsvc.Spec.Http[0].Route = routesGrpc

	resources := map[IngressResourceType][]runtime.Object{
		IstioDestinationRules: drules,
	}
	if httpAllowed && grpcAllowed {
		resources[IstioVirtualServices] = []runtime.Object{httpVsvc, grpcVsvc}
	} else if httpAllowed {
		resources[IstioVirtualServices] = []runtime.Object{httpVsvc}
	} else {
		resources[IstioVirtualServices] = []runtime.Object{grpcVsvc}
	}
	return resources, nil
}

func (i *IstioIngress) GenerateExplainerResources(pSvcName string, spec *machinelearningv1.PredictorSpec, mlDep *machinelearningv1.SeldonDeployment, seldonId string, namespace string, engineHttpPort int, engineGrpcPort int) (map[IngressResourceType][]runtime.Object, error) {
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

	istioGateway := utils.GetEnv(ENV_ISTIO_GATEWAY, "seldon-gateway")
	httpVsvc := &istio.VirtualService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      vsNameHttp,
			Namespace: namespace,
		},
		Spec: v1alpha32.VirtualService{
			Hosts:    []string{"*"},
			Gateways: []string{getAnnotation(mlDep, ANNOTATION_ISTIO_GATEWAY, istioGateway)},
			Http: []*v1alpha32.HTTPRoute{
				{
					Match: []*v1alpha32.HTTPMatchRequest{
						{
							Uri: &v1alpha32.StringMatch{MatchType: &v1alpha32.StringMatch_Prefix{Prefix: "/seldon/" + namespace + "/" + mlDep.GetName() + constants.ExplainerPathSuffix + "/" + spec.Name + "/"}},
						},
					},
					Rewrite: &v1alpha32.HTTPRewrite{Uri: "/"},
				},
			},
		},
	}

	grpcVsvc := &istio.VirtualService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      vsNameGrpc,
			Namespace: namespace,
		},
		Spec: v1alpha32.VirtualService{
			Hosts:    []string{"*"},
			Gateways: []string{getAnnotation(mlDep, ANNOTATION_ISTIO_GATEWAY, istioGateway)},
			Http: []*v1alpha32.HTTPRoute{
				{
					Match: []*v1alpha32.HTTPMatchRequest{
						{
							Uri: &v1alpha32.StringMatch{MatchType: &v1alpha32.StringMatch_Prefix{Prefix: "/seldon.protos.Seldon/"}},
							Headers: map[string]*v1alpha32.StringMatch{
								"seldon":    {MatchType: &v1alpha32.StringMatch_Exact{Exact: mlDep.GetName()}},
								"namespace": {MatchType: &v1alpha32.StringMatch_Exact{Exact: namespace}},
							},
						},
					},
				},
			},
		},
	}

	routesHttp := make([]*v1alpha32.HTTPRouteDestination, 1)
	routesGrpc := make([]*v1alpha32.HTTPRouteDestination, 1)
	drules := make([]runtime.Object, 1)

	drule := &istio.DestinationRule{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pSvcName,
			Namespace: namespace,
		},
		Spec: v1alpha32.DestinationRule{
			Host: pSvcName,
			Subsets: []*v1alpha32.Subset{
				{
					Name: spec.Name,
					Labels: map[string]string{
						"version": spec.Labels["version"],
					},
				},
			},
		},
	}

	routesHttp[0] = &v1alpha32.HTTPRouteDestination{
		Destination: &v1alpha32.Destination{
			Host:   pSvcName,
			Subset: spec.Name,
			Port: &v1alpha32.PortSelector{
				Number: uint32(engineHttpPort),
			},
		},
		Weight: int32(100),
	}
	routesGrpc[0] = &v1alpha32.HTTPRouteDestination{
		Destination: &v1alpha32.Destination{
			Host:   pSvcName,
			Subset: spec.Name,
			Port: &v1alpha32.PortSelector{
				Number: uint32(engineGrpcPort),
			},
		},
		Weight: int32(100),
	}
	drules[0] = drule

	httpVsvc.Spec.Http[0].Route = routesHttp
	grpcVsvc.Spec.Http[0].Route = routesGrpc
	vsvcs := make([]runtime.Object, 0, 2)

	// Explainer may not expose REST and grpc (presumably engine ensures predictors do?)
	if engineHttpPort > 0 {
		vsvcs = append(vsvcs, httpVsvc)
	}
	if engineGrpcPort > 0 {
		vsvcs = append(vsvcs, grpcVsvc)
	}

	resources := map[IngressResourceType][]runtime.Object{
		IstioDestinationRules: drules,
		IstioVirtualServices:  vsvcs,
	}
	return resources, nil
}

func (i *IstioIngress) CreateResources(resources map[IngressResourceType][]runtime.Object, instance *machinelearningv1.SeldonDeployment, log logr.Logger) (bool, error) {
	ready := true
	if virtualServices, ok := resources[IstioVirtualServices]; ok == true {
		for _, s := range virtualServices {
			svc := s.(*v1alpha3.VirtualService)
			if err := controllerutil.SetControllerReference(instance, svc, i.scheme); err != nil {
				return ready, err
			}
			found := &v1alpha3.VirtualService{}
			err := i.client.Get(context.TODO(), types2.NamespacedName{Name: svc.Name, Namespace: svc.Namespace}, found)
			if err != nil && errors.IsNotFound(err) {
				ready = false
				log.Info("Creating Virtual Service", "namespace", svc.Namespace, "name", svc.Name)
				err = i.client.Create(context.TODO(), svc)
				if err != nil {
					return ready, err
				}
				i.recorder.Eventf(instance, v13.EventTypeNormal, constants.EventsCreateVirtualService, "Created VirtualService %q", svc.GetName())
			} else if err != nil {
				return ready, err
			} else {
				// Update the found object and write the result back if there are any changes
				if !equality.Semantic.DeepEqual(svc.Spec, found.Spec) {
					desiredSvc := found.DeepCopy()
					found.Spec = svc.Spec
					log.Info("Updating Virtual Service", "namespace", svc.Namespace, "name", svc.Name)
					err = i.client.Update(context.TODO(), found)
					if err != nil {
						return ready, err
					}

					// Check if what came back from server modulo the defaults applied by k8s is the same or not
					if !equality.Semantic.DeepEqual(desiredSvc.Spec, found.Spec) {
						ready = false
						i.recorder.Eventf(instance, v13.EventTypeNormal, constants.EventsUpdateVirtualService, "Updated VirtualService %q", svc.GetName())
						// For debugging we will show the difference
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

		// Cleanup unused VirtualService. This should usually only happen on Operator upgrades where there is a breaking change to the names of the VirtualServices created
		if len(virtualServices) > 0 && ready {
			b, deleted, err2 := cleanupVirtualServices(i.client, i.recorder, instance, virtualServices, log)
			log.Info("Deleted VirtualServices", "deleted", deleted)
			if err2 != nil {
				return b, err2
			}
		}
	}

	if destinationRules, ok := resources[IstioDestinationRules]; ok == true {
		for _, d := range destinationRules {
			drule := d.(*istio.DestinationRule)
			if err := controllerutil.SetControllerReference(instance, drule, i.scheme); err != nil {
				return ready, err
			}
			found := &v1alpha3.DestinationRule{}
			err := i.client.Get(context.TODO(), types2.NamespacedName{Name: drule.Name, Namespace: drule.Namespace}, found)
			if err != nil && errors.IsNotFound(err) {
				ready = false
				log.Info("Creating Istio Destination Rule", "namespace", drule.Namespace, "name", drule.Name)
				err = i.client.Create(context.TODO(), drule)
				if err != nil {
					return ready, err
				}
				i.recorder.Eventf(instance, v13.EventTypeNormal, constants.EventsCreateDestinationRule, "Created DestinationRule %q", drule.GetName())
			} else if err != nil {
				return ready, err
			} else {
				// Update the found object and write the result back if there are any changes
				if !equality.Semantic.DeepEqual(drule.Spec, found.Spec) {
					desiredDrule := found.DeepCopy()
					found.Spec = drule.Spec
					log.Info("Updating Istio Destination Rule", "namespace", drule.Namespace, "name", drule.Name)
					err = i.client.Update(context.TODO(), found)
					if err != nil {
						return ready, err
					}

					// Check if what came back from server modulo the defaults applied by k8s is the same or not
					if !equality.Semantic.DeepEqual(desiredDrule.Spec, found.Spec) {
						ready = false
						i.recorder.Eventf(instance, v13.EventTypeNormal, constants.EventsUpdateDestinationRule, "Updated DestinationRule %q", drule.GetName())
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
	}

	return ready, nil
}

func cleanupVirtualServices(k8sClient client.Client, recorder record.EventRecorder, instance *machinelearningv1.SeldonDeployment, virtualServices []runtime.Object, log logr.Logger) (bool, []*istio.VirtualService, error) {
	// First gather existing list of virtual services
	var deleted []*istio.VirtualService
	vlist := &istio.VirtualServiceList{}
	err := k8sClient.List(context.Background(), vlist, &client.ListOptions{Namespace: instance.Namespace})
	if err != nil {
		return false, nil, err
	}
	for _, vsvc := range vlist.Items {
		for _, ownerRef := range vsvc.OwnerReferences {
			if ownerRef.Name == instance.Name {
				found := false
				for _, expectedVsvc := range virtualServices {
					esvc := expectedVsvc.(*v1alpha3.VirtualService)
					if esvc.Name == vsvc.Name {
						found = true
						break
					}
				}
				if !found {
					log.Info("Will delete VirtualService", "name", vsvc.Name, "namespace", vsvc.Namespace)
					_ = k8sClient.Delete(context.Background(), &vsvc, client.PropagationPolicy(metav1.DeletePropagationForeground))
					deleted = append(deleted, vsvc.DeepCopy())
				}
			}
		}
	}
	for _, vsvcDeleted := range deleted {
		recorder.Eventf(instance, v13.EventTypeNormal, constants.EventsDeleteVirtualService, "Delete VirtualService %q", vsvcDeleted.GetName())
	}
	return true, deleted, nil
}

// Istio plugin doesn't set any annotations on the service itself
func (i *IstioIngress) GenerateServiceAnnotations(mlDep *machinelearningv1.SeldonDeployment, p *machinelearningv1.PredictorSpec, serviceName string, engineHttpPort, engineGrpcPort int, isExplainer bool) (map[string]string, error) {
	return nil, nil
}
