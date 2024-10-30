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

```yaml
# operator/config/seldonconfigs/default.yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: SeldonConfig
metadata:
  name: default
spec:
  components:
  - name: seldon-dataflow-engine
    replicas: 1
    podSpec:
      containers:
      - env:
        - name: SELDON_UPSTREAM_HOST
          value: seldon-scheduler
        - name: SELDON_UPSTREAM_PORT
          value: "9008"
        - name: OTEL_JAVAAGENT_ENABLED
          valueFrom:
            configMapKeyRef:
              key: OTEL_JAVAAGENT_ENABLED
              name: seldon-tracing
        - name: OTEL_EXPORTER_OTLP_ENDPOINT
          valueFrom:
            configMapKeyRef:
              key: OTEL_EXPORTER_OTLP_ENDPOINT
              name: seldon-tracing
        - name: OTEL_EXPORTER_OTLP_PROTOCOL
          valueFrom:
            configMapKeyRef:
              key: OTEL_EXPORTER_OTLP_PROTOCOL
              name: seldon-tracing
        - name: SELDON_POD_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        image: seldonio/seldon-dataflow-engine:latest
        imagePullPolicy: Always
        name: dataflow-engine
        resources:
          limits:
            memory: 1G
          requests:
            cpu: 100m
            memory: 1G
      serviceAccountName: seldon-scheduler
      terminationGracePeriodSeconds: 5
  - name: seldon-envoy
    replicas: 1
    annotations:
        "prometheus.io/path": "/stats/prometheus"
        "prometheus.io/port": "9003"
        "prometheus.io/scrape": "true"
    podSpec:
      containers:
      - image: seldonio/seldon-envoy:latest
        imagePullPolicy: Always
        name: envoy
        ports:
        - containerPort: 9000
          name: http
        - containerPort: 9003
          name: envoy-stats
        resources:
          limits:
            memory: 128Mi
          requests:
            cpu: 100m
            memory: 128Mi
        readinessProbe:
          httpGet:
            path: /ready
            port: envoy-stats
          initialDelaySeconds: 10
          periodSeconds: 5
          failureThreshold: 3
      terminationGracePeriodSeconds: 5
  - name: hodometer
    replicas: 1
    podSpec:
      containers:
      - env:
        - name: PUBLISH_URL
          value: http://hodometer.seldon.io
        - name: SCHEDULER_HOST
          value: seldon-scheduler
        - name: SCHEDULER_PLAINTXT_PORT
          value: "9004"
        - name: SCHEDULER_TLS_PORT
          value: "9044"
        - name: POD_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        image: seldonio/seldon-hodometer:latest
        imagePullPolicy: Always
        name: hodometer
        resources:
          limits:
            memory: 32Mi
          requests:
            cpu: 1m
            memory: 32Mi
      serviceAccountName: hodometer
      terminationGracePeriodSeconds: 5
  - name: seldon-modelgateway
    replicas: 1
    podSpec:
      containers:
      - args:
        - --scheduler-host=seldon-scheduler
        - --scheduler-plaintxt-port=$(SELDON_SCHEDULER_PLAINTXT_PORT)
        - --scheduler-tls-port=$(SELDON_SCHEDULER_TLS_PORT)
        - --envoy-host=seldon-mesh
        - --envoy-port=80
        - --kafka-config-path=/mnt/kafka/kafka.json
        - --tracing-config-path=/mnt/tracing/tracing.json
        command:
        - /bin/modelgateway
        env:
        - name: SELDON_SCHEDULER_PLAINTXT_PORT
          value: "9004"
        - name: SELDON_SCHEDULER_TLS_PORT
          value: "9044"
        - name: MODELGATEWAY_MAX_NUM_CONSUMERS
          value: "100"
        - name: POD_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        image: seldonio/seldon-modelgateway:latest
        imagePullPolicy: Always
        name: modelgateway
        resources:
          limits:
            memory: 1G
          requests:
            cpu: 100m
            memory: 1G
        volumeMounts:
        - mountPath: /mnt/kafka
          name: kafka-config-volume
        - mountPath: /mnt/tracing
          name: tracing-config-volume
      serviceAccountName: seldon-scheduler
      terminationGracePeriodSeconds: 5
      volumes:
      - configMap:
          name: seldon-kafka
        name: kafka-config-volume
      - configMap:
          name: seldon-tracing
        name: tracing-config-volume
  - name: seldon-pipelinegateway
    replicas: 1
    podSpec:
      containers:
      - args:
        - --http-port=9010
        - --grpc-port=9011
        - --metrics-port=9006
        - --scheduler-host=seldon-scheduler
        - --scheduler-plaintxt-port=$(SELDON_SCHEDULER_PLAINTXT_PORT)
        - --scheduler-tls-port=$(SELDON_SCHEDULER_TLS_PORT)
        - --envoy-host=seldon-mesh
        - --envoy-port=80
        - --kafka-config-path=/mnt/kafka/kafka.json
        - --tracing-config-path=/mnt/tracing/tracing.json
        command:
        - /bin/pipelinegateway
        env:
        - name: SELDON_SCHEDULER_PLAINTXT_PORT
          value: "9004"
        - name: SELDON_SCHEDULER_TLS_PORT
          value: "9044"
        - name: POD_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        image: seldonio/seldon-pipelinegateway
        imagePullPolicy: Always
        name: pipelinegateway
        ports:
        - containerPort: 9010
          name: http
          protocol: TCP
        - containerPort: 9011
          name: grpc
          protocol: TCP
        - containerPort: 9006
          name: metrics
          protocol: TCP
        resources:
          limits:
            memory: 1G
          requests:
            cpu: 100m
            memory: 1G
        volumeMounts:
        - mountPath: /mnt/kafka
          name: kafka-config-volume
        - mountPath: /mnt/tracing
          name: tracing-config-volume
      serviceAccountName: seldon-scheduler
      terminationGracePeriodSeconds: 5
      volumes:
      - configMap:
          name: seldon-kafka
        name: kafka-config-volume
      - configMap:
          name: seldon-tracing
        name: tracing-config-volume
  - name: seldon-scheduler
    replicas: 1
    podSpec:
      containers:
      - args:
        - --pipeline-gateway-host=seldon-pipelinegateway
        - --tracing-config-path=/mnt/tracing/tracing.json
        - --db-path=/mnt/scheduler/db
        - --allow-plaintxt=$(ALLOW_PLAINTXT)
        - --kafka-config-path=/mnt/kafka/kafka.json
        - --scheduler-ready-timeout-seconds=$(SCHEDULER_READY_TIMEOUT_SECONDS)
        command:
        - /bin/scheduler
        env:
        - name: ALLOW_PLAINTXT
          value: "true"
        - name: SCHEDULER_READY_TIMEOUT_SECONDS
          value: 600
        - name: POD_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        image: seldonio/seldon-scheduler:latest
        imagePullPolicy: Always
        name: scheduler
        ports:
        - containerPort: 9002
          name: xds
        - containerPort: 9004
          name: scheduler
        - containerPort: 9044
          name: scheduler-mtls
        - containerPort: 9005
          name: agent
        - containerPort: 9055
          name: agent-mtls
        - containerPort: 9008
          name: dataflow
        resources:
          limits:
            memory: 1G
          requests:
            cpu: 100m
            memory: 1G
        volumeMounts:
        - mountPath: /mnt/kafka
          name: kafka-config-volume
        - mountPath: /mnt/tracing
          name: tracing-config-volume
        - mountPath: /mnt/scheduler
          name: scheduler-state
      serviceAccountName: seldon-scheduler
      terminationGracePeriodSeconds: 5
      volumes:
      - configMap:
          name: seldon-kafka
        name: kafka-config-volume
      - configMap:
          name: seldon-tracing
        name: tracing-config-volume
    volumeClaimTemplates:
    - name: scheduler-state
      spec:
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 1G
```
