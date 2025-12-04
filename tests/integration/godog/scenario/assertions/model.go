package assertions

import (
	"fmt"

	mlopsv1alpha1 "github.com/seldonio/seldon-core/operator/v2/apis/mlops/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func ModelReady(obj runtime.Object) (bool, error) {
	if obj == nil {
		return false, nil
	}

	model, ok := obj.(*mlopsv1alpha1.Model)
	if !ok {
		return false, fmt.Errorf("unexpected type %T, expected *mlopsv1alpha1.Model", obj)
	}

	for _, c := range model.Status.Conditions {
		if c.Type == "Ready" && c.Status == corev1.ConditionTrue {
			return true, nil
		}
	}

	return false, nil
}
