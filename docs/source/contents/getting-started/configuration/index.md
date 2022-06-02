# Configuration

Seldon can be configured via various config files.

## Kafka Configuration

We allow configuration of the Kafka integration. In general this configuration looks like:

```{literalinclude} ../../../../../scheduler/config/kafka-internal.json
:language: json
```

The top level keys are:

 * `bootstrap.servers` : the global bootstrap kafka servers to use
 * `consumer` : consumer settings
 * `producer` : producer settings
 * `streams` : KStreams settings


### Kubernetes

For Kubernetes this is controlled via a ConfigMap called `seldon-kafka` whose default value is shown below:

```{literalinclude} ../../../../../scheduler/k8s/config/kafka.yaml
:language: yaml
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

At present Java instrumentation (for the dataflow engine) is duplicated via separate keys.
