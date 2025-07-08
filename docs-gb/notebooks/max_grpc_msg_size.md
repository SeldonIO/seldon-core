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

Use the setup notebook to [Setup Cluster](https://docs.seldon.io/projects/seldon-core/en/latest/examples/seldon_core_setup.html#Setup-Cluster) with [Ambassador Ingress](https://docs.seldon.io/projects/seldon-core/en/latest/examples/seldon_core_setup.html#Ambassador) and [Install Seldon Core](https://docs.seldon.io/projects/seldon-core/en/latest/examples/seldon_core_setup.html#Install-Seldon-Core). Instructions [also online](https://docs.seldon.io/projects/seldon-core/en/latest/examples/seldon_core_setup.html).


```python
!kubectl create namespace seldon
```

    Error from server (AlreadyExists): namespaces "seldon" already exists



```python
!kubectl config set-context $(kubectl config current-context) --namespace=seldon
```

    Context "kind-kind" modified.



```python
VERSION = !cat ../version.txt
VERSION = VERSION[0]
VERSION
```




    '1.5.0-dev'



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
!kubectl rollout status deploy/$(kubectl get deploy -l seldon-deployment-id=model-long-timeout -o jsonpath='{.items[0].metadata.name}')
```

    deployment "model-long-timeout-test-0-classifier" successfully rolled out


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

Send a small request which should succeed.


```python
r = sc.predict(gateway="ambassador", transport="grpc")
assert r.success == True
print(r)
```

    Success:True message:
    Request:
    {'meta': {}, 'data': {'tensor': {'shape': [1, 1], 'values': [0.4806932754099743]}}}
    Response:
    {'meta': {}, 'data': {'names': ['proba'], 'tensor': {'shape': [1, 1], 'values': [0.08047035772935462]}}}


Send a large request which will fail as the default for the model will be 4G.


```python
r = sc.predict(gateway="ambassador", transport="grpc", shape=(1000000, 1))
print(r.success, r.msg)
```

    False <_InactiveRpcError of RPC that terminated with:
    	status = StatusCode.RESOURCE_EXHAUSTED
    	details = "Received message larger than max (8000023 vs. 4194304)"
    	debug_error_string = "{"created":"@1603887710.710555595","description":"Error received from peer ipv6:[::1]:8003","file":"src/core/lib/surface/call.cc","file_line":1061,"grpc_message":"Received message larger than max (8000023 vs. 4194304)","grpc_status":8}"
    >



```python
!kubectl delete -f resources/model_long_timeouts.json
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
  name: seldon-model
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
!kubectl create -f resources/model_grpc_size.yaml -n seldon
```

    seldondeployment.machinelearning.seldon.io/seldon-model created



```python
!kubectl rollout status deploy/$(kubectl get deploy -l seldon-deployment-id=seldon-model -o jsonpath='{.items[0].metadata.name}')
```

    Waiting for deployment "seldon-model-grpc-size-0-classifier" rollout to finish: 0 of 1 updated replicas are available...
    deployment "seldon-model-grpc-size-0-classifier" successfully rolled out


Send a request via ambassador. This should succeed.


```python
sc = SeldonClient(
    deployment_name="seldon-model",
    namespace="seldon",
    grpc_max_send_message_length=50 * 1024 * 1024,
    grpc_max_receive_message_length=50 * 1024 * 1024,
)
r = sc.predict(gateway="ambassador", transport="grpc", shape=(1000000, 1))
assert r.success == True
print(r.success)
```

    True



```python
!kubectl delete -f resources/model_grpc_size.json -n seldon
```

    seldondeployment.machinelearning.seldon.io "seldon-model" deleted



```python

```
