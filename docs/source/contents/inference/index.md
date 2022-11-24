# Inference

This section will discuss how to make inference calls against your Seldon models or pipelines.

You can make synchronous inference requests via REST or gRPC or asynchronous requests via Kafka topics.
The content of your request should be an [inference V2 protocol payload](../apis/inference/v2.md):
* REST payloads will generally be in the JSON v2 protocol format.
* gRPC and Kafka payloads **must** be in the Protobuf v2 protocol format.

## Synchronous Requests

For making synchronous requests, the process will generally be:
1. Find the appropriate service endpoint (IP address and port) for accessing the installation of Seldon Core v2.
1. Determine the appropriate headers/metadata for the request.
1. Make requests via REST or gRPC.

### Find the Seldon Service Endpoint

`````{tabs}

````{tab} Docker Compose

In the default Docker Compose setup, container ports are accessible from the host machine.
This means you can use `localhost` or `0.0.0.0` as the hostname.

The default port for sending inference requests to the Seldon system is `9000`.
This is controlled by the `ENVOY_DATA_PORT` environment variable for Compose.

Putting this together, you can send inference requests to `0.0.0.0:9000`.
````

````{tab} Kubernetes

In Kubernetes, Seldon creates a single `Service` called `seldon-mesh` in the namespace it is installed into.
By default, this namespace is also called `seldon-mesh`.

If this `Service` is exposed via a load balancer, the appropriate address and port can be found via:

```bash
kubectl get svc seldon-mesh -n seldon-mesh -o jsonpath='{.status.loadBalancer.ingress[0].ip}'
```

If you are not using a `LoadBalancer` for the `seldon-mesh` `Service`, you can still send inference requests.

For development and testing purposes, you can port-forward the `Service` locally using the below.
Inference requests can then be sent to `localhost:8080`.

```
kubectl port-forward svc/seldon-mesh -n seldon-mesh 8080:80
```
````

`````

### Make Inference Requests

Seldon routes requests to to the correct endpoint via headers in HTTP calls.
You should set the header `seldon-model` as follows:

 * Models: use the model name, e.g. for a model named `mymodel` use `seldon-model: mymodel`
 * Pipelines: use the pipeline name with the suffix `.pipeline`, e.g. for a pipeline named `mypipeline` use `seldon-model: mypipeline.pipeline`

The `seldon` CLI can be used to easily send requests to your deployed resources. See the [examples](../examples/index) and the [Seldon CLI docs](../cli/index.md).

An example curl request might look like for a model called `iris`:

```
curl -v http://0.0.0.0:9000/v2/models/iris/infer -H "Content-Type: application/json" -H "seldon-model: iris"\
        -d '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}'
```

A request to the same model using `grpcurl` might look like:

```
grpcurl -d '{"model_name":"iris","inputs":[{"name":"input","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[1,4]}]}' \
        -plaintext \
	-import-path apis \
	-proto apis/mlops/v2_dataplane/v2_dataplane.proto \
	-rpc-header seldon-model:iris \
	0.0.0.0:9000 inference.GRPCInferenceService/ModelInfer
```

The above request was run from the project root folder allowing reference to the Protobuf manifests defined in the `apis/` folder.

For pipelines a synchronous request is possible if the pipeline has an outputs section in the spec.

### Using Python Tritonclient

You can also use the Python [tritonclient](https://github.com/triton-inference-server/client) package to send inference requests.

A short self-contained example corresponding to the above requests is:
```python
import tritonclient.http as httpclient
import numpy as np

client = httpclient.InferenceServerClient(
    url="172.19.255.9:80",
    verbose=False,
)

inputs = [httpclient.InferInput("predict", (1, 4), "FP64")]
inputs[0].set_data_from_numpy(np.array([[1, 2, 3, 4]]).astype("float64"), binary_data=False)

result = client.infer("iris", inputs)
print("result is:", result.as_numpy("predict"))
```

## Asynchronous Requests

The Seldon architecture uses Kafka and therefore asynchronous requests can be sent by pushing V2 protocol payloads to the appropriate topic.
Topics have the following form:

```
seldon.<namespace>.<model|pipeline>.<name>.<inputs|outputs>
```

### Model Inference

For a local install if you have a model `iris`, you would be able to send a prediction request by pushing to the topic: `seldon.default.model.iris.inputs`.
The response will appear on `seldon.default.model.iris.outputs`.

For a Kubernetes install in `seldon-mesh` if you have a model `iris`, you would be able to send a prediction request by pushing to the topic: `seldon.seldon-mesh.model.iris.inputs`.
The response will appear on `seldon.seldon-mesh.model.iris.outputs`.


### Pipeline Inference

For a local install if you have a pipeline `mypipeline`, you would be able to send a prediction request by pushing to the topic: `seldon.default.pipeline.mypipeline.inputs`. The response will appear on `seldon.default.pipeline.mypipeline.outputs`.

For a Kubernetes install in `seldon-mesh` if you have a pipeline `mypipeline`, you would be able to send a prediction request by pushing to the topic: `seldon.seldon-mesh.pipeline.mypipeline.inputs`. The response will appear on `seldon.seldon-mesh.pipeline.mypipeline.outputs`.


## Pipeline Metadata

It may be useful to send metadata alongside your inference.

If using Kafka directly as described above, you can attach Kafka metadata to your request, which will be passed around the graph.
When making synchronous requests to your pipeline with REST or gRPC you can also do this.

 * For REST requests add HTTP headers prefixed with `X-`
 * For gRPC requests add metadata with keys starting with `X-`

You can also do this with the Seldon CLI by setting headers with the `--header` argument (and also showing response headers with the `--show-headers` argument)

```
seldon pipeline infer --show-headers --header X-foo=bar tfsimples \
    '{"inputs":[{"name":"INPUT0","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,16]},{"name":"INPUT1","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,16]}]}'
```

## Request IDs

For both model and pipeline requests the response will contain a `x-request-id` response header. For pipeline requests this can be used to inspect the pipeline steps via the CLI, e.g.:

```
seldon pipeline inspect tfsimples --request-id carjjolvqj3j2pfbut10 --offset 10
```

The `--offset` parameter specifies how many messages (from the latest) you want to search to find your request. If not specified the last request will be shown.

`x-request-id` will also appear in tracing spans.

If `x-request-id` is passed in by the caller then this will be used. It is the caller's responsibility to ensure it is unique.

The IDs generated are [XIDs](https://github.com/rs/xid).
