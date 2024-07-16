# Models

Models provide the atomic building blocks of Seldon. They represents machine learning models, drift detectors, outlier detectors, explainers, feature transformations, and more complex routing models such as multi-armed bandits.

 * Seldon can handle a wide range of [inference artifacts](./inference-artifacts/index.md)
 * Artifacts can be stored on any of the 40 or more cloud storage technologies as well as from local (mounted) folder as discussed [here](./rclone/index.md).

## Kubernetes Example

A Kubernetes yaml example is shown below for a SKLearn model for iris classification:

```{literalinclude} ../../../../samples/models/sklearn-iris-gs.yaml 
:language: yaml
```

Its Kubernetes `spec` has two core requirements

 * A `storageUri` specifying the location of the artifact. This can be any rclone URI specification.
 * A `requirements` list which provides tags that need to be matched by the Server that can run this artifact type. By default when you install Seldon we provide a set of Servers that cover a range of artifact types.


## GRPC Example

You can also load models directly over the scheduler grpc service. An example is shown below use grpcurl tool:

```bash
!grpcurl -d '{"model":{ \
              "meta":{"name":"iris"},\
              "modelSpec":{"uri":"gs://seldon-models/mlserver/iris",\
                           "requirements":["sklearn"],\
                           "memoryBytes":500},\
              "deploymentSpec":{"replicas":1}}}' \
         -plaintext \
         -import-path ../../apis \
         -proto apis/mlops/scheduler/scheduler.proto  0.0.0.0:9004 seldon.mlops.scheduler.Scheduler/LoadModel
```

The proto buffer definitions for the scheduler are outlined [here](../apis/scheduler/index.md).

```{toctree}
:maxdepth: 1
:hidden:

inference-artifacts/index.md
rclone/index.md
parameterized-models/index.md
```

## Multi-model Serving with Overcommit

Multi-model serving is an architecture pattern where one ML inference server hosts multiple models at the same time. It is a feature provided out of the box by Nvidia Triton and Seldon MLServer. Multi-model serving reduces infrastructure hardware requirements (e.g. expensive GPUs) which enables the deployment of a large number of models while making it efficient to operate the system at scale.

Seldon Core v2 leverages multi-model serving by design and it is the default option for deploying models. The system will find an appropriate server to load the model onto based on requirements that the user defines in the `Model` deployment definition.

Moreover, in many cases demand patterns allow for further Overcommit of resources. Seldon Core v2 is able to register more models than what can be served by the provisioned (memory) infrastructure and will swap models dynamically according to least used without adding significant latency overheads to inference workload.

See [Multi-model serving](mms/mms.md) for more information. 

## Autoscaling of Models

See [here](../kubernetes/autoscaling/index.md) for discussion of autoscaling of models.
