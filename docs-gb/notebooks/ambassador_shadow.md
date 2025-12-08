# Shadow Rollout with Seldon and Ambassador

This notebook shows how you can deploy "shadow" deployments to direct traffic not only to the main Seldon Deployment but also to a shadow deployment whose response will be dicarded. This allows you to test new models in a production setting and with production traffic and anlalyse how they perform before putting them live.

These are useful when you want to test a new model or higher latency inference piepline (e.g., with explanation components) with production traffic but without affecting the live deployment.


## Setup Seldon Core

Use the setup notebook to [Setup Cluster](https://docs.seldon.ai/seldon-core-1/tutorials/notebooks/seldon-core-setup#setup-cluster) with [Ambassador Ingress](https://docs.seldon.ai/seldon-core-1/tutorials/notebooks/seldon-core-setup#install-ingress).


```python
!kubectl create namespace seldon
```

    Error from server (AlreadyExists): namespaces "seldon" already exists



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




    '1.19.0-dev'



## Launch main model

We will create a very simple Seldon Deployment with a dummy model image `seldonio/mock_classifier`. This deployment is named `example`.


```python
%%writetemplate ambassador-example-model.yaml
apiVersion: machinelearning.seldon.io/v1alpha2
kind: SeldonDeployment
metadata:
  labels:
    app: seldon
  name: ambassador-example
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
!kubectl apply -f ambassador-example-model.yaml -n seldon
```

    seldondeployment.machinelearning.seldon.io/ambassador-example created



```python
!kubectl wait sdep/ambassador-example \
  --for=condition=ready \
  --timeout=120s \
  -n seldon
```

    seldondeployment.machinelearning.seldon.io/ambassador-example condition met


### Get predictions


```python
from seldon_core.seldon_client import SeldonClient

sc = SeldonClient(deployment_name="ambassador-example", namespace="seldon")
```

    2025-12-04 11:18:40.102525: E external/local_xla/xla/stream_executor/cuda/cuda_fft.cc:477] Unable to register cuFFT factory: Attempting to register factory for plugin cuFFT when one has already been registered
    WARNING: All log messages before absl::InitializeLog() is called are written to STDERR
    E0000 00:00:1764847120.119458 3736507 cuda_dnn.cc:8310] Unable to register cuDNN factory: Attempting to register factory for plugin cuDNN when one has already been registered
    E0000 00:00:1764847120.124778 3736507 cuda_blas.cc:1418] Unable to register cuBLAS factory: Attempting to register factory for plugin cuBLAS when one has already been registered
    2025-12-04 11:18:40.142665: I tensorflow/core/platform/cpu_feature_guard.cc:210] This TensorFlow binary is optimized to use available CPU instructions in performance-critical operations.
    To enable the following instructions: AVX2 FMA, in other operations, rebuild TensorFlow with the appropriate compiler flags.


#### REST Request


```python

from tenacity import retry, stop_after_delay, wait_exponential

@retry(stop=stop_after_delay(300), wait=wait_exponential(multiplier=1, min=0.5, max=5))
def make_prediction():
    r = sc.predict(gateway="ambassador", transport="rest")
    return r

r = make_prediction()
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
        values: 0.010110628961988777
      }
    }
    
    Response:
    {'data': {'names': ['proba'], 'tensor': {'shape': [1, 1], 'values': [0.051830424660114026]}}, 'meta': {'requestPath': {'classifier': 'seldonio/mock_classifier:1.19.0-dev'}}}


## Launch Shadow

We will now create a new Seldon Deployment for our Shadow deployment with a new model.


```python
%%writetemplate ambassador-example-model.yaml
apiVersion: machinelearning.seldon.io/v1alpha2
kind: SeldonDeployment
metadata:
  labels:
    app: seldon
  name: ambassador-example
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
!kubectl apply -f ambassador-example-model.yaml -n seldon
```

    seldondeployment.machinelearning.seldon.io/ambassador-example configured



```python
!kubectl wait sdep/ambassador-example \
  --for=condition=ready \
  --timeout=120s \
  -n seldon
```

    seldondeployment.machinelearning.seldon.io/ambassador-example condition met


Let's send a bunch of requests to the endpoint.


```python
import time

time.sleep(10) # wait before sending requests
for i in range(10):
    r = sc.predict(gateway="ambassador", transport="rest")
```


```python
@retry(stop=stop_after_delay(300), wait=wait_exponential(multiplier=1, min=0.5, max=5))
def get_requests_count():
    count = !kubectl logs -l seldon-app==ambassador-example-default -n seldon -c classifier --tail 1000 | grep "root.predict" | wc -l
    return int(count[0])

time.sleep(10)  # wait for logs to be flushed
default_count = get_requests_count()
print(f"main logs count {default_count}")

assert default_count == 11
```

    main logs count 11



```python
@retry(stop=stop_after_delay(300), wait=wait_exponential(multiplier=1, min=0.5, max=5))
def get_shadow_requests_count():
    count = !kubectl logs -l seldon-app==ambassador-example-shadow -n seldon -c classifier --tail 1000 | grep "root.predict" | wc -l
    return int(count[0])

time.sleep(10)  # wait for logs to be flushed
shadow_count = get_shadow_requests_count()
print(f"shadow logs count {shadow_count}")

assert shadow_count == 10
```

    shadow logs count 10


## TearDown


```python
!kubectl delete -f ambassador-example-model.yaml -n seldon
```

    seldondeployment.machinelearning.seldon.io "ambassador-example" deleted

