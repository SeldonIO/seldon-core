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

For Kubernetes this is controlled via a ConfigMap call `seldon-kafka` whose default value is shown below:

```{literalinclude} ../../../../../scheduler/k8s/config/kafka.yaml
:language: yaml
```
