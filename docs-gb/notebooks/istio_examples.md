# Example Seldon Core Deployments using Helm with Istio

Prequisites

 * [Install istio](https://istio.io/latest/docs/setup/getting-started/#download)

## Setup Cluster and Ingress

Use the setup notebook to [Setup Cluster](../notebooks/seldon-core-setup.md#setup-cluster) with [Istio Ingress](../notebooks/seldon-core-setup.md#Istio). Instructions [also online](../notebooks/seldon-core-setup.md).


```python
!kubectl create namespace seldon
```

    namespace/seldon created



```python
!kubectl config set-context $(kubectl config current-context) --namespace=seldon
```

    Context "kind-kind" modified.


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

    gateway.networking.istio.io/seldon-gateway created


Ensure the istio ingress gatewaty is port-forwarded to localhost:8004

 * Istio: `kubectl port-forward $(kubectl get pods -l istio=ingressgateway -n istio-system -o jsonpath='{.items[0].metadata.name}') -n istio-system 8004:8080`


```python
ISTIO_GATEWAY = "localhost:8004"
VERSION = !cat ../version.txt
VERSION = VERSION[0]
VERSION
```




    '1.9.0-dev'




```python
from IPython.core.magic import register_line_cell_magic


@register_line_cell_magic
def writetemplate(line, cell):
    with open(line, "w") as f:
        f.write(cell.format(**globals()))
```

## Start Seldon Core

Use the setup notebook to [Install Seldon Core](../notebooks/seldon-core-setup.md#Install-Seldon-Core) with Istio Ingress. Instructions [also online](../notebooks/seldon-core-setup.md).

## Serve Single Model


```python
!helm install mymodel ../helm-charts/seldon-single-model --set model.image=seldonio/mock_classifier:$VERSION
```

    NAME: mymodel
    LAST DEPLOYED: Wed Mar 10 16:37:01 2021
    NAMESPACE: seldon
    STATUS: deployed
    REVISION: 1
    TEST SUITE: None



```python
!helm template mymodel ../helm-charts/seldon-single-model --set model.image=seldonio/mock_classifier:$VERSION | pygmentize -l json
```

    [04m[91m-[39;49;00m[04m[91m-[39;49;00m[04m[91m-[39;49;00m
    [04m[91m#[39;49;00m [04m[91mS[39;49;00m[04m[91mo[39;49;00m[04m[91mu[39;49;00m[04m[91mr[39;49;00m[04m[91mc[39;49;00m[04m[91me[39;49;00m[04m[91m:[39;49;00m [04m[91ms[39;49;00m[04m[91me[39;49;00m[04m[91ml[39;49;00m[04m[91md[39;49;00m[04m[91mo[39;49;00m[04m[91mn[39;49;00m[04m[91m-[39;49;00m[04m[91ms[39;49;00m[04m[91mi[39;49;00m[04m[91mn[39;49;00m[04m[91mg[39;49;00m[04m[91ml[39;49;00m[04m[91me[39;49;00m[04m[91m-[39;49;00m[04m[91mm[39;49;00m[04m[91mo[39;49;00m[04m[91md[39;49;00m[04m[91me[39;49;00m[04m[91ml[39;49;00m[04m[91m/[39;49;00m[04m[91mt[39;49;00m[04m[91me[39;49;00m[04m[91mm[39;49;00m[04m[91mp[39;49;00m[04m[91ml[39;49;00m[04m[91ma[39;49;00m[04m[91mt[39;49;00m[04m[91me[39;49;00m[04m[91ms[39;49;00m[04m[91m/[39;49;00m[04m[91ms[39;49;00m[04m[91me[39;49;00m[04m[91ml[39;49;00m[04m[91md[39;49;00m[04m[91mo[39;49;00m[04m[91mn[39;49;00m[04m[91md[39;49;00m[04m[91me[39;49;00m[04m[91mp[39;49;00m[04m[91ml[39;49;00m[04m[91mo[39;49;00m[04m[91my[39;49;00m[04m[91mm[39;49;00m[04m[91me[39;49;00m[04m[91mn[39;49;00m[04m[91mt[39;49;00m[04m[91m.[39;49;00m[04m[91mj[39;49;00m[04m[91ms[39;49;00m[04m[91mo[39;49;00m[04m[91mn[39;49;00m
    {
      [94m"kind"[39;49;00m: [33m"SeldonDeployment"[39;49;00m,
      [94m"apiVersion"[39;49;00m: [33m"machinelearning.seldon.io/v1"[39;49;00m,
      [94m"metadata"[39;49;00m: {
        [94m"name"[39;49;00m: [33m"mymodel"[39;49;00m,
        [94m"namespace"[39;49;00m: [33m"seldon"[39;49;00m,
        [94m"labels"[39;49;00m: {}
      },
      [94m"spec"[39;49;00m: {
          [94m"name"[39;49;00m: [33m"mymodel"[39;49;00m,
          [94m"protocol"[39;49;00m: [33m"seldon"[39;49;00m,
        [94m"annotations"[39;49;00m: {},
        [94m"predictors"[39;49;00m: [
          {
            [94m"name"[39;49;00m: [33m"default"[39;49;00m,
            [94m"graph"[39;49;00m: {
              [94m"name"[39;49;00m: [33m"model"[39;49;00m,
              [94m"type"[39;49;00m: [33m"MODEL"[39;49;00m,
            },
            [94m"componentSpecs"[39;49;00m: [
              {
                [94m"spec"[39;49;00m: {
                  [94m"containers"[39;49;00m: [
                    {
                      [94m"name"[39;49;00m: [33m"model"[39;49;00m,
                      [94m"image"[39;49;00m: [33m"seldonio/mock_classifier:1.7.0-dev"[39;49;00m,
                      [94m"env"[39;49;00m: [
                          {
                            [94m"name"[39;49;00m: [33m"LOG_LEVEL"[39;49;00m,
                            [94m"value"[39;49;00m: [33m"INFO"[39;49;00m
                          },
                        ],
                      [94m"resources"[39;49;00m: {[94m"requests"[39;49;00m:{[94m"memory"[39;49;00m:[33m"1Mi"[39;49;00m}},
                    }
                  ]
                },
              }
            ],
            [94m"replicas"[39;49;00m: [34m1[39;49;00m
          }
        ]
      }
    }



```python
!kubectl rollout status deploy/mymodel-default-0-model
```

    Waiting for deployment "mymodel-default-0-model" rollout to finish: 0 of 1 updated replicas are available...
    deployment "mymodel-default-0-model" successfully rolled out


### Get predictions


```python
from seldon_core.seldon_client import SeldonClient

sc = SeldonClient(
    deployment_name="mymodel", namespace="seldon", gateway_endpoint=ISTIO_GATEWAY
)
```

#### REST Request


```python
r = sc.predict(gateway="istio", transport="rest")
assert r.success == True
print(r)
```

    Success:True message:
    Request:
    meta {
    }
    data {
      tensor {
        shape: 1
        shape: 1
        values: 0.721679221744617
      }
    }
    
    Response:
    {'data': {'names': ['proba'], 'tensor': {'shape': [1, 1], 'values': [0.1002015221659356]}}, 'meta': {'requestPath': {'model': 'seldonio/mock_classifier:1.7.0-dev'}}}


## gRPC Request


```python
r = sc.predict(gateway="istio", transport="grpc")
assert r.success == True
print(r)
```

    Success:True message:
    Request:
    {'meta': {}, 'data': {'tensor': {'shape': [1, 1], 'values': [0.17825624441824628]}}}
    Response:
    {'meta': {'requestPath': {'model': 'seldonio/mock_classifier:1.7.0-dev'}}, 'data': {'names': ['proba'], 'tensor': {'shape': [1, 1], 'values': [0.06074453279395597]}}}



```python
!helm delete mymodel
```

    release "mymodel" uninstalled


## Host Restriction

In this example we will restriction request to those with the Host header "seldon.io"


```python
%%writetemplate resources/model_seldon.yaml
apiVersion: machinelearning.seldon.io/v1
kind: SeldonDeployment
metadata:
  name: example-seldon
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
!kubectl apply -f resources/model_seldon.yaml
```

    seldondeployment.machinelearning.seldon.io/example-seldon created



```python
!kubectl rollout status deploy/$(kubectl get deploy -l seldon-deployment-id=example-seldon -o jsonpath='{.items[0].metadata.name}')
```

    Waiting for deployment "example-seldon-model-0-classifier" rollout to finish: 0 of 1 updated replicas are available...
    deployment "example-seldon-model-0-classifier" successfully rolled out



```python
for i in range(60):
    state = !kubectl get sdep example-seldon -o jsonpath='{.status.state}'
    state = state[0]
    print(state)
    if state == "Available":
        break
    time.sleep(1)
assert state == "Available"
```

    Available



```python
X=!curl -s -d '{"data": {"ndarray":[[1.0, 2.0, 5.0]]}}' \
   -X POST http://localhost:8003/seldon/seldon/example-seldon/api/v1.0/predictions \
   -H "Content-Type: application/json" \
assert X == []
```


```python
import json
X=!curl -s -d '{"data": {"ndarray":[[1.0, 2.0, 5.0]]}}' \
   -X POST http://localhost:8003/seldon/seldon/example-seldon/api/v1.0/predictions \
   -H "Content-Type: application/json" \
   -H "Host: seldon.io"
d=json.loads(X[0])
print(d)
assert(d["data"]["ndarray"][0][0] > 0.4)
```

    {'data': {'names': ['proba'], 'ndarray': [[0.43782349911420193]]}, 'meta': {'requestPath': {'classifier': 'seldonio/mock_classifier:1.9.0-dev'}}}



```python
!kubectl delete -f resources/model_seldon.yaml
```

    seldondeployment.machinelearning.seldon.io "example-seldon" deleted



```python

```
