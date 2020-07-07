package seldon

import (
	corev1 "k8s.io/api/core/v1"
)

func GetContainerForPredictiveUnit(p *PredictorSpec, name string) *corev1.Container {
	for j := 0; j < len(p.ComponentSpecs); j++ {
		cSpec := p.ComponentSpecs[j]
		for k := 0; k < len(cSpec.Spec.Containers); k++ {
			c := &cSpec.Spec.Containers[k]
			if c.Name == name {
				return c
			}
		}
	}
	return nil
}
