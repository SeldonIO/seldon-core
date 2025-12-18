# Example Helm Deployments

![predictor with canary](../.gitbook/assets/deploy-graph.png)

## Setup Seldon Core

Use the setup notebook to [Setup Cluster](seldon-core-setup.md#setup-cluster) with [Ambassador Ingress](seldon-core-setup.md#ambassador) and [Install Seldon Core](seldon-core-setup.md#Install-Seldon-Core). 

```python
!kubectl create namespace seldon
```

```
Error from server (AlreadyExists): namespaces "seldon" already exists
```

```python
VERSION = !cat ../version.txt
VERSION = VERSION[0]
VERSION
```




    '1.19.0-dev'



## Serve Single Model

```python

!helm upgrade -i mymodel ../helm-charts/seldon-single-model --set model.image=seldonio/mock_classifier:$VERSION --namespace seldon
```

    Release "mymodel" has been upgraded. Happy Helming!
    NAME: mymodel
    LAST DEPLOYED: Thu Dec  4 12:11:31 2025
    NAMESPACE: seldon
    STATUS: deployed
    REVISION: 3
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
    deployment_name="mymodel",
    namespace="seldon",
    gateway_endpoint="localhost:8003",
    gateway="ambassador",
)
```

#### REST Request

```python
from tenacity import retry, stop_after_delay, wait_exponential

@retry(stop=stop_after_delay(300), wait=wait_exponential(multiplier=1, min=0.5, max=5))
def predict():
    r = sc.predict(transport="rest")
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
        values: 0.71543328677795837
      }
    }
    
    Response:
    {'data': {'names': ['proba'], 'tensor': {'shape': [1, 1], 'values': [0.09963978586361734]}}, 'meta': {'requestPath': {'model': 'seldonio/mock_classifier:1.19.0-dev'}}}


Response:
{'data': {'names': ['proba'], 'tensor': {'shape': [1, 1], 'values': [0.05335370865277927]}}, 'meta': {}}
```

#### GRPC Request

```python
@retry(stop=stop_after_delay(300), wait=wait_exponential(multiplier=1, min=0.5, max=5))
def predict():
    r = sc.predict(transport="grpc")
    assert r.success == True
    return r

