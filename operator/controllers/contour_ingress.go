package controllers

import (
	"bytes"
	"context"
	"fmt"
	"github.com/go-logr/logr"
	contour "github.com/projectcontour/contour/apis/projectcontour/v1"
	machinelearningv1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	v1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	"github.com/seldonio/seldon-core/operator/constants"
	"github.com/seldonio/seldon-core/operator/utils"
	v13 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	types2 "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"knative.dev/pkg/kmp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"text/template"
)

const (
	CONTOUR_INGRESS_CLASS_ANNOTATION = "projectcontour.io/ingress.class"

	ENV_CONTOUR_ENABLED          = "CONTOUR_ENABLED"
	ENV_CONTOUR_INGRESS_CLASS    = "CONTOUR_INGRESS_CLASS"
	ENV_CONTOUR_CONTROLLER_LABEL = "CONTOUR_CONTROLLER_LABEL"

	// Environment variables for use with single virtualhost mode
	ENV_CONTOUR_VIRTUALHOST_NAMESPACE = "CONTOUR_VIRTUALHOST_NAMESPACE"
	ENV_CONTOUR_VIRTUALHOST_NAME      = "CONTOUR_VIRTUALHOST_NAME"
	ENV_CONTOUR_VIRTUALHOST_FQDN      = "CONTOUR_VIRTUALHOST_FQDN"

	// Environment variables for use with per-model virtualhost mode
	ENV_CONTOUR_PER_MODEL_VHOST_ENABLED = "CONTOUR_PER_MODEL_VHOST_ENABLED"
	ENV_CONTOUR_DISABLE_PATH_REWRITE    = "CONTOUR_DISABLE_PATH_REWRITE"
	ENV_CONTOUR_PREDICTOR_FQDN_TEMPLATE = "CONTOUR_PREDICTOR_FQDN_TEMPLATE"
	ENV_CONTOUR_EXPLAINER_FQDN_TEMPLATE = "CONTOUR_EXPLAINER_FQDN_TEMPLATE"

	LabelContourSeldonController = "seldon.io/contour-controller"
)

var (
	envContourControllerLabel       = utils.GetEnv(ENV_CONTOUR_CONTROLLER_LABEL, "seldon")
	envContourExplainerfqdnTemplate = utils.GetEnv(ENV_CONTOUR_EXPLAINER_FQDN_TEMPLATE, "{{.Name}}-explainer.{{.ObjectMeta.Namespace}}")
	envContourIngressClass          = utils.GetEnv(ENV_CONTOUR_INGRESS_CLASS, "")
	envContourPathRewriteEnabled    = utils.GetEnv(ENV_CONTOUR_DISABLE_PATH_REWRITE, "false") == "true"
	envContourPerModelVhostEnabled  = utils.GetEnv(ENV_CONTOUR_PER_MODEL_VHOST_ENABLED, "false") == "true"
	envContourPredictorfqdnTemplate = utils.GetEnv(ENV_CONTOUR_PREDICTOR_FQDN_TEMPLATE, "{{.Name}}.{{.ObjectMeta.Namespace}}")
	envContourVirtualHostFqdn       = utils.GetEnv(ENV_CONTOUR_VIRTUALHOST_FQDN, "seldon.io")
	envContourVirtualHostName       = utils.GetEnv(ENV_CONTOUR_VIRTUALHOST_NAME, "seldon")
	envContourVirtualHostNamespace  = utils.GetEnv(ENV_CONTOUR_VIRTUALHOST_NAMESPACE, "projectcontour")

	contourGrpcProtocol = "h2c"
)

// +kubebuilder:rbac:groups=projectcontour.io,resources=httpproxies,verbs=get;list;watch;create;update;patch;delete

type ContourIngress struct {
	client   client.Client
	recorder record.EventRecorder
	scheme   *runtime.Scheme
}

func NewContourIngress() Ingress {
	return &ContourIngress{}
}

func (i *ContourIngress) AddToScheme(scheme *runtime.Scheme) {
	contour.AddKnownTypes(scheme)
}

func (i *ContourIngress) SetupWithManager(mgr ctrl.Manager) ([]runtime.Object, error) {
	// Store the client, recorder and scheme for use later
	i.client = mgr.GetClient()
	i.recorder = mgr.GetEventRecorderFor(constants.ControllerName)
	i.scheme = mgr.GetScheme()

	// Index HTTPProxy by OwnerReference.Name
	if err := mgr.GetFieldIndexer().IndexField(&contour.HTTPProxy{}, ownerKey, func(rawObj runtime.Object) []string {
		httpProxy := rawObj.(*contour.HTTPProxy)
		owner := metav1.GetControllerOf(httpProxy)
		if owner == nil {
			return nil
		}
		if owner.APIVersion != apiGVStr || owner.Kind != "SeldonDeployment" {
			return nil
		}
		return []string{owner.Name}
	}); err != nil {
		return nil, err
	}
	return []runtime.Object{&contour.HTTPProxy{}}, nil
}

