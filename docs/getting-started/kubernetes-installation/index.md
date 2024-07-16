# Kubernetes Installation

## Prerequisites

### A running Kubernetes Cluster

You will need a running Kubernetes cluster or can create a local KinD one for testing via [Ansible](ansible.md).

### Install Ecosystem Components

You will also need to install our ecosystem components. For this we provide directions for [Ansible](ansible.md) to install these.

```{list-table}
:header-rows: 1

* - Component
  - Summary
* - Kafka
  - Required for inference Pipeline usage.
* - Prometheus
  - (Optional) Exposes metrics.
* - Grafana
  - (Optional) UI for metrics.
* - OpenTelemetry
  - (Optional) Exposes tracing.
* - Jaeger
  - (Optional) UI for traces.

```

### Install

To install Seldon Core V2 itself you can choose from the following. At present, all require a clone of the source repository.

 * [Helm Installation](helm.md) (recommended for production systems)
 * [Ansible](ansible.md) (recommended for test / dev / trial purposes)

The Kubernetes operator that is installed runs in namespaced mode so any resources you create need to be in the same namespace as you installed into.



### Kustomize

Our recommended and supported way to install is via Helm or Ansible. If you wish to use Kustomize then you can base your configuration on the raw yaml we create in the folder `k8s/yaml` or follow the steps in `k8s/Makefile` which illustrate how we build this yaml from our own Kustomize bases.

## Operations

 * [Security](security/index.md)

```{toctree}
:maxdepth: 1
:hidden:

helm.md
ansible.md
security/index.md
```
