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
2022-10-06 16:32:05.092402: W tensorflow/stream_executor/platform/default/dso_loader.cc:64] Could not load dynamic library 'libcudart.so.11.0'; dlerror: libcudart.so.11.0: cannot open shared object file: No such file or directory
2022-10-06 16:32:05.092425: I tensorflow/stream_executor/cuda/cudart_stub.cc:29] Ignore above cudart dlerror if you do not have a GPU set up on your machine.

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
        "uid": "ccvfa5t4nntrbfkuk5hg",
        "version": 2,
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
        "pipelineVersion": 2,
        "status": "PipelineReady",
        "reason": "created pipeline",
        "lastChangeTimestamp": "2022-10-06T15:32:39.926431438Z"
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
{'model_name': '', 'outputs': [{'data': [1.4500082e-08, 1.2525715e-09, 1.6298364e-07, 0.11529311, 1.7431313e-07, 6.1856485e-06, 0.88470024, 6.073899e-09, 7.437885e-08, 4.73172e-09, 1.2644817e-06, 4.88144e-09, 1.5153264e-09, 8.4905425e-09, 5.513056e-10, 1.1617127e-09, 5.7728744e-10, 2.8839528e-07, 0.0006148928, 0.99938357, 0.88874614, 2.5331808e-06, 0.00012967634, 0.105315745, 2.4284038e-05, 6.3332695e-06, 0.0016261397, 1.1307816e-05, 0.0013286571, 0.002809183, 2.0993398e-06, 3.680442e-08, 0.0013269989, 2.1766682e-05, 0.99841356, 0.00015300752, 6.947217e-06, 1.3277109e-05, 6.18605e-05, 3.4072772e-07, 1.1205097e-05, 0.99997175, 1.9948209e-07, 6.9880706e-08, 3.338707e-08, 5.260304e-08, 3.0352362e-07, 4.37389e-08, 5.3243895e-07, 1.5870555e-05, 0.00065251003, 0.013322163, 1.4803151e-06, 0.97663224, 4.9847342e-05, 0.00058076304, 0.008405762, 5.223446e-06, 0.00023390213, 0.0001160483, 1.6682318e-06, 5.7737415e-10, 0.9975605, 6.4556276e-05, 0.0023719606, 1.03926354e-07, 9.7479244e-08, 1.4484527e-07, 8.762438e-07, 2.4758279e-08, 5.0287703e-09, 6.8563676e-11, 5.993221e-12, 4.921233e-10, 1.4711688e-07, 2.7940907e-06, 3.456325e-09, 0.99999714, 5.9420524e-10, 9.445044e-11, 4.1854808e-05, 5.041549e-08, 8.0302314e-08, 1.2119865e-07, 6.781646e-09, 1.2616152e-08, 1.1878551e-08, 1.628573e-09, 0.9999578, 3.281738e-08, 0.08930252, 1.4065115e-07, 4.111721e-07, 0.9089835, 8.9333895e-07, 0.0015637465, 0.00013868857, 9.093003e-06, 4.875963e-07, 4.397615e-07, 0.00016094722, 3.5653608e-07, 0.076052375, 0.89274454, 0.001177757, 0.0026557152, 0.027188947, 4.18919e-06, 1.3293908e-05, 1.8564508e-06, 1.3373815e-06, 1.0251267e-07, 8.651978e-09, 4.4582193e-06, 1.4646363e-05, 1.2609522e-06, 1.0460891e-08, 0.9998946, 8.332422e-05, 3.9008867e-07, 6.53852e-05, 3.0121907e-08, 1.02472164e-07, 1.8824388e-06, 0.0004958507, 3.5334717e-05, 2.739997e-07, 0.99939275, 4.840291e-06, 3.5346593e-06, 0.00055180624, 3.15972e-07, 0.99902296, 0.0003150998, 8.0788834e-07, 1.6366099e-06, 2.7955884e-06, 6.112407e-06, 9.817333e-05, 2.6027215e-07, 0.00045619791, 5.3606173e-06, 2.8656412e-05, 0.00011604099, 6.881177e-05, 8.844783e-06, 4.4656073e-05, 3.5564542e-05, 0.0065643936, 0.9926715, 0.007300965, 1.7669344e-06, 3.0520675e-07, 0.026906408, 1.3769774e-06, 0.00027539747, 5.5836076e-06, 3.7925556e-06, 0.00038767883, 0.9651167, 0.1811405, 2.8360253e-05, 0.00019927393, 0.007685893, 0.00014663483, 3.9361246e-05, 5.94171e-05, 7.361781e-05, 0.7993662, 0.011260739, 2.3992737e-11, 7.6336457e-16, 1.4644799e-15, 1, 2.4652111e-14, 1.1785965e-10, 1.9402042e-13, 4.2408475e-15, 1.2092895e-15, 2.9042784e-15, 1.536693e-08, 1.2476195e-09, 1.3560074e-07, 0.999997, 4.3113096e-11, 2.8163587e-08, 2.4494798e-06, 1.3122778e-10, 3.808112e-07, 2.1628116e-11, 0.0004926233, 6.9424455e-06, 2.827208e-05, 0.9253418, 9.500489e-06, 0.0003613391, 0.07271345, 1.2831038e-07, 0.001045708, 2.851437e-07], 'name': 'fc10', 'shape': [20, 10], 'datatype': 'FP32'}, {'data': [0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0], 'name': 'is_outlier', 'shape': [1, 20], 'datatype': 'INT64'}]}

