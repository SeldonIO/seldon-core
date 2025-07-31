# Canary Rollout with Seldon and Ambassador


## Setup Seldon Core

Use the setup notebook to [Setup Cluster](../notebooks/seldon-core-setup.md#setup-cluster) with [Ambassador Ingress](../notebooks/seldon-core-setup.md#ambassador) and [Install Seldon Core](../notebooks/seldon-core-setup.md#Install-Seldon-Core). Instructions [also online](../notebooks/seldon-core-setup.md).


```python
!kubectl create namespace seldon
```

    Error from server (AlreadyExists): namespaces "seldon" already exists



```python
!kubectl config set-context $(kubectl config current-context) --namespace=seldon
```

    Context "kind-ansible" modified.



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




    '1.15.0-dev'



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

sc = SeldonClient(deployment_name="example", namespace="seldon")
```

    2022-08-22 17:10:39.787195: W tensorflow/stream_executor/platform/default/dso_loader.cc:64] Could not load dynamic library 'libcudart.so.11.0'; dlerror: libcudart.so.11.0: cannot open shared object file: No such file or directory
    2022-08-22 17:10:39.787229: I tensorflow/stream_executor/cuda/cudart_stub.cc:29] Ignore above cudart dlerror if you do not have a GPU set up on your machine.


#### REST Request


```python
r = sc.predict(gateway="ambassador", transport="rest")
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
        values: 0.7321696737551778
      }
    }
    
    Response:
    {'data': {'names': ['proba'], 'tensor': {'shape': [1, 1], 'values': [0.10115132850847434]}}, 'meta': {'requestPath': {'classifier': 'seldonio/mock_classifier:1.15.0-dev'}}}


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

    Warning: resource seldondeployments/example is missing the kubectl.kubernetes.io/last-applied-configuration annotation which is required by kubectl apply. kubectl apply should only be used on resources created declaratively by either kubectl create --save-config or kubectl apply. The missing annotation will be patched automatically.
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
sc.predict(gateway="ambassador", transport="rest")
```




    Success:True message:
    Request:
    meta {
    }
    data {
      tensor {
        shape: 1
        shape: 1
        values: 0.5515112951204143
      }
    }
    
    Response:
    {'data': {'names': ['proba'], 'tensor': {'shape': [1, 1], 'values': [0.08586865749102578]}}, 'meta': {'requestPath': {'classifier': 'seldonio/mock_classifier:1.15.0-dev'}}}




```python
from collections import defaultdict

counts = defaultdict(int)
n = 100
for i in range(n):
    r = sc.predict(gateway="ambassador", transport="rest")
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

    0.3246753246753247



```python
!kubectl delete -f canary.yaml
```

    seldondeployment.machinelearning.seldon.io "example" deleted



```python

```
