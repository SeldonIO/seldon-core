## Tabular Income Classifier Production Deployment

To run this notebook you need the inference data. This can be acquired in two ways:

  * Run train.ipynb
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


```python
!cat ../../models/income-preprocess.yaml
!echo "---"
!cat ../../models/income.yaml
!echo "---"
!cat ../../models/income-drift.yaml
!echo "---"
!cat ../../models/income-outlier.yaml
```

    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: income-preprocess
    spec:
      storageUri: "gs://seldon-models/scv2/examples/income/preprocessor"
      requirements:
      - sklearn
    ---
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: income
    spec:
      storageUri: "gs://seldon-models/scv2/examples/income/classifier"
      requirements:
      - sklearn
    ---
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: income-drift
    spec:
      storageUri: "gs://seldon-models/scv2/examples/income/drift-detector"
      requirements:
        - mlserver
        - alibi-detect
    ---
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: income-outlier
    spec:
      storageUri: "gs://seldon-models/scv2/examples/income/outlier-detector"
      requirements:
        - mlserver
        - alibi-detect



```python
!seldon model load -f ../../models/income-preprocess.yaml
!seldon model load -f ../../models/income.yaml
!seldon model load -f ../../models/income-drift.yaml
!seldon model load -f ../../models/income-outlier.yaml
```

    {}
    {}
    {}
    {}



