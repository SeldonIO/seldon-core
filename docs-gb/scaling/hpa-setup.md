---
description: Learn how to implement request-per-second (RPS) based autoscaling in Seldon Core 2 using Kubernetes HPA and Prometheus metrics.
---

Given Seldon Core 2 is predominantly for serving ML in Kubernetes, it is possible to leverage `HorizontalPodAutoscaler` or [HPA](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/) to define scaling logic automatically scale up and down Kubernetes resources. This requires exposing metrics such that they can be used by HPA. In this tutorial, we will explain how to expose a metric (requests per second) using Prometheus and [Prometheus Adapter](https://github.com/kubernetes-sigs/prometheus-adapter), such that it can be used to autoscale Models or Servers using HPA. 

The following workflow will require: 

* Having a Seldon Core 2 install that publishes metrics to prometheus (default). In the following, we will assume that prometheus is already installed and configured in the `seldon-monitoring` namespace.
* Installing and configuring [Prometheus Adapter](https://github.com/kubernetes-sigs/prometheus-adapter), which allows prometheus queries on relevant metrics to be published as k8s custom metrics
* Configuring HPA manifests to scale Models 

{% hint style="warning" %}
Each Kubernetes cluster supports only one active custom metrics provider. If your cluster already uses a custom metrics provider different from `prometheus-adapter`, it will need to be removed before being able to scale Core 2 models and servers via HPA. The Kubernetes community is actively exploring solutions for allowing multiple custom metrics providers to coexist.
{% endhint %}

## Installing and configuring the Prometheus Adapter

The role of the Prometheus Adapter is to expose queries on metrics in Prometheus as k8s custom or external metrics. Those can then be accessed by HPA in order to take scaling decisions.

To install through helm:

```sh
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update
helm install --set prometheus.url='http://seldon-monitoring-prometheus' hpa-metrics prometheus-community/prometheus-adapter -n seldon-monitoring
```

These commands install `prometheus-adapter` as a helm release named `hpa-metrics` in the same namespace where Prometheus is installed, and point to its service URL (without the port).

The URL is not fully qualified as it references a Prometheus instance running in the same namespace. If you are using a separately-managed Prometheus instance, please update the URL accordingly.

If you are running Prometheus on a different port than the default 9090, you can also pass `--set prometheus.port=[custom_port]` You may inspect all the options available as helm values by running `helm show values prometheus-community/prometheus-adapter`

{% hint style="warning" %}
Please check that the `metricsRelistInterval` helm value (default to 1m) works well in your setup, and update it otherwise. This value needs to be larger than or equal to your Prometheus scrape interval. The corresponding prometheus adapter command-line argument is `--metrics-relist-interval`. If the relist interval is set incorrectly, it will lead to some of the custom metrics being intermittently reported as missing.
{% endhint %}

We now need to configure the adapter to look for the correct prometheus metrics and compute per-model RPS values. On install, the adapter has created a `ConfigMap` in the same namespace as itself, named `[helm_release_name]-prometheus-adapter`. In our case, it will be `hpa-metrics-prometheus-adapter`.

Overwrite the ConfigMap as shown in the following manifest, after applying any required customizations.

{% hint style="warning" %}
Change the `name` if you've chosen a different value for the `prometheus-adapter` helm release name. Change the `namespace` to match the namespace where `prometheus-adapter` is installed.
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
            <<.Series>>{<<.LabelMatchers>>}[2m]
          )
        )
````
{% endcode %}

In this example, a single rule is defined to fetch the `seldon_model_infer_total` metric from Prometheus, compute its per second change rate based on data within a 2 minute sliding window, and expose this to Kubernetes as the `infer_rps` metric, with aggregations available at model, server, inference server pod and namespace level.

When HPA requests the `infer_rps` metric via the custom metrics API for a specific model, prometheus-adapter issues a Prometheus query in line with what it is defined in its config.

For the configuration in our example, the query for a model named `irisa0` in namespace `seldon-mesh` would be:

```
sum by (model) (
  rate (
    seldon_model_infer_total{model="irisa0", namespace="seldon-mesh"}[2m]
  )
)
```

