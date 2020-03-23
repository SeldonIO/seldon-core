# Metrics

Seldon Core exposes metrics that can be scraped by Prometheus. The core metrics are exposed by the service orchestrator (```executor```).

The metrics are:

**Prediction Requests**

 * ```seldon_api_executor_server_requests_seconds_(bucket,count,sum) ``` : Requests to the service orchestrator from an ingress, e.g. API gateway or Ambassador
 * ```seldon_api_executor_client_requests_seconds_(bucket,count,sum) ``` : Requests from the service orchestrator to a component, e.g., a model

Each metric has the following key value pairs for further filtering which will be taken from the SeldonDeployment custom resource that is running:

  * service
  * deployment_name
  * predictor_name
  * predictor_version
    * This will be derived from the predictor metadata labels
  * model_name
  * model_image
  * model_image


## Helm Analytics Chart

Seldon Core provides an example Helm analytics chart that displays the above Prometheus metrics in Grafana. You can install it with:

```bash
helm install seldon-core-analytics seldon-core-analytics \
   --repo https://storage.googleapis.com/seldon-charts \
   --namespace seldon-system
```

Once running you can expose the Grafana dashboard with:

```bash
kubectl port-forward svc/seldon-core-analytics-grafana 3000:80 -n seldon-system
```

You can then view the dashboard at http://localhost:3000/dashboard/db/prediction-analytics

![dashboard](./dashboard.png)

It is also possible expose Prometheus itself with:
```bash
kubectl port-forward svc/seldon-core-analytics-prometheus-seldon 3001:80 -n seldon-system
```

and then access it at http://localhost:3001/



## Example

There is [an example notebook you can use to test the metrics](../examples/metrics.html).
