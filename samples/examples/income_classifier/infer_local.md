## Tabular Income Classifier Production Deployment

To run this notebook you need the inference data. This can be acquired in two ways:

  * Run train.ipynb
  * Start Seldon with `export LOCAL_MODEL_FOLDER=<this folder>`

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
seldon model load -f local_resources/income-preprocess.yaml
seldon model load -f local_resources/income.yaml
seldon model load -f local_resources/income-drift.yaml
seldon model load -f local_resources/income-outlier.yaml
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
        "uid": "cc536tg4sl9u21cllbr0",
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
        "lastChangeTimestamp": "2022-08-27T15:28:52.890372305Z"
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
{'model_name': '', 'outputs': [{'data': [0, 0, 1, 1, 0, 1, 0, 0, 1, 0, 0, 0, 0, 0, 1, 1, 0, 0, 0, 1], 'name': 'predict', 'shape': [20], 'datatype': 'INT64'}, {'data': [0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0], 'name': 'is_outlier', 'shape': [1, 20], 'datatype': 'INT64'}]}

```

```bash
seldon pipeline inspect income-production.income-drift.outputs.is_drift
```

```
---
seldon.default.model.income-drift.outputs
cc53o9qqojmgf7c4m4bg:{"name":"is_drift","datatype":"INT64","shape":["1"],"contents":{"int64Contents":["0"]}}

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
{'model_name': '', 'outputs': [{'data': [0, 0, 0, 1, 1, 0, 1, 1, 1, 0, 0, 0, 0, 0, 1, 0, 0, 1, 0, 1], 'name': 'predict', 'shape': [20], 'datatype': 'INT64'}, {'data': [0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0], 'name': 'is_outlier', 'shape': [1, 20], 'datatype': 'INT64'}]}

```

```bash
seldon pipeline inspect income-production.income-drift.outputs.is_drift
```

```
---
seldon.default.model.income-drift.outputs
cc53oaiqojmgf7c4m4c0:{"name":"is_drift","datatype":"INT64","shape":["1"],"contents":{"int64Contents":["1"]}}

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
{'model_name': '', 'outputs': [{'data': [0, 0, 1, 1, 0, 0, 0, 1, 1, 1, 1, 0, 0, 0, 1, 1, 0, 0, 0, 1], 'name': 'predict', 'shape': [20], 'datatype': 'INT64'}, {'data': [1, 0, 0, 0, 1, 1, 1, 1, 1, 1, 1, 0, 0, 0, 1, 1, 1, 0, 1, 1], 'name': 'is_outlier', 'shape': [1, 20], 'datatype': 'INT64'}]}

```

```bash
seldon pipeline inspect income-production.income-drift.outputs.is_drift
```

```
---
seldon.default.model.income-drift.outputs
cc53obiqojmgf7c4m4cg:{"name":"is_drift","datatype":"INT64","shape":["1"],"contents":{"int64Contents":["0"]}}

```

### Explanations

```bash
cat local_resources/income-explainer.yaml
```

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Model
metadata:
  name: income-explainer
spec:
  storageUri: "/mnt/models/explainer"
  explainer:
    type: anchor_tabular
    modelRef: income

```

