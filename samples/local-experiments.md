## Seldon V2 Non Kubernetes Local Experiment Examples


### Model Experiment

We will use two SKlearn Iris classification models to illustrate experiments.


```bash
cat ./experiments/sklearn1.yaml
```
````{collapse} Expand to see output
```yaml
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: iris
      namespace: seldon-mesh
    spec:
      storageUri: "gs://seldon-models/mlserver/iris"
      requirements:
      - sklearn
```
````

```bash
cat ./experiments/sklearn2.yaml
```
````{collapse} Expand to see output
```yaml
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: iris2
      namespace: seldon-mesh
    spec:
      storageUri: "gs://seldon-models/mlserver/iris"
      requirements:
      - sklearn
```
````
Load both models.


```bash
seldon model load -f ./experiments/sklearn1.yaml
seldon model load -f ./experiments/sklearn2.yaml
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
      namespace: seldon-mesh
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
      "kubernetesMeta": {
        "namespace": "seldon-mesh"
      }
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

    map[:iris2_1::20 :iris_1::30]
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
    Response header Server:[envoy]
    Response header X-Seldon-Route:[:iris2_1:]
    Response header Ce-Modelid:[iris2_1]
    Response header Ce-Requestid:[f2e3f1eb-38b1-4086-9c76-1ce8fc1b2c0c]
    Response header Content-Type:[application/json]
    Response header Date:[Tue, 12 Jul 2022 18:19:40 GMT]
    Response header X-Envoy-Upstream-Service-Time:[2]
    Response header Ce-Endpoint:[iris2_1]
    Response header Ce-Type:[io.seldon.serving.inference.response]
    Response header Content-Length:[229]
    Response header Ce-Id:[f2e3f1eb-38b1-4086-9c76-1ce8fc1b2c0c]
    Response header Traceparent:[00-2e25c1e12bf3550935def5a78731d093-b6287b2cc2f9a78e-01]
    Response header X-Request-Id:[04191ae7-e6d5-4753-bd36-e7def99802ff]
    Response header Ce-Inferenceservicename:[mlserver]
    Response header Ce-Source:[io.seldon.serving.deployment.mlserver]
    Response header Ce-Specversion:[0.3]
    {
    	"model_name": "iris2_1",
    	"model_version": "1",
    	"id": "f2e3f1eb-38b1-4086-9c76-1ce8fc1b2c0c",
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

    map[:iris2_1::50]
```
````

```bash
seldon model infer iris --inference-mode grpc -s -i 50\
   '{"model_name":"iris","inputs":[{"name":"input","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[1,4]}]}' 
```
````{collapse} Expand to see output
```json

    map[:iris2_1::50]
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
      namespace: seldon-mesh
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
      namespace: seldon-mesh
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
      namespace: seldon-mesh
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
      namespace: seldon-mesh
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

    {"pipelineName":"pipeline-add10", "versions":[{"pipeline":{"name":"pipeline-add10", "uid":"cb6rm570kj6np6k8gmmg", "version":1, "steps":[{"name":"add10"}], "output":{"steps":["add10.outputs"]}, "kubernetesMeta":{"namespace":"seldon-mesh"}}, "state":{"pipelineVersion":1, "status":"PipelineReady", "reason":"Created pipeline", "lastChangeTimestamp":"2022-07-12T18:19:01.877501308Z"}}]}
    {"pipelineName":"pipeline-mul10", "versions":[{"pipeline":{"name":"pipeline-mul10", "uid":"cb6rm570kj6np6k8gmn0", "version":1, "steps":[{"name":"mul10"}], "output":{"steps":["mul10.outputs"]}, "kubernetesMeta":{"namespace":"seldon-mesh"}}, "state":{"pipelineVersion":1, "status":"PipelineReady", "reason":"Created pipeline", "lastChangeTimestamp":"2022-07-12T18:19:02.394919400Z"}}]}
```
````

```bash
seldon pipeline infer pipeline-add10 --inference-mode grpc \
 '{"model_name":"add10","inputs":[{"name":"INPUT","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]}]}' 
```
````{collapse} Expand to see output
```json

    {"outputs":[{"name":"OUTPUT", "datatype":"FP32", "shape":["4"], "contents":{"fp32Contents":[11, 12, 13, 14]}}], "rawOutputContents":["AAAwQQAAQEEAAFBBAABgQQ=="]}
```
````

