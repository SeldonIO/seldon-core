# Packaging a Python model for Seldon Core using s2i


In this guide, we illustrate the steps needed to wrap your own python model in a docker image ready for deployment with Seldon Core using [source-to-image app s2i](https://github.com/openshift/source-to-image).

If you are not familiar with s2i you can read [general instructions on using s2i](../wrappers/s2i.md) and then follow the steps below.


## Step 1 - Install s2i

 [Download and install s2i](https://github.com/openshift/source-to-image#installation)

 * Prerequisites for using s2i are:
   * Docker
   * Git (if building from a remote git repo)

To check everything is working you can run

```bash
s2i usage seldonio/seldon-core-s2i-python3:1.5.0-dev
```


## Step 2 - Create your source code

To use our s2i builder image to package your python model you will need:

 * A python file with a class that runs your model
 * Your model's dependencies and environment, which can be described using either of:
   - `requirements.txt`
   - `setup.py`
   - `environment.yaml`
 * `.s2i/environment` - model definitions used by the s2i builder to correctly wrap your model

We will go into detail for each of these steps:

### Python file
Your source code should contain a python file which defines a class of the same name as the file. For example, looking at our skeleton python model file at `wrappers/s2i/python/test/model-template-app/MyModel.py`:

```python
class MyModel(object):
    """
    Model template.
    You can load your model parameters in __init__ from a location accessible at runtime.
    """

    def __init__(self):
        """
        Add any initialization parameters.
        These will be passed at runtime from the graph definition parameters defined in your seldondeployment kubernetes resource manifest.
        """
        print("Initializing")

    def predict(self,X,features_names):
        """
        Return a prediction.

        Parameters
        ----------
        X : array-like
        feature_names : array of feature names (optional)
        """
        print("Predict called - will run identity function")
        return X
```

 * The file is called MyModel.py and it defines a class MyModel
 * The class contains a predict method that takes an array (numpy) X and feature_names and returns an array of predictions.
 * You can add any required initialization inside the class init method.
 * Your return array should be at least 2-dimensional.

### Dependencies

You can describe your model's dependencies using either of: `requirements.txt`,
`setup.py` or `environment.yaml`.

#### requirements.txt

Populate a `requirements.txt` with any software dependencies your code requires.
These will be installed via pip when creating the image.

#### setup.py

Similar to a `requirements.txt` file, you can also describe your model's
dependencies using a `setup.py` file:

```python
from setuptools import setup

setup(
  name="my-model",
  # ...
  install_requires=[
    "scikit-learn",
  ]
)
```

#### environment.yaml

Describe your Conda environment using an `environment.yaml` file:

```yaml
name: my-conda-environment
channels:
  - defaults
dependencies:
  - python=3.6
  - scikit-learn=0.19.1
```

During image creation, `s2i` will create your Conda environment, fetching all
the required dependencies.
At run time, the created Conda environment will get activated at startup.

### .s2i/environment

Define the core parameters needed by our python builder image to wrap your model. An example is:

```bash
MODEL_NAME=MyModel
API_TYPE=REST
SERVICE_TYPE=MODEL
PERSISTENCE=0
```

These values can also be provided or overridden on the command line when building the image.

## Step 3 - Build your image
Use `s2i build` to create your Docker image from source code. You will need Docker installed on the machine and optionally git if your source code is in a public git repo. You can choose from three python builder images

 * Python 3.6 : seldonio/seldon-core-s2i-python36:1.5.0-dev seldonio/seldon-core-s2i-python3:1.5.0-dev
   * Note there are [issues running TensorFlow under Python 3.7](https://github.com/tensorflow/tensorflow/issues/20444) (Nov 2018) and Python 3.7 is not officially supported by TensorFlow (Dec 2018).
 * Python 3.6 plus ONNX support via [Intel nGraph](https://github.com/NervanaSystems/ngraph) : seldonio/seldon-core-s2i-python3-ngraph-onnx:0.1

Using s2i you can build directly from a git repo or from a local source folder. See the [s2i docs](https://github.com/openshift/source-to-image/blob/master/docs/cli.md#s2i-build) for further details. The general format is:

```bash
s2i build <src-folder> seldonio/seldon-core-s2i-python3:1.5.0-dev <my-image-name>
```

Change to seldonio/seldon-core-s2i-python3 if using python 3.

An example invocation using the test template model inside seldon-core:

```bash
s2i build https://github.com/seldonio/seldon-core.git --context-dir=wrappers/s2i/python/test/model-template-app seldonio/seldon-core-s2i-python3:1.5.0-dev seldon-core-template-model
```

The above s2i build invocation:

 * uses the GitHub repo: https://github.com/seldonio/seldon-core.git and the directory `wrappers/s2i/python/test/model-template-app` inside that repo.
 * uses the builder image `seldonio/seldon-core-s2i-python3`
 * creates a docker image `seldon-core-template-model`


For building from a local source folder, an example where we clone the seldon-core repo:

```bash
git clone https://github.com/seldonio/seldon-core.git
cd seldon-core
s2i build wrappers/s2i/python/test/model-template-app seldonio/seldon-core-s2i-python3:1.5.0-dev seldon-core-template-model
```

For more help see:

```bash
s2i usage seldonio/seldon-core-s2i-python3:1.5.0-dev
s2i build --help
```

## Using with Keras/Tensorflow Models

To ensure Keras models with the Tensorflow backend work correctly you may need to call `_make_predict_function()` on your model after it is loaded. This is because Flask may call the prediction request in a separate thread from the one that initialised your model. See the [keras issue](https://github.com/keras-team/keras/issues/6462) for further discussion.

## Environment Variables
The required environment variables understood by the builder image are explained below. You can provide them in the `.s2i/environment` file or on the `s2i build` command line.


### MODEL_NAME

The name of the class containing the model. Also the name of the python file which will be imported.

### API_TYPE

API type to create. Can be REST or GRPC

### SERVICE_TYPE

The service type being created. Available options are:

 * MODEL
 * ROUTER
 * TRANSFORMER
 * COMBINER
 * OUTLIER_DETECTOR

### PERSISTENCE

Set either to 0 or 1. Default is 0. If set to 1 then your model will be saved periodically to redis and loaded from redis (if exists) or created fresh if not.

### EXTRA_INDEX_URL

.. Warning::
   ``EXTRA_INDEX_URL`` is recommended to be passed as argument to ``s2i``
   command rather than adding in ``.s2i/environment`` as a practice of avoiding
   checking in credentials in the code.

For installing packages from private/self-hosted PyPi registry.

### PIP_TRUSTED_HOST

For adding private/self-hosted unsecured PyPi registry by adding it to pip trusted-host.

```bash
s2i build \
   -e EXTRA_INDEX_URL=https://<pypi-user>:<pypi-auth>@mypypi.example.com/simple \
   -e PIP_TRUSTED_HOST=mypypi.example.com \
   <src-folder> \
   seldonio/seldon-core-s2i-python3:1.5.0-dev \
   <my-image-name>
```

### PAYLOAD_PASSTHROUGH

If enabled, the Python server won't try to decode the request payload nor
encode the response back.
That means that the `predict()` method of your `SeldonComponent` model will
receive the payload as-is and it will be responsible to decode it.
Likewise, the return value of `predict()` must be a serialised response.

By default, this option will be disabled.

## Creating different service types

### MODEL

 * [A minimal skeleton for model source code](https://github.com/SeldonIO/seldon-core/tree/master/wrappers/s2i/python/test/model-template-app)
 * [Example model notebooks](../examples/notebooks.html)

### ROUTER
 * [Description of routers in Seldon Core](../analytics/routers.html)
 * [A minimal skeleton for router source code](https://github.com/SeldonIO/seldon-core/tree/master/wrappers/s2i/python/test/router-template-app)

### TRANSFORMER

 * [A minimal skeleton for transformer source code](https://github.com/SeldonIO/seldon-core/tree/master/wrappers/s2i/python/test/transformer-template-app)
 * [Example transformers](https://github.com/SeldonIO/seldon-core/tree/master/examples/transformers)


## Advanced Usage

### Model Class Arguments
You can add arguments to your component which will be populated from the `parameters` defined in the SeldonDeloyment when you deploy your image on Kubernetes. For example, our [Python TFServing proxy](https://github.com/SeldonIO/seldon-core/tree/master/integrations/tfserving) has the class init method signature defined as below:

```python
class TfServingProxy(object):

    def __init__(self,rest_endpoint=None,grpc_endpoint=None,model_name=None,signature_name=None,model_input=None,model_output=None):
```

These arguments can be set when deploying in a Seldon Deployment. An example can be found in the [MNIST TFServing example](https://github.com/SeldonIO/seldon-core/blob/master/examples/models/tfserving-mnist/tfserving-mnist.ipynb) where the arguments are defined in the [SeldonDeployment](https://github.com/SeldonIO/seldon-core/blob/master/examples/models/tfserving-mnist/mnist_tfserving_deployment.json.template)  which is partly show below:

```json
{
  "graph": {
    "name": "tfserving-proxy",
    "endpoint": { "type": "REST" },
    "type": "MODEL",
    "children": [],
    "parameters": [
      {
        "name": "grpc_endpoint",
        "type": "STRING",
        "value": "localhost:8000"
      },
      {
        "name": "model_name",
        "type": "STRING",
        "value": "mnist-model"
      },
      {
        "name": "model_output",
        "type": "STRING",
        "value": "scores"
      },
      {
        "name": "model_input",
        "type": "STRING",
        "value": "images"
      },
      {
        "name": "signature_name",
        "type": "STRING",
        "value": "predict_images"
      }
    ]
  }
}
```


The allowable `type` values for the parameters are defined in the [proto buffer definition](https://github.com/SeldonIO/seldon-core/blob/44f7048efd0f6be80a857875058d23efc4221205/proto/seldon_deployment.proto#L117-L131).


### Local Python Dependencies
`from version 0.5-SNAPSHOT`

To use a private repository for installing Python dependencies use the following build command:

```bash
s2i build -i <python-wheel-folder>:/whl <src-folder> seldonio/seldon-core-s2i-python3:1.5.0-dev <my-image-name>
```

This command will look for local Python wheels in the `<python-wheel-folder>` and use these before searching PyPI.

### Custom Metrics
`from version 0.3`

To add custom metrics to your response you can define an optional method `metrics` in your class that returns a list of metric dicts. An example is shown below:

```python
class MyModel(object):

    def predict(self, X, features_names):
        return X

    def metrics(self):
        return [{"type": "COUNTER", "key": "mycounter", "value": 1}]
```

For more details on custom metrics and the format of the metric dict see [here](../analytics/analytics.html#custom-metrics).

There is an [example notebook illustrating a model with custom metrics in python](../examples/custom_metrics.html).

### Custom Request Tags
`from version 0.3`

To add custom request tags data you can add an optional method `tags` which can return a dict of custom meta tags as shown in the example below:

```python
class MyModel(object):

    def predict(self, X, features_names):
        return X

    def tags(self):
        return {"mytag": 1}
```
