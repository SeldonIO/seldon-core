package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis"
)

const (
	ModelReady apis.ConditionType = "ModelReady"
)

var modelConditionSet = apis.NewLivingConditionSet(
	ModelReady,
)

var _ apis.ConditionsAccessor = (*ModelStatus)(nil)

func (ms *ModelStatus) InitializeConditions() {
	modelConditionSet.Manage(ms).InitializeConditions()
}

func (ms *ModelStatus) IsReady() bool {
	return modelConditionSet.Manage(ms).IsHappy()
}

func (ms *ModelStatus) GetCondition(t apis.ConditionType) *apis.Condition {
	return modelConditionSet.Manage(ms).GetCondition(t)
}

func (ms *ModelStatus) IsConditionReady(t apis.ConditionType) bool {
	return modelConditionSet.Manage(ms).GetCondition(t) != nil && modelConditionSet.Manage(ms).GetCondition(t).Status == v1.ConditionTrue
}

func (ms *ModelStatus) SetCondition(conditionType apis.ConditionType, condition *apis.Condition) {
	switch {
	case condition == nil:
		modelConditionSet.Manage(ms).MarkUnknown(conditionType, "", "")
	case condition.Status == v1.ConditionUnknown:
		modelConditionSet.Manage(ms).MarkUnknown(conditionType, condition.Reason, condition.Message)
	case condition.Status == v1.ConditionTrue:
		modelConditionSet.Manage(ms).MarkTrueWithReason(conditionType, condition.Reason, condition.Message)
	case condition.Status == v1.ConditionFalse:
		modelConditionSet.Manage(ms).MarkFalse(conditionType, condition.Reason, condition.Message)
	}
}

func (ms *ModelStatus) CreateAndSetCondition(conditionType apis.ConditionType, isTrue bool, reason string) {
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
	ms.SetCondition(conditionType, &condition)
}
