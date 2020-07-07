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

package seldon

import (
	autoscalingv2beta2 "k8s.io/api/autoscaling/v2beta1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	Label_seldon_id          = "seldon-deployment-id"
	Label_seldon_app         = "seldon-app"
	Label_seldon_app_svc     = "seldon-app-svc"
	Label_svc_orch           = "seldon-deployment-contains-svcorch"
	Label_app                = "app"
	Label_fluentd            = "fluentd"
	Label_router             = "router"
	Label_combiner           = "combiner"
	Label_model              = "model"
	Label_transformer        = "transformer"
	Label_output_transformer = "output-transformer"
	Label_default            = "default"
	Label_shadow             = "shadow"
	Label_canary             = "canary"
	Label_explainer          = "explainer"
	Label_managed_by         = "app.kubernetes.io/managed-by"
	Label_value_seldon       = "seldon-core"

	PODINFO_VOLUME_NAME     = "seldon-podinfo"
	OLD_PODINFO_VOLUME_NAME = "podinfo"
	PODINFO_VOLUME_PATH     = "/etc/podinfo"

	ENV_PREDICTIVE_UNIT_SERVICE_PORT         = "PREDICTIVE_UNIT_SERVICE_PORT"
	ENV_PREDICTIVE_UNIT_SERVICE_PORT_METRICS = "PREDICTIVE_UNIT_METRICS_SERVICE_PORT"
	ENV_PREDICTIVE_UNIT_METRICS_ENDPOINT     = "PREDICTIVE_UNIT_METRICS_ENDPOINT"
	ENV_PREDICTIVE_UNIT_METRICS_PORT_NAME    = "PREDICTIVE_UNIT_METRICS_PORT_NAME"
	ENV_PREDICTIVE_UNIT_PARAMETERS           = "PREDICTIVE_UNIT_PARAMETERS"
	ENV_PREDICTIVE_UNIT_IMAGE                = "PREDICTIVE_UNIT_IMAGE"
	ENV_PREDICTIVE_UNIT_ID                   = "PREDICTIVE_UNIT_ID"
	ENV_PREDICTOR_ID                         = "PREDICTOR_ID"
	ENV_PREDICTOR_LABELS                     = "PREDICTOR_LABELS"
	ENV_SELDON_DEPLOYMENT_ID                 = "SELDON_DEPLOYMENT_ID"

	ANNOTATION_JAVA_OPTS       = "seldon.io/engine-java-opts"
	ANNOTATION_SEPARATE_ENGINE = "seldon.io/engine-separate-pod"
	ANNOTATION_HEADLESS_SVC    = "seldon.io/headless-svc"
	ANNOTATION_NO_ENGINE       = "seldon.io/no-engine"
	ANNOTATION_CUSTOM_SVC_NAME = "seldon.io/svc-name"
	ANNOTATION_EXECUTOR        = "seldon.io/executor"

	DeploymentNamePrefix = "seldon"
)



func GetPredictiveUnit(pu *PredictiveUnit, name string) *PredictiveUnit {
	if name == pu.Name {
		return pu
	} else {
		for i := 0; i < len(pu.Children); i++ {
			found := GetPredictiveUnit(&pu.Children[i], name)
			if found != nil {
				return found
			}
		}
		return nil
	}
}



// SeldonDeploymentSpec defines the desired state of SeldonDeployment
type SeldonDeploymentSpec struct {
	//Name is Deprecated will be removed in future
	Name        string            `json:"name,omitempty" protobuf:"string,1,opt,name=name"`
	Predictors  []PredictorSpec   `json:"predictors" protobuf:"bytes,2,opt,name=name"`
	OauthKey    string            `json:"oauth_key,omitempty" protobuf:"string,3,opt,name=oauth_key"`
	OauthSecret string            `json:"oauth_secret,omitempty" protobuf:"string,4,opt,name=oauth_secret"`
	Annotations map[string]string `json:"annotations,omitempty" protobuf:"bytes,5,opt,name=annotations"`
	Protocol    Protocol          `json:"protocol,omitempty" protobuf:"bytes,6,opt,name=protocol"`
	Transport   Transport         `json:"transport,omitempty" protobuf:"bytes,7,opt,name=transport"`
	Replicas    *int32            `json:"replicas,omitempty" protobuf:"bytes,8,opt,name=replicas"`
}

