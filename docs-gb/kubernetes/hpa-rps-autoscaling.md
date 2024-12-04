
# HPA Autoscaling in single-model serving

Learn how to jointly autoscale model and server replicas based on a metric of inference
requests per second (RPS) using HPA, when there is a one-to-one correspondence between models
and servers (single-model serving). This will require:

* Having a Seldon Core 2 install that publishes metrics to prometheus (default). In the
  following, we will assume that prometheus is already installed and configured in the
  `seldon-monitoring` namespace.
* Installing and configuring [Prometheus Adapter](https://github.com/kubernetes-sigs/prometheus-adapter),
  which allows prometheus queries on relevant metrics to be published as k8s custom metrics
* Configuring HPA manifests to scale Models and the corresponding Server replicas based on the
  custom metrics

{% hint style="warning" %}
The Core 2 HPA-based autoscaling has the following constraints/limitations:

    - HPA scaling only targets single-model serving, where there is a 1:1 correspondence between
    models and servers. Autoscaling for multi-model serving (MMS) is supported for specific models
    and workloads via the Core 2 native features described [here](autoscaling.md).
    Significant improvements to MMS autoscaling are planned for future releases.
    
    - **Only custom metrics** from Prometheus are supported. Native Kubernetes
    resource metrics such as CPU or memory are not. This limitation exists because of HPA's
    design: In order to prevent multiple HPA CRs from issuing conflicting scaling instructions,
    each HPA CR must exclusively control a set of pods which is disjoint from the pods
    controlled by other HPA CRs. In Seldon Core 2, CPU/memory metrics can be used to scale the
    number of Server replicas via HPA. However, this also means that the CPU/memory metrics
    from the same set of pods can no longer be used to scale the number of model replicas. We
    are working on improvements in Core 2 to allow both servers and models to be scaled based on
    a single HPA manifest, targeting the Model CR.
    
    - Each Kubernetes cluster supports only one active custom metrics provider. If your cluster
    already uses a custom metrics provider different from `prometheus-adapter`, it
    will need to be removed before being able to scale Core 2 models and servers via HPA. The
    Kubernetes is actively exploring solutions for allowing multiple custom metrics providers to
    coexist.
{% endhint %}

## Installing and configuring the Prometheus Adapter

The role of the Prometheus Adapter is to expose queries on metrics in Prometheus as k8s custom
or external metrics. Those can then be accessed by HPA in order to take scaling decisions.

To install through helm:

```sh
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update
helm install --set prometheus.url='http://seldon-monitoring-prometheus' hpa-metrics prometheus-community/prometheus-adapter -n seldon-monitoring
```

These commands install `prometheus-adapter` as a helm release named `hpa-metrics` in
the same namespace where Prometheus is installed, and point to its service URL (without the port).

The URL is not fully qualified as it references a Prometheus instance running in the same namespace.
If you are using a separately-managed Prometheus instance, please update the URL accordingly.

If you are running Prometheus on a different port than the default 9090, you can also pass `--set
prometheus.port=[custom_port]` You may inspect all the options available as helm values by
running `helm show values prometheus-community/prometheus-adapter`

{% hint style="warning" %}
Please check that the `metricsRelistInterval` helm value (default to 1m) works well in your
setup, and update it otherwise. This value needs to be larger than or equal to your Prometheus
scrape interval. The corresponding prometheus adapter command-line argument is
`--metrics-relist-interval`. If the relist interval is set incorrectly, it will lead to some of
the custom metrics being intermittently reported as missing.
{% endhint %}

We now need to configure the adapter to look for the correct prometheus metrics and compute
per-model RPS values. On install, the adapter has created a `ConfigMap` in the same namespace as
itself, named `[helm_release_name]-prometheus-adapter`. In our case, it will be
`hpa-metrics-prometheus-adapter`.

Overwrite the ConfigMap as shown in the following manifest, after applying any required customizations.

{% hint style="warning" %}
Change the `name` if you've chosen a different value for the `prometheus-adapter` helm release name.
Change the `namespace` to match the namespace where `prometheus-adapter` is installed.
{% endhint %}

{% code title="prometheus-adapter.config.yaml" lineNumbers="true" %}
````yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: hpa-metrics-prometheus-adapter
  namespace: seldon-monitoring
data:
  config.yaml: |-
    "rules":
    -
      "seriesQuery": |
         {__name__="seldon_model_infer_total",namespace!=""}
      "resources":
        "overrides":
          "model": {group: "mlops.seldon.io", resource: "model"}
          "server": {group: "mlops.seldon.io", resource: "server"}
          "pod": {resource: "pod"}
          "namespace": {resource: "namespace"}
      "name":
        "matches": "seldon_model_infer_total"
        "as": "infer_rps"
      "metricsQuery": |
        sum by (<<.GroupBy>>) (
          rate (
            <<.Series>>{<<.LabelMatchers>>}[1m]
          )
        )
````
{% endcode %}

In this example, a single rule is defined to fetch the `seldon_model_infer_total` metric
from Prometheus, compute its rate over a 1 minute window, and expose this to k8s as the `infer_rps`
metric, with aggregations available at model, server, inference server pod and namespace level.

When HPA requests the `infer_rps` metric via the custom metrics API for a specific model,
prometheus-adapter issues a Prometheus query in line with what it is defined in its config.

For the configuration in our example, the query for a model named `irisa0` in namespace
`seldon-mesh` would be:

```
sum by (model) (
  rate (
    seldon_model_infer_total{model="irisa0", namespace="seldon-mesh"}[1m]
  )
)
```

Before configuring prometheus-adapter via the ConfigMap, it is important to sanity-check the query by executing it against
your Prometheus instance. To do so, pick an existing model CR in your Seldon Core 2 install, and
send some inference requests towards it. Then, wait for a period equal to the Prometheus scrape
interval (Prometheus default 1 minute) so that the metric values are updated. Finally, you can
modify the model name and namespace in the query above to match the model you've picked and
execute the query.

If the query result is non-empty, you may proceed with the next steps, or customize the query
according to your needs and re-test. If the query result is empty, please adjust it until it
returns the expected metric values. Update the `metricsQuery` in the prometheus-adapter
ConfigMap to match.

A list of all the Prometheus metrics exposed by Seldon Core 2 in relation to Models, Servers and
Pipelines is available [here](../metrics/operational.md), and those may be used when customizing
the configuration.

### Customizing prometheus-adapter rule definitions

The rule definition can be broken down in four parts:

* _Discovery_ (the `seriesQuery` and `seriesFilters` keys) controls what Prometheus
    metrics are considered for exposure via the k8s custom metrics API.

  As an alternative to the example above, all the Seldon Prometheus metrics of the form `seldon_model.*_total`
  could be considered, followed by excluding metrics pre-aggregated across all models (`.*_aggregate_.*`) as well as
  the cummulative infer time per model (`.*_seconds_total`):

    ```yaml
    "seriesQuery": |
            {__name__=~"^seldon_model.*_total",namespace!=""}
        "seriesFilters":
            - "isNot": "^seldon_.*_seconds_total"
            - "isNot": "^seldon_.*_aggregate_.*"
    ...
    ```

  For RPS, we are only interested in the model inference count (`seldon_model_infer_total`)

* _Association_ (the `resources` key) controls the Kubernetes resources that a particular
    metric can be attached to or aggregated over.

  The resources key defines an association between certain labels from the Prometheus metric and
  k8s resources. For example, on line 17, `"model": {group: "mlops.seldon.io", resource: "model"}`
  lets `prometheus-adapter` know that, for the selected Prometheus metrics, the value of the
  "model" label represents the name of a k8s `model.mlops.seldon.io` CR.

  One k8s custom metric is generated for each k8s resource associated with a prometheus metric.
  In this way, it becomes possible to request the k8s custom metric values for
  `models.mlops.seldon.io/iris` or for `servers.mlops.seldon.io/mlserver`.

  The labels that *do not* refer to a `namespace` resource generate "namespaced" custom
  metrics (the label values refer to resources which are part of a namespace) -- this
  distinction becomes important when needing to fetch the metrics via kubectl, and in
  understanding how certain Prometheus query template placeholders are replaced.


* _Naming_ (the `name` key) configures the naming of the k8s custom metric.

  In the example ConfigMap, this is configured to take the Prometheus metric named
  `seldon_model_infer_total` and expose custom metric endpoints named `infer_rps`, which when
  called return the result of a query over the Prometheus metric.

  Instead of a literal match, one could also use regex group capture expressions,
  which can then be referenced in the custom metric name:

  ```yaml
  "name":
    "matches": "^seldon_model_(.*)_total"
    "as": "${1}_rps"
  ```

* _Querying_ (the `metricsQuery` key) defines how a request for a specific k8s custom metric gets
    converted into a Prometheus query.

  The query can make use of the following placeholders:

    - .Series is replaced by the discovered prometheus metric name (e.g. `seldon_model_infer_total`)
    - .LabelMatchers, when requesting a namespaced metric for resource `X` with name `x` in
    namespace `n`, is replaced by `X=~"x",namespace="n"`. For example, `model=~"iris0",
    namespace="seldon-mesh"`. When requesting the namespace resource itself, only the
    `namespace="n"` is kept.
    - .GroupBy is replaced by the resource type of the requested metric (e.g. `model`,
    `server`, `pod` or `namespace`).

  You may want to modify the query in the example to match the one that you typically use in
  your monitoring setup for RPS metrics. The example calls [`rate()`](https://prometheus.io/docs/prometheus/latest/querying/functions/#rate)
  with a 1 minute window.


For a complete reference for how `prometheus-adapter` can be configured via the `ConfigMap`, please
consult the docs [here](https://github.com/kubernetes-sigs/prometheus-adapter/blob/master/docs/config.md).


Once you have applied any necessary customizations, replace the default prometheus-adapter config
with the new one, and restart the deployment (this restart is required so that prometheus-adapter
picks up the new config):

```sh
# Replace default prometheus adapter config
kubectl replace -f prometheus-adapter.config.yaml
# Restart prometheus-adapter pods
kubectl rollout restart deployment hpa-metrics-prometheus-adapter -n seldon-monitoring
```

### Testing the install using the custom metrics API

In order to test that the prometheus adapter config works and everything is set up correctly,
you can issue raw kubectl requests against the custom metrics API

{% hint style="info" %}
**Note**: If no inference requests were issued towards any model in the Seldon install, the metrics
configured above will not be available in prometheus, and thus will also not appear when
checking via the commands below. Therefore, please first run some inference requests towards a
sample model to ensure that the metrics are available — this is only required for the testing of
the install.
{% endhint %}

Listing the available metrics:

```sh
kubectl get --raw /apis/custom.metrics.k8s.io/v1beta1/ | jq .
```

For namespaced metrics, the general template for fetching is:

```sh
kubectl get --raw "/apis/custom.metrics.k8s.io/v1beta1/namespaces/[NAMESPACE]/[API_RESOURCE_NAME]/[CR_NAME]/[METRIC_NAME]"
```

For example:

* Fetching model RPS metric for a specific `(namespace, model)` pair `(seldon-mesh, irisa0)`:

    ```sh
    kubectl get --raw /apis/custom.metrics.k8s.io/v1beta1/namespaces/seldon-mesh/models.mlops.seldon.io/irisa0/infer_rps
    ```

* Fetching model RPS metric aggregated at the `(namespace, server)` level `(seldon-mesh, mlserver)`:

    ```sh
    kubectl get --raw /apis/custom.metrics.k8s.io/v1beta1/namespaces/seldon-mesh/servers.mlops.seldon.io/mlserver/infer_rps
    ```

* Fetching model RPS metric aggregated at the `(namespace, pod)` level `(seldon-mesh, mlserver-0)`:

    ```sh
    kubectl get --raw /apis/custom.metrics.k8s.io/v1beta1/namespaces/seldon-mesh/pods/mlserver-0/infer_rps
    ```

* Fetching the same metric aggregated at `namespace` level `(seldon-mesh)`:

    ```sh
    kubectl get --raw /apis/custom.metrics.k8s.io/v1beta1/namespaces/*/metrics/infer_rps
    ```

## Configuring HPA manifests

For every (Model, Server) pair you want to autoscale, you need to apply 2 HPA manifests based on
the same metric: one scaling the Model, the other the Server. The example below only works if
the mapping between Models and Servers is 1-to-1 (i.e no multi-model serving).

Consider a model named `irisa0` with the following manifest. Please note we don’t set
`minReplicas/maxReplicas`. This disables the seldon lag-based autoscaling so that it
doesn’t interact with HPA (separate `minReplicas/maxReplicas` configs will be set on the HPA
side)

You must also explicitly define a value for `spec.replicas`. This is the key modified by HPA
to increase the number of replicas, and if not present in the manifest it will result in HPA not
working until the Model CR is modified to have `spec.replicas` defined.

{% code title="irisa0.yaml" lineNumbers="false" %}
```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Model
metadata:
  name: irisa0
  namespace: seldon-mesh
spec:
  memory: 3M
  replicas: 1
  requirements:
  - sklearn
  storageUri: gs://seldon-models/testing/iris1
```
{% endcode %}

Let’s scale this model when it is deployed on a server named `mlserver`, with a target RPS **per
replica** of 3 RPS (higher RPS would trigger scale-up, lower would trigger scale-down):

{% code title="irisa0-hpa.yaml" lineNumbers="true" %}
```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: irisa0-model-hpa
  namespace: seldon-mesh
spec:
  scaleTargetRef:
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    name: irisa0
  minReplicas: 1
  maxReplicas: 3
  metrics:
  - type: Object
    object:
      metric:
        name: infer_rps
      describedObject:
        apiVersion: mlops.seldon.io/v1alpha1
        kind: Model
        name: irisa0
      target:
        type: AverageValue
        averageValue: 3
---
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: mlserver-server-hpa
  namespace: seldon-mesh
spec:
  scaleTargetRef:
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Server
    name: mlserver
  minReplicas: 1
  maxReplicas: 3
  metrics:
  - type: Object
    object:
      metric:
        name: infer_rps
      describedObject:
        apiVersion: mlops.seldon.io/v1alpha1
        kind: Model
        name: irisa0
      target:
        type: AverageValue
        averageValue: 3
```
{% endcode %}

In the preceding HPA manifests, the scaling metric is exactly the same, and uses the exact same
parameters. This is to ensure that both the Models and the Servers are scaled up/down at
approximately the same time. Small variations in the scale-up time are expected because each HPA
samples the metrics independently, at regular intervals.

{% hint style="info" %}
**Note**: If a Model gets scaled up slightly before its corresponding Server, the model is currently
marked with the condition ModelReady "Status: False" with a "ScheduleFailed" message until new
Server replicas become available. However, the existing replicas of that model remain available
and will continue to serve inference load.
{% endhint %}

In order to ensure similar scaling behaviour between Models and Servers, the number of
`minReplicas` and `maxReplicas`, as well as any other configured scaling policies should be kept
in sync across the HPA for the model and the server.

### Details on custom metrics of type Object

The HPA manifests use metrics of type "Object" that fetch the data used in scaling
decisions by querying k8s metrics associated with a particular k8s object.  The endpoints that
HPA uses for fetching those metrics are the same ones that were tested in the previous section
using `kubectl get --raw ...`. Because you have configured the Prometheus Adapter to expose those
k8s metrics based on queries to Prometheus, a mapping exists between the information contained
in the HPA Object metric definition and the actual query that is executed against Prometheus.
This section aims to give more details on how this mapping works.

In our example, the `metric.name:infer_rps` gets mapped to the `seldon_model_infer_total` metric
on the prometheus side, based on the configuration in the `name` section of the Prometheus
Adapter ConfigMap. The prometheus metric name is then used to fill in the `<<.Series>>` template
in the query (`metricsQuery` in the same ConfigMap).

Then, the information provided in the `describedObject` is used within the Prometheus query to
select the right aggregations of the metric. For the RPS metric used to scale the Model
(and the Server because of the 1-1 mapping), it makes sense to compute the aggregate RPS across
all the replicas of a given model, so the `describedObject` references a specific Model CR.

However, in the general case, the `describedObject` does not need to be a Model. Any k8s object
listed in the `resources` section of the Prometheus Adapter ConfigMap may be used. The Prometheus
label associated with the object kind fills in the `<<.GroupBy>>` template, while the name gets
used as part of the `<<.LabelMatchers>>`. For example:

* If the described object is `{ kind: Namespace, name: seldon-mesh }`, then the Prometheus
query template configured in our example would be transformed into:

```
sum by (namespace) (
  rate (
    seldon_model_infer_total{namespace="seldon-mesh"}[1m]
  )
)
```

* If the described object is not a namespace (for example, `{ kind: Pod, name: mlserver-0 }`)
then the query will be passed the label describing the object, alongside an additional label
identifying the namespace where the HPA manifest resides in.:

```
sum by (pod) (
  rate (
    seldon_model_infer_total{pod="mlserver-0", namespace="seldon-mesh"}[1m]
  )
)
```

For the `target` of the Object metric you **must** use a `type` of `AverageValue`. The value
given in `averageValue` represents the per replica RPS scaling threshold of the `scaleTargetRef`
object (either a Model or a Server in our case), with the target number of replicas being
computed by HPA according to the following formula:

$$\texttt{targetReplicas} = \frac{\texttt{infer\_rps}}{\texttt{thresholdPerReplicaRPS}}$$

{% hint style="info" %}
**Note**: Attempting other target types does not work under the current Seldon Core 2 setup, because they
use the number of active Pods associated with the Model CR (i.e. the associated Server pods) in
the `targetReplicas` computation. However, this also means that this set of pods becomes "owned"
by the Model HPA. Once a pod is owned by a given HPA it is not available for other HPAs to use,
so we would no longer be able to scale the Server CRs using HPA.
{% endhint %}

### HPA sampling of custom metrics

Each HPA CR has it's own timer on which it samples the specified custom metrics. This timer
starts when the CR is created, with sampling of the metric being done at regular intervals (by
default, 15 seconds).

As a side effect of this, creating the Model HPA and the Server HPA (for a given model) at
different times will mean that the scaling decisions on the two are taken at different times.
Even when creating the two CRs together as part of the same manifest, there will usually be a
small delay between the point where the Model and Server `spec.replicas` values are changed.

Despite this delay, the two will converge to the same number when the decisions are taken based
on the same metric (as in the previous examples).

When showing the HPA CR information via `kubectl get`, a column of the output will display the
current metric value per replica and the target average value in the format `[per replica metric value]/[target]`.
This information is updated in accordance to the sampling rate of each HPA resource. It is
therefore expected to sometimes see different metric values for the Model and it's corresponding
Server.

{% hint style="info" %}
Some versions of k8s will display `[per pod metric value]` instead of `[per replica metric value]`,
with the number of pods being computed based on a label selector present in the target resource
CR (the `status.selector` value for the Model or Server in the Core 2 case).

HPA is designed so that multiple HPA CRs cannot target the same underlying pod with this selector
(with HPA stopping when such a condition is detected). This means that in Core 2, the Model and
Server selector cannot be the same. A design choice was made to assign the Model a unique
selector that does not match any pods.

As a result, for the k8s versions displaying `[per pod metric value]`, the information shown for
the Model HPA CR will be an overflow caused by division by zero. This is only a display artefact,
with the Model HPA continuing to work normally. The actual value of the metric can be seen by
inspecting the corresponding Server HPA CR, or by fetching the metric directly via `kubectl get
--raw`
{% endhint %}

### Advanced settings

*   Filtering metrics by additional labels on the prometheus metric:

    The prometheus metric from which the model RPS is computed has the following labels managed
    by Seldon Core 2:

    ```c-like
    seldon_model_infer_total{
        code="200", 
        container="agent", 
        endpoint="metrics", 
        instance="10.244.0.39:9006", 
        job="seldon-mesh/agent", 
        method_type="rest", 
        model="irisa0", 
        model_internal="irisa0_1", 
        namespace="seldon-mesh", 
        pod="mlserver-0", 
        server="mlserver", 
        server_replica="0"
    }
    ```

    If you want the scaling metric to be computed based on a subset of the Prometheus time
    series with particular label values (labels either managed by Seldon Core 2 or added
    automatically within your infrastructure), you can add this as a selector the HPA metric
    config. This is shown in the following example, which scales only based on the RPS of REST
    requests as opposed to REST + gRPC:

    ```yaml
      metrics:
      - type: Object
        object:
          describedObject:
            apiVersion: mlops.seldon.io/v1alpha1
            kind: Model
            name: irisa0
          metric:
            name: infer_rps
            selector:
              matchLabels:
                method_type: rest
          target:
    	    type: AverageValue
            averageValue: "3"
    ```

*   Customize scale-up / scale-down rate & properties by using scaling policies as described in
    the [HPA scaling policies docs](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/#configurable-scaling-behavior)

*   For more resources, please consult the [HPA docs](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/)
    and the [HPA walkthrough](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale-walkthrough/)


## Cluster operation guidelines when using HPA-based scaling

When deploying HPA-based scaling for Seldon Core 2 models and servers as part of a production deployment,
it is important to understand the exact interactions between HPA-triggered actions and Seldon Core 2
scheduling, as well as potential pitfalls in choosing particular HPA configurations.

Using the default scaling policy, HPA is relatively aggressive on scale-up (responding quickly
to increases in load), with a maximum replicas increase of either 4 every 15 seconds or 100% of
existing replicas within the same period (**whichever is highest**). In contrast, scaling-down
is more gradual, with HPA only scaling down to the maximum number of recommended replicas in the
most recent 5 minute rolling window, in order to avoid flapping. Those parameters can be
customized via [scaling policies](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/#configurable-scaling-behavior).

When using custom metrics such as RPS, the actual number of replicas added during scale-up or
reduced during scale-down will entirely depend, alongside the maximums imposed by the policy, on
the configured target (`averageValue` RPS per replica) and on how quickly the inferencing load
varies in your cluster. All three need to be considered jointly in order to deliver both an
efficient use of resources and meeting SLAs.

### Customizing per-replica RPS targets and replica limits

Naturally, the first thing to consider is an estimated peak inference load (including some
margins) for each of the models in the cluster. If the minimum number of model
replicas needed to serve that load without breaching latency SLAs is known, it should be set as
`spec.maxReplicas`, with the HPA `target.averageValue` set to `peak_infer_RPS`/`maxReplicas`.

If `maxReplicas` is not already known, an open-loop load test with a slowly ramping up request
rate should be done on the target model (one replica, no scaling). This would allow you to
determine the RPS (inference request throughput) when latency SLAs are breached or (depending on
the desired operation point) when latency starts increasing. You would then set the HPA
`target.averageValue` taking some margin below this saturation RPS, and compute
`spec.maxReplicas` as `peak_infer_RPS`/`target.averageValue`. The margin taken below the
saturation point is very important, because scaling-up cannot be instant (it requires spinning
up new pods, downloading model artifacts, etc.). In the period until the new replicas become
available, any load increases will still need to be absorbed by the existing replicas.

If there are multiple models which typically experience peak load in a correlated manner, you
need to ensure that sufficient cluster resources are available for k8s to concurrently schedule
the maximum number of server pods, with each pod holding one model replica. This can be ensured
by using either [Cluster Autoscaler](https://kubernetes.io/docs/concepts/cluster-administration/cluster-autoscaling/)
or, when running workloads in the cloud, any provider-specific cluster autoscaling services.

{% hint style="warning" %}
It is important for the cluster to have sufficient resources for creating the total number of
desired server replicas set by the HPA CRs across all the models at a given time.

Not having sufficient cluster resources to serve the number of replicas configured by HPA at a
given moment, in particular under aggressive scale-up HPA policies, may result in breaches of
SLAs. This is discussed in more detail in the following section.
{% endhint %}

A similar approach should be taken for setting `minReplicas`, in relation to estimated RPS in
the low-load regime. However, it's useful to balance lower resource usage to immediate
availability of replicas for inference rate increases from that lowest load point. If low-load
regimes only occur for small periods of time, and especially combined with a high rate of increase
in RPS when moving out of the low-load regime, it might be worth to set the `minReplicas` floor
higher in order to ensure SLAs are met at all times.


### Customizing HPA policy settings for ensuring correct scaling behaviour

Each `spec.replica` value change for a Model or Server triggers a rescheduling event for the
Seldon Core 2 scheduler, which considers any updates that are needed in mapping Model replicas to
Server replicas  such as rescheduling failed Model replicas, loading new ones, unloading in the case of the number of replicas going down, etc.

Two characteristics in the current implementation are important in terms of
autoscaling and configuring the HPA scale-up policy:

- The scheduler does not create new Server replicas when the existing replicas are not
    sufficient for loading a Model's replicas (one Model replica per Server replica). Whenever
    a Model requests more replicas than available on any of the available Servers, its `ModelReady`
    condition transitions to `Status: False` with a `ScheduleFailed` message. However, any
    replicas of that Model that are already loaded at that point remain available for servicing
    inference load.

- There is no _partial_ scheduling of replicas. For example, consider a model with 2 replicas,
    currently loaded on a server with 3 replicas (two of those server replicas will have the
    model loaded). If you update the model replicas to 4, the scheduler will transition the
    model to `ScheduleFailed`, seeing that it cannot satisfy the requested number of replicas.
    The existing 2 model replicas will continue to serve traffic, *but a third replica will
    not be loaded* onto the remaining server replica.

    In other words, the scheduler either schedules all the requested replicas, or,
    if unable to do so, leaves the state of the cluster unchanged.

    Introducing partial scheduling would make the overall results of assigning models to servers
    significantly less predictable and ephemeral. This is because models may end up moved
    back-and forth between servers depending on the speed with which various server replicas
    become available. Network partitions or other transient errors may also trigger large changes
    to the model-to-server assignments, making it challenging to sustain consistent data plane
    load during those periods.

Taken together, the two Core 2 scheduling characteristics, combined with a very aggressive HPA
scale-up policy and a continuously increasing RPS may lead to the following pathological case:

- Based on RPS, HPA decides to increase both the Model and Server replicas from 2 (an example
start stable state) to 8. While the 6 new Server pods get scheduled and get the Model loaded
onto them, the scheduler will transition the Model into the `ScheduleFailed` state, because it
cannot fulfill the requested replicas requirement. During this period, the initial 2 Model
replicas continue to serve load, but are using their RPS margins and getting closer to the
saturation point.
- At the same time, load continues to increase, so HPA further increases the number of
required Model and Server replicas from 8 to 12, before all of the 6 new Server pods had a chance
to become available. The new replica target for the scheduler also becomes 12, and this would
not be satisfied until all the 12 Server replicas are available. The 2 Model replicas that are
available may by now be saturated and the infer latency spikes up, breaching set SLAs.
- The process may continue until load stabilizes.
- If at any point the number of requested replicas (<=`maxReplicas`) exceeds the resource
capacity of the cluster, the requested server replica count will never be reached and thus the
Model will remain permanently in the `ScheduleFailed` state.

While most likely encountered during continuous ramp-up RPS load tests with autoscaling enabled,
the pathological case example is a good showcase for the elements that need to be taken
into account when setting the HPA policies.

- The speed with which new Server replicas can become available versus how many new replicas may
  HPA request in a given time:
    - The HPA scale-up policy should not be configured to request more replicas than can
      become available in the specified time. The following example reflects a confidence that 5
      Server pods will become available within 90 seconds, with some safety margin. The default
      scale-up config, that also adds a percentage based policy (double the existing replicas
      within the set `periodSeconds`) is not recommended because of this.
    - Perhaps more importantly, there is no reason to scale faster than the time it takes for
      replicas to become available - this is the true maximum rate with which scaling up can
      happen anyway.

{% code title="hpa-custom-policy.yaml" lineNumbers="true" %}
```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: irisa0-model-hpa
  namespace: seldon-mesh
spec:
  scaleTargetRef:
    ...
  minReplicas: 1
  maxReplicas: 3
  behavior:
    scaleUp:
      stabilizationWindowSeconds: 60
      policies:
      - type: Pods
        value: 5
        periodSeconds: 90
  metrics:
    ...
```
{% endcode %}

- The duration of transient load spikes which you might want to absorb within the existing
  per-replica RPS margins.
    - The previous example, at line 13, configures a scale-up stabilization window of one minute.
      It means that for all of the HPA recommended replicas in the last 60 second window (4
      samples of the custom metric considering the default sampling rate), only the *smallest*
      will be applied.
    - Such stabilization windows should be set depending on typical load patterns in your
        cluster: not being too aggressive in reacting to increased load will allow you to
        achieve cost savings, but has the disadvantage of a delayed reaction if the load spike turns
        out to be sustained.

- The duration of any typical/expected sustained ramp-up period, and the RPS increase rate
    during this period.
    - It is useful to consider whether the replica scale-up rate configured via the policy (line
        15 in the example) is able to keep-up with this RPS increase rate.
    - Such a scenario may appear, for example, if you are planning for a smooth traffic ramp-up
        in a blue-green deployment as you are draining the "blue" deployment and transitioning
        to the "green" one
