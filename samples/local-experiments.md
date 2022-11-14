## Seldon V2 Non Kubernetes Local Experiment Examples


### Model Experiment

We will use two SKlearn Iris classification models to illustrate experiments.


```python
!cat ./models/sklearn1.yaml
```

    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: iris
    spec:
      storageUri: "gs://seldon-models/mlserver/iris"
      requirements:
      - sklearn



```python
!cat ./models/sklearn2.yaml
```

    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: iris2
    spec:
      storageUri: "gs://seldon-models/mlserver/iris"
      requirements:
      - sklearn


Load both models.


```python
!seldon model load -f ./models/sklearn1.yaml
!seldon model load -f ./models/sklearn2.yaml
```

    {}
    {}


Wait for both models to be ready.


```python
!seldon model status iris -w ModelAvailable
!seldon model status iris2 -w ModelAvailable
```

    {}
    {}



```python
!seldon model infer iris -i 50 \
  '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}' 
```

    map[:iris_1::50]



```python
!seldon model infer iris2 -i 50 \
  '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}' 
```

    map[:iris2_1::50]


Create an experiment that modifies the iris model to add a second model splitting traffic 50/50 between the two.


```python
!cat ./experiments/ab-default-model.yaml 
```

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


Start the experiment.


```python
!seldon experiment start -f ./experiments/ab-default-model.yaml 
```

    {}


Wait for the experiment to be ready.


```python
!seldon experiment status experiment-sample -w | jq -M .
```

    {
      "experimentName": "experiment-sample",
      "active": true,
      "candidatesReady": true,
      "mirrorReady": true,
      "statusDescription": "experiment active",
      "kubernetesMeta": {}
    }


Run a set of calls and record which route the traffic took. There should be roughly a 50/50 split.


```python
!seldon model infer iris -i 50 \
  '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}' 
```

    map[:iris2_1::19 :iris_1::31]


Show sticky session header `x-seldon-route` that is returned


