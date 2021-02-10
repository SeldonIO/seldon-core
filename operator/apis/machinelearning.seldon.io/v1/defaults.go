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
	"log"
	"strconv"

	"github.com/seldonio/seldon-core/operator/constants"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetPort(name string, ports []corev1.ContainerPort) *corev1.ContainerPort {
	for i := 0; i < len(ports); i++ {
		if ports[i].Name == name {
			return &ports[i]
		}
	}
	return nil
}

//----

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

	// Backwards compatibility. We set this to grpc port if that is specified otherwise go with http port
	// The executor still uses this port to check for readiness and its needed for backwards compatibility
	// for old images that only have 1 port for http or grpc open
	// TODO: deprecate and remove and fix executor
	if pu.Endpoint.Type == GRPC || r.Transport == TransportGrpc {
		pu.Endpoint.ServicePort = portNumGrpc
	} else {
		pu.Endpoint.ServicePort = portNumHttp
	}

	pu.Endpoint.HttpPort = portNumHttp
	pu.Endpoint.GrpcPort = portNumGrpc
}

func (r *SeldonDeployment) Default() {
	seldondeploymentlog.Info("Defaulting Seldon Deployment called", "name", r.Name)

	if r.ObjectMeta.Namespace == "" {
		r.ObjectMeta.Namespace = "default"
	}
	r.Spec.DefaultSeldonDeployment(r.Name, r.ObjectMeta.Namespace)
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
						creationTime := metav1.Now()
						podSpec := SeldonPodSpec{
							Metadata: ObjectMeta{CreationTimestamp: &creationTime},
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
