## CIFAR10 Image Classification Production Deployment

![cifar10](demo.png)

We show an image classifier (CIFAR10) with associated outlier and drift detectors using a Pipeline.

 * The model is a tensorflow [CIFAR10](https://www.cs.toronto.edu/~kriz/cifar.html) image classfier
 * The outlier detector is created from the [CIFAR10 VAE Outlier example](https://docs.seldon.io/projects/alibi-detect/en/stable/examples/od_vae_cifar10.html).
 * The drift detector is created from the [CIFAR10 KS Drift example](https://docs.seldon.io/projects/alibi-detect/en/stable/examples/cd_ks_cifar10.html)

### Model Training (optional for notebook)

To run local training run the [training notebook](train.ipynb).

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

```
2023-02-03 18:04:20.979821: W tensorflow/stream_executor/platform/default/dso_loader.cc:64] Could not load dynamic library 'libcudart.so.11.0'; dlerror: libcudart.so.11.0: cannot open shared object file: No such file or directory
2023-02-03 18:04:20.979834: I tensorflow/stream_executor/cuda/cudart_stub.cc:29] Ignore above cudart dlerror if you do not have a GPU set up on your machine.

```

```python
train, test = tf.keras.datasets.cifar10.load_data()
X_train, y_train = train
X_test, y_test = test

X_train = X_train.astype('float32') / 255
X_test = X_test.astype('float32') / 255
print(X_train.shape, y_train.shape, X_test.shape, y_test.shape)
classes = (
    "plane",
    "car",
    "bird",
    "cat",
    "deer",
    "dog",
    "frog",
    "horse",
    "ship",
    "truck",
)

```

```
(50000, 32, 32, 3) (50000, 1) (10000, 32, 32, 3) (10000, 1)

```

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

```
(50000, 32, 32, 3)

```

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
    for i in range(batchSz):
        show(rows[i])
    reqJson["inputs"][0]["data"] = rows.flatten().tolist()
    reqJson["inputs"][0]["shape"] = [batchSz, 32, 32, 3]
    headers = {"Content-Type": "application/json", "seldon-model":resourceName}
    response_raw = requests.post(url, json=reqJson, headers=headers)
    print(response_raw)
    print(response_raw.json())


def show(X):
    plt.imshow(X.reshape(32, 32, 3))
    plt.axis("off")
    plt.show()

```

### Pipeline

```bash
cat ../../models/cifar10.yaml
echo "---"
cat ../../models/cifar10-outlier-detect.yaml
echo "---"
cat ../../models/cifar10-drift-detect.yaml
```

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Model
metadata:
  name: cifar10
spec:
  storageUri: "gs://seldon-models/triton/tf_cifar10"
  requirements:
  - tensorflow
---
apiVersion: mlops.seldon.io/v1alpha1
kind: Model
metadata:
  name: cifar10-outlier
spec:
  storageUri: "gs://seldon-models/scv2/examples/cifar10/outlier-detector"
  requirements:
    - mlserver
    - alibi-detect
---
apiVersion: mlops.seldon.io/v1alpha1
kind: Model
metadata:
  name: cifar10-drift
spec:
  storageUri: "gs://seldon-models/scv2/examples/cifar10/drift-detector"
  requirements:
    - mlserver
    - alibi-detect

```

```bash
seldon model load -f ../../models/cifar10.yaml
seldon model load -f ../../models/cifar10-outlier-detect.yaml
seldon model load -f ../../models/cifar10-drift-detect.yaml
```

```json
{}
{}
{}

```

```bash
seldon model status cifar10 -w ModelAvailable | jq .
seldon model status cifar10-outlier -w ModelAvailable | jq .
seldon model status cifar10-drift -w ModelAvailable | jq .
```

```json
{}
{}
{}

```

```bash
cat ../../pipelines/cifar10.yaml
```

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Pipeline
metadata:
  name: cifar10-production
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

```bash
seldon pipeline load -f ../../pipelines/cifar10.yaml
```

```json
{}

```

```bash
seldon pipeline status cifar10-production -w PipelineReady| jq -M .
```

```json
{
  "pipelineName": "cifar10-production",
  "versions": [
    {
      "pipeline": {
        "name": "cifar10-production",
        "uid": "cfekpeoq4n3c73fctno0",
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
        "kubernetesMeta": {}
      },
      "state": {
        "pipelineVersion": 1,
        "status": "PipelineReady",
        "reason": "created pipeline",
        "lastChangeTimestamp": "2023-02-03T18:04:44.043515049Z",
        "modelsReady": true
      }
    }
  ]
}

```

```python
infer("cifar10-production.pipeline",20, "normal")
```

```
![png](infer_files/infer_14_0.png)

```

```
![png](infer_files/infer_14_1.png)

```

```
![png](infer_files/infer_14_2.png)

```

```
![png](infer_files/infer_14_3.png)

```

```
![png](infer_files/infer_14_4.png)

```

```
![png](infer_files/infer_14_5.png)

```

```
![png](infer_files/infer_14_6.png)

```

```
![png](infer_files/infer_14_7.png)

```

```
![png](infer_files/infer_14_8.png)

```

```
![png](infer_files/infer_14_9.png)

```

```
![png](infer_files/infer_14_10.png)

```

```
![png](infer_files/infer_14_11.png)

```

```
![png](infer_files/infer_14_12.png)

```

```
![png](infer_files/infer_14_13.png)

```

```
![png](infer_files/infer_14_14.png)

```

```
![png](infer_files/infer_14_15.png)

```

```
![png](infer_files/infer_14_16.png)

```

```
![png](infer_files/infer_14_17.png)

```

```
![png](infer_files/infer_14_18.png)

```

```
![png](infer_files/infer_14_19.png)

```

```
<Response [200]>
{'model_name': '', 'outputs': [{'data': [1.45001495e-08, 1.2525752e-09, 1.6298458e-07, 0.11529388, 1.7431412e-07, 6.1856604e-06, 0.8846994, 6.0739285e-09, 7.437921e-08, 4.7317337e-09, 1.26449e-06, 4.8814868e-09, 1.5153439e-09, 8.490656e-09, 5.5131194e-10, 1.1617216e-09, 5.7729294e-10, 2.8839776e-07, 0.0006149016, 0.99938357, 0.888746, 2.5331951e-06, 0.00012967695, 0.10531583, 2.4284174e-05, 6.3332986e-06, 0.0016261435, 1.13079e-05, 0.0013286703, 0.0028091935, 2.0993439e-06, 3.680449e-08, 0.0013269952, 2.1766558e-05, 0.99841356, 0.00015300694, 6.9472035e-06, 1.3277059e-05, 6.1860555e-05, 3.4072806e-07, 1.1205097e-05, 0.99997175, 1.9948227e-07, 6.9880834e-08, 3.3387135e-08, 5.2603138e-08, 3.0352305e-07, 4.3738982e-08, 5.3243946e-07, 1.5870584e-05, 0.0006525102, 0.013322109, 1.480307e-06, 0.9766325, 4.9847167e-05, 0.00058075984, 0.008405659, 5.2234273e-06, 0.00023390084, 0.000116047224, 1.6682397e-06, 5.7737526e-10, 0.9975605, 6.45564e-05, 0.002371972, 1.0392675e-07, 9.747962e-08, 1.4484569e-07, 8.762438e-07, 2.4758325e-08, 5.028761e-09, 6.856381e-11, 5.9932094e-12, 4.921233e-10, 1.471166e-07, 2.7940719e-06, 3.4563383e-09, 0.99999714, 5.9420524e-10, 9.445026e-11, 4.1854888e-05, 5.041549e-08, 8.0302314e-08, 1.2119854e-07, 6.781646e-09, 1.2616152e-08, 1.1878505e-08, 1.628573e-09, 0.9999578, 3.281738e-08, 0.08930307, 1.4065135e-07, 4.1117343e-07, 0.90898305, 8.933351e-07, 0.0015637449, 0.00013868928, 9.092981e-06, 4.8759745e-07, 4.3976044e-07, 0.00016094849, 3.5653954e-07, 0.0760521, 0.8927447, 0.0011777573, 0.00265573, 0.027189083, 4.1892267e-06, 1.329405e-05, 1.8564688e-06, 1.3373891e-06, 1.0251247e-07, 8.651912e-09, 4.458202e-06, 1.4646349e-05, 1.260957e-06, 1.046087e-08, 0.9998946, 8.332438e-05, 3.900894e-07, 6.53852e-05, 3.012202e-08, 1.0247197e-07, 1.8824371e-06, 0.0004958526, 3.533475e-05, 2.739997e-07, 0.99939275, 4.840305e-06, 3.5346695e-06, 0.0005518078, 3.1597017e-07, 0.99902296, 0.00031509742, 8.07886e-07, 1.6366084e-06, 2.795575e-06, 6.112367e-06, 9.817249e-05, 2.602709e-07, 0.0004561966, 5.360607e-06, 2.8656412e-05, 0.000116040654, 6.881144e-05, 8.844774e-06, 4.4655946e-05, 3.5564542e-05, 0.006564381, 0.9926715, 0.007300911, 1.766928e-06, 3.0520596e-07, 0.026906287, 1.3769699e-06, 0.00027539674, 5.583593e-06, 3.792553e-06, 0.0003876767, 0.9651169, 0.18114138, 2.8360228e-05, 0.00019927241, 0.007685872, 0.00014663498, 3.9361137e-05, 5.941682e-05, 7.36174e-05, 0.79936546, 0.01126067, 2.3992783e-11, 7.6336457e-16, 1.4644799e-15, 1, 2.4652159e-14, 1.1786078e-10, 1.9402116e-13, 4.2408636e-15, 1.209294e-15, 2.9042784e-15, 1.5366902e-08, 1.2476195e-09, 1.3560152e-07, 0.999997, 4.3113017e-11, 2.8163534e-08, 2.4494727e-06, 1.3122828e-10, 3.8081083e-07, 2.1628158e-11, 0.0004926238, 6.9424555e-06, 2.827196e-05, 0.92534137, 9.500486e-06, 0.00036133997, 0.072713904, 1.2831057e-07, 0.0010457055, 2.8514464e-07], 'name': 'fc10', 'shape': [20, 10], 'datatype': 'FP32'}, {'data': [0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0], 'name': 'is_outlier', 'shape': [1, 20], 'datatype': 'INT64'}]}

```

```bash
seldon pipeline inspect cifar10-production.cifar10-drift.outputs.is_drift
```

```
seldon.default.model.cifar10-drift.outputs	cfekpg8fh5ss739vr5tg	{"name":"is_drift","datatype":"INT64","shape":["1","1"],"contents":{"int64Contents":["0"]}}

```

```python
infer("cifar10-production.pipeline",20, "drift")
```

```
![png](infer_files/infer_16_0.png)

```

```
![png](infer_files/infer_16_1.png)

```

```
![png](infer_files/infer_16_2.png)

```

```
![png](infer_files/infer_16_3.png)

```

```
![png](infer_files/infer_16_4.png)

```

```
![png](infer_files/infer_16_5.png)

```

```
![png](infer_files/infer_16_6.png)

```

```
![png](infer_files/infer_16_7.png)

```

```
![png](infer_files/infer_16_8.png)

```

```
![png](infer_files/infer_16_9.png)

```

```
![png](infer_files/infer_16_10.png)

```

```
![png](infer_files/infer_16_11.png)

```

```
![png](infer_files/infer_16_12.png)

```

```
![png](infer_files/infer_16_13.png)

```

```
![png](infer_files/infer_16_14.png)

```

```
![png](infer_files/infer_16_15.png)

```

```
![png](infer_files/infer_16_16.png)

```

```
![png](infer_files/infer_16_17.png)

```

```
![png](infer_files/infer_16_18.png)

```

```
![png](infer_files/infer_16_19.png)

```

```
<Response [200]>
{'model_name': '', 'outputs': [{'data': [1.6211208e-09, 4.0089754e-10, 0.0004583845, 0.932375, 4.1514153e-07, 8.396962e-06, 0.0671577, 9.0561736e-10, 2.3456077e-09, 1.2709157e-10, 0.0013329197, 8.311341e-06, 0.92835414, 0.004928552, 0.006884555, 0.00035946266, 0.00044588794, 1.1627396e-05, 0.057670258, 4.246253e-06, 6.174382e-06, 4.222969e-06, 0.009117617, 0.6239796, 3.5445457e-05, 0.00017580668, 1.2707611e-05, 8.81675e-07, 0.3666669, 6.16647e-07, 4.3034826e-05, 1.1155572e-06, 0.9716082, 0.0036068736, 0.00029337086, 0.000113066584, 0.00059629563, 1.5656573e-05, 0.02372147, 9.690217e-07, 1.378388e-06, 9.712984e-08, 0.1105977, 0.8890721, 4.8204645e-05, 3.7169924e-05, 0.0002424484, 2.9864822e-07, 4.902796e-07, 1.5026151e-08, 5.9766074e-07, 1.7625035e-07, 0.27393404, 0.7257613, 6.0214446e-05, 0.00011405888, 0.00012225499, 1.4559914e-06, 5.9373265e-06, 3.2447105e-08, 2.6817952e-08, 1.9999208e-08, 0.00013637744, 0.9135113, 2.4218636e-07, 3.201824e-07, 0.086351745, 4.2980517e-09, 5.0295816e-09, 2.8233047e-09, 0.0006709972, 4.0341e-05, 0.85213655, 0.09436244, 0.021573544, 0.002507777, 0.0023639384, 0.00019957365, 0.026104767, 4.0060215e-05, 1.2040656e-06, 5.3641735e-08, 0.2256023, 0.76613724, 0.00012398, 6.541455e-05, 0.00806873, 4.4204538e-07, 5.6476705e-07, 4.8697732e-08, 4.0559004e-05, 1.2895354e-06, 0.13935447, 0.23470876, 0.016103173, 0.006059173, 0.6034049, 9.316382e-06, 0.00031046954, 7.855437e-06, 4.9287573e-05, 1.0727578e-06, 0.15807185, 0.8390138, 0.0009805716, 0.0005327644, 0.00088317494, 4.255425e-06, 0.00046240495, 7.847849e-07, 3.1769723e-05, 1.3339641e-06, 0.4634989, 0.011107896, 0.5023665, 3.508458e-05, 0.022363823, 2.6286734e-05, 0.0005025035, 6.600347e-05, 7.5162056e-08, 5.3888712e-08, 0.3991555, 0.6007053, 5.19881e-05, 5.069957e-06, 8.031406e-05, 5.197525e-07, 1.135769e-06, 1.950354e-08, 3.5058285e-08, 2.6659237e-08, 0.0018089323, 0.99746084, 3.7529335e-05, 1.1204907e-06, 0.00068862905, 8.0413884e-07, 1.9952115e-06, 2.2501927e-08, 3.2673633e-08, 1.2915636e-09, 0.9924817, 0.0031439376, 0.004059353, 2.6027578e-09, 0.0002997318, 8.470597e-08, 8.1981225e-06, 6.9291964e-06, 3.1462416e-06, 1.5913636e-07, 0.01710981, 0.97929186, 0.00026076322, 2.3728294e-06, 0.0033292156, 3.3021465e-07, 2.3155342e-06, 4.8264152e-08, 4.82286e-10, 1.3997306e-10, 0.00040846158, 0.9983327, 2.9435267e-07, 1.4357024e-08, 0.0012586339, 1.6031172e-09, 1.0569556e-09, 8.9859686e-10, 8.6732534e-08, 6.643228e-09, 0.75877184, 0.24098857, 0.00016080732, 2.8336305e-07, 7.7741395e-05, 9.77266e-08, 6.172077e-07, 3.2260452e-08, 3.4826983e-06, 6.830953e-07, 0.7558993, 0.2424847, 0.0005995873, 0.000119913275, 0.0007165305, 1.6284379e-06, 0.00017393717, 1.7342293e-07, 2.4049295e-06, 1.3600355e-07, 0.7256288, 0.2741098, 1.2869033e-05, 2.2397038e-05, 0.00019238646, 2.3288433e-07, 3.0866846e-05, 6.76577e-09], 'name': 'fc10', 'shape': [20, 10], 'datatype': 'FP32'}, {'data': [1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 1, 0, 0, 0, 0], 'name': 'is_outlier', 'shape': [1, 20], 'datatype': 'INT64'}]}

```

```bash
seldon pipeline inspect cifar10-production.cifar10-drift.outputs.is_drift
```

```
seldon.default.model.cifar10-drift.outputs	cfekpiofh5ss739vr5u0	{"name":"is_drift","datatype":"INT64","shape":["1","1"],"contents":{"int64Contents":["1"]}}

```

```python
infer("cifar10-production.pipeline",1, "outlier")
```

```
![png](infer_files/infer_18_0.png)

```

```
<Response [200]>
{'model_name': '', 'outputs': [{'data': [2.544475e-06, 2.7632382e-07, 1.9947075e-07, 0.0002308716, 2.7517972e-08, 9.416619e-06, 0.9997557, 6.393948e-08, 3.56468e-07, 4.794506e-07], 'name': 'fc10', 'shape': [1, 10], 'datatype': 'FP32'}, {'data': [1], 'name': 'is_outlier', 'shape': [1, 1], 'datatype': 'INT64'}]}

```

```python
infer("cifar10-production.pipeline",1, "ok")
```

```
![png](infer_files/infer_19_0.png)

```

```
<Response [200]>
{'model_name': '', 'outputs': [{'data': [1.45001495e-08, 1.2525752e-09, 1.6298458e-07, 0.11529388, 1.7431412e-07, 6.1856604e-06, 0.8846994, 6.0739285e-09, 7.43792e-08, 4.7317337e-09], 'name': 'fc10', 'shape': [1, 10], 'datatype': 'FP32'}, {'data': [0], 'name': 'is_outlier', 'shape': [1, 1], 'datatype': 'INT64'}]}

```

Use the seldon CLI to look at the outputs from the CIFAR10 model. It will decode the Triton binary outputs for us.

```bash
seldon pipeline inspect cifar10-production.cifar10.outputs
```

```
seldon.default.model.cifar10.outputs	cfekplofh5ss739vr5v0	{"modelName":"cifar10_1","modelVersion":"1","outputs":[{"name":"fc10","datatype":"FP32","shape":["1","10"],"contents":{"fp32Contents":[1.45001495e-8,1.2525752e-9,1.6298458e-7,0.11529388,1.7431412e-7,0.0000061856604,0.8846994,6.0739285e-9,7.43792e-8,4.7317337e-9]}}]}

```

```bash
seldon pipeline unload cifar10-production
```

```json
{}

```

```bash
seldon model unload cifar10
seldon model unload cifar10-outlier
seldon model unload cifar10-drift
```

```json
{}
{}
{}

```

```python

```
