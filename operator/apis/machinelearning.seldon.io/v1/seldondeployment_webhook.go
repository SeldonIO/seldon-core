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

package v1

import (
	"os"

	"github.com/seldonio/seldon-core/operator/constants"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

var (
	// log is for logging in this package.
	seldondeploymentlog                 = logf.Log.WithName("seldondeployment")
	ControllerNamespace                 = GetEnv("POD_NAMESPACE", "seldon-system")
	C                                   client.Client
	envPredictiveUnitHttpServicePort    = os.Getenv(ENV_PREDICTIVE_UNIT_HTTP_SERVICE_PORT)
	envPredictiveUnitGrpcServicePort    = os.Getenv(ENV_PREDICTIVE_UNIT_GRPC_SERVICE_PORT)
	envPredictiveUnitServicePortMetrics = os.Getenv(ENV_PREDICTIVE_UNIT_SERVICE_PORT_METRICS)
	envPredictiveUnitMetricsPortName    = GetEnv(ENV_PREDICTIVE_UNIT_METRICS_PORT_NAME, constants.DefaultMetricsPortName)
)

// Get an environment variable given by key or return the fallback.
func GetEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func (r *SeldonDeployment) SetupWebhookWithManager(mgr ctrl.Manager) error {
	C = mgr.GetClient()
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

func GetContainerForPredictiveUnit(p *PredictorSpec, name string) *corev1.Container {
	for j := 0; j < len(p.ComponentSpecs); j++ {
		cSpec := p.ComponentSpecs[j]
		for k := 0; k < len(cSpec.Spec.Containers); k++ {
			c := &cSpec.Spec.Containers[k]
			if c.Name == name {
				return c
			}
		}
	}
	return nil
}

// --- Validating

// Check the predictive units to ensure the graph matches up with defined containers.
func (r *SeldonDeploymentSpec) checkPredictiveUnits(pu *PredictiveUnit, p *PredictorSpec, fldPath *field.Path, allErrs field.ErrorList) field.ErrorList {

	if pu.Implementation == nil || *pu.Implementation == UNKNOWN_IMPLEMENTATION {

		if GetContainerForPredictiveUnit(p, pu.Name) == nil {
			allErrs = append(allErrs, field.Invalid(fldPath, pu.Name, "Can't find container for Predictive Unit"))
		}

		if pu.Type != nil && *pu.Type == UNKNOWN_TYPE && (pu.Methods == nil || len(*pu.Methods) == 0) {
			allErrs = append(allErrs, field.Invalid(fldPath, pu.Name, "Predictive Unit has no implementation methods defined. Change to a known type or add what methods it defines"))
		}

	} else if IsPrepack(pu) {
		// Only HuggingFace server is allowed for no ModelURI as it can load from Hub
		if pu.ModelURI == "" && (pu.Implementation == nil || *pu.Implementation != PrepackHuggingFaceName) {
			allErrs = append(allErrs, field.Invalid(fldPath, pu.Name, "Predictive unit modelUri required when using standalone servers"))
		}
		c := GetContainerForPredictiveUnit(p, pu.Name)

		//Current non tensorflow serving prepack servers can not handle tensorflow protocol
		if r.Protocol == ProtocolTensorflow && (*pu.Implementation == PrepackSklearnName || *pu.Implementation == PrepackXgboostName || *pu.Implementation == PrepackMlflowName || *pu.Implementation == PrepackHuggingFaceName) {
			allErrs = append(allErrs, field.Invalid(fldPath, pu.Name, "Prepackaged server does not handle tensorflow protocol "+string(*pu.Implementation)))
		}

		if c == nil || c.Image == "" {

			ServersConfigs, err := getPredictorServerConfigs()

			if err != nil {
				seldondeploymentlog.Error(err, "Failed to read prepacked model servers from configmap")
			}

			_, ok := ServersConfigs[string(*pu.Implementation)]
			if !ok {
				allErrs = append(allErrs, field.Invalid(fldPath, pu.Name, "No entry in predictors map for "+string(*pu.Implementation)))
			}
		}
	}

	if pu.Logger != nil {
		if pu.Logger.Mode == "" {
			allErrs = append(allErrs, field.Invalid(fldPath, pu.Logger.Mode, "No logger mode specified"))
		}
	}

	for i := 0; i < len(pu.Children); i++ {
		allErrs = r.checkPredictiveUnits(&pu.Children[i], p, fldPath.Index(i), allErrs)
	}

	return allErrs
}

func checkTraffic(spec *SeldonDeploymentSpec, fldPath *field.Path, allErrs field.ErrorList) field.ErrorList {
	var trafficSum int32 = 0
	var shadows int = 0
	for i := 0; i < len(spec.Predictors); i++ {
		p := spec.Predictors[i]

		if p.Shadow == true {
			shadows += 1
			if shadows > 1 {
				allErrs = append(allErrs, field.Invalid(fldPath, spec.Predictors[i].Name, "Multiple shadows are not allowed"))
			}
			if p.Traffic < 0 || p.Traffic > 100 {
				allErrs = append(allErrs, field.Invalid(fldPath, spec.Predictors[i].Name, "shadow traffic is illegal, the traffic number should be between [0, 100]"))
			}
		} else {
			trafficSum = trafficSum + p.Traffic
		}
	}

	if trafficSum != 100 && (len(spec.Predictors)-shadows) > 1 {
		allErrs = append(allErrs, field.Invalid(fldPath, spec.Predictors[0].Name, "Traffic must sum to 100 for multiple predictors"))
	}
	if trafficSum > 0 && trafficSum < 100 && len(spec.Predictors) == 1 {
		allErrs = append(allErrs, field.Invalid(fldPath, spec.Predictors[0].Name, "Traffic must sum be 100 for a single predictor when set"))
	}

	return allErrs
}

func sizeOfGraph(p *PredictiveUnit) int {
	count := 0
	for _, child := range p.Children {
		count = count + sizeOfGraph(&child)
	}
	return count + 1
}

func collectTransports(pu *PredictiveUnit, transportsFound map[EndpointType]bool) {
	if pu.Endpoint != nil && pu.Endpoint.Type != "" {
		transportsFound[pu.Endpoint.Type] = true
	}
	for _, c := range pu.Children {
		collectTransports(&c, transportsFound)
	}
}

const (
	ENV_KAFKA_BROKER       = "KAFKA_BROKER"
	ENV_KAFKA_INPUT_TOPIC  = "KAFKA_INPUT_TOPIC"
	ENV_KAFKA_OUTPUT_TOPIC = "KAFKA_OUTPUT_TOPIC"
)

func (r *SeldonDeploymentSpec) validateKafka(allErrs field.ErrorList) field.ErrorList {
	if r.ServerType == ServerTypeKafka {
		for i, p := range r.Predictors {
			if len(p.SvcOrchSpec.Env) == 0 {
				fldPath := field.NewPath("spec").Child("predictors").Index(i)
				allErrs = append(allErrs, field.Invalid(fldPath, p.Name, "For kafka please supply svcOrchSpec envs KAFKA_BROKER, KAFKA_INPUT_TOPIC, KAFKA_OUTPUT_TOPIC"))
			} else {
				found := 0
				for _, env := range p.SvcOrchSpec.Env {
					switch env.Name {
					case ENV_KAFKA_BROKER, ENV_KAFKA_INPUT_TOPIC, ENV_KAFKA_OUTPUT_TOPIC:
						found = found + 1
					}
				}
				if found < 3 {
					fldPath := field.NewPath("spec").Child("predictors").Index(i)
					allErrs = append(allErrs, field.Invalid(fldPath, p.Name, "For kafka please supply svcOrchSpec envs KAFKA_BROKER, KAFKA_INPUT_TOPIC, KAFKA_OUTPUT_TOPIC"))
				}
			}
		}
	}
	return allErrs
}

func (r *SeldonDeploymentSpec) validateShadow(allErrs field.ErrorList) field.ErrorList {
	if len(r.Predictors) == 1 && r.Predictors[0].Shadow {
		fldPath := field.NewPath("spec").Child("predictors").Index(0)
		allErrs = append(allErrs, field.Invalid(fldPath, r.Predictors[0].Name, "Shadow can not exist as only predictor"))
	}
	return allErrs
}

func (r *SeldonDeploymentSpec) ValidateSeldonDeployment() error {
	var allErrs field.ErrorList

	switch r.Protocol {
	case "", ProtocolSeldon, ProtocolTensorflow, ProtocolKfserving, ProtocolV2: // do nothing, no error
		seldondeploymentlog.Info("Protocol is valid", "Protocol", r.Protocol)
	default:
		fldPath := field.NewPath("spec")
		allErrs = append(allErrs, field.Invalid(fldPath, r.Protocol, "Invalid protocol"))
	}

	switch r.Transport {
	case "", TransportRest, TransportGrpc: // do nothing, no error
		seldondeploymentlog.Info("Transport is valid", "Transport", r.Transport)
	default:
		fldPath := field.NewPath("spec")
		allErrs = append(allErrs, field.Invalid(fldPath, r.Transport, "Invalid transport"))
	}

	switch r.ServerType {
	case "", ServerTypeRPC, ServerTypeKafka, ServerTypeRabbitMQ: // do nothing, no error
		seldondeploymentlog.Info("Server Type is valid", "ServerType", r.ServerType)
	default:
		fldPath := field.NewPath("spec")
		allErrs = append(allErrs, field.Invalid(fldPath, r.ServerType, "Invalid serverType"))
	}

	// TODO validate rabbitmq
	allErrs = r.validateKafka(allErrs)
	allErrs = r.validateShadow(allErrs)

	transports := make(map[EndpointType]bool)

	if len(r.Predictors) == 0 {
		fldPath := field.NewPath("spec")
		allErrs = append(allErrs, field.Invalid(fldPath, r.Transport, "Graph contains no predictors"))
	}

	predictorNames := make(map[string]bool)
	for i, p := range r.Predictors {

		collectTransports(&p.Graph, transports)

		_, noEngine := p.Annotations[ANNOTATION_NO_ENGINE]
		if noEngine && sizeOfGraph(&p.Graph) > 1 {
			fldPath := field.NewPath("spec").Child("predictors").Index(i)
			allErrs = append(allErrs, field.Invalid(fldPath, p.Name, "Running without engine only valid for single element graphs"))
		}

		if _, present := predictorNames[p.Name]; present {
			fldPath := field.NewPath("spec").Child("predictors").Index(i)
			allErrs = append(allErrs, field.Invalid(fldPath, p.Name, "Duplicate predictor name"))
		}
		predictorNames[p.Name] = true

		allErrs = r.checkPredictiveUnits(&p.Graph, &p, field.NewPath("spec").Child("predictors").Index(i).Child("graph"), allErrs)
	}

	if len(transports) > 1 {
		fldPath := field.NewPath("spec")
		allErrs = append(allErrs, field.Invalid(fldPath, "", "Multiple endpoint.types found - can only have 1 type in graph. Please use spec.transport"))
	} else if len(transports) == 1 && r.Transport != "" {
		for k := range transports {
			if (k == REST && r.Transport != TransportRest) || (k == GRPC && r.Transport != TransportGrpc) {
				fldPath := field.NewPath("spec")
				allErrs = append(allErrs, field.Invalid(fldPath, "", "Mixed transport types found. Remove graph endpoint.types if transport set at deployment level"))
			}
		}
	}

	allErrs = checkTraffic(r, field.NewPath("spec"), allErrs)

	if len(allErrs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(
		schema.GroupKind{Group: "machinelearing.seldon.io", Kind: "SeldonDeployment"},
		r.Name, allErrs)

}

/// ---

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// +kubebuilder:webhook:webhookVersions=v1,verbs=create;update,path=/validate-machinelearning-seldon-io-v1-seldondeployment,mutating=false,failurePolicy=fail,sideEffects=None,admissionReviewVersions=v1;v1beta1,groups=machinelearning.seldon.io,resources=seldondeployments,versions=v1,name=v1.vseldondeployment.kb.io

var _ webhook.Validator = &SeldonDeployment{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *SeldonDeployment) ValidateCreate() error {
	seldondeploymentlog.Info("Validating v1 Webhook called for CREATE", "name", r.Name)
	return r.Spec.ValidateSeldonDeployment()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *SeldonDeployment) ValidateUpdate(old runtime.Object) error {
	seldondeploymentlog.Info("Validating v1 webhook called for UPDATE", "name", r.Name)
	return r.Spec.ValidateSeldonDeployment()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *SeldonDeployment) ValidateDelete() error {
	seldondeploymentlog.Info("Validating v1 webhook called for DELETE", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}
