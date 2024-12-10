# Kubernetes Installation

## Prerequisites

* Ensure that the version of the Kubernetes cluster is v1.27 or later. You can create a [KinD](https://kind.sigs.k8s.io/docs/user/quick-start/#installation) cluster on your local computer for testing with [Ansible](ansible.md). 
* Install the ecosystem components using [Ansible](ansible.md).

## Core 2 Dependencies

Here is a list of components that Seldon Core 2 depends on, with minimum and maximum supported versions.

| Component | Minimum Version | Maximum Version | Notes |
| - | - | - | - |
| Kubernetes | 1.27 | 1.31 | Required |
| Envoy`*` | 1.32.2 | 1.32.2 | Required |
| Rclone`*` | 1.68.2 | 1.68.2 | Required |
| Kafka`**` | 3.4 | 3.8 | Optional |
| Prometheus | 2.0 | 2.x | Optional |
| Grafana | 10.0 | `***` | Optional |
| Prometheus-adapter | 0.12 | 0.12 | Optional |
| Opentelemetry Collector | 0.68 | `***` | Optional |

`*` These components are shipped as part of Seldon Core 2 docker images set, users should not install them separately but they need to be aware of the configuration options that are supported by these versions.
`**` Kafka is only required to operate Seldon Core 2 dataflow Pipelines. If not required then users should not install seldon-modelgateway, seldon-pipelinegateway, seldon-dataflow-engine.
`***` Not hard limit on the maximum version to be used.


## Install Ecosystem Components

You also need to install our ecosystem components. For this we provide directions for [Ansible](ansible.md) to install these.

| Component  | Summary |
| - | - |
| Kafka | Required for inference Pipeline usage. |
| Prometheus | (Optional) Exposes metrics. |
| Grafana | (Optional) UI for metrics. |
| OpenTelemetry | (Optional) Exposes tracing. |
| Jaeger | (Optional) UI for traces. |


### Install

To install Seldon Core 2 from the [source repository](https://github.com/SeldonIO/seldon-core), you can choose one of the following methods:

* [Helm](helm.md)(recommended for production systems)
* [Ansible](ansible.md)(recommended for testing, development, or trial)

The Kubernetes operator that is installed runs in namespaced mode so any resources you create
need to be in the same namespace as you installed into.

### Kustomize

Seldon recommends installing Seldon Core 2 using Helm or Ansible. If you prefer to use Kustomize, you can base your configuration on the raw YAML files you generate in the k8s/yaml folder. Alternatively, you can follow the steps in the k8s/Makefile, which demonstrates how to build the YAML files from the Kustomize bases.

## Operations

* [Security](security/README.md)
