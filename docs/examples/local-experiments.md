# Seldon V2 Non Kubernetes Local Experiment Examples

### Model Experiment

We will use two SKlearn Iris classification models to illustrate experiments.

```bash
cat ./models/sklearn1.yaml
```

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Model
metadata:
  name: iris
spec:
  storageUri: "gs://seldon-models/mlserver/iris"
  requirements:
  - sklearn

```

```bash
cat ./models/sklearn2.yaml
```

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Model
metadata:
  name: iris2
spec:
  storageUri: "gs://seldon-models/mlserver/iris"
  requirements:
  - sklearn

```

Load both models.

```bash
seldon model load -f ./models/sklearn1.yaml
seldon model load -f ./models/sklearn2.yaml
```

```json
{}
{}

```

Wait for both models to be ready.

```bash
seldon model status iris -w ModelAvailable
seldon model status iris2 -w ModelAvailable
```

```json
{}
{}

```

```bash
seldon model infer iris -i 50 \
  '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}'
```

```
Success: map[:iris_1::50]

```

```bash
seldon model infer iris2 -i 50 \
  '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}'
```

```
Success: map[:iris2_1::50]

```

Create an experiment that modifies the iris model to add a second model splitting traffic 50/50 between the two.

```bash
cat ./experiments/ab-default-model.yaml
```

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Experiment
metadata:
  name: experiment-sample
spec:
  default: iris
  candidates:
  - name: iris
    weight: 50
  - name: iris2
    weight: 50

```

Start the experiment.

```bash
seldon experiment start -f ./experiments/ab-default-model.yaml
```

Wait for the experiment to be ready.

```bash
seldon experiment status experiment-sample -w | jq -M .
```

```json
{
  "experimentName": "experiment-sample",
  "active": true,
  "candidatesReady": true,
  "mirrorReady": true,
  "statusDescription": "experiment active",
  "kubernetesMeta": {}
}

```

Run a set of calls and record which route the traffic took. There should be roughly a 50/50 split.

```bash
seldon model infer iris -i 50 \
  '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}'
```

```
Success: map[:iris2_1::27 :iris_1::23]

```

Show sticky session header `x-seldon-route` that is returned

```bash
seldon model infer iris --show-headers \
  '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}'
```

```
> POST /v2/models/iris/infer HTTP/1.1
> Host: localhost:9000
> Content-Type:[application/json]
> Seldon-Model:[iris]

< X-Seldon-Route:[:iris_1:]
< Ce-Id:[463e96ad-645f-4442-8890-4c340b58820b]
< Traceparent:[00-fe9e87fcbe4be98ed82fb76166e15ceb-d35e7ac96bd8b718-01]
< X-Envoy-Upstream-Service-Time:[3]
< Ce-Specversion:[0.3]
< Date:[Thu, 29 Jun 2023 14:03:03 GMT]
< Ce-Source:[io.seldon.serving.deployment.mlserver]
< Content-Type:[application/json]
< Server:[envoy]
< X-Request-Id:[cieou5ofh5ss73fbjdu0]
< Ce-Endpoint:[iris_1]
< Ce-Modelid:[iris_1]
< Ce-Type:[io.seldon.serving.inference.response]
< Content-Length:[213]
< Ce-Inferenceservicename:[mlserver]
< Ce-Requestid:[463e96ad-645f-4442-8890-4c340b58820b]

{
	"model_name": "iris_1",
	"model_version": "1",
	"id": "463e96ad-645f-4442-8890-4c340b58820b",
	"parameters": {},
	"outputs": [
		{
			"name": "predict",
			"shape": [
				1,
				1
			],
			"datatype": "INT64",
			"parameters": {
				"content_type": "np"
			},
			"data": [
				2
			]
		}
	]
}

```

Use sticky session key passed by last infer request to ensure same route is taken each time.

```bash
seldon model infer iris -s -i 50 \
  '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}'
```

```
Success: map[:iris_1::50]

```

```bash
seldon model infer iris --inference-mode grpc -s -i 50\
   '{"model_name":"iris","inputs":[{"name":"input","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[1,4]}]}'
