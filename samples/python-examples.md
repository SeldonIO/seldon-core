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
2022-11-16 19:28:57.916579: I tensorflow/core/platform/cpu_feature_guard.cc:193] This TensorFlow binary is optimized with oneAPI Deep Neural Network Library (oneDNN) to use the following CPU instructions in performance-critical operations:  AVX2 AVX_VNNI FMA
To enable them in other operations, rebuild TensorFlow with the appropriate compiler flags.
2022-11-16 19:28:58.079213: I tensorflow/core/util/util.cc:169] oneDNN custom operations are on. You may see slightly different numerical results due to floating-point round-off errors from different computation orders. To turn them off, set the environment variable `TF_ENABLE_ONEDNN_OPTS=0`.
2022-11-16 19:28:58.123952: W tensorflow/stream_executor/platform/default/dso_loader.cc:64] Could not load dynamic library 'libcudart.so.11.0'; dlerror: libcudart.so.11.0: cannot open shared object file: No such file or directory
2022-11-16 19:28:58.123978: I tensorflow/stream_executor/cuda/cudart_stub.cc:29] Ignore above cudart dlerror if you do not have a GPU set up on your machine.
2022-11-16 19:28:58.159891: E tensorflow/stream_executor/cuda/cuda_blas.cc:2981] Unable to register cuBLAS factory: Attempting to register factory for plugin cuBLAS when one has already been registered
2022-11-16 19:28:58.870452: W tensorflow/stream_executor/platform/default/dso_loader.cc:64] Could not load dynamic library 'libnvinfer.so.7'; dlerror: libnvinfer.so.7: cannot open shared object file: No such file or directory
2022-11-16 19:28:58.870490: W tensorflow/stream_executor/platform/default/dso_loader.cc:64] Could not load dynamic library 'libnvinfer_plugin.so.7'; dlerror: libnvinfer_plugin.so.7: cannot open shared object file: No such file or directory
2022-11-16 19:28:58.870493: W tensorflow/compiler/tf2tensorrt/utils/py_utils.cc:38] TF-TRT Warning: Cannot dlopen some TensorRT libraries. If you would like to use Nvidia GPU with TensorRT, please make sure the missing libraries mentioned above are installed properly.

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
Downloading data from https://www.cs.toronto.edu/~kriz/cifar-10-python.tar.gz
170498071/170498071 [==============================] - 9s 0us/step
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

```
[1;39m{}[0m

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
{'model_name': 'cifar10_1', 'model_version': '1', 'outputs': [{'name': 'fc10', 'datatype': 'FP32', 'shape': [2, 10], 'data': [1.4500082023971572e-08, 1.252571490972798e-09, 1.6298334060138586e-07, 0.11529310792684555, 1.7431312926419196e-07, 6.185642178024864e-06, 0.8847002387046814, 6.0738876150878696e-09, 7.437885329864002e-08, 4.731711022998297e-09, 1.2644841262954287e-06, 4.881440140991344e-09, 1.51533230408063e-09, 8.490559366691741e-09, 5.51305612273012e-10, 1.1617149464626664e-09, 5.772873845621973e-10, 2.8839554033766035e-07, 0.0006148957181721926, 0.9993835687637329]}]}

```

```bash
seldon model unload cifar10
```

```json
{}

```

```python

```
