# Configuration

Seldon can be configured via various config files.

## Kafka Configuration

We allow configuration of the Kafka integration. In general this configuration looks like:

```{literalinclude} ../../../../../scheduler/config/kafka-internal.json
:language: json
```

The top level keys are:

 * `topicPrefix` : the prefix to add to kafka topics created by Seldon
 * `consumerGroupIdPrefix` : the prefix to add to kafka consumer group ids created by Seldon
 * `bootstrap.servers` : the global bootstrap kafka servers to use
 * `consumer` : consumer settings
 * `producer` : producer settings
 * `streams` : KStreams settings

For `topicPrefix` you can use any acceptable kafka topic characters which are `a-z, A-Z, 0-9, . (dot), _ (underscore), and - (dash)`. We use `.` (dot) internally as topic naming separator so we would suggest you don't end your topic prefix with a dot for clarity. For illustration, an example topic could be `seldon.default.model.mymodel.inputs` where `seldon` is the topic prefix.

The `consumerGroupIdPrefix` will ensure that all consumer groups created have a given prefix.

### Kubernetes

For Kubernetes this is controlled via a ConfigMap called `seldon-kafka` whose default values are defined in the `SeldonConfig` custom resource.

```{literalinclude} ../../../../../k8s/yaml/components.yaml
:language: yaml
:start-after:     kafkaConfig
:end-before:     serviceConfig
```

When the `SeldonRuntime` is installed in a namespace a configMap will be created with these settings for Kafka configuration.

To customize the settings you can add and modify the Kafka configuration via Helm, for example below is a custom Helm values file that add compression for producers:

```{literalinclude} ../../../../../k8s/samples/values-runtime-kafka-compression.yaml
:language: yaml
```
To use this with the SeldonRuntime Helm chart:

```
helm install seldon-v2-runtime k8s/helm-charts/seldon-core-v2-runtime \
    --namespace seldon-mesh \
    --values k8s/samples/values-runtime-kafka-compression.yaml
```

## Tracing Configuration

We allow configuration of tracing. This file looks like:

```{literalinclude} ../../../../../scheduler/config/tracing-internal.json
:language: json
```

The top level keys are:

 * `enable` : whether to enable tracing
 * `otelExporterEndpoint` : The host and port for the OTEL exporter
 * `ratio` : The ratio of requests to trace. Takes values between 0 and 1 inclusive.



### Kubernetes

For Kubernetes this is controlled via a ConfigMap call `seldon-tracing` whose default value is shown below:

```{literalinclude} ../../../../../scheduler/k8s/config/tracing.yaml
:language: yaml
```

Note, this ConfigMap is created via our Helm charts and there is usually no need to modify it manually.

At present Java instrumentation (for the dataflow engine) is duplicated via separate keys.
