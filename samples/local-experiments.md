## Seldon V2 Non Kubernetes Local Experiment Examples


### Model Experiment

We will use two SKlearn Iris classification models to illustrate experiments.


```bash
cat ./models/sklearn1.yaml
```
````{collapse} Expand to see output
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
````

```bash
cat ./models/sklearn2.yaml
```
````{collapse} Expand to see output
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
````
Load both models.


```bash
seldon model load -f ./models/sklearn1.yaml
seldon model load -f ./models/sklearn2.yaml
```
````{collapse} Expand to see output
```json

    {}
    {}
```
````
Wait for both models to be ready.


```bash
seldon model status iris -w ModelAvailable
seldon model status iris2 -w ModelAvailable
```
````{collapse} Expand to see output
```json

    {}
    {}
```
````
Create an experiment that modifies the iris model to add a second model splitting traffic 50/50 between the two.


```bash
cat ./experiments/ab-default-model.yaml 
```
````{collapse} Expand to see output
```yaml
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Experiment
    metadata:
      name: experiment-sample
    spec:
      default: iris
      candidates:
      - modelName: iris
        weight: 50
      - modelName: iris2
        weight: 50
```
````
Start the experiment.


```bash
seldon experiment start -f ./experiments/ab-default-model.yaml 
```
````{collapse} Expand to see output
```json

    {}
```
````
Wait for the experiment to be ready.


```bash
seldon experiment status experiment-sample -w | jq -M .
```
````{collapse} Expand to see output
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
````
Run a set of calls and record which route the traffic took. There should be roughly a 50/50 split.


```bash
seldon model infer iris -i 50 \
  '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}' 
```
````{collapse} Expand to see output
```json

    map[:iris2_1::24 :iris_1::26]
```
````
Show sticky session header `x-seldon-route` that is returned


```bash
seldon model infer iris --show-headers \
  '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}' 
```
````{collapse} Expand to see output
```json

    Request header Content-Type:[application/json]
    Request header Seldon-Model:[iris]
    Response header Ce-Specversion:[0.3]
    Response header Server:[envoy]
    Response header Traceparent:[00-d4bddd5e505e5cd7322344d826898496-acbc6dd4ed24dfde-01]
    Response header X-Seldon-Route:[:iris_1:]
    Response header Ce-Id:[e621cc3b-c488-46f8-bb5a-4fde4b02dcb9]
    Response header Ce-Source:[io.seldon.serving.deployment.mlserver]
    Response header Content-Length:[228]
    Response header Content-Type:[application/json]
    Response header Date:[Fri, 05 Aug 2022 06:14:39 GMT]
    Response header X-Envoy-Upstream-Service-Time:[3]
    Response header X-Request-Id:[573e72c5-a339-4ff7-b90a-cd6319a42f48]
    Response header Ce-Modelid:[iris_1]
    Response header Ce-Requestid:[e621cc3b-c488-46f8-bb5a-4fde4b02dcb9]
    Response header Ce-Endpoint:[iris_1]
    Response header Ce-Type:[io.seldon.serving.inference.response]
    Response header Ce-Inferenceservicename:[mlserver]
    {
    	"model_name": "iris_1",
    	"model_version": "1",
    	"id": "e621cc3b-c488-46f8-bb5a-4fde4b02dcb9",
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
````
Use sticky session key passed by last infer request to ensure same route is taken each time.


```bash
seldon model infer iris -s -i 50 \
  '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}' 
```
````{collapse} Expand to see output
```json

    map[:iris_1::50]
```
````

```bash
seldon model infer iris --inference-mode grpc -s -i 50\
   '{"model_name":"iris","inputs":[{"name":"input","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[1,4]}]}' 
```
````{collapse} Expand to see output
```json

    map[:iris_1::50]
```
````
Stop the experiment


```bash
seldon experiment stop experiment-sample
```
````{collapse} Expand to see output
```json

    {}
```
````
Unload both models.


```bash
seldon model unload iris
seldon model unload iris2
```
````{collapse} Expand to see output
```json

    {}
    {}
```
````
### Pipeline Experiment


```bash
cat ./models/add10.yaml
```
````{collapse} Expand to see output
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
````

```bash
cat ./models/mul10.yaml
```
````{collapse} Expand to see output
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
````

```bash
seldon model load -f ./models/add10.yaml
seldon model load -f ./models/mul10.yaml
```
````{collapse} Expand to see output
```json

    {}
    {}
```
````

```bash
seldon model status add10 -w ModelAvailable
seldon model status mul10 -w ModelAvailable
```
````{collapse} Expand to see output
```json

    {}
    {}
```
````

```bash
cat ./pipelines/mul10.yaml
```
````{collapse} Expand to see output
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
````

```bash
cat ./pipelines/add10.yaml
```
````{collapse} Expand to see output
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
````

```bash
seldon pipeline load -f ./pipelines/add10.yaml
seldon pipeline load -f ./pipelines/mul10.yaml
```
````{collapse} Expand to see output
```json

    {}
    {}
