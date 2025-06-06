# Model Explainers

![cat](cat.png)
![explanation](cat_explanation.png)

Seldon provides model explanations using its [Alibi](https://github.com/SeldonIO/alibi) library.

The v1 explainer server supports explainers saved with Python 3.7. However, for the Open Inference Protocol (or V2 protocol) using MLServer, this requirement is no longer necessary.

| Package | Version |
| ------ | ----- |
| `alibi` | `0.6.4` |


## Available Methods

Seldon Core supports a subset of the methods currently available in [Alibi](https://github.com/SeldonIO/alibi). Presently this the following:


| Method | Explainer Key |
|--------|---------------||
| [Anchor Tabular](https://docs.seldon.io/projects/alibi/en/latest/methods/Anchors.html) | `AnchorTabular` |
| [Anchor Text](https://docs.seldon.io/projects/alibi/en/latest/methods/Anchors.html) | `AnchorText` |
| [Anchor Images](https://docs.seldon.io/projects/alibi/en/latest/methods/Anchors.html) | `AnchorImages` |
| [kernel Shap](https://docs.seldon.io/projects/alibi/en/latest/methods/KernelSHAP.html) | `KernelShap` |
| [Integrated Gradients](https://docs.seldon.io/projects/alibi/en/latest/methods/IntegratedGradients.html) | `IntegratedGradients` |
| [Tree Shap](https://docs.seldon.io/projects/alibi/en/latest/methods/TreeSHAP.html) | `TreeShap` |

## Creating your explainer

For Alibi explainers that need to be trained you should

 1. Use python 3.7 as the Seldon Alibi Explain Server also runs in python 3.7.10 when it loads your explainer.
 1. Follow the [Alibi docs](https://docs.seldon.io/projects/alibi/en/latest/index.html) for your particular desired explainer. The Seldon Wrapper presently supports: Anchors (Tabular, Text and Image), KernelShap and Integrated Gradients.
 1. Save your explainer using [explainer.save](https://docs.seldon.io/projects/alibi/en/latest/overview/saving.html) method and store in the object store or PVC in your cluster. We support various cloud storage solutions through our [init container](../servers/overview.html).

The runtime environment in our [Alibi Explain Server](https://github.com/SeldonIO/seldon-core/tree/master/components/alibi-explain-server) is locked using [Poetry](https://python-poetry.org/). See our e2e example [here](../examples/iris_explainer_poetry.html) on how to use that definition to train your explainers.

### Open Inference Protocol for explainer using [MLServer](https://github.com/SeldonIO/MLServer)

The support for Open Inference Protocol is now handled with MLServer moving forward. This is experimental
and only works for black-box explainers.

For an e2e example, please check AnchorTabular notebook [here](../examples/iris_anchor_tabular_explainer_v2.html).

## Explain API

**Note**: Seldon has adopted the industry-standard Open Inference Protocol (OIP) and is no longer maintaining the Seldon and TensorFlow protocols. This transition allows for greater interoperability among various model serving runtimes, such as MLServer. To learn more about implementing OIP for model serving in Seldon Core 1, see [MLServer](https://docs.seldon.ai/mlserver).

We strongly encourage you to adopt the OIP, which provides seamless integration across diverse model serving runtimes, supports the development of versatile client and benchmarking tools, and ensures a high-performance, consistent, and unified inference experience.


For the Seldon Protocol an endpoint path will be exposed for:

```
http://<ingress-gateway>/seldon/<namespace>/<deployment name>/<predictor name>/api/v1.0/explain
```

So for example if you deployed:

```yaml
apiVersion: machinelearning.seldon.io/v1
kind: SeldonDeployment
metadata:
  name: income
  namespace: seldon
spec:
  name: income
  annotations:
    seldon.io/rest-timeout: "100000"
  predictors:
  - graph:
      children: []
      implementation: SKLEARN_SERVER
      modelUri: gs://seldon-models/sklearn/income/model-0.23.2
      name: classifier
    explainer:
      type: AnchorTabular
      modelUri: gs://seldon-models/sklearn/income/explainer-py36-0.5.2
    name: default
    replicas: 1
```

If you were port forwarding to Ambassador or istio on localhost:8003 then the API call would be:

```
http://localhost:8003/seldon/seldon/income-explainer/default/api/v1.0/explain
```

The explain method is also supported for tensorflow and Open Inference protocols. The full list of endpoint URIs is:

| Protocol | URI |
| ------ | ----- |
| v2 | `http://<host>/<ingress-path>/v2/models/<model-name>/infer` |
| seldon | `http://<host>/<ingress-path>/api/v1.0/explain` |
| tensorflow | `http://<host>/<ingress-path>/v1/models/<model-name>:explain` |


Note: for `tensorflow` protocol we support similar non-standard extension as for the [prediction API](../graph/protocols.md#rest-and-grpc-tensorflow-protocol), `http://<host>/<ingress-path>/v1/models/:explain`.
