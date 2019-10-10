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

package v1alpha2

import (
	"encoding/json"
	"fmt"
	"github.com/seldonio/seldon-core/operator/constants"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"os"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"strconv"
)

// log is for logging in this package.
var seldondeploymentlog = logf.Log.WithName("seldondeployment")

func (r *SeldonDeployment) SetupWebhookWithManager(mgr ctrl.Manager) error {
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
	return *pu.Implementation == SKLEARN_SERVER || *pu.Implementation == XGBOOST_SERVER || *pu.Implementation == TENSORFLOW_SERVER || *pu.Implementation == MLFLOW_SERVER
}

func SetImageNameForPrepackContainer(pu *PredictiveUnit, c *corev1.Container) {
	//Add missing fields
	// Add image
	if c.Image == "" {
		if *pu.Implementation == SKLEARN_SERVER {

			if pu.Endpoint.Type == REST {
				c.Image = constants.DefaultSKLearnServerImageNameRest
			} else {
				c.Image = constants.DefaultSKLearnServerImageNameGrpc
			}

		} else if *pu.Implementation == XGBOOST_SERVER {

			if pu.Endpoint.Type == REST {
				c.Image = constants.DefaultXGBoostServerImageNameRest
			} else {
				c.Image = constants.DefaultXGBoostServerImageNameGrpc
			}

		} else if *pu.Implementation == TENSORFLOW_SERVER {

			if pu.Endpoint.Type == REST {
				c.Image = constants.DefaultTFServerImageNameRest
			} else {
				c.Image = constants.DefaultTFServerImageNameGrpc
			}

		} else if *pu.Implementation == MLFLOW_SERVER {

			if pu.Endpoint.Type == REST {
				c.Image = constants.DefaultMLFlowServerImageNameRest
			} else {
				c.Image = constants.DefaultMLFlowServerImageNameGrpc
			}

		}
	}
}

// -----