```
````

```bash
seldon pipeline status pipeline-add10 -w PipelineReady 
seldon pipeline status pipeline-mul10 -w PipelineReady 
```
````{collapse} Expand to see output
```json

    {"pipelineName":"pipeline-add10","versions":[{"pipeline":{"name":"pipeline-add10","uid":"cbmbaq2l0p8os8jr7do0","version":1,"steps":[{"name":"add10"}],"output":{"steps":["add10.outputs"]},"kubernetesMeta":{}},"state":{"pipelineVersion":1,"status":"PipelineReady","reason":"Created pipeline","lastChangeTimestamp":"2022-08-05T06:15:04.824730014Z"}}]}
    {"pipelineName":"pipeline-mul10","versions":[{"pipeline":{"name":"pipeline-mul10","uid":"cbmbaq2l0p8os8jr7dog","version":1,"steps":[{"name":"mul10"}],"output":{"steps":["mul10.outputs"]},"kubernetesMeta":{}},"state":{"pipelineVersion":1,"status":"PipelineReady","reason":"Created pipeline","lastChangeTimestamp":"2022-08-05T06:15:05.334401790Z"}}]}
```
````

```bash
seldon pipeline infer pipeline-add10 --inference-mode grpc \
 '{"model_name":"add10","inputs":[{"name":"INPUT","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]}]}' 
```
````{collapse} Expand to see output
```json

    {"outputs":[{"name":"OUTPUT","datatype":"FP32","shape":["4"],"contents":{"fp32Contents":[11,12,13,14]}}],"rawOutputContents":["AAAwQQAAQEEAAFBBAABgQQ=="]}
```
````

```bash
seldon pipeline infer pipeline-mul10 --inference-mode grpc \
 '{"model_name":"add10","inputs":[{"name":"INPUT","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]}]}' 
```
````{collapse} Expand to see output
```json

    {"outputs":[{"name":"OUTPUT","datatype":"FP32","shape":["4"],"contents":{"fp32Contents":[10,20,30,40]}}],"rawOutputContents":["AAAgQQAAoEEAAPBBAAAgQg=="]}
```
````

```bash
cat ./experiments/addmul10.yaml
```
````{collapse} Expand to see output
```yaml
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Experiment
    metadata:
      name: addmul10
    spec:
      default: pipeline-add10
      resourceType: pipeline
      candidates:
      - modelName: pipeline-add10
        weight: 50
      - modelName: pipeline-mul10
        weight: 50
```
````

```bash
seldon experiment start -f ./experiments/addmul10.yaml 
```
````{collapse} Expand to see output
```json

    {}
```
````

```bash
seldon experiment status addmul10 -w | jq -M .
```
````{collapse} Expand to see output
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
````

```bash
seldon pipeline infer pipeline-add10 -i 50 --inference-mode grpc \
 '{"model_name":"add10","inputs":[{"name":"INPUT","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]}]}' 
```
````{collapse} Expand to see output
```json

    map[:add10_1::28 :mul10_1::22 :pipeline-add10.pipeline::28 :pipeline-mul10.pipeline::22]
```
````
Use sticky session key passed by last infer request to ensure same route is taken each time.


```bash
seldon pipeline infer pipeline-add10 --show-headers --inference-mode grpc \
 '{"model_name":"add10","inputs":[{"name":"INPUT","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]}]}' 
```
````{collapse} Expand to see output
```json

    Request metadata seldon-model:[pipeline-add10.pipeline]
    {"outputs":[{"name":"OUTPUT","datatype":"FP32","shape":["4"],"contents":{"fp32Contents":[10,20,30,40]}}],"rawOutputContents":["AAAgQQAAoEEAAPBBAAAgQg=="]}
    Response header date:[Fri, 05 Aug 2022 06:15:57 GMT]
    Response header server:[envoy]
    Response header content-type:[application/grpc]
    Response header x-forwarded-proto:[http]
    Response header x-request-id:[5bcf4c9e-3f74-4edb-83d4-7e804be6471e]
    Response header x-envoy-expected-rq-timeout-ms:[60000]
    Response header x-seldon-route:[:mul10_1: :pipeline-mul10.pipeline:]
    Response header x-envoy-upstream-service-time:[10]
```
````

```bash
seldon pipeline infer pipeline-add10 -s --show-headers --inference-mode grpc \
 '{"model_name":"add10","inputs":[{"name":"INPUT","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]}]}' 
