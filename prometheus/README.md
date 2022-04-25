# Prometheus

## Install

We recommend install [kube-prometheus](https://github.com/prometheus-community/helm-charts/tree/main/charts/kube-prometheus-stack) that provides an all-in-one package with the [Prometheus operator](https://github.com/prometheus-operator/prometheus-operator).

```
kubectl apply --server-side -f manifests/setup
until kubectl get servicemonitors --all-namespaces ; do date; sleep 1; echo ""; done
kubectl apply -f manifests/
```

### RBAC

You will need to modify the default RBAC installed by kube-prometheus as described [here](https://github.com/prometheus-operator/prometheus-operator/blob/main/Documentation/user-guides/getting-started.md#enable-rbac-rules-for-prometheus-pods).

```
kubectl apply -f rbac/cr.yaml
```

## Monitors

We use a PodMonitor for scrapping agent metrics. The envoy and server monitors are there for completeness but not present needed.

```
kubectl apply -f monitors
```

Includes:

 * Agent pod monitor. Monitors the metrics port of server inference pods.
 * Server pod monitor. Monitors the server-metrics port of server inference pods.
 * Envoy service monitor. Monitors the Envoy proxies.

Pod monitors were chosen as ports for metrics are not exposed at service level as we do not have a top level service for server replicas but 1 headless service per replica. Future discussions could reference [this](https://github.com/prometheus-operator/prometheus-operator/issues/3119).


# Reference

 * [Prometheus CRDs](https://github.com/prometheus-operator/prometheus-operator/blob/main/Documentation/api.md)

# Access the dashboards

 * Access to dashboards is explained [here](https://github.com/prometheus-operator/kube-prometheus#access-the-dashboards)

# TODO

 * Provide Ansible installation


