package v1

import (
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	autoscalingv2beta1 "k8s.io/api/autoscaling/v2beta1"
)

func ConvertMetricSpecSlice(metrics []autoscalingv2beta1.MetricSpec) []autoscalingv2.MetricSpec {
	var metricsv2 []autoscalingv2.MetricSpec
	for _, ms := range metrics {
		metricsv2 = append(metricsv2, convertMetricSpec(ms))
	}
	return metricsv2
}

func convertMetricSpec(spec autoscalingv2beta1.MetricSpec) autoscalingv2.MetricSpec {
	return autoscalingv2.MetricSpec{
		Type:              convertType(spec.Type),
		Object:            convertObjectMetricSource(spec.Object),
		Pods:              convertPodMetricSource(spec.Pods),
		Resource:          convertResourceMetricSource(spec.Resource),
		ContainerResource: convertContainerResourceMetricSource(spec.ContainerResource),
		External:          convertExternalMetricSource(spec.External),
	}
}

func convertExternalMetricSource(ems *autoscalingv2beta1.ExternalMetricSource) *autoscalingv2.ExternalMetricSource {
	if ems != nil {
		emsV2 := &autoscalingv2.ExternalMetricSource{
			Metric: autoscalingv2.MetricIdentifier{
				Name:     ems.MetricName,
				Selector: ems.MetricSelector,
			},
			Target: autoscalingv2.MetricTarget{
				AverageValue: ems.TargetAverageValue,
				Value:        ems.TargetValue,
			},
		}
		if ems.TargetValue != nil {
			emsV2.Target.Type = autoscalingv2.ValueMetricType
		} else {
			emsV2.Target.Type = autoscalingv2.AverageValueMetricType
		}
		return emsV2
	}
	return nil
}

func convertContainerResourceMetricSource(crms *autoscalingv2beta1.ContainerResourceMetricSource) *autoscalingv2.ContainerResourceMetricSource {
	if crms != nil {
		crmsV2 := &autoscalingv2.ContainerResourceMetricSource{
			Name:      crms.Name,
			Container: crms.Container,
			Target: autoscalingv2.MetricTarget{
				AverageUtilization: crms.TargetAverageUtilization,
				AverageValue:       crms.TargetAverageValue,
			},
		}
		if crms.TargetAverageValue != nil {
			crmsV2.Target.Type = autoscalingv2.AverageValueMetricType
		} else {
			crmsV2.Target.Type = autoscalingv2.UtilizationMetricType
		}
		return crmsV2
	}
	return nil
}

func convertResourceMetricSource(rms *autoscalingv2beta1.ResourceMetricSource) *autoscalingv2.ResourceMetricSource {
	if rms != nil {
		rmsV2 := &autoscalingv2.ResourceMetricSource{
			Name: rms.Name,
			Target: autoscalingv2.MetricTarget{
				AverageUtilization: rms.TargetAverageUtilization,
				AverageValue:       rms.TargetAverageValue,
			},
		}
		if rms.TargetAverageValue != nil {
			rmsV2.Target.Type = autoscalingv2.AverageValueMetricType
		} else {
			rmsV2.Target.Type = autoscalingv2.UtilizationMetricType
		}
		return rmsV2
	}
	return nil
}

func convertPodMetricSource(pms *autoscalingv2beta1.PodsMetricSource) *autoscalingv2.PodsMetricSource {
	if pms != nil {
		return &autoscalingv2.PodsMetricSource{
			Target: autoscalingv2.MetricTarget{
				Type:         autoscalingv2.AverageValueMetricType,
				AverageValue: &pms.TargetAverageValue,
			},
			Metric: autoscalingv2.MetricIdentifier{
				Name:     pms.MetricName,
				Selector: pms.Selector,
			},
		}
	}
	return nil
}

func convertObjectMetricSource(oms *autoscalingv2beta1.ObjectMetricSource) *autoscalingv2.ObjectMetricSource {
	if oms != nil {
		omsV2 := &autoscalingv2.ObjectMetricSource{
			DescribedObject: autoscalingv2.CrossVersionObjectReference{
				Kind:       oms.Target.Kind,
				Name:       oms.Target.Name,
				APIVersion: oms.Target.APIVersion,
			},
			Target: autoscalingv2.MetricTarget{
				Value:        &oms.TargetValue,
				AverageValue: oms.AverageValue,
			},
			Metric: autoscalingv2.MetricIdentifier{
				Name:     oms.MetricName,
				Selector: oms.Selector,
			},
		}
		if oms.AverageValue != nil {
			omsV2.Target.Type = autoscalingv2.AverageValueMetricType
		} else {
			omsV2.Target.Type = autoscalingv2.ValueMetricType
		}
		return omsV2
	}
	return nil
}

func convertType(ty autoscalingv2beta1.MetricSourceType) autoscalingv2.MetricSourceType {
	switch ty {
	case autoscalingv2beta1.ObjectMetricSourceType:
		return autoscalingv2.ObjectMetricSourceType
	case autoscalingv2beta1.ContainerResourceMetricSourceType:
		return autoscalingv2.ContainerResourceMetricSourceType
	case autoscalingv2beta1.ExternalMetricSourceType:
		return autoscalingv2.ExternalMetricSourceType
	case autoscalingv2beta1.PodsMetricSourceType:
		return autoscalingv2.PodsMetricSourceType
	case autoscalingv2beta1.ResourceMetricSourceType:
		return autoscalingv2.ResourceMetricSourceType
	}
	//Should not reach here as we cover all types
	return autoscalingv2.ObjectMetricSourceType
}
