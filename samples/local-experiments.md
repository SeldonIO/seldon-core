## Seldon V2 Non Kubernetes Local Experiment Examples

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

```yaml
Success: map[:iris_1::50]

```

```bash
seldon model infer iris2 -i 50 \
  '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}'
```

```yaml
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

```json
{}

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

```yaml
Success: map[:iris2_1::30 :iris_1::20]

```

Show sticky session header `x-seldon-route` that is returned

```bash
seldon model infer iris --show-headers \
  '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}'
```

```yaml
> POST /v2/models/iris/infer HTTP/1.1
> Host: 0.0.0.0:9000
> Content-Type:[application/json]
> Seldon-Model:[iris]

< Ce-Requestid:[3568c090-f916-44db-95b6-cbcc2b2e4eda]
< Ce-Type:[io.seldon.serving.inference.response]
< Content-Type:[application/json]
< X-Request-Id:[cdnmtbv2c1bs73a62cb0]
< X-Seldon-Route:[:iris2_1:]
< Ce-Id:[3568c090-f916-44db-95b6-cbcc2b2e4eda]
< Ce-Modelid:[iris2_1]
< Ce-Specversion:[0.3]
< Server:[envoy]
< Ce-Endpoint:[iris2_1]
< Ce-Source:[io.seldon.serving.deployment.mlserver]
< Date:[Sat, 12 Nov 2022 10:00:15 GMT]
< Traceparent:[00-f927e47b237c00f371dcc88f3c0ec2ac-ccc466a3ec1cf492-01]
< Ce-Inferenceservicename:[mlserver]
< Content-Length:[229]
< X-Envoy-Upstream-Service-Time:[2]

{
	"model_name": "iris2_1",
	"model_version": "1",
	"id": "3568c090-f916-44db-95b6-cbcc2b2e4eda",
	"parameters": {
		"content_type": null,
		"headers": null
	},
	"outputs": [
		{
			"name": "predict",
			"shape": [
				1
			],
			"datatype": "INT64",
			"parameters": null,
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

```yaml
Success: map[:iris2_1::50]

```

```bash
seldon model infer iris --inference-mode grpc -s -i 50\
   '{"model_name":"iris","inputs":[{"name":"input","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[1,4]}]}'
```

```yaml
Success: map[:iris2_1::50]

```

Stop the experiment

```bash
seldon experiment stop experiment-sample
```

```json
{}

```

Unload both models.

```bash
seldon model unload iris
seldon model unload iris2
```

```json
{}
{}

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
  storageUri: "gs://seldon-models/triton/add10"
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
  storageUri: "gs://seldon-models/triton/mul10"
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

```json
{}
{}

```

```bash
seldon pipeline status pipeline-add10 -w PipelineReady
seldon pipeline status pipeline-mul10 -w PipelineReady
```

```json
{"pipelineName":"pipeline-add10", "versions":[{"pipeline":{"name":"pipeline-add10", "uid":"cdnmtgv7uvcc73er9lu0", "version":1, "steps":[{"name":"add10"}], "output":{"steps":["add10.outputs"]}, "kubernetesMeta":{}}, "state":{"pipelineVersion":1, "status":"PipelineReady", "reason":"created pipeline", "lastChangeTimestamp":"2022-11-12T10:00:35.844121227Z"}}]}
{"pipelineName":"pipeline-mul10", "versions":[{"pipeline":{"name":"pipeline-mul10", "uid":"cdnmtgv7uvcc73er9lug", "version":1, "steps":[{"name":"mul10"}], "output":{"steps":["mul10.outputs"]}, "kubernetesMeta":{}}, "state":{"pipelineVersion":1, "status":"PipelineReady", "reason":"created pipeline", "lastChangeTimestamp":"2022-11-12T10:00:36.037598129Z"}}]}

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

```json
{}

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

```yaml
Success: map[:add10_1::34 :mul10_1::16 :pipeline-add10.pipeline::34 :pipeline-mul10.pipeline::16]

```

Use sticky session key passed by last infer request to ensure same route is taken each time.

```bash
seldon pipeline infer pipeline-add10 --show-headers --inference-mode grpc \
 '{"model_name":"add10","inputs":[{"name":"INPUT","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]}]}'
```

