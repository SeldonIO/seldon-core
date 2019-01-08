# Seldon Core Analytics

This is a Prometheus and Grafana installation with a basic Grafana dashboard showing the default Prometheus metrics exposed by Seldon's Service Orchestrator for each Seldon Deployment graph that is run.

## Installation

The Helm chart takes the following parameters:

  * ```grafana_prom_admin_password``` : The password for logging into Grafana with the admin account
  * ```persistence.enabled``` : Whether to try and run with persistence. If ```true``` then you need to provide a persistent volume with a claim name ```seldon-claim``` which Prometheus will mount to store metrics.


An example install is shown below:

```
helm install seldon-core-analytics --name seldon-core-analytics --set grafana_prom_admin_password=password --set persistence.enabled=false --repo https://storage.googleapis.com/seldon-charts --namespace seldon
```

To access the Grafana dashboard port-forward to the Grafana pod:

```
kubectl port-forward $(kubectl get pods -n seldon -l app=grafana-prom-server -o jsonpath='{.items[0].metadata.name}') -n seldon 3000:3000
```

You can then open http://localhost:3000 to log into Grafana using your set password.


