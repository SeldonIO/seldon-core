# Seldon Python Component

To create your component to run under Seldon you should create a class that implements the signatures needed for the type of component you are creating.

## Model

To wrap your machine learning model create a Class that has a predict method with the following signature:

```python
    def predict(self, X: np.ndarray, names: Iterable[str], meta: Dict = None) -> Union[np.ndarray, List, str, bytes]:
```

Your predict method will receive a numpy array `X` with iterable set of column names (if they exist in the input features) and optional Dictionary of meta data. It should return the result of the prediction as either:

  * Numpy array
  * List of values
  * String or Bytes

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

  You can follow [various notebook examples](../examples/notebooks.html).

## Transformers

Seldon Core allows you to create components to transform features either in the input request direction (input transformer) or the output response direction (output transformer). For these components create methods with the signatures below:

```python
    def transform_input(self, X: np.ndarray, names: Iterable[str], meta: Dict = None) -> Union[np.ndarray, List, str, bytes]:

    def transform_output(self, X: np.ndarray, names: Iterable[str], meta: Dict = None) -> Union[np.ndarray, List, str, bytes]:
```

## Combiners

Seldon Core allows you to create components that combine responses from multiple models into a single response. To create a class for this add a method with signature below:

```python
    def aggregate(self, features_list: List[Union[np.ndarray, str, bytes]], feature_names_list: List) -> Union[np.ndarray, List, str, bytes]:
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

This method should return a Dictionary of metrics as described in the [custom metrics](../analytics/custom_metrics.md) docs.

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
            {"type":"COUNTER","key":"mycounter","value":1}, # a counter which will increase by the given value
            {"type":"GAUGE","key":"mygauge","value":100}, # a gauge which will be set to given value
            {"type":"TIMER","key":"mytimer","value":20.2}, # a timer which will add sum and count metrics - assumed millisecs
            ]

```

## Returning Tags
If we wish to add arbitrary tags to the returned metadata you can provide a `tags` method with signature as shown below:

```python
    def tags(self) -> Dict:
```

A simple example is shown below:

```python
class ModelWithMetrics(object):

    def predict(self,X,features_names):
        return X

    def tags(self,X):
    	return {"system":"production"}
```




## Low level Methods
If you want more control you can provide a low-level methods that will provide as input the raw proto buffer payloads. The signatures for these are shown below for release `sedon_core>=0.2.6.1`:

```python
    def predict_raw(self, msg: prediction_pb2.SeldonMessage) -> prediction_pb2.SeldonMessage:

    def send_feedback_raw(self, feedback: prediction_pb2.Feedback) -> prediction_pb2.SeldonMessage:

    def transform_input_raw(self, msg: prediction_pb2.SeldonMessage) -> prediction_pb2.SeldonMessage:

    def transform_output_raw(self, msg: prediction_pb2.SeldonMessage) -> prediction_pb2.SeldonMessage:

    def route_raw(self, msg: prediction_pb2.SeldonMessage) -> prediction_pb2.SeldonMessage:

    def aggregate_raw(self, msgs: prediction_pb2.SeldonMessageList) -> prediction_pb2.SeldonMessage:
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
### Gunicorn (Alpha Feature)

To run your class under gunicorn set the environment variable `GUNICORN_WORKERS` to an integer value > 1.

```
apiVersion: machinelearning.seldon.io/v1alpha2
kind: SeldonDeployment
metadata:
  name: gunicorn
spec:
  name: worker
  predictors:
  - componentSpecs:
    - spec:
        containers:
        - image: seldonio/mock_classifier:1.0
          name: classifier
          env:
          - name: GUNICORN_WORKERS
            value: '4'
        terminationGracePeriodSeconds: 1
    graph:
      children: []
      endpoint:
        type: REST
      name: classifier
      type: MODEL
    labels:
      version: v1
    name: example
    replicas: 1

```



## Gunicorn and load

If the wrapped python class is run under [gunicorn](https://gunicorn.org/) then as part of initialization of each gunicorn worker a `load` method will be called on your class if it has it. You should use this method to load and initialise your model. This is important for Tensorflow models which need their session created in each worker process. The [Tensorflow MNIST example](../examples/deep_mnist.html) does this.

```
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

## Multi-value numpy arrays

By default, when using the data ndarray parameter, the conversion to ndarray (by default) converts all inner types into the same type. With models that may take as input arrays with different value types, you will be able to do so by overriding the `predict_raw` function yourself which gives you access to the raw request, and creating the numpy array as follows:

```
import numpy as np
class Model:
    def predict_raw(self, request):
        data = request.get("data", {}).get("ndarray")
        if data:
            mult_types_array = array(data, dtype=object)

        # Handle other data types as required + your logic
```

## Next Steps

After you have created the Component you need to create a Docker image that can be managed by Seldon Core. Follow the documentation to do this with [s2i](./python_wrapping_s2i.md) or [Docker](./python_wrapping_docker.md).