```yaml
> /inference.GRPCInferenceService/ModelInfer HTTP/2
> Host: 0.0.0.0:9000
> seldon-model:[pipeline-add10.pipeline]

< server:[envoy]
< content-type:[application/grpc]
< x-seldon-route:[:add10_1: :pipeline-add10.pipeline:]
< x-envoy-expected-rq-timeout-ms:[60000]
< x-forwarded-proto:[http]
< x-request-id:[cdnmtl95h2ks73fq2810]
< x-envoy-upstream-service-time:[9]
< date:[Sat, 12 Nov 2022 10:00:53 GMT]

{"outputs":[{"name":"OUTPUT", "datatype":"FP32", "shape":["4"], "contents":{"fp32Contents":[11, 12, 13, 14]}}]}

```

```bash
seldon pipeline infer pipeline-add10 -s --show-headers --inference-mode grpc \
 '{"model_name":"add10","inputs":[{"name":"INPUT","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]}]}'
```

```yaml
> /inference.GRPCInferenceService/ModelInfer HTTP/2
> Host: 0.0.0.0:9000
> x-seldon-route:[:add10_1: :pipeline-add10.pipeline:]
> seldon-model:[pipeline-add10.pipeline]

< x-forwarded-proto:[http]
< x-envoy-expected-rq-timeout-ms:[60000]
< x-request-id:[cdnmtmp5h2ks73fq281g]
< x-envoy-upstream-service-time:[9]
< date:[Sat, 12 Nov 2022 10:00:59 GMT]
< server:[envoy]
< content-type:[application/grpc]
< x-seldon-route:[:add10_1: :pipeline-add10.pipeline: :pipeline-add10.pipeline:]

{"outputs":[{"name":"OUTPUT", "datatype":"FP32", "shape":["4"], "contents":{"fp32Contents":[11, 12, 13, 14]}}]}

```

```bash
seldon pipeline infer pipeline-add10 -s -i 50 --inference-mode grpc \
 '{"model_name":"add10","inputs":[{"name":"INPUT","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]}]}'
```

```yaml
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

```json
{}

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

```yaml
Success: map[:add10_1::20 :add20_1::30]

```

```bash
seldon pipeline infer pipeline-add10 -i 100 --inference-mode grpc \
 '{"model_name":"add10","inputs":[{"name":"INPUT","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]}]}'
```

```yaml
Success: map[:add10_1::25 :add20_1::26 :mul10_1::49 :pipeline-add10.pipeline::51 :pipeline-mul10.pipeline::49]

```

```bash
seldon pipeline infer pipeline-add10 --show-headers --inference-mode grpc \
 '{"model_name":"add10","inputs":[{"name":"INPUT","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]}]}'
```

```yaml
> /inference.GRPCInferenceService/ModelInfer HTTP/2
> Host: 0.0.0.0:9000
> seldon-model:[pipeline-add10.pipeline]

< x-envoy-expected-rq-timeout-ms:[60000]
< x-seldon-route:[:add20_1: :pipeline-add10.pipeline:]
< x-request-id:[cdnmtqp5h2ks73fq2ad0]
< x-envoy-upstream-service-time:[6]
< date:[Sat, 12 Nov 2022 10:01:15 GMT]
< server:[envoy]
< content-type:[application/grpc]
< x-forwarded-proto:[http]

{"outputs":[{"name":"OUTPUT", "datatype":"FP32", "shape":["4"], "contents":{"fp32Contents":[21, 22, 23, 24]}}]}

```

```bash
seldon pipeline infer pipeline-add10 -s --show-headers --inference-mode grpc \
 '{"model_name":"add10","inputs":[{"name":"INPUT","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]}]}'
```

