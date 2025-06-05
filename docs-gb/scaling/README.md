---
---

# Autoscaling in Seldon Core 2

Seldon Core 2 provides multiple approaches to scaling your machine learning deployments, allowing you to optimize resource utilization and handle varying workloads efficiently. In Core 2, we separate out Models and Servers, and Servers can have multiple Models loaded on them (Multi-Model Serving). Given this, setting up autoscaling requires defining the logic by which you want to scale your Models and then configuring the autoscaling of Servers such that they autoscale in a coordinated way. The following steps can be followed to set up autoscaling based on specific requirements:

1. **Identify metrics** that you want to scale Models on. There are a couple of different options here:
    1. Core 2 natively supports scaling based on **Inference Lag**, meaning the difference between incoming and outgoing requests for a model in a given period of time.
    2. Users can expose **custom or Kubernetes-native metrics**, and then target the scaling of models based on those metrics by using `HorizontalPodAutoscaler`. This requires exposing the right metrics, using the monitoring tool of your choice (e.g. Prometheus).
2. **Implement Server Scaling** by either:
    1. Enabling Autoscaling of Servers based on Model needs. This is managed by Seldon's scheduler, and is enabled by setting `minReplicas` and `maxReplicas` in the Server Custom Resource.
    2. If Models and Servers are to have a one-to-one mapping (no Multi-Model Serving) then users can also define scaling of Servers using an HPA manifest that matches the HPA applied to the associated Models. This approach is outlined [here](./hpa-rps-autoscaling.md). This approach will only work with custom metrics, as Kubernetes does not allow mutliple HPAs to target the same metrics from Kubernetes directly.

The above options result in the following three options for coordinated autoscaling of Models and Servers:

![Seldon Core Autoscaling](core-model-server-autoscaling.png) ![HPA Autoscaling for Single-Model Serving](model-server-hpa-scaling.png) ![HPA Autoscaling for Models, Servers Autoscaled by Seldon Core](model-hpa-server-autscaled.png)


# Scaling Seldon Services

When running Core 2 at scale, it is important to understand the scaling behaviour of Seldon's services as well as the scaling of the Models and Servers themselves. This is outlined in the [Scaling Core Services](scaling-core-services.md) page.