```

```bash
seldon pipeline inspect cifar10-production.cifar10-drift.outputs.is_drift
```

```
seldon.default.model.cifar10-drift.outputs	ccvfa7pi3pn0vadt0di0	{"name":"is_drift", "datatype":"INT64", "shape":["1"], "contents":{"int64Contents":["0"]}}

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
{'model_name': '', 'outputs': [{'data': [1.6211126e-09, 4.0089623e-10, 0.00045838102, 0.93237555, 4.15139e-07, 8.396934e-06, 0.067157224, 9.056127e-10, 2.3455957e-09, 1.2709092e-10, 0.0013329195, 8.31132e-06, 0.92835444, 0.0049285158, 0.0068845307, 0.0003594609, 0.00044588748, 1.1627344e-05, 0.057670027, 4.2462384e-06, 6.1743613e-06, 4.222963e-06, 0.009117634, 0.6239805, 3.544544e-05, 0.00017580659, 1.2707604e-05, 8.816747e-07, 0.36666605, 6.166467e-07, 4.3034703e-05, 1.115552e-06, 0.9716082, 0.003606853, 0.0002933709, 0.000113066366, 0.00059629255, 1.5656527e-05, 0.023721391, 9.69018e-07, 1.3783852e-06, 9.712964e-08, 0.11059789, 0.88907194, 4.8204773e-05, 3.7169844e-05, 0.00024244905, 2.9864788e-07, 4.902785e-07, 1.5026147e-08, 5.976571e-07, 1.7624993e-07, 0.27393368, 0.72576165, 6.0214476e-05, 0.00011405839, 0.00012225457, 1.455988e-06, 5.937307e-06, 3.244706e-08, 2.6818164e-08, 1.9999252e-08, 0.00013637826, 0.9135115, 2.4218687e-07, 3.2018372e-07, 0.08635152, 4.298069e-09, 5.029612e-09, 2.8233054e-09, 0.00067099655, 4.0340925e-05, 0.85213655, 0.09436264, 0.021573491, 0.0025077735, 0.0023639349, 0.00019957346, 0.026104674, 4.006014e-05, 1.2040629e-06, 5.3641614e-08, 0.22560245, 0.766137, 0.00012397974, 6.5414526e-05, 0.008068763, 4.4204523e-07, 5.647658e-07, 4.8697718e-08, 4.0559004e-05, 1.2895365e-06, 0.13935485, 0.23470965, 0.016103141, 0.0060591903, 0.6034037, 9.316398e-06, 0.00031046892, 7.855436e-06, 4.9287453e-05, 1.072755e-06, 0.15807153, 0.8390142, 0.0009805678, 0.00053276104, 0.000883172, 4.255407e-06, 0.00046240314, 7.847815e-07, 3.1769618e-05, 1.3339647e-06, 0.46349868, 0.011107955, 0.5023667, 3.5084562e-05, 0.022363717, 2.6286772e-05, 0.00050250255, 6.600338e-05, 7.516225e-08, 5.3888748e-08, 0.39915618, 0.6007046, 5.1988183e-05, 5.0699696e-06, 8.031388e-05, 5.197544e-07, 1.135772e-06, 1.9503554e-08, 3.505842e-08, 2.6659439e-08, 0.0018089443, 0.99746084, 3.7529408e-05, 1.120497e-06, 0.00068863365, 8.041411e-07, 1.995225e-06, 2.2502057e-08, 3.267345e-08, 1.291554e-09, 0.9924819, 0.003143899, 0.0040593147, 2.6027382e-09, 0.00029972926, 8.470566e-08, 8.198109e-06, 6.9291314e-06, 3.146221e-06, 1.5913548e-07, 0.017109655, 0.979292, 0.00026076124, 2.372814e-06, 0.0033292002, 3.302128e-07, 2.3155103e-06, 4.8263697e-08, 4.8228505e-10, 1.3997253e-10, 0.00040846117, 0.9983327, 2.9435154e-07, 1.4356915e-08, 0.001258629, 1.6031141e-09, 1.0569535e-09, 8.9859686e-10, 8.67323e-08, 6.6431984e-09, 0.75877273, 0.2409877, 0.0001608066, 2.8336203e-07, 7.774119e-05, 9.772625e-08, 6.1720607e-07, 3.2260367e-08, 3.4826974e-06, 6.8309384e-07, 0.75589913, 0.24248488, 0.000599586, 0.00011991303, 0.00071652926, 1.6284345e-06, 0.0001739368, 1.7342272e-07, 2.4049314e-06, 1.3600392e-07, 0.7256294, 0.27410924, 1.2869068e-05, 2.2397078e-05, 0.000192387, 2.3288452e-07, 3.0866962e-05, 6.7657884e-09], 'name': 'fc10', 'shape': [20, 10], 'datatype': 'FP32'}, {'data': [1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 1, 0, 0, 0, 0], 'name': 'is_outlier', 'shape': [1, 20], 'datatype': 'INT64'}]}