```python
!seldon model infer iris --show-headers \
  '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}' 
```

    Request header Content-Type:[application/json]
    Request header Seldon-Model:[iris]
    Response header Ce-Requestid:[0757e893-64c9-411f-8937-f0f4774852ef]
    Response header Server:[envoy]
    Response header Ce-Endpoint:[iris_1]
    Response header Date:[Mon, 29 Aug 2022 13:12:01 GMT]
    Response header X-Envoy-Upstream-Service-Time:[2]
    Response header Ce-Specversion:[0.3]
    Response header Ce-Modelid:[iris_1]
    Response header Ce-Source:[io.seldon.serving.deployment.mlserver]
    Response header Ce-Type:[io.seldon.serving.inference.response]
    Response header X-Request-Id:[b3255545-9531-4cbe-9895-96c7101e19b8]
    Response header Ce-Inferenceservicename:[mlserver]
    Response header Content-Length:[228]
    Response header Content-Type:[application/json]
    Response header Traceparent:[00-364448c1aff0e9276eb505a0b64421c1-bf3b5e0412c650fd-01]
    Response header X-Seldon-Route:[:iris_1:]
    Response header Ce-Id:[0757e893-64c9-411f-8937-f0f4774852ef]
    {
    	"model_name": "iris_1",
    	"model_version": "1",
    	"id": "0757e893-64c9-411f-8937-f0f4774852ef",
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


Use sticky session key passed by last infer request to ensure same route is taken each time.


```python
!seldon model infer iris -s -i 50 \
  '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}' 
```

    map[:iris_1::50]



```python
!seldon model infer iris --inference-mode grpc -s -i 50\
   '{"model_name":"iris","inputs":[{"name":"input","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[1,4]}]}' 
```

    map[:iris_1::50]


Stop the experiment


```python
!seldon experiment stop experiment-sample
```

    {}


Unload both models.


```python
!seldon model unload iris
!seldon model unload iris2
```

    {}
    {}


### Pipeline Experiment


```python
!cat ./models/add10.yaml
```

    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: add10
    spec:
      storageUri: "gs://seldon-models/triton/add10"
      requirements:
      - triton
      - python



```python
!cat ./models/mul10.yaml
```

    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: mul10
    spec:
      storageUri: "gs://seldon-models/triton/mul10"
      requirements:
      - triton
      - python



```python
!seldon model load -f ./models/add10.yaml
!seldon model load -f ./models/mul10.yaml
```

    {}
    {}



```python
!seldon model status add10 -w ModelAvailable
!seldon model status mul10 -w ModelAvailable
```

    {}
    {}



```python
!cat ./pipelines/mul10.yaml
```

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



```python
!cat ./pipelines/add10.yaml
```

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



```python
!seldon pipeline load -f ./pipelines/add10.yaml
!seldon pipeline load -f ./pipelines/mul10.yaml
```

    {}
    {}



```python
!seldon pipeline status pipeline-add10 -w PipelineReady 
!seldon pipeline status pipeline-mul10 -w PipelineReady 
```

    {"pipelineName":"pipeline-add10","versions":[{"pipeline":{"name":"pipeline-add10","uid":"cc6bmcs5em8of75v7pi0","version":1,"steps":[{"name":"add10"}],"output":{"steps":["add10.outputs"]},"kubernetesMeta":{}},"state":{"pipelineVersion":1,"status":"PipelineReady","reason":"created pipeline","lastChangeTimestamp":"2022-08-29T13:12:19.395809013Z"}}]}
    {"pipelineName":"pipeline-mul10","versions":[{"pipeline":{"name":"pipeline-mul10","uid":"cc6bmcs5em8of75v7pig","version":1,"steps":[{"name":"mul10"}],"output":{"steps":["mul10.outputs"]},"kubernetesMeta":{}},"state":{"pipelineVersion":1,"status":"PipelineReady","reason":"created pipeline","lastChangeTimestamp":"2022-08-29T13:12:19.632179449Z"}}]}



```python
!seldon pipeline infer pipeline-add10 --inference-mode grpc \
 '{"model_name":"add10","inputs":[{"name":"INPUT","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]}]}' | jq -M . 
```

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
      ],
      "rawOutputContents": [
        "AAAwQQAAQEEAAFBBAABgQQ=="
      ]
    }



```python
!seldon pipeline infer pipeline-mul10 --inference-mode grpc \
 '{"model_name":"add10","inputs":[{"name":"INPUT","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]}]}' | jq -M .
```

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
      ],
      "rawOutputContents": [
        "AAAgQQAAoEEAAPBBAAAgQg=="
      ]
    }



```python
!cat ./experiments/addmul10.yaml
```

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



```python
!seldon experiment start -f ./experiments/addmul10.yaml 
```

    {}



```python
!seldon experiment status addmul10 -w | jq -M .
```

    {
      "experimentName": "addmul10",
      "active": true,
      "candidatesReady": true,
      "mirrorReady": true,
      "statusDescription": "experiment active",
      "kubernetesMeta": {}
    }



```python
!seldon pipeline infer pipeline-add10 -i 50 --inference-mode grpc \
 '{"model_name":"add10","inputs":[{"name":"INPUT","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]}]}' 
```

    map[:add10_1::25 :mul10_1::25 :pipeline-add10.pipeline::25 :pipeline-mul10.pipeline::25]


Use sticky session key passed by last infer request to ensure same route is taken each time.


```python
!seldon pipeline infer pipeline-add10 --show-headers --inference-mode grpc \
 '{"model_name":"add10","inputs":[{"name":"INPUT","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]}]}' 
```

    Request metadata seldon-model:[pipeline-add10.pipeline]
    {"outputs":[{"name":"OUTPUT","datatype":"FP32","shape":["4"],"contents":{"fp32Contents":[11,12,13,14]}}],"rawOutputContents":["AAAwQQAAQEEAAFBBAABgQQ=="]}
    Response header x-seldon-route:[:add10_1: :pipeline-add10.pipeline:]
    Response header x-forwarded-proto:[http]
    Response header x-request-id:[76414998-7068-49e7-9731-c6830278017e]
    Response header x-envoy-expected-rq-timeout-ms:[60000]
    Response header date:[Mon, 29 Aug 2022 13:12:58 GMT]
    Response header server:[envoy]
    Response header content-type:[application/grpc]
    Response header x-envoy-upstream-service-time:[10]



