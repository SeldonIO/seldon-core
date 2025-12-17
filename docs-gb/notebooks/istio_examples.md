# Example Seldon Core Deployments using Helm with Istio

Prequisites

 * [Install istio](https://istio.io/latest/docs/setup/getting-started/#download)

## Setup Cluster and Ingress

Use the setup notebook to [Setup Cluster](https://docs.seldon.ai/seldon-core-1/tutorials/notebooks/seldon-core-setup#setup-cluster) with [Istio Ingress](https://docs.seldon.ai/seldon-core-1/tutorials/notebooks/seldon-core-setup#istio).


```python
!kubectl create namespace seldon
```

    Error from server (AlreadyExists): namespaces "seldon" already exists


## Configure Istio

For this example we will create the default istio gateway for seldon which needs to be called `seldon-gateway`. You can supply your own gateway by adding to your SeldonDeployments resources the annotation `seldon.io/istio-gateway` with values the name of your istio gateway.

Create a gateway for our istio-ingress


```python
%%writefile resources/seldon-gateway.yaml
apiVersion: networking.istio.io/v1alpha3
kind: Gateway
metadata:
  name: seldon-gateway
  namespace: istio-system
spec:
  selector:
    istio: ingressgateway # use istio default controller
  servers:
  - port:
      number: 80
      name: http
      protocol: HTTP
    hosts:
    - "*"
```

    Overwriting resources/seldon-gateway.yaml



```python
!kubectl create -f resources/seldon-gateway.yaml -n istio-system 
```

    Error from server (AlreadyExists): error when creating "resources/seldon-gateway.yaml": gateways.networking.istio.io "seldon-gateway" already exists


Ensure the istio ingress gatewaty is port-forwarded to localhost:8004

 * Istio: `kubectl port-forward $(kubectl get pods -l istio=ingressgateway -n istio-system -o jsonpath='{.items[0].metadata.name}') -n istio-system 8004:8080`


```python
ISTIO_GATEWAY = "localhost:8004"

VERSION = !cat ../version.txt
VERSION = VERSION[0]
VERSION
```




    '1.19.0-dev'




```python
from IPython.core.magic import register_line_cell_magic


@register_line_cell_magic
def writetemplate(line, cell):
    with open(line, "w") as f:
        f.write(cell.format(**globals()))
```

## Serve Single Model


```python
!helm upgrade -i mymodel ../helm-charts/seldon-single-model --set model.image=seldonio/mock_classifier:$VERSION --namespace seldon
```

    Release "mymodel" does not exist. Installing it now.
    NAME: mymodel
    LAST DEPLOYED: Thu Dec  4 09:49:01 2025
    NAMESPACE: seldon
    STATUS: deployed
    REVISION: 1
    TEST SUITE: None



```python
!helm template mymodel ../helm-charts/seldon-single-model --set model.image=seldonio/mock_classifier:$VERSION | pygmentize -l json
```

    [34m---[39;49;00m[37m[39;49;00m
    [04m[91m#[39;49;00m[37m [39;49;00m[04m[91mS[39;49;00m[04m[91mo[39;49;00m[04m[91mu[39;49;00m[04m[91mr[39;49;00m[04m[91mc[39;49;00m[04m[91me[39;49;00m:[37m [39;49;00m[04m[91ms[39;49;00m[04m[91me[39;49;00m[04m[91ml[39;49;00m[04m[91md[39;49;00m[04m[91mo[39;49;00m[34mn[39;49;00m[34m-[39;49;00m[04m[91ms[39;49;00m[04m[91mi[39;49;00m[34mn[39;49;00m[04m[91mg[39;49;00m[04m[91ml[39;49;00m[04m[91me[39;49;00m[34m-[39;49;00m[04m[91mm[39;49;00m[04m[91mo[39;49;00m[04m[91md[39;49;00m[04m[91me[39;49;00m[04m[91ml[39;49;00m[04m[91m/[39;49;00m[34mte[39;49;00m[04m[91mm[39;49;00m[04m[91mp[39;49;00m[04m[91ml[39;49;00m[04m[91ma[39;49;00m[34mtes[39;49;00m[04m[91m/[39;49;00m[04m[91ms[39;49;00m[04m[91me[39;49;00m[04m[91ml[39;49;00m[04m[91md[39;49;00m[04m[91mo[39;49;00m[34mn[39;49;00m[04m[91md[39;49;00m[04m[91me[39;49;00m[04m[91mp[39;49;00m[04m[91ml[39;49;00m[04m[91mo[39;49;00m[04m[91my[39;49;00m[04m[91mm[39;49;00m[04m[91me[39;49;00m[34mnt[39;49;00m[04m[91m.[39;49;00m[04m[91mj[39;49;00m[04m[91ms[39;49;00m[04m[91mo[39;49;00m[34mn[39;49;00m[37m[39;49;00m
    {[37m[39;49;00m
    [37m  [39;49;00m[94m"kind"[39;49;00m:[37m [39;49;00m[33m"SeldonDeployment"[39;49;00m,[37m[39;49;00m
    [37m  [39;49;00m[94m"apiVersion"[39;49;00m:[37m [39;49;00m[33m"machinelearning.seldon.io/v1"[39;49;00m,[37m[39;49;00m
    [37m  [39;49;00m[94m"metadata"[39;49;00m:[37m [39;49;00m{[37m[39;49;00m
    [37m    [39;49;00m[94m"name"[39;49;00m:[37m [39;49;00m[33m"mymodel"[39;49;00m,[37m[39;49;00m
    [37m    [39;49;00m[94m"namespace"[39;49;00m:[37m [39;49;00m[33m"seldon"[39;49;00m,[37m[39;49;00m
    [37m    [39;49;00m[94m"labels"[39;49;00m:[37m [39;49;00m{}[37m[39;49;00m
    [37m  [39;49;00m},[37m[39;49;00m
    [37m  [39;49;00m[94m"spec"[39;49;00m:[37m [39;49;00m{[37m[39;49;00m
    [37m      [39;49;00m[94m"name"[39;49;00m:[37m [39;49;00m[33m"mymodel"[39;49;00m,[37m[39;49;00m
    [37m      [39;49;00m[94m"protocol"[39;49;00m:[37m [39;49;00m[33m"seldon"[39;49;00m,[37m[39;49;00m
    [37m    [39;49;00m[94m"annotations"[39;49;00m:[37m [39;49;00m{},[37m[39;49;00m
    [37m    [39;49;00m[94m"predictors"[39;49;00m:[37m [39;49;00m[[37m[39;49;00m
    [37m      [39;49;00m{[37m[39;49;00m
    [37m        [39;49;00m[94m"name"[39;49;00m:[37m [39;49;00m[33m"default"[39;49;00m,[37m[39;49;00m
    [37m        [39;49;00m[94m"graph"[39;49;00m:[37m [39;49;00m{[37m[39;49;00m
    [37m          [39;49;00m[94m"name"[39;49;00m:[37m [39;49;00m[33m"model"[39;49;00m,[37m[39;49;00m
    [37m          [39;49;00m[94m"type"[39;49;00m:[37m [39;49;00m[33m"MODEL"[39;49;00m,[37m[39;49;00m
    [37m        [39;49;00m},[37m[39;49;00m
    [37m        [39;49;00m[94m"componentSpecs"[39;49;00m:[37m [39;49;00m[[37m[39;49;00m
    [37m          [39;49;00m{[37m[39;49;00m
    [37m            [39;49;00m[94m"spec"[39;49;00m:[37m [39;49;00m{[37m[39;49;00m
    [37m              [39;49;00m[94m"containers"[39;49;00m:[37m [39;49;00m[[37m[39;49;00m
    [37m                [39;49;00m{[37m[39;49;00m
    [37m                  [39;49;00m[94m"name"[39;49;00m:[37m [39;49;00m[33m"model"[39;49;00m,[37m[39;49;00m
    [37m                  [39;49;00m[94m"image"[39;49;00m:[37m [39;49;00m[33m"seldonio/mock_classifier:1.19.0-dev"[39;49;00m,[37m[39;49;00m
    [37m                  [39;49;00m[94m"env"[39;49;00m:[37m [39;49;00m[[37m[39;49;00m
    [37m                      [39;49;00m{[37m[39;49;00m
    [37m                        [39;49;00m[94m"name"[39;49;00m:[37m [39;49;00m[33m"LOG_LEVEL"[39;49;00m,[37m[39;49;00m
    [37m                        [39;49;00m[94m"value"[39;49;00m:[37m [39;49;00m[33m"INFO"[39;49;00m[37m[39;49;00m
    [37m                      [39;49;00m},[37m[39;49;00m
    [37m                    [39;49;00m],[37m[39;49;00m
    [37m                  [39;49;00m[94m"resources"[39;49;00m:[37m [39;49;00m{[94m"requests"[39;49;00m:{[94m"memory"[39;49;00m:[33m"1Mi"[39;49;00m}},[37m[39;49;00m
    [37m                [39;49;00m}[37m[39;49;00m
    [37m              [39;49;00m][37m[39;49;00m
    [37m            [39;49;00m},[37m[39;49;00m
    [37m          [39;49;00m}[37m[39;49;00m
    [37m        [39;49;00m],[37m[39;49;00m
    [37m        [39;49;00m[94m"replicas"[39;49;00m:[37m [39;49;00m[34m1[39;49;00m[37m[39;49;00m
    [37m      [39;49;00m}[37m[39;49;00m
    [37m    [39;49;00m][37m[39;49;00m
    [37m  [39;49;00m}[37m[39;49;00m
    }[37m[39;49;00m



```python
!kubectl wait sdep/mymodel \
  --for=condition=ready \
  --timeout=120s \
  -n seldon
```

    seldondeployment.machinelearning.seldon.io/mymodel condition met


### Get predictions


```python
from seldon_core.seldon_client import SeldonClient

sc = SeldonClient(
    deployment_name="mymodel", namespace="seldon", gateway_endpoint=ISTIO_GATEWAY
)
```

    2025-12-04 09:50:10.240979: E external/local_xla/xla/stream_executor/cuda/cuda_fft.cc:477] Unable to register cuFFT factory: Attempting to register factory for plugin cuFFT when one has already been registered
    WARNING: All log messages before absl::InitializeLog() is called are written to STDERR
    E0000 00:00:1764841810.258347 3602796 cuda_dnn.cc:8310] Unable to register cuDNN factory: Attempting to register factory for plugin cuDNN when one has already been registered
    E0000 00:00:1764841810.263401 3602796 cuda_blas.cc:1418] Unable to register cuBLAS factory: Attempting to register factory for plugin cuBLAS when one has already been registered
    2025-12-04 09:50:10.281927: I tensorflow/core/platform/cpu_feature_guard.cc:210] This TensorFlow binary is optimized to use available CPU instructions in performance-critical operations.
    To enable the following instructions: AVX2 FMA, in other operations, rebuild TensorFlow with the appropriate compiler flags.


#### REST Request


```python
from tenacity import retry, stop_after_delay, wait_exponential

@retry(stop=stop_after_delay(300), wait=wait_exponential(multiplier=1, min=0.5, max=5))
def predict():
    r = sc.predict(gateway="istio", transport="rest")
    assert r.success == True
    return r

predict()
```




    Success:True message:
    Request:
    meta {
    }
    data {
      tensor {
        shape: 1
        shape: 1
        values: 0.30196264156462915
      }
    }
    
    Response:
    {'data': {'names': ['proba'], 'tensor': {'shape': [1, 1], 'values': [0.06819806874238313]}}, 'meta': {'requestPath': {'model': 'seldonio/mock_classifier:1.19.0-dev'}}}



## gRPC Request


```python
from tenacity import retry, stop_after_delay, wait_exponential

@retry(stop=stop_after_delay(300), wait=wait_exponential(multiplier=1, min=0.5, max=5))
def predict():
    r = sc.predict(gateway="istio", transport="grpc")
    assert r.success == True
    return r

predict()
```




    Success:True message:
    Request:
    {'meta': {}, 'data': {'tensor': {'shape': [1, 1], 'values': [0.2596814235407022]}}}
    Response:
    {'meta': {'requestPath': {'model': 'seldonio/mock_classifier:1.19.0-dev'}}, 'data': {'names': ['proba'], 'tensor': {'shape': [1, 1], 'values': [0.0655597808283028]}}}




```python
!helm delete mymodel -n seldon
```

    release "mymodel" uninstalled


## Host Restriction

In this example we will restriction request to those with the Host header "seldon.io"


```python
%%writetemplate resources/model_seldon.yaml
apiVersion: machinelearning.seldon.io/v1
kind: SeldonDeployment
metadata:
  name: mock-classifier-restricted
  annotations:
    "seldon.io/istio-host": "seldon.io"
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
!kubectl apply -f resources/model_seldon.yaml --namespace seldon
```

    seldondeployment.machinelearning.seldon.io/mock-classifier-restricted created



```python
!kubectl wait sdep/mock-classifier-restricted \
  --for=condition=ready \
  --timeout=120s \
  -n seldon
```

    seldondeployment.machinelearning.seldon.io/mock-classifier-restricted condition met



```python
sc = SeldonClient(
    deployment_name="mock-classifier-restricted", namespace="seldon", gateway_endpoint="localhost:8003"
)
```


```python
@retry(stop=stop_after_delay(300), wait=wait_exponential(multiplier=1, min=0.5, max=5))
def predict():
   X=!curl -s -d '{"data": {"ndarray":[[1.0, 2.0, 5.0]]}}' \
      -X POST http://localhost:8003/seldon/seldon/mock-classifier-restricted/api/v1.0/predictions \
      -H "Content-Type: application/json" \
   assert X == []

predict()
```


```python
import json

@retry(stop=stop_after_delay(300), wait=wait_exponential(multiplier=1, min=0.5, max=5))
def predict():
   X=!curl -s -d '{"data": {"ndarray":[[1.0, 2.0, 5.0]]}}' \
      -X POST http://localhost:8003/seldon/seldon/mock-classifier-restricted/api/v1.0/predictions \
      -H "Content-Type: application/json" \
      -H "Host: seldon.io"
   d=json.loads(X[0])
   assert(d["data"]["ndarray"][0][0] > 0.4)

   return d

predict()
```




    {'data': {'names': ['proba'], 'ndarray': [[0.43782349911420193]]},
     'meta': {'requestPath': {'classifier': 'seldonio/mock_classifier:1.19.0-dev'}}}




```python
!kubectl delete -f resources/model_seldon.yaml -n seldon
```

    seldondeployment.machinelearning.seldon.io "mock-classifier-restricted" deleted

