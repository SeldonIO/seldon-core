# Autoscaling

Autoscaling in Seldon applies to various concerns:

 * Autoscaling of inference servers
 * Autoscaling of models
 * Model memory overcommit

## Autoscaling of models

```{note}
Autoscaling of models is in the roadmap.
```

As each server can serve multiple models, models can scale across the available replicas of the server.

Autoscaling will be provided internally as well as allowing external use via HPA or KEDA.

## Model memory Overcommit

Servers can hold more models than available memory if overcommit is swictched on (default yes). This allows under utilized models to be moved from inference server memory to allow for other models to take their place. If traffic patterns for inference of models vary then this can allow more models than available server memory to be run on the Seldon system.