```bash
seldon model load -f local_resources/income-explainer.yaml
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
{'model_name': 'income-explainer_1', 'model_version': '1', 'id': '2cba4255-d972-468f-a9e0-5b0212bcd10f', 'parameters': {'content_type': None, 'headers': None}, 'outputs': [{'name': 'explanation', 'shape': [1], 'datatype': 'BYTES', 'parameters': {'content_type': 'str', 'headers': None}, 'data': ['{"meta": {"name": "AnchorTabular", "type": ["blackbox"], "explanations": ["local"], "params": {"seed": 1, "disc_perc": [25, 50, 75], "threshold": 0.95, "delta": 0.1, "tau": 0.15, "batch_size": 100, "coverage_samples": 10000, "beam_size": 1, "stop_on_first": false, "max_anchor_size": null, "min_samples_start": 100, "n_covered_ex": 10, "binary_cache_size": 10000, "cache_margin": 1000, "verbose": false, "verbose_every": 1, "kwargs": {}}, "version": "0.7.0"}, "data": {"anchor": ["Marital Status = Never-Married", "Relationship = Own-child", "Capital Gain <= 0.00", "Capital Loss <= 0.00"], "precision": 1.0, "coverage": 0.06720071206052515, "raw": {"feature": [3, 5, 8, 9], "mean": [0.7980769230769231, 0.9152542372881356, 0.9972144846796658, 1.0], "precision": [0.7980769230769231, 0.9152542372881356, 0.9972144846796658, 1.0], "coverage": [0.3037383177570093, 0.07165109034267912, 0.06853582554517133, 0.06720071206052515], "examples": [{"covered_true": [[23, 7, 1, 1, 5, 1, 4, 0, 0, 0, 32, 9], [44, 4, 1, 1, 8, 0, 4, 1, 0, 0, 40, 9], [39, 7, 1, 1, 8, 1, 4, 0, 0, 0, 40, 9], [25, 4, 5, 1, 5, 1, 4, 0, 0, 0, 40, 9], [44, 2, 1, 1, 5, 0, 4, 1, 0, 0, 40, 9], [49, 6, 1, 1, 5, 0, 4, 1, 0, 0, 40, 9], [31, 4, 1, 1, 8, 1, 4, 0, 0, 0, 40, 9], [26, 4, 5, 1, 8, 3, 4, 0, 0, 0, 40, 9], [31, 4, 1, 1, 4, 1, 4, 1, 0, 0, 45, 9], [56, 4, 1, 1, 5, 0, 4, 1, 0, 1902, 65, 9]], "covered_false": [[50, 4, 5, 1, 8, 0, 4, 1, 15024, 0, 65, 9], [43, 4, 5, 1, 5, 4, 4, 0, 0, 2547, 40, 9], [62, 6, 1, 1, 2, 0, 4, 1, 0, 0, 55, 9], [29, 4, 1, 1, 8, 1, 4, 0, 0, 2258, 45, 9], [45, 7, 1, 1, 8, 0, 4, 1, 0, 1977, 60, 9], [67, 6, 2, 1, 5, 0, 4, 1, 10605, 0, 35, 9], [38, 4, 1, 1, 5, 0, 4, 1, 0, 0, 45, 9], [38, 4, 5, 1, 8, 0, 4, 1, 0, 1977, 60, 9], [34, 5, 1, 1, 5, 1, 4, 0, 0, 0, 62, 9], [63, 0, 1, 1, 0, 0, 4, 1, 7688, 0, 54, 9]], "uncovered_true": [], "uncovered_false": []}, {"covered_true": [[35, 1, 2, 1, 5, 3, 0, 0, 0, 0, 60, 9], [37, 4, 5, 1, 5, 3, 1, 1, 0, 0, 40, 4], [24, 4, 1, 1, 8, 3, 4, 1, 0, 0, 40, 9], [26, 4, 5, 1, 5, 3, 4, 0, 0, 0, 40, 9], [25, 4, 1, 1, 5, 3, 4, 0, 0, 0, 15, 9], [43, 4, 1, 1, 5, 3, 4, 1, 0, 0, 40, 9], [45, 4, 1, 1, 8, 3, 4, 1, 0, 0, 60, 9], [40, 7, 5, 1, 5, 3, 4, 0, 0, 0, 45, 9], [34, 4, 1, 1, 8, 3, 1, 0, 0, 0, 40, 7], [53, 4, 1, 1, 7, 3, 4, 1, 0, 0, 40, 1]], "covered_false": [[32, 4, 1, 1, 6, 3, 2, 1, 15024, 0, 50, 9], [63, 0, 1, 1, 0, 3, 4, 1, 7688, 0, 54, 9], [52, 5, 1, 1, 6, 3, 4, 1, 15024, 0, 50, 9], [41, 4, 1, 1, 6, 3, 4, 1, 7688, 0, 50, 9], [40, 6, 1, 1, 5, 3, 4, 1, 7298, 0, 50, 9], [32, 4, 1, 1, 8, 3, 4, 1, 0, 2444, 50, 9], [46, 2, 1, 1, 5, 3, 4, 1, 4787, 0, 45, 9], [36, 4, 1, 1, 8, 3, 4, 1, 8614, 0, 40, 9], [59, 5, 5, 1, 8, 3, 4, 1, 15024, 0, 80, 9], [50, 5, 2, 1, 5, 3, 4, 1, 15024, 0, 60, 9]], "uncovered_true": [], "uncovered_false": []}, {"covered_true": [[61, 4, 1, 1, 1, 3, 4, 1, 0, 0, 40, 9], [35, 4, 1, 1, 6, 3, 4, 1, 0, 0, 40, 9], [66, 6, 1, 1, 6, 3, 4, 1, 0, 0, 40, 9], [54, 7, 1, 1, 8, 3, 4, 1, 0, 0, 38, 9], [32, 2, 1, 1, 5, 3, 4, 0, 0, 0, 45, 9], [38, 4, 1, 1, 6, 3, 4, 1, 0, 0, 60, 9], [57, 4, 1, 1, 8, 3, 4, 1, 0, 0, 40, 9], [67, 0, 1, 1, 0, 3, 4, 1, 0, 0, 60, 9], [36, 4, 5, 1, 8, 3, 4, 1, 0, 0, 45, 9], [25, 2, 1, 1, 5, 3, 4, 1, 0, 0, 45, 9]], "covered_false": [], "uncovered_true": [], "uncovered_false": []}, {"covered_true": [[26, 2, 5, 1, 5, 3, 4, 1, 0, 0, 50, 9], [40, 4, 1, 1, 8, 3, 4, 1, 0, 0, 75, 9], [36, 7, 1, 1, 4, 3, 2, 0, 0, 0, 40, 9], [53, 4, 1, 1, 6, 3, 4, 1, 0, 0, 50, 9], [36, 4, 1, 1, 1, 3, 4, 0, 0, 0, 40, 9], [45, 5, 1, 1, 8, 3, 4, 1, 0, 0, 40, 9], [70, 4, 1, 1, 1, 3, 4, 1, 0, 0, 40, 9], [49, 2, 1, 1, 4, 3, 2, 1, 0, 0, 40, 9], [23, 4, 1, 1, 6, 3, 4, 0, 0, 0, 40, 9], [28, 4, 1, 1, 7, 3, 2, 1, 0, 0, 40, 9]], "covered_false": [[34, 4, 1, 1, 6, 3, 4, 1, 0, 0, 40, 9]], "uncovered_true": [], "uncovered_false": []}], "all_precision": 0, "num_preds": 1000000, "success": true, "names": ["Marital Status = Never-Married", "Relationship = Own-child", "Capital Gain <= 0.00", "Capital Loss <= 0.00"], "prediction": [0], "instance": [47.0, 4.0, 1.0, 1.0, 1.0, 3.0, 4.0, 1.0, 0.0, 0.0, 40.0, 9.0], "instances": [[47.0, 4.0, 1.0, 1.0, 1.0, 3.0, 4.0, 1.0, 0.0, 0.0, 40.0, 9.0]]}}}']}]}

```

### Cleanup

```bash
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

```

```python

```
