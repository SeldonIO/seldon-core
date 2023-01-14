## CLI Tests

## List resources

```bash
seldon model load -f ./models/sklearn-iris-gs.yaml
seldon model load -f ./models/tfsimple1.yaml
seldon model load -f ./models/sklearn2.yaml
seldon model load -f ./models/cifar10.yaml
seldon experiment start -f ./experiments/ab-default-model.yaml
seldon pipeline load -f ./pipelines/cifar10.yaml
```

```json
{}
{}
{}
{}
{}
{}

```

```bash
seldon model list
```

```
model		state		reason
-----		-----		------
iris		ModelAvailable
tfsimple1	ModelAvailable
iris2		ModelAvailable
cifar10		ModelAvailable

```

```bash
seldon model metadata cifar10
```

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

```bash
seldon server list
```

```
server		replicas	models
------		--------	------
mlserver	1		2
triton		1		2

```

```bash
seldon experiment list
```

```
experiment		active
----------		------
experiment-sample	true

```

```bash
seldon pipeline list
```

```
pipeline		state		reason
--------		-----		------
cifar10-production	PipelineReady	created pipeline

```

```bash
seldon model unload iris
seldon model unload tfsimple1
seldon model unload iris2
seldon model unload cifar10
seldon experiment stop experiment-sample
seldon pipeline unload cifar10-production
```

```json
{}
{}
{}
{}
{}
{}

```

### Resource Errors from CLI

```bash
cat ./models/error-bad-spec.yaml
```

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Model
metadata:
  name: iris
spec:
  storagUri: "gs://seldon-models/mlserver/iris"
  requirements:
  - sklearn

```

```bash
seldon model load -f ./models/error-bad-spec.yaml
```

```
Error: json: unknown field "storagUri"

```

```bash
cat ./pipelines/error-bad-spec.yaml
```

```yaml
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

```

```bash
seldon pipeline load -f ./pipelines/error-bad-spec.yaml
```

```
Error: json: unknown field "input"

```

```bash
cat ./experiments/error-bad-spec.yaml
```

```yaml
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

```

```bash
seldon experiment start -f ./experiments/error-bad-spec.yaml
```

```
Error: json: unknown field "candidate"

```

```bash
cat ./pipelines/error-step-name.yaml
```

```yaml
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

```

```bash
seldon pipeline load -f ./pipelines/error-step-name.yaml
```

```
Error: rpc error: code = FailedPrecondition desc = pipeline iris must not have a step name with the same name as pipeline name

```

```bash
cat ./pipelines/error-empty-input.yaml
```

```yaml
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

```

```bash
seldon pipeline load -f ./pipelines/error-empty-input.yaml
```

```
Error: rpc error: code = FailedPrecondition desc = pipeline iris-pipeline step second has an empty input

```

```bash
cat ./pipelines/error-empty-trigger.yaml
```

```yaml
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

```

```bash
seldon pipeline load -f ./pipelines/error-empty-trigger.yaml
```

```
Error: rpc error: code = FailedPrecondition desc = pipeline iris-pipeline step third has an empty trigger

```

## Failed scheduling errors from CLI

```bash
seldon model load -f ./models/error-bad-capabilities.yaml
```

```
Error: rpc error: code = FailedPrecondition desc = failed to schedule model badcapabilities. [failed replica filter RequirementsReplicaFilter for server replica mlserver:0 : model requirements [foobar] replica capabilities [mlserver alibi-detect alibi-explain huggingface lightgbm mlflow python sklearn spark-mlib xgboost] failed replica filter RequirementsReplicaFilter for server replica triton:0 : model requirements [foobar] replica capabilities [triton dali fil onnx openvino python pytorch tensorflow tensorrt]]

```

```bash
seldon model unload badcapabilities
```

```json
{}

```

```python

```
