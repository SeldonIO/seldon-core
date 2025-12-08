# Custom Header Routing with Seldon and Ambassador

This notebook shows how you can deploy Seldon Deployments which can have custom routing via Ambassador's custom header routing.


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

    2025-12-04 11:15:20.327410: E external/local_xla/xla/stream_executor/cuda/cuda_fft.cc:477] Unable to register cuFFT factory: Attempting to register factory for plugin cuFFT when one has already been registered
    WARNING: All log messages before absl::InitializeLog() is called are written to STDERR
    E0000 00:00:1764846920.344554 3730764 cuda_dnn.cc:8310] Unable to register cuDNN factory: Attempting to register factory for plugin cuDNN when one has already been registered
    E0000 00:00:1764846920.350034 3730764 cuda_blas.cc:1418] Unable to register cuBLAS factory: Attempting to register factory for plugin cuBLAS when one has already been registered
    2025-12-04 11:15:20.369431: I tensorflow/core/platform/cpu_feature_guard.cc:210] This TensorFlow binary is optimized to use available CPU instructions in performance-critical operations.
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
        values: 0.39252928301617063
      }
    }
    
    Response:
    {'data': {'names': ['proba'], 'tensor': {'shape': [1, 1], 'values': [0.07418328803460136]}}, 'meta': {'requestPath': {'classifier': 'seldonio/mock_classifier:1.19.0-dev'}}}



## Launch Model with Custom Routing

To make it a canary of the original `example` deployment we add two annotations

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
!kubectl create -f model_with_header.yaml -n seldon
```

    seldondeployment.machinelearning.seldon.io/example-header created



```python
!kubectl wait sdep/example-header \
  --for=condition=ready \
  --timeout=120s \
  -n seldon
```

    seldondeployment.machinelearning.seldon.io/example-header condition met


Check a request without a header goes to the existing model.


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
        values: 0.00838913842242972
      }
    }
    
    Response:
    {'data': {'names': ['proba'], 'tensor': {'shape': [1, 1], 'values': [0.0517458889162658]}}, 'meta': {'requestPath': {'classifier': 'seldonio/mock_classifier:1.19.0-dev'}}}




```python
import time

@retry(stop=stop_after_delay(300), wait=wait_exponential(multiplier=1, min=0.5, max=5))
def get_requests_count():
    default_count = !kubectl logs -l seldon-app=example-single -n seldon -c classifier --tail 1000 | grep "root:predict" | wc -l
    return int(default_count[0])

time.sleep(10)  # wait for logs to be flushed
default_count = get_requests_count()
print(f"main logs count {default_count}")

assert default_count == 2
```

    main logs count 2


Check a REST request with the required header gets routed to the new model.


```python
@retry(stop=stop_after_delay(300), wait=wait_exponential(multiplier=1, min=0.5, max=5))
def make_prediction():
    r = sc.predict(gateway="ambassador", transport="rest", headers={"location": "london"})
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
        values: 0.86335349963397356
      }
    }
    
    Response:
    {'data': {'names': ['proba'], 'tensor': {'shape': [1, 1], 'values': [0.11371803202813073]}}, 'meta': {'requestPath': {'classifier': 'seldonio/mock_classifier:1.19.0-dev'}}}




```python
@retry(stop=stop_after_delay(300), wait=wait_exponential(multiplier=1, min=0.5, max=5))
def get_requests_count_for_specific_header():
    count = !kubectl logs -l seldon-app=example-header-single -n seldon -c classifier --tail 1000 | grep "root:predict" | wc -l
    return int(count[0])

time.sleep(10)  # wait for logs to be flushed
header_count = get_requests_count_for_specific_header()
print(f"header logs count {header_count}")

assert header_count == 1
```

    header logs count 1



```python
!kubectl delete -f model.yaml -n seldon
```

    seldondeployment.machinelearning.seldon.io "example" deleted



```python
!kubectl delete -f model_with_header.yaml -n seldon
```

    seldondeployment.machinelearning.seldon.io "example-header" deleted

