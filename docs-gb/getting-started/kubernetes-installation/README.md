# Kubernetes Installation

## Prerequisites

* Ensure that the version of the Kubernetes cluster is v1.27 or later. Seldon Core 2 supports Kubernetes versions 1.27, 1.28, 1.29, 1.30, and 1.31. You can create a [KinD](https://kind.sigs.k8s.io/docs/user/quick-start/#installation) cluster on your local computer for testing with [Ansible](ansible.md). 
* Install the ecosystem components using [Ansible](ansible.md).

## Install Ecosystem Components¶

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
