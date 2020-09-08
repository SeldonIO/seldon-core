# XGBoost Server

If you have a trained XGBoost model saved you can deploy it simply using Seldon's prepackaged XGBoost server.

Prequisites:

  * The model must be named `model.bst`
  * You must save your model using `bst.save_model(file_path)`
  * The model is loaded with `xgb.Booster(model_file=model_file)`
  * Dependencies (otherwise it may not work):
      + scikit-learn == 0.20.3
      + numpy == 1.15.4
      + xgboost == 1.0.1

An example for a saved Iris prediction model:

```
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


Try out a [worked notebook](../examples/server_examples.html)
