# Python example

We will test a simple Pipeline with a cifar10 image classification model with batch requests. We assume a locally running Seldon.

```python
import requests
import json
from typing import Dict, List
import numpy as np
import os
import tensorflow as tf
tf.keras.backend.clear_session()
```

```
2023-03-10 10:17:52.202780: W tensorflow/stream_executor/platform/default/dso_loader.cc:64] Could not load dynamic library 'libcudart.so.11.0'; dlerror: libcudart.so.11.0: cannot open shared object file: No such file or directory
2023-03-10 10:17:52.202792: I tensorflow/stream_executor/cuda/cudart_stub.cc:29] Ignore above cudart dlerror if you do not have a GPU set up on your machine.

```

```python
train, test = tf.keras.datasets.cifar10.load_data()
X_train, y_train = train
X_test, y_test = test

X_train = X_train.astype('float32') / 255
X_test = X_test.astype('float32') / 255
print(X_train.shape, y_train.shape, X_test.shape, y_test.shape)
```

```
(50000, 32, 32, 3) (50000, 1) (10000, 32, 32, 3) (10000, 1)

```

```bash
cat ./models/cifar10.yaml
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

```

```bash
seldon model load -f ./models/cifar10.yaml
```

```json
{}

```

```bash
seldon model status cifar10 -w ModelAvailable | jq .
```

```json
{}

```

```python
headers = {"Content-Type": "application/json", "seldon-model":"cifar10"}
reqJson = json.loads('{"inputs":[{"name":"input_1","data":[],"datatype":"FP32","shape":[]}]}')
url = "http://0.0.0.0:9000/v2/models/cifar10/infer"
```

```python
batchSz = 2
rows = X_train[0:0+batchSz]
reqJson["inputs"][0]["data"] = rows.flatten().tolist()
reqJson["inputs"][0]["shape"] = [batchSz, 32, 32, 3]
```

```python
response_raw = requests.post(url, json=reqJson, headers=headers)
print(response_raw)
print(response_raw.json())
```

```
<Response [200]>
{'model_name': 'cifar10_1', 'model_version': '1', 'outputs': [{'name': 'fc10', 'datatype': 'FP32', 'shape': [2, 10], 'data': [1.4500123768357298e-08, 1.2525751547087793e-09, 1.6298442062634422e-07, 0.11529377847909927, 1.7431396770462015e-07, 6.185648544487776e-06, 0.8846994638442993, 6.073928471295176e-09, 7.437921567543526e-08, 4.7317341156372095e-09, 1.2644900380109902e-06, 4.881486770358379e-09, 1.51534673697995e-09, 8.490656178139488e-09, 5.513119405442524e-10, 1.161723717224561e-09, 5.772940459003451e-10, 2.883980414480902e-07, 0.0006149015971459448, 0.9993835687637329]}]}

```

```bash
seldon model unload cifar10
```

```json
{}

```

```python

```