```
````{collapse} Expand to see output
```json

    Request metadata x-seldon-route:[:mul10_1: :pipeline-mul10.pipeline:]
    {"outputs":[{"name":"OUTPUT","datatype":"FP32","shape":["4"],"contents":{"fp32Contents":[10,20,30,40]}}],"rawOutputContents":["AAAgQQAAoEEAAPBBAAAgQg=="]}
    Response header x-seldon-route:[:mul10_1: :pipeline-mul10.pipeline:]
    Response header x-forwarded-proto:[http]
    Response header x-request-id:[5978548e-0ce4-442a-a567-af0d186d116c]
    Response header x-envoy-expected-rq-timeout-ms:[60000]
    Response header x-envoy-upstream-service-time:[12]
    Response header date:[Fri, 05 Aug 2022 06:16:03 GMT]
    Response header server:[envoy]
    Response header content-type:[application/grpc]
```
````

```bash
seldon pipeline infer pipeline-add10 -s -i 50 --inference-mode grpc \
 '{"model_name":"add10","inputs":[{"name":"INPUT","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]}]}' 
```
````{collapse} Expand to see output
```json

    map[:mul10_1::50 :pipeline-mul10.pipeline::50]
```
````

```bash
cat ./models/add20.yaml
```
````{collapse} Expand to see output
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
````

```bash
seldon model load -f ./models/add20.yaml
```
````{collapse} Expand to see output
```json

    {}
```
````

```bash
seldon model status add20 -w ModelAvailable
```
````{collapse} Expand to see output
```json

    {}
```
````

```bash
cat ./experiments/add1020.yaml
```
````{collapse} Expand to see output
```yaml
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Experiment
    metadata:
      name: add1020
    spec:
      default: add10
      candidates:
      - modelName: add10
        weight: 50
      - modelName: add20
        weight: 50
```
````

```bash
seldon experiment start -f ./experiments/add1020.yaml
```
````{collapse} Expand to see output
```json

    {}
```
````

```bash
seldon experiment status add1020 -w | jq -M .
```
````{collapse} Expand to see output
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
````

```bash
seldon model infer add10 -i 50  --inference-mode grpc \
  '{"model_name":"add10","inputs":[{"name":"INPUT","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]}]}' 
```
````{collapse} Expand to see output
```json

    map[:add10_1::24 :add20_1::26]
```
````

```bash
seldon pipeline infer pipeline-add10 -i 100 --inference-mode grpc \
 '{"model_name":"add10","inputs":[{"name":"INPUT","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]}]}' 
```
````{collapse} Expand to see output
```json

    map[:add10_1::31 :add20_1::27 :mul10_1::42 :pipeline-add10.pipeline::58 :pipeline-mul10.pipeline::42]
```
````

```bash
seldon pipeline infer pipeline-add10 --show-headers --inference-mode grpc \
 '{"model_name":"add10","inputs":[{"name":"INPUT","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]}]}' 
```
````{collapse} Expand to see output
```json

    Request metadata seldon-model:[pipeline-add10.pipeline]
    {"outputs":[{"name":"OUTPUT","datatype":"FP32","shape":["4"],"contents":{"fp32Contents":[10,20,30,40]}}],"rawOutputContents":["AAAgQQAAoEEAAPBBAAAgQg=="]}
    Response header x-forwarded-proto:[http]
    Response header x-request-id:[84fede6e-e54c-4063-b23a-168f312b4d95]
    Response header x-envoy-expected-rq-timeout-ms:[60000]
    Response header x-envoy-upstream-service-time:[8]
    Response header x-seldon-route:[:mul10_1: :pipeline-mul10.pipeline:]
    Response header date:[Fri, 05 Aug 2022 06:20:23 GMT]
    Response header server:[envoy]
    Response header content-type:[application/grpc]
```
````

```bash
seldon pipeline infer pipeline-add10 -s --show-headers --inference-mode grpc \
 '{"model_name":"add10","inputs":[{"name":"INPUT","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]}]}' 
```
````{collapse} Expand to see output
```json

    Request metadata x-seldon-route:[:mul10_1: :pipeline-mul10.pipeline:]
    {"outputs":[{"name":"OUTPUT","datatype":"FP32","shape":["4"],"contents":{"fp32Contents":[10,20,30,40]}}],"rawOutputContents":["AAAgQQAAoEEAAPBBAAAgQg=="]}
    Response header x-seldon-route:[:mul10_1: :pipeline-mul10.pipeline:]
    Response header x-forwarded-proto:[http]
    Response header date:[Fri, 05 Aug 2022 06:20:25 GMT]
    Response header server:[envoy]
    Response header content-type:[application/grpc]
    Response header x-request-id:[08b2e88e-e3c8-4de9-b44d-99f8563c2053]
    Response header x-envoy-upstream-service-time:[7]
    Response header x-envoy-expected-rq-timeout-ms:[60000]
```
````

```bash
seldon experiment stop addmul10
seldon experiment stop add1020
seldon pipeline unload pipeline-add10
seldon pipeline unload pipeline-mul10
seldon model unload add10
seldon model unload add20
seldon model unload mul10
```
````{collapse} Expand to see output
```json

    {}
    {}
    {}
    {}
    {}
    {}
    {}
```
````

```python

```
