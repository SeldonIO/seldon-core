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
	"github.com/seldonio/seldon-core/operator/constants"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"log"
	"os"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"strconv"
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

var _ webhook.Defaulter = &SeldonDeployment{}

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

func GetPort(name string, ports []corev1.ContainerPort) *corev1.ContainerPort {
	for i := 0; i < len(ports); i++ {
		if ports[i].Name == name {
			return &ports[i]
		}
	}
	return nil
}

// -----

func addDefaultsToGraph(pu *PredictiveUnit) {
	if pu.Type == nil && pu.Methods == nil && pu.Implementation == nil {
		ty := MODEL
		pu.Type = &ty
	}
	if pu.Implementation == nil {
		im := UNKNOWN_IMPLEMENTATION
		pu.Implementation = &im
	} else if IsPrepack(pu) {
		ty := MODEL
		pu.Type = &ty
	}
	for i := 0; i < len(pu.Children); i++ {
		addDefaultsToGraph(&pu.Children[i])
	}
}

func getUpdatePortNumMap(name string, nextPortNum *int32, portMap map[string]int32) int32 {
	if _, present := portMap[name]; !present {
		portMap[name] = *nextPortNum
		*nextPortNum++
	}
	return portMap[name]
}

func addMetricsPortAndIncrement(nextMetricsPortNum *int32, con *corev1.Container) {
	existingMetricPort := GetPort(envPredictiveUnitMetricsPortName, con.Ports)
	if existingMetricPort == nil {
		con.Ports = append(con.Ports, corev1.ContainerPort{
			Name:          envPredictiveUnitMetricsPortName,
			ContainerPort: *nextMetricsPortNum,
			Protocol:      corev1.ProtocolTCP,
		})
		*nextMetricsPortNum++
	}
}

func (r *SeldonDeploymentSpec) setContainerPredictiveUnitDefaults(compSpecIdx int,
	portNumHttp int32, portNumGrpc int32, nextMetricsPortNum *int32, mldepName string, namespace string,
	p *PredictorSpec, pu *PredictiveUnit, con *corev1.Container) {

	if pu.Endpoint == nil {
		pu.Endpoint = &Endpoint{}
	}

	existingHttpPort := GetPort(constants.HttpPortName, con.Ports)
	if existingHttpPort != nil {
		portNumHttp = existingHttpPort.ContainerPort
	}

	existingGrpcPort := GetPort(constants.GrpcPortName, con.Ports)
	if existingGrpcPort != nil {
		portNumGrpc = existingGrpcPort.ContainerPort
	}

	volFound := false
	for _, vol := range con.VolumeMounts {
		if vol.Name == PODINFO_VOLUME_NAME {
			volFound = true
		}
	}
	//SeldonDeployments first deployed before 1.2 have OLD_PODINFO_VOLUME_NAME
	//they retain that name indefinitely
	oldVolIndex := -1
	for idx, vol := range con.VolumeMounts {
		if vol.Name == OLD_PODINFO_VOLUME_NAME {
			log.Println("found old vol of name " + OLD_PODINFO_VOLUME_NAME)
			oldVolIndex = idx
		}
	}
	if oldVolIndex > -1 {
		con.VolumeMounts[oldVolIndex] = con.VolumeMounts[len(con.VolumeMounts)-1] // Copy last element to index i.
		con.VolumeMounts[len(con.VolumeMounts)-1] = corev1.VolumeMount{}          // Erase last element (write zero value).
		con.VolumeMounts = con.VolumeMounts[:len(con.VolumeMounts)-1]             // Truncate slice.
	}

	if !volFound {
		con.VolumeMounts = append(con.VolumeMounts, corev1.VolumeMount{
			Name:      PODINFO_VOLUME_NAME,
			MountPath: PODINFO_VOLUME_PATH,
		})
	}

	//Add metrics port if missing
	addMetricsPortAndIncrement(nextMetricsPortNum, con)

	// Set ports and hostname in predictive unit so engine can read it from SDep
	// if this is the first componentSpec then it's the one to put the engine in - note using outer loop counter here
	if _, hasSeparateEnginePod := r.Annotations[ANNOTATION_SEPARATE_ENGINE]; compSpecIdx == 0 && !hasSeparateEnginePod {
		pu.Endpoint.ServiceHost = constants.DNSLocalHost
	} else {
		containerServiceValue := GetContainerServiceName(mldepName, *p, con)
		pu.Endpoint.ServiceHost = containerServiceValue + "." + namespace + constants.DNSClusterLocalSuffix
	}
	// deprecated
	pu.Endpoint.ServicePort = portNumHttp

	pu.Endpoint.HttpPort = portNumHttp
	pu.Endpoint.GrpcPort = portNumGrpc

}