```bash
seldon pipeline infer pipeline-mul10 --inference-mode grpc \
 '{"model_name":"add10","inputs":[{"name":"INPUT","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]}]}' 
```
````{collapse} Expand to see output
```json

    {"outputs":[{"name":"OUTPUT", "datatype":"FP32", "shape":["4"], "contents":{"fp32Contents":[10, 20, 30, 40]}}], "rawOutputContents":["AAAgQQAAoEEAAPBBAAAgQg=="]}
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
      namespace: seldon-mesh
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
      "kubernetesMeta": {
        "namespace": "seldon-mesh"
      }
    }
```
````

```bash
seldon pipeline infer pipeline-add10 -i 50 --inference-mode grpc \
 '{"model_name":"add10","inputs":[{"name":"INPUT","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]}]}' 
```
````{collapse} Expand to see output
```json

    map[:add10_1::24 :mul10_1::26 :pipeline-add10.pipeline::24 :pipeline-mul10.pipeline::26]
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
    {"outputs":[{"name":"OUTPUT", "datatype":"FP32", "shape":["4"], "contents":{"fp32Contents":[10, 20, 30, 40]}}], "rawOutputContents":["AAAgQQAAoEEAAPBBAAAgQg=="]}
    Response header x-forwarded-proto:[http]
    Response header x-request-id:[bd99da3a-ab25-45d9-a276-b900ad628929]
    Response header x-envoy-expected-rq-timeout-ms:[60000]
    Response header x-seldon-route:[:mul10_1: :pipeline-mul10.pipeline:]
    Response header x-envoy-upstream-service-time:[11]
    Response header date:[Tue, 12 Jul 2022 18:19:11 GMT]
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
    {"outputs":[{"name":"OUTPUT", "datatype":"FP32", "shape":["4"], "contents":{"fp32Contents":[10, 20, 30, 40]}}], "rawOutputContents":["AAAgQQAAoEEAAPBBAAAgQg=="]}
    Response header x-seldon-route:[:mul10_1: :pipeline-mul10.pipeline:]
    Response header x-forwarded-proto:[http]
    Response header x-request-id:[32e3f339-b40b-4f85-8e44-78f0e9c7b843]
    Response header x-envoy-upstream-service-time:[11]
    Response header date:[Tue, 12 Jul 2022 18:19:12 GMT]
    Response header server:[envoy]
    Response header content-type:[application/grpc]
    Response header x-envoy-expected-rq-timeout-ms:[60000]
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
      namespace: seldon-mesh
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
      namespace: seldon-mesh
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
      "kubernetesMeta": {
        "namespace": "seldon-mesh"
      }
    }
```
````

```bash
seldon model infer add10 -i 50  --inference-mode grpc \
  '{"model_name":"add10","inputs":[{"name":"INPUT","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]}]}' 
```
````{collapse} Expand to see output
```json

    map[:add10_1::23 :add20_1::27]
```
````

```bash
seldon pipeline infer pipeline-add10 -i 100 --inference-mode grpc \
 '{"model_name":"add10","inputs":[{"name":"INPUT","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]}]}' 
```
````{collapse} Expand to see output
```json

    map[:add10_1::33 :add20_1::23 :mul10_1::44 :pipeline-add10.pipeline::56 :pipeline-mul10.pipeline::44]
```
````

```bash
seldon pipeline infer pipeline-add10 -s --show-headers --inference-mode grpc \
 '{"model_name":"add10","inputs":[{"name":"INPUT","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]}]}' 
```
````{collapse} Expand to see output
```json

    Request metadata x-seldon-route:[:add10_1: :pipeline-add10.pipeline:]
    {"modelName":"add10_1", "modelVersion":"1", "outputs":[{"name":"OUTPUT", "datatype":"FP32", "shape":["4"], "contents":{"fp32Contents":[11, 12, 13, 14]}}], "rawOutputContents":["AAAwQQAAQEEAAFBBAABgQQ=="]}
    Response header x-request-id:[27cdbf2c-57f0-4797-9bf1-05f80fdead48]
    Response header content-type:[application/grpc]
    Response header x-envoy-upstream-service-time:[1]
    Response header x-seldon-route:[:add10_1:]
    Response header date:[Tue, 12 Jul 2022 18:19:24 GMT]
    Response header server:[envoy]
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
