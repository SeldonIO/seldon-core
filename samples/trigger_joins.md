## Trigger Joins

In this notebook we show use of trigger joins of type "any" where we wait for any of the inputs tobe present. We illustrate where we send various combinations of 2 inputs and show we always get an output.

```python
import json

def get_request_string(use_trigger_1=True, use_trigger_2=True):
    request = {
        "model_name": "whatever",
        "inputs": [
            {"name": "INPUT1", "contents": {"int64Contents": [1]}, "datatype": "INT64", "shape": [1]},
            {"name": "INPUT2", "contents": {"int64Contents": [1]}, "datatype": "INT64", "shape": [1]},
        ]
    }

    if use_trigger_1:
        request["inputs"].append({"name": "TRIGGER1", "contents": {"boolContents": [True]}, "datatype": "BOOL", "shape": [1]})

    if use_trigger_2:
        request["inputs"].append({"name": "TRIGGER2", "contents": {"boolContents": [True]}, "datatype": "BOOL", "shape": [1]})

    request_string = json.dumps(request)
    return request_string
```

Load models and pipelines

```bash
seldon model load -f ./models/id1_node.yaml
seldon model load -f ./models/id2_node.yaml
seldon model load -f ./models/join_node.yaml
```

```json
{}
{}
{}

```

```bash
seldon model status join_node -w ModelAvailable
seldon model status id1_node -w ModelAvailable
seldon model status id2_node -w ModelAvailable
```

```json
{}
{}
{}

```

```bash
seldon pipeline load -f ./pipelines/triggers_join_inputs.yaml
```

```json
{}

```

```bash
seldon pipeline status triggers_join_inputs -w PipelineReady | jq .
```

```json
{
  "pipelineName": "triggers_join_inputs",
  "versions": [
    {
      "pipeline": {
        "name": "triggers_join_inputs",
        "uid": "cg5g8ak6dpcs73c4qhpg",
        "version": 1,
        "steps": [
          {
            "name": "join_node",
            "inputs": [
              "triggers_join_inputs.inputs.INPUT1",
              "triggers_join_inputs.inputs.INPUT2"
            ],
            "triggers": [
              "triggers_join_inputs.inputs.TRIGGER1",
              "triggers_join_inputs.inputs.TRIGGER2"
            ],
            "triggersJoin": "ANY"
          }
        ],
        "output": {
          "steps": [
            "join_node.outputs"
          ]
        },
        "kubernetesMeta": {}
      },
      "state": {
        "pipelineVersion": 1,
        "status": "PipelineReady",
        "reason": "created pipeline",
        "lastChangeTimestamp": "2023-03-10T10:19:22.997419447Z",
        "modelsReady": true
      }
    }
  ]
}

```

```bash
cat ./pipelines/triggers_join_inputs.yaml
```

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Pipeline
metadata:
  name: triggers_join_inputs
spec:
  steps:
    - name: join_node
      inputs:
        - triggers_join_inputs.inputs.INPUT1
        - triggers_join_inputs.inputs.INPUT2
      triggers:
        - triggers_join_inputs.inputs.TRIGGER1
        - triggers_join_inputs.inputs.TRIGGER2
      triggersJoinType: any
  output:
    steps:
      - join_node

```

```bash
request_string = get_request_string(use_trigger_1=True, use_trigger_2=True)

seldon pipeline infer triggers_join_inputs --inference-mode grpc '{request_string}'
```

```json
{"outputs":[{"name":"OUTPUT1","datatype":"INT64","shape":["1"],"contents":{"int64Contents":["2"]}}]}

```

```bash
request_string = get_request_string(use_trigger_1=True, use_trigger_2=False)

seldon pipeline infer triggers_join_inputs --inference-mode grpc '{request_string}'
```

```json
{"outputs":[{"name":"OUTPUT1","datatype":"INT64","shape":["1"],"contents":{"int64Contents":["2"]}}]}

```

```bash
request_string = get_request_string(use_trigger_1=False, use_trigger_2=True)