func (r *SeldonDeploymentSpec) DefaultSeldonDeployment(mldepName string, namespace string) {

	var firstHttpPuPortNum int32 = constants.FirstHttpPortNumber

	if envPredictiveUnitHttpServicePort != "" {
		portNum, err := strconv.Atoi(envPredictiveUnitHttpServicePort)
		if err != nil {
			seldondeploymentlog.Error(err, "Failed to decode predictive unit service port will use default", "envar", ENV_PREDICTIVE_UNIT_HTTP_SERVICE_PORT, "value", envPredictiveUnitHttpServicePort)
		} else {
			firstHttpPuPortNum = int32(portNum)
		}
	}
	nextHttpPortNum := firstHttpPuPortNum

	var firstGrpcPuPortNum int32 = constants.FirstGrpcPortNumber
	if envPredictiveUnitGrpcServicePort != "" {
		portNum, err := strconv.Atoi(envPredictiveUnitGrpcServicePort)
		if err != nil {
			seldondeploymentlog.Error(err, "Failed to decode grpc predictive unit service port will use default", "envar", ENV_PREDICTIVE_UNIT_GRPC_SERVICE_PORT, "value", envPredictiveUnitGrpcServicePort)
		} else {
			firstGrpcPuPortNum = int32(portNum)
		}
	}
	nextGrpcPortNum := firstGrpcPuPortNum

	var firstMetricsPuPortNum int32 = constants.FirstMetricsPortNumber
	if envPredictiveUnitServicePortMetrics != "" {
		portNum, err := strconv.Atoi(envPredictiveUnitServicePortMetrics)
		if err != nil {
			seldondeploymentlog.Error(err, "Failed to decode PREDICTIVE_UNIT_SERVICE_PORT_METRICS will use default", "value", envPredictiveUnitServicePortMetrics)
		} else {
			firstMetricsPuPortNum = int32(portNum)
		}
	}
	nextMetricsPortNum := firstMetricsPuPortNum
	portMapHttp := map[string]int32{}
	portMapGrpc := map[string]int32{}

	for i := 0; i < len(r.Predictors); i++ {
		p := r.Predictors[i]

		// Add version label for predictor if not present
		if p.Labels == nil {
			p.Labels = map[string]string{}
		}
		if _, present := p.Labels["version"]; !present {
			p.Labels["version"] = p.Name
		}

		addDefaultsToGraph(&p.Graph)

		for j := 0; j < len(p.ComponentSpecs); j++ {
			cSpec := r.Predictors[i].ComponentSpecs[j]

			// add service details for each container - looping this way as if containers in same pod and its the engine pod both need to be localhost
			for k := 0; k < len(cSpec.Spec.Containers); k++ {
				con := &cSpec.Spec.Containers[k]

				getUpdatePortNumMap(con.Name, &nextHttpPortNum, portMapHttp)
				httpPortNum := portMapHttp[con.Name]

				getUpdatePortNumMap(con.Name, &nextGrpcPortNum, portMapGrpc)
				grpcPortNum := portMapGrpc[con.Name]

				pu := GetPredictiveUnit(&p.Graph, con.Name)

				if pu != nil {
					r.setContainerPredictiveUnitDefaults(j, httpPortNum, grpcPortNum, &nextMetricsPortNum, mldepName, namespace, &p, pu, con)
				}
			}
		}

		pus := GetPredictiveUnitList(&p.Graph)

		//some pus might not have a container spec so pick those up
		for l := 0; l < len(pus); l++ {
			pu := pus[l]

			if IsPrepack(pu) {

				con := GetContainerForPredictiveUnit(&p, pu.Name)

				existing := con != nil
				if !existing {
					con = &corev1.Container{
						Name: pu.Name,
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      PODINFO_VOLUME_NAME,
								MountPath: PODINFO_VOLUME_PATH,
							},
						},
					}
				}

				getUpdatePortNumMap(pu.Name, &nextHttpPortNum, portMapHttp)
				httpPortNum := portMapHttp[pu.Name]

				getUpdatePortNumMap(con.Name, &nextGrpcPortNum, portMapGrpc)
				grpcPortNum := portMapGrpc[con.Name]

				r.setContainerPredictiveUnitDefaults(0, httpPortNum, grpcPortNum, &nextMetricsPortNum, mldepName, namespace, &p, pu, con)
				//Only set image default for non tensorflow graphs
				if r.Protocol != ProtocolTensorflow {
					serverConfig := GetPrepackServerConfig(string(*pu.Implementation))
					if serverConfig != nil {
						if con.Image == "" {
							con.Image = serverConfig.PrepackImageName(r.Protocol, pu)
						}
					}
				}

				// if new Add container to componentSpecs
				if !existing {
					if len(p.ComponentSpecs) > 0 {
						p.ComponentSpecs[0].Spec.Containers = append(p.ComponentSpecs[0].Spec.Containers, *con)
					} else {
						podSpec := SeldonPodSpec{
							Metadata: metav1.ObjectMeta{CreationTimestamp: metav1.Now()},
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{*con},
							},
						}
						p.ComponentSpecs = []*SeldonPodSpec{&podSpec}

						// p is a copy so update the entry
						r.Predictors[i] = p
					}
				}
			}
		}

		r.Predictors[i] = p
	}
}

