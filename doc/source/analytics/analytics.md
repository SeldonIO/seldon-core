# Seldon Core Analytics

Seldon Core exposes metrics that can be scraped by Prometheus. The core metrics are exposed by the service orchestrator (```engine```) and API gateway (```server_ingress```).

The metrics are:

**Prediction Requests**

 * ```seldon_api_engine_server_requests_duration_seconds_(bucket,count,sum) ``` : Requests to the service orchestrator from an ingress, e.g. API gateway or Ambassador
 * ```seldon_api_engine_client_requests_duration_seconds_(bucket,count,sum) ``` : Requests from the service orchestrator to a component, e.g., a model
 * ```seldon_api_server_ingress_requests_duration_seconds_(bucket,count,sum) ``` : Requests to the API Gateway from an external client

**Feedback Requests**

 * ```seldon_api_model_feedback_reward_total``` : Reward sent via Feedback API
 * ```seldon_api_model_feedback_total``` : Total feedback requests

Each metric has the following key value pairs for further filtering which will be taken from the SeldonDeployment custom resource that is running:

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
helm install seldon-core-analytics --name seldon-core-analytics \
     --repo https://storage.googleapis.com/seldon-charts \
     --set grafana_prom_admin_password=password \
     --set persistence.enabled=false \
```

The available parameters are:

 * ```grafana_prom_admin_password``` : The admin user Grafana password to use.
 * ```persistence.enabled``` : Whether Prometheus persistence is enabled.

Once running you can expose the Grafana dashboard with:

```bash
kubectl port-forward $(kubectl get pods -l app=grafana-prom-server -o jsonpath='{.items[0].metadata.name}') 3000:3000
```

You can then view the dashboard at http://localhost:3000/dashboard/db/prediction-analytics?refresh=5s&orgId=1

![dashboard](./dashboard.png)

