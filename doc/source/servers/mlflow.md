# MLflow Server

If you have a trained an MLflow model you are able to deploy one (or several)
of the versions saved using Seldon's prepackaged MLflow server.
During initialisation, the built-in reusable server will create the [Conda
environment](https://www.mlflow.org/docs/latest/projects.html#project-environments)
specified on your `conda.yaml` file.

## Pre-requisites

To use the built-in MLflow server the following pre-requisites need to be met:

- Your [MLmodel artifact
  folder](https://www.mlflow.org/docs/latest/models.html) needs to be
  accessible remotely (e.g. as `gs://seldon-models/mlflow/elasticnet_wine`).
- Your model needs to be compatible with the [python_function
  flavour](https://www.mlflow.org/docs/latest/models.html#python-function-python-function).
- Your `MLproject` environment needs to be specified using Conda.

## Conda environment creation

The MLflow built-in server will create the Conda environment specified on your
`MLmodel`'s `conda.yaml` file during initialisation.
Note that this approach may slow down your Kubernetes `SeldonDeployment`
startup time considerably.

In some cases, it may be worth consider creating your own custom reusable
server using
[s2i](https://docs.seldon.io/projects/seldon-core/en/latest/python/python_wrapping_s2i.html)
with the `seldonio/seldon-core-s2i-conda` template image.
For example, when the Conda environment can be considered stable, you can
create your own image with a fixed set of dependencies.
This image can then be re-used across different model versions using the same
pre-loaded environment.

## Examples

An example for a saved Iris prediction model can be found below:

```yaml
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

You can also try out a [worked
notebook](../examples/server_examples.html#Serve-MLflow-Elasticnet-Wines-Model)
or check our [talk at the Spark + AI Summit
2019](https://www.youtube.com/watch?v=D6eSfd9w9eA).
