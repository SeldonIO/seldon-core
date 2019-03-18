# Deployment

Once you have created your inference graph as a JSON or YAML Seldon Deployment resource you can deploy to Kubernetes with `kubectl`. For example, if your deployment is packaged in `my_ml_deployment.yaml`:

```bash
kubectl apply -f my_ml_deployment.yaml
```

## Helm Charts

You can use Helm to manage your deployment as illustrated in the [Helm examples notebook](../examples/helm_examples.html).

We have a selection of [templated helm charts](../graph/helm_charts.html) you can use as a basis for your deployments.

## Ksonnet Apps

You can use Ksonnet to manage your deployments as illustrated in the [Ksonnet examples notebook](../notebooks/ksonnet_examples.ipynb).

We have a selection of [Ksonnet prototypes](../graph/ksonnet_templates.html) you can use as a basis for your deployments.


## Validate your Deployment

You can check the status of the running deployments using kubectl

For example:

```bash
kubectl get sdep -o jsonpath='{.items[].status}'
```





