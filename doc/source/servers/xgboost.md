# XGBoost Server

If you have a trained XGBoost model saved you can deploy it simply using
Seldon's prepackaged XGBoost server.

## Prerequisites

Seldon expects that your model has been saved as `model.bst`, using XGBoost's
`bst.save_model()` method.
Note that this is the [recommended approach to serialise
models](https://xgboost.readthedocs.io/en/latest/tutorials/saving_model.html).

To maximise compatibility between the serialised model and the serving runtime,
it's recommended to use the same toolkit versions at both training and
inference time. 
The expected dependency versions in the latest XGBoost pre-packaged server are
as follows:

| Package | Version |
| ------ | ----- |
| `xgboost` | `1.4.2` |

## Usage

To use the pre-packaged XGBoost server, it's enough to declare `XGBOOST_SERVER`
as the `implementation` for your model.
For example, for a saved Iris model, you could consider the following config:

```yaml
apiVersion: machinelearning.seldon.io/v1alpha2
kind: SeldonDeployment
metadata:
  name: xgboost
spec:
  name: iris
  predictors:
  - graph:
      children: []
      implementation: XGBOOST_SERVER
      modelUri: gs://seldon-models/xgboost/iris
      name: classifier
    name: default
    replicas: 1
```

You can try out a [worked notebook](../examples/server_examples.html) with a
similar example.

## V2 protocol

The XGBoost server can also be used to expose an API compatible with the [V2
protocol](../graph/protocols.md#v2-protocol).
Note that, under the hood, it will use the [Seldon
MLServer](https://github.com/SeldonIO/MLServer) runtime.

In order to enable support for the V2 protocol, it's enough to
specify the `protocol` of the `SeldonDeployment` to use `v2`.
For example,

```yaml
apiVersion: machinelearning.seldon.io/v1alpha2
kind: SeldonDeployment
metadata:
  name: xgboost
spec:
  name: iris
  protocol: v2 # Activate the V2 protocol
  predictors:
  - graph:
      children: []
      implementation: XGBOOST_SERVER
      modelUri: gs://seldon-models/xgboost/iris
      name: classifier
    name: default
    replicas: 1
```

You can try a similar example in [this worked
notebook](../examples/server_examples.html).
