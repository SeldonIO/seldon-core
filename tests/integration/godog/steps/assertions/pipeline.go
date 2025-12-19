/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package assertions

import (
	"fmt"

	schedulerAPI "github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"
	"github.com/seldonio/seldon-core/operator/v2/apis/mlops/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
)

func PipelineReady(obj runtime.Object) (bool, error) {
	if obj == nil {
		return false, nil
	}

	pipeline, ok := obj.(*v1alpha1.Pipeline)
	if !ok {
		return false, fmt.Errorf("unexpected type %T, expected *v1alpha1.Pipeline", obj)
	}

	if pipeline.Status.IsReady() {
		return true, nil
	}

	return false, nil
}

func PipelineNotReady(obj runtime.Object) (bool, error) {
	if obj == nil {
		// pipeline does not exist (yet) or has been deleted: treat as "not ready"
		return false, nil // or true,nil if you want "absence = not ready"
	}

	pipeline, ok := obj.(*v1alpha1.Pipeline)
	if !ok {
		return false, fmt.Errorf("unexpected type %T, expected *v1alpha1.Pipeline", obj)
	}

	if pipeline.Status.IsReady() {
		return false, nil
	}

	return true, nil
}

func PipelineFailed(obj runtime.Object) (bool, error) {
	if obj == nil {
		return false, nil
	}

	pipeline, ok := obj.(*v1alpha1.Pipeline)
	if !ok {
		return false, fmt.Errorf("unexpected type %T, expected *v1alpha1.Pipeline", obj)
	}

	condition := pipeline.Status.GetCondition(v1alpha1.PipelineReady)
	if condition == nil {
		return false, nil
	}

	if condition.Message == schedulerAPI.PipelineVersionState_PipelineFailed.String() {
		return true, nil
	}

	return false, nil
}

func PipelineFailedTerminate(obj runtime.Object) (bool, error) {
	if obj == nil {
		return false, nil
	}

	pipeline, ok := obj.(*v1alpha1.Pipeline)
	if !ok {
		return false, fmt.Errorf("unexpected type %T, expected *v1alpha1.Pipeline", obj)
	}

	condition := pipeline.Status.GetCondition(v1alpha1.PipelineReady)

	if condition == nil {
		return false, nil
	}

	if condition.Message == schedulerAPI.PipelineVersionState_PipelineFailed.String() {
		return true, nil
	}

	return false, nil
}

// PipelineDeleted returns done=true when the pipeline no longer exists in the store.
// When called with obj == nil (delete event or never present), we treat it as deleted.
func PipelineDeleted(obj runtime.Object) (bool, error) {
	if obj == nil {
		// The object is not present in the store: consider it deleted.
		return true, nil
	}

	// If you want to be extra strict and only consider it "deleted" when actually gone,
	// you could also check DeletionTimestamp here, but it's usually unnecessary.
	pipeline, ok := obj.(*v1alpha1.Pipeline)
	if !ok {
		return false, fmt.Errorf("unexpected type %T, expected *v1alpha1.Pipeline", obj)
	}

	// If you want to assert it's *in the process* of being deleted first:
	if pipeline.DeletionTimestamp != nil {
		// still present but marked for deletion; NOT done yet if you want full delete
		return false, nil
	}

	// still present and not being deleted -> not done
	return false, nil
}

//func PipelineReadyMessageCondition(expectedMessage string) k8sclient.ConditionFunc {
//	return func(obj runtime.Object) (bool, error) {
//		if obj == nil {
//			return false, nil
//		}
//
//		pipeline, ok := obj.(*v1alpha1.Pipeline)
//		if !ok {
//			return false, fmt.Errorf("unexpected type %T, expected *mlopsv1alpha1.Model", obj)
//		}
//
//		pipeline.Status.Status.
//
//		return false, nil
//	}
//}
