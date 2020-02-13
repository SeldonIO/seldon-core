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
   --set grafana_prom_admin_password=password \
   --set persistence.enabled=false \
   --namespace seldon-system
```

The available parameters are:

 * ```grafana_prom_admin_password``` : The admin user Grafana password to use.
 * ```persistence.enabled``` : Whether Prometheus persistence is enabled.

Once running you can expose the Grafana dashboard with:

```bash
kubectl port-forward $(kubectl get pods -n seldon-system -l app=grafana-prom-server -o jsonpath='{.items[0].metadata.name}') 3000:3000 -n seldon-system
```

You can then view the dashboard at http://localhost:3000/dashboard/db/prediction-analytics?refresh=5s&orgId=1

![dashboard](./dashboard.png)

## Example

There is [an example notebook you can use to test the metrics](../examples/metrics.html).