```python
!seldon pipeline infer pipeline-add10 -s --show-headers --inference-mode grpc \
 '{"model_name":"add10","inputs":[{"name":"INPUT","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]}]}' 
```

    Request metadata x-seldon-route:[:add10_1: :pipeline-add10.pipeline:]
    Request metadata seldon-model:[pipeline-add10.pipeline]
    {"outputs":[{"name":"OUTPUT","datatype":"FP32","shape":["4"],"contents":{"fp32Contents":[11,12,13,14]}}],"rawOutputContents":["AAAwQQAAQEEAAFBBAABgQQ=="]}
    Response header server:[envoy]
    Response header content-type:[application/grpc]
    Response header x-seldon-route:[:add10_1: :pipeline-add10.pipeline: :pipeline-add10.pipeline:]
    Response header x-forwarded-proto:[http]
    Response header x-request-id:[f33d3527-fcb0-4693-8d1e-b76d96418ee4]
    Response header x-envoy-expected-rq-timeout-ms:[60000]
    Response header x-envoy-upstream-service-time:[11]
    Response header date:[Mon, 29 Aug 2022 13:12:58 GMT]



```python
!seldon pipeline infer pipeline-add10 -s -i 50 --inference-mode grpc \
 '{"model_name":"add10","inputs":[{"name":"INPUT","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]}]}' 
```

    map[:add10_1::50 :pipeline-add10.pipeline::150]



```python
!cat ./models/add20.yaml
```

    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: add20
    spec:
      storageUri: "gs://seldon-models/triton/add20"
      requirements:
      - triton
      - python



```python
!seldon model load -f ./models/add20.yaml
```

    {}



```python
!seldon model status add20 -w ModelAvailable
```

    {}



```python
!cat ./experiments/add1020.yaml
```

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



```python
!seldon experiment start -f ./experiments/add1020.yaml
```

    {}



```python
!seldon experiment status add1020 -w | jq -M .
```

    {
      "experimentName": "add1020",
      "active": true,
      "candidatesReady": true,
      "mirrorReady": true,
      "statusDescription": "experiment active",
      "kubernetesMeta": {}
    }



```python
!seldon model infer add10 -i 50  --inference-mode grpc \
  '{"model_name":"add10","inputs":[{"name":"INPUT","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]}]}' 
```

    map[:add10_1::20 :add20_1::30]



```python
!seldon pipeline infer pipeline-add10 -i 100 --inference-mode grpc \
 '{"model_name":"add10","inputs":[{"name":"INPUT","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]}]}' 
```

    map[:add10_1::27 :add20_1::31 :mul10_1::42 :pipeline-add10.pipeline::58 :pipeline-mul10.pipeline::42]



```python
!seldon pipeline infer pipeline-add10 --show-headers --inference-mode grpc \
 '{"model_name":"add10","inputs":[{"name":"INPUT","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]}]}' 
```

    Request metadata seldon-model:[pipeline-add10.pipeline]
    {"outputs":[{"name":"OUTPUT","datatype":"FP32","shape":["4"],"contents":{"fp32Contents":[10,20,30,40]}}],"rawOutputContents":["AAAgQQAAoEEAAPBBAAAgQg=="]}
    Response header x-forwarded-proto:[http]
    Response header x-request-id:[e130e61e-a20f-480c-9cc6-a276004e9b9f]
    Response header x-envoy-expected-rq-timeout-ms:[60000]
    Response header x-envoy-upstream-service-time:[9]
    Response header x-seldon-route:[:mul10_1: :pipeline-mul10.pipeline:]
    Response header date:[Mon, 29 Aug 2022 13:13:20 GMT]
    Response header server:[envoy]
    Response header content-type:[application/grpc]



