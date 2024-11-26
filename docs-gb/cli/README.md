---
description: Learn more about using Seldon CLI commands
---

# CLI

Seldon provides a CLI for easy management and testing of model, experiment, and pipeline resources. The Seldon CLI allows you to view information about underlying Seldon resources and make changes to them through the scheduler in non-Kubernetes environments. However, it cannot modify underlying manifests within a Kubernetes cluster. Therefore, using the Seldon CLI for control plane operations in a Kubernetes environment is not recommended and is disabled by default. 

While Seldon CLI control plane operations (e.g. `load` and `unload`) in a Kubernetes environment are not recommended, there are other use cases (e.g. inspecting kafka topics in a pipeline) that are enabled easily with Seldon CLI. It offers out-of-the box deserialisation of the kafka messages according to Open Inference Protocol (OIP), allowing to test these pipelines in these environments. 

&#x20;The following table provides more information about when and where to use these command line tools.

| Usage                    | Seldon CLI                                                                                                                   | kubectl                                                                                                        |
| ------------------------ | ---------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------- |
| Primary purpose          | Simplifies management for non-Kubernetes users, abstracting control plane operations like load, unload, status.              | Kubernetes-native, manages resources via Kubernetes Custom Resources (CRs) like Deployments, Pods, etc.        |
| Control Plane Operations | Executes operations such as `load` and `unload`models through scheduler gRPC endpoints, without interaction with Kubernetes. This is disabled by default and the user has to explicitly specify `--force` flag as argument (or set envar `SELDON_FORCE_CONTROL_PLANE=true`). | Interacts with Kubernetes, creating and managing CRs such as SeldonDeployments and other Kubernetes resources. |
| Data Plane Operations    | Abstracts open inference protocol to issue `infer` or `inspect` requests for testing purposes. This is useful for example when inspecting intermediate kafka messages in a multi-step pipeline.                               | Used indirectly for data plane operations by exposing Kubernetes services and interacting with them.           |
| Visibility of Resources  | Resources created using Seldon  CLI are internal to the scheduler and not visible as Kubernetes resources.                   | All resources that are created using `kubectl` are visible and manageable within the Kubernetes environment.   |

## Running Seldon CLI as Kubernetes Deployment (Experimental)

Seldon CLI can be deployed in the same namespace along side a Core 2 deployment, which would allow users to have access to the different CLI commands with minimal setup (i.e. envars set appropriately). 

### Deployment

`seldonio/seldon-cli` Docker image has prepackaged Seldon CLI suitable for deployment in a Kubernetes cluster. Consider the following Kubernetes Deployment manifest: 

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: seldon-cli
spec:
  containers:
  - image: seldonio/seldon-cli:latest
    command:
      - tail
      - "-f"
      - "/dev/null"
    imagePullPolicy: IfNotPresent
    name: seldon-cli
    env:
    - name: KAFKA_SECURITY_PROTOCOL
      value: SSL
    - name: KAFKA_SASL_MECHANISM
      value: SCRAM-SHA-512
    - name: KAFKA_CLIENT_TLS_ENDPOINT_IDENTIFICATION_ALGORITHM
      value: ''
    - name: KAFKA_CLIENT_TLS_SECRET_NAME                       
      value: seldon
    - name: KAFKA_CLIENT_TLS_KEY_LOCATION
      value: /tmp/certs/kafka/client/user.key
    - name: KAFKA_CLIENT_TLS_CRT_LOCATION
      value: /tmp/certs/kafka/client/user.crt
    - name: KAFKA_CLIENT_TLS_CA_LOCATION
      value: /tmp/certs/kafka/client/ca.crt
    - name: KAFKA_CLIENT_SASL_USERNAME
      value: seldon
    - name: KAFKA_CLIENT_SASL_SECRET_NAME
      value: ''
    - name: KAFKA_CLIENT_SASL_PASSWORD_LOCATION
      value: password
    - name: KAFKA_BROKER_TLS_SECRET_NAME
      value: seldon-cluster-ca-cert
    - name: KAFKA_BROKER_TLS_CA_LOCATION                        
      value: /tmp/certs/kafka/broker/ca.crt
    - name: SELDON_KAFKA_CONFIG_PATH
      value: /mnt/kafka/kafka.json
    # The following environment variables are used to configure the seldon-cli from the already existing environment variables
    - name: SELDON_SCHEDULE_HOST
      value: $(SELDON_SCHEDULER_SERVICE_HOST):$(SELDON_SCHEDULER_SERVICE_PORT_SCHEDULER)
    # this envar is used for both TLS and seldon-cli
    - name: POD_NAMESPACE
      valueFrom:
        fieldRef:
          fieldPath: metadata.namespace
    volumeMounts:
      - mountPath: /mnt/kafka
        name: kafka-config-volume
  volumes:
  - configMap:
      name: seldon-kafka
    name: kafka-config-volume
  restartPolicy: Always
  serviceAccountName: seldon-scheduler
```

which can be deployed with 
```bash
kubectl apply -f <file> -n <namespace>
```
**Notes:**
- The above manifest sets up envars with the correct ip/port for the scheduler control plane (`SELDON_SCHEDULE_HOST`).
- It mounts `seldon-kafka` ConfigMap in `/mnt/kafka/kafka.json` so that Seldon CLI reads the relevant details from it (specifically the kafka bootstrap servers ip/port). 
- It also sets Kafka TLS envars, which should be identical to `seldon-modelgateway` deployment.
- Service account `seldon-scheduler` is required to access Kafka TLS secrets if any.

### Usage 

In this case users can run any Seldon CLI commands via `kubectl exec` utility, For example:

```bash
kubectl exec -n seldon-cli -n <namespace> -- seldon pipeline list
```

In particular this mode is also useful for inspecting data in kafka topics that form a pipeline. This is enabled because all the relevant arguments are setup so that Seldon CLI can consume data from these topics using the same settings used in the corresponding Core 2 deployment in the same namespace.

check `operator/config/cli` for a helper script and example manifest.




## Environment Variables for Services

The CLI talks to 3 backend services on default endpoints:

1. The Seldon Core 2 Scheduler: default `0.0.0.0:9004`
2. The Seldon Core inference endpoint: default `0.0.0.0:9000`
3. The Seldon Kafka broker: default: `0.0.0.0:9092`

These defaults will be correct when Seldon Core 2 is installed locally as per the docs. For Kubernetes, you will need to change these by defining environment variables for the following.

```bash
	SELDON_INFER_HOST
	SELDON_KAFKA_BROKER
	SELDON_SCHEDULE_HOST
```

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

You can create a config file to manage connections to running Seldon Core 2 installs. The settings will override any environment variable settings.

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

type KafkaConfig struct {
	Bootstrap    string `json:"bootstrap,omitempty"`
	Namespace    string `json:"namespace,omitempty"`
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
  }
}
```
## Additional Kafka Consumer Config (for `pipeline inspect`)

Additional Kafka configuration specifying Kafka broker and relevant consumer config can be specified as path to a json formatted file. This is currently only supported for inspecting data in topics that form a specific pipeline by passing `--kafka-config-path`. For example:

```json
{
  "bootstrap.servers": "seldon-kafka-bootstrap.seldon-mesh.svc.cluster.local:9093",
  "consumer": {
    "message.max.bytes": "1000000000",
  },
  "topicPrefix": "seldon"
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


