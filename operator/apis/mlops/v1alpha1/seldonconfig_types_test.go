/*
Copyright 2023 Seldon Technologies Ltd.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"testing"

	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
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
					Disable:              true,
					OtelExporterEndpoint: "bar",
				},
				KafkaConfig: KafkaConfig{},
			},
			expected: SeldonConfiguration{
				TracingConfig: TracingConfig{
					Disable:              true,
					OtelExporterEndpoint: "bar",
					Ratio:                "0.5",
				},
			},
		},
		{
			name: "no tracing overrides",
			defaults: SeldonConfiguration{
				TracingConfig: TracingConfig{
					Disable:              false,
					OtelExporterEndpoint: "foo",
					Ratio:                "0.5",
				},
			},
			runtime: SeldonConfiguration{},
			expected: SeldonConfiguration{
				TracingConfig: TracingConfig{
					Disable:              false,
					OtelExporterEndpoint: "foo",
					Ratio:                "0.5",
				},
			},
		},
		{
			name: "kafka overrides",
			defaults: SeldonConfiguration{
				KafkaConfig: KafkaConfig{
					BootstrapServers:      "h1,h2",
					ConsumerGroupIdPrefix: "foo",
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
					ConsumerGroupIdPrefix: "bar",
					Consumer:              map[string]intstr.IntOrString{},
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
					BootstrapServers:      "h1,h2",
					ConsumerGroupIdPrefix: "bar",
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
					BootstrapServers:      "h1,h2",
					ConsumerGroupIdPrefix: "foo",
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
					BootstrapServers:      "h1,h2",
					ConsumerGroupIdPrefix: "foo",
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
