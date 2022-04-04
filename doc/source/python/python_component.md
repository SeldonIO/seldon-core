# Seldon Python Component

In order to run a custom python model on Seldon Core, you first need to wrap the model in it's own Python Class.

## Model

To wrap your machine learning model create a Class that has a predict method with the following signature:

```python
    def predict(self, X: Union[np.ndarray, List, str, bytes, Dict], names: Optional[List[str]], meta: Optional[Dict] = None) -> Union[np.ndarray, List, str, bytes, Dict]:
```

Your predict method will receive a numpy array `X` with iterable set of column names (if they exist in the input features) and optional Dictionary of meta data. It should return the result of the prediction as either:

- Numpy array
- List of values
- String or Bytes
- Dictionary

A simple example is shown below:

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

    def predict(self, X, features_names=None):
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

### Returning class names

You can also provide a method to return the column names for your prediction with a method `class_names` with signature show below:

```python
    def class_names(self) -> Iterable[str]:
```

### Examples

You can follow [various notebook examples](../examples/notebooks.html#python-language-wrapper-examples).

## Transformers

Seldon Core allows you to create components to transform features either in the input request direction (input transformer) or the output response direction (output transformer). For these components create methods with the signatures below:

```python
    def transform_input(self, X: Union[np.ndarray, List, str, bytes, Dict], names: Optional[List[str]], meta: Optional[Dict] = None) -> Union[np.ndarray, List, str, bytes, Dict]:

    def transform_output(self, X: Union[np.ndarray, List, str, bytes, Dict], names: Optional[List[str]], meta: Optional[Dict] = None) -> Union[np.ndarray, List, str, bytes, Dict]:
```

## Combiners

Seldon Core allows you to create components that combine responses from multiple models into a single response. To create a class for this add a method with signature below:

```python
    def aggregate(self, features_list: List[Union[np.ndarray, str, bytes]], feature_names_list: List) -> Union[np.ndarray, List, str, bytes, Dict]:
```

A simple example that averages a set of responses is shown below:

```python
import numpy as np
logger = logging.getLogger(__name__)

class ImageNetCombiner(object):

    def aggregate(self, Xs, features_names):
        return (np.reshape(Xs[0],(1,-1)) + np.reshape(Xs[1], (1,-1)))/2.0
```

## Routers

Routers provide functionality to direct a request to one of a set of child components. For this you should create a method with signature as shown below that returns the `id` for the child component to route the request to. The `id` is the index of children connected to the router.

```python
    def route(self, features: Union[np.ndarray, str, bytes], feature_names: Iterable[str]) -> int:
```

To see examples of this you can follow the various [example routers](https://github.com/SeldonIO/seldon-core/tree/master/components/routers) that are part of Seldon Core.

## Adding Custom Metrics

To return metrics associated with a call create a method with signature as shown below:

```python
    def metrics(self) -> List[Dict]:
```

This method should return a Dictionary of metrics as described in the [custom metrics](../analytics/analytics.md#custom-metrics) docs.

An illustrative example is shown below:

```python
class ModelWithMetrics(object):

    def __init__(self):
        print("Initialising")

    def predict(self,X,features_names):
        print("Predict called")
        return X

    def metrics(self):
        return [
            {"type": "COUNTER", "key": "mycounter", "value": 1}, # a counter which will increase by the given value
            {"type": "GAUGE", "key": "mygauge", "value": 100},   # a gauge which will be set to given value
            {"type": "TIMER", "key": "mytimer", "value": 20.2},  # a timer which will add sum and count metrics - assumed millisecs
        ]
```

Note: prior to Seldon Core 1.1 custom metrics have always been returned to client. From SC 1.1 you can control this behaviour setting `INCLUDE_METRICS_IN_CLIENT_RESPONSE` environmental variable to either `true` or `false`. Despite value of this environmental variable custom metrics will always be exposed to Prometheus.

Prior to Seldon Core 1.1.0 not implementing custom metrics logs a message at the info level at each predict call. Starting with Seldon Core 1.1.0 this is logged at the debug level. To suppress this warning implement a metrics function returning an empty list:

```python
def metrics(self):
    return []
```

## Returning Tags

If we wish to add arbitrary tags to the returned metadata you can provide a `tags` method with signature as shown below:

```python
    def tags(self) -> Dict:
```

A simple example is shown below:

```python
class ModelWithTags(object):

    def predict(self,X,features_names):
        return X

    def tags(self):
        return {"system":"production"}
```


## Runtime Metrics and Tags

Starting from SC 1.3 `metrics` and `tags` can also be defined on the output of `predict`, `transform_input`, `transform_output`, `send_feedback`, `route` and `aggregate`.

This is thread-safe.

```python
from seldon_core.user_model import SeldonResponse


class Model:
    def predict(self, X, names=[], meta={}):
        runtime_metrics = {"type": "COUNTER", "key": "instance_counter", "value": len(X)},
        runtime_tags = {"runtime": "tag", "shared": "right one"}
        return SeldonResponse(data=X, metrics=runtime_metrics, tags=runtime_tags)

    def metrics(self):
        return [{"type": "COUNTER", "key": "requests_counter", "value": 1}]

    def tags(self):
        return {"static": "tag", "shared": "not right one"}
```

Note that `tags` and `metrics` defined through `SeldonResponse` take priority.
In above examples returned tags will be:
```json
{"runtime":"tag", "shared":"right one", "static":"tag"}
```



## REST Health Endpoint
If you wish to add a REST health point, you can implement the `health_status` method with signature as shown below:
```python
    def health_status(self) -> Union[np.ndarray, List, str, bytes]:
```

You can use this to verify that your service can respond to HTTP calls after you have built your docker image and also
as kubernetes liveness and readiness probes to verify that your model is healthy.

A simple example is shown below:

```python
class ModelWithHealthEndpoint(object):
    def predict(self, X, features_names):
        return X

    def health_status(self):
        response = self.predict([1, 2], ["f1", "f2"])
        assert len(response) == 2, "health check returning bad predictions" # or some other simple validation
        return response
```

When you use `seldon-core-microservice` to start the HTTP server, you can verify that the model is up and running by
checking the `/health/status` endpoint:
```console
$ curl localhost:5000/health/status
{"data":{"names":[],"tensor":{"shape":[2],"values":[1,2]}},"meta":{}}
```

Additionally, you can also use the `/health/ping` endpoint if you want a lightweight call that just checks that
the HTTP server is up:

```console
$ curl localhost:5000/health/ping
pong%
```

You can also override the default liveness and readiness probes and use HTTP health endpoints by adding them in your
`SeldonDeployment` YAML. You can modify the parameters for the probes to suit your reliability needs without putting
too much stress on the container. Read more about these probes in the
[kubernetes documentation](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/).
An example is shown below:

```yaml
apiVersion: machinelearning.seldon.io/v1alpha2
kind: SeldonDeployment
spec:
  name: my-app
  predictors:
  - componentSpecs:
    - spec:
        containers:
        - image: my-app-image:version
          name: classifier
          livenessProbe:
            failureThreshold: 3
            initialDelaySeconds: 60
            periodSeconds: 5
            successThreshold: 1
            httpGet:
              path: /health/status
              port: http
              scheme: HTTP
            timeoutSeconds: 1
          readinessProbe:
            failureThreshold: 3
            initialDelaySeconds: 20
            periodSeconds: 5
            successThreshold: 1
            httpGet:
              path: /health/status
              port: http
              scheme: HTTP
            timeoutSeconds: 1
```

## Low level Methods

If you want more control you can provide a low-level methods that will provide as input the raw proto buffer payloads. The signatures for these are shown below for release `seldon_core>=0.2.6.1`:

```python
    def predict_raw(self, msg: Union[Dict, prediction_pb2.SeldonMessage]) -> prediction_pb2.SeldonMessage:

    def send_feedback_raw(self, feedback: prediction_pb2.Feedback) -> prediction_pb2.SeldonMessage:

    def transform_input_raw(self, msg: Union[Dict, prediction_pb2.SeldonMessage]) -> prediction_pb2.SeldonMessage:

    def transform_output_raw(self, msg: Union[Dict, prediction_pb2.SeldonMessage]) -> prediction_pb2.SeldonMessage:

    def route_raw(self, msg: prediction_pb2.SeldonMessage) -> prediction_pb2.SeldonMessage:

    def aggregate_raw(self, msgs: prediction_pb2.SeldonMessageList) -> prediction_pb2.SeldonMessage:

    def health_status_raw(self) -> prediction_pb2.SeldonMessage:
```

## User Defined Exceptions

If you want to handle custom exceptions define a field `model_error_handler` as shown below:

```python
    model_error_handler = flask.Blueprint('error_handlers', __name__)
```

An example is as follows:

```python
"""
Model Template
"""
class MyModel(Object):

    """
    The field is used to register custom exceptions
    """
    model_error_handler = flask.Blueprint('error_handlers', __name__)

    """
    Register the handler for an exception
    """
    @model_error_handler.app_errorhandler(UserCustomException)
    def handleCustomError(error):
        response = jsonify(error.to_dict())
        response.status_code = error.status_code
        return response

    def __init__(self, metrics_ok=True, ret_nparray=False, ret_meta=False):
        pass

    def predict(self, X, features_names, **kwargs):
        raise UserCustomException('Test-Error-Msg',1402,402)
        return X
```

```python
"""
User Defined Exception
"""
class UserCustomException(Exception):

    status_code = 404

    def __init__(self, message, application_error_code,http_status_code):
        Exception.__init__(self)
        self.message = message
        if http_status_code is not None:
            self.status_code = http_status_code
        self.application_error_code = application_error_code

    def to_dict(self):
        rv = {"status": {"status": self.status_code, "message": self.message,
                         "app_code": self.application_error_code}}
        return rv

```

## Multi-value numpy arrays

By default, when using the data ndarray parameter, the conversion to ndarray (by default) converts all inner types into the same type. With models that may take as input arrays with different value types, you will be able to do so by overriding the `predict_raw` function yourself which gives you access to the raw request, and creating the numpy array as follows:

```python
import numpy as np

class Model:
    def predict_raw(self, request):
        data = request.get("data", {}).get("ndarray")
        if data:
            mult_types_array = np.array(data, dtype=object)

        # Handle other data types as required + your logic
```

## Gunicorn and load

If the wrapped python class is served under [Gunicorn](https://gunicorn.org/) then as
part of initialization of each gunicorn worker a `load` method will be called
on your class if it has it.
You should use this method to load and initialise your model.
This is important for Tensorflow models which need their session created in
each worker process.
The [Tensorflow MNIST example](../examples/deep_mnist.html) does this.

```python
import tensorflow as tf
import numpy as np
import os

class DeepMnist(object):
    def __init__(self):
        self.loaded = False
        self.class_names = ["class:{}".format(str(i)) for i in range(10)]

    def load(self):
        print("Loading model",os.getpid())
        self.sess = tf.Session()
        saver = tf.train.import_meta_graph("model/deep_mnist_model.meta")
        saver.restore(self.sess,tf.train.latest_checkpoint("./model/"))
        graph = tf.get_default_graph()
        self.x = graph.get_tensor_by_name("x:0")
        self.y = graph.get_tensor_by_name("y:0")
        self.loaded = True
        print("Loaded model")

    def predict(self,X,feature_names):
        if not self.loaded:
            self.load()
        predictions = self.sess.run(self.y,feed_dict={self.x:X})
        return predictions.astype(np.float64)
```

## Integer numbers

The `json` package in Python, parses numbers with no decimal part as integers.
Therefore, a tensor containing only numbers without a decimal part will get
parsed as an integer tensor.

To illustrate the above, we can consider the following example:

```json
{
  "data": {
    "ndarray": [0, 1, 2, 3]
  }
}
```

By default, the `json` package will parse the array in the `data.ndarray`
field as an array of Python `Integer` values.
Since there are no floating-point values, `numpy` will then create a tensor
with `dtype = np.int32`.

If we want to force a different behaviour, we can use the underlying `predict_raw()`
method to control the deserialisation of the input payload.
As an example, using the example above, we could force the resulting tensor to always
using `dtype = np.float64` by implementing `predict_raw()` as:

```python
import numpy as np

class Model:
    def predict_raw(self, request):
        data = request.get("data", {}).get("ndarray")
        if data:
            float_array = np.array(data, dtype=np.float64)

        # Make predictions using float_array
```


## Incubating features


### REST Metadata Endpoint
The python wrapper will automatically expose a `/metadata` endpoint to return metadata about the loaded model.
It is up to the developer to implement a `metadata` method in their class to provide a `dict` back containing the model metadata.

See metadata [documentation](../reference/apis/metadata.md) for more details.


#### Example format:
```python
class Model:
    ...

    def init_metadata(self):

        meta = {
            "name": "model-name",
            "versions": ["model-version"],
            "platform": "platform-name",
            "inputs": [{"name": "input", "datatype": "BYTES", "shape": [1]}],
            "outputs": [{"name": "output", "datatype": "BYTES", "shape": [1]}],
        }

        return meta
```

#### Validation
Output of developer-defined `metadata` method will be validated to follow the [V2 dataplane proposal](https://github.com/kserve/kserve/blob/master/docs/predict-api/v2/required_api.md#model-metadata) protocol, see [this](https://github.com/SeldonIO/seldon-core/issues/1638) GitHub issue for details:
```javascript
$metadata_model_response =
{
  "name" : $string,
  "versions" : [ $string, ... ], // optional
  "platform" : $string,
  "inputs" : [ $metadata_tensor, ... ],
  "outputs" : [ $metadata_tensor, ... ]
}
```
with
```javascript
$metadata_tensor =
{
  "name" : $string,
  "datatype" : $string,
  "shape" : [ $number, ... ]
}
```

If validation fails server will reply with `500` response `MICROSERVICE_BAD_METADATA` when requested for `metadata`.

#### Examples:
- [Basic Examples for Model with Metadata](../examples/metadata.html)
- [SKLearn Server example with MinIO](../examples/minio-sklearn.html)



## Next Steps

After you have created the Component you need to create a Docker image that can be managed by Seldon Core. Follow the documentation to do this with [s2i](./python_wrapping_s2i.md) or [Docker](./python_wrapping_docker.md).
