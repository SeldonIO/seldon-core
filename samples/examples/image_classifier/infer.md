## CIFAR10 Image Classification Production Deployment

We show an image classifier (CIFAR10) with associated outlier and drift detectors using a Pipeline.

 * The model is a tensorflow [CIFAR10](https://www.cs.toronto.edu/~kriz/cifar.html) image classfier 
 * The outlier detector is created from the [CIFAR10 VAE Outlier example](https://docs.seldon.io/projects/alibi-detect/en/stable/examples/od_vae_cifar10.html).
 * The drift detector is created from the [CIFAR10 KS Drift example](https://docs.seldon.io/projects/alibi-detect/en/stable/examples/cd_ks_cifar10.html)
 


```python
import requests
import json
from typing import Dict, List
import numpy as np
import os
import tensorflow as tf
from alibi_detect.utils.perturbation import apply_mask
from alibi_detect.datasets import fetch_cifar10c
import matplotlib.pyplot as plt
tf.keras.backend.clear_session()
```

    2022-05-26 11:27:37.239516: W tensorflow/stream_executor/platform/default/dso_loader.cc:64] Could not load dynamic library 'libcudart.so.11.0'; dlerror: libcudart.so.11.0: cannot open shared object file: No such file or directory
    2022-05-26 11:27:37.239536: I tensorflow/stream_executor/cuda/cudart_stub.cc:29] Ignore above cudart dlerror if you do not have a GPU set up on your machine.
```
````

```python
train, test = tf.keras.datasets.cifar10.load_data()
X_train, y_train = train
X_test, y_test = test

X_train = X_train.astype('float32') / 255
X_test = X_test.astype('float32') / 255
print(X_train.shape, y_train.shape, X_test.shape, y_test.shape)
```

    (50000, 32, 32, 3) (50000, 1) (10000, 32, 32, 3) (10000, 1)
```
````

```python
outliers = []
for idx in range(0,X_train.shape[0]):
    X_mask, mask = apply_mask(X_train[idx].reshape(1, 32, 32, 3),
                                  mask_size=(12,12),
                                  n_masks=1,
                                  channels=[0,1,2],
                                  mask_type='normal',
                                  noise_distr=(0,1),
                                  clip_rng=(0,1))
    outliers.append(X_mask)
X_outliers = np.vstack(outliers)
X_outliers.shape
```




    (50000, 32, 32, 3)
```
````


```python
corruption = ['gaussian_noise']
X_corr, y_corr = fetch_cifar10c(corruption=corruption, severity=5, return_X_y=True)
X_corr = X_corr.astype('float32') / 255
```


```python
reqJson = json.loads('{"inputs":[{"name":"input_1","data":[],"datatype":"FP32","shape":[]}]}')
url = "http://0.0.0.0:9000/v2/models/model/infer"
```


```python
def infer(resourceName: str, batchSz: int, requestType: str):
    if requestType == "outlier":
        rows = X_outliers[0:0+batchSz]
    elif requestType == "drift":
        rows = X_corr[0:0+batchSz]
    else:
        rows = X_train[0:0+batchSz]
    reqJson["inputs"][0]["data"] = rows.flatten().tolist()
    reqJson["inputs"][0]["shape"] = [batchSz, 32, 32, 3]
    headers = {"Content-Type": "application/json", "seldon-model":resourceName}
    response_raw = requests.post(url, json=reqJson, headers=headers)
    print(response_raw)
    print(response_raw.json())
```

### Pipeline


```bash
cat ../../models/cifar10.yaml
echo "---"
cat ../../models/cifar10-outlier-detect.yaml
echo "---"
cat ../../models/cifar10-drift-detect.yaml
```
````{collapse} Expand to see output
```yaml
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: cifar10
      namespace: seldon-mesh
    spec:
      storageUri: "gs://seldon-models/triton/tf_cifar10"
      requirements:
      - tensorflow
    ---
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: cifar10-outlier
      namespace: seldon-mesh
    spec:
      storageUri: "gs://seldon-models/mlserver/alibi-detect/cifar10-outlier"
      requirements:
        - mlserver
        - alibi-detect
    ---
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: cifar10-drift
      namespace: seldon-mesh
    spec:
      storageUri: "gs://seldon-models/mlserver/alibi-detect/cifar10-drift"
      requirements:
        - mlserver
        - alibi-detect
```
````

```bash
seldon model load -f ../../models/cifar10.yaml
seldon model load -f ../../models/cifar10-outlier-detect.yaml
seldon model load -f ../../models/cifar10-drift-detect.yaml
```
````{collapse} Expand to see output
```json

    {}
    {}
    {}
```
````

```bash
seldon model status cifar10 -w ModelAvailable | jq .
seldon model status cifar10-outlier -w ModelAvailable | jq .
seldon model status cifar10-drift -w ModelAvailable | jq .
```
````{collapse} Expand to see output
```json

    [1;39m{}[0m
    [1;39m{}[0m
    [1;39m{}[0m
```
````

```bash
cat ../../pipelines/cifar10.yaml
```
````{collapse} Expand to see output
```yaml
    apiVersion: mlops.seldon.io/v1alpha1
    kind: Pipeline
    metadata:
      name: cifar10-production
      namespace: seldon-mesh
    spec:
      steps:
        - name: cifar10
        - name: cifar10-outlier
        - name: cifar10-drift
          batch:
            size: 20
      output:
        steps:
        - cifar10
        - cifar10-outlier.outputs.is_outlier
```
````

```bash
seldon pipeline load -f ../../pipelines/cifar10.yaml
```
````{collapse} Expand to see output
```json

    {}
```
````

```bash
seldon pipeline status cifar10-production -w PipelineReady| jq -M .
```
````{collapse} Expand to see output
```json

    {
      "pipelineName": "cifar10-production",
      "versions": [
        {
          "pipeline": {
            "name": "cifar10-production",
            "uid": "ca7lvtv7a4umaeddt0gg",
            "version": 1,
            "steps": [
              {
                "name": "cifar10"
              },
              {
                "name": "cifar10-drift",
                "batch": {
                  "size": 20
                }
              },
              {
                "name": "cifar10-outlier"
              }
            ],
            "output": {
              "steps": [
                "cifar10.outputs",
                "cifar10-outlier.outputs.is_outlier"
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
            "lastChangeTimestamp": "2022-05-26T11:09:44.988745665Z"
          }
        }
      ]
    }
```
````

```python
infer("cifar10-production.pipeline",20, "drift")
```

    <Response [200]>
    {'model_name': '', 'outputs': [{'data': None, 'name': 'fc10', 'shape': [20, 10], 'datatype': 'FP32'}, {'data': [1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 1, 0, 0, 0, 0], 'name': 'is_outlier', 'shape': [1, 20], 'datatype': 'INT64'}], 'rawOutputContents': ['1M3eMCFl3C/cUvA5KrBuP0Xg3jSP4Aw3uomJPdTueDAoMCEx6rwLL1u1rjrZcAs3o6htP2N/oTuhl+E7Bna8OQHG6TkQE0M3aDdsPel6jjZcLc82+rKNNiFiFTwwvR8/QKsUOLlYODm+MlU3OaxsNae7uz6jhyU1KYA0OBO6lTVRu3g/9GBsO5PPmTkLHu04hFAcOiNWgzddU8I8NA+CNfQAuTWWldAzJoHiPTiaYz97L0o40eYbOOw5fjnaVaA0r5sDNdkSgTKvbiA1Nj89NAlBjD6Eyzk/zY58OKIy7zh2MQA5XmvDNRQ5xzbrWwszxF3mMtLKqzLDAA855NtpP+kFgjS35as0EdmwPTeukzHv0KwxHQRCMdHlLzq9Myk4nyVaPzNBwT3kurA8dVkkOz/sGjuYRFE5eNnVPD8GKDhJm6E1lGNmM1QEZz6OIUQ/jwACOSgviTjYMgQ8PFLtNGmaFzW1J1Ez5h0qOCMUrTUKsw4+uldwPrzqgzwsjMY7qngaP6BNHDdvxqI518oDNwC6TjiV+481gd0hPqLJVj9lhoA6/agLOquEZzqryY42s27yOcapUjVeQAU4rgqzNbNP7T4k/jU8G5sAP8MnEzgdNLc8aoLcN2C6AzpLa4o44mihM05zZzMzXsw+x8cZP+MNWji1Hqo2MG6oOECFCzXUcJg1xYinMiSTFjO6AOUyGRrtOphZfz/laB04/GOWNWyFNDoo3Fc1tOUFNoxKwTLXVAwzhIKxMEsTfj/fCU47AASFO9vbMjH7JJ05bue1M5yKCTfJgOg2qiNTNtTeKjSMKYw84bJ6P8i2iDmiPB82ti5aOxpIsTQqZBs2f0pPM8uRBDDE5hkvuybWObuSfz9eB540YaZ2MpX4pDqQVNwwP0SRMCIBdzCWQbozMULkMe4+Qj97xXY+MZ4oOfsgmDT8CKM4k93RMxSuJTWmjgozSbhpNtpdNzWbgkE/9U14PoktHTrRefs4d9U7OpWQ2jXOYjY5EzY6NGhkITZ5CBI02cI5PwxYjD466Fc3XOG7N367STnPDno0JncBOIF46DE=']}
```
````

```bash
seldon pipeline inspect cifar10-production.cifar10-drift.outputs.is_drift
```
````{collapse} Expand to see output
```json

    ---
    seldon.default.model.cifar10-drift.outputs
    {"name":"is_drift", "datatype":"INT64", "shape":["1"], "contents":{"int64Contents":["1"]}}
```
````

```bash
seldon pipeline unload -p cifar10-production
```
````{collapse} Expand to see output
```json

    {}
```
````

```bash
seldon model unload cifar10
seldon model unload cifar10-outlier
seldon model unload cifar10-drift
```
````{collapse} Expand to see output
```json

    {}
    {}
    {}
```
````

```python

```