type PredictorSpec struct {
	Name            string                  `json:"name" protobuf:"string,1,opt,name=name"`
	Graph           PredictiveUnit          `json:"graph" protobuf:"bytes,2,opt,name=predictiveUnit"`
	ComponentSpecs  []*SeldonPodSpec        `json:"componentSpecs,omitempty" protobuf:"bytes,3,opt,name=componentSpecs"`
	Replicas        *int32                  `json:"replicas,omitempty" protobuf:"string,4,opt,name=replicas"`
	Annotations     map[string]string       `json:"annotations,omitempty" protobuf:"bytes,5,opt,name=annotations"`
	EngineResources v1.ResourceRequirements `json:"engineResources,omitempty" protobuf:"bytes,6,opt,name=engineResources"`
	Labels          map[string]string       `json:"labels,omitempty" protobuf:"bytes,7,opt,name=labels"`
	SvcOrchSpec     SvcOrchSpec             `json:"svcOrchSpec,omitempty" protobuf:"bytes,8,opt,name=svcOrchSpec"`
	Traffic         int32                   `json:"traffic,omitempty" protobuf:"bytes,9,opt,name=traffic"`
	Explainer       *Explainer              `json:"explainer,omitempty" protobuf:"bytes,10,opt,name=explainer"`
	Shadow          bool                    `json:"shadow,omitempty" protobuf:"bytes,11,opt,name=shadow"`
}

type Protocol string

const (
	ProtocolSeldon     Protocol = "seldon"
	ProtocolTensorflow Protocol = "tensorflow"
)

type Transport string

const (
	TransportRest Transport = "rest"
	TransportGrpc Transport = "grpc"
)

type SvcOrchSpec struct {
	Resources *v1.ResourceRequirements `json:"resources,omitempty" protobuf:"bytes,1,opt,name=resources"`
	Env       []*v1.EnvVar             `json:"env,omitempty" protobuf:"bytes,2,opt,name=env"`
	Replicas  *int32                   `json:"replicas,omitempty" protobuf:"bytes,3,opt,name=replicas"`
}

type AlibiExplainerType string

const (
	AlibiAnchorsTabularExplainer  AlibiExplainerType = "AnchorTabular"
	AlibiAnchorsImageExplainer    AlibiExplainerType = "AnchorImages"
	AlibiAnchorsTextExplainer     AlibiExplainerType = "AnchorText"
	AlibiCounterfactualsExplainer AlibiExplainerType = "Counterfactuals"
	AlibiContrastiveExplainer     AlibiExplainerType = "Contrastive"
)

type Explainer struct {
	Type               AlibiExplainerType `json:"type,omitempty" protobuf:"string,1,opt,name=type"`
	ModelUri           string             `json:"modelUri,omitempty" protobuf:"string,2,opt,name=modelUri"`
	ServiceAccountName string             `json:"serviceAccountName,omitempty" protobuf:"string,3,opt,name=serviceAccountName"`
	ContainerSpec      v1.Container       `json:"containerSpec,omitempty" protobuf:"bytes,4,opt,name=containerSpec"`
	Config             map[string]string  `json:"config,omitempty" protobuf:"bytes,5,opt,name=config"`
	Endpoint           *Endpoint          `json:"endpoint,omitempty" protobuf:"bytes,6,opt,name=endpoint"`
	EnvSecretRefName   string             `json:"envSecretRefName,omitempty" protobuf:"bytes,7,opt,name=envSecretRefName"`
}

