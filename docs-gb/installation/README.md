
# Installation
Seldon Core 2 can be installed in various setups to suit different stages of the development lifecycle. The most common modes include:

## Local Environment
Ideal for development and testing purposes, a local setup allows for quick iteration and experimentation with minimal overhead. Common tools include:
* Docker Compose:
Simplifies deployment by orchestrating Seldon Core components and dependencies in Docker containers. Suitable for environments without Kubernetes, providing a lightweight alternative.
* Kind (Kubernetes IN Docker):
Runs a Kubernetes cluster inside Docker, offering a realistic testing environment.
Ideal for experimenting with Kubernetes-native features.

## Production Environment
Designed for high-availability and scalable deployments, a production setup ensures security, reliability, and resource efficiency. Typical tools and setups include:
* Managed Kubernetes Clusters:
Platforms like GKE (Google Kubernetes Engine), EKS (Amazon Elastic Kubernetes Service), and AKS (Azure Kubernetes Service) provide managed Kubernetes solutions.
Suitable for enterprises requiring scalability and cloud integration.
* On-Premises Kubernetes Clusters:
For organizations with strict compliance or data sovereignty requirements.
Can be deployed on platforms like OpenShift or custom Kubernetes setups.

By selecting the appropriate installation mode—whether it's Docker Compose for simplicity, Kind for local Kubernetes experimentation, or production-grade Kubernetes for scalability—you can effectively leverage Seldon Core 2 to meet your specific needs.

## Helm Charts

| **Name of the Helm Chart**              | **Description**                                                                                                                                              |
|-----------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `seldon-core-v2-crds`      | Cluster-wide installation of custom resources.                                                                                                              |
| `seldon-core-v2-setup`     | Installation of the manager to manage resources in the namespace or cluster-wide. This also installs default `SeldonConfig` and `ServerConfig` resources, allowing *Runtimes* and *Servers* to be installed on demand. |
| `seldon-core-v2-runtime`   | Installs a `SeldonRuntime` custom resource that creates the core components in a namespace.                                                                 |
| `seldon-core-v2-servers`   | Installs `Server` custom resources providing example core servers to load models.                                                                            |
                                                                    

For more information, see the published [Helm charts](https://github.com/SeldonIO/helm-charts).

For the description of (some) values that can be configured for these charts, see this [helm parameters section](helm/README.md).

## Seldon Core 2 Dependencies

Here is a list of components that Seldon Core 2 requires, along with the minimum and maximum supported versions:

| **Component**              | **Minimum Version** | **Maximum Version** | **Notes**                                                                                                                                               |
|-----------------------------|---------------------|---------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------|
| **Kubernetes**             | 1.27               | 1.33.0               | Required                                                                                                                                                |
| **Envoy***                 | 1.32.2             | 1.32.2             | Required                                                                                                                                                |
| **Rclone***                | 1.68.2             | 1.69.0            | Required                                                                                                                                                |
| **Kafka**                  | 3.4                | 3.8                | Recommended (only required for operating Seldon Core 2 dataflow Pipelines)                                                                             |
| **Prometheus**             | 2.0                | 2.x                | Optional                                                                                                                                                |
| **Grafana**                | 10.0               | ***                | Optional (no hard limit on the maximum version to be used)                                                                                              |
| **Prometheus-adapter**     | 0.12               | 0.12               | Optional                                                                                                                                                |
| **Opentelemetry Collector**| 0.68               | ***                | Optional (no hard limit on the maximum version to be used)                                                                                              |

**Notes**:
- **Envoy** and **Rclone**: These components are included as part of the Seldon Core 2 Docker images. You are not required to install them separately but must be aware of the configuration options supported by these versions.
- **Kafka**: Only required for operating Seldon Core 2 dataflow Pipelines. If not needed, you should avoid installing `seldon-modelgateway`, `seldon-pipelinegateway`, and `seldon-dataflow-engine`.
- **Maximum Versions** marked with `***` indicates no hard limit on the version that can be used.


### Get started

<table data-view="cards"><thead><tr><th></th><th></th><th data-hidden data-card-cover data-type="files"></th><th data-hidden></th><th data-hidden data-card-target data-type="content-ref"></th></tr></thead><tbody>
<tr><td><strong>Learning environment</strong></td>
<td>Install Seldon Core 2 in Docker Compose, or Kind </td>
<td></td><td></td><td><a href="learning-environment/README.md">README.md</a></td></tr>
<tr><td><strong>Production environment</strong></td>
<td>Install Seldon Core 2 in a Managed Kubernetes cluster, or On-Premises Kubernetes cluster</td>
<td></td><td></td>
<td><a href="production-environment/README.md">README.md</a></td></tr>
</tbody></table>
