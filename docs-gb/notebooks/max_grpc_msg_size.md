# Increasing the Maximum Message Size for gRPC


## Running this notebook

You will need to start Jupyter with settings to allow for large payloads, for example:

```
jupyter notebook --NotebookApp.iopub_data_rate_limit=1000000000
```


```python
from IPython.core.magic import register_line_cell_magic


@register_line_cell_magic
def writetemplate(line, cell):
    with open(line, "w") as f:
        f.write(cell.format(**globals()))
```

## Setup Seldon Core

Use the setup notebook to [Setup Cluster](https://docs.seldon.ai/seldon-core-1/tutorials/notebooks/seldon-core-setup#setup-cluster) with [Ambassador Ingress](https://docs.seldon.ai/seldon-core-1/tutorials/notebooks/seldon-core-setup#install-ingress).


```python
!kubectl create namespace seldon
```

    Error from server (AlreadyExists): namespaces "seldon" already exists



```python
VERSION = !cat ../version.txt
VERSION = VERSION[0]
VERSION
```




    '1.19.0-dev'



We now add in our model config file the annotations `"seldon.io/rest-timeout":"100000"` and `"seldon.io/grpc-timeout":"100000"`


```python
%%writetemplate resources/model_long_timeouts.yaml
apiVersion: machinelearning.seldon.io/v1alpha2
kind: SeldonDeployment
metadata:
  labels:
    app: seldon
  name: model-long-timeout
spec:
  annotations:
    deployment_version: v1
    seldon.io/grpc-timeout: '100000'
    seldon.io/rest-timeout: '100000'
  name: long-to
  predictors:
  - annotations:
      predictor_version: v1
    componentSpecs:
    - spec:
        containers:
        - image: seldonio/mock_classifier:{VERSION}
          imagePullPolicy: IfNotPresent
          name: classifier
          resources:
            requests:
              memory: 1Mi
        terminationGracePeriodSeconds: 20
    graph:
      children: []
      name: classifier
      type: MODEL
    name: test
    replicas: 1

```

## Create Seldon Deployment

Deploy the runtime graph to kubernetes.


```python
!kubectl apply -f resources/model_long_timeouts.yaml -n seldon
```

    seldondeployment.machinelearning.seldon.io/model-long-timeout created



```python
!kubectl wait sdep/model-long-timeout \
  --for=condition=ready \
  --timeout=120s \
  -n seldon
```

    seldondeployment.machinelearning.seldon.io/model-long-timeout condition met


## Get predictions


```python
from seldon_core.seldon_client import SeldonClient

sc = SeldonClient(
    deployment_name="model-long-timeout",
    namespace="seldon",
    grpc_max_send_message_length=50 * 1024 * 1024,
    grpc_max_receive_message_length=50 * 1024 * 1024,
)
```

    2025-12-04 11:06:35.990484: E external/local_xla/xla/stream_executor/cuda/cuda_fft.cc:477] Unable to register cuFFT factory: Attempting to register factory for plugin cuFFT when one has already been registered
    WARNING: All log messages before absl::InitializeLog() is called are written to STDERR
    E0000 00:00:1764846396.007483 3716075 cuda_dnn.cc:8310] Unable to register cuDNN factory: Attempting to register factory for plugin cuDNN when one has already been registered
    E0000 00:00:1764846396.013021 3716075 cuda_blas.cc:1418] Unable to register cuBLAS factory: Attempting to register factory for plugin cuBLAS when one has already been registered
    2025-12-04 11:06:36.032110: I tensorflow/core/platform/cpu_feature_guard.cc:210] This TensorFlow binary is optimized to use available CPU instructions in performance-critical operations.
    To enable the following instructions: AVX2 FMA, in other operations, rebuild TensorFlow with the appropriate compiler flags.


Send a small request which should succeed.


```python
from tenacity import retry, stop_after_delay, wait_exponential

@retry(stop=stop_after_delay(300), wait=wait_exponential(multiplier=1, min=0.5, max=5))
def predict():
    r = sc.predict(gateway="ambassador", transport="grpc")
    assert r.success == True
    
predict()
```

Send a large request which will fail as the default for the model will be 4G.


```python
@retry(stop=stop_after_delay(300), wait=wait_exponential(multiplier=1, min=0.5, max=5))
def predict():
    r = sc.predict(gateway="ambassador", transport="grpc", shape=(1000000, 1))
    assert r.success == False

predict()
```


```python
!kubectl delete -f resources/model_long_timeouts.yaml -n seldon
```

    seldondeployment.machinelearning.seldon.io "model-long-timeout" deleted


## Allowing larger gRPC messages

Now we change our SeldonDeployment to include a annotation for max grpx message size.


```python
%%writetemplate resources/model_grpc_size.yaml
apiVersion: machinelearning.seldon.io/v1alpha2
kind: SeldonDeployment
metadata:
  labels:
    app: seldon
  name: seldon-model-grpc-size
spec:
  annotations:
    seldon.io/grpc-max-message-size: '10000000'
    seldon.io/grpc-timeout: '100000'
    seldon.io/rest-timeout: '100000'
  name: test-deployment
  predictors:
  - annotations:
      predictor_version: v1
    componentSpecs:
    - spec:
        containers:
        - image: seldonio/mock_classifier:{VERSION}
          imagePullPolicy: IfNotPresent
          name: classifier
          resources:
            requests:
              memory: 1Mi
        terminationGracePeriodSeconds: 20
    graph:
      children: []
      endpoint:
        type: GRPC
      name: classifier
      type: MODEL
    name: grpc-size
    replicas: 1

```


```python
!kubectl apply -f resources/model_grpc_size.yaml -n seldon
```

    seldondeployment.machinelearning.seldon.io/seldon-model-grpc-size created



```python
!kubectl wait sdep/seldon-model-grpc-size \
  --for=condition=ready \
  --timeout=120s \
  -n seldon
```

    seldondeployment.machinelearning.seldon.io/seldon-model-grpc-size condition met


Send a request via ambassador. This should succeed.


```python
sc = SeldonClient(
    deployment_name="seldon-model-grpc-size",
    namespace="seldon",
    grpc_max_send_message_length=50 * 1024 * 1024,
    grpc_max_receive_message_length=50 * 1024 * 1024,
)
```


```python
@retry(stop=stop_after_delay(300), wait=wait_exponential(multiplier=1, min=0.5, max=5))
def predict():
    r = sc.predict(gateway="ambassador", transport="grpc", shape=(1000000, 1))
    assert r.success == True

predict()
```


```python
!kubectl delete -f resources/model_grpc_size.yaml -n seldon
```

    seldondeployment.machinelearning.seldon.io "seldon-model-grpc-size" deleted