// --- Validating

// Check the predictive units to ensure the graph matches up with defined containers.
func (r *SeldonDeploymentSpec) checkPredictiveUnits(pu *PredictiveUnit, p *PredictorSpec, fldPath *field.Path, allErrs field.ErrorList) field.ErrorList {
	if *pu.Implementation == UNKNOWN_IMPLEMENTATION {

		if GetContainerForPredictiveUnit(p, pu.Name) == nil {
			allErrs = append(allErrs, field.Invalid(fldPath, pu.Name, "Can't find container for Predictive Unit"))
		}

		if *pu.Type == UNKNOWN_TYPE && (pu.Methods == nil || len(*pu.Methods) == 0) {
			allErrs = append(allErrs, field.Invalid(fldPath, pu.Name, "Predictive Unit has no implementation methods defined. Change to a known type or add what methods it defines"))
		}

	} else if IsPrepack(pu) {
		if pu.ModelURI == "" {
			allErrs = append(allErrs, field.Invalid(fldPath, pu.Name, "Predictive unit modelUri required when using standalone servers"))
		}
		c := GetContainerForPredictiveUnit(p, pu.Name)

		//Current non tensorflow serving prepack servers can not handle tensorflow protocol
		if r.Protocol == ProtocolTensorflow && (*pu.Implementation == PrepackSklearnName || *pu.Implementation == PrepackXgboostName || *pu.Implementation == PrepackMlflowName) {
			allErrs = append(allErrs, field.Invalid(fldPath, pu.Name, "Prepackaged server does not handle tendorflow protocol "+string(*pu.Implementation)))
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
		trafficSum = trafficSum + p.Traffic

		if p.Shadow == true {
			shadows += 1
			if shadows > 1 {
				allErrs = append(allErrs, field.Invalid(fldPath, spec.Predictors[i].Name, "Multiple shadows are not allowed"))
			}
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
	if r.ServerType == ServerKafka {
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

	if r.Protocol != "" && !(r.Protocol == ProtocolSeldon || r.Protocol == ProtocolTensorflow || r.Protocol == ProtocolKfserving) {
		fldPath := field.NewPath("spec")
		allErrs = append(allErrs, field.Invalid(fldPath, r.Protocol, "Invalid protocol"))
	}

	if r.Transport != "" && !(r.Transport == TransportRest || r.Transport == TransportGrpc) {
		fldPath := field.NewPath("spec")
		allErrs = append(allErrs, field.Invalid(fldPath, r.Transport, "Invalid transport"))
	}

	if r.ServerType != "" && !(r.ServerType == ServerRPC || r.ServerType == ServerKafka) {
		fldPath := field.NewPath("spec")
		allErrs = append(allErrs, field.Invalid(fldPath, r.ServerType, "Invalid serverType"))
	}

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

// +kubebuilder:webhook:path=/mutate-machinelearning-seldon-io-v1-seldondeployment,mutating=true,failurePolicy=fail,groups=machinelearning.seldon.io,resources=seldondeployments,verbs=create;update,versions=v1,name=v1.mseldondeployment.kb.io

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *SeldonDeployment) Default() {
	seldondeploymentlog.Info("Defaulting v1 web hook called", "name", r.Name)

	if r.ObjectMeta.Namespace == "" {
		r.ObjectMeta.Namespace = "default"
	}
	r.Spec.DefaultSeldonDeployment(r.Name, r.ObjectMeta.Namespace)
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// +kubebuilder:webhook:verbs=create;update,path=/validate-machinelearning-seldon-io-v1-seldondeployment,mutating=false,failurePolicy=fail,groups=machinelearning.seldon.io,resources=seldondeployments,versions=v1,name=v1.vseldondeployment.kb.io

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