type SeldonPodSpec struct {
	Metadata metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Spec     v1.PodSpec        `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
	HpaSpec  *SeldonHpaSpec    `json:"hpaSpec,omitempty" protobuf:"bytes,3,opt,name=hpaSpec"`
	Replicas *int32            `json:"replicas,omitempty" protobuf:"bytes,4,opt,name=replicas"`
}

type SeldonHpaSpec struct {
	MinReplicas *int32                          `json:"minReplicas,omitempty" protobuf:"int,1,opt,name=minReplicas"`
	MaxReplicas int32                           `json:"maxReplicas" protobuf:"int,2,opt,name=maxReplicas"`
	Metrics     []autoscalingv2beta2.MetricSpec `json:"metrics,omitempty" protobuf:"bytes,3,opt,name=metrics"`
}

type PredictiveUnitType string

const (
	UNKNOWN_TYPE       PredictiveUnitType = "UNKNOWN_TYPE"
	ROUTER             PredictiveUnitType = "ROUTER"
	COMBINER           PredictiveUnitType = "COMBINER"
	MODEL              PredictiveUnitType = "MODEL"
	TRANSFORMER        PredictiveUnitType = "TRANSFORMER"
	OUTPUT_TRANSFORMER PredictiveUnitType = "OUTPUT_TRANSFORMER"
)

type PredictiveUnitImplementation string

const (
	UNKNOWN_IMPLEMENTATION PredictiveUnitImplementation = "UNKNOWN_IMPLEMENTATION"
	SIMPLE_MODEL           PredictiveUnitImplementation = "SIMPLE_MODEL"
	SIMPLE_ROUTER          PredictiveUnitImplementation = "SIMPLE_ROUTER"
	RANDOM_ABTEST          PredictiveUnitImplementation = "RANDOM_ABTEST"
	AVERAGE_COMBINER       PredictiveUnitImplementation = "AVERAGE_COMBINER"
)

type PredictiveUnitMethod string

const (
	TRANSFORM_INPUT  PredictiveUnitMethod = "TRANSFORM_INPUT"
	TRANSFORM_OUTPUT PredictiveUnitMethod = "TRANSFORM_OUTPUT"
	ROUTE            PredictiveUnitMethod = "ROUTE"
	AGGREGATE        PredictiveUnitMethod = "AGGREGATE"
	SEND_FEEDBACK    PredictiveUnitMethod = "SEND_FEEDBACK"
)

type EndpointType string

const (
	REST EndpointType = "REST"
	GRPC EndpointType = "GRPC"
)

type Endpoint struct {
	ServiceHost string       `json:"service_host,omitempty" protobuf:"string,1,opt,name=service_host"`
	ServicePort int32        `json:"service_port,omitempty" protobuf:"int32,2,opt,name=service_port"`
	Type        EndpointType `json:"type,omitempty" protobuf:"int,3,opt,name=type"`
}

type ParmeterType string

const (
	INT    ParmeterType = "INT"
	FLOAT  ParmeterType = "FLOAT"
	DOUBLE ParmeterType = "DOUBLE"
	STRING ParmeterType = "STRING"
	BOOL   ParmeterType = "BOOL"
)

type Parameter struct {
	Name  string       `json:"name" protobuf:"string,1,opt,name=name"`
	Value string       `json:"value" protobuf:"string,2,opt,name=value"`
	Type  ParmeterType `json:"type" protobuf:"int,3,opt,name=type"`
}

type PredictiveUnit struct {
	Name               string                        `json:"name" protobuf:"string,1,opt,name=name"`
	Children           []PredictiveUnit              `json:"children,omitempty" protobuf:"bytes,2,opt,name=children"`
	Type               *PredictiveUnitType           `json:"type,omitempty" protobuf:"int,3,opt,name=type"`
	Implementation     *PredictiveUnitImplementation `json:"implementation,omitempty" protobuf:"int,4,opt,name=implementation"`
	Methods            *[]PredictiveUnitMethod       `json:"methods,omitempty" protobuf:"int,5,opt,name=methods"`
	Endpoint           *Endpoint                     `json:"endpoint,omitempty" protobuf:"bytes,6,opt,name=endpoint"`
	Parameters         []Parameter                   `json:"parameters,omitempty" protobuf:"bytes,7,opt,name=parameters"`
	ModelURI           string                        `json:"modelUri,omitempty" protobuf:"bytes,8,opt,name=modelUri"`
	ServiceAccountName string                        `json:"serviceAccountName,omitempty" protobuf:"bytes,9,opt,name=serviceAccountName"`
	EnvSecretRefName   string                        `json:"envSecretRefName,omitempty" protobuf:"bytes,10,opt,name=envSecretRefName"`
	// Request/response  payload logging. v2alpha1 feature that is added to v1 for backwards compatibility while v1 is the storage version.
	Logger *Logger `json:"logger,omitempty"`
}

type LoggerMode string

const (
	LogAll      LoggerMode = "all"
	LogRequest  LoggerMode = "request"
	LogResponse LoggerMode = "response"
)

// Logger provides optional payload logging for all endpoints
// +experimental
type Logger struct {
	// URL to send request logging CloudEvents
	// +optional
	Url *string `json:"url,omitempty"`
	// What payloads to log
	Mode LoggerMode `json:"mode,omitempty"`
}

type DeploymentStatus struct {
	Name              string `json:"name,omitempty" protobuf:"string,1,opt,name=name"`
	Status            string `json:"status,omitempty" protobuf:"string,2,opt,name=status"`
	Description       string `json:"description,omitempty" protobuf:"string,3,opt,name=description"`
	Replicas          int32  `json:"replicas,omitempty" protobuf:"string,4,opt,name=replicas"`
	AvailableReplicas int32  `json:"availableReplicas,omitempty" protobuf:"string,5,opt,name=availableRelicas"`
	ExplainerFor      string `json:"explainerFor,omitempty" protobuf:"string,6,opt,name=explainerFor"`
}

type ServiceStatus struct {
	SvcName      string `json:"svcName,omitempty" protobuf:"string,1,opt,name=svcName"`
	HttpEndpoint string `json:"httpEndpoint,omitempty" protobuf:"string,2,opt,name=httpEndpoint"`
	GrpcEndpoint string `json:"grpcEndpoint,omitempty" protobuf:"string,3,opt,name=grpcEndpoint"`
	ExplainerFor string `json:"explainerFor,omitempty" protobuf:"string,4,opt,name=explainerFor"`
}

type StatusState string

// CRD Status values
const (
	StatusStateAvailable StatusState = "Available"
	StatusStateCreating  StatusState = "Creating"
	StatusStateFailed    StatusState = "Failed"
)

// Addressable placeholder until duckv1 issue is fixed:
//    https://github.com/kubernetes-sigs/controller-tools/issues/391
type SeldonAddressable struct {
	URL string `json:"url,omitempty"`
}

// SeldonDeploymentStatus defines the observed state of SeldonDeployment
type SeldonDeploymentStatus struct {
	State            StatusState                 `json:"state,omitempty" protobuf:"string,1,opt,name=state"`
	Description      string                      `json:"description,omitempty" protobuf:"string,2,opt,name=description"`
	DeploymentStatus map[string]DeploymentStatus `json:"deploymentStatus,omitempty" protobuf:"bytes,3,opt,name=deploymentStatus"`
	ServiceStatus    map[string]ServiceStatus    `json:"serviceStatus,omitempty" protobuf:"bytes,4,opt,name=serviceStatus"`
	Replicas         int32                       `json:"replicas,omitempty" protobuf:"string,5,opt,name=replicas"`
	Address          *SeldonAddressable          `json:"address,omitempty"`
}

// SeldonDeployment is the Schema for the seldondeployments API
type SeldonDeployment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SeldonDeploymentSpec   `json:"spec,omitempty"`
	Status SeldonDeploymentStatus `json:"status,omitempty"`
}


// SeldonDeploymentList contains a list of SeldonDeployment
type SeldonDeploymentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SeldonDeployment `json:"items"`
}
