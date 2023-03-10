## Tabular Income Classifier Production Deployment

To run this notebook you need the inference data. This can be acquired in two ways:

  * Run `make train` or,
  * `gsutil cp -R gs://seldon-models/scv2/examples/income/infer-data .`

```python
import numpy as np
import json
import requests
```

```python
with open('./infer-data/test.npy', 'rb') as f:
    x_ref = np.load(f)
    x_h1 = np.load(f)
    y_ref = np.load(f)
    x_outlier = np.load(f)
```

```python
reqJson = json.loads('{"inputs":[{"name":"input_1","data":[],"datatype":"FP32","shape":[]}]}')
url = "http://0.0.0.0:9000/v2/models/model/infer"
```

```python
def infer(resourceName: str, batchSz: int, requestType: str):
    if requestType == "outlier":
        rows = x_outlier[0:0+batchSz]
    elif requestType == "drift":
        rows = x_h1[0:0+batchSz]
    else:
        rows = x_ref[0:0+batchSz]
    reqJson["inputs"][0]["data"] = rows.flatten().tolist()
    reqJson["inputs"][0]["shape"] = [batchSz, rows.shape[1]]
    headers = {"Content-Type": "application/json", "seldon-model":resourceName}
    response_raw = requests.post(url, json=reqJson, headers=headers)
    print(response_raw)
    print(response_raw.json())
```

### Pipeline with model, drift detector and outlier detector

```bash
cat ../../models/income-preprocess.yaml
echo "---"
cat ../../models/income.yaml
echo "---"
cat ../../models/income-drift.yaml
echo "---"
cat ../../models/income-outlier.yaml
```

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Model
metadata:
  name: income-preprocess
spec:
  storageUri: "gs://seldon-models/scv2/examples/mlserver_1.2.4/income/preprocessor"
  requirements:
  - sklearn
---
apiVersion: mlops.seldon.io/v1alpha1
kind: Model
metadata:
  name: income
spec:
  storageUri: "gs://seldon-models/scv2/examples/mlserver_1.2.4/income/classifier"
  requirements:
  - sklearn
---
apiVersion: mlops.seldon.io/v1alpha1
kind: Model
metadata:
  name: income-drift
spec:
  storageUri: "gs://seldon-models/scv2/examples/mlserver_1.2.4/income/drift-detector"
  requirements:
    - mlserver
    - alibi-detect
---
apiVersion: mlops.seldon.io/v1alpha1
kind: Model
metadata:
  name: income-outlier
spec:
  storageUri: "gs://seldon-models/scv2/examples/mlserver_1.2.4/income/outlier-detector"
  requirements:
    - mlserver
    - alibi-detect

```

```bash
seldon model load -f ../../models/income-preprocess.yaml
seldon model load -f ../../models/income.yaml
seldon model load -f ../../models/income-drift.yaml
seldon model load -f ../../models/income-outlier.yaml
```

```json
{}
{}
{}
{}

```

```bash
seldon model status income-preprocess -w ModelAvailable | jq .
seldon model status income -w ModelAvailable | jq .
seldon model status income-drift -w ModelAvailable | jq .
seldon model status income-outlier -w ModelAvailable | jq .
```

```json
{}
{}
{}
{}

```

```bash
cat ../../pipelines/income.yaml
```

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Pipeline
metadata:
  name: income-production
spec:
  steps:
    - name: income
    - name: income-preprocess
    - name: income-outlier
      inputs:
      - income-preprocess
    - name: income-drift
      batch:
        size: 20
  output:
    steps:
    - income
    - income-outlier.outputs.is_outlier

```

```bash
seldon pipeline load -f ../../pipelines/income.yaml
```

```json
{}

```

```bash
seldon pipeline status income-production -w PipelineReady| jq -M .
```

```json
{
  "pipelineName": "income-production",
  "versions": [
    {
      "pipeline": {
        "name": "income-production",
        "uid": "cg536k95aeqs73avlo9g",
        "version": 1,
        "steps": [
          {
            "name": "income"
          },
          {
            "name": "income-drift",
            "batch": {
              "size": 20
            }
          },
          {
            "name": "income-outlier",
            "inputs": [
              "income-preprocess.outputs"
            ]
          },
          {
            "name": "income-preprocess"
          }
        ],
        "output": {
          "steps": [
            "income.outputs",
            "income-outlier.outputs.is_outlier"
          ]
        },
        "kubernetesMeta": {}
      },
      "state": {
        "pipelineVersion": 1,
        "status": "PipelineReady",
        "reason": "created pipeline",
        "lastChangeTimestamp": "2023-03-09T19:28:17.908580977Z",
        "modelsReady": true
      }
    }
  ]
}

```

Show predictions from reference set. Should not be drift or outliers.

```python
batchSz=20
print(y_ref[0:batchSz])
infer("income-production.pipeline",batchSz,"normal")
```

