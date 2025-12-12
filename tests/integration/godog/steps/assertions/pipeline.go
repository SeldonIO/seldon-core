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
