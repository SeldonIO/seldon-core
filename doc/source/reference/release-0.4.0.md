# Seldon Core Release 0.4.0

A summary of the main contributions to the [Seldon Core release 0.4.0](https://github.com/SeldonIO/seldon-core/releases/tag/v0.4.0).

## Prepackaged Model Servers

Seldon provides several prepacked servers you can use to deploy trained models:

 * [SKLearn Server](../servers/sklearn.html)
 * [XGBoost Server](../servers/xgboost.html)
 * [Tensorflow Serving](../servers/tensorflow.html)
 * [MLFlow Server](../servers/mlflow.html)

For these servers you only need the location of the saved model in a local filestore, Google bucket or S3 bucket. An example manifest with an sklearn server is shown below:

```
apiVersion: machinelearning.seldon.io/v1alpha2
kind: SeldonDeployment
metadata:
  name: sklearn
spec:
  name: iris
  predictors:
  - graph:
      children: []
      implementation: SKLEARN_SERVER
      modelUri: gs://seldon-models/sklearn/iris
      name: classifier
    name: default
    replicas: 1
```

The `modelUri` specifies the bucket containing the saved model, in this case `gs://seldon-models/sklearn/iris`.

`modeluri` supports the following three object storage providers:

  * Google Cloud Storage (using `gs://`)
  * S3-comptaible (using `s3://`)
  * Azure Blob storage (using `https://(.+?).blob.core.windows.net/(.+)`)
  

## Gunicorn Alpha Feature

We have provided an early alpha release for the python language wrapper to run under [gunicorn](https://gunicorn.org/) rather than Flask. For further details see our [gunicorn documentation](../python/python_component.html#gunicorn-alpha-feature).

## Kustomize Integration

We have a [kustomize resource](https://github.com/SeldonIO/seldon-core/tree/master/kustomize/seldon-core-operator) you can use and extend for your own particular setup for installing Seldon Core.

## More Example Integrations

Our range of example has expanded to include:

 * [Tabular Model Explanations using Seldon Alibi](../examples/alibi_anchor_tabular.html)
 * [Alibaba MNIST](../examples/alibaba_ack_deep_mnist.html)
 * [MLFLow Model Server](../examples/mlflow_server_ab_test_ambassador.html)





