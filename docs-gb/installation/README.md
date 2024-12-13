---
icon: screwdriver-wrench
  title:
    visible: true
  description:
    visible: false
  tableOfContents:
    visible: true
  outline:
    visible: true
  pagination:
    visible: true
---

# Installation
Seldon Core 2 can be installed in various setups to suit different stages of the development lifecycle. The most common modes include:

## Local Environment
Ideal for development and testing purposes, a local setup allows for quick iteration and experimentation with minimal overhead. Common tools include:
* Docker Compose:
Simplifies deployment by orchestrating Seldon Core components and dependencies in Docker containers. Suitable for environments without Kubernetes, providing a lightweight alternative.
* Kind (Kubernetes IN Docker):
Runs a Kubernetes cluster inside Docker, offering a realistic testing environment.
Ideal for experimenting with Kubernetes-native features.
* Orbstack:
Optimized for macOS, Orbstack offers lightweight virtualized environments to test Kubernetes workloads.This mode emphasizes simplicity and fast iteration cycles, making it an excellent choice for developers and data scientists building and testing models.
## Production Environment
Designed for high-availability and scalable deployments, a production setup ensures security, reliability, and resource efficiency. Typical tools and setups include:
* Fully-Managed Kubernetes Clusters:
Platforms like GKE (Google Kubernetes Engine), EKS (Amazon Elastic Kubernetes Service), and AKS (Azure Kubernetes Service) provide managed Kubernetes solutions.
Suitable for enterprises requiring scalability and cloud integration.
* On-Premises Kubernetes:
For organizations with strict compliance or data sovereignty requirements.
Can be deployed on platforms like OpenShift or custom Kubernetes setups.
* Advanced Networking and Scaling:
Requires an ingress controller for routing external traffic. Incorporates monitoring stacks like Prometheus and Grafana to manage workloads.
Production setups prioritize robust features like rolling updates, canary deployments, and disaster recovery to ensure smooth operations at scale.

By selecting the appropriate installation mode—whether it's Docker Compose for simplicity, Kind for local Kubernetes experimentation, or production-grade Kubernetes for scalability—you can effectively leverage Seldon Core 2 to meet your specific needs.

### Get started

<table data-view="cards"><thead><tr><th></th><th></th><th data-hidden data-card-cover data-type="files"></th><th data-hidden></th><th data-hidden data-card-target data-type="content-ref"></th></tr></thead><tbody><tr><td><strong>Local environment</strong></td><td>Install Seldon Core 2 in Docker Compose, Kind, or Orbstack </td><td></td><td></td><td><a href="learning-environment/README.md">README.md</a></td></tr><tr><td><strong>Production environment</strong></td><td>Install Seldon COre 2 in a Managed Kubernetes cluster, or On-Premises Kubernetes cluster</td><td></td><td></td><td><a href="production-environment/README.md">README.md</a></td></tr></tbody></table>
