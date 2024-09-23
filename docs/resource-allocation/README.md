---
description: >-
  Learn more about using taints and tolerations with node affinity or node
  selector to allocate resources in a Kubernetes cluster.
---

# Resource allocation

When deploying machine learning models in Kubernetes, you may need to control which infrastructure resources these models use. This is especially important in environments where certain workloads, such as resource-intensive models, should be isolated from others or where specific hardware such as  GPUs, needs to be dedicated to particular tasks. Without fine-grained control over workload placement, models might end up running on suboptimal nodes, leading to inefficiencies or resource contention.

For example, you may want to:

* Isolate inference workloads from control plane components or other services to prevent resource contention.
* Ensure that GPU nodes are reserved exclusively for models that require hardware acceleration.
* Keep business-critical models on dedicated nodes to ensure performance and reliability.
* Run external dependencies like Kafka on separate nodes to avoid interference with inference workloads.

To solve these problems, Kubernetes provides mechanisms such as taints, tolerations, and `nodeAffinity` or `nodeSelector` to control resource allocation and workload scheduling.&#x20;

[Taints ](https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/)are applied to nodes and [tolerations](https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/) to Pods to control which Pods can be scheduled on specific nodes within the Kubernetes cluster. Pods without a matching toleration for a node’s taint are scheduled on that node. For instance, if a node has GPUs or other specialized hardware, you can prevent Pods that don’t need these resources from running on that node to avoid unnecessary resource usage.

{% hint style="info" %}
[Taints and tolerations](https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/) alone do not ensure that a Pod runs on a tainted node. Even if a Pod has the correct toleration, Kubernetes may still schedule it on other nodes without taints. To ensure a Pod runs on a specific node, you need to also use [node affinity](https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/#affinity-and-anti-affinity) and [node selector](https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/#nodeselector) rules.
{% endhint %}

When used together, taints and tolerations with `nodeAffinity` or `nodeSelector` can effectively allocate certain Pods to specific nodes, while preventing other Pods from being scheduled on those nodes.

In a Kubernetes cluster running Seldon Core 2, this involves two key configurations:

1. Configuring servers with specific nodes using mechanisms like taints, tolerations, and `nodeAffinity` or `nodeSelector`.
2. Configuring models so that they are scheduled and loaded on the appropriate servers.&#x20;

This ensures that models are deployed on the optimal infrastructure and servers that meet their requirements.

