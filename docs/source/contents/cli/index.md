# CLI

Seldon provides a CLI to allow easy management and testing of Model, Experiment, and Pipeline resources.

At present this needs to be built by hand from the operator folder.

```
make build-seldon
```

Then place the `bin/seldon` executable in your path.

 * [cli docs](./docs/seldon.md)


## Environment Variables and Services

The CLI talks to 3 backend services on default endpoints:

 1. The Seldon Core V2 Scheduler: default 0.0.0.0:9004
 2. The Seldon Core inference endpoint: default 0.0.0.0:9000
 3. The Seldon Kafka broker: default: 0.0.0.0:9092

These defaults will be correct when Seldon Core v2 is installed locally as per the docs. For Kubernetes, you will need to change these by defining environment variables.

```{literalinclude} ../../../../operator/cmd/seldon/cli/flags.go
:language: golang
:start-after: // Defaults
```

## Kubernetes Usage

### Inference Service

For a default install into the `seldon-mesh` namespace if you have exposed the inference svc as a loadbalancer you will find it at:

```
kubectl get svc seldon-mesh -n seldon-mesh -o jsonpath='{.status.loadBalancer.ingress[0].ip}'
```

So use above ip at port 80:

```
export SELDON_INFER_HOST=<ip>:80
```

### Scheduler Service

For a default install into the `seldon-mesh` namespace if you have exposed the scheduler svc as a loadbalancer you will find it at:

```
kubectl get svc seldon-scheduler -n seldon-mesh -o jsonpath='{.status.loadBalancer.ingress[0].ip}
```

So use above ip at port 9000:

```
export SELDON_SCHEDULE_HOST=<ip>:80
```

### Kafka Broker

The kafka broker will depend on how you have installed Kafka into your Kubernetes cluster. Find the broker IP and use

```
export SELDON_KAFKA_BROKER=<ip>:<port>
```

## Config file

You can create a config file in `$HOME/.config/seldon/cli`

The definition is shown below:

  ```{literalinclude} ../../../../operator/pkg/cli/config.go
   :language: golang
   :start-after: // start config struct
   :end-before: // end config struct
   ```

An example below shows an example where we connect via TLS to the Seldon scheduler using our scheduler client certificate:

```
{
    "schedulerHost": "seldon-scheduler.svc:9044",
    "tlsKeyPath": "/home/seldon/certs/tls.key",
    "tlsCrtPath": "/home/seldon/certs/tls.crt",
    "caCrtPath": "/home/seldon/certs//ca.crt"
}

```

```{toctree}
:maxdepth: 1
:hidden:

docs/seldon.md
docs/seldon_model.md
docs/seldon_experiment.md
docs/seldon_pipeline.md
docs/seldon_server.md
docs/seldon_model_infer.md
docs/seldon_model_load.md
docs/seldon_model_status.md
docs/seldon_model_unload.md
docs/seldon_experiment_start.md
docs/seldon_experiment_status.md
docs/seldon_experiment_stop.md
docs/seldon_pipeline_infer.md
docs/seldon_pipeline_load.md
docs/seldon_pipeline_status.md
docs/seldon_pipeline_unload.md
docs/seldon_pipeline_inspect.md
docs/seldon_server_status.md
```

