# Seldon Core Native Integration with Kubeflow

The Seldon Core team is part of the core Kubeflow team, which means that we are continuously ensuring Seldon Core is fully integrated with the Kubeflow Stack.

Seldon Core comes installed with Kubeflow. The [Seldon Core documentation site](https://docs.seldon.io/projects/seldon-core/en/latest/) provides full documentation for running Seldon Core inference.

If you have a saved model in a PersistentVolume (PV), Google Cloud Storage bucket or Amazon S3 Storage you can use one of the [prepackaged model servers provided by Seldon Core](https://docs.seldon.io/projects/seldon-core/en/latest/servers/overview.html).

Seldon Core also provides [language specific model wrappers](../wrappers/language_wrappers.html) to wrap your inference code for it to run in Seldon Core.

## Kubeflow specifics

You need to ensure the namespace where your models will be served has:

* An Istio gateway named kubeflow-gateway
* A label set as `serving.kubeflow.org/inferenceservice=enabled`

The following example applies the label `my-namespace` to the namespace for serving:

```console
kubectl label namespace my-namespace serving.kubeflow.org/inferenceservice=enabled
```

Create a gateway called `kubeflow-gateway` in namespace `my-namespace`:

```yaml
apiVersion: networking.istio.io/v1alpha3
kind: Gateway
metadata:
  name: kubeflow-gateway
  namespace: my-namespace
spec:
  selector:
    istio: ingressgateway
  servers:
  - hosts:
    - '*'
    port:
      name: http
      number: 80
      protocol: HTTP
```

Save the above resource and apply it with `kubectl`.
