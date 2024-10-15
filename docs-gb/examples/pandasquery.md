# Conditional Pipeline with Pandas Query Model

The model is defined as an MLServer custom runtime and allows the user to pass in a custom
pandas query as a parameter defined at model creation to be used to filter the data passed to the model.

```python
from mlserver import MLModel
from mlserver.types import InferenceRequest, InferenceResponse
from mlserver.codecs import PandasCodec
from mlserver.errors import MLServerError
import pandas as pd
from fastapi import status
from mlserver.logging import logger

QUERY_KEY = "query"


class ModelParametersMissing(MLServerError):
  def __init__(self, model_name: str, reason: str):
    super().__init__(
      f"Parameters missing for model {model_name} {reason}", status.HTTP_400_BAD_REQUEST
    )

class PandasQueryRuntime(MLModel):

  async def load(self) -> bool:
    logger.info("Loading with settings %s", self.settings)
    if self.settings.parameters is None or \
      self.settings.parameters.extra is None:
      raise ModelParametersMissing(self.name, "no settings.parameters.extra found")
    self.query = self.settings.parameters.extra[QUERY_KEY]
    if self.query is None:
      raise ModelParametersMissing(self.name, "no settings.parameters.extra.query found")
    self.ready = True

    return self.ready

  async def predict(self, payload: InferenceRequest) -> InferenceResponse:
    input_df: pd.DataFrame = PandasCodec.decode_request(payload)
    # run query on input_df and save in output_df
    output_df = input_df.query(self.query)
    if output_df.empty:
      output_df = pd.DataFrame({'status':["no rows satisfied " + self.query]})
    else:
      output_df["status"] = "row satisfied " + self.query
    return PandasCodec.encode_response(self.name, output_df, self.version)
```

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