```

```
Success: map[:iris_1::50]

```

Stop the experiment

```bash
seldon experiment stop experiment-sample
```

Unload both models.

```bash
seldon model unload iris
seldon model unload iris2
```

### Pipeline Experiment

```bash
cat ./models/add10.yaml
```

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Model
metadata:
  name: add10
spec:
  storageUri: "gs://seldon-models/scv2/samples/triton_23-03/add10"
  requirements:
  - triton
  - python

```

```bash
cat ./models/mul10.yaml
```

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Model
metadata:
  name: mul10
spec:
  storageUri: "gs://seldon-models/scv2/samples/triton_23-03/mul10"
  requirements:
  - triton
  - python

```

```bash
seldon model load -f ./models/add10.yaml
seldon model load -f ./models/mul10.yaml
```

```json
{}
{}

```

```bash
seldon model status add10 -w ModelAvailable
seldon model status mul10 -w ModelAvailable
```

```json
{}
{}

```

```bash
cat ./pipelines/mul10.yaml
```

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Pipeline
metadata:
  name: pipeline-mul10
spec:
  steps:
    - name: mul10
  output:
    steps:
    - mul10

```

```bash
cat ./pipelines/add10.yaml
```

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Pipeline
metadata:
  name: pipeline-add10
spec:
  steps:
    - name: add10
  output:
    steps:
    - add10

```

```bash
seldon pipeline load -f ./pipelines/add10.yaml
seldon pipeline load -f ./pipelines/mul10.yaml
```

```bash
seldon pipeline status pipeline-add10 -w PipelineReady
seldon pipeline status pipeline-mul10 -w PipelineReady
```

```json
{"pipelineName":"pipeline-add10", "versions":[{"pipeline":{"name":"pipeline-add10", "uid":"cieov47l80lc739juklg", "version":1, "steps":[{"name":"add10"}], "output":{"steps":["add10.outputs"]}, "kubernetesMeta":{}}, "state":{"pipelineVersion":1, "status":"PipelineReady", "reason":"created pipeline", "lastChangeTimestamp":"2023-06-29T14:05:04.460868091Z", "modelsReady":true}}]}
{"pipelineName":"pipeline-mul10", "versions":[{"pipeline":{"name":"pipeline-mul10", "uid":"cieov47l80lc739jukm0", "version":1, "steps":[{"name":"mul10"}], "output":{"steps":["mul10.outputs"]}, "kubernetesMeta":{}}, "state":{"pipelineVersion":1, "status":"PipelineReady", "reason":"created pipeline", "lastChangeTimestamp":"2023-06-29T14:05:04.631980330Z", "modelsReady":true}}]}

```

```bash
seldon pipeline infer pipeline-add10 --inference-mode grpc \
 '{"model_name":"add10","inputs":[{"name":"INPUT","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]}]}' | jq -M .
```

```json
{
  "outputs": [
    {
      "name": "OUTPUT",
      "datatype": "FP32",
      "shape": [
        "4"
      ],
      "contents": {
        "fp32Contents": [
          11,
          12,
          13,
          14
        ]
      }
    }
  ]
}

```

```bash
seldon pipeline infer pipeline-mul10 --inference-mode grpc \
 '{"model_name":"add10","inputs":[{"name":"INPUT","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]}]}' | jq -M .
```

```json
{
  "outputs": [
    {
      "name": "OUTPUT",
      "datatype": "FP32",
      "shape": [
        "4"
      ],
      "contents": {
        "fp32Contents": [
          10,
          20,
          30,
          40
        ]
      }
    }
  ]
}

```

```bash
cat ./experiments/addmul10.yaml
```

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Experiment
metadata:
  name: addmul10
spec:
  default: pipeline-add10
  resourceType: pipeline
  candidates:
  - name: pipeline-add10
    weight: 50
  - name: pipeline-mul10
    weight: 50

```

```bash
seldon experiment start -f ./experiments/addmul10.yaml
```

```bash
seldon experiment status addmul10 -w | jq -M .
```

