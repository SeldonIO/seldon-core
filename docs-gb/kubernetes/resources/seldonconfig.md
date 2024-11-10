# Seldon Config

{% hint style="info" %}
**Note**: This section is for advanced usage where you want to define how seldon is installed in each namespace.
{% endhint %}

The SeldonConfig resource defines the core installation components installed by Seldon. If you wish to
install Seldon, you can use the [SeldonRuntime](seldonruntime.md) resource which allows easy
overriding of some parts defined in this specification. In general, we advise core DevOps to use
the default SeldonConfig or customize it for their usage. Individual installation of Seldon can
then use the SeldonRuntime with a few overrides for special customisation needed in that namespace.

The specification contains core PodSpecs for each core component and a section for general configuration
including the ConfigMaps that are created for the Agent (rclone defaults), Kafka and Tracing (open telemetry).

```go
type SeldonConfigSpec struct {
	Components []*ComponentDefn    `json:"components,omitempty"`
	Config     SeldonConfiguration `json:"config,omitempty"`
}

type SeldonConfiguration struct {
	TracingConfig TracingConfig      `json:"tracingConfig,omitempty"`
	KafkaConfig   KafkaConfig        `json:"kafkaConfig,omitempty"`
	AgentConfig   AgentConfiguration `json:"agentConfig,omitempty"`
	ServiceConfig ServiceConfig      `json:"serviceConfig,omitempty"`
}

type ServiceConfig struct {
	GrpcServicePrefix string         `json:"grpcServicePrefix,omitempty"`
	ServiceType       v1.ServiceType `json:"serviceType,omitempty"`
}

type KafkaConfig struct {
	BootstrapServers      string                        `json:"bootstrap.servers,omitempty"`
	ConsumerGroupIdPrefix string                        `json:"consumerGroupIdPrefix,omitempty"`
	Debug                 string                        `json:"debug,omitempty"`
	Consumer              map[string]intstr.IntOrString `json:"consumer,omitempty"`
	Producer              map[string]intstr.IntOrString `json:"producer,omitempty"`
	Streams               map[string]intstr.IntOrString `json:"streams,omitempty"`
	TopicPrefix           string                        `json:"topicPrefix,omitempty"`
}

type AgentConfiguration struct {
	Rclone RcloneConfiguration `json:"rclone,omitempty" yaml:"rclone,omitempty"`
}

type RcloneConfiguration struct {
	ConfigSecrets []string `json:"config_secrets,omitempty" yaml:"config_secrets,omitempty"`
	Config        []string `json:"config,omitempty" yaml:"config,omitempty"`
}

type TracingConfig struct {
	Disable              bool   `json:"disable,omitempty"`
	OtelExporterEndpoint string `json:"otelExporterEndpoint,omitempty"`
	OtelExporterProtocol string `json:"otelExporterProtocol,omitempty"`
	Ratio                string `json:"ratio,omitempty"`
}

type ComponentDefn struct {
	// +kubebuilder:validation:Required

	Name                 string                  `json:"name"`
	Labels               map[string]string       `json:"labels,omitempty"`
	Annotations          map[string]string       `json:"annotations,omitempty"`
	Replicas             *int32                  `json:"replicas,omitempty"`
	PodSpec              *v1.PodSpec             `json:"podSpec,omitempty"`
	VolumeClaimTemplates []PersistentVolumeClaim `json:"volumeClaimTemplates,omitempty"`
}
```

Some of these values can be overridden on a per namespace basis via the SeldonRuntime resource. Labels and annotations
can also be set at the component level - these will be merged with the labels and annotations from the SeldonConfig
resource in which they are defined and added to the component's corresponding Deployment, or StatefulSet.

The default configuration is shown below.

{% @github-files/github-code-block url="https://github.com/SeldonIO/seldon-core/blob/v2/operator/config/seldonconfigs/default.yaml" %}

