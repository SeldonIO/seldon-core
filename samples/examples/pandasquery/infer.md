## Conditional Pipeline using PandasQuery

```bash
cat ../../models/choice1.yaml
echo "---"
cat ../../models/choice2.yaml
echo "---"
cat ../../models/add10.yaml
echo "---"
cat ../../models/mul10.yaml
```

```yaml
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
  storageUri: "gs://seldon-models/scv2/samples/triton_23-03/add10"
  requirements:
  - triton
  - python
---
apiVersion: mlops.seldon.io/v1alpha1
kind: Model
metadata:
  name: mul10
spec:
  storageUri: "gs://seldon-models/scv2/samples/triton_23-03/mul10"
  requirements:
  - triton
  - python

```

```bash
seldon model load -f ../../models/choice1.yaml
seldon model load -f ../../models/choice2.yaml
seldon model load -f ../../models/add10.yaml
seldon model load -f ../../models/mul10.yaml
```

```json
{}
{}
{}
{}

```

```bash
seldon model status choice-is-one -w ModelAvailable
seldon model status choice-is-two -w ModelAvailable
seldon model status add10 -w ModelAvailable
seldon model status mul10 -w ModelAvailable
```

```json
{}
{}
{}
{}

```

```bash
cat ../../pipelines/choice.yaml
```

```yaml
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

```

```bash
seldon pipeline load -f ../../pipelines/choice.yaml
```

```bash
seldon pipeline status choice -w PipelineReady | jq -M .
```

```json
{
  "pipelineName": "choice",
  "versions": [
    {
      "pipeline": {
        "name": "choice",
        "uid": "cifel9aufmbc73e5intg",
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
        "lastChangeTimestamp": "2023-06-30T14:45:57.284684328Z",
        "modelsReady": true
      }
    }
  ]
}

```

```bash
seldon pipeline infer choice --inference-mode grpc \
 '{"model_name":"choice","inputs":[{"name":"choice","contents":{"int_contents":[1]},"datatype":"INT32","shape":[1]},{"name":"INPUT","contents":{"fp32_contents":[5,6,7,8]},"datatype":"FP32","shape":[4]}]}' | jq -M .
```

```json
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
  ]
}

```

```bash
seldon pipeline infer choice --inference-mode grpc \
 '{"model_name":"choice","inputs":[{"name":"choice","contents":{"int_contents":[2]},"datatype":"INT32","shape":[1]},{"name":"INPUT","contents":{"fp32_contents":[5,6,7,8]},"datatype":"FP32","shape":[4]}]}' | jq -M .
```

```json
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
  ]
}

```

```bash
seldon model unload choice-is-one
seldon model unload choice-is-two
seldon model unload add10
seldon model unload mul10
seldon pipeline unload choice
```

```python

```
