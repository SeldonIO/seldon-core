# Seldon Core Analytics

This is a Prometheus and Grafana installation with a basic Grafana dashboard showing the default Prometheus metrics exposed by Seldon's Service Orchestrator for each Seldon Deployment graph that is run.

## Installation

The Helm chart parameters for Prometheus and Grafana can be found and edited in values.yaml, examples of which can be seen for Prometheus [here](https://github.com/helm/charts/blob/master/stable/prometheus/values.yaml) and for Grafana [here](https://github.com/helm/charts/blob/master/stable/grafana/values.yaml).

An example install is shown below:

```
kubectl create namespace seldon
```

```
helm install seldon-core-analytics . -n seldon
```

To access the Grafana dashboard port-forward to the Grafana pod:

```
kubectl port-forward $(kubectl get pods -l app=grafana -n seldon -o jsonpath='{.items[0].metadata.name}') -n seldon 3000:3000
```

You can then open http://localhost:3000 to log into Grafana using your set password from values.yaml.


