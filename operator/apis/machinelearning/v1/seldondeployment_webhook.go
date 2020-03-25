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
	"context"
	"encoding/json"
	"fmt"
	"github.com/seldonio/seldon-core/operator/constants"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	k8types "k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"
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
	ControllerConfigMapName             = "seldon-config"
	C                                   client.Client
	envPredictiveUnitServicePort        = os.Getenv(ENV_PREDICTIVE_UNIT_SERVICE_PORT)
	envPredictiveUnitServicePortMetrics = os.Getenv(ENV_PREDICTIVE_UNIT_SERVICE_PORT_METRICS)
)

const PredictorServerConfigMapKeyName = "predictor_servers"

type PredictorImageConfig struct {
	ContainerImage      string `json:"image"`
	DefaultImageVersion string `json:"defaultImageVersion"`
}

type PredictorServerConfig struct {
	Tensorflow      bool                 `json:"tensorflow,omitempty"`
	TensorflowImage string               `json:"tfImage,omitempty"`
	RestConfig      PredictorImageConfig `json:"rest,omitempty"`
	GrpcConfig      PredictorImageConfig `json:"grpc,omitempty"`
}

// Get an environment variable given by key or return the fallback.
func GetEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getPredictorServerConfigs() (map[string]PredictorServerConfig, error) {
	configMap := &corev1.ConfigMap{}

	err := C.Get(context.TODO(), k8types.NamespacedName{Name: ControllerConfigMapName, Namespace: ControllerNamespace}, configMap)

	if err != nil {
		fmt.Println("Failed to find config map " + ControllerConfigMapName)
		fmt.Println(err)
		return map[string]PredictorServerConfig{}, err
	}
	return getPredictorServerConfigsFromMap(configMap)
}