predict()
```




    Success:True message:
    Request:
    {'meta': {}, 'data': {'tensor': {'shape': [1, 1], 'values': [0.555819139294561]}}}
    Response:
    {'meta': {'requestPath': {'model': 'seldonio/mock_classifier:1.19.0-dev'}}, 'data': {'names': ['proba'], 'tensor': {'shape': [1, 1], 'values': [0.08620740652415673]}}}




```python
!helm delete mymodel --namespace seldon
```

```
release "mymodel" uninstalled
```

## Serve REST AB Test

```python
!helm upgrade -i myabtest ../helm-charts/seldon-abtest --namespace seldon
```

    Release "myabtest" does not exist. Installing it now.
    NAME: myabtest
    LAST DEPLOYED: Thu Dec  4 12:12:21 2025
    NAMESPACE: seldon
    STATUS: deployed
    REVISION: 1
    TEST SUITE: None



```python
!helm template ../helm-charts/seldon-abtest | pygmentize -l json
```

    [34m---[39;49;00m[37m[39;49;00m
    [04m[91m#[39;49;00m[37m [39;49;00m[04m[91mS[39;49;00m[04m[91mo[39;49;00m[04m[91mu[39;49;00m[04m[91mr[39;49;00m[04m[91mc[39;49;00m[04m[91me[39;49;00m:[37m [39;49;00m[04m[91ms[39;49;00m[04m[91me[39;49;00m[04m[91ml[39;49;00m[04m[91md[39;49;00m[04m[91mo[39;49;00m[34mn[39;49;00m[34m-[39;49;00m[04m[91ma[39;49;00m[04m[91mb[39;49;00m[34mtest[39;49;00m[04m[91m/[39;49;00m[34mte[39;49;00m[04m[91mm[39;49;00m[04m[91mp[39;49;00m[04m[91ml[39;49;00m[04m[91ma[39;49;00m[34mtes[39;49;00m[04m[91m/[39;49;00m[04m[91ma[39;49;00m[04m[91mb[39;49;00m[04m[91m_[39;49;00m[34mtest[39;49;00m[04m[91m_[39;49;00m[34m2[39;49;00m[04m[91mp[39;49;00m[04m[91mo[39;49;00m[04m[91md[39;49;00m[04m[91ms[39;49;00m[04m[91m.[39;49;00m[04m[91mj[39;49;00m[04m[91ms[39;49;00m[04m[91mo[39;49;00m[34mn[39;49;00m[37m[39;49;00m
    {[37m[39;49;00m
    [37m    [39;49;00m[94m"apiVersion"[39;49;00m:[37m [39;49;00m[33m"machinelearning.seldon.io/v1alpha2"[39;49;00m,[37m[39;49;00m
    [37m    [39;49;00m[94m"kind"[39;49;00m:[37m [39;49;00m[33m"SeldonDeployment"[39;49;00m,[37m[39;49;00m
    [37m    [39;49;00m[94m"metadata"[39;49;00m:[37m [39;49;00m{[37m[39;49;00m
    [37m	[39;49;00m[94m"labels"[39;49;00m:[37m [39;49;00m{[37m[39;49;00m
    [37m	    [39;49;00m[94m"app"[39;49;00m:[37m [39;49;00m[33m"seldon"[39;49;00m[37m[39;49;00m
    [37m	[39;49;00m},[37m[39;49;00m
    [37m	[39;49;00m[94m"name"[39;49;00m:[37m [39;49;00m[33m"release-name"[39;49;00m[37m[39;49;00m
    [37m    [39;49;00m},[37m[39;49;00m
    [37m    [39;49;00m[94m"spec"[39;49;00m:[37m [39;49;00m{[37m[39;49;00m
    [37m	[39;49;00m[94m"name"[39;49;00m:[37m [39;49;00m[33m"release-name"[39;49;00m,[37m[39;49;00m
    [37m	[39;49;00m[94m"predictors"[39;49;00m:[37m [39;49;00m[[37m[39;49;00m
    [37m	    [39;49;00m{[37m[39;49;00m
    [37m		[39;49;00m[94m"name"[39;49;00m:[37m [39;49;00m[33m"default"[39;49;00m,[37m[39;49;00m
    [37m		[39;49;00m[94m"replicas"[39;49;00m:[37m [39;49;00m[34m1[39;49;00m,[37m[39;49;00m
    [37m		[39;49;00m[94m"componentSpecs"[39;49;00m:[37m [39;49;00m[{[37m[39;49;00m
    [37m		    [39;49;00m[94m"spec"[39;49;00m:[37m [39;49;00m{[37m[39;49;00m
    [37m			[39;49;00m[94m"containers"[39;49;00m:[37m [39;49;00m[[37m[39;49;00m
    [37m			    [39;49;00m{[37m[39;49;00m
    [37m                                [39;49;00m[94m"image"[39;49;00m:[37m [39;49;00m[33m"seldonio/mock_classifier:1.19.0-dev"[39;49;00m,[37m[39;49;00m
    [37m				[39;49;00m[94m"imagePullPolicy"[39;49;00m:[37m [39;49;00m[33m"IfNotPresent"[39;49;00m,[37m[39;49;00m
    [37m				[39;49;00m[94m"name"[39;49;00m:[37m [39;49;00m[33m"classifier-1"[39;49;00m,[37m[39;49;00m
    [37m				[39;49;00m[94m"resources"[39;49;00m:[37m [39;49;00m{[37m[39;49;00m
    [37m				    [39;49;00m[94m"requests"[39;49;00m:[37m [39;49;00m{[37m[39;49;00m
    [37m					[39;49;00m[94m"memory"[39;49;00m:[37m [39;49;00m[33m"1Mi"[39;49;00m[37m[39;49;00m
    [37m				    [39;49;00m}[37m[39;49;00m
    [37m				[39;49;00m}[37m[39;49;00m
    [37m			    [39;49;00m}],[37m[39;49;00m
    [37m			[39;49;00m[94m"terminationGracePeriodSeconds"[39;49;00m:[37m [39;49;00m[34m20[39;49;00m[37m[39;49;00m
    [37m		    [39;49;00m}},[37m[39;49;00m
    [37m	        [39;49;00m{[37m[39;49;00m
    [37m		    [39;49;00m[94m"metadata"[39;49;00m:{[37m[39;49;00m
    [37m			[39;49;00m[94m"labels"[39;49;00m:{[37m[39;49;00m
    [37m			    [39;49;00m[94m"version"[39;49;00m:[33m"v2"[39;49;00m[37m[39;49;00m
    [37m			[39;49;00m}[37m[39;49;00m
    [37m		    [39;49;00m},[37m    [39;49;00m
    [37m			[39;49;00m[94m"spec"[39;49;00m:{[37m[39;49;00m
    [37m			    [39;49;00m[94m"containers"[39;49;00m:[[37m[39;49;00m
    [37m				[39;49;00m{[37m[39;49;00m
    [37m                                [39;49;00m[94m"image"[39;49;00m:[37m [39;49;00m[33m"seldonio/mock_classifier:1.19.0-dev"[39;49;00m,[37m[39;49;00m
    [37m				[39;49;00m[94m"imagePullPolicy"[39;49;00m:[37m [39;49;00m[33m"IfNotPresent"[39;49;00m,[37m[39;49;00m
    [37m				[39;49;00m[94m"name"[39;49;00m:[37m [39;49;00m[33m"classifier-2"[39;49;00m,[37m[39;49;00m
    [37m				[39;49;00m[94m"resources"[39;49;00m:[37m [39;49;00m{[37m[39;49;00m
    [37m				    [39;49;00m[94m"requests"[39;49;00m:[37m [39;49;00m{[37m[39;49;00m
    [37m					[39;49;00m[94m"memory"[39;49;00m:[37m [39;49;00m[33m"1Mi"[39;49;00m[37m[39;49;00m
    [37m				    [39;49;00m}[37m[39;49;00m
    [37m				[39;49;00m}[37m[39;49;00m
    [37m			    [39;49;00m}[37m[39;49;00m
    [37m			[39;49;00m],[37m[39;49;00m
    [37m			[39;49;00m[94m"terminationGracePeriodSeconds"[39;49;00m:[37m [39;49;00m[34m20[39;49;00m[37m[39;49;00m
    [37m				   [39;49;00m}[37m[39;49;00m
    [37m				   [39;49;00m}],[37m[39;49;00m
    [37m		[39;49;00m[94m"graph"[39;49;00m:[37m [39;49;00m{[37m[39;49;00m
    [37m		    [39;49;00m[94m"name"[39;49;00m:[37m [39;49;00m[33m"release-name"[39;49;00m,[37m[39;49;00m
    [37m		    [39;49;00m[94m"implementation"[39;49;00m:[33m"RANDOM_ABTEST"[39;49;00m,[37m[39;49;00m
    [37m		    [39;49;00m[94m"parameters"[39;49;00m:[37m [39;49;00m[[37m[39;49;00m
    [37m			[39;49;00m{[37m[39;49;00m
    [37m			    [39;49;00m[94m"name"[39;49;00m:[33m"ratioA"[39;49;00m,[37m[39;49;00m
    [37m			    [39;49;00m[94m"value"[39;49;00m:[33m"0.5"[39;49;00m,[37m[39;49;00m
    [37m			    [39;49;00m[94m"type"[39;49;00m:[33m"FLOAT"[39;49;00m[37m[39;49;00m
    [37m			[39;49;00m}[37m[39;49;00m
    [37m		    [39;49;00m],[37m[39;49;00m
    [37m		    [39;49;00m[94m"children"[39;49;00m:[37m [39;49;00m[[37m[39;49;00m
    [37m			[39;49;00m{[37m[39;49;00m
    [37m			    [39;49;00m[94m"name"[39;49;00m:[37m [39;49;00m[33m"classifier-1"[39;49;00m,[37m[39;49;00m
    [37m			    [39;49;00m[94m"type"[39;49;00m:[33m"MODEL"[39;49;00m,[37m[39;49;00m
    [37m			    [39;49;00m[94m"children"[39;49;00m:[][37m[39;49;00m
    [37m			[39;49;00m},[37m[39;49;00m
    [37m			[39;49;00m{[37m[39;49;00m
    [37m			    [39;49;00m[94m"name"[39;49;00m:[37m [39;49;00m[33m"classifier-2"[39;49;00m,[37m[39;49;00m
    [37m			    [39;49;00m[94m"type"[39;49;00m:[33m"MODEL"[39;49;00m,[37m[39;49;00m
    [37m			    [39;49;00m[94m"children"[39;49;00m:[][37m[39;49;00m
    [37m			[39;49;00m}[37m   [39;49;00m
    [37m		    [39;49;00m][37m[39;49;00m
    [37m		[39;49;00m}[37m[39;49;00m
    [37m	    [39;49;00m}[37m[39;49;00m
    [37m	[39;49;00m][37m[39;49;00m
    [37m    [39;49;00m}[37m[39;49;00m
    }[37m[39;49;00m



```python
!kubectl wait sdep/myabtest \
  --for=condition=ready \
  --timeout=120s \
  -n seldon