func (i *ContourIngress) GeneratePredictorResources(mlDep *v1.SeldonDeployment, seldonId string, namespace string, ports []httpGrpcPorts, httpAllowed bool, grpcAllowed bool) (map[IngressResourceType][]runtime.Object, error) {
	var virtualHost *contour.VirtualHost
	if envContourPerModelVhostEnabled {
		fqdn, err := templateFqdnContour(envContourPredictorfqdnTemplate, mlDep)
		if err != nil {
			return nil, err
		}
		virtualHost = &contour.VirtualHost{
			Fqdn: fqdn,
		}
	}

	// Set ingress.class if configured
	var annotations map[string]string
	if envContourIngressClass != "" {
		annotations = map[string]string{
			CONTOUR_INGRESS_CLASS_ANNOTATION: envContourIngressClass,
		}
	}

	var httpServices []contour.Service
	var grpcServices []contour.Service
	var routes []contour.Route

	for i, predictor := range mlDep.Spec.Predictors {
		predictorServiceName := v1.GetPredictorKey(mlDep, &predictor)
		httpServices = append(httpServices, contour.Service{
			Name:   predictorServiceName,
			Mirror: predictor.Shadow,
			Weight: int64(predictor.Traffic),
			Port:   ports[i].httpPort,
		})
		grpcServices = append(grpcServices, contour.Service{
			Name:     predictorServiceName,
			Mirror:   predictor.Shadow,
			Weight:   int64(predictor.Traffic),
			Port:     ports[i].grpcPort,
			Protocol: &contourGrpcProtocol,
		})
	}

	if httpAllowed {
		var pathRewritePolicy *contour.PathRewritePolicy
		prefix := "/"
		// If path-rewriting is enabled then override the route prefix and rewrite
		if envContourPathRewriteEnabled {
			prefix = "/seldon/" + namespace + "/" + mlDep.Name + "/"
			pathRewritePolicy = &contour.PathRewritePolicy{ReplacePrefix: []contour.ReplacePrefix{
				{
					Prefix:      prefix,
					Replacement: "/",
				},
			}}
		}
		routes = append(routes, contour.Route{
			Conditions: []contour.Condition{{
				Prefix: prefix,
			}},
			PathRewritePolicy: pathRewritePolicy,
			Services:          httpServices,
		})
	}

	if grpcAllowed {
		routes = append(routes, []contour.Route{{
			Conditions: []contour.Condition{{Prefix: constants.GRPCPathPrefixSeldon}},
			Services:   grpcServices,
		}, {
			Conditions: []contour.Condition{{Prefix: constants.GRPCPathPrefixTensorflow}},
			Services:   grpcServices,
		}}...)
	}

	return map[IngressResourceType][]runtime.Object{
		ContourHTTPProxies: {&contour.HTTPProxy{
			ObjectMeta: v12.ObjectMeta{
				Name:        seldonId,
				Namespace:   namespace,
				Annotations: annotations,
				Labels: map[string]string{
					LabelContourSeldonController: envContourControllerLabel,
				},
			},
			Spec: contour.HTTPProxySpec{
				VirtualHost: virtualHost,
				Routes:      routes,
			},
		}},
	}, nil
}

func templateFqdnContour(templateString string, mlDep *v1.SeldonDeployment) (string, error) {
	tmpl, err := template.New("fqdn").Parse(templateString)
	if err != nil {
		return "", fmt.Errorf("error parsing Contour FQDN template: %s", err)
	}
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, mlDep)
	if err != nil {
		return "", fmt.Errorf("error executing Contour FQDN template: %s", err)
	}
	return buf.String(), nil
}

