## CLI Tests


## List resources


```python
!seldon model load -f ./models/sklearn-iris-gs.yaml
!seldon model load -f ./models/tfsimple1.yaml
!seldon model load -f ./experiments/sklearn2.yaml
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

    model		state			reason
    -----		-----			------
    iris2		ModelAvailable		
    cifar10		ModelProgressing	
    iris		ModelAvailable		
    tfsimple1	ModelAvailable		



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
    triton		1		2
    mlserver	1		2



```python
!seldon experiment list
```

    experiment		active	
    ----------		------	
    experiment-sample	true



```python
!seldon pipeline list
```

    pipeline
    --------
    cifar10-production



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


### Badly formed model



```python
!cat ./models/error-bad-spec.yaml
```

    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: iris
      namespace: seldon-mesh
    spec:
      storagUri: "gs://seldon-models/mlserver/iris"
      requirements:
      - sklearn



```python
!seldon model load -f ./models/error-bad-spec.yaml
```

    Error: json: unknown field "storagUri"
    Usage:
      seldon model load [flags]
    
    Flags:
      -f, --file-path string        model file to load
      -h, --help                    help for load
          --scheduler-host string   seldon scheduler host (default "0.0.0.0:9004")
    
    Global Flags:
      -r, --show-request    show request
      -o, --show-response   show response (default true)
    
    json: unknown field "storagUri"



```python
!cat ./pipelines/error-bad-spec.yaml
```

    apiVersion: mlops.seldon.io/v1alpha1
    kind: Pipeline
    metadata:
      name: tfsimple-conditional
      namespace: seldon-mesh
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
    Usage:
      seldon pipeline load [flags]
    
    Flags:
      -f, --file-path string        pipeline file to load
      -h, --help                    help for load
          --scheduler-host string   seldon scheduler host (default "0.0.0.0:9004")
    
    Global Flags:
      -r, --show-request    show request
      -o, --show-response   show response (default true)
    
    json: unknown field "input"



```python
!cat ./experiments/error-bad-spec.yaml
```

    apiVersion: mlops.seldon.io/v1alpha1
    kind: Experiment
    metadata:
      name: experiment-sample
      namespace: seldon-mesh
    spec:
      defaultModel: iris
      candidate:
      - modelName: iris
        weight: 50
      - modelName: iris2
        weight: 50



```python
!seldon experiment start -f ./experiments/error-bad-spec.yaml
```

    Error: json: unknown field "candidate"
    Usage:
      seldon experiment start [flags]
    
    Flags:
      -f, --file-path string        model file to load
      -h, --help                    help for start
          --scheduler-host string   seldon scheduler host (default "0.0.0.0:9004")
    
    Global Flags:
      -r, --show-request    show request
      -o, --show-response   show response (default true)
    
    json: unknown field "candidate"



```python
!seldon pipeline load -f ./pipelines/error-step-name.yaml
```

    Error: rpc error: code = FailedPrecondition desc = pipeline iris must not have a step name with the same name
    Usage:
      seldon pipeline load [flags]
    
    Flags:
      -f, --file-path string        pipeline file to load
      -h, --help                    help for load
          --scheduler-host string   seldon scheduler host (default "0.0.0.0:9004")
    
    Global Flags:
      -r, --show-request    show request
      -o, --show-response   show response (default true)
    
    rpc error: code = FailedPrecondition desc = pipeline iris must not have a step name with the same name


## Failed scheduling


```python
!seldon model load -f ./models/error-bad-capabilities.yaml
```

    Error: rpc error: code = FailedPrecondition desc = failed to schedule model badcapabilities. [failed replica filter RequirementsReplicaFilter for server replica triton:0 : model requirements [foobar] replica capabilities [triton dali fil onnx openvino python pytorch tensorflow tensorrt] failed replica filter RequirementsReplicaFilter for server replica mlserver:0 : model requirements [foobar] replica capabilities [mlserver alibi-detect lightgbm mlflow python sklearn spark-mlib xgboost]]
    Usage:
      seldon model load [flags]
    
    Flags:
      -f, --file-path string        model file to load
      -h, --help                    help for load
          --scheduler-host string   seldon scheduler host (default "0.0.0.0:9004")
    
    Global Flags:
      -r, --show-request    show request
      -o, --show-response   show response (default true)
    
    rpc error: code = FailedPrecondition desc = failed to schedule model badcapabilities. [failed replica filter RequirementsReplicaFilter for server replica triton:0 : model requirements [foobar] replica capabilities [triton dali fil onnx openvino python pytorch tensorflow tensorrt] failed replica filter RequirementsReplicaFilter for server replica mlserver:0 : model requirements [foobar] replica capabilities [mlserver alibi-detect lightgbm mlflow python sklearn spark-mlib xgboost]]



```python
!seldon model unload badcapabilities
```

    {}

