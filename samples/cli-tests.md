## CLI Tests


## List resources


```python
!seldon model load -f ./models/sklearn-iris-gs.yaml
!seldon model load -f ./models/tfsimple1.yaml
!seldon model load -f ./models/sklearn2.yaml
!seldon model load -f ./models/cifar10.yaml
!seldon experiment start -f ./experiments/ab-default-model.yaml 
!seldon pipeline load -f ./pipelines/cifar10.yaml
```

    {}
    {}
    {}
    {}
    {}
    {}



```python
!seldon model list
```

    model		state		reason
    -----		-----		------
    iris		ModelAvailable	
    tfsimple1	ModelAvailable	
    iris2		ModelAvailable	
    cifar10		ModelAvailable	



```python
!seldon model metadata cifar10 
```

    {
    	"name": "cifar10_1",
    	"versions": [
    		"1"
    	],
    	"platform": "tensorflow_savedmodel",
    	"inputs": [
    		{
    			"name": "input_1",
    			"datatype": "FP32",
    			"shape": [
    				-1,
    				32,
    				32,
    				3
    			]
    		}
    	],
    	"outputs": [
    		{
    			"name": "fc10",
    			"datatype": "FP32",
    			"shape": [
    				-1,
    				10
    			]
    		}
    	]
    }



```python
!seldon server list
```

    server		replicas	models
    ------		--------	------
    mlserver	1		2
    triton		1		2



```python
!seldon experiment list
```

    experiment		active	
    ----------		------	
    experiment-sample	true



```python
!seldon pipeline list
```

    pipeline		state		reason
    --------		-----		------
    cifar10-production	PipelineReady	created pipeline



```python
!seldon model unload iris
!seldon model unload tfsimple1
!seldon model unload iris2
!seldon model unload cifar10
!seldon experiment stop experiment-sample
!seldon pipeline unload cifar10-production
```

    {}
    {}
    {}
    {}
    {}
    {}


### Resource Errors from CLI


```python
!cat ./models/error-bad-spec.yaml
```

    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: iris
    spec:
      storagUri: "gs://seldon-models/mlserver/iris"
      requirements:
      - sklearn



```python
!seldon model load -f ./models/error-bad-spec.yaml
```

    Error: json: unknown field "storagUri"



```python
!cat ./pipelines/error-bad-spec.yaml
```

    apiVersion: mlops.seldon.io/v1alpha1
    kind: Pipeline
    metadata:
      name: tfsimple-conditional
    spec:
      steps:
      - name: conditional
      - name: mul10
        input:
        - conditional.outputs.OUTPUT0
        tensorMap:
          conditional.outputs.OUTPUT0: INPUT
      - name: add10
        inputs:
        - conditional.outputs.OUTPUT1
        tensorMap:
          conditional.outputs.OUTPUT1: INPUT
      output:
        steps:
        - mul10
        - add10
        stepsJoin: any



```python
!seldon pipeline load -f ./pipelines/error-bad-spec.yaml
```

    Error: json: unknown field "input"



```python
!cat ./experiments/error-bad-spec.yaml
```

    apiVersion: mlops.seldon.io/v1alpha1
    kind: Experiment
    metadata:
      name: experiment-sample
    spec:
      default: iris
      candidate:
      - name: iris
        weight: 50
      - name: iris2
        weight: 50



```python
!seldon experiment start -f ./experiments/error-bad-spec.yaml
```

    Error: json: unknown field "candidate"



```python
!cat ./pipelines/error-step-name.yaml
```

    apiVersion: mlops.seldon.io/v1alpha1
    kind: Pipeline
    metadata:
      name: iris
    spec:
      steps:
        - name: iris
      output:
        steps:
        - iris



```python
!seldon pipeline load -f ./pipelines/error-step-name.yaml
```

    Error: rpc error: code = FailedPrecondition desc = pipeline iris must not have a step name with the same name as pipeline name



```python
!cat ./pipelines/error-empty-input.yaml
```

    apiVersion: mlops.seldon.io/v1alpha1
    kind: Pipeline
    metadata:
      name: iris-pipeline
    spec:
      steps:
        - name: first
        - name: second
          inputs:
            -
      output:
        steps:
        - iris



```python
!seldon pipeline load -f ./pipelines/error-empty-input.yaml
```

    Error: rpc error: code = FailedPrecondition desc = pipeline iris-pipeline step second has an empty input



```python
!cat ./pipelines/error-empty-trigger.yaml
```

    apiVersion: mlops.seldon.io/v1alpha1
    kind: Pipeline
    metadata:
      name: iris-pipeline
    spec:
      steps:
        - name: first
        - name: second
        - name: third
          inputs:
            - first
          triggers:
            -
      output:
        steps:
        - iris



```python
!seldon pipeline load -f ./pipelines/error-empty-trigger.yaml
```

    Error: rpc error: code = FailedPrecondition desc = pipeline iris-pipeline step third has an empty trigger


## Failed scheduling errors from CLI


```python
!seldon model load -f ./models/error-bad-capabilities.yaml
```

    Error: rpc error: code = FailedPrecondition desc = failed to schedule model badcapabilities. [failed replica filter RequirementsReplicaFilter for server replica mlserver:0 : model requirements [foobar] replica capabilities [mlserver alibi-detect alibi-explain huggingface lightgbm mlflow python sklearn spark-mlib xgboost] failed replica filter RequirementsReplicaFilter for server replica triton:0 : model requirements [foobar] replica capabilities [triton dali fil onnx openvino python pytorch tensorflow tensorrt]]



```python
!seldon model unload badcapabilities
```

    {}



```python

```
