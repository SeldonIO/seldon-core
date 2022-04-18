# Metrics

Metrics are exposed for scrapping by [Prometheus](https://prometheus.io/).

## Example Installation

We recommend to install [kube-prometheus](https://github.com/prometheus-operator/kube-prometheus#installing) that provides an all-in-one package with the [Prometheus operator](https://github.com/prometheus-operator/prometheus-operator).


### RBAC

You will need to modify the default RBAC installed by kube-prometheus as described [here](https://github.com/prometheus-operator/prometheus-operator/blob/main/Documentation/user-guides/getting-started.md#enable-rbac-rules-for-prometheus-pods).

From the `prometheus` folder in the project run:

```bash
kubectl apply -f rbac/cr.yaml
```

## Monitors

We use a PodMonitor for scrapping agent metrics. The envoy and server monitors are there for completeness but not presently needed.

```bash
kubectl apply -f monitors
```

Includes:

 * Agent pod monitor. Monitors the metrics port of server inference pods.
 * Server pod monitor. Monitors the server-metrics port of inference server pods.
 * Envoy service monitor. Monitors the Envoy gateway proxies.

Pod monitors were chosen as ports for metrics are not exposed at service level as we do not have a top level service for server replicas but 1 headless service per replica. Future discussions could reference [this](https://github.com/prometheus-operator/prometheus-operator/issues/3119).

## Example Grafana Dashboard

TBC

## Reference

 * [Prometheus CRDs](https://github.com/prometheus-operator/prometheus-operator/blob/main/Documentation/api.md)
