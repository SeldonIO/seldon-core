# Packaging a Python model for Seldon Core using Docker


In this guide, we illustrate the steps needed to wrap your own python model in a docker image ready for deployment with Seldon Core using Docker.

## Step 1 - Create your source code

You will need:

 * A python file with a class that runs your model
 * A requirements.txt with a seldon-core entry

We will go into detail for each of these steps:

### Python file
Your source code should contain a python file which defines a class of the same name as the file. For example, looking at our skeleton python model file at `wrappers/s2i/python/test/model-template-app/MyModel.py`:

```python
class MyModel(object):
    """
    Model template. You can load your model parameters in __init__ from a location accessible at runtime
    """

    def __init__(self):
        """
        Add any initialization parameters. These will be passed at runtime from the graph definition parameters defined in your seldondeployment kubernetes resource manifest.
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

### requirements.txt
Populate a requirements.txt with any software dependencies your code requires. At a minimum the file should contain:

```text
seldon-core
```

## Step 2 - Define the Dockerfile

Define a Dockerfile in the same directory as your source code and requirements.txt. It will define the core parameters needed by our python builder image to wrap your model as env vars. An example is:

```dockerfile
FROM python:3.7-slim
WORKDIR /app

# Install python packages
COPY requirements.txt requirements.txt
RUN pip install -r requirements.txt

# Copy source code
COPY . .

# Port for GRPC
EXPOSE 5000
# Port for REST
EXPOSE 9000

# Define environment variables
ENV MODEL_NAME MyModel
ENV SERVICE_TYPE MODEL

# Changing folder to default user
RUN chown -R 8888 /app

CMD exec seldon-core-microservice $MODEL_NAME --service-type $SERVICE_TYPE
```


## Step 3 - Build your image
Use `docker build . -t $ORG/$MODEL_NAME:$TAG` to create your Docker image from source code. A simple name can be used but convention is to use the ORG/IMAGE:TAG format.

## Using with Keras/Tensorflow Models

To ensure Keras models with the Tensorflow backend work correctly you may need to call `_make_predict_function()` on your model after it is loaded. This is because Flask may call the prediction request in a separate thread from the one that initialised your model. See the [keras issue](https://github.com/keras-team/keras/issues/6462) for further discussion.

## Environment Variables
The required environment variables understood by the builder image are explained below. You can provide them in the Dockerfile or as `-e` parameters to `docker run`.


### MODEL_NAME
The name of the class containing the model. Also the name of the python file which will be imported.

### SERVICE_TYPE

The service type being created. Available options are:

 * MODEL
 * ROUTER
 * TRANSFORMER
 * COMBINER
 * OUTLIER_DETECTOR


### Flask Settings

See [Flask - Builtin Configuration Values](https://flask.palletsprojects.com/config/#builtin-configuration-values) for possible configurations; the following are configurable when prefixed with the `FLASK_` string (e.g. `FLASK_JSON_SORT_KEYS` translates to `JSON_SORT_KEYS` in Flask):

 * DEBUG
 * EXPLAIN_TEMPLATE_LOADING
 * JSONIFY_PRETTYPRINT_REGULAR
 * JSON_SORT_KEYS
 * PROPAGATE_EXCEPTIONS
 * PRESERVE_CONTEXT_ON_EXCEPTION
 * SESSION_COOKIE_HTTPONLY
 * SESSION_COOKIE_SECURE
 * SESSION_REFRESH_EACH_REQUEST
 * TEMPLATES_AUTO_RELOAD
 * TESTING
 * TRAP_HTTP_EXCEPTIONS
 * TRAP_BAD_REQUEST_ERRORS

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
You can add arguments to your component which will be populated from the `parameters` defined in the SeldonDeloyment when you deploy your image on Kubernetes. For example, our [Python TFServing proxy](https://github.com/SeldonIO/seldon-core/tree/master/servers/tfserving) has the class init method signature defined as below:

```python
class TfServingProxy(object):

    def __init__(self,rest_endpoint=None,grpc_endpoint=None,model_name=None,signature_name=None,model_input=None,model_output=None):
```

These arguments can be set when deploying in a Seldon Deployment. An example can be found in the [MNIST TFServing example](https://github.com/SeldonIO/seldon-core/blob/master/examples/models/tfserving-mnist/tfserving-mnist.ipynb) where the arguments are defined in the [SeldonDeployment](https://github.com/SeldonIO/seldon-core/blob/master/examples/models/tfserving-mnist/mnist_tfserving_deployment.json.template)  which is partly show below:

```json
{
  "graph": {
    "name": "tfserving-proxy",
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
