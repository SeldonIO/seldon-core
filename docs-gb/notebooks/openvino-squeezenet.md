# OpenVINO example with Squeezenet Model

This notebook illustrates how you can serve [OpenVINO](https://software.intel.com/en-us/openvino-toolkit) optimized models for Imagenet with Seldon Core.

![car](car.png)

   
To run all of the notebook successfully you will need to start it with
```
jupyter notebook --NotebookApp.iopub_data_rate_limit=100000000
```

## Setup Seldon Core

Use the setup notebook to [Setup Cluster](../notebooks/seldon-core-setup.md#setup-cluster) with [Ambassador Ingress](../notebooks/seldon-core-setup.md#ambassador) and [Install Seldon Core](../notebooks/seldon-core-setup.md#Install-Seldon-Core). Instructions [also online](../notebooks/seldon-core-setup.md).


```python
!kubectl create namespace seldon
```


```python
!kubectl config set-context $(kubectl config current-context) --namespace=seldon
```

## Deploy Seldon Intel OpenVINO Graph


```python
!helm install openvino-squeezenet ../../../helm-charts/seldon-openvino \
    --set openvino.model.src=gs://seldon-models/openvino/squeezenet \
    --set openvino.model.path=/opt/ml/squeezenet \
    --set openvino.model.name=squeezenet1.1 \
    --set openvino.model.input=data \
    --set openvino.model.output=prob 
```


```python
!helm template openvino-squeezenet ../../../helm-charts/seldon-openvino \
    --set openvino.model.src=gs://seldon-models/openvino/squeezenet \
    --set openvino.model.path=/opt/ml/squeezenet \
    --set openvino.model.name=squeezenet1.1 \
    --set openvino.model.input=data \
    --set openvino.model.output=prob | pygmentize -l json
```


```python
!kubectl rollout status deploy/$(kubectl get deploy -l seldon-deployment-id=openvino-model -o jsonpath='{.items[0].metadata.name}')
```

## Test



```python
%matplotlib inline
import json
import sys

import matplotlib.pyplot as plt
import numpy as np
from keras.applications.imagenet_utils import decode_predictions, preprocess_input
from keras.preprocessing import image

from seldon_core.seldon_client import SeldonClient


def getImage(path):
    img = image.load_img(path, target_size=(227, 227))
    x = image.img_to_array(img)
    plt.imshow(x / 255.0)
    x = np.expand_dims(x, axis=0)
    x = preprocess_input(x)
    return x


X = getImage("car.png")
X = X.transpose((0, 3, 1, 2))
print(X.shape)

sc = SeldonClient(deployment_name="openvino-model", namespace="seldon")

response = sc.predict(
    gateway="ambassador", transport="grpc", data=X, client_return_type="proto"
)

result = response.response.data.tensor.values

result = np.array(result)
result = result.reshape(1, 1000)

with open("imagenet_classes.json") as f:
    cnames = eval(f.read())

    for i in range(result.shape[0]):
        single_result = result[[i], ...]
        ma = np.argmax(single_result)
        print("\t", i, cnames[ma])
        assert cnames[ma] == "sports car, sport car"
```


```python
!helm delete openvino-squeezenet
```


```python

```
