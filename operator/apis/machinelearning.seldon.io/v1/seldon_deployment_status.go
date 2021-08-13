package v1

import (
	v1 "k8s.io/api/core/v1"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
)

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
	duckv1.Status    `json:",inline"`
}

const (
	DeploymentsReady apis.ConditionType = "DeploymentsReady"
	ServicesReady    apis.ConditionType = "ServicesReady"
)

// InferenceService Ready condition is depending on predictor and route readiness condition
var conditionSet = apis.NewLivingConditionSet(
	DeploymentsReady,
	ServicesReady,
)

var _ apis.ConditionsAccessor = (*SeldonDeploymentStatus)(nil)

func (ss *SeldonDeploymentStatus) InitializeConditions() {
	conditionSet.Manage(ss).InitializeConditions()
}

// IsReady returns if the service is ready to serve the requested configuration.
func (ss *SeldonDeploymentStatus) IsReady() bool {
	return conditionSet.Manage(ss).IsHappy()
}

// GetCondition returns the condition by name.
func (ss *SeldonDeploymentStatus) GetCondition(t apis.ConditionType) *apis.Condition {
	return conditionSet.Manage(ss).GetCondition(t)
}

// IsConditionReady returns the readiness for a given condition
func (ss *SeldonDeploymentStatus) IsConditionReady(t apis.ConditionType) bool {
	return conditionSet.Manage(ss).GetCondition(t) != nil && conditionSet.Manage(ss).GetCondition(t).Status == v1.ConditionTrue
}

func (ss *SeldonDeploymentStatus) SetCondition(conditionType apis.ConditionType, condition *apis.Condition) {
	switch {
	case condition == nil:
	case condition.Status == v1.ConditionUnknown:
		conditionSet.Manage(ss).MarkUnknown(conditionType, condition.Reason, condition.Message)
	case condition.Status == v1.ConditionTrue:
		conditionSet.Manage(ss).MarkTrue(conditionType)
	case condition.Status == v1.ConditionFalse:
		conditionSet.Manage(ss).MarkFalse(conditionType, condition.Reason, condition.Message)
	}
}
