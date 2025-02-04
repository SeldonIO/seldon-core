# Model Scheduling

Core 2 architecture is built around decoupling `Model` and `Server` CRs to allow for multi-model deployment, 
enabling multiple models to be loaded and served on one server replica or a single `Pod`. Multi-model serving
allows for more efficient use of resources, see [Multi Model Serving](mms.md) for more information.

This architecture requires that Core 2 handles scheduling of models onto server pods natively. In particular Core 2 implements different sorters and filters which are used to find the best Server that is able to host a given Model. This process we describe in the following section.

## Scheduling Process

### Overview

The scheduling process in Core 2 attempts to find a candidate server for a given model. This is based on multiple steps that sort and filter 
servers mainly based on the following criterias:

- Server has matching Capabilities with Model `spec.requirements`.
- Server has enough replicas to load the desired `spec.replicas` of the Model.
- Each replica of Server has enough available memory to load one replica of the model defined in `spec.memory`.
- Server that already hosts the Model is preferred to reduce flip-flops between different candidate servers.

Once a candidate Server is identified for a given Model, Core 2 will attempt to load this Model onto it. Otherwise if there is no matching Server, the Model will be marked as `ScheduleFailed`.

This process is extensible and in future versions we might add other filters that can be used for scheduling decisions.

{% hint style="info" %}
**Note**: A specific Model can only be assigned to at most one Server and therefore this Server will require to have enough replicas to host all replicas of the Model.
{% endhint %}

## Partial Scheduling

Core 2 (from `2.9`) is able to do partial scheduling of Models. Partial scheduling is defined as the loading of enough replicas of the model above `spec.minReplicas` and upto the number of available Server replicas. This allows the user a little bit more flexibility in serving traffic while optimising infrastructure provisioning. 

To enable partial scheduling, `spec.minReplicas` needs to be defined as it provides Core 2 the minimum replicas of the model that is required for serving. 

There is no explicit state for partial scheduling though; the Model will be marked as `ModelAvailable` and ready to serve traffic. The user can inspect the status of the Model CR, where we populate DESIRED REPLICAS and AVAILABLE REPLICAS to expose information on how many replicas are currently loaded in Core 2. Accordingly, the following logic can be performed:

- *Fully Scheduled*: READY is True and DESIRED REPLICAS is equal to AVAILABLE REPLICAS (STATUS is `ModelAvailable`)
- *Partially Scheduled*: READY is TRUE and DESIRED REPLICAS is greater than AVAILABLE REPLICAS (STATUS is `ModelAvailable`)
- *Not Scheduled*: Ready is False (Status is `ScheduleFailed`)

