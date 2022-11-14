## Conditional Pipeline using PandasQuery



```python
!cat ../../models/choice1.yaml
!echo "---"
!cat ../../models/choice2.yaml
!echo "---"
!cat ../../models/add10.yaml
!echo "---"
!cat ../../models/mul10.yaml
```

    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: choice-is-one
    spec:
      storageUri: "gs://seldon-models/scv2/examples/pandasquery"
      requirements:
      - mlserver
      - python
      parameters:
      - name: query
        value: "choice == 1"
    ---
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: choice-is-two
    spec:
      storageUri: "gs://seldon-models/scv2/examples/pandasquery"
      requirements:
      - mlserver
      - python
      parameters:
      - name: query
        value: "choice == 2"
    ---
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: add10
    spec:
      storageUri: "gs://seldon-models/triton/add10"
      requirements:
      - triton
      - python
    ---
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
!seldon model load -f ../../models/choice1.yaml
!seldon model load -f ../../models/choice2.yaml
!seldon model load -f ../../models/add10.yaml
!seldon model load -f ../../models/mul10.yaml
```

    {}
    {}
    {}
    {}



```python
!seldon model status choice-is-one -w ModelAvailable 
!seldon model status choice-is-two -w ModelAvailable 
!seldon model status add10 -w ModelAvailable 
!seldon model status mul10 -w ModelAvailable 
```

    {}
    {}
    {}
    {}



```python
!cat ../../pipelines/choice.yaml
```

    apiVersion: mlops.seldon.io/v1alpha1
    kind: Pipeline
    metadata:
      name: choice
    spec:
      steps:
      - name: choice-is-one
      - name: mul10
        inputs:
        - choice.inputs.INPUT
        triggers:
        - choice-is-one.outputs.choice
      - name: choice-is-two
      - name: add10
        inputs:
        - choice.inputs.INPUT
        triggers:
        - choice-is-two.outputs.choice
      output:
        steps:
        - mul10
        - add10
        stepsJoin: any



```python
!seldon pipeline load -f ../../pipelines/choice.yaml
```

    {}



```python
!seldon pipeline status choice -w PipelineReady | jq -M .
```

    {
      "pipelineName": "choice",
      "versions": [
        {
          "pipeline": {
            "name": "choice",
            "uid": "cc67qd45em8of75v7phg",
            "version": 1,
            "steps": [
              {
                "name": "add10",
                "inputs": [
                  "choice.inputs.INPUT"
                ],
                "triggers": [
                  "choice-is-two.outputs.choice"
                ]
              },
              {
                "name": "choice-is-one"
              },
              {
                "name": "choice-is-two"
              },
              {
                "name": "mul10",
                "inputs": [
                  "choice.inputs.INPUT"
                ],
                "triggers": [
                  "choice-is-one.outputs.choice"
                ]
              }
            ],
            "output": {
              "steps": [
                "mul10.outputs",
                "add10.outputs"
              ],
              "stepsJoin": "ANY"
            },
            "kubernetesMeta": {}
          },
          "state": {
            "pipelineVersion": 1,
            "status": "PipelineReady",
            "reason": "created pipeline",
            "lastChangeTimestamp": "2022-08-29T08:47:48.496640155Z"
          }
        }
      ]
    }



```python
!seldon pipeline infer choice --inference-mode grpc \
 '{"model_name":"choice","inputs":[{"name":"choice","contents":{"int_contents":[1]},"datatype":"INT32","shape":[1]},{"name":"INPUT","contents":{"fp32_contents":[5,6,7,8]},"datatype":"FP32","shape":[4]}]}' | jq -M .
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
              50,
              60,
              70,
              80
            ]
          }
        }
      ],
      "rawOutputContents": [
        "AABIQgAAcEIAAIxCAACgQg=="
      ]
    }



```python
!seldon pipeline infer choice --inference-mode grpc \
 '{"model_name":"choice","inputs":[{"name":"choice","contents":{"int_contents":[2]},"datatype":"INT32","shape":[1]},{"name":"INPUT","contents":{"fp32_contents":[5,6,7,8]},"datatype":"FP32","shape":[4]}]}' | jq -M .
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
              15,
              16,
              17,
              18
            ]
          }
        }
      ],
      "rawOutputContents": [
        "AABwQQAAgEEAAIhBAACQQQ=="
      ]
    }



```python
!seldon model unload choice-is-one
!seldon model unload choice-is-two
!seldon model unload add10
!seldon model unload mul10
!seldon pipeline unload choice
```

    {}
    {}
    {}
    {}
    {}



```python

```
