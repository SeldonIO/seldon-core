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
      storageUri: "gs://seldon-models/scv2/examples/mlserver_1.2.3/income/preprocessor"
      requirements:
      - sklearn
    ---
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: income
    spec:
      storageUri: "gs://seldon-models/scv2/examples/mlserver_1.2.3/income/classifier"
      requirements:
      - sklearn
    ---
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: income-drift
    spec:
      storageUri: "gs://seldon-models/scv2/examples/mlserver_1.2.3/income/drift-detector"
      requirements:
        - mlserver
        - alibi-detect
    ---
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: income-outlier
    spec:
      storageUri: "gs://seldon-models/scv2/examples/mlserver_1.2.3/income/outlier-detector"
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
            "uid": "cf384ralig2s738ul620",
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
            "lastChangeTimestamp": "2023-01-17T11:11:42.467005603Z",
            "modelsReady": true
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
    {'model_name': '', 'outputs': [{'data': [0, 0, 1, 1, 0, 1, 0, 0, 1, 0, 0, 0, 0, 0, 1, 1, 0, 0, 0, 1], 'name': 'predict', 'shape': [20, 1], 'datatype': 'INT64'}, {'data': [0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0], 'name': 'is_outlier', 'shape': [1, 20], 'datatype': 'INT64'}]}



```python
!seldon pipeline inspect income-production.income-drift.outputs.is_drift
```

    seldon.default.model.income-drift.outputs	cf384sgbcf7s73dfr3fg	{"name":"is_drift", "datatype":"INT64", "shape":["1", "1"], "contents":{"int64Contents":["0"]}}


Show predictions from drift data. Should be drift and probably not outliers.


```python
batchSz=20
print(y_ref[0:batchSz])
infer("income-production.pipeline",batchSz,"drift")
```

    [0 0 1 1 0 1 0 0 1 0 0 0 0 0 1 1 0 0 0 1]
    <Response [200]>
    {'model_name': '', 'outputs': [{'data': [0, 0, 0, 1, 1, 0, 1, 1, 1, 0, 0, 0, 0, 0, 1, 0, 0, 1, 0, 1], 'name': 'predict', 'shape': [20, 1], 'datatype': 'INT64'}, {'data': [0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0], 'name': 'is_outlier', 'shape': [1, 20], 'datatype': 'INT64'}]}



```python
!seldon pipeline inspect income-production.income-drift.outputs.is_drift
```

    seldon.default.model.income-drift.outputs	cf38510bcf7s73dfr3g0	{"name":"is_drift", "datatype":"INT64", "shape":["1", "1"], "contents":{"int64Contents":["1"]}}


Show predictions from outlier data. Should be outliers and probably not drift.


```python
batchSz=20
print(y_ref[0:batchSz])
infer("income-production.pipeline",batchSz,"outlier")
```

    [0 0 1 1 0 1 0 0 1 0 0 0 0 0 1 1 0 0 0 1]
    <Response [200]>
    {'model_name': '', 'outputs': [{'data': [0, 0, 1, 1, 0, 0, 1, 0, 1, 0, 1, 0, 0, 1, 1, 1, 0, 0, 0, 1], 'name': 'predict', 'shape': [20, 1], 'datatype': 'INT64'}, {'data': [1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0, 1, 1, 0, 1, 1, 1, 0, 1], 'name': 'is_outlier', 'shape': [1, 20], 'datatype': 'INT64'}]}



```python
!seldon pipeline inspect income-production.income-drift.outputs.is_drift
```

    seldon.default.model.income-drift.outputs	cf38540bcf7s73dfr3gg	{"name":"is_drift", "datatype":"INT64", "shape":["1", "1"], "contents":{"int64Contents":["0"]}}


### Explanations