```
[0 0 1 1 0 1 0 0 1 0 0 0 0 0 1 1 0 0 0 1]
<Response [200]>
{'model_name': '', 'outputs': [{'data': [0, 0, 1, 1, 0, 1, 0, 0, 1, 0, 0, 0, 0, 0, 1, 1, 0, 0, 0, 1], 'name': 'predict', 'shape': [20, 1], 'datatype': 'INT64'}, {'data': [0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0], 'name': 'is_outlier', 'shape': [1, 20], 'datatype': 'INT64'}]}

```

```bash
seldon pipeline inspect income-production.income-drift.outputs.is_drift
```

```
seldon.default.model.income-drift.outputs	cg536m0fh5ss73f8bd9g	{"name":"is_drift","datatype":"INT64","shape":["1","1"],"contents":{"int64Contents":["0"]}}

```

Show predictions from drift data. Should be drift and probably not outliers.

```python
batchSz=20
print(y_ref[0:batchSz])
infer("income-production.pipeline",batchSz,"drift")
```

```
[0 0 1 1 0 1 0 0 1 0 0 0 0 0 1 1 0 0 0 1]
<Response [200]>
{'model_name': '', 'outputs': [{'data': [0, 0, 0, 1, 1, 0, 1, 1, 1, 0, 0, 0, 0, 0, 1, 0, 0, 1, 0, 1], 'name': 'predict', 'shape': [20, 1], 'datatype': 'INT64'}, {'data': [0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0], 'name': 'is_outlier', 'shape': [1, 20], 'datatype': 'INT64'}]}

```

```bash
seldon pipeline inspect income-production.income-drift.outputs.is_drift
```

```
seldon.default.model.income-drift.outputs	cg536nofh5ss73f8bda0	{"name":"is_drift","datatype":"INT64","shape":["1","1"],"contents":{"int64Contents":["1"]}}

```

Show predictions from outlier data. Should be outliers and probably not drift.

```python
batchSz=20
print(y_ref[0:batchSz])
infer("income-production.pipeline",batchSz,"outlier")
```

```
[0 0 1 1 0 1 0 0 1 0 0 0 0 0 1 1 0 0 0 1]
<Response [200]>
{'model_name': '', 'outputs': [{'data': [0, 0, 1, 0, 0, 1, 0, 1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 0, 0, 1], 'name': 'predict', 'shape': [20, 1], 'datatype': 'INT64'}, {'data': [1, 1, 1, 1, 0, 1, 1, 1, 0, 1, 0, 1, 1, 1, 1, 1, 0, 1, 0, 1], 'name': 'is_outlier', 'shape': [1, 20], 'datatype': 'INT64'}]}

```

```bash
seldon pipeline inspect income-production.income-drift.outputs.is_drift
```

```
seldon.default.model.income-drift.outputs	cg536pofh5ss73f8bdag	{"name":"is_drift","datatype":"INT64","shape":["1","1"],"contents":{"int64Contents":["1"]}}

```

### Explanations

```bash
cat ../../models/income-explainer.yaml
```

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Model
metadata:
  name: income-explainer
spec:
  storageUri: "gs://seldon-models/scv2/examples/mlserver_1.2.4/income/explainer"
  explainer:
    type: anchor_tabular
    modelRef: income

```

```bash
seldon model load -f ../../models/income-explainer.yaml
```

```json
{}

```

```bash
seldon model status income-explainer -w ModelAvailable | jq .
```

```json
{}