```python
!seldon model status income-preprocess -w ModelAvailable | jq .
!seldon model status income -w ModelAvailable | jq .
!seldon model status income-drift -w ModelAvailable | jq .
!seldon model status income-outlier -w ModelAvailable | jq .
```

    [1;39m{}[0m
    [1;39m{}[0m
    [1;39m{}[0m
    [1;39m{}[0m



```python
!cat ../../pipelines/income.yaml
```

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



```python
!seldon pipeline load -f ../../pipelines/income.yaml
```

    {}



```python
!seldon pipeline status income-production -w PipelineReady| jq -M .
```

    {
      "pipelineName": "income-production",
      "versions": [
        {
          "pipeline": {
            "name": "income-production",
            "uid": "ccss58q4u1vus3q1e1ig",
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
            "lastChangeTimestamp": "2022-10-02T16:56:03.396751091Z"
          }
        }
      ]
    }


Show predictions from reference set. Should not be drift or outliers.


```python
batchSz=20
print(y_ref[0:batchSz])
infer("income-production.pipeline",batchSz,"normal")
```

    [0 0 1 1 0 1 0 0 1 0 0 0 0 0 1 1 0 0 0 1]
    <Response [200]>
    {'model_name': '', 'outputs': [{'data': [0, 0, 1, 1, 0, 1, 0, 0, 1, 0, 0, 0, 0, 0, 1, 1, 0, 0, 0, 1], 'name': 'predict', 'shape': [20], 'datatype': 'INT64'}, {'data': [0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0], 'name': 'is_outlier', 'shape': [1, 20], 'datatype': 'INT64'}]}



```python
!seldon pipeline inspect income-production.income-drift.outputs.is_drift
```

    seldon.default.model.income-drift.outputs	ccss59patu79k6fs0pjg	{"name":"is_drift","datatype":"INT64","shape":["1"],"contents":{"int64Contents":["0"]}}


Show predictions from drift data. Should be drift and probably not outliers.


```python
batchSz=20
print(y_ref[0:batchSz])
infer("income-production.pipeline",batchSz,"drift")
```

    [0 0 1 1 0 1 0 0 1 0 0 0 0 0 1 1 0 0 0 1]
    <Response [200]>
    {'model_name': '', 'outputs': [{'data': [0, 0, 0, 1, 1, 0, 1, 1, 1, 0, 0, 0, 0, 0, 1, 0, 0, 1, 0, 1], 'name': 'predict', 'shape': [20], 'datatype': 'INT64'}, {'data': [0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0], 'name': 'is_outlier', 'shape': [1, 20], 'datatype': 'INT64'}]}



```python
!seldon pipeline inspect income-production.income-drift.outputs.is_drift
```

    seldon.default.model.income-drift.outputs	ccss5dpatu79k6fs0pk0	{"name":"is_drift","datatype":"INT64","shape":["1"],"contents":{"int64Contents":["1"]}}


Show predictions from outlier data. Should be outliers and probably not drift.


```python
batchSz=20
print(y_ref[0:batchSz])
infer("income-production.pipeline",batchSz,"outlier")
```

    [0 0 1 1 0 1 0 0 1 0 0 0 0 0 1 1 0 0 0 1]
    <Response [200]>
    {'model_name': '', 'outputs': [{'data': [0, 0, 1, 1, 0, 0, 0, 1, 1, 1, 1, 0, 0, 0, 1, 1, 0, 0, 0, 1], 'name': 'predict', 'shape': [20], 'datatype': 'INT64'}, {'data': [1, 0, 0, 0, 1, 1, 1, 1, 1, 1, 1, 0, 0, 0, 1, 1, 1, 0, 1, 1], 'name': 'is_outlier', 'shape': [1, 20], 'datatype': 'INT64'}]}



```python
!seldon pipeline inspect income-production.income-drift.outputs.is_drift
```

    seldon.default.model.income-drift.outputs	ccss5ghatu79k6fs0pkg	{"name":"is_drift","datatype":"INT64","shape":["1"],"contents":{"int64Contents":["0"]}}


### Explanations


```python
!cat ../../models/income-explainer.yaml
```

    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: income-explainer
    spec:
      storageUri: "gs://seldon-models/scv2/examples/income/explainer"
      explainer:
        type: anchor_tabular
        modelRef: income



```python
!seldon model load -f ../../models/income-explainer.yaml
```

    {}



```python
!seldon model status income-explainer -w ModelAvailable | jq .
```

    [1;39m{}[0m



```python
batchSz=1
print(y_ref[0:batchSz])
infer("income-explainer",batchSz,"normal")
```

    [0]
    <Response [200]>
    {'model_name': 'income-explainer_1', 'model_version': '1', 'id': '4e5362b0-e2c7-41e0-82d5-f1a240f60834', 'parameters': {'content_type': None, 'headers': None}, 'outputs': [{'name': 'explanation', 'shape': [1], 'datatype': 'BYTES', 'parameters': {'content_type': 'str', 'headers': None}, 'data': ['{"meta": {"name": "AnchorTabular", "type": ["blackbox"], "explanations": ["local"], "params": {"seed": 1, "disc_perc": [25, 50, 75], "threshold": 0.95, "delta": 0.1, "tau": 0.15, "batch_size": 100, "coverage_samples": 10000, "beam_size": 1, "stop_on_first": false, "max_anchor_size": null, "min_samples_start": 100, "n_covered_ex": 10, "binary_cache_size": 10000, "cache_margin": 1000, "verbose": false, "verbose_every": 1, "kwargs": {}}, "version": "0.7.0"}, "data": {"anchor": ["Marital Status = Never-Married", "Relationship = Own-child"], "precision": 1.0, "coverage": 0.07165109034267912, "raw": {"feature": [3, 5], "mean": [0.8044052863436123, 1.0], "precision": [0.8044052863436123, 1.0], "coverage": [0.3037383177570093, 0.07165109034267912], "examples": [{"covered_true": [[24, 4, 1, 1, 6, 3, 4, 1, 0, 0, 40, 9], [31, 4, 1, 1, 8, 1, 4, 1, 0, 0, 40, 9], [28, 4, 1, 1, 8, 1, 4, 0, 0, 0, 50, 9], [49, 4, 5, 1, 5, 1, 4, 1, 0, 0, 60, 9], [24, 0, 1, 1, 0, 1, 1, 1, 0, 0, 25, 2], [41, 4, 5, 1, 1, 1, 4, 0, 0, 0, 40, 9], [34, 1, 1, 1, 8, 4, 1, 1, 1471, 0, 40, 7], [61, 4, 1, 1, 1, 1, 4, 1, 0, 0, 40, 9], [55, 5, 1, 1, 6, 0, 4, 1, 0, 0, 60, 9], [26, 4, 5, 1, 5, 0, 4, 1, 0, 0, 20, 9]], "covered_false": [[47, 2, 5, 1, 5, 1, 4, 1, 0, 0, 35, 9], [30, 4, 1, 1, 8, 4, 4, 0, 0, 0, 45, 9], [49, 7, 2, 1, 8, 1, 4, 0, 0, 2258, 50, 9], [59, 1, 5, 1, 5, 1, 4, 0, 0, 0, 40, 0], [42, 4, 1, 1, 8, 0, 4, 1, 5178, 0, 40, 9], [36, 4, 5, 1, 6, 0, 4, 1, 15024, 0, 45, 9], [33, 4, 1, 1, 5, 0, 4, 1, 15024, 0, 75, 9], [40, 4, 1, 1, 6, 1, 3, 0, 0, 0, 55, 9], [32, 4, 1, 1, 6, 0, 4, 1, 99999, 0, 50, 9], [30, 4, 1, 1, 4, 1, 4, 0, 0, 0, 40, 9]], "uncovered_true": [], "uncovered_false": []}, {"covered_true": [[37, 4, 5, 1, 5, 3, 4, 1, 0, 0, 48, 9], [28, 4, 1, 1, 4, 3, 4, 0, 0, 0, 40, 9], [64, 0, 1, 1, 0, 3, 2, 0, 0, 0, 8, 9], [69, 4, 2, 1, 5, 3, 4, 1, 0, 0, 40, 2], [48, 7, 5, 1, 5, 3, 4, 1, 0, 0, 40, 9], [45, 4, 1, 1, 5, 3, 4, 0, 0, 0, 50, 9], [59, 0, 1, 1, 0, 3, 4, 1, 0, 0, 40, 9], [55, 5, 5, 1, 8, 3, 4, 1, 0, 0, 60, 9], [43, 4, 1, 1, 6, 3, 4, 1, 0, 0, 44, 9], [36, 4, 1, 1, 8, 3, 4, 1, 0, 0, 55, 9]], "covered_false": [[26, 4, 1, 1, 5, 3, 4, 1, 10520, 0, 64, 9], [49, 7, 5, 1, 5, 3, 2, 0, 7430, 0, 40, 9], [62, 4, 2, 1, 5, 3, 4, 1, 15024, 0, 50, 9], [63, 6, 1, 1, 6, 3, 4, 1, 10605, 0, 40, 9], [41, 4, 1, 1, 6, 3, 4, 1, 15024, 0, 45, 9], [39, 4, 1, 1, 8, 3, 4, 1, 15024, 0, 60, 9], [44, 4, 5, 1, 5, 3, 4, 1, 7688, 0, 55, 9], [46, 4, 5, 1, 5, 3, 4, 0, 7688, 0, 35, 9], [32, 4, 1, 1, 8, 3, 4, 1, 7298, 0, 45, 9], [51, 5, 5, 1, 8, 3, 4, 1, 15024, 0, 50, 9]], "uncovered_true": [], "uncovered_false": []}], "all_precision": 0, "num_preds": 1000000, "success": true, "names": ["Marital Status = Never-Married", "Relationship = Own-child"], "prediction": [0], "instance": [47.0, 4.0, 1.0, 1.0, 1.0, 3.0, 4.0, 1.0, 0.0, 0.0, 40.0, 9.0], "instances": [[47.0, 4.0, 1.0, 1.0, 1.0, 3.0, 4.0, 1.0, 0.0, 0.0, 40.0, 9.0]]}}}']}]}


### Cleanup


```python
!seldon model unload income-preprocess
!seldon model unload income
!seldon model unload income-drift
!seldon model unload income-outlier
!seldon model unload income-explainer
```

    {}
    {}
    {}
    {}
    {}



```python

```
