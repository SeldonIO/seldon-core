# Autoscaling

Autoscaling in Seldon applies to various concerns:

 * Autoscaling of inference servers
 * Autoscaling of models
 * Model memory overcommit

## Autoscaling of servers

All servers have HPA autoscalers attached and will scale within the bounds of defined `minReplicas` (>0) and `maxReplicas` based on CPU utilization: at present a default average utilization of 90%.

## Autoscaling of models

```{note}
Autoscaling of models is in the roadmap.
```

As each server can serve multiple models, models can scale across the available replicas of the server.

## Model memory Overcommit

Servers can hold more models than available memory if overcommit is swictched on (default yes). This allows under utilized models to be moved from inference server memory to allow for other models to take their place. If traffic patterns for inference of models vary then this can allow more models than available server memory to be run on the Seldon system.

