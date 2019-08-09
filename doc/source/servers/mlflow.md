# SKLearn Server

If you have a trained an MLFlow model you are able to deploy one (or several) of the versions saved using Seldon's prepackaged MLFlow server.

Pre-requisites:

  * The direct path to the selected MLFlow model should be provided (for example, `gs://mlruns/0/540ee112155e46e682b35b2768ae7f4d/artefacts/model`).
  * The model should be compatible with MLFlow's [load_model](https://www.mlflow.org/docs/latest/python_api/mlflow.pyfunc.html#mlflow.pyfunc.load_model) function
  * The input to the model is set to be pandas by default, so the numpy array passed will be converted into a pandas dataframe
  * The model server was built with Pandas version 0.25, so model should be compatible with that version
  * The model server was built using MLFlow version 1.1.0, so the model should be compatible with that version

An example for a saved Iris prediction model:

```
apiVersion: machinelearning.seldon.io/v1alpha2
kind: SeldonDeployment
metadata:
  name: mlflow
spec:
  name: wines
  predictors:
  - graph:
      children: []
      implementation: MLFLOW_SERVER
      modelUri: gs://seldon-models/mlflow/elasticnet_wine
      name: classifier
    name: default
    replicas: 1

```

Try out a [worked notebook](../examples/server_examples.html)
