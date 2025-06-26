---
description: Learn how to leverage Core 2's native autoscaling functionality for Models  
---

# Autoscaling Models based on Inference Lag

In order to set up autoscaling, users should first identify which metric they would want to scale their models on. Seldon Core provides an approach to autoscale models based on **Inference Lag**, or supports more custom scaling logic by leveraging HPA, (or [Horizontal Pod Autoscaler](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/)), whereby you can use custom metrics to automatically scale Kubernetes resources. This page will go through the first approach. **Inference Lag** refers to the difference in incoming vs. outgoing requests in a given period of time. If choosing this approach, it is recommended to configure autoscaling for Servers, so that Models scale on Inference Lag, and in turn set up [autoscaling for Servers](./core-autoscaling-servers.md) to scale based on model needs.

This implementation of autoscaling is enabled if Core 2 is installed with the `autoscaling.autoscalingModelEnabled` helm value set to `true` (default is `false`) and at least `MinReplicas` or `MaxReplicas` is set in the Model Custom Resource. Then according to lag (how much the model "falls behind" in terms of serving inference requests) the system will scale the number of `Replicas` within this range. As an example, the following model will be deployed at first with 1 replica and will autoscale according to lag.

```yaml
# samples/models/tfsimple_scaling.yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Model
metadata:
  name: tfsimple
spec:
  storageUri: "gs://seldon-models/triton/simple"
  requirements:
  - tensorflow
  memory: 100Ki
  minReplicas: 1
  replicas: 1
```

When the system autoscales, the initial model spec is not changed (e.g. the number of `replicas`) and therefore the user cannot reset the number of replicas back to the initial specified value without an explicit change. If only `replicas` is specified by the user, autoscaling of models is disabled and the system will have exactly the number of replicas of this model deployed regardless of inference load.

The scale-up and scale-down logic, and it's configurability is described below:

- **Scale Up**: To trigger scale up with the approach described above, we use **Inference Lag** as the metrics. **Inference Lag** is the difference between incoming and outgoing requests in a given time period. If the lag crosses a threshold, then we trigger a model scale up event. This threshold can be defined via `SELDON_MODEL_INFERENCE_LAG_THRESHOLD` inference server environment variable. The threshold used will apply to all the models hosted on the Server where the lag was configured.

- **Scale Down**: When using Model autoscaling that is managed by Seldon Core, model scale down events are triggered if a model has not been used for a number of seconds. This is defined in `SELDON_MODEL_INACTIVE_SECONDS_THRESHOLD` inference server environment variable.

- **Rate of metrics calculation**: Each agent checks the above stats periodically and if any model hits the corresponding threshold, then the agent sends an event to the scheduler to request model scaling. How often this process executes can be defined via `SELDON_SCALING_STATS_PERIOD_SECONDS` inference server environment variable.

Based on the logic above, the scheduler will trigger model autoscaling if:
* The model is stable (no state change in the last 5 minutes) and available.
* The desired number of replicas is within range. Note we always have a least 1 replica of any deployed model and we rely on over commit to reduce the resources used further.
* For scaling up, there is enough capacity for the new model replica.

{% hint style="danger" %}
If autoscaling models with the approach above, it is recommended to autoscale servers based on using Seldon's Server autoscaling (configured by setting `MinReplicas` and `MaxReplicas` for the Server CR - see below). Without Server autoscaling configured, the required number of servers will not necessarily spin up, even if the desired number of model replicas cannot be currently fulfilled by the current provisioned number of servers. Setting up Server Autoscaling is described in more detail below.
{% endhint %}
