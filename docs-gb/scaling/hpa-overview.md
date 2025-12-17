---
description: Overview of Horizontal Pod Autoscaler (HPA) scaling options in Seldon Core 2
---

# Using HPA for Autoscaling

Given Seldon Core 2 is predominantly for serving ML in Kubernetes, it is possible to leverage `HorizontalPodAutoscaler` or [HPA](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/) to define scaling logic that automatically scales up and down Kubernetes resources. HPA targets Kubernetes or custom metrics to trigger scale-up or scale-down events for specified resources. Using HPA is recommnended if **custom Scaling Metrics** are required. These would be exposed using Prometheus, and [Prometheus Adapter](https://github.com/kubernetes-sigs/prometheus-adapter) or similar tools for explosing metrics to HPA. If these tools cause conflicts, [autoscaling functionality native to Core 2](core-autoscaling.md) that does not require exposing custom metrics is recommended.

Seldon Core 2 provides two main approaches to leveraging Kubernetes Horizontal Pod Autoscaler (HPA) for autoscaling. It is important to remember that since in Core 2 Models and Servers are separate, autoscaling of _both_ Models and Servers, in a coordinated way, needs to be accounted for when implementing autoscaling. In order to implement either approach, metrics first need to be exposed - this is explained in the [HPA Setup](hpa-setup.md) guide which explains the fundamental requirements and configuration needed to enable HPA-based scaling in Seldon Core 2.

## 1. Model Autoscaling with HPA

The [Model Autoscaling with HPA](model-hpa-autoscaling.md) approach enables users to scale Models based on custom metrics. This approach, along with [Server Autoscaling](core-autoscaling-servers.md), enables users to customize the scaling logic for models, and automate the scaling of Servers based on the needs of the Models hosted on them.

![Model Autoscaling with HPA, Servers autoscaled by Core 2](../.gitbook/assets/model-hpa-server-autoscaled.png)

## 2. Model and Server Autoscaling with HPA

The [Model and Server Autoscaling with HPA](single-model-serving-hpa.md) approach leverages HPA to autoscale for Models and Servers in a coordinated way. This requires a **1-1 Mapping of Models and Servers** (no Multi-Model Serving). In this case, setting up HPA can be set up for a Model and its associated Server, targetting the same _custom_ metric (this is possible for Kubernetes-native metrics).

![Model and Server autoscaling with HPA, for single-model serving](../.gitbook/assets/model-server-hpa-scaling.png)

s