func addDefaultsToGraph(pu *PredictiveUnit) {
	if pu.Type == nil {
		ty := UNKNOWN_TYPE
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

func (r *SeldonDeployment) DefaultSeldonDeployment() {

	var firstPuPortNum int32 = 9000
	if env_preditive_unit_service_port, ok := os.LookupEnv("PREDICTIVE_UNIT_SERVICE_PORT"); ok {
		portNum, err := strconv.Atoi(env_preditive_unit_service_port)
		if err != nil {
			seldondeploymentlog.Error(err, "Failed to decode PREDICTIVE_UNIT_SERVICE_PORT will use default 9000", "value", env_preditive_unit_service_port)
		} else {
			firstPuPortNum = int32(portNum)
		}
	}
	nextPortNum := firstPuPortNum

	portMap := map[string]int32{}

	if r.ObjectMeta.Namespace == "" {
		r.ObjectMeta.Namespace = "default"
	}

	for i := 0; i < len(r.Spec.Predictors); i++ {
		p := r.Spec.Predictors[i]
		if p.Graph.Type == nil {
			ty := UNKNOWN_TYPE
			p.Graph.Type = &ty
		}
		// Add version label for predictor if not present
		if p.Labels == nil {
			p.Labels = map[string]string{}
		}
		if _, present := p.Labels["version"]; !present {
			p.Labels["version"] = p.Name
		}
		addDefaultsToGraph(p.Graph)

		fmt.Println("predictor is now")
		jstr, _ := json.Marshal(p)
		fmt.Println(string(jstr))

		r.Spec.Predictors[i] = p

		for j := 0; j < len(p.ComponentSpecs); j++ {
			cSpec := r.Spec.Predictors[i].ComponentSpecs[j]

			// add service details for each container - looping this way as if containers in same pod and its the engine pod both need to be localhost
			for k := 0; k < len(cSpec.Spec.Containers); k++ {
				con := &cSpec.Spec.Containers[k]

				getUpdatePortNumMap(con.Name, &nextPortNum, portMap)

				portNum := portMap[con.Name]

				pu := GetPredictiveUnit(p.Graph, con.Name)

				if pu != nil {

					if pu.Endpoint == nil {
						pu.Endpoint = &Endpoint{Type: REST}
					}
					var portType string
					if pu.Endpoint.Type == GRPC {
						portType = "grpc"
					} else {
						portType = "http"
					}

					if con != nil {
						existingPort := GetPort(portType, con.Ports)
						if existingPort != nil {
							portNum = existingPort.ContainerPort
						}
					}

					// Set ports and hostname in predictive unit so engine can read it from SDep
					// if this is the first componentSpec then it's the one to put the engine in - note using outer loop counter here
					if _, hasSeparateEnginePod := r.Spec.Annotations[ANNOTATION_SEPARATE_ENGINE]; j == 0 && !hasSeparateEnginePod {
						pu.Endpoint.ServiceHost = "localhost"
					} else {
						containerServiceValue := GetContainerServiceName(r, p, con)
						pu.Endpoint.ServiceHost = containerServiceValue + "." + r.ObjectMeta.Namespace + ".svc.cluster.local."
					}
					pu.Endpoint.ServicePort = portNum
				}
			}

			// Add defaultMode to volumes if not set to ensure no changes when comparing later in controller
			for k := 0; k < len(cSpec.Spec.Volumes); k++ {
				vol := &cSpec.Spec.Volumes[k]
				if vol.Secret != nil && vol.Secret.DefaultMode == nil {
					var defaultMode = corev1.SecretVolumeSourceDefaultMode
					vol.Secret.DefaultMode = &defaultMode
				} else if vol.ConfigMap != nil && vol.ConfigMap.DefaultMode == nil {
					var defaultMode = corev1.ConfigMapVolumeSourceDefaultMode
					vol.ConfigMap.DefaultMode = &defaultMode
				} else if vol.DownwardAPI != nil && vol.DownwardAPI.DefaultMode == nil {
					var defaultMode = corev1.DownwardAPIVolumeSourceDefaultMode
					vol.DownwardAPI.DefaultMode = &defaultMode
				} else if vol.Projected != nil && vol.Projected.DefaultMode == nil {
					var defaultMode = corev1.ProjectedVolumeSourceDefaultMode
					vol.Projected.DefaultMode = &defaultMode
				}
			}
		}

		pus := GetPredictiveUnitList(p.Graph)

		//some pus might not have a container spec so pick those up
		for l := 0; l < len(pus); l++ {
			pu := pus[l]

			con := GetContainerForPredictiveUnit(&p, pu.Name)

			// want to set host and port for engine to use in orchestration
			//only assign host and port if there's a container or it's a prepackaged model server
			if !IsPrepack(pu) && (con == nil || con.Name == "") {
				continue
			}

			if _, present := portMap[pu.Name]; !present {
				portMap[pu.Name] = nextPortNum
				nextPortNum++
			}
			portNum := portMap[pu.Name]
			// Add a default REST endpoint if none provided
			// pu needs to have an endpoint as engine reads it from SDep in order to direct graph traffic
			// probes etc will be added later by controller
			if pu.Endpoint == nil {
				pu.Endpoint = &Endpoint{Type: REST}
			}
			var portType string
			if pu.Endpoint.Type == GRPC {
				portType = "grpc"
			} else {
				portType = "http"
			}

			if con != nil {
				existingPort := GetPort(portType, con.Ports)
				if existingPort != nil {
					portNum = existingPort.ContainerPort
				}

				//downward api used to make pod info available to container
				volMount := false
				for _, vol := range con.VolumeMounts {
					if vol.Name == PODINFO_VOLUME_NAME {
						volMount = true
					}
				}
				if !volMount {
					con.VolumeMounts = append(con.VolumeMounts, corev1.VolumeMount{
						Name:      PODINFO_VOLUME_NAME,
						MountPath: PODINFO_VOLUME_PATH,
					})
				}
			}
			// Set ports and hostname in predictive unit so engine can read it from SDep
			// if this is the firstPuPortNum then we've not added engine yet so put the engine in here
			if pu.Endpoint.ServiceHost == "" {
				if _, hasSeparateEnginePod := r.Spec.Annotations[ANNOTATION_SEPARATE_ENGINE]; portNum == firstPuPortNum && !hasSeparateEnginePod {
					pu.Endpoint.ServiceHost = "localhost"
				} else {
					containerServiceValue := GetContainerServiceName(r, p, con)
					pu.Endpoint.ServiceHost = containerServiceValue + "." + r.ObjectMeta.Namespace + ".svc.cluster.local."
				}
			}
			if pu.Endpoint.ServicePort == 0 {
				pu.Endpoint.ServicePort = portNum
			}

			// for prepack servers we want to add a container name and image to correspond to grafana dashboards
			if IsPrepack(pu) {

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
						r.Spec.Predictors[i] = p
					}
				}
			}

		}

	}

}

