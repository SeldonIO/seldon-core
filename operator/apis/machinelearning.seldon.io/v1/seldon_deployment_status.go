package v1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	DeploymentsReady        apis.ConditionType = "DeploymentsReady"
	ServicesReady           apis.ConditionType = "ServicesReady"
	KedaReady               apis.ConditionType = "KedaReady"
	VirtualServicesReady    apis.ConditionType = "istioVirtualServicesReady"
	HpasReady               apis.ConditionType = "HpasReady"
	PdbsReady               apis.ConditionType = "PdbsReady"
	AmbassadorMappingsReady apis.ConditionType = "AmbassadorMappingsReady"

	SvcNotReadyReason           string = "Not all services created"
	SvcReadyReason              string = "All services created"
	KedaNotDefinedReason        string = "No KEDA resources defined"
	KedaNotReadyReason          string = "KEDA resources not ready"
	KedaReadyReason             string = "All KEDA resources ready"
	HpaNotDefinedReason         string = "No HPAs defined"
	HpaNotReadyReason           string = "HPAs not ready"
	HpaReadyReason              string = "All HPAs resources ready"
	PdbNotDefinedReason         string = "No PDBs defined"
	PdbNotReadyReason           string = "PDBs not ready"
	PdbReadyReason              string = "All PDBs resources ready"
	VirtualServiceNotDefined    string = "No VirtualServices defined"
	VirtualServiceNotReady      string = "Not all VirtualServices created"
	VirtualServiceReady         string = "All VirtualServices created"
	AmbassadorMappingNotDefined string = "No Ambassador Mappaings defined"
	AmbassadorMappingNotReady   string = "Not all Ambassador Mappings created"
	AmbassadorMappingReady      string = "All Ambassador Mappings created"
)

// InferenceService Ready condition is depending on predictor and route readiness condition
var conditionSet = apis.NewLivingConditionSet(
	DeploymentsReady,
	ServicesReady,
	KedaReady,
	VirtualServicesReady,
	HpasReady,
	PdbsReady,
	AmbassadorMappingsReady,
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
		conditionSet.Manage(ss).MarkUnknown(conditionType, "", "")
	case condition.Status == v1.ConditionUnknown:
		conditionSet.Manage(ss).MarkUnknown(conditionType, condition.Reason, condition.Message)
	case condition.Status == v1.ConditionTrue:
		conditionSet.Manage(ss).MarkTrueWithReason(conditionType, condition.Reason, condition.Message)
	case condition.Status == v1.ConditionFalse:
		conditionSet.Manage(ss).MarkFalse(conditionType, condition.Reason, condition.Message)
	}
}

func (ss *SeldonDeploymentStatus) CreateCondition(conditionType apis.ConditionType, isTrue bool, reason string) {
	condition := apis.Condition{}
	if isTrue {
		condition.Status = v1.ConditionTrue
	} else {
		condition.Status = v1.ConditionFalse
	}
	condition.Type = conditionType
	condition.Reason = reason
	condition.LastTransitionTime = apis.VolatileTime{
		Inner: metav1.Now(),
	}
	ss.SetCondition(conditionType, &condition)
}
