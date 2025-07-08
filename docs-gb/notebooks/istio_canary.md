# Canary Rollout with Seldon and Istio


## Setup Seldon Core

Use the setup notebook to [Setup Cluster](https://docs.seldon.io/projects/seldon-core/en/latest/examples/seldon_core_setup.html#Setup-Cluster) with [Istio Ingress](https://docs.seldon.io/projects/seldon-core/en/latest/examples/seldon_core_setup.html#Istio) and [Install Seldon Core](https://docs.seldon.io/projects/seldon-core/en/latest/examples/seldon_core_setup.html#Install-Seldon-Core). Instructions [also online](https://docs.seldon.io/projects/seldon-core/en/latest/examples/seldon_core_setup.html).


```python
!kubectl create namespace seldon
```

    namespace/seldon created



```python
!kubectl config set-context $(kubectl config current-context) --namespace=seldon
```

    Context "kind-kind" modified.



```python
from IPython.core.magic import register_line_cell_magic


@register_line_cell_magic
def writetemplate(line, cell):
    with open(line, "w") as f:
        f.write(cell.format(**globals()))
```

Ensure the istio ingress gatewaty is port-forwarded to localhost:8004



* Istio: `kubectl port-forward $(kubectl get pods -l istio=ingressgateway -n istio-system -o jsonpath='{.items[0].metadata.name}') -n istio-system 8004:8080`



```python
ISTIO_GATEWAY = "localhost:8004"

VERSION = !cat ../../../version.txt
VERSION = VERSION[0]
VERSION
```




    '1.7.0-dev'



## Launch main model

We will create a very simple Seldon Deployment with a dummy model image `seldonio/mock_classifier:1.0`. This deployment is named `example`.


```python
%%writetemplate model.yaml
apiVersion: machinelearning.seldon.io/v1alpha2
kind: SeldonDeployment
metadata:
  labels:
    app: seldon
  name: example
spec:
  name: canary-example
  predictors:
  - componentSpecs:
    - spec:
        containers:
        - image: seldonio/mock_classifier:{VERSION}
          imagePullPolicy: IfNotPresent
          name: classifier
        terminationGracePeriodSeconds: 1
    graph:
      children: []
      endpoint:
        type: REST
      name: classifier
      type: MODEL
    name: main
    replicas: 1

```


```python
!kubectl create -f model.yaml
```

    seldondeployment.machinelearning.seldon.io/example created



```python
!kubectl rollout status deploy/$(kubectl get deploy -l seldon-deployment-id=example -o jsonpath='{.items[0].metadata.name}')
```

    Waiting for deployment "example-main-0-classifier" rollout to finish: 0 of 1 updated replicas are available...
    deployment "example-main-0-classifier" successfully rolled out


### Get predictions


```python
from seldon_core.seldon_client import SeldonClient

sc = SeldonClient(
    deployment_name="example", namespace="seldon", gateway_endpoint=ISTIO_GATEWAY
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
        values: 0.6670563912281003
      }
    }
    
    Response:
    {'data': {'names': ['proba'], 'tensor': {'shape': [1, 1], 'values': [0.09538308704053941]}}, 'meta': {'requestPath': {'classifier': 'seldonio/mock_classifier:1.7.0-dev'}}}


## Launch Canary

We will now extend the existing graph and add a new predictor as a canary using a new model `seldonio/mock_classifier_rest:1.1`. We will add traffic values to split traffic 75/25 to the main and canary.


```python
%%writetemplate canary.yaml
apiVersion: machinelearning.seldon.io/v1alpha2
kind: SeldonDeployment
metadata:
  labels:
    app: seldon
  name: example
spec:
  name: canary-example
  predictors:
  - componentSpecs:
    - spec:
        containers:
        - image: seldonio/mock_classifier:{VERSION}
          imagePullPolicy: IfNotPresent
          name: classifier
        terminationGracePeriodSeconds: 1
    graph:
      children: []
      endpoint:
        type: REST
      name: classifier
      type: MODEL
    name: main
    replicas: 1
    traffic: 75
  - componentSpecs:
    - spec:
        containers:
        - image: seldonio/mock_classifier:{VERSION}
          imagePullPolicy: IfNotPresent
          name: classifier
        terminationGracePeriodSeconds: 1
    graph:
      children: []
      endpoint:
        type: REST
      name: classifier
      type: MODEL
    name: canary
    replicas: 1
    traffic: 25

```


```python
!kubectl apply -f canary.yaml
```

    Warning: kubectl apply should be used on resource created by either kubectl create --save-config or kubectl apply
    seldondeployment.machinelearning.seldon.io/example configured



```python
!kubectl rollout status deploy/$(kubectl get deploy -l seldon-deployment-id=example -o jsonpath='{.items[0].metadata.name}')
!kubectl rollout status deploy/$(kubectl get deploy -l seldon-deployment-id=example -o jsonpath='{.items[1].metadata.name}')
```

    Waiting for deployment "example-canary-0-classifier" rollout to finish: 0 of 1 updated replicas are available...
    deployment "example-canary-0-classifier" successfully rolled out
    deployment "example-main-0-classifier" successfully rolled out


Show our REST requests are now split with roughly 25% going to the canary.


```python
sc.predict(gateway="istio", transport="rest")
```




    Success:True message:
    Request:
    meta {
    }
    data {
      tensor {
        shape: 1
        shape: 1
        values: 0.5754642896429739
      }
    }
    
    Response:
    {'data': {'names': ['proba'], 'tensor': {'shape': [1, 1], 'values': [0.08776759944872958]}}, 'meta': {'requestPath': {'classifier': 'seldonio/mock_classifier:1.7.0-dev'}}}




```python
from collections import defaultdict

counts = defaultdict(int)
n = 100
for i in range(n):
    r = sc.predict(gateway="istio", transport="rest")
```

Following checks number of prediction requests processed by default/canary predictors respectively.


```python
default_count = !kubectl logs $(kubectl get pod -lseldon-app=example-main -o jsonpath='{.items[0].metadata.name}') classifier | grep "root:predict" | wc -l
```


```python
canary_count = !kubectl logs $(kubectl get pod -lseldon-app=example-canary -o jsonpath='{.items[0].metadata.name}') classifier | grep "root:predict" | wc -l
```


```python
canary_percentage = float(canary_count[0]) / float(default_count[0])
print(canary_percentage)
assert canary_percentage > 0.1 and canary_percentage < 0.5
```

    0.275



```python
!kubectl delete -f canary.yaml
```

    seldondeployment.machinelearning.seldon.io "example" deleted

