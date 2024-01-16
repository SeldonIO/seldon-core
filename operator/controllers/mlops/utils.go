/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

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
