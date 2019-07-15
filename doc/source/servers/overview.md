# Prepackaged Model Servers

Seldon provides several prepacked servers you can use to deploy trained models:

 * [SKLearn Server](./sklearn.html)
 * [XGBoost Server](xgboost.html)
 * [Tensorflow Serving](tensorflow.html)


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

If you want to customize the resources for the server you can add a skeleton `Container` with the same name to your podSpecs, e.g.

```
apiVersion: machinelearning.seldon.io/v1alpha2
kind: SeldonDeployment
metadata:
  name: sklearn
spec:
  name: iris
  predictors:
  - componentSpecs:
    - spec:
        containers:
        - name: classifier
          resources:
            requests:
              memory: 50Mi
    graph:
      children: []
      implementation: SKLEARN_SERVER
      modelUri: gs://seldon-models/sklearn/iris
      name: classifier
    name: default
    replicas: 1

```

The image name and other details will be added when this is deployed automatically.

Next steps:

   * [Worked notebook](../examples/server_examples.html)
   * [SKLearn Server](./sklearn.html)
   * [XGBoost Server](xgboost.html)
   * [Tensorflow Serving](tensorflow.html)

If your use case does not fall into the above standard servers then you can create your own component using our wrappers.

