package v1

import (
	autoscaling "k8s.io/api/autoscaling/v2"
	autoscalingv2beta1 "k8s.io/api/autoscaling/v2beta1"
	"k8s.io/apimachinery/pkg/api/resource"
	"testing"

	. "github.com/onsi/gomega"
)

func TestHPAConversion(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name     string
		input    autoscalingv2beta1.MetricSpec
		expected autoscaling.MetricSpec
	}
	value := resource.MustParse("100")
	tests := []test{
		{
			name: "object",
			input: autoscalingv2beta1.MetricSpec{
				Type: autoscalingv2beta1.ObjectMetricSourceType,
				Object: &autoscalingv2beta1.ObjectMetricSource{
					Target: autoscalingv2beta1.CrossVersionObjectReference{
						Kind:       "sdep",
						APIVersion: "v1",
						Name:       "foo",
					},
					MetricName:  "metric",
					TargetValue: resource.MustParse("100"),
				},
			},
			expected: autoscaling.MetricSpec{
				Type: autoscaling.ObjectMetricSourceType,
				Object: &autoscaling.ObjectMetricSource{
					DescribedObject: autoscaling.CrossVersionObjectReference{
						Kind:       "sdep",
						APIVersion: "v1",
						Name:       "foo",
					},
					Target: autoscaling.MetricTarget{
						Type:  autoscaling.ValueMetricType,
						Value: &value,
					},
					Metric: autoscaling.MetricIdentifier{
						Name: "metric",
					},
				},
			},
		},
		{
			name: "pods",
			input: autoscalingv2beta1.MetricSpec{
				Type: autoscalingv2beta1.PodsMetricSourceType,
				Pods: &autoscalingv2beta1.PodsMetricSource{
					MetricName:         "metric",
					TargetAverageValue: resource.MustParse("100"),
				},
			},
			expected: autoscaling.MetricSpec{
				Type: autoscaling.PodsMetricSourceType,
				Pods: &autoscaling.PodsMetricSource{
					Target: autoscaling.MetricTarget{
						Type:         autoscaling.AverageValueMetricType,
						AverageValue: &value,
					},
					Metric: autoscaling.MetricIdentifier{
						Name: "metric",
					},
				},
			},
		},
		{
			name: "resource",
			input: autoscalingv2beta1.MetricSpec{
				Type: autoscalingv2beta1.ResourceMetricSourceType,
				Resource: &autoscalingv2beta1.ResourceMetricSource{
					Name:               "resource",
					TargetAverageValue: &value,
				},
			},
			expected: autoscaling.MetricSpec{
				Type: autoscaling.ResourceMetricSourceType,
				Resource: &autoscaling.ResourceMetricSource{
					Name: "resource",
					Target: autoscaling.MetricTarget{
						Type:         autoscaling.AverageValueMetricType,
						AverageValue: &value,
					},
				},
			},
		},
		{
			name: "container",
			input: autoscalingv2beta1.MetricSpec{
				Type: autoscalingv2beta1.ContainerResourceMetricSourceType,
				ContainerResource: &autoscalingv2beta1.ContainerResourceMetricSource{
					Name:               "resource",
					Container:          "container",
					TargetAverageValue: &value,
				},
			},
			expected: autoscaling.MetricSpec{
				Type: autoscaling.ContainerResourceMetricSourceType,
				ContainerResource: &autoscaling.ContainerResourceMetricSource{
					Name:      "resource",
					Container: "container",
					Target: autoscaling.MetricTarget{
						Type:         autoscaling.AverageValueMetricType,
						AverageValue: &value,
					},
				},
			},
		},
		{
			name: "external",
			input: autoscalingv2beta1.MetricSpec{
				Type: autoscalingv2beta1.ExternalMetricSourceType,
				External: &autoscalingv2beta1.ExternalMetricSource{
					MetricName:         "metric",
					TargetAverageValue: &value,
				},
			},
			expected: autoscaling.MetricSpec{
				Type: autoscaling.ExternalMetricSourceType,
				External: &autoscaling.ExternalMetricSource{
					Metric: autoscaling.MetricIdentifier{
						Name: "metric",
					},
					Target: autoscaling.MetricTarget{
						Type:         autoscaling.AverageValueMetricType,
						AverageValue: &value,
					},
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res := convertMetricSpec(test.input)
			g.Expect(res).To(Equal(test.expected))
		})
	}
}
