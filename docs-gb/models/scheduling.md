<!--- cSpell:enable --->

# Model Scheduling

Core 2 architecture is built around decoupling `Model` and `Server` CRs to allow for multi-model deployement, 
i.e. enabling multiple models to be loaded and served on one server replica (`Pod`). 
This architecture requires that Core 2 handles scheduling of models onto server pods natively, which we describe next.

## Scheduling Process

### Overview

The scheduling process in Core 2 attempts to find a candidate server for a given model. This is based on multiple steps that filter 
servers mainly based on the following criterias:

- Server has matching Capabilities with Model `spec.requirements`.
- Server has enough replicas to load the desired `spec.replicas` of the Model.
- Each replica of Server has enough available memory to load one replica of the model defined in `spec.memory`.

Once a candidate Server is identified for a given Model, Core 2 will attempt to load this Model onto it. Otherwise if there is no matching Server, the Model will be marked as `ScheduleFailed`.

## Partial Scheduling