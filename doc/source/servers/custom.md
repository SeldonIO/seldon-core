# Custom inference servers

Out of the box, Seldon offers support for several [pre-packaged inference
servers](./overview.md).
However, there may be cases where it makes sense to rollout your own re-usable
inference server.
For example, you may need particular dependencies, specific versions, a custom
process to download your model weights, etc.

To support these use cases, Seldon allows you to easily build your own
inference servers, which can then be configured to be used as you would do the
pre-packaged ones.
That is, by using the `implementation` key and passing through model parameters
(e.g. `modelUri`).

```yaml
apiVersion: machinelearning.seldon.io/v1alpha2
kind: SeldonDeployment
metadata:
  name: nlp-model
spec:
  predictors:
    - graph:
        children: []
        implementation: CUSTOM_INFERENCE_SERVER
        modelUri: s3://our-custom-models/nlp-model
        name: model
      name: default
      replicas: 1
```

## Building a new inference server

To build a new custom inference server you can follow the same instructions as
the ones on [how to wrap a model](../wrappers/language_wrappers.md).
The only difference is that the server will now receive a set of model
parameters, like `modelUri`.

As inspiration, you can see how the [SKLearn
server](https://github.com/SeldonIO/seldon-core/blob/d84b97431c49602d25f6f5397ba540769ec695d9/servers/sklearnserver/sklearnserver/SKLearnServer.py#L16-L23)
and the [other pre-packaged inference servers](./overview.md) handle these as
part of their `__init__()` method:

```python
def __init__(self, model_uri: str = None,  method: str = "predict_proba"):
        super().__init__()
        self.model_uri = model_uri
        self.method = method
        self.ready = False
        self.load()
```

## Adding a new inference server

The list of available inference servers in Seldon Core is maintained in the
`seldon-config` configmap, which lives in the same namespace as your Seldon
Core operator.
In particular, the `predictor_servers` key holds the JSON config for each
inference server.

The `predictor_servers` key will hold a JSON dictionary similar to the one
below:

```json
{
  "SKLEARN_SERVER": {
    "grpc": {
      "defaultImageVersion": "0.2",
      "image": "seldonio/sklearnserver_grpc"
    },
    "rest": {
      "defaultImageVersion": "0.2",
      "image": "seldonio/sklearnserver_rest"
    }
  }
}
```

Adding a new inference server is just a matter of adding a new key to the dict
above.
For example:

```json
{
  "CUSTOM_INFERENCE_SERVER": {
    "grpc": {
      "defaultImageVersion": "1.0",
      "image": "org/custom-server-grpc"
    },
    "rest": {
      "defaultImageVersion": "1.0",
      "image": "org/custom-server-rest"
    }
  },
  "SKLEARN_SERVER": {
    "grpc": {
      "defaultImageVersion": "0.2",
      "image": "seldonio/sklearnserver_grpc"
    },
    "rest": {
      "defaultImageVersion": "0.2",
      "image": "seldonio/sklearnserver_rest"
    }
  }
}
```

## Worked Example

A worked example to build a LighGBM Model server can be found [here](../examples/custom_server.html)