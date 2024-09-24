---
description: Learn more about using Seldon CLI commands
---

# CLI

Seldon provides a CLI for easy management and testing of model, experiment, and pipeline resources. You can use Seldon CLI in testing environments, while `kubectl` is recommended for managing Seldon Core 2 resources in a Kubernetes production environment. The following table provides more information about when and where to use these command line tools.

| Usage                    | Seldon CLI                                                                                                                   | kubectl                                                                                                        |
| ------------------------ | ---------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------- |
| Production environment   | Not recommended for production environments in Kubernetes. Ideal for use outside Kubernetes.                                 | Recommended for production in Kubernetes, offering full control over resources and deployments.                |
| Primary purpose          | Simplifies management for non-Kubernetes users, abstracting control plane operations like load, unload, status.              | Kubernetes-native, manages resources via Kubernetes Custom Resources (CRs) like Deployments, Pods, etc.        |
| Control Plane Operations | Executes operations such as `load` and `unload`models through scheduler gRPC endpoints, without interaction with Kubernetes. | Interacts with Kubernetes, creating and managing CRs such as SeldonDeployments and other Kubernetes resources. |
| Data Plane Operations    | Abstracts open inference protocol to issue `infer` or `inspect` requests for testing purposes.                               | Used indirectly for data plane operations by exposing Kubernetes services and interacting with them.           |
| Scope of Operations      | Does not create or manage any Kubernetes resources or CRs; operates on internal scheduler state.                             | Manages and operates on actual Kubernetes resources, making it ideal for production use in Kubernetes.         |
| Visibility of Resources  | Resources created using Seldon  CLI are internal to the scheduler and not visible as Kubernetes resources.                   | All resources that are created using `kubectl` are visible and manageable within the Kubernetes environment.   |

## Environment Variables and Services

The CLI talks to 3 backend services on default endpoints:

1. The Seldon Core V2 Scheduler: default `0.0.0.0:9004`
2. The Seldon Core inference endpoint: default `0.0.0.0:9000`
3. The Seldon Kafka broker: default: `0.0.0.0:9092`

These defaults will be correct when Seldon Core v2 is installed locally as per the docs. For Kubernetes, you will need to change these by defining environment variables.

```go
const (
	defaultInferHost     = "0.0.0.0:9000"
	defaultKafkaHost     = "0.0.0.0:9092"
	defaultSchedulerHost = "0.0.0.0:9004"
)
```

## Kubernetes Usage

### Inference Service

For a default install into the `seldon-mesh` namespace if you have exposed the inference `svc` as a loadbalancer you will find it at:

```sh
kubectl get svc seldon-mesh -n seldon-mesh -o jsonpath='{.status.loadBalancer.ingress[0].ip}'
```

Use above IP at port 80:

```sh
export SELDON_INFER_HOST=<ip>:80
```

### Scheduler Service

For a default install into the `seldon-mesh` namespace if you have exposed the scheduler svc as a loadbalancer you will find it at:

```sh
kubectl get svc seldon-scheduler -n seldon-mesh -o jsonpath='{.status.loadBalancer.ingress[0].ip}'
```

Use above IP at port 9004:

```sh
export SELDON_SCHEDULE_HOST=<ip>:9004
```

### Kafka Broker

The Kafka broker will depend on how you have installed Kafka into your Kubernetes cluster. Find the broker IP and use:

```sh
export SELDON_KAFKA_BROKER=<ip>:<port>
```

## Config file

You can create a config file to manage connections to running seldon core v2 installs. The settings will override any environment variable settings.

The definition is shown below:

```go
type SeldonCLIConfig struct {
	Dataplane    *Dataplane    `json:"dataplane,omitempty"`
	Controlplane *ControlPlane `json:"controlplane,omitempty"`
	Kafka        *KafkaConfig  `json:"kafka,omitempty"`
}

type Dataplane struct {
	InferHost     string `json:"inferHost,omitempty"`
	Tls           bool   `json:"tls,omitempty"`
	SkipSSLVerify bool   `json:"skipSSLVerify,omitempty"`
	KeyPath       string `json:"keyPath,omitempty"`
	CrtPath       string `json:"crtPath,omitempty"`
	CaPath        string `json:"caPath,omitempty"`
}

type ControlPlane struct {
	SchedulerHost string `json:"schedulerHost,omitempty"`
	Tls           bool   `json:"tls,omitempty"`
	KeyPath       string `json:"keyPath,omitempty"`
	CrtPath       string `json:"crtPath,omitempty"`
	CaPath        string `json:"caPath,omitempty"`
}

const (
	KafkaConfigProtocolSSL          = "ssl"
	KafkaConfigProtocolSASLSSL      = "sasl_ssl"
	KafkaConfigProtocolSASLPlaintxt = "sasl_plaintxt"
)

type KafkaConfig struct {
	Bootstrap    string `json:"bootstrap,omitempty"`
	Namespace    string `json:"namespace,omitempty"`
	Protocol     string `json:"protocol,omitempty"`
	KeyPath      string `json:"keyPath,omitempty"`
	CrtPath      string `json:"crtPath,omitempty"`
	CaPath       string `json:"caPath,omitempty"`
	SaslUsername string `json:"saslUsername,omitempty"`
	SaslPassword string `json:"saslPassword,omitempty"`
	TopicPrefix  string `json:"topicPrefix,omitempty"`
}
```

An example below shows an example where we connect via TLS to the Seldon scheduler using our scheduler client certificate:

```json
{
  "controlplane": {
  	"schedulerHost": "seldon-scheduler.svc:9044",
  	"tls": true,
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

```sh
./download-k8s-certs.sh <namespace> <secret> [<folder>]
```

e.g.:

```sh
./download-k8s-certs.sh seldon-mesh seldon-scheduler-client
```