```json
{
  "experimentName": "addmul10",
  "active": true,
  "candidatesReady": true,
  "mirrorReady": true,
  "statusDescription": "experiment active",
  "kubernetesMeta": {}
}

```

```bash
seldon pipeline infer pipeline-add10 -i 50 --inference-mode grpc \
 '{"model_name":"add10","inputs":[{"name":"INPUT","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]}]}'
```

```
Success: map[:add10_1::28 :mul10_1::22 :pipeline-add10.pipeline::28 :pipeline-mul10.pipeline::22]

```

Use sticky session key passed by last infer request to ensure same route is taken each time.

```bash
seldon pipeline infer pipeline-add10 --show-headers --inference-mode grpc \
 '{"model_name":"add10","inputs":[{"name":"INPUT","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]}]}'
```

```
> /inference.GRPCInferenceService/ModelInfer HTTP/2
> Host: localhost:9000
> seldon-model:[pipeline-add10.pipeline]

< x-envoy-expected-rq-timeout-ms:[60000]
< x-request-id:[cieov8ofh5ss739277i0]
< date:[Thu, 29 Jun 2023 14:05:23 GMT]
< server:[envoy]
< content-type:[application/grpc]
< x-envoy-upstream-service-time:[6]
< x-seldon-route:[:add10_1: :pipeline-add10.pipeline:]
< x-forwarded-proto:[http]

{"outputs":[{"name":"OUTPUT", "datatype":"FP32", "shape":["4"], "contents":{"fp32Contents":[11, 12, 13, 14]}}]}

```

```bash
seldon pipeline infer pipeline-add10 -s --show-headers --inference-mode grpc \
 '{"model_name":"add10","inputs":[{"name":"INPUT","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]}]}'
```

```
> /inference.GRPCInferenceService/ModelInfer HTTP/2
> Host: localhost:9000
> x-seldon-route:[:add10_1: :pipeline-add10.pipeline:]
> seldon-model:[pipeline-add10.pipeline]

< content-type:[application/grpc]
< x-forwarded-proto:[http]
< x-envoy-expected-rq-timeout-ms:[60000]
< x-seldon-route:[:add10_1: :pipeline-add10.pipeline: :pipeline-add10.pipeline:]
< x-request-id:[cieov90fh5ss739277ig]
< x-envoy-upstream-service-time:[7]
< date:[Thu, 29 Jun 2023 14:05:24 GMT]
< server:[envoy]

{"outputs":[{"name":"OUTPUT", "datatype":"FP32", "shape":["4"], "contents":{"fp32Contents":[11, 12, 13, 14]}}]}

```

```bash
seldon pipeline infer pipeline-add10 -s -i 50 --inference-mode grpc \
 '{"model_name":"add10","inputs":[{"name":"INPUT","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]}]}'
```

```
Success: map[:add10_1::50 :pipeline-add10.pipeline::150]

```

```bash
cat ./models/add20.yaml
```

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Model
metadata:
  name: add20
spec:
  storageUri: "gs://seldon-models/triton/add20"
  requirements:
  - triton
  - python

```

```bash
seldon model load -f ./models/add20.yaml
```

```json
{}

```

```bash
seldon model status add20 -w ModelAvailable
```

```json
{}

```

```bash
cat ./experiments/add1020.yaml
```

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Experiment
metadata:
  name: add1020
spec:
  default: add10
  candidates:
  - name: add10
    weight: 50
  - name: add20
    weight: 50

```

```bash
seldon experiment start -f ./experiments/add1020.yaml
```

```bash
seldon experiment status add1020 -w | jq -M .
```

```json
{
  "experimentName": "add1020",
  "active": true,
  "candidatesReady": true,
  "mirrorReady": true,
  "statusDescription": "experiment active",
  "kubernetesMeta": {}
}

```

```bash
seldon model infer add10 -i 50  --inference-mode grpc \
  '{"model_name":"add10","inputs":[{"name":"INPUT","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]}]}'
```

```
Success: map[:add10_1::22 :add20_1::28]

```

