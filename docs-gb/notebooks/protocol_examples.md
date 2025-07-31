# Basic Examples with Different Protocols

## Prerequisites

 * A kubernetes cluster with kubectl configured
 * curl
 * grpcurl
 * pygmentize
 
## Examples

  * [Open Inference Protocol or V2 Protocol](#V2-Protocol-Model)
  * [Seldon Protocol](#Seldon-Protocol-Model)
  * [Tensorflow Protocol](#Tensorflow-Protocol-Model)

**Note**:Seldon has adopted the industry-standard Open Inference Protocol (OIP) and is no longer maintaining the Seldon and TensorFlow protocols. This transition allows for greater interoperability among various model serving runtimes, such as MLServer. To learn more about implementing OIP for model serving in Seldon Core 1, see [MLServer](https://docs.seldon.ai/mlserver).

We strongly encourage you to adopt the OIP, which provides seamless integration across diverse model serving runtimes, supports the development of versatile client and benchmarking tools, and ensures a high-performance, consistent, and unified inference experience.

## Setup Seldon Core

Use the setup notebook to [Setup Cluster](../notebooks/seldon-core-setup.md) to setup Seldon Core with an ingress - either Ambassador or Istio.

Then port-forward to that ingress on localhost:8003 in a separate terminal either with:

 * Ambassador: `kubectl port-forward $(kubectl get pods -n seldon -l app.kubernetes.io/name=ambassador -o jsonpath='{.items[0].metadata.name}') -n seldon 8003:8080`
 * Istio: `kubectl port-forward $(kubectl get pods -l istio=ingressgateway -n istio-system -o jsonpath='{.items[0].metadata.name}') -n istio-system 8003:8080`


```python
!kubectl create namespace seldon
```

    Error from server (AlreadyExists): namespaces "seldon" already exists



```python
!kubectl config set-context $(kubectl config current-context) --namespace=seldon
```

    Context "kind-ansible" modified.



```python
import json
import time
```


```python
from IPython.core.magic import register_line_cell_magic

@register_line_cell_magic
def writetemplate(line, cell):
    with open(line, 'w') as f:
        f.write(cell.format(**globals()))
```


```python
VERSION=!cat ../version.txt
VERSION=VERSION[0]
VERSION
```




    '1.17.0-dev'



## Seldon Protocol Model

We will deploy a REST model that uses the SELDON Protocol namely by specifying the attribute `protocol: seldon`


```python
%%writetemplate resources/model_seldon.yaml
apiVersion: machinelearning.seldon.io/v1
kind: SeldonDeployment
metadata:
  name: example-seldon
spec:
  protocol: seldon
  predictors:
  - componentSpecs:
    - spec:
        containers:
        - image: seldonio/mock_classifier:{VERSION}
          name: classifier
    graph:
      name: classifier
      type: MODEL
    name: model
    replicas: 1
```


```python
!kubectl apply -f resources/model_seldon.yaml
```

    seldondeployment.machinelearning.seldon.io/example-seldon created



```python
!kubectl wait --for condition=ready --timeout=300s sdep --all -n seldon
```

    seldondeployment.machinelearning.seldon.io/example-seldon condition met



```python
X=!curl -s -d '{"data": {"ndarray":[[1.0, 2.0, 5.0]]}}' \
   -X POST http://localhost:8003/seldon/seldon/example-seldon/api/v1.0/predictions \
   -H "Content-Type: application/json"
d=json.loads(X[0])
print(d)
assert(d["data"]["ndarray"][0][0] > 0.4)
```

    {'data': {'names': ['proba'], 'ndarray': [[0.43782349911420193]]}, 'meta': {'requestPath': {'classifier': 'seldonio/mock_classifier:1.17.0-dev'}}}



```python
X=!cd ../executor/proto && grpcurl -d '{"data":{"ndarray":[[1.0,2.0,5.0]]}}' \
         -rpc-header seldon:example-seldon -rpc-header namespace:seldon \
         -plaintext \
         -proto ./prediction.proto  0.0.0.0:8003 seldon.protos.Seldon/Predict
d=json.loads("".join(X))
print(d)
assert(d["data"]["ndarray"][0][0] > 0.4)
```

    {'meta': {'requestPath': {'classifier': 'seldonio/mock_classifier:1.17.0-dev'}}, 'data': {'names': ['proba'], 'ndarray': [[0.43782349911420193]]}}



```python
!kubectl delete -f resources/model_seldon.yaml
```

    seldondeployment.machinelearning.seldon.io "example-seldon" deleted


## Seldon protocol Model with ModelUri with two custom models


```python
%%writetemplate resources/model_seldon.yaml
apiVersion: machinelearning.seldon.io/v1
kind: SeldonDeployment
metadata:
  name: example-seldon
spec:
  protocol: seldon
  predictors:
  - componentSpecs:
    - spec:
        containers:
        - image: seldonio/mock_classifier:{VERSION}
          name: classifier
        - image: seldonio/mock_classifier:{VERSION}
          name: classifier2
    graph:
      name: classifier
      type: MODEL
      modelUri: gs://seldon-models/v1.17.0-dev/sklearn/iris
      children:
      - name: classifier2
        type: MODEL
        modelUri: gs://seldon-models/v1.17.0-dev/sklearn/iris
    name: model
    replicas: 1
```


```python
!kubectl apply -f resources/model_seldon.yaml
```

    seldondeployment.machinelearning.seldon.io/example-seldon unchanged



```python
!kubectl wait --for condition=ready --timeout=300s sdep --all -n seldon
```

    seldondeployment.machinelearning.seldon.io/example-seldon condition met



```python
X=!curl -s -d '{"data": {"ndarray":[[1.0, 2.0, 5.0]]}}' \
   -X POST http://localhost:8003/seldon/seldon/example-seldon/api/v1.0/predictions \
   -H "Content-Type: application/json"
d=json.loads(X[0])
print(d)
```

    {'data': {'names': ['proba'], 'ndarray': [[0.07735472603574542]]}, 'meta': {'requestPath': {'classifier': 'seldonio/mock_classifier:1.17.0-dev', 'classifier2': 'seldonio/mock_classifier:1.17.0-dev'}}}



```python
X=!cd ../executor/proto && grpcurl -d '{"data":{"ndarray":[[1.0,2.0,5.0]]}}' \
         -rpc-header seldon:example-seldon -rpc-header namespace:seldon \
         -plaintext \
         -proto ./prediction.proto  0.0.0.0:8003 seldon.protos.Seldon/Predict
d=json.loads("".join(X))
print(d)
```

    {'meta': {'requestPath': {'classifier': 'seldonio/mock_classifier:1.17.0-dev', 'classifier2': 'seldonio/mock_classifier:1.17.0-dev'}}, 'data': {'names': ['proba'], 'ndarray': [[0.07735472603574542]]}}



```python
!kubectl delete -f resources/model_seldon.yaml
```

    seldondeployment.machinelearning.seldon.io "example-seldon" deleted


## Tensorflow Protocol Model
We will deploy a model that uses the TENSORLFOW Protocol namely by specifying the attribute `protocol: tensorflow`


```python
%%writefile resources/model_tfserving.yaml
apiVersion: machinelearning.seldon.io/v1
kind: SeldonDeployment
metadata:
  name: example-tfserving
spec:
  protocol: tensorflow
  predictors:
  - componentSpecs:
    - spec:
        containers:
        - args: 
          - --port=8500
          - --rest_api_port=8501
          - --model_name=halfplustwo
          - --model_base_path=gs://seldon-models/tfserving/half_plus_two
          image: tensorflow/serving
          name: halfplustwo
          ports:
          - containerPort: 8501
            name: http
            protocol: TCP
          - containerPort: 8500
            name: grpc
            protocol: TCP
    graph:
      name: halfplustwo
      type: MODEL
      endpoint:
        httpPort: 8501
        grpcPort: 8500
    name: model
    replicas: 1
```

    Writing resources/model_tfserving.yaml



```python
!kubectl apply -f resources/model_tfserving.yaml
```

    seldondeployment.machinelearning.seldon.io/example-tfserving created



```python
!kubectl wait --for condition=ready --timeout=300s sdep --all -n seldon
```

    seldondeployment.machinelearning.seldon.io/example-tfserving condition met



```python
X=!curl -s -d '{"instances": [1.0, 2.0, 5.0]}' \
   -X POST http://localhost:8003/seldon/seldon/example-tfserving/v1/models/halfplustwo/:predict \
   -H "Content-Type: application/json"
d=json.loads("".join(X))
print(d)
assert(d["predictions"][0] == 2.5)
```

    {'predictions': [2.5, 3.0, 4.5]}



```python
X=!cd ../executor/proto && grpcurl \
   -d '{"model_spec":{"name":"halfplustwo"},"inputs":{"x":{"dtype": 1, "tensor_shape": {"dim":[{"size": 3}]}, "floatVal" : [1.0, 2.0, 3.0]}}}' \
   -rpc-header seldon:example-tfserving -rpc-header namespace:seldon \
   -plaintext -proto ./prediction_service.proto \
   0.0.0.0:8003 tensorflow.serving.PredictionService/Predict
d=json.loads("".join(X))
print(d)
assert(d["outputs"]["x"]["floatVal"][0] == 2.5)
```

    {'outputs': {'x': {'dtype': 'DT_FLOAT', 'tensorShape': {'dim': [{'size': '3'}]}, 'floatVal': [2.5, 3, 3.5]}}, 'modelSpec': {'name': 'halfplustwo', 'version': '123', 'signatureName': 'serving_default'}}



```python
!kubectl delete -f resources/model_tfserving.yaml
```

    seldondeployment.machinelearning.seldon.io "example-tfserving" deleted


## V2 Protocol Model

We will deploy a REST model that uses the V2 Protocol namely by specifying the attribute `protocol: v2`


```python
%%writefile resources/model_v2.yaml
apiVersion: machinelearning.seldon.io/v1alpha2
kind: SeldonDeployment
metadata:
  name: triton
spec:
  protocol: v2
  predictors:
  - graph:
      children: []
      implementation: TRITON_SERVER
      modelUri: gs://seldon-models/trtis/simple-model
      name: simple
    name: simple
    replicas: 1
```

    Writing resources/model_v2.yaml



```python
!kubectl apply -f resources/model_v2.yaml
```

    seldondeployment.machinelearning.seldon.io/triton created



```python
!kubectl wait --for condition=ready --timeout=300s sdep --all -n seldon
```

    seldondeployment.machinelearning.seldon.io/triton condition met



```python
X=!curl -s -d '{"inputs":[{"name":"INPUT0","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,16]},{"name":"INPUT1","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,16]}]}'  \
        -X POST http://0.0.0.0:8003/seldon/seldon/triton/v2/models/simple/infer \
        -H "Content-Type: application/json"
d=json.loads(X[0])
print(d)
assert(d["outputs"][0]["data"][0]==2)
```

    {'model_name': 'simple', 'model_version': '1', 'outputs': [{'name': 'OUTPUT0', 'datatype': 'INT32', 'shape': [1, 16], 'data': [2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32]}, {'name': 'OUTPUT1', 'datatype': 'INT32', 'shape': [1, 16], 'data': [0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0]}]}



```python
X=!cd ../executor/api/grpc/kfserving/inference && \
        grpcurl -d '{"model_name":"simple","inputs":[{"name":"INPUT0","contents":{"int_contents":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]},"datatype":"INT32","shape":[1,16]},{"name":"INPUT1","contents":{"int_contents":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]},"datatype":"INT32","shape":[1,16]}]}' \
        -plaintext -proto ./grpc_service.proto \
        -rpc-header seldon:triton -rpc-header namespace:seldon \
        0.0.0.0:8003 inference.GRPCInferenceService/ModelInfer
X="".join(X)
print(X)
```

    {  "modelName": "simple",  "modelVersion": "1",  "outputs": [    {      "name": "OUTPUT0",      "datatype": "INT32",      "shape": [        "1",        "16"      ]    },    {      "name": "OUTPUT1",      "datatype": "INT32",      "shape": [        "1",        "16"      ]    }  ],  "rawOutputContents": [    "AgAAAAQAAAAGAAAACAAAAAoAAAAMAAAADgAAABAAAAASAAAAFAAAABYAAAAYAAAAGgAAABwAAAAeAAAAIAAAAA==",    "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=="  ]}



```python
!kubectl delete -f resources/model_v2.yaml
```

    seldondeployment.machinelearning.seldon.io "triton" deleted



```python
!kubectl delete -f resources/model_tfserving.yaml
```

    seldondeployment.machinelearning.seldon.io "example-tfserving" deleted



```python

```
