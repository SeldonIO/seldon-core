# Shadow Rollout with Seldon and Ambassador

This notebook shows how you can deploy "shadow" deployments to direct traffic not only to the main Seldon Deployment but also to a shadow deployment whose response will be dicarded. This allows you to test new models in a production setting and with production traffic and anlalyse how they perform before putting them live.

These are useful when you want to test a new model or higher latency inference piepline (e.g., with explanation components) with production traffic but without affecting the live deployment.


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
  name: production-model
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
    name: default
    replicas: 1

```


```python
!kubectl apply -f model.yaml
```

    seldondeployment.machinelearning.seldon.io/example created



```python
!kubectl rollout status deploy/$(kubectl get deploy -l seldon-deployment-id=example -o jsonpath='{.items[0].metadata.name}')
```

    Waiting for deployment "example-default-0-classifier" rollout to finish: 0 of 1 updated replicas are available...
    deployment "example-default-0-classifier" successfully rolled out


### Get predictions


```python
from seldon_core.seldon_client import SeldonClient

sc = SeldonClient(deployment_name="example", namespace="seldon")
```

#### REST Request


```python
r = sc.predict(gateway="ambassador", transport="rest")
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
        values: 0.6853937460651732
      }
    }
    
    Response:
    {'data': {'names': ['proba'], 'tensor': {'shape': [1, 1], 'values': [0.0969771108807172]}}, 'meta': {'requestPath': {'classifier': 'seldonio/mock_classifier:1.15.0-dev'}}}


## Launch Shadow

We will now create a new Seldon Deployment for our Shadow deployment with a new model.


```python
%%writetemplate shadow.yaml
apiVersion: machinelearning.seldon.io/v1alpha2
kind: SeldonDeployment
metadata:
  labels:
    app: seldon
  name: example
spec:
  name: shadow-model
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
    name: default
    replicas: 1
  - componentSpecs:
    - spec:
        containers:
        - image: seldonio/mock_classifier:{VERSION}
          imagePullPolicy: IfNotPresent
          name: classifier
    graph:
      children: []
      endpoint:
        type: REST
      name: classifier
      type: MODEL
    name: shadow
    replicas: 1
    shadow: true
    traffic: 100

```


```python
!kubectl apply -f shadow.yaml
```

    seldondeployment.machinelearning.seldon.io/example configured



```python
!kubectl rollout status deploy/$(kubectl get deploy -l seldon-deployment-id=example -o jsonpath='{.items[0].metadata.name}')
!kubectl rollout status deploy/$(kubectl get deploy -l seldon-deployment-id=example -o jsonpath='{.items[1].metadata.name}')
```

    deployment "example-default-0-classifier" successfully rolled out
    Waiting for deployment "example-shadow-0-classifier" rollout to finish: 0 of 1 updated replicas are available...
    deployment "example-shadow-0-classifier" successfully rolled out


Let's send a bunch of requests to the endpoint.


```python
for i in range(10):
    r = sc.predict(gateway="ambassador", transport="rest")
```


```python
default_count = !kubectl logs $(kubectl get pod -lseldon-app=example-default -o jsonpath='{.items[0].metadata.name}') classifier | grep "root.predict" | wc -l
```


```python
shadow_count = !kubectl logs $(kubectl get pod -lseldon-app=example-shadow -o jsonpath='{.items[0].metadata.name}') classifier | grep "root.predict" | wc -l
```


```python
print(shadow_count)
print(default_count)
assert int(shadow_count[0]) == 10
assert int(default_count[0]) == 11
```

    ['10']
    ['11']


## TearDown


```python
!kubectl delete -f model.yaml
```

    seldondeployment.machinelearning.seldon.io "example" deleted



```python

```