func getPredictorServerConfigsFromMap(configMap *corev1.ConfigMap) (map[string]PredictorServerConfig, error) {
	predictorServerConfig := make(map[string]PredictorServerConfig)
	if predictorConfig, ok := configMap.Data[PredictorServerConfigMapKeyName]; ok {
		err := json.Unmarshal([]byte(predictorConfig), &predictorServerConfig)
		if err != nil {
			panic(fmt.Errorf("Unable to unmarshall %v json string due to %v ", PredictorServerConfigMapKeyName, err))
		}
	}

	return predictorServerConfig, nil
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

func IsPrepack(pu *PredictiveUnit) bool {
	isPrepack := len(*pu.Implementation) > 0 && *pu.Implementation != SIMPLE_MODEL && *pu.Implementation != SIMPLE_ROUTER && *pu.Implementation != RANDOM_ABTEST && *pu.Implementation != AVERAGE_COMBINER && *pu.Implementation != UNKNOWN_IMPLEMENTATION
	return isPrepack
}

func GetPrepackServerConfig(serverName string) PredictorServerConfig {
	ServersConfigs, err := getPredictorServerConfigs()

	if err != nil {
		seldondeploymentlog.Error(err, "Failed to read prepacked model servers from configmap")
	}
	ServerConfig, ok := ServersConfigs[serverName]
	if !ok {
		seldondeploymentlog.Error(nil, "No entry in predictors map for "+serverName)
	}
	return ServerConfig
}

func SetImageNameForPrepackContainer(pu *PredictiveUnit, c *corev1.Container) {
	//Add missing fields
	// Add image
	if c.Image == "" {

		ServerConfig := GetPrepackServerConfig(string(*pu.Implementation))

		if pu.Endpoint.Type == REST {
			c.Image = ServerConfig.RestConfig.ContainerImage + ":" + ServerConfig.RestConfig.DefaultImageVersion
		} else {
			c.Image = ServerConfig.GrpcConfig.ContainerImage + ":" + ServerConfig.GrpcConfig.DefaultImageVersion
		}

	}
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
	existingMetricPort := GetPort(constants.MetricsPortName, con.Ports)
	if existingMetricPort == nil {
		con.Ports = append(con.Ports, corev1.ContainerPort{
			Name:          constants.MetricsPortName,
			ContainerPort: *nextMetricsPortNum,
			Protocol:      corev1.ProtocolTCP,
		})
		*nextMetricsPortNum++
	}
}

func (r *SeldonDeploymentSpec) setContainerPredictiveUnitDefaults(compSpecIdx int,
	portNum int32, nextMetricsPortNum *int32, mldepName string, namespace string,
	p *PredictorSpec, pu *PredictiveUnit, con *corev1.Container) {

	if pu.Endpoint == nil {
		if r.Transport == TransportGrpc {
			pu.Endpoint = &Endpoint{Type: GRPC}
		} else {
			pu.Endpoint = &Endpoint{Type: REST}
		}
	}
	var portType string
	if pu.Endpoint.Type == GRPC {
		portType = constants.GrpcPortName
	} else {
		portType = constants.HttpPortName
	}

	existingPort := GetPort(portType, con.Ports)
	if existingPort != nil {
		portNum = existingPort.ContainerPort
	}

	volFound := false
	for _, vol := range con.VolumeMounts {
		if vol.Name == PODINFO_VOLUME_NAME {
			volFound = true
		}
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
	pu.Endpoint.ServicePort = portNum

}

func (r *SeldonDeploymentSpec) DefaultSeldonDeployment(mldepName string, namespace string) {

	var firstPuPortNum int32 = constants.FirstPortNumber
	if envPredictiveUnitServicePort != "" {
		portNum, err := strconv.Atoi(envPredictiveUnitServicePort)
		if err != nil {
			seldondeploymentlog.Error(err, "Failed to decode predictive unit service port will use default", "envar", ENV_PREDICTIVE_UNIT_SERVICE_PORT, "value", envPredictiveUnitServicePort)
		} else {
			firstPuPortNum = int32(portNum)
		}
	}
	nextPortNum := firstPuPortNum

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
	portMap := map[string]int32{}

	for i := 0; i < len(r.Predictors); i++ {
		p := r.Predictors[i]

		// Add version label for predictor if not present
		if p.Labels == nil {
			p.Labels = map[string]string{}
		}
		if _, present := p.Labels["version"]; !present {
			p.Labels["version"] = p.Name
		}
		addDefaultsToGraph(p.Graph)

		r.Predictors[i] = p

		for j := 0; j < len(p.ComponentSpecs); j++ {
			cSpec := r.Predictors[i].ComponentSpecs[j]

			// add service details for each container - looping this way as if containers in same pod and its the engine pod both need to be localhost
			for k := 0; k < len(cSpec.Spec.Containers); k++ {
				con := &cSpec.Spec.Containers[k]

				getUpdatePortNumMap(con.Name, &nextPortNum, portMap)
				portNum := portMap[con.Name]

				pu := GetPredictiveUnit(p.Graph, con.Name)

				if pu != nil {
					r.setContainerPredictiveUnitDefaults(j, portNum, &nextMetricsPortNum, mldepName, namespace, &p, pu, con)
				}
			}
		}

		pus := GetPredictiveUnitList(p.Graph)

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

				getUpdatePortNumMap(con.Name, &nextPortNum, portMap)
				portNum := portMap[pu.Name]

				r.setContainerPredictiveUnitDefaults(0, portNum, &nextMetricsPortNum, mldepName, namespace, &p, pu, con)
				SetImageNameForPrepackContainer(pu, con)

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
	}
}

// --- Validating

// Check the predictive units to ensure the graph matches up with defined containers.
func checkPredictiveUnits(pu *PredictiveUnit, p *PredictorSpec, fldPath *field.Path, allErrs field.ErrorList) field.ErrorList {
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
		allErrs = checkPredictiveUnits(&pu.Children[i], p, fldPath.Index(i), allErrs)
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

func (r *SeldonDeploymentSpec) ValidateSeldonDeployment() error {
	var allErrs field.ErrorList

	if r.Protocol != "" && !(r.Protocol == ProtocolSeldon || r.Protocol == ProtocolTensorflow) {
		fldPath := field.NewPath("spec")
		allErrs = append(allErrs, field.Invalid(fldPath, r.Protocol, "Invalid protocol"))
	}

	if r.Transport != "" && !(r.Transport == TransportRest || r.Transport == TransportGrpc) {
		fldPath := field.NewPath("spec")
		allErrs = append(allErrs, field.Invalid(fldPath, r.Transport, "Invalid transport"))
	}

	transports := make(map[EndpointType]bool)

	predictorNames := make(map[string]bool)
	for i, p := range r.Predictors {

		collectTransports(p.Graph, transports)

		_, noEngine := p.Annotations[ANNOTATION_NO_ENGINE]
		if noEngine && sizeOfGraph(p.Graph) > 1 {
			fldPath := field.NewPath("spec").Child("predictors").Index(i)
			allErrs = append(allErrs, field.Invalid(fldPath, p.Name, "Running without engine only valid for single element graphs"))
		}

		if _, present := predictorNames[p.Name]; present {
			fldPath := field.NewPath("spec").Child("predictors").Index(i)
			allErrs = append(allErrs, field.Invalid(fldPath, p.Name, "Duplicate predictor name"))
		}
		predictorNames[p.Name] = true
		allErrs = checkPredictiveUnits(p.Graph, &p, field.NewPath("spec").Child("predictors").Index(i).Child("graph"), allErrs)
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