```

```python
batchSz=1
print(y_ref[0:batchSz])
infer("income-explainer",batchSz,"normal")
```

```
[0]
<Response [200]>
{'model_name': 'income-explainer_1', 'model_version': '1', 'id': 'b47e6a65-2699-4839-8382-6867673da3e0', 'parameters': {}, 'outputs': [{'name': 'explanation', 'shape': [1, 1], 'datatype': 'BYTES', 'parameters': {'content_type': 'str'}, 'data': ['{"meta": {"name": "AnchorTabular", "type": ["blackbox"], "explanations": ["local"], "params": {"seed": 1, "disc_perc": [25, 50, 75], "threshold": 0.95, "delta": 0.1, "tau": 0.15, "batch_size": 100, "coverage_samples": 10000, "beam_size": 1, "stop_on_first": false, "max_anchor_size": null, "min_samples_start": 100, "n_covered_ex": 10, "binary_cache_size": 10000, "cache_margin": 1000, "verbose": false, "verbose_every": 1, "kwargs": {}}, "version": "0.9.0"}, "data": {"anchor": ["Marital Status = Never-Married", "Relationship = Own-child", "Capital Gain <= 0.00"], "precision": 0.9970501474926253, "coverage": 0.06853582554517133, "raw": {"feature": [3, 5, 8], "mean": [0.8027586206896552, 0.9420849420849421, 0.9970501474926253], "precision": [0.8027586206896552, 0.9420849420849421, 0.9970501474926253], "coverage": [0.3037383177570093, 0.07165109034267912, 0.06853582554517133], "examples": [{"covered_true": [[34, 4, 1, 1, 7, 0, 4, 1, 0, 0, 30, 9], [38, 4, 1, 1, 2, 0, 4, 1, 4508, 0, 40, 9], [25, 0, 1, 1, 0, 1, 4, 1, 0, 0, 40, 1], [35, 6, 1, 1, 5, 0, 4, 1, 0, 0, 40, 9], [27, 4, 1, 1, 2, 3, 2, 1, 0, 0, 40, 9], [46, 2, 5, 1, 5, 4, 4, 0, 0, 0, 43, 9], [42, 4, 1, 1, 6, 0, 4, 1, 0, 0, 60, 9], [54, 6, 1, 1, 8, 0, 4, 1, 0, 0, 40, 9], [27, 4, 5, 1, 5, 1, 4, 0, 0, 1504, 45, 9], [36, 7, 5, 1, 5, 1, 4, 0, 0, 0, 40, 9]], "covered_false": [[56, 5, 5, 1, 8, 1, 4, 1, 27828, 0, 60, 9], [39, 5, 1, 1, 8, 0, 4, 1, 7298, 0, 40, 9], [46, 4, 1, 1, 8, 0, 4, 1, 0, 0, 50, 9], [35, 4, 5, 1, 6, 0, 4, 1, 0, 0, 55, 9], [39, 6, 1, 1, 5, 0, 4, 1, 15024, 0, 50, 6], [33, 4, 5, 1, 5, 0, 4, 1, 15024, 0, 40, 9], [49, 4, 5, 1, 5, 1, 4, 1, 0, 0, 40, 9], [45, 4, 2, 1, 5, 4, 4, 1, 15020, 0, 40, 6], [50, 4, 5, 1, 8, 0, 4, 1, 0, 0, 50, 9], [67, 5, 5, 1, 6, 0, 4, 1, 0, 0, 48, 9]], "uncovered_true": [], "uncovered_false": []}, {"covered_true": [[54, 4, 1, 1, 7, 3, 4, 1, 0, 0, 40, 9], [45, 4, 1, 1, 2, 3, 1, 1, 0, 0, 40, 6], [22, 4, 1, 1, 6, 3, 4, 1, 0, 0, 45, 9], [33, 4, 1, 1, 8, 3, 4, 0, 0, 0, 40, 9], [36, 1, 5, 1, 5, 3, 4, 1, 0, 0, 45, 9], [55, 6, 1, 1, 5, 3, 4, 0, 0, 0, 8, 9], [53, 7, 5, 1, 5, 3, 4, 0, 0, 0, 40, 9], [23, 4, 1, 1, 8, 3, 4, 0, 0, 0, 40, 9], [23, 4, 1, 1, 2, 3, 4, 1, 0, 0, 40, 9], [38, 1, 5, 1, 5, 3, 4, 1, 0, 0, 40, 6]], "covered_false": [[51, 4, 5, 1, 8, 3, 4, 1, 15024, 0, 60, 9], [78, 6, 1, 1, 8, 3, 4, 1, 99999, 0, 20, 9], [57, 4, 5, 1, 5, 3, 4, 1, 10520, 0, 32, 9], [50, 4, 5, 1, 8, 3, 4, 1, 99999, 0, 60, 9], [32, 4, 1, 1, 5, 3, 4, 1, 99999, 0, 50, 9]], "uncovered_true": [], "uncovered_false": []}, {"covered_true": [[25, 4, 1, 1, 1, 3, 4, 1, 0, 0, 40, 9], [38, 2, 1, 1, 4, 3, 4, 0, 0, 0, 43, 9], [42, 4, 2, 1, 5, 3, 4, 1, 0, 0, 50, 9], [26, 4, 1, 1, 8, 3, 4, 1, 0, 0, 40, 9], [67, 4, 1, 1, 5, 3, 4, 1, 0, 0, 2, 9], [31, 4, 1, 1, 4, 3, 4, 1, 0, 0, 37, 9], [44, 4, 1, 1, 1, 3, 4, 0, 0, 0, 40, 6], [24, 4, 1, 1, 8, 3, 4, 0, 0, 0, 40, 9], [38, 2, 5, 1, 8, 3, 4, 1, 0, 0, 70, 9], [61, 4, 1, 1, 5, 3, 4, 1, 0, 0, 40, 9]], "covered_false": [], "uncovered_true": [], "uncovered_false": []}], "all_precision": 0, "num_preds": 1000000, "success": true, "names": ["Marital Status = Never-Married", "Relationship = Own-child", "Capital Gain <= 0.00"], "prediction": [0], "instance": [47.0, 4.0, 1.0, 1.0, 1.0, 3.0, 4.0, 1.0, 0.0, 0.0, 40.0, 9.0], "instances": [[47.0, 4.0, 1.0, 1.0, 1.0, 3.0, 4.0, 1.0, 0.0, 0.0, 40.0, 9.0]]}}}']}]}

```

### Cleanup

```bash
seldon pipeline unload income-production
seldon model unload income-preprocess
seldon model unload income
seldon model unload income-drift
seldon model unload income-outlier
seldon model unload income-explainer
```

```json
{}
{}
{}
{}
{}
{}

```

```python

```