```yaml
> /inference.GRPCInferenceService/ModelInfer HTTP/2
> Host: 0.0.0.0:9000
> x-seldon-route:[:add20_1: :pipeline-add10.pipeline:]
> seldon-model:[pipeline-add10.pipeline]

< x-forwarded-proto:[http]
< x-envoy-upstream-service-time:[6]
< x-request-id:[cdnmtrh5h2ks73fq2adg]
< date:[Sat, 12 Nov 2022 10:01:18 GMT]
< server:[envoy]
< content-type:[application/grpc]
< x-envoy-expected-rq-timeout-ms:[60000]
< x-seldon-route:[:add20_1: :pipeline-add10.pipeline: :add10_1: :pipeline-add10.pipeline:]

{"outputs":[{"name":"OUTPUT", "datatype":"FP32", "shape":["4"], "contents":{"fp32Contents":[11, 12, 13, 14]}}]}

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

```json
{}
{}
{}
{}
{}
{}
{}

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

```json
{}

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

```yaml
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

```json
{}

```

Unload both models.

```bash
seldon model unload iris
seldon model unload iris2
```

```json
{}
{}

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
  storageUri: "gs://seldon-models/triton/add10"
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
  storageUri: "gs://seldon-models/triton/mul10"
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

```json
{}
{}

```

```bash
seldon pipeline status pipeline-add10 -w PipelineReady
seldon pipeline status pipeline-mul10 -w PipelineReady
```

```json
{"pipelineName":"pipeline-add10","versions":[{"pipeline":{"name":"pipeline-add10","uid":"ce35uho4mlvc73d8elh0","version":1,"steps":[{"name":"add10"}],"output":{"steps":["add10.outputs"]},"kubernetesMeta":{}},"state":{"pipelineVersion":1,"status":"PipelineReady","reason":"created pipeline","lastChangeTimestamp":"2022-11-29T19:36:39.645836943Z","modelsReady":true}}]}
{"pipelineName":"pipeline-mul10","versions":[{"pipeline":{"name":"pipeline-mul10","uid":"ce35uho4mlvc73d8elhg","version":1,"steps":[{"name":"mul10"}],"output":{"steps":["mul10.outputs"]},"kubernetesMeta":{}},"state":{"pipelineVersion":1,"status":"PipelineReady","reason":"created pipeline","lastChangeTimestamp":"2022-11-29T19:36:39.824124774Z","modelsReady":true}}]}

```

```bash
seldon pipeline infer pipeline-add10 --inference-mode grpc \
 '{"model_name":"add10","inputs":[{"name":"INPUT","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]}]}'
```

```json
{"outputs":[{"name":"OUTPUT","datatype":"FP32","shape":["4"],"contents":{"fp32Contents":[11,12,13,14]}}]}

```

```bash
seldon pipeline infer pipeline-mul10 --inference-mode grpc \
 '{"model_name":"add10","inputs":[{"name":"INPUT","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]}]}'
```

```json
{"outputs":[{"name":"OUTPUT","datatype":"FP32","shape":["4"],"contents":{"fp32Contents":[10,20,30,40]}}]}

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

```json
{}

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
{"outputs":[{"name":"OUTPUT","datatype":"FP32","shape":["4"],"contents":{"fp32Contents":[11,12,13,14]}}]}

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

```yaml
seldon.default.model.mul10.inputs	ce35uksj0jvc73bqr0pg	{"inputs":[{"name":"INPUT","datatype":"FP32","shape":["4"],"contents":{"fp32Contents":[1,2,3,4]}}]}
seldon.default.model.mul10.outputs	ce35uksj0jvc73bqr0pg	{"modelName":"mul10_1","modelVersion":"1","outputs":[{"name":"OUTPUT","datatype":"FP32","shape":["4"],"contents":{"fp32Contents":[10,20,30,40]}}]}
seldon.default.pipeline.pipeline-mul10.inputs	ce35uksj0jvc73bqr0pg	{"inputs":[{"name":"INPUT","datatype":"FP32","shape":["4"],"contents":{"fp32Contents":[1,2,3,4]}}]}
seldon.default.pipeline.pipeline-mul10.outputs	ce35uksj0jvc73bqr0pg	{"outputs":[{"name":"OUTPUT","datatype":"FP32","shape":["4"],"contents":{"fp32Contents":[10,20,30,40]}}]}

```

```bash
seldon experiment stop addmul10-mirror
seldon pipeline unload pipeline-add10
seldon pipeline unload pipeline-mul10
seldon model unload add10
seldon model unload mul10
```

```json
{}
{}
{}
{}
{}

```

```python

```
