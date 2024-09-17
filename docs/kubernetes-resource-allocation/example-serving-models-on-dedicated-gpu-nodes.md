---
description: >-
  This example illustrates how to use taints, tolerations with node affinity or
  node selector to assign GPU nodes to specific models.
---

# Example: Serving models on dedicated GPU nodes

{% hint style="info" %}
**Note**: Configuration options depend on your cluster setup and the desired outcome. The Seldon CRDs for Seldon Core 2 Pods offer complete customization of Pod specifications, allowing you to apply additional Kubernetes customizations as needed.
{% endhint %}

Serving models on dedicated GPU nodes:

1. [Configuring inference servers ](example-serving-models-on-dedicated-gpu-nodes.md#configure-inference-servers)
2. [Configuring models ](example-serving-models-on-dedicated-gpu-nodes.md#configuring-models)

## Configure inference servers

{% hint style="info" %}
**Note**: To dedicate a set of nodes to run only a specific group of inference servers, you must first provision an additional set of nodes within the Kubernetes cluster for the remaining Seldon Core 2 components. For more information about adding labels and taint to the GPU nodes in your Kubernetes cluster refer to the respective cloud provider documentation.
{% endhint %}

1.  Add taint to the GPU node.\
    You can add the taint when you are creating the node or after the node has been provisioned. You can apply the same taint to multiple nodes, not just a single node. A common approach is to define the taint at the node pool level. \
    **Note:**  When you apply a `NoSchedule` taint to a node after it is created it may result in existing Pods that do not have a matching toleration to remain on the node without being evicted. To ensure that such Pods are removed, you can use the `NoExecute` taint effect instead. \
    In this example, the node includes several labels that are used later for node affinity settings. You may need to specify some labels, while others are usually added by the cloud provider or a GPU operator installed in the cluster. \


    ```yaml

    apiVersion: v1
    kind: Node
    metadata:
    	...
    	labels:
    		...
    		# manually-added labels
    		pool: infer-srv    # sample custom label, could be any key-value pair
    		...
    		# other labels, perhaps added by the cloud provider or the NVIDIA GPU operator
    		# you wouldn't typically see all of those at the same time
    		nvidia.com/gpu.product: A100-SXM4-40GB-MIG-1g.5gb-SHARED # sample label as added by gpu-feature-discovery when using the NVIDIA GPU Operator
    		cloud.google.com/gke-accelerator: nvidia-a100-80gb  # GKE without NVIDIA GPU operator
    		cloud.google.com/gke-accelerator-count: 2		
    spec:
    	...
    	taints:
    	- effect: NoSchedule
    		key: seldon-gpu-srv
    		value: "true"
    ```
2. Configure inference server custom resources. \
   To ensure a specific inference server Pod runs only on the nodes you've configured, you can modify one of these:\
   \- **Seldon Server custom resource**: using `nodeSelector` or `nodeAffinity` with `toleration` in each of the inference servers.\
   \- **ServerConfig custom resource**: using `nodeSelector` or `nodeAffinity` with `toleration` across multiple inference servers.

**Configuring Seldon Server custom resource**\
While `nodeSelector` requires an exact match of node labels for the server Pods to select that node, `nodeAffinity` provides flexibility.

{% tabs %}
{% tab title="nodeSelector" %}
In this example, a `nodeSelector` and a `toleration` is set for the Seldon Server custom resource.&#x20;

```yaml

apiVersion: mlops.seldon.io/v1alpha1
kind: Server
metadata:
  name: mlserver-llm-local-gpu     # <server name>
  namespace: seldon-mesh           # <seldon runtime namespace>
spec:
  replicas: 1
  serverConfig: mlserver     # <reference Serverconfig CR>
  extraCapabilities:
    - model-on-gpu           # custom capability that can be used for matching Model to this server
  podSpec:
    nodeSelector:            # only run mlserver-llm-local-gpu pods on nodes that have all those labels  
      pool: infer-srv
      cloud.google.com/gke-accelerator: nvidia-a100-80gb  # example requesting specific GPU on GKE, not required
      # cloud.google.com/gke-accelerator-count: 2   # also request node with label denoting a specific GPU count
    ...
    tolerations:             # allow mlserver-llm-local-gpu pods to be scheduled on nodes with the matching taint
    - effect: NoSchedule
      key: seldon-gpu-srv
      operator: Equal
      value: "true"
    ...
    containers:              # if needed, override settings from Serverconfig, for this specific Server
	  - name: mlserver
		  resources:
			  requests:
				  nvidia.com/gpu: 1  # in particular, have the mlserver container request a GPU
				  cpu: 40
				  memory: 360Gi
				  ephemeral-storage: 290Gi
				limits:
					nvidia.com/gpu: 2
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
  namespace: seldon-mesh           # <seldon runtime namespace>
spec:
  replicas: 1
  serverConfig: mlserver     # <reference Serverconfig CR>
  extraCapabilities:
    - model-on-gpu           # custom capability that can be used for matching Model to this server
  podSpec:
    nodeSelector:            # only run mlserver-llm-local-gpu pods on nodes that have all those labels  
      pool: infer-srv
      cloud.google.com/gke-accelerator: nvidia-a100-80gb  # example requesting specific GPU on GKE, not required
      # cloud.google.com/gke-accelerator-count: 2   # also request node with label denoting a specific GPU count
    ...
    tolerations:             # allow mlserver-llm-local-gpu pods to be scheduled on nodes with the matching taint
    - effect: NoSchedule
      key: seldon-gpu-srv
      operator: Equal
      value: "true"
    ...
    containers:              # if needed, override settings from Serverconfig, for this specific Server
	  - name: mlserver
		  resources:
			  requests:
				  nvidia.com/gpu: 1  # in particular, have the mlserver container request a GPU
				  cpu: 40
				  memory: 360Gi
				  ephemeral-storage: 290Gi
				limits:
					nvidia.com/gpu: 2
					cpu: 40
					memory: 360Gi
					
```

You can configure more advanced Pod selection using `nodeAffinity`, as in this example:

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Server
metadata:
  name: mlserver-llm-local-gpu     # <server name>
  namespace: seldon-mesh           # <seldon runtime namespace>
spec:
	...
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
	            values: ["true"] # (optional) only schedule on nodes with PCI device 10
    
    ...
    tolerations:             #** allow mlserver-llm-local-gpu pods to be scheduled on nodes with the matching taint
    - effect: NoSchedule
      key: seldon-gpu-srv
      operator: Equal
      value: "true"
    ...
    containers:
	  - name: mlserver
		  env:
			  ...
			image: ...
		  resources:
			  requests:
				  nvidia.com/gpu: 1
				  cpu: 40
				  memory: 360Gi
				  ephemeral-storage: 290Gi
				limits:
					nvidia.com/gpu: 2
					cpu: 40
					memory: 360Gi
		... # many other configs
```
{% endtab %}
{% endtabs %}

**Configuring ServerConfig custom resource**

This configuration automatically affects all servers using that `ServerConfig`, unless you specify server-specific overrides, which takes precedence.

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: ServerConfig
metadata:
  name: mlserver-llm               # <ServerConfig name>**
  namespace: seldon-mesh           # <seldon runtime namespace>**
spec:
  podSpec:
    nodeSelector:           # only run mlserver-llm-local-gpu pods on nodes that have all those labels  
      pool: infer-srv
      cloud.google.com/gke-accelerator: nvidia-a100-80gb# example requesting specific GPU on GKE, not required
      # cloud.google.com/gke-accelerator-count: 2   # also request node with label denoting a specific GPU count
   ...
    tolerations:             #** allow mlserver-llm-local-gpu pods to be scheduled on nodes with the matching taint
    - effect: NoSchedule
      key: seldon-gpu-srv
      operator: Equal
      value: "true"
    ...
    containers:
	  - name: mlserver
		  env:
			  ...
			image: ...
		  resources:
			  requests:
				  nvidia.com/gpu: 1
				  cpu: 40
				  memory: 360Gi
				  ephemeral-storage: 290Gi
				limits:
					nvidia.com/gpu: 2
					cpu: 40
					memory: 360Gi
		... # many other configs
```

### Configuring models&#x20;

When you have a set of inference servers running exclusively on GPU nodes, you can assign a model to one of those servers in two ways:

* Custom model requirements (recommended)
* Explicit server pinning

Here's the distinction between the two methods of assigning models to servers.

| Method                        | Behavior                                                                                                                                        |
| ----------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------- |
| **Explicit pinning**          | If the specified server lacks sufficient memory or resources, the model load fails without trying another server.                               |
| **Custom model requirements** | If the assigned server cannot load the model due to insufficient resources, another similarly-capable server can be selected to load the model. |

{% tabs %}
{% tab title="Custom model requirements" %}
When you specify a requirement matching a server capability in the model custom resource it loads the model on any inference server with a capability matching the requirements.

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Model
metadata:
  name: llama3           # <model name>**
  namespace: seldon-mesh # <seldon runtime namespace>**
spec:
  ...
  requirements:
  - model-on-gpu         # requirement matching a Server capability
  ...

```

Ensure that the additional capability that matches the requirement label is added to the Server custom resource.&#x20;

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Server
metadata:
  name: mlserver-llm-local-gpu     # **<server name>**
  namespace: seldon-mesh           # **<seldon runtime namespace>**
spec:
	serverConfig: mlserver     # **<reference Serverconfig CR>**
  extraCapabilities:
    - model-on-gpu           # custom capability that can be used for matching Model to this server
  ...
```

Instead of adding a capability using `extraCapabilities` on a Server custom resource, you may also add to the list of capabilities in the associated ServerConfig custom resource. This applies to all servers referencing the configuration.

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: ServerConfig
metadata:
  name: mlserver-llm               # **<ServerConfig name>**
  namespace: seldon-mesh           # **<seldon runtime namespace>**
spec:
  podSpec:
    ...
    containers:
	  - name: agent                  #** note the setting is applied to the **agent** container
		  env:
			- name: SELDON_SERVER_CAPABILITIES
			  value: mlserver,alibi-detect,...,xgboost,model-on-gpu  #** add capability to the list
			...  
			image: ...
		... # many other configs
```
{% endtab %}

{% tab title="Explicit pinning" %}
With these specifications, the model is loaded on replicas of inference servers created by the referenced Server custom resource.

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Model
metadata:
  name: llama3           # **<model name>**
  namespace: seldon-mesh # **<seldon runtime namespace>**
spec:
  ...
  server: mlserver-llm-local-gpu # **<reference Server CR>**
  ...

```
{% endtab %}
{% endtabs %}