```

    seldondeployment.machinelearning.seldon.io/myabtest condition met


### Get predictions

```python
from seldon_core.seldon_client import SeldonClient

sc = SeldonClient(
    deployment_name="myabtest",
    namespace="seldon",
    gateway_endpoint="localhost:8003",
    gateway="ambassador",
)
```

#### REST Request

```python
@retry(stop=stop_after_delay(300), wait=wait_exponential(multiplier=1, min=0.5, max=5))
def predict():
    r = sc.predict(transport="rest")
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
        values: 0.26095295840328658
      }
    }
    
    Response:
    {'data': {'names': ['proba'], 'tensor': {'shape': [1, 1], 'values': [0.0656377202541611]}}, 'meta': {'requestPath': {'classifier-2': 'seldonio/mock_classifier:1.19.0-dev'}}}


Response:
{'data': {'names': ['proba'], 'tensor': {'shape': [1, 1], 'values': [0.11299965170860979]}}, 'meta': {}}
```

#### gRPC Request

```python
@retry(stop=stop_after_delay(300), wait=wait_exponential(multiplier=1, min=0.5, max=5))
def predict():
    r = sc.predict(transport="grpc")
    assert r.success == True
    return r

predict()
```




    Success:True message:
    Request:
    {'meta': {}, 'data': {'tensor': {'shape': [1, 1], 'values': [0.163692647815478]}}}
    Response:
    {'meta': {'requestPath': {'classifier-2': 'seldonio/mock_classifier:1.19.0-dev'}}, 'data': {'names': ['proba'], 'tensor': {'shape': [1, 1], 'values': [0.05991890833984506]}}}




```python
!helm delete myabtest --namespace seldon
```

```
release "myabtest" uninstalled
```

## Serve REST Multi-Armed Bandit

```python
!helm upgrade -i mymab ../helm-charts/seldon-mab --namespace seldon
```

    Release "mymab" does not exist. Installing it now.
    NAME: mymab
    LAST DEPLOYED: Thu Dec  4 12:13:04 2025
    NAMESPACE: seldon
    STATUS: deployed
    REVISION: 1
    TEST SUITE: None



```python
!helm template ../helm-charts/seldon-mab | pygmentize -l json
```

    [34m---[39;49;00m[37m[39;49;00m
    [04m[91m#[39;49;00m[37m [39;49;00m[04m[91mS[39;49;00m[04m[91mo[39;49;00m[04m[91mu[39;49;00m[04m[91mr[39;49;00m[04m[91mc[39;49;00m[04m[91me[39;49;00m:[37m [39;49;00m[04m[91ms[39;49;00m[04m[91me[39;49;00m[04m[91ml[39;49;00m[04m[91md[39;49;00m[04m[91mo[39;49;00m[34mn[39;49;00m[34m-[39;49;00m[04m[91mm[39;49;00m[04m[91ma[39;49;00m[04m[91mb[39;49;00m[04m[91m/[39;49;00m[34mte[39;49;00m[04m[91mm[39;49;00m[04m[91mp[39;49;00m[04m[91ml[39;49;00m[04m[91ma[39;49;00m[34mtes[39;49;00m[04m[91m/[39;49;00m[04m[91mm[39;49;00m[04m[91ma[39;49;00m[04m[91mb[39;49;00m[04m[91m.[39;49;00m[04m[91mj[39;49;00m[04m[91ms[39;49;00m[04m[91mo[39;49;00m[34mn[39;49;00m[37m[39;49;00m
    {[37m[39;49;00m
    [37m    [39;49;00m[94m"apiVersion"[39;49;00m:[37m [39;49;00m[33m"machinelearning.seldon.io/v1alpha2"[39;49;00m,[37m[39;49;00m
    [37m    [39;49;00m[94m"kind"[39;49;00m:[37m [39;49;00m[33m"SeldonDeployment"[39;49;00m,[37m[39;49;00m
    [37m    [39;49;00m[94m"metadata"[39;49;00m:[37m [39;49;00m{[37m[39;49;00m
    [37m		[39;49;00m[94m"labels"[39;49;00m:[37m [39;49;00m{[94m"app"[39;49;00m:[33m"seldon"[39;49;00m},[37m[39;49;00m
    [37m		[39;49;00m[94m"name"[39;49;00m:[37m [39;49;00m[33m"release-name"[39;49;00m[37m[39;49;00m
    [37m    [39;49;00m},[37m[39;49;00m
    [37m    [39;49;00m[94m"spec"[39;49;00m:[37m [39;49;00m{[37m[39;49;00m
    [37m	[39;49;00m[94m"name"[39;49;00m:[37m [39;49;00m[33m"release-name"[39;49;00m,[37m[39;49;00m
    [37m	[39;49;00m[94m"predictors"[39;49;00m:[37m [39;49;00m[[37m[39;49;00m
    [37m	    [39;49;00m{[37m[39;49;00m
    [37m		[39;49;00m[94m"name"[39;49;00m:[37m [39;49;00m[33m"default"[39;49;00m,[37m[39;49;00m
    [37m		[39;49;00m[94m"replicas"[39;49;00m:[37m [39;49;00m[34m1[39;49;00m,[37m[39;49;00m
    [37m		[39;49;00m[94m"componentSpecs"[39;49;00m:[37m [39;49;00m[{[37m[39;49;00m
    [37m		    [39;49;00m[94m"spec"[39;49;00m:[37m [39;49;00m{[37m[39;49;00m
    [37m			[39;49;00m[94m"containers"[39;49;00m:[37m [39;49;00m[[37m[39;49;00m
    [37m			    [39;49;00m{[37m[39;49;00m
    [37m                                [39;49;00m[94m"image"[39;49;00m:[37m [39;49;00m[33m"seldonio/mock_classifier:1.19.0-dev"[39;49;00m,[37m				[39;49;00m
    [37m				[39;49;00m[94m"imagePullPolicy"[39;49;00m:[37m [39;49;00m[33m"IfNotPresent"[39;49;00m,[37m[39;49;00m
    [37m				[39;49;00m[94m"name"[39;49;00m:[37m [39;49;00m[33m"classifier-1"[39;49;00m,[37m[39;49;00m
    [37m				[39;49;00m[94m"resources"[39;49;00m:[37m [39;49;00m{[37m[39;49;00m
    [37m				    [39;49;00m[94m"requests"[39;49;00m:[37m [39;49;00m{[37m[39;49;00m
    [37m					[39;49;00m[94m"memory"[39;49;00m:[37m [39;49;00m[33m"1Mi"[39;49;00m[37m[39;49;00m
    [37m				    [39;49;00m}[37m[39;49;00m
    [37m				[39;49;00m}[37m[39;49;00m
    [37m			    [39;49;00m}],[37m[39;49;00m
    [37m			[39;49;00m[94m"terminationGracePeriodSeconds"[39;49;00m:[37m [39;49;00m[34m20[39;49;00m[37m[39;49;00m
    [37m		    [39;49;00m}},[37m[39;49;00m
    [37m	        [39;49;00m{[37m[39;49;00m
    [37m			[39;49;00m[94m"spec"[39;49;00m:{[37m[39;49;00m
    [37m			    [39;49;00m[94m"containers"[39;49;00m:[[37m[39;49;00m
    [37m				[39;49;00m{[37m[39;49;00m
    [37m                                [39;49;00m[94m"image"[39;49;00m:[37m [39;49;00m[33m"seldonio/mock_classifier:1.19.0-dev"[39;49;00m,[37m								    [39;49;00m
    [37m				[39;49;00m[94m"imagePullPolicy"[39;49;00m:[37m [39;49;00m[33m"IfNotPresent"[39;49;00m,[37m[39;49;00m
    [37m				[39;49;00m[94m"name"[39;49;00m:[37m [39;49;00m[33m"classifier-2"[39;49;00m,[37m[39;49;00m
    [37m				[39;49;00m[94m"resources"[39;49;00m:[37m [39;49;00m{[37m[39;49;00m
    [37m				    [39;49;00m[94m"requests"[39;49;00m:[37m [39;49;00m{[37m[39;49;00m
    [37m					[39;49;00m[94m"memory"[39;49;00m:[37m [39;49;00m[33m"1Mi"[39;49;00m[37m[39;49;00m
    [37m				    [39;49;00m}[37m[39;49;00m
    [37m				[39;49;00m}[37m[39;49;00m
    [37m			    [39;49;00m}[37m[39;49;00m
    [37m			[39;49;00m],[37m[39;49;00m
    [37m			[39;49;00m[94m"terminationGracePeriodSeconds"[39;49;00m:[37m [39;49;00m[34m20[39;49;00m[37m[39;49;00m
    [37m			[39;49;00m}[37m[39;49;00m
    [37m		[39;49;00m},[37m[39;49;00m
    [37m	        [39;49;00m{[37m[39;49;00m
    [37m		    [39;49;00m[94m"spec"[39;49;00m:{[37m[39;49;00m
    [37m			[39;49;00m[94m"containers"[39;49;00m:[37m [39;49;00m[{[37m[39;49;00m
    [37m                            [39;49;00m[94m"image"[39;49;00m:[37m [39;49;00m[33m"seldonio/mab_epsilon_greedy:1.19.0-dev"[39;49;00m,[37m								    			    [39;49;00m
    [37m			    [39;49;00m[94m"name"[39;49;00m:[37m [39;49;00m[33m"eg-router"[39;49;00m[37m[39;49;00m
    [37m			[39;49;00m}],[37m[39;49;00m
    [37m			[39;49;00m[94m"terminationGracePeriodSeconds"[39;49;00m:[37m [39;49;00m[34m20[39;49;00m[37m[39;49;00m
    [37m		    [39;49;00m}}[37m[39;49;00m
    [37m	        [39;49;00m],[37m[39;49;00m
    [37m		[39;49;00m[94m"graph"[39;49;00m:[37m [39;49;00m{[37m[39;49;00m
    [37m		    [39;49;00m[94m"name"[39;49;00m:[37m [39;49;00m[33m"eg-router"[39;49;00m,[37m[39;49;00m
    [37m		    [39;49;00m[94m"type"[39;49;00m:[33m"ROUTER"[39;49;00m,[37m[39;49;00m
    [37m		    [39;49;00m[94m"parameters"[39;49;00m:[37m [39;49;00m[[37m[39;49;00m
    [37m			[39;49;00m{[37m[39;49;00m
    [37m			    [39;49;00m[94m"name"[39;49;00m:[37m [39;49;00m[33m"n_branches"[39;49;00m,[37m[39;49;00m
    [37m			    [39;49;00m[94m"value"[39;49;00m:[37m [39;49;00m[33m"2"[39;49;00m,[37m[39;49;00m
    [37m			    [39;49;00m[94m"type"[39;49;00m:[37m [39;49;00m[33m"INT"[39;49;00m[37m[39;49;00m
    [37m			[39;49;00m},[37m[39;49;00m
    [37m			[39;49;00m{[37m[39;49;00m
    [37m			    [39;49;00m[94m"name"[39;49;00m:[37m [39;49;00m[33m"epsilon"[39;49;00m,[37m[39;49;00m
    [37m			    [39;49;00m[94m"value"[39;49;00m:[37m [39;49;00m[33m"0.2"[39;49;00m,[37m[39;49;00m
    [37m			    [39;49;00m[94m"type"[39;49;00m:[37m [39;49;00m[33m"FLOAT"[39;49;00m[37m[39;49;00m
    [37m			[39;49;00m},[37m[39;49;00m
    [37m			[39;49;00m{[37m[39;49;00m
    [37m			    [39;49;00m[94m"name"[39;49;00m:[37m [39;49;00m[33m"verbose"[39;49;00m,[37m[39;49;00m
    [37m			    [39;49;00m[94m"value"[39;49;00m:[37m [39;49;00m[33m"1"[39;49;00m,[37m[39;49;00m
    [37m			    [39;49;00m[94m"type"[39;49;00m:[37m [39;49;00m[33m"BOOL"[39;49;00m[37m[39;49;00m
    [37m			[39;49;00m}[37m[39;49;00m
    [37m		    [39;49;00m],[37m[39;49;00m
    [37m		    [39;49;00m[94m"children"[39;49;00m:[37m [39;49;00m[[37m[39;49;00m
    [37m			[39;49;00m{[37m[39;49;00m
    [37m			    [39;49;00m[94m"name"[39;49;00m:[37m [39;49;00m[33m"classifier-1"[39;49;00m,[37m[39;49;00m
    [37m			    [39;49;00m[94m"type"[39;49;00m:[33m"MODEL"[39;49;00m,[37m[39;49;00m
    [37m			    [39;49;00m[94m"children"[39;49;00m:[][37m[39;49;00m
    [37m			[39;49;00m},[37m[39;49;00m
    [37m			[39;49;00m{[37m[39;49;00m
    [37m			    [39;49;00m[94m"name"[39;49;00m:[37m [39;49;00m[33m"classifier-2"[39;49;00m,[37m[39;49;00m
    [37m			    [39;49;00m[94m"type"[39;49;00m:[33m"MODEL"[39;49;00m,[37m[39;49;00m
    [37m			    [39;49;00m[94m"children"[39;49;00m:[][37m[39;49;00m
    [37m			[39;49;00m}[37m   [39;49;00m
    [37m		    [39;49;00m][37m[39;49;00m
    [37m		[39;49;00m},[37m[39;49;00m
    [37m		[39;49;00m[94m"svcOrchSpec"[39;49;00m:[37m [39;49;00m{[37m[39;49;00m
    [37m		[39;49;00m[94m"resources"[39;49;00m:[37m [39;49;00m{[94m"requests"[39;49;00m:{[94m"cpu"[39;49;00m:[33m"0.1"[39;49;00m}},[37m[39;49;00m
    [94m"env"[39;49;00m:[37m [39;49;00m[[37m[39;49;00m
    {[37m[39;49;00m
    [94m"name"[39;49;00m:[37m [39;49;00m[33m"SELDON_LOG_MESSAGES_EXTERNALLY"[39;49;00m,[37m[39;49;00m
    [94m"value"[39;49;00m:[37m [39;49;00m[33m"false"[39;49;00m[37m[39;49;00m
    },[37m[39;49;00m
    {[37m[39;49;00m
    [94m"name"[39;49;00m:[37m [39;49;00m[33m"SELDON_LOG_MESSAGE_TYPE"[39;49;00m,[37m[39;49;00m
    [94m"value"[39;49;00m:[37m [39;49;00m[33m"seldon.message.pair"[39;49;00m[37m[39;49;00m
    },[37m[39;49;00m
    {[37m[39;49;00m
    [94m"name"[39;49;00m:[37m [39;49;00m[33m"SELDON_LOG_REQUESTS"[39;49;00m,[37m[39;49;00m
    [94m"value"[39;49;00m:[37m [39;49;00m[33m"false"[39;49;00m[37m[39;49;00m
    },[37m[39;49;00m
    {[37m[39;49;00m
    [94m"name"[39;49;00m:[37m [39;49;00m[33m"SELDON_LOG_RESPONSES"[39;49;00m,[37m[39;49;00m
    [94m"value"[39;49;00m:[37m [39;49;00m[33m"false"[39;49;00m[37m[39;49;00m
    },[37m[39;49;00m
    ][37m[39;49;00m
    },[37m[39;49;00m
    [37m		[39;49;00m[94m"labels"[39;49;00m:[37m [39;49;00m{[94m"fluentd"[39;49;00m:[33m"true"[39;49;00m,[94m"version"[39;49;00m:[33m"1.19.0-dev"[39;49;00m}[37m[39;49;00m
    [37m	    [39;49;00m}[37m[39;49;00m
    [37m	[39;49;00m][37m[39;49;00m
    [37m    [39;49;00m}[37m[39;49;00m
    }[37m[39;49;00m



```python
!kubectl wait sdep/mymab \
  --for=condition=ready \
  --timeout=120s \
  -n seldon
