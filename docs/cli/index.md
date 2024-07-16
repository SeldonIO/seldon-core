# CLI

Seldon provides a CLI to allow easy management and testing of Model, Experiment, and Pipeline resources.

At present this needs to be built by hand from the operator folder.

```
make build-seldon     # for linux/macOS amd64
make build-seldon-arm # for macOS ARM
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
:end-before: // Help statements
```

## Kubernetes Usage

### Inference Service

For a default install into the `seldon-mesh` namespace if you have exposed the inference `svc` as a loadbalancer you will find it at:

```
kubectl get svc seldon-mesh -n seldon-mesh -o jsonpath='{.status.loadBalancer.ingress[0].ip}'
```

Use above IP at port 80:

```
export SELDON_INFER_HOST=<ip>:80
```

### Scheduler Service

For a default install into the `seldon-mesh` namespace if you have exposed the scheduler svc as a loadbalancer you will find it at:

```
kubectl get svc seldon-scheduler -n seldon-mesh -o jsonpath='{.status.loadBalancer.ingress[0].ip}'
```

Use above IP at port 9004:

```
export SELDON_SCHEDULE_HOST=<ip>:9004
```

### Kafka Broker

The Kafka broker will depend on how you have installed Kafka into your Kubernetes cluster. Find the broker IP and use:

```
export SELDON_KAFKA_BROKER=<ip>:<port>
```

## Config file

You can create a config file to manage connections to running seldon core v2 installs. The settings will override any environment variable settings.

The definition is shown below:

  ```{literalinclude} ../../../../operator/pkg/cli/config.go
   :language: golang
   :start-after: // start config struct
   :end-before: // end config struct
   ```

An example below shows an example where we connect via TLS to the Seldon scheduler using our scheduler client certificate:

```
{
    "controlplane":{
	"schedulerHost": "seldon-scheduler.svc:9044",
	"tls"; true,
	"keyPath": "/home/certs/seldon-scheduler-client/tls.key",
	"crtPath": "/home/certs/seldon-scheduler-client/tls.crt",
	"caPath": "/home/certs/seldon-scheduler-client/ca.crt"
    }
}

```

To manage config files and activate them you can use the CLI command `seldon config` which has subcommands to list, add, remove, activate and decative configs.

For example:

```
$ seldon config list
config		path						active
------		----						------
kind-sasl	/home/work/seldon/cli/config-sasl.json		*

$ seldon config deactivate kind-sasl

$ seldon config list
config		path						active
------		----						------
kind-sasl	/home/work/seldon/cli/config-sasl.json

$ seldon config add gcp-scv2 ~/seldon/cli/gcp.json

$ seldon config list
config		path						active
------		----						------
gcp-scv2	/home/work/seldon/cli/gcp.json
kind-sasl	/home/work/seldon/cli/config-sasl.json

$ seldon config activate gcp-scv2

$ seldon config list
config		path						active
------		----						------
gcp-scv2	/home/work/seldon/cli/gcp.json	    		*
kind-sasl	/home/work/seldon/cli/config-sasl.json

$ seldon config list kind-sasl
{
  "controlplane": {
    "schedulerHost": "172.19.255.2:9004"
  },
  "kafka": {
    "bootstrap": "172.19.255.3:9093",
    "caPath": "/home/work/gcp/scv2/certs/seldon-cluster-ca-cert/ca.crt"
  }
}
```

## TLS Certificates for Local Use

For running with Kubernetes TLS connections on the control and/or data plane, certificates will need to be downloaded locally. We provide an example script which will download certificates from a Kubernetes secret and store them in a folder. It can be found in `hack/download-k8s-certs.sh` and takes 2 or 3 arguments:

```
./download-k8s-certs.sh <namespace> <secret> [<folder>]
```

e.g.:

```
./download-k8s-certs.sh seldon-mesh seldon-scheduler-client
```

```{toctree}
:maxdepth: 1
:hidden:

docs/seldon.md
docs/seldon_config.md
docs/seldon_model.md
docs/seldon_experiment.md
docs/seldon_pipeline.md
docs/seldon_server.md
docs/seldon_config_activate.md
docs/seldon_config_add.md
docs/seldon_config_deactivate.md
docs/seldon_config_list.md
docs/seldon_config_remove.md
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
