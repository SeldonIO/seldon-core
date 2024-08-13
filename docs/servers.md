# Servers

By default Seldon installs two server farms using MLServer and Triton with 1 replica each. Models are scheduled onto servers based on the server's resources and whether the capabilities of the server matches the requirements specified in the Model request. For example:

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Model
metadata:
  name: iris
spec:
  storageUri: "gs://seldon-models/scv2/samples/mlserver_1.5.0/iris-sklearn"
  requirements:
  - sklearn
  memory: 100Ki
```

This model specifies the requirement `sklearn`

There is a default capabilities for each server as follows:

* MLServer
  ```yaml
  - name: SELDON_SERVER_CAPABILITIES
    value: "mlserver,alibi-detect,alibi-explain,huggingface,lightgbm,mlflow,python,sklearn,spark-mlib,xgboost"
  ```
* Triton
  ```yaml
  - name: SELDON_SERVER_CAPABILITIES
    value: "triton,dali,fil,onnx,openvino,python,pytorch,tensorflow,tensorrt"
  ```

## Custom Capabilities
Servers can be defined with a `capabilities` field to indicate custom configurations (e.g. Python dependencies). For instance:

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Server
metadata:
  name: mlserver-134
spec:
  serverConfig: mlserver
  capabilities:
  - mlserver-1.3.4
  podSpec:
    containers:
    - image: seldonio/mlserver:1.3.4
      name: mlserver
```

These `capabilities` override the ones from the `serverConfig: mlserver`. A model that takes advantage of this is shown below:

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Model
metadata:
  name: iris
spec:
  storageUri: "gs://seldon-models/mlserver/iris"
  requirements:
  - mlserver-1.3.4
```

This above model will be matched with the previous custom server `mlserver-134`.

Servers can also be set up with the `extraCapabilities` that add to existing capabilities from the referenced ServerConfig. For instance:

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Server
metadata:
  name: mlserver-extra
spec:
  serverConfig: mlserver
  extraCapabilities:
  - extra
```
This server, `mlserver-extra`, inherits a default set of capabilities via `serverConfig: mlserver`.
These defaults are discussed above.
The `extraCapabilities` are appended to these to create a single list of capabilities for this server.

Models can then specify requirements to select a server that satisfies those requirements as follows.
```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Model
metadata:
  name: extra-model-requirements
spec:
  storageUri: "gs://seldon-models/mlserver/iris"
  requirements:
  - extra
```

The `capabilities` field takes precedence over the `extraCapabilities` field.

For some examples see [here](../examples/custom-servers.md).


## Autoscaling of Servers

Within docker we don't support this but for Kubernetes see [here](../kubernetes/autoscaling/README.md)