```

    seldondeployment.machinelearning.seldon.io/mymab condition met


### Get predictions

```python
from seldon_core.seldon_client import SeldonClient

sc = SeldonClient(
    deployment_name="mymab",
    namespace="seldon",
    gateway_endpoint="localhost:8003",
    gateway="ambassador",
)
```

#### REST Request

```python
@retry(stop=stop_after_delay(300), wait=wait_exponential(multiplier=1, min=0.5, max=5))
def predict():
    r = sc.predict(transport="rest")
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
        values: 0.13298918130078008
      }
    }
    
    Response:
    {'data': {'names': ['proba'], 'tensor': {'shape': [1, 1], 'values': [0.058212613572549546]}}, 'meta': {'requestPath': {'classifier-1': 'seldonio/mock_classifier:1.19.0-dev'}}}


Response:
{'data': {'names': ['proba'], 'tensor': {'shape': [1, 1], 'values': [0.05643175042558145]}}, 'meta': {}}
```

#### gRPC Request

```python
@retry(stop=stop_after_delay(300), wait=wait_exponential(multiplier=1, min=0.5, max=5))
def predict():
    r = sc.predict(transport="grpc")
    assert r.success == True
    return r

predict()
```




    Success:True message:
    Request:
    {'meta': {}, 'data': {'tensor': {'shape': [1, 1], 'values': [0.1279822023238696]}}}
    Response:
    {'meta': {'requestPath': {'classifier-1': 'seldonio/mock_classifier:1.19.0-dev'}}, 'data': {'names': ['proba'], 'tensor': {'shape': [1, 1], 'values': [0.05793871786684459]}}}




```python
!helm delete mymab --namespace seldon
```

    release "mymab" uninstalled

