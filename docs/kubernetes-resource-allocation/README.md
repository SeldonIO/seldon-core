---
description: >-
  Learn more about using taints and tolerations with node affinity or node
  selector to allocate resources in a Kubernetes cluster.
---

# Kubernetes resource allocation

###



{% hint style="info" %}
[Taints and tolerations](https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/) alone do not ensure that a Pod will run on a tainted node. Even if a Pod has the correct toleration, Kubernetes may still schedule it on other nodes without taints. To ensure a Pod runs on a specific node, you need to also use [node affinity](https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/#affinity-and-anti-affinity) and [node selector](https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/#nodeselector) rules.
{% endhint %}

* Taints are applied to nodes and tolerations to Pods to control which Pods can be scheduled on specific nodes within the cluster. Pods without a matching toleration for a node’s taint will not be scheduled on that node. For instance, if a node has GPUs or other specialized hardware, you can prevent Pods that don’t need these resources from running on that node to avoid unnecessary resource usage.
* [Node affinity](https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/#affinity-and-anti-affinity) or [node selector](https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/#nodeselector) is used to ensure that specific Pods are scheduled on particular nodes, typically to meet the unique requirements of their workloads.

When used together, taints and tolerations with node affinity or node selector can effectively allocate certain Pods to specific nodes, while preventing other Pods from being scheduled on those nodes.

#### Use cases

In a Kubernetes cluster running Seldon Core 2, you might configure taints, tolerations, and node affinity for use cases such as:

* Isolating Seldon Core 2 control plane or data-plane routing Pods from those serving inference workloads.&#x20;
* Running resource-intensive services such as Kafka on separate nodes to prevent resource contention with inference workloads.
* &#x20;Assigning business-critical inference servers to dedicated nodes, isolating them from other servers.
* &#x20;Reserving GPU nodes solely for models that require hardware acceleration.
