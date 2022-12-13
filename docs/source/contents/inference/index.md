# Inference

This section will discuss how to make inference calls against your Seldon models or pipelines.

You can make synchronous inference requests via REST or gRPC or asynchronous requests via Kafka topics.
The content of your request should be an [inference v2 protocol payload](../apis/inference/v2.md):
* REST payloads will generally be in the JSON v2 protocol format.
* gRPC and Kafka payloads **must** be in the Protobuf v2 protocol format.

## Synchronous Requests

For making synchronous requests, the process will generally be:
1. Find the appropriate service endpoint (IP address and port) for accessing the installation of Seldon Core v2.
2. Determine the appropriate headers/metadata for the request.
3. Make requests via REST or gRPC.

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

If you are using a service mesh like Istio or Ambassador, you will need to use the IP address of the service mesh ingress and determine the appropriate port.
````

`````

### Make Inference Requests

Let us imagine making inference requests to a model called `iris`.

This `iris` model has the following schema, which can be set in a `model-settings.json` file for MLServer:

```
{
    "name": "iris",
    "implementation": "mlserver_sklearn.SKLearnModel",
    "inputs": [
        {
            "name": "predict",
            "datatype": "FP32",
            "shape": [-1, 4]
        }
    ],
    "outputs": [
        {
            "name": "predict",
            "datatype": "INT64",
            "shape": [-1, 1]
        }
    ],
    "parameters": {
        "version": "1"
    }
}
```

Examples are given below for some common tools for making requests.

`````{tabs}

````{group-tab} Seldon CLI

An example `seldon` request might look like this:

```
seldon model infer iris \
        '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}'
```

The default inference mode is REST, but you can also send gRPC requests like this:

```
seldon model infer iris \
        --inference-mode grpc \
        '{"model_name":"iris","inputs":[{"name":"input","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[1,4]}]}'
```
````

````{group-tab} cURL

An example `curl` request might look like this:

```
curl -v http://0.0.0.0:9000/v2/models/iris/infer \
        -H "Content-Type: application/json" \
        -d '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}'
```
````

````{group-tab} grpcurl

An example `grpcurl` request might look like this:

```
grpcurl \
	-d '{"model_name":"iris","inputs":[{"name":"input","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[1,4]}]}' \
	-plaintext \
	-import-path apis \
	-proto apis/mlops/v2_dataplane/v2_dataplane.proto \
	0.0.0.0:9000 inference.GRPCInferenceService/ModelInfer
```

The above request was run from the project root folder allowing reference to the Protobuf manifests defined in the `apis/` folder.
````

````{group-tab} Python tritonclient