```bash
seldon pipeline infer pipeline-add10 -i 100 --inference-mode grpc \
 '{"model_name":"add10","inputs":[{"name":"INPUT","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]}]}'
```

```
Success: map[:add10_1::24 :add20_1::32 :mul10_1::44 :pipeline-add10.pipeline::56 :pipeline-mul10.pipeline::44]

```

```bash
seldon pipeline infer pipeline-add10 --show-headers --inference-mode grpc \
 '{"model_name":"add10","inputs":[{"name":"INPUT","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]}]}'
```

```
> /inference.GRPCInferenceService/ModelInfer HTTP/2
> Host: localhost:9000
> seldon-model:[pipeline-add10.pipeline]

< x-request-id:[cieovf0fh5ss739279u0]
< x-envoy-upstream-service-time:[5]
< x-seldon-route:[:add10_1: :pipeline-add10.pipeline:]
< date:[Thu, 29 Jun 2023 14:05:48 GMT]
< server:[envoy]
< content-type:[application/grpc]
< x-forwarded-proto:[http]
< x-envoy-expected-rq-timeout-ms:[60000]

{"outputs":[{"name":"OUTPUT", "datatype":"FP32", "shape":["4"], "contents":{"fp32Contents":[11, 12, 13, 14]}}]}

```

```bash
seldon pipeline infer pipeline-add10 -s --show-headers --inference-mode grpc \
 '{"model_name":"add10","inputs":[{"name":"INPUT","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]}]}'
```

```
> /inference.GRPCInferenceService/ModelInfer HTTP/2
> Host: localhost:9000
> x-seldon-route:[:add10_1: :pipeline-add10.pipeline:]
> seldon-model:[pipeline-add10.pipeline]

< x-forwarded-proto:[http]
< x-envoy-expected-rq-timeout-ms:[60000]
< x-request-id:[cieovf8fh5ss739279ug]
< x-envoy-upstream-service-time:[6]
< date:[Thu, 29 Jun 2023 14:05:49 GMT]
< server:[envoy]
< content-type:[application/grpc]
< x-seldon-route:[:add10_1: :pipeline-add10.pipeline: :add20_1: :pipeline-add10.pipeline:]

{"outputs":[{"name":"OUTPUT", "datatype":"FP32", "shape":["4"], "contents":{"fp32Contents":[21, 22, 23, 24]}}]}

```

```bash
seldon experiment stop addmul10
seldon experiment stop add1020
seldon pipeline unload pipeline-add10
seldon pipeline unload pipeline-mul10
seldon model unload add10
seldon model unload add20
seldon model unload mul10
```

### Model Mirror Experiment

We will use two SKlearn Iris classification models to illustrate a model with a mirror.

```bash
cat ./models/sklearn1.yaml
```

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Model
metadata:
  name: iris
spec:
  storageUri: "gs://seldon-models/mlserver/iris"
  requirements:
  - sklearn

```

```bash
cat ./models/sklearn2.yaml
```

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Model
metadata:
  name: iris2
spec:
  storageUri: "gs://seldon-models/mlserver/iris"
  requirements:
  - sklearn

```

Load both models.

```bash
seldon model load -f ./models/sklearn1.yaml
seldon model load -f ./models/sklearn2.yaml
```

```json
{}
{}

```

Wait for both models to be ready.

```bash
seldon model status iris -w ModelAvailable
seldon model status iris2 -w ModelAvailable
```

```json
{}
{}

```

Create an experiment that modifies in which we mirror traffic to iris also to iris2.

```bash
cat ./experiments/sklearn-mirror.yaml
```

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Experiment
metadata:
  name: sklearn-mirror
spec:
  default: iris
  candidates:
  - name: iris
    weight: 100
  mirror:
    name: iris2
    percent: 100

```

Start the experiment.

```bash
seldon experiment start -f ./experiments/sklearn-mirror.yaml
```

Wait for the experiment to be ready.

```bash
seldon experiment status sklearn-mirror -w | jq -M .
```

```json
{
  "experimentName": "sklearn-mirror",
  "active": true,
  "candidatesReady": true,
  "mirrorReady": true,
  "statusDescription": "experiment active",
  "kubernetesMeta": {}
}

