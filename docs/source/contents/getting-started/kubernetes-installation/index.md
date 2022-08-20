# Kubernetes Installation

You will need a running Kubernetes cluster or can create a local KinD one for testing via [Ansible](ansible.md).

You will also need to install our ecosystem components. For this we provide directions for [Ansible](ansible.md) to install these which includes Kafka, Prometheus, OpenTelemetry and Jeager. These are all optional except for Kafka if you wish to use Pipelines.

To install Seldon Core V2 itself you can choose:

 * [Helm chart](helm.md)
 * [Raw yaml](raw.md)
 * [Ansible](ansible.md)
 
The Kubernetes operator that is installed runs in namespaced mode so any resources you create need to be in the same namespace as you installed into.

## Operations

 * [TLS](tls.md)

```{toctree}
:maxdepth: 1
:hidden:

helm.md
raw.md
ansible.md
```
