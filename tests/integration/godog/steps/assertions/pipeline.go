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