```python
!seldon pipeline infer pipeline-add10 -s --show-headers --inference-mode grpc \
 '{"model_name":"add10","inputs":[{"name":"INPUT","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]}]}' 
```

    Request metadata x-seldon-route:[:mul10_1: :pipeline-mul10.pipeline:]
    Request metadata seldon-model:[pipeline-add10.pipeline]
    {"outputs":[{"name":"OUTPUT","datatype":"FP32","shape":["4"],"contents":{"fp32Contents":[10,20,30,40]}}],"rawOutputContents":["AAAgQQAAoEEAAPBBAAAgQg=="]}
    Response header content-type:[application/grpc]
    Response header x-request-id:[a0929eea-0b8f-4856-9a75-1087b4e3fe2b]
    Response header x-envoy-expected-rq-timeout-ms:[60000]
    Response header x-seldon-route:[:mul10_1: :pipeline-mul10.pipeline: :pipeline-mul10.pipeline:]
    Response header x-forwarded-proto:[http]
    Response header x-envoy-upstream-service-time:[8]
    Response header date:[Mon, 29 Aug 2022 13:13:22 GMT]
    Response header server:[envoy]



```python
!seldon experiment stop addmul10
!seldon experiment stop add1020
!seldon pipeline unload pipeline-add10
!seldon pipeline unload pipeline-mul10
!seldon model unload add10
!seldon model unload add20
!seldon model unload mul10
```

    {}
    {}
    {}
    {}
    {}
    {}
    {}


### Model Mirror Experiment

We will use two SKlearn Iris classification models to illustrate a model with a mirror.


```python
!cat ./models/sklearn1.yaml
```

    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: iris
    spec:
      storageUri: "gs://seldon-models/mlserver/iris"
      requirements:
      - sklearn



```python
!cat ./models/sklearn2.yaml
```

    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: iris2
    spec:
      storageUri: "gs://seldon-models/mlserver/iris"
      requirements:
      - sklearn


Load both models.


```python
!seldon model load -f ./models/sklearn1.yaml
!seldon model load -f ./models/sklearn2.yaml
```

    {}
    {}


Wait for both models to be ready.


```python
!seldon model status iris -w ModelAvailable
!seldon model status iris2 -w ModelAvailable
```

    {}
    {}


Create an experiment that modifies in which we mirror traffic to iris also to iris2.


```python
!cat ./experiments/sklearn-mirror.yaml 
```

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
    


Start the experiment.


```python
!seldon experiment start -f ./experiments/sklearn-mirror.yaml
```

    {}


Wait for the experiment to be ready.


```python
!seldon experiment status sklearn-mirror -w | jq -M .
```

    {
      "experimentName": "sklearn-mirror",
      "active": true,
      "candidatesReady": true,
      "mirrorReady": true,
      "statusDescription": "experiment active",
      "kubernetesMeta": {}
    }


We get responses from iris but all requests would also have been mirrored to iris2


```python
!seldon model infer iris -i 50 \
  '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}' 
```

    map[:iris_1::50]


We can check the local prometheus port from the agent to validate requests went to iris2


```python
!curl -s 0.0.0:9006/metrics | grep seldon_model_infer_api_seconds_count | grep iris2_1
```

    seldon_model_infer_api_seconds_count{code="200",method_type="rest",model="iris",model_internal="iris2_1",server="mlserver",server_replica="0"} 50


Stop the experiment


```python
!seldon experiment stop sklearn-mirror
```

    {}


Unload both models.


```python
!seldon model unload iris
!seldon model unload iris2
```

    {}
    {}


## Pipeline Mirror Experiment


```python
!cat ./models/add10.yaml
```

    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: add10
    spec:
      storageUri: "gs://seldon-models/triton/add10"
      requirements:
      - triton
      - python



```python
!cat ./models/mul10.yaml
```

    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: mul10
    spec:
      storageUri: "gs://seldon-models/triton/mul10"
      requirements:
      - triton
      - python



