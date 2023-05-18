package v1alpha1

import (
	"testing"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	. "github.com/onsi/gomega"
)

func TestSeldonConfigurationAddDefaults(t *testing.T) {
	g := NewGomegaWithT(t)
	type test struct {
		name     string
		defaults SeldonConfiguration
		runtime  SeldonConfiguration
		expected SeldonConfiguration
	}

	tests := []test{
		{
			name: "tracing overrides",
			defaults: SeldonConfiguration{
				TracingConfig: TracingConfig{
					OtelExporterEndpoint: "foo",
					Ratio:                "0.5",
				},
			},
			runtime: SeldonConfiguration{
				TracingConfig: TracingConfig{
					OtelExporterEndpoint: "bar",
				},
				KafkaConfig: KafkaConfig{},
			},
			expected: SeldonConfiguration{
				TracingConfig: TracingConfig{
					OtelExporterEndpoint: "bar",
					Ratio:                "0.5",
				},
			},
		},
		{
			name: "kafka overrides",
			defaults: SeldonConfiguration{
				KafkaConfig: KafkaConfig{
					BootstrapServers: "h1,h2",
					Consumer: map[string]intstr.IntOrString{
						"key1": intstr.FromInt(1000),
					},
					Producer: map[string]intstr.IntOrString{
						"key2": intstr.FromString("val"),
					},
					Streams: map[string]intstr.IntOrString{
						"key3": intstr.FromString("val2"),
					},
				},
			},
			runtime: SeldonConfiguration{
				KafkaConfig: KafkaConfig{
					Consumer: map[string]intstr.IntOrString{},
					Producer: map[string]intstr.IntOrString{
						"key4": intstr.FromString("val"),
					},
					Streams: map[string]intstr.IntOrString{
						"key3": intstr.FromString("val1"),
					},
				},
			},
			expected: SeldonConfiguration{
				KafkaConfig: KafkaConfig{
					BootstrapServers: "h1,h2",
					Consumer: map[string]intstr.IntOrString{
						"key1": intstr.FromInt(1000),
					},
					Producer: map[string]intstr.IntOrString{
						"key2": intstr.FromString("val"),
						"key4": intstr.FromString("val"),
					},
					Streams: map[string]intstr.IntOrString{
						"key3": intstr.FromString("val1"),
					},
				},
			},
		},
		{
			name: "kafka no overrides",
			defaults: SeldonConfiguration{
				KafkaConfig: KafkaConfig{
					BootstrapServers: "h1,h2",
					Consumer: map[string]intstr.IntOrString{
						"key1": intstr.FromInt(1000),
					},
					Producer: map[string]intstr.IntOrString{
						"key2": intstr.FromString("val"),
					},
					Streams: map[string]intstr.IntOrString{
						"key3": intstr.FromString("val2"),
					},
				},
			},
			runtime: SeldonConfiguration{
				KafkaConfig: KafkaConfig{},
			},
			expected: SeldonConfiguration{
				KafkaConfig: KafkaConfig{
					BootstrapServers: "h1,h2",
					Consumer: map[string]intstr.IntOrString{
						"key1": intstr.FromInt(1000),
					},
					Producer: map[string]intstr.IntOrString{
						"key2": intstr.FromString("val"),
					},
					Streams: map[string]intstr.IntOrString{
						"key3": intstr.FromString("val2"),
					},
				},
			},
		},
		{
			name: "agent overrides",
			defaults: SeldonConfiguration{
				AgentConfig: AgentConfiguration{
					Rclone: RcloneConfiguration{
						ConfigSecrets: []string{"sec1"},
					},
				},
			},
			runtime: SeldonConfiguration{
				AgentConfig: AgentConfiguration{
					Rclone: RcloneConfiguration{
						ConfigSecrets: []string{"sec2"},
					},
				},
			},
			expected: SeldonConfiguration{
				AgentConfig: AgentConfiguration{
					Rclone: RcloneConfiguration{
						ConfigSecrets: []string{"sec2", "sec1"},
					},
				},
			},
		},
		{
			name: "service overrides",
			defaults: SeldonConfiguration{
				ServiceConfig: ServiceConfig{
					GrpcServicePrefix: "grpc-",
					ServiceType:       v1.ServiceTypeClusterIP,
				},
			},
			runtime: SeldonConfiguration{
				ServiceConfig: ServiceConfig{
					ServiceType: v1.ServiceTypeLoadBalancer,
				},
			},
			expected: SeldonConfiguration{
				ServiceConfig: ServiceConfig{
					GrpcServicePrefix: "grpc-",
					ServiceType:       v1.ServiceTypeLoadBalancer,
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.runtime.AddDefaults(test.defaults)
			g.Expect(test.runtime).To(Equal(test.expected))
		})
	}
}
