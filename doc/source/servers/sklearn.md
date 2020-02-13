# SKLearn Server

If you have a trained SKLearn model saved as a pickle you can deploy it simply using Seldon's prepackaged SKLearn server.

Pre-requisites:

  * The model pickle must be saved using joblib and presently be named `model.joblib`
  * We presently use sklearn version 0.20.3. Your pickled model must be compatible with this version

An example for a saved Iris prediction model:

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

## Sklearn Method

By default the server will call `predict_proba` on your loaded model/pipeline. If you wish for it to call `predict` instead you can pass a parameter `method` and set it to `predict`. For example:

```
apiVersion: machinelearning.seldon.io/v1alpha2
kind: SeldonDeployment
metadata:
  name: sklearn
spec:
  name: iris-predict
  predictors:
  - graph:
      children: []
      implementation: SKLEARN_SERVER
      modelUri: gs://seldon-models/sklearn/iris
      name: classifier
      parameters:
        - name: method
          type: STRING
          value: predict
    name: default
    replicas: 1
```

Try out a [worked notebook](../examples/server_examples.html)
