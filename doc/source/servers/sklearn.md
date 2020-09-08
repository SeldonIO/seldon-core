# SKLearn Server

If you have a trained SKLearn model saved as a pickle you can deploy it simply using Seldon's prepackaged SKLearn server.

Pre-requisites:

  * The model pickle must be saved using joblib and presently be named `model.joblib`
  * Installed dependencies (may not work if versions don't match):
      + sklearn == 0.23.2
      + joblib == 0.16.0
      + numpy >= 1.8.2

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

Acceptable values for the `method` parameter are `predict`, `predict_proba`, `decision_function`.

Try out a [worked notebook](../examples/server_examples.html)

## Version

The version of sklearn used depends on the version of seldon install as follows:

| Seldon Version | SKLearn Version |
| -------------- | --------------- |
| >=1.3          | 0.23.2          |
| <1.3 (latest 1.2.3)          | 0.20.3          |

If you wish to use an earlier sklearn image from seldon you can set the image in the componentSpecs, e.g.

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
         image: seldonio/sklearnserver_rest:1.2.3
    graph:
      children: []
      implementation: SKLEARN_SERVER
      modelUri: gs://seldon-models/sklearn/iris
      name: classifier
    name: default
    replicas: 1
    svcOrchSpec: 
      env: 
      - name: SELDON_LOG_LEVEL
        value: DEBUG
```

If you wish to use a different version of sklearn then you should build your own image from the code in https://github.com/SeldonIO/seldon-core/tree/master/servers/sklearnserver and set that image as above.

If you wish the server image for the sklearn server to be globally changed you can also change the configMap used by the Seldon Operator. For the helm chart this can be done by editing the `values.yaml` which contains the images to use for each server. For example:

```
  SKLEARN_SERVER:
    grpc:
      defaultImageVersion: "1.2.3"
      image: seldonio/sklearnserver_grpc
    rest:
      defaultImageVersion: "1.2.3"
      image: seldonio/sklearnserver_rest
```
