---
description: >-
  This example illustrates how to use taints,  tolerations with nodeAffinity or
  nodeSelector to assign GPU nodes to specific models.
---

# Example: Serving models on dedicated GPU nodes

{% hint style="info" %}
**Note**: Configuration options depend on your cluster setup and the desired outcome. The Seldon CRDs for Seldon Core 2 Pods offer complete customization of Pod specifications, allowing you to apply additional Kubernetes customizations as needed.
{% endhint %}

To serve a model on a dedicated GPU node, you should follow these steps:

1. [Configuring the node](example-serving-models-on-dedicated-gpu-nodes.md#configuring-the-gpu-node)
2. [Configuring inference servers ](example-serving-models-on-dedicated-gpu-nodes.md#configure-inference-servers)
3. [Configuring models ](example-serving-models-on-dedicated-gpu-nodes.md#configuring-models)



### Configuring the GPU node

{% hint style="info" %}
**Note**: To dedicate a set of nodes to run only a specific group of inference servers, you must first provision an additional set of nodes within the Kubernetes cluster for the remaining Seldon Core 2 components. For more information about adding labels and taint to the GPU nodes in your Kubernetes cluster refer to the respective cloud provider documentation.
{% endhint %}

\
You can add the taint when you are creating the node or after the node has been provisioned. You can apply the same taint to multiple nodes, not just a single node. A common approach is to define the taint at the node pool level.&#x20;

When you apply a `NoSchedule` taint to a node after it is created it may result in existing Pods that do not have a matching toleration to remain on the node without being evicted. To ensure that such Pods are removed, you can use the `NoExecute` taint effect instead.&#x20;

\
In this example, the node includes several labels that are used later for node affinity settings. You may choose to specify some labels, while others are usually added by the cloud provider or a GPU operator installed in the cluster. \


```yaml

apiVersion: v1
kind: Node
metadata:
  name: example-node         # Replace with the actual node name
  labels:
    pool: infer-srv          # Custom label
    nvidia.com/gpu.product: A100-SXM4-40GB-MIG-1g.5gb-SHARED  # Sample label from GPU discovery
    cloud.google.com/gke-accelerator: nvidia-a100-80gb      # GKE without NVIDIA GPU operator
    cloud.google.com/gke-accelerator-count: "2"              # Accelerator count
spec:
  taints:
    - effect: NoSchedule
      key: seldon-gpu-srv
      value: "true"
```

## Configure inference servers

\
To ensure a specific inference server Pod runs only on the nodes you've configured, you can use `nodeSelector` or `nodeAffinity` together with a `toleration` by modifying one of the following:

* **Seldon Server custom resource**: Apply changes to each individual inference server.
* **ServerConfig custom resource**:  Apply settings across multiple inference servers at once.

**Configuring Seldon Server custom resource**\
While `nodeSelector` requires an exact match of node labels for server Pods to select a node, `nodeAffinity` offers more fine-grained control. It enables a conditional approach by using logical operators in the node selection process. For more information, see [Affinity and anti-affinity](https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/#affinity-and-anti-affinity).

{% tabs %}
{% tab title="nodeSelector" %}
In this example, a `nodeSelector` and a `toleration` is set for the Seldon Server custom resource.&#x20;

```yaml

apiVersion: mlops.seldon.io/v1alpha1
kind: Server
metadata:
  name: mlserver-llm-local-gpu     # <server name>
  namespace: seldon-mesh            # <seldon runtime namespace>
spec:
  replicas: 1
  serverConfig: mlserver            # <reference Serverconfig CR>
  extraCapabilities:
    - model-on-gpu                  # Custom capability for matching Model to this server
  podSpec:
    nodeSelector:                   # Schedule pods only on nodes with these labels
      pool: infer-srv
      cloud.google.com/gke-accelerator: nvidia-a100-80gb  # Example requesting specific GPU on GKE
      # cloud.google.com/gke-accelerator-count: 2          # Optional GPU count
    tolerations:                    # Allow scheduling on nodes with the matching taint
      - effect: NoSchedule
        key: seldon-gpu-srv
        operator: Equal
        value: "true"
    containers:                     # Override settings from Serverconfig if needed
      - name: mlserver
        resources:
          requests:
            nvidia.com/gpu: 1       # Request a GPU for the mlserver container
            cpu: 40
            memory: 360Gi
            ephemeral-storage: 290Gi
          limits:
            nvidia.com/gpu: 2       # Limit to 2 GPUs
            cpu: 40
            memory: 360Gi

					
```
{% endtab %}

{% tab title="nodeAffinity" %}
In this example, a `nodeAffinity` and a `toleration` is set for the Seldon Server custom resource.&#x20;

```yaml

apiVersion: mlops.seldon.io/v1alpha1
kind: Server
metadata:
  name: mlserver-llm-local-gpu     # <server name>
  namespace: seldon-mesh            # <seldon runtime namespace>
spec:
  podSpec:
    affinity:
      nodeAffinity:
        requiredDuringSchedulingIgnoredDuringExecution:
          nodeSelectorTerms:
          - matchExpressions:
            - key: "pool"
              operator: In
              values:
              - infer-srv
            - key: "cloud.google.com/gke-accelerator"
              operator: In
              values:
              - nvidia-a100-80gb
    tolerations:                     # Allow mlserver-llm-local-gpu pods to be scheduled on nodes with the matching taint
    - effect: NoSchedule
      key: seldon-gpu-srv
      operator: Equal
      value: "true"
    containers:                      # If needed, override settings from ServerConfig for this specific Server
      - name: mlserver
        resources:
          requests:
            nvidia.com/gpu: 1        # Request a GPU for the mlserver container
            cpu: 40
            memory: 360Gi
            ephemeral-storage: 290Gi
          limits:
            nvidia.com/gpu: 2        # Limit to 2 GPUs
            cpu: 40
            memory: 360Gi

					
```

You can configure more advanced Pod selection using `nodeAffinity`, as in this example:

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Server
metadata:
  name: mlserver-llm-local-gpu     # <server name>
  namespace: seldon-mesh            # <seldon runtime namespace>
spec:
  podSpec:
    affinity:
      nodeAffinity:
        requiredDuringSchedulingIgnoredDuringExecution:
          nodeSelectorTerms:
            - matchExpressions:
              - key: "cloud.google.com/gke-accelerator-count"
                operator: Gt       # (greater than)
                values: ["1"]
              - key: "gpu.gpu-vendor.example/installed-memory"
                operator: Gt
                values: ["75000"]
              - key: "feature.node.kubernetes.io/pci-10.present" # NFD Feature label
                operator: In
                values: ["true"] # (optional) only schedule on nodes with PCI device 10

    tolerations:                     # Allow mlserver-llm-local-gpu pods to be scheduled on nodes with the matching taint
    - effect: NoSchedule
      key: seldon-gpu-srv
      operator: Equal
      value: "true"

    containers:                      # If needed, override settings from ServerConfig for this specific Server
      - name: mlserver
        env:
          ...                        # Add your environment variables here
        image: ...                   # Specify your container image here
        resources:
          requests:
            nvidia.com/gpu: 1        # Request a GPU for the mlserver container
            cpu: 40
            memory: 360Gi
            ephemeral-storage: 290Gi
          limits:
            nvidia.com/gpu: 2        # Limit to 2 GPUs
            cpu: 40
            memory: 360Gi
        ...                           # Other configurations can go here

```
{% endtab %}
{% endtabs %}

**Configuring ServerConfig custom resource**

This configuration automatically affects all servers using that `ServerConfig`, unless you specify server-specific overrides, which takes precedence.

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: ServerConfig
metadata:
  name: mlserver-llm              # <ServerConfig name>
  namespace: seldon-mesh           # <seldon runtime namespace>
spec:
  podSpec:
    nodeSelector:                  # Schedule pods only on nodes with these labels
      pool: infer-srv
      cloud.google.com/gke-accelerator: nvidia-a100-80gb  # Example requesting specific GPU on GKE
      # cloud.google.com/gke-accelerator-count: 2          # Optional GPU count
    tolerations:                   # Allow scheduling on nodes with the matching taint
      - effect: NoSchedule
        key: seldon-gpu-srv
        operator: Equal
        value: "true"
    containers:                    # Define the container specifications
      - name: mlserver
        env:                       # Environment variables (fill in as needed)
          ...
        image: ...                 # Specify the container image
        resources:
          requests:
            nvidia.com/gpu: 1      # Request a GPU for the mlserver container
            cpu: 40
            memory: 360Gi
            ephemeral-storage: 290Gi
          limits:
            nvidia.com/gpu: 2      # Limit to 2 GPUs
            cpu: 40
            memory: 360Gi
        ...                        # Additional container configurations

```

### Configuring models&#x20;

When you have a set of inference servers running exclusively on GPU nodes, you can assign a model to one of those servers in two ways:

* Custom model requirements (recommended)
* Explicit server pinning

Here's the distinction between the two methods of assigning models to servers.

| Method                        | Behavior                                                                                                                                        |
| ----------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------- |
| **Custom model requirements** | If the assigned server cannot load the model due to insufficient resources, another similarly-capable server can be selected to load the model. |
| **Explicit pinning**          | If the specified server lacks sufficient memory or resources, the model load fails without trying another server.                               |

{% tabs %}
{% tab title="Custom model requirements" %}
When you specify a requirement matching a server capability in the model custom resource it loads the model on any inference server with a capability matching the requirements.

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Model
metadata:
  name: llama3           # <model name>
  namespace: seldon-mesh # <seldon runtime namespace>
spec:
  requirements:
  - model-on-gpu         # requirement matching a Server capability

```

Ensure that the additional capability that matches the requirement label is added to the Server custom resource.&#x20;

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Server
metadata:
  name: mlserver-llm-local-gpu     # <server name>
  namespace: seldon-mesh           # <seldon runtime namespace>
spec:
  serverConfig: mlserver           # <reference ServerConfig CR>
  extraCapabilities:
    - model-on-gpu                 # custom capability that can be used for matching Model to this server
  # Other fields would go here
```

Instead of adding a capability using `extraCapabilities` on a Server custom resource, you may also add to the list of capabilities in the associated ServerConfig custom resource. This applies to all servers referencing the configuration.

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: ServerConfig
metadata:
  name: mlserver-llm               # <ServerConfig name>
  namespace: seldon-mesh           # <seldon runtime namespace>
spec:
  podSpec:
    containers:
      - name: agent                # note the setting is applied to the agent container
        env:
          - name: SELDON_SERVER_CAPABILITIES
            value: mlserver,alibi-detect,...,xgboost,model-on-gpu  # add capability to the list
        image: ...
    # Other configurations go here
```
{% endtab %}

{% tab title="Explicit pinning" %}
With these specifications, the model is loaded on replicas of inference servers created by the referenced Server custom resource.

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Model
metadata:
  name: llama3           # <model name>
  namespace: seldon-mesh # <seldon runtime namespace>
spec:
  server: mlserver-llm-local-gpu   # <reference Server CR>
  requirements:
    - model-on-gpu                # requirement matching a Server capability

```
{% endtab %}
{% endtabs %}

