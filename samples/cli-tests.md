## CLI Tests


## List resources


```bash
seldon model load -f ./models/sklearn-iris-gs.yaml
seldon model load -f ./models/tfsimple1.yaml
seldon model load -f ./experiments/sklearn2.yaml
seldon model load -f ./models/cifar10.yaml
seldon experiment start -f ./experiments/ab-default-model.yaml 
seldon pipeline load -f ./pipelines/cifar10.yaml
```
````{collapse} Expand to see output
```json

    {}
    {}
    {}
    {}
    {}
    {}
```
````

```bash
seldon model list
```
````{collapse} Expand to see output
```json

    model		state			reason
    -----		-----			------
    iris2		ModelAvailable		
    cifar10		ModelProgressing	
    iris		ModelAvailable		
    tfsimple1	ModelAvailable		
```
````

```bash
seldon model metadata cifar10 
```
````{collapse} Expand to see output
```json

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
```
````

```bash
seldon server list
```
````{collapse} Expand to see output
```json

    server		replicas	models
    ------		--------	------
    triton		1		2
    mlserver	1		2
```
````

```bash
seldon experiment list
```
````{collapse} Expand to see output
```json

    experiment		active	
    ----------		------	
    experiment-sample	true
```
````

```bash
seldon pipeline list
```
````{collapse} Expand to see output
```json

    pipeline
    --------
    cifar10-production
```
````

```bash
seldon model unload iris
seldon model unload tfsimple1
seldon model unload iris2
seldon model unload cifar10
seldon experiment stop experiment-sample
seldon pipeline unload cifar10-production
```
````{collapse} Expand to see output
```json

    {}
    {}
    {}
    {}
    {}
    {}
```
````
### Badly formed model



```bash
cat ./models/error-bad-spec.yaml
```
````{collapse} Expand to see output
```yaml
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: iris
      namespace: seldon-mesh
    spec:
      storagUri: "gs://seldon-models/mlserver/iris"
      requirements:
      - sklearn
```
````

```bash
seldon model load -f ./models/error-bad-spec.yaml
```
````{collapse} Expand to see output
```json

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
```
````

```bash
cat ./pipelines/error-bad-spec.yaml
```
````{collapse} Expand to see output
```yaml
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
```
````

```bash
seldon pipeline load -f ./pipelines/error-bad-spec.yaml
```
````{collapse} Expand to see output
```json

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
```
````

```bash
cat ./experiments/error-bad-spec.yaml
```
````{collapse} Expand to see output
```yaml
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
```
````

```bash
seldon experiment start -f ./experiments/error-bad-spec.yaml
```
````{collapse} Expand to see output
```json

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
```
````
## Failed scheduling


```bash
seldon model load -f ./models/error-bad-capabilities.yaml
```
````{collapse} Expand to see output
```json

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
```
````

```bash
seldon model unload badcapabilities
```
````{collapse} Expand to see output
```json

    {}
```
````

```python

```
