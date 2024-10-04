
# HPA Autoscaling in single-model serving

Learn about how to jointly autoscale model and server replicas based on a metric of inference
requests per second (RPS) using HPA, when there is a one-to-one correspondence between models
and servers (single-model serving). This will require:

* Having a Seldon Core 2 install that publishes metrics to prometheus (default). In the
  following, we will assume that prometheus is already installed and configured in the
  `seldon-monitoring` namespace.
* Installing and configuring [Prometheus Adapter](https://github.com/kubernetes-sigs/prometheus-adapter),
  which allows prometheus queries on relevant metrics to be published as k8s custom metrics
* Configuring HPA manifests to scale Models and the corresponding Server replicas based on the
  custom metrics

### Installing and configuring the Prometheus Adapter

The role of the Prometheus Adapter is to expose queries on metrics in prometheus as k8s custom
or external metrics. Those can then be accessed by HPA in order to take scaling decisions.

To install through helm:

```sh
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update
helm install --set prometheus.url='http://seldon-monitoring-prometheus' hpa-metrics prometheus-community/prometheus-adapter -n seldon-monitoring
```

In the commands above, we install `prometheus-adapter` as a helm release named `hpa-metrics` in
the same namespace as our prometheus install, and point to its service URL (without the port).

If running prometheus on a different port than the default 9090, you can also pass `--set
prometheus.port=[custom_port]` You may inspect all the options available as helm values by
running `helm show values prometheus-community/prometheus-adapter`

We now need to configure the adapter to look for the correct prometheus metrics and compute
per-model RPS values. On install, the adapter has created a `ConfigMap` in the same namespace as
itself, named `[helm_release_name]-prometheus-adapter`. In our case, it will be
`hpa-metrics-prometheus-adapter`.

We want to overwrite this ConfigMap with the content below (please change the name if your helm
release has a different one). The manifest contains embedded documentation, highlighting how we
match the `seldon_model_infer_total` metric in Prometheus, compute a rate via a `metricsQuery`
and expose this to k8s as the `infer_rps` metric, on a per (model, namespace) basis.

Other aggregations on per (server, namespace) and (pod, namespace) are also exposed and may be
used in HPA, but we will focus on the (model, namespace) aggregation in the examples below.

You may want to modify some of the settings to match the prometheus query that you typically use
for RPS metrics. For example, the `metricsQuery` below computes the RPS by calling [`rate()`]
(https://prometheus.io/docs/prometheus/latest/querying/functions/#rate) with a 1 minute window.

````yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: hpa-metrics-prometheus-adapter
  namespace: seldon-monitoring
data:
  config.yaml: |-
    "rules":
    # Rule matching Seldon inference requests-per-second metrics and exposing aggregations for
    # specific k8s models, servers, pods and namespaces
    #
    # Uses the prometheus-side `seldon_model_(.*)_total` inference request count metrics to
    # compute and expose k8s custom metrics on inference RPS `${1}_rps`. A prometheus metric named
    # `seldon_model_infer_total` will be exposed as multiple `[group-by-k8s-resource]/infer_rps`
    # k8s metrics, for consumption by HPA.
    #
    # One k8s metric is generated for each k8s resource associated with a prometheus metric, as
    # defined in the "Association" section below. Because this association is defined based on
    # labels present in the prometheus metric, the number of generated k8s metrics will vary
    # depending on what labels are available in each discovered prometheus metric.
    #
    # The resources associated through this rule (when available as labels for each of the
    # discovered prometheus metrics) are:
    # - models
    # - servers
    # - pods (inference server pods)
    # - namespaces
    #
    # For example, you will get aggregated metrics for `models.mlops.seldon.io/iris0/infer_rps`,
    # `servers.mlops.seldon.io/mlserver/infer_rps`, `pods/mlserver-0/infer_rps`,
    # `namespaces/seldon-mesh/infer_rps`
    #
    # Metrics associated with any resource except the namespace one (models, servers and pods)
    # need to be requested in the context of a particular namespace.
    #
    # To fetch those k8s metrics manually once the prometheus-adapter is running, you can run:
    #
    # For "namespaced" resources, i.e. models, servers and pods (replace values in brackets):
    # ```
    # kubectl get --raw
    # "/apis/custom.metrics.k8s.io/v1beta1/namespaces/[NAMESPACE]/[RESOURCE_NAME]/[CR_NAME]/infer_rps"
    # ```
    #
    # For example:
    # ```
    # kubectl get --raw
    # "/apis/custom.metrics.k8s.io/v1beta1/namespaces/seldon-mesh/models.mlops.seldon.io/iris0/infer_rps"
    # ```
    #
    # For the namespace resource, you can get the namespace-level aggregation of the metric with:
    # ```
    # kubectl get --raw "/apis/custom.metrics.k8s.io/v1beta1/namespaces/*/metrics/infer_rps"
    # ```
    -
      # Metric discovery: selects subset of metrics exposed in Prometheus, based on name and
      # filters
      "seriesQuery": |
         {__name__=~"^seldon_model.*_total",namespace!=""}
      "seriesFilters":
        - "isNot": "^seldon_.*_seconds_total"
        - "isNot": "^seldon_.*_aggregate_.*"
      # Association: maps label values in the Prometheus metric to K8s resources (native or CRs)
      # Below, we associate the "model" prometheus metric label to the corresponding Seldon Model
      # CR, the "server" label to the Seldon Server CR, etc.
      "resources":
        "overrides":
          "model": {group: "mlops.seldon.io", resource: "model"}
          "server": {group: "mlops.seldon.io", resource: "server"}
          "pod": {resource: "pod"}
          "namespace": {resource: "namespace"}
      # Rename prometheus metrics to get k8s metric names that reflect the processing done via
      # the query applied to those metrics (actual query below under the "metricsQuery" key)
      "name":
        "matches": "^seldon_model_(.*)_total"
        "as": "${1}_rps"
      # The actual query to be executed against Prometheus to retrieve the metric value
      # Here:
      #   - .Series is replaced by the discovered prometheus metric name (e.g.
      #     `seldon_model_infer_total`)
      #   - .LabelMatchers, when requesting a metric for a namespaced resource X with name x in
      #     namespace n, is replaced by `X=~"x",namespace="n"`. For example, `model=~"iris0",
      #     namespace="seldon-mesh"`. When requesting the namespace resource itself, only the
      #     `namespace="n"` is kept.
      #   - .GroupBy is replaced by the resource type of the requested metric (e.g. `model`,
      #     `server`, `pod` or `namespace`).
      "metricsQuery": |
        sum by (<<.GroupBy>>) (
          rate (
            <<.Series>>{<<.LabelMatchers>>}[1m]
          )
        )
````

Apply the config, and restart the prometheus adapter deployment (this restart is required so
that prometheus-adapter picks up the new config):

```sh
# Apply prometheus adapter config
kubectl apply -f prometheus-adapter.config.yaml
# Restart prom-adapter pods
kubectl rollout restart deployment hpa-metrics-prometheus-adapter -n seldon-monitoring
```

In order to test that the prometheus adapter config works and everything is set up correctly,
you can issue raw kubectl requests against the custom metrics API, as described below.

{% hint style="info" %}
If no inference requests were issued towards any model in the Seldon install, the metrics
configured above will not be available in prometheus, and thus will also not appear when
checking via the commands below. Therefore, please first run some inference requests towards a
sample model to ensure that the metrics are available — this is only required for the testing of
the install.
{% endhint %}

**Testing the prometheus-adapter install using the custom metrics API**

List available metrics

```sh
kubectl get --raw /apis/custom.metrics.k8s.io/v1beta1/ | jq .
```

Fetching model RPS metric for specific (namespace, model) pair:

```sh
kubectl get --raw /apis/custom.metrics.k8s.io/v1beta1/namespaces/seldon-mesh/models.mlops.seldon.io/irisa0/infer_rps
```

Fetching model RPS metric aggregated at the (namespace, server) level:

```sh
kubectl get --raw /apis/custom.metrics.k8s.io/v1beta1/namespaces/seldon-mesh/servers.mlops.seldon.io/mlserver/infer_rps
```

Fetching model RPS metric aggregated at the (namespace, pod) level:

```sh
kubectl get --raw /apis/custom.metrics.k8s.io/v1beta1/namespaces/seldon-mesh/pods/mlserver-0/infer_rps
```

Fetching the same metric aggregated at namespace level:

```sh
kubectl get --raw /apis/custom.metrics.k8s.io/v1beta1/namespaces/*/metrics/infer_rps
```

### Configuring HPA manifests

For every (Model, Server) pair you want to autoscale, you need to apply 2 HPA manifests based on
the same metric: one scaling the Model, the other the Server. The example below only works if
the mapping between Models and Servers is 1-to-1 (i.e no multi-model serving).

Consider a model named `irisa0` with the following manifest. Please note we don’t set
`minReplicas/maxReplicas` this is in order to disable the seldon-specific autoscaling so that it
doesn’t interact with HPA.

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

Let’s scale this model when it is deployed on a server named `mlserver`, with a target RPS **per
replica** of 3 RPS (higher RPS would trigger scale-up, lower would trigger scale-down):

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

In the preceding HPA manifests, the scaling metric is exactly the same, and uses the exact same
parameters: this is to ensure that both the Models and the Servers are scaled up/down at
approximately the same time. Small variations in the scale-up time are expected because each HPA
samples the metrics independently, at regular intervals.

{% hint style="info" %}
If a Model gets scaled up slightly before its corresponding Server, the model is currently
marked with the condition ModelReady "Status: False" with a "ScheduleFailed" message until new
Server replicas become available. However, the existing replicas of that model remain available
and will continue to serve inference load.
{% endhint %}

In order to ensure similar scaling behaviour between Models and Servers, the number of
`minReplicas` and `maxReplicas`, as well as any other configured scaling policies should be kept
in sync across the HPA for the model and the server.

#### Details on custom metrics of type Object

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
select the right aggregations of the metric. For the RPS metric we have used to scale the Model
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
Attempting other target types does not work under the current Seldon Core 2 setup, because they
use the number of active Pods associated with the Model CR (i.e. the associated Server pods) in
the `targetReplicas` computation. However, this also means that this set of pods becomes "owned"
by the Model HPA. Once a pod is owned by a given HPA it is not available for other HPAs to use,
so we would no longer be able to scale the Server CRs using HPA.
{% endhint %}

#### Advanced settings

*   Filtering metrics by additional labels on the prometheus metric

    The prometheus metric from which the model RPS is computed has the following labels:

    ```yaml
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

    If you want the scaling metric to be computed based on inferences with a particular value
    for any of those labels, you can add this in the HPA metric config, as in the example
    (targeting `method_type="rest"`):

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
*   Customise scale-up / scale-down rate & properties by using scaling policies as described in
    the [HPA scaling policies docs](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/#configurable-scaling-behavior)

*   For more resources, please consult the [HPA docs](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/)
    and the [HPA walkthrough](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale-walkthrough/)
