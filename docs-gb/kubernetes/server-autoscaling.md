# Server _Native_ Autoscaling

## Overview
Core 2 runs with long lived server replicas that can host multiple models (MMS). These server replicas can be autoscaled natively to host these dynamically loaded models, allowing users to seamlessly optimise infrastructure cost associated with their deployment. This document outlines the autoscaling policies and mechanisms that are available for server replicas. These policies are designed to ensure that the server replicas are increased (scaled up) or decreased (scaled down) in response to changes in the number of model replicas that are required. In other words if a given model is scaled up, the system will scale up the server replicas to host these models. Similarly, if a given model is scaled down, the system will scale down the server replicas that are hosting these models.

{% hint style="info" %}
**Note**: Native autoscaling of servers is required in the case of MMS as the models are dynamically loaded and unloaded onto these server replicas. In this case Core 2 would autoscale server replicas according to changes to the model replicas that are required. This is in contrast to single model serving where the server replicas can be autoscaled using HPA.

In fact, in the case of single model serving, where the user is using HPA to autoscale the model replicas, the server replicas can be scaled by Core 2 without the use of HPA. This simplifies the autoscaling process as the user only needs to manage the model replicas (via HPA).
{% endhint %}

## Requirements

To enable autoscaling of server replicas, the following requirements need to be met:
1. Setting `minReplicas` and `maxReplicas` in the `Server` CR. This will define the minimum and maximum number of server replicas that can be created.
2. Setting `autoscaling.autoscalingServerEnabled` to `true` (default) during installation of Core 2. This will enable the autoscaling of server replicas.

An example of a `Server` CR with autoscaling enabled is shown below:

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Server
metadata:
  name: mlserver
  namespace: seldon
spec:
  replicas: 2
  minReplicas: 1
  maxReplicas: 4
maxReplicas: 4
  serverConfig: mlserver
```

{% hint style="info" %}
**Note**: Not setting `minReplicas` and/or `maxReplicas` will also effectively disable autoscaling of server replicas. In this case, the user will need to manually scale the server replicas by setting the `replicas` field in the `Server` CR. In this case this allows external autoscaling mechanisms to be used e.g. HPA. In future versions of Core 2, we might relax this requirement and allow native autoscaling of server replicas without setting both `minReplicas` and `maxReplicas`.
{% endhint %}

## Server Scale Up

### Overview

When we want to scale up a model replicas, the associated servers might not have enough capacity (replicas) to host these models. In this case we need to scale up the server replicas to host these models.

### Policies
There is currently only one policy for scaling up server replicas:
1. **Model Replica Count**:
    
    This policy scales up the server replicas to match the number of model replicas that are required. In other words, if a model is scaled up, the system will scale up the server replicas to host these models. This policy is simple to implement and ensures that the server replicas are scaled up in response to changes in the number of model replicas that are required.

    During the scale up process, the system will create new server replicas to host the new model replicas. The new server replicas will be created with the same configuration as the existing server replicas. This includes the server configuration, resources, etc. The new server replicas will be added to the existing server replicas and will be used to host the new model replicas. 

    There is a period of time where the new server replicas are being created and the new model replicas are being loaded onto these server replicas. During this period, the system will ensure that the existing server replicas are still serving load. This ensures that there is no downtime during the scale up process. This is achieved orthogonally by using partial scheduling of the new model replicas onto the new server replicas. This ensures that the new server replicas are gradually loaded with the new model replicas and that the existing server replicas are still serving load. Check the [Partial Scheduling](../models/scheduling.md) document for more details.

## Server Scale Down

### Overview

When we want to scale down a model replicas, the associated servers might be left empty not used. In this case these extra server pods (especially if they use GPUs) are wasted causing infrastructure cost for the user.

Scaling down servers in sync with models is not straight forward in the case of multi model serving. Scaling down one model does not necessarily mean that we also need to scale down the corresponding server replica as this server replica might be still serving load for other models. 

Therefore we need to define some heuristics that can be used to scale down servers if we think that they are not properly used.

{% hint style="info" %}
**Note**: Scaling down the number of replicas for a model server does not necessarily mean that the system is going to remove a specific replica that we want. 

As currently we have the model server deployed as StatefulSets, scaling down the number of replicas will mean that we are removing a pod with the largest index.

The system will rebalance afterwards, where the models from this draining server replica will be rescheduled but then there is a period of time where this is happening. This draining process should allows for no downtime as the models are being rescheduled onto other server replicas before the draining server replica is removed.
{% endhint %}

### Policies

1. **Empty Server Replica**:
    
    In the simplest case we can remove a server replica if it does not host any models. This guarantees that there is no load on a particular server replica before removing it.
    
    This policy works best in the case of single model serving where the server replicas are only hosting a single model. In this case, if the model is scaled down, the server replica will be empty and can be removed.

    However in the case of MMS it can lead to less optimal packing of models onto server replicas. This is because the system will not pack models onto a fewer set of replicas. This can lead to more server replicas being used than necessary. This can be mitigated by the lightly loaded server replicas policy.    
    
2. **Lightly Loaded Server Replicas** (Experimental):

    {% hint style="warning" %}
    **Warning**: This policy is experimental and is not enabled by default. It can be enabled by setting `autoscaling.serverPackingEnabled` to `true` and `autoscaling.serverPackingPercentage` to a value between 0 and 100. This policy is still under development and might in some cases increase latencies. Use with caution.
    {% endhint %}
    
    In MMS different model replicas will be hosted on potentially different server replicas and as we scale these models up and down the system can end up in a situation where the models are not consolidated to the an optimised number of servers. Take for example the case of 3 Models: $A$, $B$ and $C$. We have 1 server $S$ with 2 replicas: $S_1$ and $S_2$ that can host these models. Assuming that at the start we have only $A$ and $B$ have 1 replica and $C$ has 2 replicas therefore the initial assignment is:
        
    Initial assignment:

    - $S_1$: $A_1$, $C_1$
    - $S_2$: $B_1$, $C_2$
        
    Now if the user unloads Model $C$ the assignment is:
        
    - $S_1$: $A_1$
    - $S_2$: $B_1$
        
    There is a strong argument that this is not optimised and in MMS the assignment should really be:
        
    - $S_1$: $A_1$, $B_1$
    - $S_2$: removed
    
    As the system evolves this imbalance can get bigger and will could cause the serving infrastructure to be less optimised. 
    
    In fact this is not directly related to autoscaling per se as the example above is actually not related to autoscaling. However autoscaling will aggravate the issue causing more imbalance.

    The above imbalance can be mitigated by making by the following observation: If the max number of replicas of any given model (assigned to a server from a logical point of view) is less than the number of replicas for this server, then we can pack the models hosted onto a smaller set of replicas.
    
    In other words consider the following example, for models $A$ and $B$ having 2 replicas each and we have 3 server $S$ replicas, the following assignment is not potentially optimised.
    

    - $S_1$: $A_1$, $B_1$
    - $S_2$: $A_2$
    - $S_3$: $B_2$
    
    In this case we could trigger removal of $S_3$ for the server which could pack the models more appropriately
    
    - $S_1$: $A_1$, $B_1$
    - $S_2$: $A_2$, $B_2$
    - $S_3$: removed
    
    While this heuristic is going to pack models onto a fewer set of replicas, which allows us to scale models down, there is still the risk that the packing could increase latencies, trigger a later scale up. Core 2 tries to make sure that we are not flip-flopping between these states. This can be done by setting `autoscaling.serverPackingPercentage`.
