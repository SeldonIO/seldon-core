# Custom Header Routing with Seldon and Ambassador

This notebook shows how you can deploy Seldon Deployments which can have custom routing via Ambassador's custom header routing.


## Setup Seldon Core

Use the setup notebook to [Setup Cluster](https://docs.seldon.io/projects/seldon-core/en/latest/examples/seldon_core_setup.html#Setup-Cluster) with [Ambassador Ingress](https://docs.seldon.io/projects/seldon-core/en/latest/examples/seldon_core_setup.html#Ambassador) and [Install Seldon Core](https://docs.seldon.io/projects/seldon-core/en/latest/examples/seldon_core_setup.html#Install-Seldon-Core). Instructions [also online](https://docs.seldon.io/projects/seldon-core/en/latest/examples/seldon_core_setup.html).


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
    name: single
    replicas: 1

```


```python
!kubectl create -f model.yaml
```

    seldondeployment.machinelearning.seldon.io/example created



```python
!kubectl rollout status deploy/$(kubectl get deploy -l seldon-deployment-id=example -o jsonpath='{.items[0].metadata.name}')
```

    Waiting for deployment "example-single-0-classifier" rollout to finish: 0 of 1 updated replicas are available...
    deployment "example-single-0-classifier" successfully rolled out


### Get predictions


```python
from seldon_core.seldon_client import SeldonClient

sc = SeldonClient(deployment_name="example", namespace="seldon")
```

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
        values: 0.4691931399866406
      }
    }
    
    Response:
    {'data': {'names': ['proba'], 'tensor': {'shape': [1, 1], 'values': [0.0796235017124669]}}, 'meta': {'requestPath': {'classifier': 'seldonio/mock_classifier:1.15.0-dev'}}}


## Launch Model with Custom Routing

We will now create a new graph for our Canary with a new model `seldonio/mock_classifier_rest:1.1`. To make it a canary of the original `example` deployment we add two annotations

```
"annotations": {
	    "seldon.io/ambassador-header":"location:london"
	    "seldon.io/ambassador-service-name":"example"	    
	},	
```

The first annotation says we want to route traffic that has the header `location:london`. The second says we want to use `example` as our service endpoint rather than the default which would be our deployment name - in this case `example-canary`. This will ensure that this Ambassador setting will apply to the same prefix as the previous one.


```python
%%writetemplate model_with_header.yaml
apiVersion: machinelearning.seldon.io/v1alpha2
kind: SeldonDeployment
metadata:
  labels:
    app: seldon
  name: example-header
spec:
  annotations:
    seldon.io/ambassador-header: 'location:london'
    seldon.io/ambassador-service-name: example
  name: header-model
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
    name: single
    replicas: 1
```


```python
!kubectl create -f model_with_header.yaml
```

    seldondeployment.machinelearning.seldon.io/example-header created



```python
!kubectl rollout status deploy/$(kubectl get deploy -l seldon-deployment-id=example-header -o jsonpath='{.items[0].metadata.name}')
```

    Waiting for deployment "example-header-single-0-classifier" rollout to finish: 0 of 1 updated replicas are available...
    deployment "example-header-single-0-classifier" successfully rolled out


Check a request without a header goes to the existing model.


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
        values: 0.17492496126262558
      }
    }
    
    Response:
    {'data': {'names': ['proba'], 'tensor': {'shape': [1, 1], 'values': [0.06055474553922779]}}, 'meta': {'requestPath': {'classifier': 'seldonio/mock_classifier:1.15.0-dev'}}}



```python
default_count = !kubectl logs $(kubectl get pod -lseldon-app=example-single -o jsonpath='{.items[0].metadata.name}') classifier | grep "root.predict" | wc -l
```


```python
print(default_count)
assert int(default_count[0]) == 2
```

    ['2']


Check a REST request with the required header gets routed to the new model.


```python
r = sc.predict(gateway="ambassador", transport="rest", headers={"location": "london"})
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
        values: 0.8493669886048304
      }
    }
    
    Response:
    {'data': {'names': ['proba'], 'tensor': {'shape': [1, 1], 'values': [0.11231598191770942]}}, 'meta': {'requestPath': {'classifier': 'seldonio/mock_classifier:1.15.0-dev'}}}



```python
header_count = !kubectl logs $(kubectl get pod -lseldon-app=example-header-single -o jsonpath='{.items[0].metadata.name}') classifier | grep "root.predict" | wc -l
```


```python
print(header_count)
assert int(header_count[0]) == 1
```

    ['1']



```python
!kubectl delete -f model.yaml
```

    seldondeployment.machinelearning.seldon.io "example" deleted



```python
!kubectl delete -f model_with_header.yaml
```

    seldondeployment.machinelearning.seldon.io "example-header" deleted



```python

```
