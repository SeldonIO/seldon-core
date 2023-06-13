## Inference Examples

We will show:

 * Model inference to a Tensorflow model
   * REST and gRPC using seldon CLI, curl and grpcurl
 * Pipeline inference
   * REST and gRPC using seldon CLI, curl and grpcurl


```python
%env INFER_ENDPOINT=0.0.0.0:9000
```

```
env: INFER_ENDPOINT=0.0.0.0:9000

```

### Tensorflow Model

```bash
cat ./models/tfsimple1.yaml
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

```

Load the model.

```bash
seldon model load -f ./models/tfsimple1.yaml
```

```json
{}

```

Wait for the model to be ready.

```bash
seldon model status tfsimple1 -w ModelAvailable | jq -M .
```

```json
{}

```

```bash
seldon model infer tfsimple1 --inference-host ${INFER_ENDPOINT} \
    '{"inputs":[{"name":"INPUT0","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,16]},{"name":"INPUT1","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,16]}]}'
```

```json
{
	"model_name": "tfsimple1_1",
	"model_version": "1",
	"outputs": [
		{
			"name": "OUTPUT0",
			"datatype": "INT32",
			"shape": [
				1,
				16
			],
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
			]
		},
		{
			"name": "OUTPUT1",
			"datatype": "INT32",
			"shape": [
				1,
				16
			],
			"data": [
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0
			]
		}
	]
}

```

```bash
seldon model infer tfsimple1 --inference-mode grpc  --inference-host ${INFER_ENDPOINT} \
    '{"model_name":"tfsimple1","inputs":[{"name":"INPUT0","contents":{"intContents":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]},"datatype":"INT32","shape":[1,16]},{"name":"INPUT1","contents":{"intContents":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]},"datatype":"INT32","shape":[1,16]}]}'
```

```json
{"modelName":"tfsimple1_1","modelVersion":"1","outputs":[{"name":"OUTPUT0","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[2,4,6,8,10,12,14,16,18,20,22,24,26,28,30,32]}},{"name":"OUTPUT1","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0]}}]}

```

```bash
curl http://${INFER_ENDPOINT}/v2/models/tfsimple1/infer -H "Content-Type: application/json" -H "seldon-model: tfsimple1" \
        -d '{"inputs":[{"name":"INPUT0","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,16]},{"name":"INPUT1","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,16]}]}'
```

```json
{"model_name":"tfsimple1_1","model_version":"1","outputs":[{"name":"OUTPUT0","datatype":"INT32","shape":[1,16],"data":[2,4,6,8,10,12,14,16,18,20,22,24,26,28,30,32]},{"name":"OUTPUT1","datatype":"INT32","shape":[1,16],"data":[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0]}]}

```

```bash
grpcurl -d '{"model_name":"tfsimple1","inputs":[{"name":"INPUT0","contents":{"int_contents":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]},"datatype":"INT32","shape":[1,16]},{"name":"INPUT1","contents":{"int_contents":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]},"datatype":"INT32","shape":[1,16]}]}' \
    -plaintext \
    -import-path ../apis \
    -proto ../apis/mlops/v2_dataplane/v2_dataplane.proto \
    -rpc-header seldon-model:tfsimple1 \
    ${INFER_ENDPOINT} inference.GRPCInferenceService/ModelInfer
```

```json
{
  "modelName": "tfsimple1_1",
  "modelVersion": "1",
  "outputs": [
    {
      "name": "OUTPUT0",
      "datatype": "INT32",
      "shape": [
        "1",
        "16"
      ]
    },
    {
      "name": "OUTPUT1",
      "datatype": "INT32",
      "shape": [
        "1",
        "16"
      ]
    }
  ],
  "rawOutputContents": [
    "AgAAAAQAAAAGAAAACAAAAAoAAAAMAAAADgAAABAAAAASAAAAFAAAABYAAAAYAAAAGgAAABwAAAAeAAAAIAAAAA==",
    "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=="
  ]
}

```

```bash
cat ./pipelines/tfsimple.yaml
```

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Pipeline
metadata:
  name: tfsimple
spec:
  steps:
    - name: tfsimple1
  output:
    steps:
    - tfsimple1

```

```bash
seldon pipeline load -f ./pipelines/tfsimple.yaml
```

```json
{}

```

```bash
seldon pipeline status tfsimple -w PipelineReady
```

```json
{"pipelineName":"tfsimple","versions":[{"pipeline":{"name":"tfsimple","uid":"cg5fm6c6dpcs73c4qhe0","version":1,"steps":[{"name":"tfsimple1"}],"output":{"steps":["tfsimple1.outputs"]},"kubernetesMeta":{}},"state":{"pipelineVersion":1,"status":"PipelineReady","reason":"created pipeline","lastChangeTimestamp":"2023-03-10T09:40:41.317797761Z","modelsReady":true}}]}

```

```bash
seldon pipeline infer tfsimple  --inference-host ${INFER_ENDPOINT} \
    '{"inputs":[{"name":"INPUT0","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,16]},{"name":"INPUT1","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,16]}]}'
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
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0,
				0
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
seldon pipeline infer tfsimple --inference-mode grpc  --inference-host ${INFER_ENDPOINT} \
    '{"model_name":"tfsimple1","inputs":[{"name":"INPUT0","contents":{"int_contents":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]},"datatype":"INT32","shape":[1,16]},{"name":"INPUT1","contents":{"int_contents":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]},"datatype":"INT32","shape":[1,16]}]}'
```

```json
{"outputs":[{"name":"OUTPUT0","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[2,4,6,8,10,12,14,16,18,20,22,24,26,28,30,32]}},{"name":"OUTPUT1","datatype":"INT32","shape":["1","16"],"contents":{"intContents":[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0]}}]}

```

```bash
curl http://${INFER_ENDPOINT}/v2/models/tfsimple1/infer -H "Content-Type: application/json" -H "seldon-model: tfsimple.pipeline" \
        -d '{"inputs":[{"name":"INPUT0","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,16]},{"name":"INPUT1","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,16]}]}'
```

```json
{"model_name":"","outputs":[{"data":[2,4,6,8,10,12,14,16,18,20,22,24,26,28,30,32],"name":"OUTPUT0","shape":[1,16],"datatype":"INT32"},{"data":[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0],"name":"OUTPUT1","shape":[1,16],"datatype":"INT32"}]}

```

```bash
grpcurl -d '{"model_name":"tfsimple1","inputs":[{"name":"INPUT0","contents":{"int_contents":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]},"datatype":"INT32","shape":[1,16]},{"name":"INPUT1","contents":{"int_contents":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]},"datatype":"INT32","shape":[1,16]}]}' \
    -plaintext \
    -import-path ../apis \
    -proto ../apis/mlops/v2_dataplane/v2_dataplane.proto \
    -rpc-header seldon-model:tfsimple.pipeline \
    ${INFER_ENDPOINT} inference.GRPCInferenceService/ModelInfer
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
      ]
    },
    {
      "name": "OUTPUT1",
      "datatype": "INT32",
      "shape": [
        "1",
        "16"
      ]
    }
  ],
  "rawOutputContents": [
    "AgAAAAQAAAAGAAAACAAAAAoAAAAMAAAADgAAABAAAAASAAAAFAAAABYAAAAYAAAAGgAAABwAAAAeAAAAIAAAAA==",
    "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=="
  ]
}

```

```bash
seldon pipeline unload tfsimple
seldon model unload tfsimple1
```

```json
{}
{}

```

```python

```