func (i *ContourIngress) GenerateExplainerResources(pSvcName string, p *v1.PredictorSpec, mlDep *v1.SeldonDeployment, seldonId string, namespace string, engineHttpPort int, engineGrpcPort int) (map[IngressResourceType][]runtime.Object, error) {
	var virtualHost *contour.VirtualHost
	if envContourPerModelVhostEnabled {
		fqdn, err := templateFqdnContour(envContourExplainerfqdnTemplate, mlDep)
		if err != nil {
			return nil, err
		}
		virtualHost = &contour.VirtualHost{
			Fqdn: fqdn,
		}
	}

	// Set ingress.class if configured
	var annotations map[string]string
	if envContourIngressClass != "" {
		annotations = map[string]string{
			CONTOUR_INGRESS_CLASS_ANNOTATION: envContourIngressClass,
		}
	}

	var routes []contour.Route

	if engineHttpPort > 0 {
		var pathRewritePolicy *contour.PathRewritePolicy
		prefix := "/"
		// If path-rewriting is enabled then override the route prefix and rewrite
		if envContourPathRewriteEnabled {
			prefix = "/seldon/" + namespace + "/" + mlDep.GetName() + constants.ExplainerPathSuffix + "/" + p.Name + "/"
			pathRewritePolicy = &contour.PathRewritePolicy{ReplacePrefix: []contour.ReplacePrefix{
				{
					Prefix:      prefix,
					Replacement: "/",
				},
			}}
		}
		routes = append(routes, contour.Route{
			Conditions: []contour.Condition{{
				Prefix: prefix,
			}},
			PathRewritePolicy: pathRewritePolicy,
			Services: []contour.Service{
				{
					Name:   pSvcName,
					Weight: int64(100),
					Port:   engineHttpPort,
				},
			},
		})
	}

	if engineGrpcPort > 0 {
		routes = append(routes, contour.Route{
			Conditions: []contour.Condition{{
				Prefix: "/",
			}},
			Services: []contour.Service{{
				Name:     pSvcName,
				Weight:   int64(100),
				Port:     engineGrpcPort,
				Protocol: &contourGrpcProtocol,
			}},
		})
	}

	return map[IngressResourceType][]runtime.Object{
		ContourHTTPProxies: {&contour.HTTPProxy{
			ObjectMeta: v12.ObjectMeta{
				Name:        pSvcName,
				Namespace:   namespace,
				Annotations: annotations,
				Labels: map[string]string{
					LabelContourSeldonController: envContourControllerLabel,
				},
			},
			Spec: contour.HTTPProxySpec{
				VirtualHost: virtualHost,
				Routes:      routes,
			},
		}},
	}, nil
}

