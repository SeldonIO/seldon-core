# MLflow Server

If you have a trained MLflow model you are able to deploy one (or several)
of the versions saved using Seldon's prepackaged MLflow server.
During initialisation, the built-in reusable server will create the [Conda
environment](https://www.mlflow.org/docs/latest/projects.html#project-environments)
specified on your `conda.yaml` file.

## Pre-requisites

To use the built-in MLflow server the following pre-requisites need to be met:

- Your [MLmodel artifact
  folder](https://www.mlflow.org/docs/latest/models.html) needs to be
  accessible remotely (e.g. as `gs://seldon-models/mlflow/elasticnet_wine_1.8.0`).
- Your model needs to be compatible with the [python_function
  flavour](https://www.mlflow.org/docs/latest/models.html#python-function-python-function).
- Your `MLproject` environment needs to be specified using Conda.

## Conda environment creation

The MLflow built-in server will create the Conda environment specified on your
`MLmodel`'s `conda.yaml` file during initialisation.
Note that this approach may slow down your Kubernetes `SeldonDeployment`
startup time considerably.

In some cases, it may be worth to consider [creating your own custom reusable
server](./custom.md).
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
        modelUri: gs://seldon-models/mlflow/elasticnet_wine_1.8.0
        name: classifier
      name: default
      replicas: 1
```

## MLFlow xtype

By default the server will call your loaded model's predict function with a `numpy.ndarray`. If you wish for it to call it with `pandas.DataFrame` instead, you can pass a parameter `xtype` and set it to `DataFrame`. For example:   

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
        modelUri: gs://seldon-models/mlflow/elasticnet_wine_1.8.0
        name: classifier
        parameters:
        - name: xtype
          type: STRING
          value: DataFrame
      name: default
      replicas: 1
```

You can also try out a [worked
notebook](../examples/server_examples.html#Serve-MLflow-Elasticnet-Wines-Model)
or check our [talk at the Spark + AI Summit
2019](https://www.youtube.com/watch?v=D6eSfd9w9eA).

## V2 KFServing protocol [Incubating]

.. Warning:: 
  Support for the V2 KFServing protocol is still considered an incubating
  feature.
  This means that some parts of Seldon Core may still not be supported (e.g.
  tracing, graphs, etc.).

The MLFlow server can also be used to expose an API compatible with the [V2
KFServing Protocol](../graph/protocols.md#v2-kfserving-protocol).
Note that, under the hood, it will use the [Seldon
MLServer](https://github.com/SeldonIO/MLServer) runtime.

### Create a model using `mlflow` and deploy to `seldon-core`
As an example we are going to use the elasticnet wine model.

1. Create a `conda` environment
```bash
$ conda -y create -n python3.8-mlflow-example python=3.8
$ conda activate python3.8-mlflow-example
```
2. Install `mlflow`
```bash
$ pip install mlflow
```
3. Train the elasticnet wine example
```bash
$ git clone https://github.com/mlflow/mlflow
$ cd mlflow/examples
$ python sklearn_elasticnet_wine/train.py
```
After the script ends, there will be a models persisted at `mlruns/0/<uuid>/artifacts/model`. This can
be fetched from the ui (`mlflow ui`)
4. Install additional packaged required to deploy and ack the conda environment using [conda-pack](https://conda.github.io/conda-pack/)
```bash
$ pip install conda-pack
$ pip install mlserver
$ pip install mlserver-mlflow
$ cd mlflow/examples/mlruns/0/<uuid>/artifacts/model
$ conda pack -o environment.tar.gz -f
```
This will pack the current conda environment to `environment.tar.gz`, this will be required by `mlserver` to create the same environment used during train for serving the model.
5. copy the model directory to a Google Storage bucket that is accessible by seldon-core
```bash
$ gsutil cp -r ../model gs://seldon-models/test/elasticnet_wine_<uuid>
```
6. deploy the model to seldon-core

In order to enable support for the V2 KFServing protocol, it's enough to
specify the `protocol` of the `SeldonDeployment` to use `kfserving`.
For example,

```yaml
apiVersion: machinelearning.seldon.io/v1alpha2
kind: SeldonDeployment
metadata:
  name: mlflow
spec:
  protocol: kfserving  # Activate the v2 protocol
  name: wines
  predictors:
    - graph:
        children: []
        implementation: MLFLOW_SERVER
        modelUri: gs://seldon-models/test/elasticnet_wine_<uuid>
        name: classifier
      name: default
      replicas: 1
```
7. get predictions from the deployed model
```python
import json

import requests

inference_request = {
    "parameters": {
        "content_type": "pd"
    },
    "inputs": [
        {
          "name": "fixed acidity",
          "shape": [1],
          "datatype": "FP32",
          "data": [7.4],
          "parameters": {
              "content_type": "np"
          }
        },
        {
          "name": "volatile acidity",
          "shape": [1],
          "datatype": "FP32",
          "data": [0.7000],
          "parameters": {
              "content_type": "np"
          }
        },
        {
          "name": "citric acidity",
          "shape": [1],
          "datatype": "FP32",
          "data": [0],
          "parameters": {
              "content_type": "np"
          }
        },
        {
          "name": "residual sugar",
          "shape": [1],
          "datatype": "FP32",
          "data": [1.9],
          "parameters": {
              "content_type": "np"
          }
        },
        {
          "name": "chlorides",
          "shape": [1],
          "datatype": "FP32",
          "data": [0.076],
          "parameters": {
              "content_type": "np"
          }
        },
        {
          "name": "free sulfur dioxide",
          "shape": [1],
          "datatype": "FP32",
          "data": [11],
          "parameters": {
              "content_type": "np"
          }
        },
        {
          "name": "total sulfur dioxide",
          "shape": [1],
          "datatype": "FP32",
          "data": [34],
          "parameters": {
              "content_type": "np"
          }
        },
        {
          "name": "density",
          "shape": [1],
          "datatype": "FP32",
          "data": [0.9978],
          "parameters": {
              "content_type": "np"
          }
        },
        {
          "name": "pH",
          "shape": [1],
          "datatype": "FP32",
          "data": [3.51],
          "parameters": {
              "content_type": "np"
          }
        },
        {
          "name": "sulphates",
          "shape": [1],
          "datatype": "FP32",
          "data": [0.56],
          "parameters": {
              "content_type": "np"
          }
        },
        {
          "name": "alcohol",
          "shape": [1],
          "datatype": "FP32",
          "data": [9.4],
          "parameters": {
              "content_type": "np"
          }
        },
    ]
}

endpoint = "http://localhost:8003/seldon/seldon/mlflow/v2/models/infer"
response = requests.post(endpoint, json=inference_request)

print(json.dumps(response.json(), indent=2))
```

### Caveats
- The version of `mlserver` installed in the conda environment will need to match the supported version in `seldon-core`. We are working on tooling to make this more seamless.
- Check the caveats of using [`conda-pack`](https://conda.github.io/conda-pack/#caveats)