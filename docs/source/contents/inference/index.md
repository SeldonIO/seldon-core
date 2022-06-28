# Inference

This section will discuss how to make inference calls against your seldon models, pipelines or explainers.

You can make synchronous inference requests via REST or gRPC or asynchronous requests via Kafka topics.

## Synchronous requests

  1. Find the Seldon service endpoint
  2. Make requests via REST or gRPC

### Find Seldon service endpoint

 1. If you are runing Seldon locally via Docker compose by deafult the endpoint will be `0.0.0.0:9000`
 2. If you are running in Kubernetes Seldon creates a single Service `seldon-mesh` in the namespace Seldon is installed to, usually `seldon-mesh`. If this has be exposed via a load balancer this can be found via:
 
 ```bash
 kubectl get svc seldon-mesh -n seldon-mesh -o jsonpath='{.status.loadBalancer.ingress[0].ip}'
 ```

### Make inference requests

Seldon routes requests to to the correct enpoint via headers in HTTP calls. You should set the header `seldon-model` as follow:

 * For Models: use the model name, e.g. for a Model names `mymodel`, `seldon-model: mymodel`
 * For Pipelines: use the Pipeline name with the suffix `.pipeline`, e.g. for a Pipeline named `mypipeline`, `seldon-model: mypipeline.pipeline`
 * For Explainers: use the Explainer name with the suffix `.explainer`, e.g. for an Explainer named `myexplainer`, `seldon-m odel: myexplainer.explainer`

The content of your request should be a [V2 protocol payload](../apis/inference/v2.md).

The `seldon` CLI can be used to easily send requests to your deployed resources. See the [examples](../examples/index) and the [seldon CLI docs](../cli/index.md).

An example curl request might look like for a Model called `iris`:

```
curl -v http://0.0.0.0:9000/v2/models/iris/infer -H "Content-Type: application/json" -H "seldon-model: iris"\
        -d '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}'
```

A similar gRPC request using grpcurl to the same model might look like:

```
grpcurl -d '{"model_name":"iris","inputs":[{"name":"input","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[1,4]}]}' \
        -plaintext \
	-import-path apis \	
	-proto apis/mlops/v2_dataplane/v2_dataplane.proto \
	-rpc-header seldon-model:iris \
	0.0.0.0:9000 inference.GRPCInferenceService/ModelInfer
```

The above request was run from the project root folder allowing reference to the protos defined in the apis folder.

For Pipelines a synchronous request is possible if the Pipeline has an outputs section in the spec.

## Asynchronous requests

The Seldon architetcure uses Kafka and therefore asynchronous requests can be sent by pushing V2 proto payloads to the appropriate topic.

```
seldon.<namespace>.<model|pipeline|explainer>.<name>.<inputs|outputs>
```

### Model Inference

For a local install if you have a Model `iris`, you would be able to send a prediction request by pushing to the topic: `seldon.default.model.iris.inputs`. The response will appear on `seldon.default.model.iris.outputs`.

For a Kubernetes install in `seldon-mesh` if you have a Model `iris`, you would be able to send a prediction request by pushing to the topic: `seldon.seldon-mesh.model.iris.inputs`. The response will appear on `seldon.seldon-mesh.model.iris.outputs`.


### Pipeline Inference

For a local install if you have a Pipeline `mypipeline`, you would be able to send a prediction request by pushing to the topic: `seldon.default.pipeline.mypipeline.inputs`. The response will appear on `seldon.default.pipeline.mypipeline.outputs`.

For a Kubernetes install in `seldon-mesh` if you have a Pipeline `pipeline`, you would be able to send a prediction request by pushing to the topic: `seldon.seldon-mesh.pipeline.mypipeline.inputs`. The response will appear on `seldon.seldon-mesh.pipeline.mypipeline.outputs`.


## Pipeline Metadata

It may be useful to send metadata alongside your inference. If using Kafka directly as described above you can attach kafka metadata to your request which will be passed oround the graph. When making synchronous requests to your pipeline with REST or gRPC you can also do this.

 * For REST requests add HTTP headers prefixe with `X-`
 * For gRPC requests add meatdata with keys starting with `X-`

You can also do this with the seldon CLI by setting headers with the `--header` argument (and also showing response headers with the `--show-headers` argument)

```
seldon pipeline infer --show-headers --header X-foo=bar tfsimples \
    '{"inputs":[{"name":"INPUT0","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,16]},{"name":"INPUT1","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,16]}]}'
```

## Request IDs

For both model and pipeline request the response will contain a `x-request-id` response header. For pipeline requests this can be used to inspect the pipeline steps via the CLI, e.g.:

```
seldon pipeline inspect tfsimples --request-id carjjolvqj3j2pfbut10 --offset 10
```

The `--offset` parameter specifies how many messages (from the latest) you want to search to find your request. If not specified the last request will be shown.

`x-request-id` will also appear in tracing spans.
