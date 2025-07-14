# Custom LightGBM Prepackaged Model Server

**Note**: Seldon has adopted the industry-standard Open Inference Protocol (OIP) and is no longer maintaining the Seldon and TensorFlow protocols. This transition allows for greater interoperability among various model serving runtimes, such as MLServer. To learn more about implementing OIP for model serving in Seldon Core 1, see [MLServer](https://docs.seldon.ai/mlserver).

We strongly encourage you to adopt the OIP, which provides seamless integration across diverse model serving runtimes, supports the development of versatile client and benchmarking tools, and ensures a high-performance, consistent, and unified inference experience.

In this notebook we create a new custom LIGHTGBM_SERVER prepackaged server with two versions:
   * A Seldon protocol LightGBM model server
   * A KfServing Open Inference protocol or V2 protocol version using MLServer for running lightgbm models

The Seldon model server is in defined in `lightgbmserver` folder.

## Prerequisites

 * A kubernetes cluster with kubectl configured
 * curl

## Setup Seldon Core

Use the setup notebook to [Setup Cluster](https://docs.seldon.io/projects/seldon-core/en/latest/examples/seldon_core_setup.html) to setup Seldon Core with an ingress - either Ambassador or Istio.

Then port-forward to that ingress on localhost:8003 in a separate terminal either with:

 * Ambassador: `kubectl port-forward $(kubectl get pods -n seldon -l app.kubernetes.io/name=ambassador -o jsonpath='{.items[0].metadata.name}') -n seldon 8003:8080`
 * Istio: `kubectl port-forward $(kubectl get pods -l istio=ingressgateway -n istio-system -o jsonpath='{.items[0].metadata.name}') -n istio-system 8003:8080`


```python
!kubectl create namespace seldon
```


```python
from IPython.core.magic import register_line_cell_magic


@register_line_cell_magic
def writetemplate(line, cell):
    with open(line, "w") as f:
        f.write(cell.format(**globals()))
```


```python
VERSION = !cat ../../../version.txt
VERSION = VERSION[0]
VERSION
```

## Training (can be skipped)


```python
TRAIN_MODEL = False
if TRAIN_MODEL:
    import os

    import joblib
    import lightgbm as lgb
    from sklearn import datasets
    from sklearn.datasets import load_iris
    from sklearn.model_selection import train_test_split

    model_dir = "./artifacts"
    BST_FILE = "model.txt"

    iris = load_iris()
    y = iris["target"]
    X = iris["data"]
    X_train, X_test, y_train, y_test = train_test_split(X, y, test_size=0.1)
    dtrain = lgb.Dataset(X_train, label=y_train)

    params = {"objective": "multiclass", "metric": "softmax", "num_class": 3}
    lgb_model = lgb.train(params=params, train_set=dtrain)
    model_file = os.path.join(model_dir, BST_FILE)
    lgb_model.save_model(model_file)
```

## Update Seldon Core with Custom Model


```python
%%writetemplate values.yaml
predictor_servers:
  MLFLOW_SERVER:
    protocols:
      seldon:
        defaultImageVersion: "{VERSION}"
        image: seldonio/mlflowserver
  SKLEARN_SERVER:
    protocols:
      seldon:
        defaultImageVersion: "{VERSION}"
        image: seldonio/sklearnserver
      kfserving:
        defaultImageVersion: "0.3.2"
        image: seldonio/mlserver
  TENSORFLOW_SERVER:
    protocols:
      seldon:
        defaultImageVersion: "{VERSION}"
        image: seldonio/tfserving-proxy
      tensorflow: 
        defaultImageVersion: 2.1.0
        image:  tensorflow/serving
  XGBOOST_SERVER:
    protocols:
      seldon:
        defaultImageVersion: "{VERSION}"
        image: seldonio/xgboostserver
      kfserving:
        defaultImageVersion: "0.3.2"
        image: seldonio/mlserver
  LIGHTGBM_SERVER:
    protocols:
      seldon:
        defaultImageVersion: "{VERSION}"
        image: seldonio/lighgbmserver
      kfserving:
        defaultImageVersion: "0.3.2"
        image: seldonio/mlserver
  TRITON_SERVER:
    protocols:
      kfserving:
        defaultImageVersion: "21.08-py3"
        image: nvcr.io/nvidia/tritonserver
  TEMPO_SERVER:
    protocols:
      kfserving:
        defaultImageVersion: "0.3.2"
        image: seldonio/mlserver

```


```python
!helm upgrade seldon-core  \
    ../../../helm-charts/seldon-core-operator \
    --namespace seldon-system \
    --values values.yaml \
    --set istio.enabled=true
```

## DeployLightGBM Model with Seldon Protocol


```python
!cat model_seldon_v1.yaml
```

Wait for new webhook certificates to be loaded


```python
import time

time.sleep(60)
```


```python
!kubectl create -f model_seldon_v1.yaml -n seldon
```


```python
!kubectl rollout status deploy/$(kubectl get deploy -l seldon-deployment-id=iris -o jsonpath='{.items[0].metadata.name}' -n seldon) -n seldon
```


```python
for i in range(60):
    state = !kubectl get sdep iris -n seldon -o jsonpath='{.status.state}'
    state = state[0]
    print(state)
    if state == "Available":
        break
    time.sleep(1)
assert state == "Available"
```


```python
import json
X=!curl -s -d '{"data": {"ndarray":[[1.0, 2.0, 3.0, 4.0]]}}' \
   -X POST http://localhost:8003/seldon/seldon/iris/api/v1.0/predictions \
   -H "Content-Type: application/json"
d=json.loads(X[0])
print(d)
```


```python
!kubectl delete -f model_seldon_v1.yaml
```

## Deploy Model with KFserving Protocol


```python
!cat model_seldon_v2.yaml
```


```python
!kubectl create -f model_seldon_v2.yaml -n seldon
```


```python
!kubectl rollout status deploy/$(kubectl get deploy -l seldon-deployment-id=iris -o jsonpath='{.items[0].metadata.name}' -n seldon) -n seldon
```


```python
for i in range(60):
    state = !kubectl get sdep iris -n seldon -o jsonpath='{.status.state}'
    state = state[0]
    print(state)
    if state == "Available":
        break
    time.sleep(1)
assert state == "Available"
```


```python
import json
X=!curl -s -d '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}'\
   -X POST http://localhost:8003/seldon/seldon/iris/v2/models/infer \
   -H "Content-Type: application/json"
d=json.loads(X[0])
print(d)
```


```python
!kubectl delete -f model_seldon_v2.yaml
```


```python

```
