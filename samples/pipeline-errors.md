## Seldon V2 Pipeline Errors

```bash
cat ./models/tfsimple1.yaml
cat ./models/tfsimple2.yaml
```

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Model
metadata:
  name: tfsimple1
spec:
  storageUri: "gs://seldon-models/triton/simple"
  requirements:
  - tensorflow
  memory: 100Ki
apiVersion: mlops.seldon.io/v1alpha1
kind: Model
metadata:
  name: tfsimple2
spec:
  storageUri: "gs://seldon-models/triton/simple"
  requirements:
  - tensorflow
  memory: 100Ki

```

```bash
seldon model load -f ./models/tfsimple1.yaml
seldon model load -f ./models/tfsimple2.yaml
```

```json
{}
{}

```

```bash
seldon model status tfsimple1 -w ModelAvailable | jq -M .
seldon model status tfsimple2 -w ModelAvailable | jq -M .
```

```json
{}
{}

```

```bash
cat ./pipelines/tfsimples.yaml
```

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Pipeline
metadata:
  name: tfsimples
spec:
  steps:
    - name: tfsimple1
    - name: tfsimple2
      inputs:
      - tfsimple1
      tensorMap:
        tfsimple1.outputs.OUTPUT0: INPUT0
        tfsimple1.outputs.OUTPUT1: INPUT1
  output:
    steps:
    - tfsimple2

```

```bash
seldon pipeline load -f ./pipelines/tfsimples.yaml
```

```json
{}

```

```bash
seldon pipeline status tfsimples -w PipelineReady| jq -M .
```

```json
{
  "pipelineName": "tfsimples",
  "versions": [
    {
      "pipeline": {
        "name": "tfsimples",
        "uid": "cdqjic9qa12c739ab3og",
        "version": 1,
        "steps": [
          {
            "name": "tfsimple1"
          },
          {
            "name": "tfsimple2",
            "inputs": [
              "tfsimple1.outputs"
            ],
            "tensorMap": {
              "tfsimple1.outputs.OUTPUT0": "INPUT0",
              "tfsimple1.outputs.OUTPUT1": "INPUT1"
            }
          }
        ],
        "output": {
          "steps": [
            "tfsimple2.outputs"
          ]
        },
        "kubernetesMeta": {}
      },
      "state": {
        "pipelineVersion": 1,
        "status": "PipelineReady",
        "reason": "created pipeline",
        "lastChangeTimestamp": "2022-11-16T19:25:37.908093300Z"
      }
    }
  ]
}

```

```bash
seldon pipeline infer tfsimples \
```yaml
'{"inputs":[{"name":"INPUT0","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,16]},{"name":"INPUT1","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,16]}]}'
```
```

```json
{
	"model_name": "",
	"outputs": [
		{
			"data": [
				2,
				4,
				6,
				8,
				10,
				12,
				14,
				16,
				18,
				20,
				22,
				24,
				26,
				28,
				30,
				32
			],
			"name": "OUTPUT0",
			"shape": [
				1,
				16
			],
			"datatype": "INT32"
		},
		{
			"data": [
				2,
				4,
				6,
				8,
				10,
				12,
				14,
				16,
				18,
				20,
				22,
				24,
				26,
				28,
				30,
				32
			],
			"name": "OUTPUT1",
			"shape": [
				1,
				16
			],
			"datatype": "INT32"
		}
	]
}

```

```bash
seldon pipeline infer tfsimples \
```yaml
'{"inputs":[{"name":"INPUT0","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,20]},{"name":"INPUT1","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,16]}]}'
```
```

```yaml
Error: V2 server error: 400 tfsimple1 : rpc error: code = InvalidArgument desc = unexpected shape for input 'INPUT0' for model 'tfsimple1_1'. Expected [-1,16], got [1,20]

```

```bash
seldon pipeline infer tfsimples --inference-mode grpc \
```yaml
'{"model_name":"simple","inputs":[{"name":"INPUT0","contents":{"int_contents":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]},"datatype":"INT32","shape":[1,16]},{"name":"INPUT1","contents":{"int_contents":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]},"datatype":"INT32","shape":[1,16]}]}' | jq -M .
```
```

```json
{
  "outputs": [
    {
      "name": "OUTPUT0",
      "datatype": "INT32",
      "shape": [
        "1",
        "16"
      ],
      "contents": {
        "intContents": [
          2,
          4,
          6,
          8,
          10,
          12,
          14,
          16,
          18,
          20,
          22,
          24,
          26,
          28,
          30,
          32
        ]
      }
    },
    {
      "name": "OUTPUT1",
      "datatype": "INT32",
      "shape": [
        "1",
        "16"
      ],
      "contents": {
        "intContents": [
          2,
          4,
          6,
          8,
          10,
          12,
          14,
          16,
          18,
          20,
          22,
          24,
          26,
          28,
          30,
          32
        ]
      }
    }
  ]
}

```

```bash
seldon pipeline infer tfsimples --inference-mode grpc \
```yaml
'{"model_name":"simple","inputs":[{"name":"INPUT0","contents":{"int_contents":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]},"datatype":"INT32","shape":[1,20]},{"name":"INPUT1","contents":{"int_contents":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]},"datatype":"INT32","shape":[1,16]}]}'
```
```

```yaml
Error: rpc error: code = Unknown desc = tfsimple1 : rpc error: code = InvalidArgument desc = unexpected shape for input 'INPUT0' for model 'tfsimple1_1'. Expected [-1,16], got [1,20]

```

```bash
seldon pipeline unload tfsimples
seldon model unload tfsimple1
seldon model unload tfsimple2
```

```json
{}
{}
{}

```

```python

```