```

```bash
seldon pipeline inspect cifar10-production.cifar10-drift.outputs.is_drift
```

```
seldon.default.model.cifar10-drift.outputs	ccvfac1i3pn0vadt0dig	{"name":"is_drift", "datatype":"INT64", "shape":["1"], "contents":{"int64Contents":["1"]}}

```

```python
infer("cifar10-production.pipeline",1, "outlier")
```

```
![png](infer_files/infer_18_0.png)

```

```
<Response [200]>
{'model_name': '', 'outputs': [{'data': [0.0007882897, 0.00011240715, 8.967017e-05, 0.015603337, 3.67151e-05, 0.008521079, 0.97396415, 6.24793e-06, 0.0008641859, 1.3902612e-05], 'name': 'fc10', 'shape': [1, 10], 'datatype': 'FP32'}, {'data': [1], 'name': 'is_outlier', 'shape': [1, 1], 'datatype': 'INT64'}]}

```

```python
infer("cifar10-production.pipeline",1, "ok")
```

```
![png](infer_files/infer_19_0.png)

```

```
<Response [200]>
{'model_name': '', 'outputs': [{'data': [1.4500107e-08, 1.2525738e-09, 1.6298331e-07, 0.115293205, 1.7431327e-07, 6.185636e-06, 0.8847001, 6.0738867e-09, 7.437898e-08, 4.7317195e-09], 'name': 'fc10', 'shape': [1, 10], 'datatype': 'FP32'}, {'data': [0], 'name': 'is_outlier', 'shape': [1, 1], 'datatype': 'INT64'}]}

```

Use the seldon CLI to look at the outputs from the CIFAR10 model. It will decide the Triton binary outputs for us.

```bash
seldon pipeline inspect cifar10-production.cifar10.outputs
```

```
seldon.default.model.cifar10.outputs	ccvfafpi3pn0vadt0djg	{"modelName":"cifar10_1", "modelVersion":"1", "outputs":[{"name":"fc10", "datatype":"FP32", "shape":["1", "10"], "contents":{"fp32Contents":[1.4500107e-8, 1.2525738e-9, 1.6298331e-7, 0.115293205, 1.7431327e-7, 0.000006185636, 0.8847001, 6.0738867e-9, 7.437898e-8, 4.7317195e-9]}}]}

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
