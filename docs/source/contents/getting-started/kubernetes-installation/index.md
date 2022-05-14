# Kubernetes Installation

You will need a running Kubernetes cluster or can create a local KinD one for testing via [Ansible](ansible.md).

You will also need to install our ecosystem components. For this we provide directions for [Ansible](ansible.md) to install these which includes Kafka, Prometheus, OpenTelemetry and Jeager. These are all optional except for Kafka if you wish to use Pipelines.

To install Seldon Core V2 itself you can choose:

 * [Helm chart](helm.md)
 * [Ansible](ansible.md)
 * [Raw yaml](raw.md)


```{toctree}
:maxdepth: 1
:hidden:

ansible.md
helm.md
raw.md
```
