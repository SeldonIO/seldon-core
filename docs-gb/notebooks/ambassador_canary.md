# Canary Rollout with Seldon and Ambassador


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
!kubectl apply -f model.yaml -n seldon
```

    seldondeployment.machinelearning.seldon.io/example created



```python
!kubectl wait sdep/example \
  --for=condition=ready \
  --timeout=120s \
  -n seldon
```

    seldondeployment.machinelearning.seldon.io/example condition met


### Get predictions


```python
from seldon_core.seldon_client import SeldonClient

sc = SeldonClient(deployment_name="example", namespace="seldon")
```

    2025-12-04 11:09:46.205128: E external/local_xla/xla/stream_executor/cuda/cuda_fft.cc:477] Unable to register cuFFT factory: Attempting to register factory for plugin cuFFT when one has already been registered
    WARNING: All log messages before absl::InitializeLog() is called are written to STDERR
    E0000 00:00:1764846586.222049 3721772 cuda_dnn.cc:8310] Unable to register cuDNN factory: Attempting to register factory for plugin cuDNN when one has already been registered
    E0000 00:00:1764846586.227263 3721772 cuda_blas.cc:1418] Unable to register cuBLAS factory: Attempting to register factory for plugin cuBLAS when one has already been registered
    2025-12-04 11:09:46.246078: I tensorflow/core/platform/cpu_feature_guard.cc:210] This TensorFlow binary is optimized to use available CPU instructions in performance-critical operations.
    To enable the following instructions: AVX2 FMA, in other operations, rebuild TensorFlow with the appropriate compiler flags.


#### REST Request


```python
from tenacity import retry, stop_after_delay, wait_exponential

@retry(stop=stop_after_delay(300), wait=wait_exponential(multiplier=1, min=0.5, max=5))
def make_prediction():
    r = sc.predict(gateway="ambassador", transport="rest")
    assert r.success == True
    return r

make_prediction()
```




    Success:True message:
    Request:
    meta {
    }
    data {
      tensor {
        shape: 1
        shape: 1
        values: 0.68551705136886776
      }
    }
    
    Response:
    {'data': {'names': ['proba'], 'tensor': {'shape': [1, 1], 'values': [0.09698790957732062]}}, 'meta': {'requestPath': {'classifier': 'seldonio/mock_classifier:1.19.0-dev'}}}



## Launch Canary

We will now extend the existing graph and add a new predictor as a canary using a new model `seldonio/mock_classifier_rest`. We will add traffic values to split traffic 75/25 to the main and canary.


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
!kubectl apply -f canary.yaml -n seldon
```

    seldondeployment.machinelearning.seldon.io/example configured



```python
!kubectl wait sdep/example \
  --for=condition=ready \
  --timeout=120s \
  -n seldon
```

    seldondeployment.machinelearning.seldon.io/example condition met


Show our REST requests are now split with roughly 25% going to the canary.


```python
@retry(stop=stop_after_delay(300), wait=wait_exponential(multiplier=1, min=0.5, max=5))
def make_prediction():
    r = sc.predict(gateway="ambassador", transport="rest")
    assert r.success == True
    return r

make_prediction()
```




    Success:True message:
    Request:
    meta {
    }
    data {
      tensor {
        shape: 1
        shape: 1
        values: 0.83946571577703166
      }
    }
    
    Response:
    {'data': {'names': ['proba'], 'tensor': {'shape': [1, 1], 'values': [0.11133259692564312]}}, 'meta': {'requestPath': {'classifier': 'seldonio/mock_classifier:1.19.0-dev'}}}




```python
from collections import defaultdict
import time

counts = defaultdict(int)
n = 100
time.sleep(10)  # wait before sending requests
for i in range(n):
    r = sc.predict(gateway="ambassador", transport="rest")
```

Following checks number of prediction requests processed by default/canary predictors respectively.


```python
import time

@retry(stop=stop_after_delay(300), wait=wait_exponential(multiplier=1, min=0.5, max=5))
def get_requests_cound_for_main():
    default_count = !kubectl logs -l seldon-app=example-main -n seldon -c classifier --tail 1000 | grep "root:predict" | wc -l
    return float(default_count[0])


time.sleep(10) # wait for logs to be flushed
default_count = get_requests_cound_for_main()
print(f"main logs count {default_count}")
```

    main logs count 75.0



```python
@retry(stop=stop_after_delay(300), wait=wait_exponential(multiplier=1, min=0.5, max=5))
def get_requests_cound_for_canary():
    canary_count = !kubectl logs -l seldon-app=example-canary -n seldon -c classifier --tail 1000 | grep "root:predict" | wc -l
    return float(canary_count[0])

time.sleep(10) # wait for logs to be flushed
canary_count = get_requests_cound_for_canary()
print(f"canary logs count {canary_count}")
```

    canary logs count 27.0



```python
canary_percentage = canary_count / default_count
print(f"canary percentage {canary_percentage}")
assert canary_percentage > 0.1 and canary_percentage < 0.5
```

    canary percentage 0.36



```python
!kubectl delete -f canary.yaml -n seldon
```

    seldondeployment.machinelearning.seldon.io "example" deleted

