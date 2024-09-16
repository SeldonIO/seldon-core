---
description: >-
  This example illustrates how to use taints, tolerations, and node affinity to
  assign GPU nodes to specific models.
---

# Example: Serving models on dedicated GPU nodes

{% hint style="info" %}
**Note**: Configuration options depend on your cluster setup and the desired outcome. The Seldon CRDs for Seldon Core 2 Pods offer complete customization of Pod specifications, allowing you to apply additional Kubernetes customizations as needed.
{% endhint %}

Serving models on dedicated GPU nodes:

1. [Configuring inference servers ](example-serving-models-on-dedicated-gpu-nodes.md#configure-inference-servers)
2. Configuring models&#x20;

## Configure inference servers

{% hint style="info" %}
To dedicate a set of nodes to run only a specific group of inference servers, you must first provision an additional set of nodes within the Kubernetes cluster for the remaining Seldon Core 2 components.
{% endhint %}

1.  Add taint to the GPU node.\
    You can add the taint when you are creating the node or after the node has been provisioned. You can apply the same taint to multiple nodes, not just a single node. A common approach is to define the taint at the node pool level.\
    **Note:**  When you apply a `NoSchedule` taint to a node after it is created it may result in existing Pods that do not have a matching toleration to remain on the node without being evicted. To ensure that such Pods are removed, you can use the `NoExecute` taint effect instead. \
    In this example, the node includes several labels that can be used later for node affinity settings. You may need to specify some labels, while others are usually added by the cloud provider or a GPU operator installed in the cluster.\


    ```yaml

    apiVersion: v1
    kind: Node
    metadata:
    	...
    	labels:
    		...
    		# manually-added labels
    		**pool: infer-srv**    # sample custom label, could be any key-value pair
    		**...**
    		# other labels, perhaps added by the cloud provider or the NVIDIA GPU operator
    		# you wouldn't typically see all of those at the same time
    		**nvidia.com/gpu.product: A100-SXM4-40GB-MIG-1g.5gb-SHARED** # sample label as added by gpu-feature-discovery when using the NVIDIA GPU Operator
    		**cloud.google.com/gke-accelerator: nvidia-a100-80gb**  # GKE without NVIDIA GPU operator
    		**cloud.google.com/gke-accelerator-count: 2**		
    spec:
    	...
    	**taints:
    	- effect: NoSchedule
    		key: seldon-gpu-srv
    		value: "true"**
    ```