```python
!cat ../../models/income-explainer.yaml
```

    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: income-explainer
    spec:
      storageUri: "gs://seldon-models/scv2/examples/mlserver_1.2.3/income/explainer"
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
    {'model_name': 'income-explainer_1', 'model_version': '1', 'id': 'cafea910-c625-4e5d-9a79-c988204cb955', 'parameters': {}, 'outputs': [{'name': 'explanation', 'shape': [1, 1], 'datatype': 'BYTES', 'parameters': {'content_type': 'str'}, 'data': ['{"meta": {"name": "AnchorTabular", "type": ["blackbox"], "explanations": ["local"], "params": {"seed": 1, "disc_perc": [25, 50, 75], "threshold": 0.95, "delta": 0.1, "tau": 0.15, "batch_size": 100, "coverage_samples": 10000, "beam_size": 1, "stop_on_first": false, "max_anchor_size": null, "min_samples_start": 100, "n_covered_ex": 10, "binary_cache_size": 10000, "cache_margin": 1000, "verbose": false, "verbose_every": 1, "kwargs": {}}, "version": "0.9.0"}, "data": {"anchor": ["Marital Status = Never-Married", "Hours per week <= 40.00", "Relationship = Own-child"], "precision": 0.9810126582278481, "coverage": 0.05941255006675567, "raw": {"feature": [3, 10, 5], "mean": [0.8278301886792453, 0.900611620795107, 0.9810126582278481], "precision": [0.8278301886792453, 0.900611620795107, 0.9810126582278481], "coverage": [0.3037383177570093, 0.2071651090342679, 0.05941255006675567], "examples": [{"covered_true": [[58, 4, 1, 1, 8, 1, 4, 1, 0, 0, 38, 9], [39, 4, 5, 1, 5, 5, 4, 0, 0, 1977, 24, 9], [40, 4, 1, 1, 5, 0, 4, 1, 0, 0, 46, 9], [28, 4, 1, 1, 8, 5, 4, 0, 0, 0, 60, 9], [35, 0, 1, 1, 0, 0, 4, 1, 0, 0, 20, 9], [47, 4, 1, 1, 8, 0, 4, 1, 0, 1887, 40, 9], [23, 4, 1, 1, 8, 1, 4, 0, 0, 0, 45, 9], [36, 7, 5, 1, 5, 3, 4, 0, 0, 0, 30, 9], [38, 6, 1, 1, 5, 0, 4, 1, 0, 0, 70, 9], [43, 4, 5, 1, 4, 1, 4, 0, 0, 0, 40, 9]], "covered_false": [[53, 4, 5, 1, 6, 0, 4, 1, 7298, 0, 60, 9], [43, 4, 5, 1, 5, 0, 4, 1, 0, 0, 50, 9], [55, 7, 5, 1, 8, 0, 4, 1, 0, 0, 45, 9], [39, 4, 1, 1, 5, 5, 4, 0, 7688, 0, 32, 9], [59, 4, 5, 1, 6, 1, 4, 0, 27828, 0, 45, 9], [46, 4, 5, 1, 8, 0, 4, 1, 0, 0, 55, 9], [47, 4, 5, 1, 8, 0, 4, 1, 99999, 0, 50, 9], [32, 4, 1, 1, 6, 0, 4, 1, 99999, 0, 60, 9], [43, 4, 5, 1, 8, 1, 4, 1, 0, 0, 50, 9], [49, 4, 5, 1, 5, 1, 4, 1, 0, 0, 40, 9]], "uncovered_true": [], "uncovered_false": []}, {"covered_true": [[40, 4, 1, 1, 6, 1, 4, 1, 0, 0, 16, 9], [32, 4, 1, 1, 8, 4, 4, 0, 0, 0, 25, 1], [60, 2, 5, 1, 5, 0, 4, 1, 0, 0, 35, 9], [63, 0, 5, 1, 0, 1, 4, 0, 0, 0, 40, 9], [24, 4, 1, 1, 6, 1, 4, 1, 0, 0, 40, 9], [40, 4, 1, 1, 1, 4, 2, 1, 0, 0, 40, 9], [76, 0, 1, 1, 0, 0, 4, 1, 0, 0, 40, 9], [52, 1, 1, 1, 8, 0, 4, 1, 0, 0, 40, 9], [40, 7, 5, 1, 5, 4, 4, 0, 0, 0, 30, 9], [52, 2, 5, 1, 5, 3, 4, 0, 0, 0, 40, 9]], "covered_false": [[44, 4, 1, 1, 8, 0, 4, 1, 7298, 0, 35, 9], [64, 6, 2, 1, 8, 0, 4, 1, 0, 0, 24, 9], [57, 6, 2, 1, 6, 0, 4, 1, 0, 1902, 40, 9], [42, 1, 1, 1, 8, 0, 4, 1, 7298, 0, 40, 9], [78, 5, 1, 1, 8, 0, 4, 1, 0, 2392, 40, 9], [60, 7, 2, 1, 5, 0, 4, 1, 7688, 0, 30, 9], [42, 2, 1, 1, 5, 0, 4, 1, 5178, 0, 40, 9], [50, 4, 1, 1, 8, 0, 4, 1, 7298, 0, 30, 9]], "uncovered_true": [], "uncovered_false": []}, {"covered_true": [[35, 1, 1, 1, 4, 3, 1, 1, 0, 0, 40, 0], [23, 4, 1, 1, 1, 3, 4, 0, 0, 0, 40, 9], [25, 2, 1, 1, 5, 3, 4, 0, 0, 0, 40, 9], [25, 4, 1, 1, 6, 3, 4, 1, 0, 0, 40, 9], [36, 2, 1, 1, 5, 3, 4, 1, 0, 0, 35, 9], [50, 6, 1, 1, 6, 3, 4, 1, 0, 0, 40, 9], [24, 2, 1, 1, 5, 3, 4, 0, 0, 0, 40, 9], [45, 2, 5, 1, 5, 3, 4, 1, 0, 0, 38, 9], [27, 7, 1, 1, 5, 3, 4, 1, 0, 0, 40, 9], [36, 2, 1, 1, 5, 3, 4, 0, 0, 0, 21, 9]], "covered_false": [[49, 4, 5, 1, 8, 3, 4, 1, 15024, 0, 35, 9], [39, 4, 5, 1, 5, 3, 4, 1, 99999, 0, 40, 9], [55, 4, 5, 1, 8, 3, 4, 1, 99999, 0, 40, 9]], "uncovered_true": [], "uncovered_false": []}], "all_precision": 0, "num_preds": 1000000, "success": true, "names": ["Marital Status = Never-Married", "Hours per week <= 40.00", "Relationship = Own-child"], "prediction": [0], "instance": [47.0, 4.0, 1.0, 1.0, 1.0, 3.0, 4.0, 1.0, 0.0, 0.0, 40.0, 9.0], "instances": [[47.0, 4.0, 1.0, 1.0, 1.0, 3.0, 4.0, 1.0, 0.0, 0.0, 40.0, 9.0]]}}}']}]}


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
