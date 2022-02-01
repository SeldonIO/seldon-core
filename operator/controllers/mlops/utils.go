package mlops

import (
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func setControllerReferences(controller v1.Object, controllees []v1.Object, scheme *runtime.Scheme) error {
	for _, controllee := range controllees {
		err := controllerutil.SetControllerReference(controller, controllee, scheme)
		if err != nil {
			return err
		}
	}
	return nil
}