You may want to modify the query in the example to match the one that you typically use in your monitoring setup for RPS metrics. The example calls [`rate()`](https://prometheus.io/docs/prometheus/latest/querying/functions/#rate) with a 2 minute sliding window. Values scraped at the beginning and end of the 2 minute window before query time are used to compute the RPS.

It is important to sanity-check the query by executing it against your Prometheus instance. To do so, pick an existing model CR in your Seldon Core 2 install, and send some inference requests towards it. Then, wait for a period equal to at least twice the Prometheus scrape interval (Prometheus default 1 minute), so that two values from the series are captured and a rate can be computed. Finally, you can modify the model name and namespace in the query above to match the model you've picked and execute the query.

If the query result is empty, please adjust it until it consistently returns the expected metric values. Pay special attention to the window size (2 minutes in the example): if it is smaller than twice the Prometheus scrape interval, the query may return no results. A compromise needs to be reached to set the window size large enough to reject noise but also small enough to make the result responsive to quick changes in load.

Update the `metricsQuery` in the prometheus-adapter ConfigMap to match any query changes you have made during tests.

A list of all the Prometheus metrics exposed by Seldon Core 2 in relation to Models, Servers and Pipelines is available [here](../metrics/operational.md), and those may be used when customizing the configuration.

### Customizing prometheus-adapter rule definitions

The rule definition can be broken down in four parts:

1. **Discovery** (the `seriesQuery` and `seriesFilters` keys) controls what Prometheus metrics are considered for exposure via the k8s custom metrics API.

  As an alternative to the example above, all the Seldon Prometheus metrics of the form `seldon_model.*_total` could be considered, followed by excluding metrics pre-aggregated across all models (`.*_aggregate_.*`) as well as the cummulative infer time per model (`.*_seconds_total`):

    ```yaml
    "seriesQuery": |
            {__name__=~"^seldon_model.*_total",namespace!=""}
        "seriesFilters":
            - "isNot": "^seldon_.*_seconds_total"
            - "isNot": "^seldon_.*_aggregate_.*"
    ...
    ```

  For RPS, we are only interested in the model inference count (`seldon_model_infer_total`)

2. **Association** (the `resources` key) controls the Kubernetes resources that a particular metric can be attached to or aggregated over.

  The resources key defines an association between certain labels from the Prometheus metric and k8s resources. For example, on line 17, `"model": {group: "mlops.seldon.io", resource: "model"}` lets `prometheus-adapter` know that, for the selected Prometheus metrics, the value of the "model" label represents the name of a k8s `model.mlops.seldon.io` CR.

  One k8s custom metric is generated for each k8s resource associated with a prometheus metric. In this way, it becomes possible to request the k8s custom metric values for `models.mlops.seldon.io/iris` or for `servers.mlops.seldon.io/mlserver`.

  The labels that *do not* refer to a `namespace` resource generate "namespaced" custom metrics (the label values refer to resources which are part of a namespace) -- this distinction becomes important when needing to fetch the metrics via kubectl, and in understanding how certain Prometheus query template placeholders are replaced.

3. **Naming** (the `name` key) configures the naming of the k8s custom metric.

  In the example ConfigMap, this is configured to take the Prometheus metric named `seldon_model_infer_total` and expose custom metric endpoints named `infer_rps`, which when called return the result of a query over the Prometheus metric. Instead of a literal match, one could also use regex group capture expressions, which can then be referenced in the custom metric name:

  ```yaml
  "name":
    "matches": "^seldon_model_(.*)_total"
    "as": "${1}_rps"
  ```

4. **Querying** (the `metricsQuery` key) defines how a request for a specific k8s custom metric gets converted into a Prometheus query.

  The query can make use of the following placeholders:

    - .Series is replaced by the discovered prometheus metric name (e.g. `seldon_model_infer_total`)
    - .LabelMatchers, when requesting a namespaced metric for resource `X` with name `x` in namespace `n`, is replaced by `X=~"x",namespace="n"`. For example, `model=~"iris0", namespace="seldon-mesh"`. When requesting the namespace resource itself, only the `namespace="n"` is kept.
    - .GroupBy is replaced by the resource type of the requested metric (e.g. `model`, `server`, `pod` or `namespace`).

For a complete reference for how `prometheus-adapter` can be configured via the `ConfigMap`, please consult the docs [here](https://github.com/kubernetes-sigs/prometheus-adapter/blob/master/docs/config.md).

Once you have applied any necessary customizations, replace the default prometheus-adapter config with the new one, and restart the deployment (this restart is required so that prometheus-adapter picks up the new config):

```sh
# Replace default prometheus adapter config
kubectl replace -f prometheus-adapter.config.yaml
# Restart prometheus-adapter pods
kubectl rollout restart deployment hpa-metrics-prometheus-adapter -n seldon-monitoring
```

### Testing the install using the custom metrics API

In order to test that the prometheus adapter config works and everything is set up correctly, you can issue raw kubectl requests against the custom metrics API

{% hint style="info" %}
**Note**: If no inference requests were issued towards any model in the Seldon install, the metrics configured above will not be available in prometheus, and thus will also not appear when checking via the commands below. Therefore, please first run some inference requests towards a sample model to ensure that the metrics are available — this is only required for the testing of the install.
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