// -----

// --- Validating

// Check the predictive units to ensure the graph matches up with defined containers.
func checkPredictiveUnits(pu *PredictiveUnit, p *PredictorSpec, fldPath *field.Path, allErrs field.ErrorList) field.ErrorList {
	if *pu.Implementation == UNKNOWN_IMPLEMENTATION {

		if GetContainerForPredictiveUnit(p, pu.Name) == nil {
			allErrs = append(allErrs, field.Invalid(fldPath, pu.Name, "Can't find container for Predictive Unit"))
		}

		if *pu.Type == UNKNOWN_TYPE && (pu.Methods == nil || len(*pu.Methods) == 0) {
			allErrs = append(allErrs, field.Invalid(fldPath, pu.Name, "Predictive Unit has no implementation methods defined. Change to a know type or add what methods it defines"))
		}

	} else if *pu.Implementation == SKLEARN_SERVER ||
		*pu.Implementation == XGBOOST_SERVER ||
		*pu.Implementation == TENSORFLOW_SERVER ||
		*pu.Implementation == MLFLOW_SERVER {
		if pu.ModelURI == "" {
			allErrs = append(allErrs, field.Invalid(fldPath, pu.Name, "Predictive unit modelUri required when using standalone servers"))
		}
	}

	for i := 0; i < len(pu.Children); i++ {
		allErrs = checkPredictiveUnits(&pu.Children[i], p, fldPath.Index(i), allErrs)
	}

	return allErrs
}

func checkTraffic(mlDep *SeldonDeployment, fldPath *field.Path, allErrs field.ErrorList) field.ErrorList {
	var trafficSum int32 = 0
	for i := 0; i < len(mlDep.Spec.Predictors); i++ {
		p := mlDep.Spec.Predictors[i]
		trafficSum = trafficSum + p.Traffic
	}
	if trafficSum != 100 && len(mlDep.Spec.Predictors) > 1 {
		allErrs = append(allErrs, field.Invalid(fldPath, mlDep.Name, "Traffic must sum to 100 for multiple predictors"))
	}
	if trafficSum > 0 && trafficSum < 100 && len(mlDep.Spec.Predictors) == 1 {
		allErrs = append(allErrs, field.Invalid(fldPath, mlDep.Name, "Traffic must sum be 100 for a single predictor when set"))
	}

	return allErrs
}

func (r *SeldonDeployment) validateSeldonDeployment() error {
	var allErrs field.ErrorList

	predictorNames := make(map[string]bool)
	for i, p := range r.Spec.Predictors {
		if _, present := predictorNames[p.Name]; present {
			fldPath := field.NewPath("spec").Child("predictors").Index(i)
			allErrs = append(allErrs, field.Invalid(fldPath, p.Name, "Duplicate predictor name"))
		}
		predictorNames[p.Name] = true
		allErrs = checkPredictiveUnits(p.Graph, &p, field.NewPath("spec").Child("predictors").Index(i).Child("graph"), allErrs)
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

// +kubebuilder:webhook:path=/mutate-machinelearning-seldon-io-v1alpha2-seldondeployment,mutating=true,failurePolicy=fail,groups=machinelearning.seldon.io,resources=seldondeployments,verbs=create;update,versions=v1alpha2,name=mseldondeployment.kb.io

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *SeldonDeployment) Default() {
	seldondeploymentlog.Info("default", "name", r.Name)

	r.DefaultSeldonDeployment()
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// +kubebuilder:webhook:verbs=create;update,path=/validate-machinelearning-seldon-io-v1alpha2-seldondeployment,mutating=false,failurePolicy=fail,groups=machinelearning.seldon.io,resources=seldondeployments,versions=v1alpha2,name=vseldondeployment.kb.io

var _ webhook.Validator = &SeldonDeployment{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *SeldonDeployment) ValidateCreate() error {
	seldondeploymentlog.Info("validate create", "name", r.Name)

	return r.validateSeldonDeployment()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *SeldonDeployment) ValidateUpdate(old runtime.Object) error {
	seldondeploymentlog.Info("validate update", "name", r.Name)

	return r.validateSeldonDeployment()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *SeldonDeployment) ValidateDelete() error {
	seldondeploymentlog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}