You can use the Python [tritonclient](https://github.com/triton-inference-server/client) package to send inference requests.

A short, self-contained example is:

```python
import tritonclient.http as httpclient
import numpy as np

client = httpclient.InferenceServerClient(
    url="localhost:8080",
    verbose=False,
)

inputs = [httpclient.InferInput("predict", (1, 4), "FP64")]
inputs[0].set_data_from_numpy(
    np.array([[1, 2, 3, 4]]).astype("float64"),
    binary_data=False,
)

result = client.infer("iris", inputs)
print("result is:", result.as_numpy("predict"))
```
````
`````

```{tip}
For pipelines, a synchronous request is possible if the pipeline has an `outputs` section defined in its spec.
```

### Request Routing

#### Seldon Routes

Seldon needs to determine where to route requests to, as models and pipelines might have the same name.
There are two ways of doing this: header-based routing (preferred) and path-based routing.

`````{tabs}

````{tab} Headers

Seldon can route requests to the correct endpoint via headers in HTTP calls, both for REST (HTTP/1.1) and gRPC (HTTP/2).

Use the `Seldon-Model` header as follows:
* For models, use the model name as the value.
  For example, to send requests to a model named `foo` use the header `Seldon-Model: foo`.
* For pipelines, use the pipeline name followed by `.pipeline` as the value.
  For example, to send requests to a pipeline named `foo` use the header `Seldon-Model: foo.pipeline`.

The `seldon` CLI is aware of these rules and can be used to easily send requests to your deployed resources.
See the [examples](../examples/index) and the [Seldon CLI docs](../cli/index.md) for more information.
````

````{tab} Paths

The inference v2 protocol is only aware of models, thus has no concept of pipelines.
Seldon works around this limitation by introducing _virtual_ endpoints for pipelines.
Virtual means that Seldon understands them, but other v2 protocol-compatible components like inference servers do not.

Use the following rules for paths to route to models and pipelines:
* For models, use the path prefix `/v2/models/{model name}`.
  This is normal usage of the inference v2 protocol.
* For pipelines, you can use the path prefix `/v2/pipelines/{pipeline name}`.
  Otherwise calling pipelines looks just like the inference v2 protocol for models.
  Do **not** use any suffix for the pipeline name as you would for routing headers.
* For pipelines, you can also use the path prefix `/v2/models/{pipeline name}.pipeline`.
  Again, this form looks just like the inference v2 protocol for models.
````

`````

Extending our examples from [above](#make-inference-requests), the requests may look like the below when using header-based routing.

`````{tabs}

````{group-tab} Seldon CLI

No changes are required as the `seldon` CLI already understands how to set the appropriate gRPC and REST headers.

````

````{group-tab} cURL

Note the header in the last line:

```{code-block}
:emphasize-lines: 4

curl -v http://0.0.0.0:9000/v2/models/iris/infer \
        -H "Content-Type: application/json" \
        -d '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}' \
        -H "Seldon-Model: iris"
```
````

````{group-tab} grpcurl

Note the `rpc-header` flag in the penultimate line:

```{code-block}
:emphasize-lines: 6

grpcurl \
	-d '{"model_name":"iris","inputs":[{"name":"input","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[1,4]}]}' \
	-plaintext \
	-import-path apis \
	-proto apis/mlops/v2_dataplane/v2_dataplane.proto \
	-rpc-header seldon-model:iris \
	0.0.0.0:9000 inference.GRPCInferenceService/ModelInfer
```
````

````{group-tab} Python tritonclient

Note the `headers` dictionary in the `client.infer()` call:

```{code-block}
:emphasize-lines: 18

import tritonclient.http as httpclient
import numpy as np

client = httpclient.InferenceServerClient(
    url="localhost:8080",
    verbose=False,
)

inputs = [httpclient.InferInput("predict", (1, 4), "FP64")]
inputs[0].set_data_from_numpy(
    np.array([[1, 2, 3, 4]]).astype("float64"),
    binary_data=False,
)

result = client.infer(
    "iris",
    inputs,
    headers={"Seldon-Model": "iris"},
)
print("result is:", result.as_numpy("predict"))
```
````

`````

#### Ingress Routes

If you are using an ingress controller to make inference requests with Seldon, you will need to configure the routing rules correctly.

There are many ways to do this, but custom path prefixes will not work with gRPC.
This is because gRPC determines the path based on the Protobuf definition.
Some gRPC implementations permit manipulating paths when sending requests, but this is by no means universal.

If you want to expose your inference endpoints via gRPC and REST in a consistent way, you should use virtual hosts, subdomains, or headers.

The downside of using only paths is that you cannot differentiate between different installations of Seldon Core v2 or between traffic to Seldon and any other inference endpoints you may have exposed via the same ingress.

You might want to use a mixture of these methods; the choice is yours.

`````{tabs}

````{tab} Virtual Hosts

Virtual hosts are a way of differentiating between logical services accessed via the same physical machine(s).

Virtual hosts are defined by the `Host` header for [HTTP/1](https://www.rfc-editor.org/rfc/rfc7230#section-5.4) and the `:authority` pseudo-header for [HTTP/2](https://www.rfc-editor.org/rfc/rfc9113.html#section-8.3.1).
These represent the same thing, and the HTTP/2 specification defines how to translate these when converting between protocol versions.

Many tools and libraries treat these headers as special and have particular ways of handling them.
Some common ones are given below:

* The `seldon` CLI has an `--authority` flag which applies to both REST and gRPC inference calls.
* `curl` accepts `Host` as a normal header.
* `grpcurl` has an `-authority` flag.
* In Go, the standard library's `http.Request` struct has a `Host` field and ignores attempts to set this value via headers.
* In Python, the `requests` library accepts the host as a normal header.

Be sure to check the documentation for how to set this with your preferred tools and languages.
````

````{tab} Subdomains

Subdomain names constitute a part of the overall host name.
As such, specifying a subdomain name for requests will involve setting the appropriate host in the URI.

For example, you may expose inference services in the namespaces `seldon-1` and `seldon-2` as in the following snippets:

```
curl https://seldon-1.example.com/v2/models/iris/infer ...

seldon model infer --inference-host https://seldon-2.example.com/v2/models/iris/infer ...
```

Many popular ingresses support subdomain-based routing, including Istio and Nginx.
Please refer to the documentation for your ingress of choice for further information.
````

````{tab} Headers

Many ingress controllers and service meshes support routing on headers.
You can use whatever headers you prefer, so long as they do not conflict with any Seldon relies upon.

Many tools and libraries support adding custom headers to requests.
Some common ones are given below:
* The `seldon` CLI accepts headers using the `--header` flag, which can be specified multiple times.
* `curl` accepts headers using the `-H` or `--header` flags.
* `grpcurl` accepts headers using the `-H` flag, which can be specified multiple times.
````

````{tab} Paths

It is possible to route on paths by using well-known path prefixes defined by the inference v2 protocol.
For gRPC, the full path (or "method") for an inference call is:
```
/inference.GRPCInferenceService/ModelInfer
```

This corresponds to the package (`inference`), service (`GRPCInferenceService`), and RPC name (`ModelInfer`) in the Protobuf definition of the inference v2 protocol.

You could use an exact match or a regex like `.*inference.*` to match this path, for example.
````

`````

## Asynchronous Requests

The Seldon architecture uses Kafka and therefore asynchronous requests can be sent by pushing inference v2 protocol payloads to the appropriate topic.
Topics have the following form:

```
seldon.<namespace>.<model|pipeline>.<name>.<inputs|outputs>
```

```{note}
If writing to a pipeline topic, you will need to include a Kafka header with the key `pipeline` and the value being the name of the pipeline.
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