```

We get responses from iris but all requests would also have been mirrored to iris2

```bash
seldon model infer iris -i 50 \
  '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}'
```

```
Success: map[:iris_1::50]

```

We can check the local prometheus port from the agent to validate requests went to iris2

```bash
curl -s 0.0.0:9006/metrics | grep seldon_model_infer_total | grep iris2_1
```

```
seldon_model_infer_total{code="200",method_type="rest",model="iris",model_internal="iris2_1",server="mlserver",server_replica="0"} 50

```

Stop the experiment

```bash
seldon experiment stop sklearn-mirror
```

Unload both models.

```bash
seldon model unload iris
seldon model unload iris2
```

## Pipeline Mirror Experiment

```bash
cat ./models/add10.yaml
```

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Model
metadata:
  name: add10
spec:
  storageUri: "gs://seldon-models/scv2/samples/triton_23-03/add10"
  requirements:
  - triton
  - python

```

```bash
cat ./models/mul10.yaml
```

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Model
metadata:
  name: mul10
spec:
  storageUri: "gs://seldon-models/scv2/samples/triton_23-03/mul10"
  requirements:
  - triton
  - python

```

```bash
seldon model load -f ./models/add10.yaml
seldon model load -f ./models/mul10.yaml
```

```json
{}
{}

```

```bash
seldon model status add10 -w ModelAvailable
seldon model status mul10 -w ModelAvailable
```

```json
{}
{}

```

```bash
cat ./pipelines/mul10.yaml
```

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Pipeline
metadata:
  name: pipeline-mul10
spec:
  steps:
    - name: mul10
  output:
    steps:
    - mul10

```

```bash
cat ./pipelines/add10.yaml
```

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Pipeline
metadata:
  name: pipeline-add10
spec:
  steps:
    - name: add10
  output:
    steps:
    - add10

```

```bash
seldon pipeline load -f ./pipelines/add10.yaml
seldon pipeline load -f ./pipelines/mul10.yaml
```

```bash
seldon pipeline status pipeline-add10 -w PipelineReady
seldon pipeline status pipeline-mul10 -w PipelineReady
```

```json
{"pipelineName":"pipeline-add10", "versions":[{"pipeline":{"name":"pipeline-add10", "uid":"ciep072i8ufs73flaipg", "version":1, "steps":[{"name":"add10"}], "output":{"steps":["add10.outputs"]}, "kubernetesMeta":{}}, "state":{"pipelineVersion":1, "status":"PipelineReady", "reason":"created pipeline", "lastChangeTimestamp":"2023-06-29T14:07:24.903503109Z", "modelsReady":true}}]}
{"pipelineName":"pipeline-mul10", "versions":[{"pipeline":{"name":"pipeline-mul10", "uid":"ciep072i8ufs73flaiq0", "version":1, "steps":[{"name":"mul10"}], "output":{"steps":["mul10.outputs"]}, "kubernetesMeta":{}}, "state":{"pipelineVersion":1, "status":"PipelineReady", "reason":"created pipeline", "lastChangeTimestamp":"2023-06-29T14:07:25.082642153Z", "modelsReady":true}}]}

```

```bash
seldon pipeline infer pipeline-add10 --inference-mode grpc \
 '{"model_name":"add10","inputs":[{"name":"INPUT","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]}]}'
```

```json
{"outputs":[{"name":"OUTPUT", "datatype":"FP32", "shape":["4"], "contents":{"fp32Contents":[11, 12, 13, 14]}}]}

```

```bash
seldon pipeline infer pipeline-mul10 --inference-mode grpc \
 '{"model_name":"add10","inputs":[{"name":"INPUT","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]}]}'
```

```json
{"outputs":[{"name":"OUTPUT", "datatype":"FP32", "shape":["4"], "contents":{"fp32Contents":[10, 20, 30, 40]}}]}

