## Tabular Income Classifier Production Deployment

 * Optionally run the train.ipynb to show how models are created


```python
import numpy as np
import json
import requests
```


```python
with open('test.npy', 'rb') as f:
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
````{collapse} Expand to see output
```yaml
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: income-preprocess
      namespace: seldon-mesh
    spec:
      storageUri: "/mnt/models/preprocessor"
      requirements:
      - sklearn
    ---
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: income
      namespace: seldon-mesh
    spec:
      storageUri: "/mnt/models/classifier"
      requirements:
      - sklearn
    ---
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: income-drift
      namespace: seldon-mesh
    spec:
      storageUri: "/mnt/models/drift-detector"
      requirements:
        - mlserver
        - alibi-detect
    ---
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: income-outlier
      namespace: seldon-mesh
    spec:
      storageUri: "/mnt/models/outlier-detector"
      requirements:
        - mlserver
        - alibi-detect
```
````

```bash
seldon model load -f ../../models/income-preprocess.yaml
seldon model load -f ../../models/income.yaml
seldon model load -f ../../models/income-drift.yaml
seldon model load -f ../../models/income-outlier.yaml
```
````{collapse} Expand to see output
```json

    {}
    {}
    {}
    {}
```
````

```bash
seldon model status income-preprocess -w ModelAvailable | jq .
seldon model status income -w ModelAvailable | jq .
seldon model status income-drift -w ModelAvailable | jq .
seldon model status income-outlier -w ModelAvailable | jq .
```
````{collapse} Expand to see output
```json

    [1;39m{}[0m
    [1;39m{}[0m
    [1;39m{}[0m
    [1;39m{}[0m
```
````

```bash
cat ../../pipelines/income.yaml
```
````{collapse} Expand to see output
```yaml
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Pipeline
    metadata:
      name: income-production
      namespace: seldon-mesh
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
````

```bash
seldon pipeline load -f ../../pipelines/income.yaml
```
````{collapse} Expand to see output
```json

    {}
```
````

```bash
seldon pipeline status income-production -w PipelineReady| jq -M .
```
````{collapse} Expand to see output
```json

    {
      "pipelineName": "income-production",
      "versions": [
        {
          "pipeline": {
            "name": "income-production",
            "uid": "cbimdu958n94aofi9tjg",
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
            "kubernetesMeta": {
              "namespace": "seldon-mesh"
            }
          },
          "state": {
            "pipelineVersion": 1,
            "status": "PipelineReady",
            "reason": "Created pipeline",
            "lastChangeTimestamp": "2022-07-30T17:14:35.712344628Z"
          }
        }
      ]
    }
```
````
Show predictions from reference set. Should not be drift or outliers.


```python
batchSz=20
print(y_ref[0:batchSz])
infer("income-production.pipeline",batchSz,"normal")
```

    [0 0 1 1 0 1 0 0 1 0 0 0 0 0 1 1 0 0 0 1]
    <Response [200]>
    {'model_name': '', 'outputs': [{'data': [0, 0, 1, 1, 0, 1, 0, 0, 1, 0, 0, 0, 0, 0, 1, 1, 0, 0, 0, 1], 'name': 'predict', 'shape': [20], 'datatype': 'INT64'}, {'data': [0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0], 'name': 'is_outlier', 'shape': [1, 20], 'datatype': 'INT64'}]}
```
````

```bash
seldon pipeline inspect income-production.income-drift.outputs.is_drift
```
````{collapse} Expand to see output
```json

    ---
    seldon.default.model.income-drift.outputs
    cbime95vqj3pei2v1gcg:{"name":"is_drift","datatype":"INT64","shape":["1"],"contents":{"int64Contents":["0"]}}
```
````
Show predictions from drift data. Should be drift and probably not outliers.


```python
batchSz=20
print(y_ref[0:batchSz])
infer("income-production.pipeline",batchSz,"drift")
```

    [0 0 1 1 0 1 0 0 1 0 0 0 0 0 1 1 0 0 0 1]
    <Response [200]>
    {'model_name': '', 'outputs': [{'data': [0, 0, 0, 1, 1, 0, 1, 1, 1, 0, 0, 0, 0, 0, 1, 0, 0, 1, 0, 1], 'name': 'predict', 'shape': [20], 'datatype': 'INT64'}, {'data': [0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0], 'name': 'is_outlier', 'shape': [1, 20], 'datatype': 'INT64'}]}
```
````

```bash
seldon pipeline inspect income-production.income-drift.outputs.is_drift
```
````{collapse} Expand to see output
```json

    ---
    seldon.default.model.income-drift.outputs
    cbimeptvqj3pei2v1gd0:{"name":"is_drift","datatype":"INT64","shape":["1"],"contents":{"int64Contents":["1"]}}
```
````
Show predictions from outlier data. Should be outliers and probably not drift.


```python
batchSz=20
print(y_ref[0:batchSz])
infer("income-production.pipeline",batchSz,"outlier")
```

    [0 0 1 1 0 1 0 0 1 0 0 0 0 0 1 1 0 0 0 1]
    <Response [200]>
    {'model_name': '', 'outputs': [{'data': [0, 0, 1, 1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 1, 1, 0, 0, 0, 1], 'name': 'predict', 'shape': [20], 'datatype': 'INT64'}, {'data': [1, 1, 1, 0, 1, 1, 1, 1, 1, 0, 1, 1, 0, 1, 1, 1, 1, 1, 1, 1], 'name': 'is_outlier', 'shape': [1, 20], 'datatype': 'INT64'}]}
```
````

```bash
seldon pipeline inspect income-production.income-drift.outputs.is_drift
```
````{collapse} Expand to see output
```json

    ---
    seldon.default.model.income-drift.outputs
    cbimm05vqj3pei2v1gdg:{"name":"is_drift","datatype":"INT64","shape":["1"],"contents":{"int64Contents":["0"]}}
```
````
### Explanations


```bash
cat ../../models/income-explainer.yaml
```
````{collapse} Expand to see output
```yaml
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: income-explainer
      namespace: seldon-mesh
    spec:
      storageUri: "/mnt/models/explainer"
      explainer:
        type: anchor_tabular
        modelRef: income
```
````

```bash
seldon model load -f ../../models/income-explainer.yaml
```
````{collapse} Expand to see output
```json

    {}
```
````

```bash
seldon model status income-explainer -w ModelAvailable | jq .
```
````{collapse} Expand to see output
```json

    [1;39m{}[0m
```
````

```python
batchSz=1
print(y_ref[0:batchSz])
infer("income-explainer",batchSz,"normal")
```

    [0]
    <Response [200]>
    {'model_name': 'income-explainer_1', 'model_version': '1', 'id': '577e5c5a-6c15-44d8-93e6-f9c722071ff2', 'parameters': {'content_type': None, 'headers': None}, 'outputs': [{'name': 'explanation', 'shape': [1], 'datatype': 'BYTES', 'parameters': {'content_type': 'str', 'headers': None}, 'data': ['{"meta": {"name": "AnchorTabular", "type": ["blackbox"], "explanations": ["local"], "params": {"seed": 1, "disc_perc": [25, 50, 75], "threshold": 0.95, "delta": 0.1, "tau": 0.15, "batch_size": 100, "coverage_samples": 10000, "beam_size": 1, "stop_on_first": false, "max_anchor_size": null, "min_samples_start": 100, "n_covered_ex": 10, "binary_cache_size": 10000, "cache_margin": 1000, "verbose": false, "verbose_every": 1, "kwargs": {}}, "version": "0.7.0"}, "data": {"anchor": ["Marital Status = Never-Married", "Relationship = Own-child", "Capital Gain <= 0.00"], "precision": 0.9937106918238994, "coverage": 0.06853582554517133, "raw": {"feature": [3, 5, 8], "mean": [0.8059701492537313, 0.9416666666666667, 0.9937106918238994], "precision": [0.8059701492537313, 0.9416666666666667, 0.9937106918238994], "coverage": [0.3037383177570093, 0.07165109034267912, 0.06853582554517133], "examples": [{"covered_true": [[44, 4, 1, 1, 5, 0, 4, 1, 0, 1848, 40, 9], [34, 4, 1, 1, 2, 0, 1, 1, 0, 0, 40, 9], [33, 2, 1, 1, 5, 1, 0, 1, 0, 0, 60, 9], [32, 1, 1, 1, 8, 1, 4, 0, 0, 0, 60, 9], [54, 4, 1, 1, 8, 0, 4, 1, 0, 0, 40, 9], [25, 4, 1, 1, 8, 1, 4, 1, 0, 0, 35, 9], [24, 4, 1, 1, 7, 3, 4, 0, 0, 0, 30, 9], [33, 2, 1, 1, 1, 0, 4, 1, 0, 1848, 45, 9], [36, 4, 5, 1, 4, 4, 1, 0, 0, 0, 40, 1], [29, 4, 1, 1, 8, 5, 4, 0, 0, 0, 50, 9]], "covered_false": [[50, 1, 1, 1, 8, 1, 4, 1, 0, 0, 55, 9], [46, 6, 5, 1, 8, 0, 4, 1, 0, 0, 50, 9], [50, 5, 1, 1, 8, 0, 4, 1, 15024, 0, 60, 9], [45, 7, 1, 1, 8, 0, 4, 1, 0, 1977, 60, 9], [45, 4, 5, 1, 8, 0, 4, 1, 0, 0, 65, 9], [42, 2, 1, 1, 1, 1, 4, 0, 99999, 0, 40, 9], [57, 4, 5, 1, 8, 1, 4, 1, 0, 2824, 50, 9], [74, 6, 1, 1, 2, 1, 4, 1, 15831, 0, 8, 3], [45, 6, 1, 1, 6, 1, 4, 1, 14084, 0, 45, 9], [44, 4, 1, 1, 8, 0, 4, 1, 15024, 0, 50, 9]], "uncovered_true": [], "uncovered_false": []}, {"covered_true": [[28, 4, 1, 1, 8, 3, 2, 0, 0, 0, 45, 0], [51, 4, 1, 1, 6, 3, 4, 1, 0, 0, 40, 9], [41, 4, 1, 1, 1, 3, 4, 0, 0, 0, 40, 9], [41, 7, 1, 1, 8, 3, 4, 1, 0, 0, 40, 9], [55, 6, 1, 1, 8, 3, 4, 1, 0, 0, 40, 9], [22, 4, 1, 1, 1, 3, 4, 0, 0, 0, 35, 9], [32, 4, 1, 1, 8, 3, 4, 0, 0, 0, 50, 9], [45, 4, 1, 1, 6, 3, 4, 1, 0, 1902, 40, 9], [29, 4, 2, 1, 5, 3, 1, 1, 0, 0, 60, 1], [36, 7, 5, 1, 5, 3, 4, 1, 5455, 0, 30, 9]], "covered_false": [[44, 4, 1, 1, 6, 3, 4, 1, 7688, 0, 50, 9], [50, 5, 1, 1, 8, 3, 4, 1, 15024, 0, 60, 9], [39, 4, 5, 1, 8, 3, 2, 0, 15020, 0, 60, 9], [66, 4, 1, 1, 8, 3, 4, 1, 99999, 0, 55, 0], [47, 4, 5, 1, 8, 3, 4, 1, 15024, 0, 55, 9], [32, 4, 1, 1, 5, 3, 4, 1, 99999, 0, 50, 9]], "uncovered_true": [], "uncovered_false": []}, {"covered_true": [[31, 4, 1, 1, 8, 3, 4, 1, 0, 1977, 55, 9], [59, 4, 2, 1, 5, 3, 4, 1, 0, 2415, 45, 9], [37, 4, 1, 1, 5, 3, 3, 1, 0, 0, 50, 0], [26, 4, 1, 1, 5, 3, 4, 1, 0, 0, 40, 9], [35, 4, 1, 1, 6, 3, 1, 1, 0, 0, 50, 0], [43, 4, 5, 1, 8, 3, 1, 1, 0, 0, 40, 1], [61, 2, 5, 1, 5, 3, 4, 0, 0, 0, 70, 9], [49, 6, 1, 1, 2, 3, 4, 1, 0, 0, 50, 9], [45, 2, 5, 1, 5, 3, 4, 1, 0, 0, 37, 9], [48, 2, 5, 1, 5, 3, 4, 0, 0, 1380, 40, 9]], "covered_false": [[36, 4, 1, 1, 6, 3, 4, 1, 0, 2415, 45, 9]], "uncovered_true": [], "uncovered_false": []}], "all_precision": 0, "num_preds": 1000000, "success": true, "names": ["Marital Status = Never-Married", "Relationship = Own-child", "Capital Gain <= 0.00"], "prediction": [0], "instance": [47.0, 4.0, 1.0, 1.0, 1.0, 3.0, 4.0, 1.0, 0.0, 0.0, 40.0, 9.0], "instances": [[47.0, 4.0, 1.0, 1.0, 1.0, 3.0, 4.0, 1.0, 0.0, 0.0, 40.0, 9.0]]}}}']}]}
```
````
### Cleanup


```bash
seldon model unload income-preprocess
seldon model unload income
seldon model unload income-drift
seldon model unload income-outlier
seldon model unload income-explainer
```
````{collapse} Expand to see output
```json

    {}
    {}
    {}
    {}
    {}
```
````

```python

```