func (i *ContourIngress) CreateResources(resources map[IngressResourceType][]runtime.Object, instance *v1.SeldonDeployment, log logr.Logger) (bool, error) {
	ready := true
	if httpProxies, ok := resources[ContourHTTPProxies]; ok == true {
		for _, s := range httpProxies {
			httpProxy := s.(*contour.HTTPProxy)
			if err := controllerutil.SetControllerReference(instance, httpProxy, i.scheme); err != nil {
				return ready, err
			}
			found := &contour.HTTPProxy{}
			err := i.client.Get(context.TODO(), types2.NamespacedName{Name: httpProxy.Name, Namespace: httpProxy.Namespace}, found)
			if err != nil && errors.IsNotFound(err) {
				ready = false
				log.Info("Creating HTTPProxy", "namespace", httpProxy.Namespace, "name", httpProxy.Name)
				err = i.client.Create(context.TODO(), httpProxy)
				if err != nil {
					return ready, err
				}
				i.recorder.Eventf(instance, v13.EventTypeNormal, constants.EventsCreateHTTPProxy, "Created HTTPProxy %q", httpProxy.GetName())
			} else if err != nil {
				return ready, err
			} else {
				// Update the found object and write the result back if there are any changes
				if !equality.Semantic.DeepEqual(httpProxy.Spec, found.Spec) {
					desiredSvc := found.DeepCopy()
					found.Spec = httpProxy.Spec
					log.Info("Updating HTTPProxy", "namespace", httpProxy.Namespace, "name", httpProxy.Name)
					err = i.client.Update(context.TODO(), found)
					if err != nil {
						return ready, err
					}

					// Check if what came back from server modulo the defaults applied by k8s is the same or not
					if !equality.Semantic.DeepEqual(desiredSvc.Spec, found.Spec) {
						ready = false
						i.recorder.Eventf(instance, v13.EventTypeNormal, constants.EventsUpdateHTTPProxy, "Updated HTTPProxy %q", httpProxy.GetName())
						// For debugging we will show the difference
						diff, err := kmp.SafeDiff(desiredSvc.Spec, found.Spec)
						if err != nil {
							log.Error(err, "Failed to diff")
						} else {
							log.Info(fmt.Sprintf("Difference in HTTPProxy: %v", diff))
						}
					} else {
						log.Info("The HTTPProxy objects are the same - API server defaults ignored")
					}
				} else {
					log.Info("Found identical HTTPProxy", "namespace", found.Namespace, "name", found.Name)
				}
			}
		}

		httpProxyList := &contour.HTTPProxyList{}
		_ = i.client.List(context.Background(), httpProxyList, &client.ListOptions{Namespace: instance.Namespace})
		var deleted []*contour.HTTPProxy
		for _, httpProxy := range httpProxyList.Items {
			for _, ownerRef := range httpProxy.OwnerReferences {
				if ownerRef.Name == instance.Name {
					found := false
					for _, p := range httpProxies {
						expectedHttpProxy := p.(*contour.HTTPProxy)
						if expectedHttpProxy.Name == httpProxy.Name {
							found = true
							break
						}
					}
					if !found {
						log.Info("Will delete HTTPProxy", "name", httpProxy.Name, "namespace", httpProxy.Namespace)
						_ = i.client.Delete(context.Background(), &httpProxy, client.PropagationPolicy(metav1.DeletePropagationForeground))
						deleted = append(deleted, &httpProxy)
					}
				}
			}
		}
		for _, deletedHttpProxy := range deleted {
			i.recorder.Eventf(instance, v13.EventTypeNormal, constants.EventsDeleteHTTPProxy, "Delete HTTPProxy %q", deletedHttpProxy)
		}
	}

	// Reconcile the root HTTPProxy if per-model virtualhosts isn't enabled
	if !envContourPerModelVhostEnabled {
		labelSelector, err := labels.Parse(LabelContourSeldonController + "=" + envContourControllerLabel)
		if err != nil {
			return false, err
		}
		// Get all HTTPProxies managed by this Seldon controller that are not the root HTTPProxy
		httpProxyList := &contour.HTTPProxyList{}
		_ = i.client.List(context.Background(), httpProxyList, &client.ListOptions{Namespace: instance.Namespace, LabelSelector: labelSelector, FieldSelector: fields.OneTermNotEqualSelector("metadata.name", envContourVirtualHostName)})

		// Generate list of included HTTPProxies
		includes := make([]contour.Include, len(httpProxyList.Items), len(httpProxyList.Items))
		for _, childHttpProxy := range httpProxyList.Items {
			includes = append(includes, contour.Include{Name: childHttpProxy.Name, Namespace: childHttpProxy.Namespace})
		}

		// Get the root HTTPProxy
		httpProxy := &contour.HTTPProxy{
			ObjectMeta: v12.ObjectMeta{
				Name:      envContourVirtualHostName,
				Namespace: envContourVirtualHostNamespace,
				Labels: map[string]string{
					LabelContourSeldonController: envContourControllerLabel,
				},
			},
			Spec: contour.HTTPProxySpec{
				VirtualHost: &contour.VirtualHost{
					Fqdn: envContourVirtualHostFqdn,
				},
				Includes: includes,
			},
		}
		found := &contour.HTTPProxy{}
		err = i.client.Get(context.TODO(), types2.NamespacedName{Name: httpProxy.Name, Namespace: httpProxy.Namespace}, found)
		if err != nil && errors.IsNotFound(err) {
			ready = false
			log.Info("Creating HTTPProxy", "namespace", httpProxy.Namespace, "name", httpProxy.Name)
			err = i.client.Create(context.TODO(), httpProxy)
			if err != nil {
				return ready, err
			}
		} else if err != nil {
			return ready, err
		} else {
			// Update the found object and write the result back if there are any changes
			if !equality.Semantic.DeepEqual(httpProxy.Spec, found.Spec) {
				desiredSvc := found.DeepCopy()
				found.Spec = httpProxy.Spec
				log.Info("Updating HTTPProxy", "namespace", httpProxy.Namespace, "name", httpProxy.Name)
				err = i.client.Update(context.TODO(), found)
				if err != nil {
					return ready, err
				}
				// Check if what came back from server modulo the defaults applied by k8s is the same or not
				if !equality.Semantic.DeepEqual(desiredSvc.Spec, found.Spec) {
					ready = false
					// For debugging we will show the difference
					diff, err := kmp.SafeDiff(desiredSvc.Spec, found.Spec)
					if err != nil {
						log.Error(err, "Failed to diff")
					} else {
						log.Info(fmt.Sprintf("Difference in HTTPProxy: %v", diff))
					}
				} else {
					log.Info("The HTTPProxy objects are the same - API server defaults ignored")
				}
			} else {
				log.Info("Found identical HTTPProxy", "namespace", found.Namespace, "name", found.Name)
			}
		}
	}
	return true, nil
}

func (i *ContourIngress) GenerateServiceAnnotations(mlDep *machinelearningv1.SeldonDeployment, p *machinelearningv1.PredictorSpec, serviceName string, engineHttpPort, engineGrpcPort int, isExplainer bool) (map[string]string, error) {
	return nil, nil
}