```

```bash
cat ./experiments/addmul10-mirror.yaml
```

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Experiment
metadata:
  name: addmul10-mirror
spec:
  default: pipeline-add10
  resourceType: pipeline
  candidates:
  - name: pipeline-add10
    weight: 100
  mirror:
    name: pipeline-mul10
    percent: 100

```

```bash
seldon experiment start -f ./experiments/addmul10-mirror.yaml
```

```bash
seldon experiment status addmul10-mirror -w | jq -M .
```

```json
{
  "experimentName": "addmul10-mirror",
  "active": true,
  "candidatesReady": true,
  "mirrorReady": true,
  "statusDescription": "experiment active",
  "kubernetesMeta": {}
}

```

```bash
seldon pipeline infer pipeline-add10 -i 1 --inference-mode grpc \
 '{"model_name":"add10","inputs":[{"name":"INPUT","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]}]}'
```

```json
{"outputs":[{"name":"OUTPUT", "datatype":"FP32", "shape":["4"], "contents":{"fp32Contents":[11, 12, 13, 14]}}]}

```

Let's check that the mul10 model was called.

```bash
curl -s 0.0.0:9007/metrics | grep seldon_model_infer_total | grep mul10_1
```

```
seldon_model_infer_total{code="OK",method_type="grpc",model="mul10",model_internal="mul10_1",server="triton",server_replica="0"} 2

```

```bash
curl -s 0.0.0:9007/metrics | grep seldon_model_infer_total | grep add10_1
```

```
seldon_model_infer_total{code="OK",method_type="grpc",model="add10",model_internal="add10_1",server="triton",server_replica="0"} 2

```

Let's do an http call and check agaib the two models

```bash
seldon pipeline infer pipeline-add10 -i 1 \
 '{"model_name":"add10","inputs":[{"name":"INPUT","data":[1,2,3,4],"datatype":"FP32","shape":[4]}]}'
```

```json
{
	"model_name": "",
	"outputs": [
		{
			"data": [
				11,
				12,
				13,
				14
			],
			"name": "OUTPUT",
			"shape": [
				4
			],
			"datatype": "FP32"
		}
	]
}

```

```bash
curl -s 0.0.0:9007/metrics | grep seldon_model_infer_total | grep mul10_1
```

```
seldon_model_infer_total{code="OK",method_type="grpc",model="mul10",model_internal="mul10_1",server="triton",server_replica="0"} 3

```

```bash
curl -s 0.0.0:9007/metrics | grep seldon_model_infer_total | grep add10_1
```

```
seldon_model_infer_total{code="OK",method_type="grpc",model="add10",model_internal="add10_1",server="triton",server_replica="0"} 3

```

```bash
seldon pipeline inspect pipeline-mul10
```

```
seldon.default.model.mul10.inputs	ciep0bofh5ss73dpdiq0	{"inputs":[{"name":"INPUT", "datatype":"FP32", "shape":["4"], "contents":{"fp32Contents":[1, 2, 3, 4]}}]}
seldon.default.model.mul10.outputs	ciep0bofh5ss73dpdiq0	{"modelName":"mul10_1", "modelVersion":"1", "outputs":[{"name":"OUTPUT", "datatype":"FP32", "shape":["4"], "contents":{"fp32Contents":[10, 20, 30, 40]}}]}
seldon.default.pipeline.pipeline-mul10.inputs	ciep0bofh5ss73dpdiq0	{"inputs":[{"name":"INPUT", "datatype":"FP32", "shape":["4"], "contents":{"fp32Contents":[1, 2, 3, 4]}}]}
seldon.default.pipeline.pipeline-mul10.outputs	ciep0bofh5ss73dpdiq0	{"outputs":[{"name":"OUTPUT", "datatype":"FP32", "shape":["4"], "contents":{"fp32Contents":[10, 20, 30, 40]}}]}

```

```bash
seldon experiment stop addmul10-mirror
seldon pipeline unload pipeline-add10
seldon pipeline unload pipeline-mul10
seldon model unload add10
seldon model unload mul10
```

```python

```
