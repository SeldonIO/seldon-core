# Server (native) Autoscaling

## Overview
Core 2 runs with long lived server replicas that can host multiple models (MMS). These server replicas can be autoscaled natively to host these dynamically loaded models, allowing users to seamlessly optimise infrastructure cost associated with their deployment. This document outlines the autoscaling policies and mechanisms that are available for server replicas. 

{% hint style="info" %}
**Note**: Native autoscaling of servers is required in the case of MMS as the models are dynamically loaded and unloaded onto these server replicas. In this case Core 2 would autoscale server replicas according to changes to the model replicas that are required. This is in contrast to single model serving where the server replicas can be autoscaled using HPA.

In fact, in the case of single model serving, where the user is using HPA to autoscale the model replicas, the server replicas can be scaled by Core 2 without the use of HPA. This simplifies the autoscaling process as the user only needs to manage the model replicas (via HPA).
{% endhint %}

## Requirements

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
  serverConfig: mlserver
```

## Scale Up

## Scale Down

### Overview

When we want to scale down a model replicas, the associated servers might be left empty not used. In this case these extra server pods (especially if they use GPUs) are wasted causing infrastructure cost for the user.

Scaling down servers in sync with models is not straight forward in the case of multi model serving. Scaling down one model does not necessarily mean that we also need to scale down the corresponding server replica as this server replica might be still serving load for other models. 

Therefore we need to define some heuristics that can be used to scale down servers if we think that they are not properly used.

{% hint style="info" %}
**Note**: Scaling down the number of replicas for a model server does not necessarily mean that the system is going to remove a specific replica that w e want. 

As currently we have the model server deployed as StatefulSets, scaling down the number of replicas will mean that we are removing a pod with the largest idx.

The system will rebalance afterwards, where the models from this draining server replica will be rescheduled but then there is a period of time where this is happening.

This is less critical as by definition we are scaling down and therefore the load in general is not high.
{% endhint %}

### Policies

1. **Empty Server Replica**:
    
    In the simplest case we can remove a server replica if it doesn't host any more models. This guarantees that there is no load on a particular server replica before removing it.
    
    This policy is simple to implement however it suffers from the following issues:
    
    - In MMS different model replicas will be hosted on potentially different server replicas and as we scale these models up and down the system can end up in a situation where there the models are not consolidated to the an optimised number of servers. Take for example the case of Models A, B and C. We have 2 server replicas 1 and 2 that can host these models. Assuming that at the start we have only A and B have 1 replica and C has 2 replicas therefore the initial assignment is:
        - Replica 1: A1, C1
        - Replica 2: B1, C2
        
        Now if the user unloads Model C the assignment is:
        
        - Replica 1: A1
        - Replica 2: B1
        
        There is a strong argument that this is not optimised and in MMS the assignment should really be:
        
        - Replica 1: A1, B1
        - Replica 1 removed
    
    As the system evolves this imbalance can get bigger and will could cause the serving infrastructure to be less optimised. 
    
    In fact this is not directly related to autoscaling per se  as the example above is actually not related to autoscaling. However autoscaling will aggravate the issue causing more imbalance.
    
2. **Lightly Loaded Server Replicas**
    
    The above imbalance can be mitigated by making by the following observation: If the max number of replicas of any given model (assigned to a server from a logical point of view) is less than the number of replicas for this server, then we can pack the models hosted onto a smaller set of replicas.
    
    In other words consider the following example, for models A and B having 2 replicas each and we have 3 server replicas, the following assignment is not potentially optimised.
    
    - Replica  1: A1, B1
    - Replica 2: A2
    - Replica 3: B2
    
    In this case we could trigger removal of Replica 3 for the server which could pack the models more appropriately
    
    - Replica 1: A1, B1
    - Replica 2: A2, B2
    - Replica 3: removed
    
    While this heuristic is going to pack models onto a fewer set of replicas, which allows us to scale models down, there is still the risk that the packing could increase latencies, trigger a later scale up. We need to make sure that we are not flip-flopping between these states. 
    
    This packing process can be triggered on the following:
    
    - Scale down event for a model