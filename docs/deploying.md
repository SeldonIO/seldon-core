# Deployment


 1. Deploy your machine learning model inference graph
 1. Validate successful deployment


## Deploy

You can manage your deployment resource via the standard Kuberntes tools:

### Kubectl

You can manage your deployments via the standard Kubernetes CLI kubectl, e.g.

```bash
kubectl apply -f my_ml_deployment.yaml
```

### Helm

You can use Helm to manage your deployment as illustrated in the [Helm examples notebook](../notebooks/helm_examples.ipynb).

We have a selection of [templated helm charts](../helm-charts/README.md#seldon-core-inference-graph-templates) you can use as a basis for your deployments.

### Ksonnet

You can use Ksonnet to manage your deployments as illustrated in the [Ksonnet examples notebook](../notebooks/ksonnet_examples.ipynb).

We have a selection of [Ksonnet prototypes](../seldon-core/seldon-core/README.md) you can use as a basis for your deployments.


## Validate

You can check the status of the running deployments using kubectl

For example:

```
kubectl get sdep -o jsonpath='{.items[].status}'
```





