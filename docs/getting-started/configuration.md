# Configuration

Seldon can be configured via various config files.

## Kafka Configuration

We allow configuration of the Kafka integration. In general this configuration looks like:

{% @github-files/github-code-block url="https://github.com/SeldonIO/seldon-core/blob/v2/scheduler/config/kafka-internal.json" %}

The top level keys are:

* `topicPrefix` : the prefix to add to kafka topics created by Seldon
* `consumerGroupIdPrefix` : the prefix to add to Kafka consumer group IDs created by Seldon
* `bootstrap.servers` : the global bootstrap kafka servers to use
* `consumer` : consumer settings
* `producer` : producer settings
* `streams` : KStreams settings

For `topicPrefix` you can use any acceptable kafka topic characters which are
`a-z, A-Z, 0-9, . (dot), _ (underscore), and - (dash)`. We use `.` (dot) internally as topic
naming separator so we would suggest you don't end your topic prefix with a dot for clarity. For
illustration, an example topic could be `seldon.default.model.mymodel.inputs` where `seldon` is the topic prefix.

The `consumerGroupIdPrefix` will ensure that all consumer groups created have a given prefix.

### Kubernetes

For Kubernetes this is controlled via a ConfigMap called `seldon-kafka` whose default values are
defined in the `SeldonConfig` custom resource.

{% @github-files/github-code-block url="https://github.com/SeldonIO/seldon-core/blob/v2/k8s/yaml/components.yaml" %}

When the `SeldonRuntime` is installed in a namespace a configMap will be created with these
settings for Kafka configuration.

To customize the settings you can add and modify the Kafka configuration via Helm, for example
below is a custom Helm values file that add compression for producers:

{% @github-files/github-code-block url="https://github.com/SeldonIO/seldon-core/blob/v2/k8s/samples/values-runtime-kafka-compression.yaml" %}

To use this with the SeldonRuntime Helm chart:

```sh
helm install seldon-v2-runtime k8s/helm-charts/seldon-core-v2-runtime \
    --namespace seldon-mesh \
    --values k8s/samples/values-runtime-kafka-compression.yaml
```
### Topic and consumer isolation

If you use a shared Kafka cluster with other applications you may want to isolate the topic
names and consumer group IDs from other users of the cluster to ensure there is no name
clash. For this we provide two settings:

* `topicPrefix`: set a prefix for all topics
* `consumerGroupIdPrefix`: set a prefix for all consumer groups

An example to set this in the configuration when using the helm installation is showm below for creating the default `SeldonConfig`:

```sh
helm upgrade --install seldon-v2 k8s/helm-charts/seldon-core-v2-setup/ -n seldon-mesh \
    --set controller.clusterwide=true \
    --set kafka.topicPrefix=myorg \
    --set kafka.consumerGroupIdPrefix=myorg
```

You can find a worked example [here](../examples/k8s-clusterwide.md).

You can create alternate `SeldonConfig`s with different values or override values for particular `SeldonRuntime` installs.

## Tracing Configuration

We allow configuration of tracing. This file looks like:

{% @github-files/github-code-block url="https://github.com/SeldonIO/seldon-core/blob/v2/scheduler/config/tracing-internal.json" %}

The top level keys are:

* `enable` : whether to enable tracing
* `otelExporterEndpoint` : The host and port for the OTEL exporter
* `otelExporterProtocol` : The protocol for the OTEL exporter. Currently used for
    jvm-based components only (such as dataflow-engine), because `opentelemetry-java-instrumentation`
    requires a http(s) URI for the endpoint but defaults to `http/protobuf` as a protocol.
    Because of this, gRPC connections (over http) can only be set up by setting this option to `grpc`
* `ratio` : The ratio of requests to trace. Takes values between 0 and 1 inclusive.



### Kubernetes

For Kubernetes this is controlled via a ConfigMap call `seldon-tracing` whose default value is shown below:

{% @github-files/github-code-block url="https://github.com/SeldonIO/seldon-core/blob/v2/scheduler/k8s/config/tracing.yaml" %}

Note, this `ConfigMap` is created via our Helm charts and there is usually no need to modify it manually.

At present Java instrumentation (for the dataflow engine) is duplicated via separate keys.