```python
!seldon model load -f ./models/add10.yaml
!seldon model load -f ./models/mul10.yaml
```

    {}
    {}



```python
!seldon model status add10 -w ModelAvailable
!seldon model status mul10 -w ModelAvailable
```

    {}
    {}



```python
!cat ./pipelines/mul10.yaml
```

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



```python
!cat ./pipelines/add10.yaml
```

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



```python
!seldon pipeline load -f ./pipelines/add10.yaml
!seldon pipeline load -f ./pipelines/mul10.yaml
```

    {}
    {}



```python
!seldon pipeline status pipeline-add10 -w PipelineReady 
!seldon pipeline status pipeline-mul10 -w PipelineReady 
```

    {"pipelineName":"pipeline-add10", "versions":[{"pipeline":{"name":"pipeline-add10", "uid":"cc0a78ui50579svh4i5g", "version":1, "steps":[{"name":"add10"}], "output":{"steps":["add10.outputs"]}, "kubernetesMeta":{}}, "state":{"pipelineVersion":1, "status":"PipelineReady", "reason":"Created pipeline", "lastChangeTimestamp":"2022-08-20T09:10:25.432802482Z"}}]}
    {"pipelineName":"pipeline-mul10", "versions":[{"pipeline":{"name":"pipeline-mul10", "uid":"cc0a78ui50579svh4i60", "version":1, "steps":[{"name":"mul10"}], "output":{"steps":["mul10.outputs"]}, "kubernetesMeta":{}}, "state":{"pipelineVersion":1, "status":"PipelineReady", "reason":"Created pipeline", "lastChangeTimestamp":"2022-08-20T09:10:26.057188908Z"}}]}



```python
!seldon pipeline infer pipeline-add10 --inference-mode grpc \
 '{"model_name":"add10","inputs":[{"name":"INPUT","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]}]}' 
```

    {"outputs":[{"name":"OUTPUT", "datatype":"FP32", "shape":["4"], "contents":{"fp32Contents":[11, 12, 13, 14]}}], "rawOutputContents":["AAAwQQAAQEEAAFBBAABgQQ=="]}



```python
!seldon pipeline infer pipeline-mul10 --inference-mode grpc \
 '{"model_name":"add10","inputs":[{"name":"INPUT","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]}]}' 
```

    {"outputs":[{"name":"OUTPUT", "datatype":"FP32", "shape":["4"], "contents":{"fp32Contents":[10, 20, 30, 40]}}], "rawOutputContents":["AAAgQQAAoEEAAPBBAAAgQg=="]}



```python
!cat ./experiments/addmul10-mirror.yaml
```

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
    



```python
!seldon experiment start -f ./experiments/addmul10-mirror.yaml 
```

    {}



```python
!seldon experiment status addmul10-mirror -w | jq -M .
```

    {
      "experimentName": "addmul10-mirror",
      "active": true,
      "candidatesReady": true,
      "mirrorReady": true,
      "statusDescription": "experiment active",
      "kubernetesMeta": {}
    }



```python
!seldon pipeline infer pipeline-add10 -i 50 --inference-mode grpc \
 '{"model_name":"add10","inputs":[{"name":"INPUT","contents":{"fp32_contents":[1,2,3,4]},"datatype":"FP32","shape":[4]}]}' 
```

    map[:add10_1::50 :pipeline-add10.pipeline::50]


Let's check that the mul10 model was called.


```python
!curl -s 0.0.0:9007/metrics | grep seldon_model_infer_api_seconds_count | grep mul10_1
```

    seldon_model_infer_api_seconds_count{code="OK",method_type="grpc",model="mul10",model_internal="mul10_1",server="triton",server_replica="0"} 52



```python
!seldon experiment stop addmul10-mirror
!seldon pipeline unload pipeline-add10
!seldon pipeline unload pipeline-mul10
!seldon model unload add10
!seldon model unload mul10
```

    {}
    {}
    {}
    {}
    {}



```python

```