seldon pipeline infer triggers_join_inputs --inference-mode grpc '{request_string}'
```

```json
{"outputs":[{"name":"OUTPUT1","datatype":"INT64","shape":["1"],"contents":{"int64Contents":["2"]}}]}

```

```bash
seldon pipeline unload triggers_join_inputs
```

```json
{}

```

```bash
seldon pipeline load -f ./pipelines/triggers_join_internal.yaml
```

```json
{}

```

```bash
seldon pipeline status triggers_join_internal -w PipelineReady | jq .
```

```json
{
  "pipelineName": "triggers_join_internal",
  "versions": [
    {
      "pipeline": {
        "name": "triggers_join_internal",
        "uid": "cg5g8fs6dpcs73c4qhq0",
        "version": 1,
        "steps": [
          {
            "name": "id1_node",
            "inputs": [
              "triggers_join_internal.inputs.TRIGGER1"
            ],
            "tensorMap": {
              "triggers_join_internal.inputs.TRIGGER1": "INPUT1"
            }
          },
          {
            "name": "id2_node",
            "inputs": [
              "triggers_join_internal.inputs.TRIGGER2"
            ],
            "tensorMap": {
              "triggers_join_internal.inputs.TRIGGER2": "INPUT1"
            }
          },
          {
            "name": "join_node",
            "inputs": [
              "triggers_join_internal.inputs.INPUT1",
              "triggers_join_internal.inputs.INPUT2"
            ],
            "triggers": [
              "id1_node.outputs.OUTPUT1",
              "id2_node.outputs.OUTPUT1"
            ],
            "triggersJoin": "ANY"
          }
        ],
        "output": {
          "steps": [
            "join_node.outputs"
          ]
        },
        "kubernetesMeta": {}
      },
      "state": {
        "pipelineVersion": 1,
        "status": "PipelineReady",
        "reason": "created pipeline",
        "lastChangeTimestamp": "2023-03-10T10:19:43.307156600Z",
        "modelsReady": true
      }
    }
  ]
}

```

```bash
cat ./pipelines/triggers_join_internal.yaml
```

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Pipeline
metadata:
  name: triggers_join_internal
spec:
  steps:
    - name: id1_node
      inputs:
        - triggers_join_internal.inputs.TRIGGER1
      tensorMap:
        triggers_join_internal.inputs.TRIGGER1: INPUT1
    - name: id2_node
      inputs:
        - triggers_join_internal.inputs.TRIGGER2
      tensorMap:
        triggers_join_internal.inputs.TRIGGER2: INPUT1
    - name: join_node
      inputs:
        - triggers_join_internal.inputs.INPUT1
        - triggers_join_internal.inputs.INPUT2
      triggers:
        - id1_node.outputs.OUTPUT1
        - id2_node.outputs.OUTPUT1
      triggersJoinType: any
  output:
    steps:
      - join_node

```

```bash
request_string = get_request_string(use_trigger_1=True, use_trigger_2=True)

seldon pipeline infer triggers_join_internal --inference-mode grpc '{request_string}'
```

```json
{"outputs":[{"name":"OUTPUT1","datatype":"INT64","shape":["1"],"contents":{"int64Contents":["2"]}}]}

```

```bash
request_string = get_request_string(use_trigger_1=True, use_trigger_2=False)

seldon pipeline infer triggers_join_internal --inference-mode grpc '{request_string}'
```

```json
{"outputs":[{"name":"OUTPUT1","datatype":"INT64","shape":["1"],"contents":{"int64Contents":["2"]}}]}

```

```bash
request_string = get_request_string(use_trigger_1=False, use_trigger_2=True)

seldon pipeline infer triggers_join_internal --inference-mode grpc '{request_string}'
```

```json
{"outputs":[{"name":"OUTPUT1","datatype":"INT64","shape":["1"],"contents":{"int64Contents":["2"]}}]}

```

```bash
seldon pipeline unload triggers_join_internal
```

```json
{}

```

```bash
seldon model unload id1_node
seldon model unload id2_node
seldon model unload join_node
```

```json
{}
{}
{}

```

```python

```
