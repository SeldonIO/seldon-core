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


```python
!seldon model load -f ./models/id1_node.yaml
!seldon model load -f ./models/id2_node.yaml
!seldon model load -f ./models/join_node.yaml
```

    {}
    {}
    {}



```python
!seldon model status join_node -w ModelAvailable
!seldon model status id1_node -w ModelAvailable
!seldon model status id2_node -w ModelAvailable
```

    {}
    {}
    {}



```python
!seldon pipeline load -f ./pipelines/triggers_join_inputs.yaml
```

    {}



```python
!seldon pipeline status triggers_join_inputs -w PipelineReady | jq .
```

    [1;39m{
      [0m[34;1m"pipelineName"[0m[1;39m: [0m[0;32m"triggers_join_inputs"[0m[1;39m,
      [0m[34;1m"versions"[0m[1;39m: [0m[1;39m[
        [1;39m{
          [0m[34;1m"pipeline"[0m[1;39m: [0m[1;39m{
            [0m[34;1m"name"[0m[1;39m: [0m[0;32m"triggers_join_inputs"[0m[1;39m,
            [0m[34;1m"uid"[0m[1;39m: [0m[0;32m"cdqjkdhqa12c739ab3t0"[0m[1;39m,
            [0m[34;1m"version"[0m[1;39m: [0m[0;39m1[0m[1;39m,
            [0m[34;1m"steps"[0m[1;39m: [0m[1;39m[
              [1;39m{
                [0m[34;1m"name"[0m[1;39m: [0m[0;32m"join_node"[0m[1;39m,
                [0m[34;1m"inputs"[0m[1;39m: [0m[1;39m[
                  [0;32m"triggers_join_inputs.inputs.INPUT1"[0m[1;39m,
                  [0;32m"triggers_join_inputs.inputs.INPUT2"[0m[1;39m
                [1;39m][0m[1;39m,
                [0m[34;1m"triggers"[0m[1;39m: [0m[1;39m[
                  [0;32m"triggers_join_inputs.inputs.TRIGGER1"[0m[1;39m,
                  [0;32m"triggers_join_inputs.inputs.TRIGGER2"[0m[1;39m
                [1;39m][0m[1;39m,
                [0m[34;1m"triggersJoin"[0m[1;39m: [0m[0;32m"ANY"[0m[1;39m
              [1;39m}[0m[1;39m
            [1;39m][0m[1;39m,
            [0m[34;1m"output"[0m[1;39m: [0m[1;39m{
              [0m[34;1m"steps"[0m[1;39m: [0m[1;39m[
                [0;32m"join_node.outputs"[0m[1;39m
              [1;39m][0m[1;39m
            [1;39m}[0m[1;39m,
            [0m[34;1m"kubernetesMeta"[0m[1;39m: [0m[1;39m{}[0m[1;39m
          [1;39m}[0m[1;39m,
          [0m[34;1m"state"[0m[1;39m: [0m[1;39m{
            [0m[34;1m"pipelineVersion"[0m[1;39m: [0m[0;39m1[0m[1;39m,
            [0m[34;1m"status"[0m[1;39m: [0m[0;32m"PipelineReady"[0m[1;39m,
            [0m[34;1m"reason"[0m[1;39m: [0m[0;32m"created pipeline"[0m[1;39m,
            [0m[34;1m"lastChangeTimestamp"[0m[1;39m: [0m[0;32m"2022-11-16T19:29:58.778539435Z"[0m[1;39m
          [1;39m}[0m[1;39m
        [1;39m}[0m[1;39m
      [1;39m][0m[1;39m
    [1;39m}[0m



```python
!cat ./pipelines/triggers_join_inputs.yaml
```

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



```python
request_string = get_request_string(use_trigger_1=True, use_trigger_2=True)

!seldon pipeline infer triggers_join_inputs --inference-mode grpc '{request_string}'
```

    {"outputs":[{"name":"OUTPUT1", "datatype":"INT64", "shape":["1"], "contents":{"int64Contents":["2"]}}]}



```python
request_string = get_request_string(use_trigger_1=True, use_trigger_2=False)

!seldon pipeline infer triggers_join_inputs --inference-mode grpc '{request_string}'
```

    {"outputs":[{"name":"OUTPUT1", "datatype":"INT64", "shape":["1"], "contents":{"int64Contents":["2"]}}]}



```python
request_string = get_request_string(use_trigger_1=False, use_trigger_2=True)

!seldon pipeline infer triggers_join_inputs --inference-mode grpc '{request_string}'
```

    {"outputs":[{"name":"OUTPUT1", "datatype":"INT64", "shape":["1"], "contents":{"int64Contents":["2"]}}]}



```python
!seldon pipeline unload triggers_join_inputs
```

    {}



```python
!seldon pipeline load -f ./pipelines/triggers_join_internal.yaml
```

    {}



```python
!seldon pipeline status triggers_join_internal -w PipelineReady | jq .
```

    [1;39m{
      [0m[34;1m"pipelineName"[0m[1;39m: [0m[0;32m"triggers_join_internal"[0m[1;39m,
      [0m[34;1m"versions"[0m[1;39m: [0m[1;39m[
        [1;39m{
          [0m[34;1m"pipeline"[0m[1;39m: [0m[1;39m{
            [0m[34;1m"name"[0m[1;39m: [0m[0;32m"triggers_join_internal"[0m[1;39m,
            [0m[34;1m"uid"[0m[1;39m: [0m[0;32m"cdqjkqpqa12c739ab3tg"[0m[1;39m,
            [0m[34;1m"version"[0m[1;39m: [0m[0;39m1[0m[1;39m,
            [0m[34;1m"steps"[0m[1;39m: [0m[1;39m[
              [1;39m{
                [0m[34;1m"name"[0m[1;39m: [0m[0;32m"id1_node"[0m[1;39m,
                [0m[34;1m"inputs"[0m[1;39m: [0m[1;39m[
                  [0;32m"triggers_join_internal.inputs.TRIGGER1"[0m[1;39m
                [1;39m][0m[1;39m,
                [0m[34;1m"tensorMap"[0m[1;39m: [0m[1;39m{
                  [0m[34;1m"triggers_join_internal.inputs.TRIGGER1"[0m[1;39m: [0m[0;32m"INPUT1"[0m[1;39m
                [1;39m}[0m[1;39m
              [1;39m}[0m[1;39m,
              [1;39m{
                [0m[34;1m"name"[0m[1;39m: [0m[0;32m"id2_node"[0m[1;39m,
                [0m[34;1m"inputs"[0m[1;39m: [0m[1;39m[
                  [0;32m"triggers_join_internal.inputs.TRIGGER2"[0m[1;39m
                [1;39m][0m[1;39m,
                [0m[34;1m"tensorMap"[0m[1;39m: [0m[1;39m{
                  [0m[34;1m"triggers_join_internal.inputs.TRIGGER2"[0m[1;39m: [0m[0;32m"INPUT1"[0m[1;39m
                [1;39m}[0m[1;39m
              [1;39m}[0m[1;39m,
              [1;39m{
                [0m[34;1m"name"[0m[1;39m: [0m[0;32m"join_node"[0m[1;39m,
                [0m[34;1m"inputs"[0m[1;39m: [0m[1;39m[
                  [0;32m"triggers_join_internal.inputs.INPUT1"[0m[1;39m,
                  [0;32m"triggers_join_internal.inputs.INPUT2"[0m[1;39m
                [1;39m][0m[1;39m,
                [0m[34;1m"triggers"[0m[1;39m: [0m[1;39m[
                  [0;32m"id1_node.outputs.OUTPUT1"[0m[1;39m,
                  [0;32m"id2_node.outputs.OUTPUT1"[0m[1;39m
                [1;39m][0m[1;39m,
                [0m[34;1m"triggersJoin"[0m[1;39m: [0m[0;32m"ANY"[0m[1;39m
              [1;39m}[0m[1;39m
            [1;39m][0m[1;39m,
            [0m[34;1m"output"[0m[1;39m: [0m[1;39m{
              [0m[34;1m"steps"[0m[1;39m: [0m[1;39m[
                [0;32m"join_node.outputs"[0m[1;39m
              [1;39m][0m[1;39m
            [1;39m}[0m[1;39m,
            [0m[34;1m"kubernetesMeta"[0m[1;39m: [0m[1;39m{}[0m[1;39m
          [1;39m}[0m[1;39m,
          [0m[34;1m"state"[0m[1;39m: [0m[1;39m{
            [0m[34;1m"pipelineVersion"[0m[1;39m: [0m[0;39m1[0m[1;39m,
            [0m[34;1m"status"[0m[1;39m: [0m[0;32m"PipelineReady"[0m[1;39m,
            [0m[34;1m"reason"[0m[1;39m: [0m[0;32m"created pipeline"[0m[1;39m,
            [0m[34;1m"lastChangeTimestamp"[0m[1;39m: [0m[0;32m"2022-11-16T19:30:51.996346140Z"[0m[1;39m
          [1;39m}[0m[1;39m
        [1;39m}[0m[1;39m
      [1;39m][0m[1;39m
    [1;39m}[0m



```python
!cat ./pipelines/triggers_join_internal.yaml
```

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



```python
request_string = get_request_string(use_trigger_1=True, use_trigger_2=True)

!seldon pipeline infer triggers_join_internal --inference-mode grpc '{request_string}'
```

    {"outputs":[{"name":"OUTPUT1", "datatype":"INT64", "shape":["1"], "contents":{"int64Contents":["2"]}}]}



```python
request_string = get_request_string(use_trigger_1=True, use_trigger_2=False)

!seldon pipeline infer triggers_join_internal --inference-mode grpc '{request_string}'
```

    {"outputs":[{"name":"OUTPUT1", "datatype":"INT64", "shape":["1"], "contents":{"int64Contents":["2"]}}]}



```python
request_string = get_request_string(use_trigger_1=False, use_trigger_2=True)

!seldon pipeline infer triggers_join_internal --inference-mode grpc '{request_string}'
```

    {"outputs":[{"name":"OUTPUT1", "datatype":"INT64", "shape":["1"], "contents":{"int64Contents":["2"]}}]}



```python
!seldon pipeline unload triggers_join_internal
```

    {}



```python
!seldon model unload id1_node
!seldon model unload id2_node
!seldon model unload join_node
```

    {}
    {}
    {}



```python

```
