/*
Copyright 2022 Seldon Technologies Ltd.

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

func (ms *ModelStatus) CreateAndSetCondition(
	conditionType apis.ConditionType,
	isTrue bool,
	message string,
	reason string,
) {
	condition := apis.Condition{}
	if isTrue {
		condition.Status = v1.ConditionTrue
	} else {
		condition.Status = v1.ConditionFalse
	}
	condition.Type = conditionType
	condition.Message = message
	condition.Reason = reason
	condition.LastTransitionTime = apis.VolatileTime{
		Inner: metav1.Now(),
	}
	ms.SetCondition(conditionType, &condition)
}
