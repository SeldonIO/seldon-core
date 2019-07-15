# SKLearn Server

If you have a trained SKLearn model saved as a pickle you can deploy it simply using Seldon's prepackaged SKLearn server.

Prequisites:

  * The model pickle must be saved using joblib and presently be named `model.joblib`
  * We presently use sklearn version 0.20.3. Your pickled model must be compatbible with this version

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


Try out a [worked notebook](../examples/server_examples.html)