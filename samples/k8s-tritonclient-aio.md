## Tritonclient Examples with Seldon Core V2 (Asyncio)

- Note: for compatibility of Tritonclient check this issue https://github.com/SeldonIO/seldon-core-v2/issues/471


```python
import os
os.environ["NAMESPACE"] = "seldon-mesh"
```


```python
MESH_IP=!kubectl get svc seldon-mesh -n ${NAMESPACE} -o jsonpath='{.status.loadBalancer.ingress[0].ip}'
MESH_IP=MESH_IP[0]
import os
os.environ['MESH_IP'] = MESH_IP
MESH_IP
```




    '172.19.255.14'



## With MLServer

- Note: GRPC support with MLServer is blocked by https://github.com/SeldonIO/MLServer/issues/48
- Note: binary data support in HTTP is blocked by https://github.com/SeldonIO/MLServer/issues/324

### Deploy Model and Pipeline


```python
!cat models/sklearn-iris-gs.yaml
```

    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: iris
    spec:
      storageUri: "gs://seldon-models/mlserver/iris"
      requirements:
      - sklearn
      memory: 100Ki



```python
!cat pipelines/iris.yaml
```

    apiVersion: mlops.seldon.io/v1alpha1
    kind: Pipeline
    metadata:
      name: iris-pipeline
    spec:
      steps:
        - name: iris
      output:
        steps:
        - iris



```python
!kubectl apply -f models/sklearn-iris-gs.yaml -n ${NAMESPACE}
!kubectl apply -f pipelines/iris.yaml -n ${NAMESPACE}
```

    model.mlops.seldon.io/iris created
    pipeline.mlops.seldon.io/iris-pipeline created



```python
!kubectl wait --for condition=ready --timeout=300s model iris -n ${NAMESPACE}
!kubectl wait --for condition=ready --timeout=300s pipelines iris-pipeline -n ${NAMESPACE}
```

    model.mlops.seldon.io/iris condition met
    pipeline.mlops.seldon.io/iris-pipeline condition met


### HTTP Transport Protocol


```python
import tritonclient.http.aio as httpclient
import numpy as np

http_triton_client = httpclient.InferenceServerClient(
    url=f"{MESH_IP}:80",
    verbose=False,
)

print("model ready:", await http_triton_client.is_model_ready("iris"))
print("model metadata:", await http_triton_client.get_model_metadata("iris"))
```

    model ready: True
    model metadata: {'name': 'iris_1', 'versions': [], 'platform': '', 'inputs': [], 'outputs': [], 'parameters': {'content_type': None, 'headers': None}}


#### Against Model


```python
headers = {"content-type": "application/json"}

binary_data = False

inputs = [httpclient.InferInput("predict", (1, 4), "FP64")]
inputs[0].set_data_from_numpy(np.array([[1, 2, 3, 4]]).astype("float64"), binary_data=binary_data)

outputs = [httpclient.InferRequestedOutput("predict", binary_data=binary_data)]

result = await http_triton_client.infer("iris", inputs, outputs=outputs, headers=headers)
result.as_numpy("predict")
```




    array([2])



#### Against Pipeline


```python
headers = {"content-type": "application/json"}

binary_data = False

inputs = [httpclient.InferInput("predict", (1, 4), "FP64")]
inputs[0].set_data_from_numpy(np.array([[1, 2, 3, 4]]).astype("float64"), binary_data=binary_data)

outputs = [httpclient.InferRequestedOutput("predict", binary_data=binary_data)]

result = await http_triton_client.infer("iris-pipeline.pipeline", inputs, outputs=outputs, headers=headers)
result.as_numpy("predict")
```




    array([2])



### GRPC Transport Protocol

// Not supported with MLServer currently. 

### Cleanup


```python
await http_triton_client.close()
```


```python
!kubectl delete -f models/sklearn-iris-gs.yaml -n seldon-mesh
!kubectl delete -f pipelines/iris.yaml -n seldon-mesh
```

    model.mlops.seldon.io "iris" deleted
    pipeline.mlops.seldon.io "iris-pipeline" deleted


## With Tritonserver

### Deploy Model and Pipeline


```python
!cat models/tfsimple1.yaml
```

    apiVersion: mlops.seldon.io/v1alpha1
    kind: Model
    metadata:
      name: tfsimple1
    spec:
      storageUri: "gs://seldon-models/triton/simple"
      requirements:
      - tensorflow
      memory: 100Ki



```python
!cat pipelines/tfsimple.yaml
```

    apiVersion: mlops.seldon.io/v1alpha1
    kind: Pipeline
    metadata:
      name: tfsimple
    spec:
      steps:
        - name: tfsimple1
      output:
        steps:
        - tfsimple1



```python
!kubectl apply -f models/tfsimple1.yaml -n seldon-mesh
!kubectl apply -f pipelines/tfsimple.yaml -n seldon-mesh
```

    model.mlops.seldon.io/tfsimple1 created
    pipeline.mlops.seldon.io/tfsimple created



```python
!kubectl wait --for condition=ready --timeout=300s model tfsimple1 -n seldon-mesh
!kubectl wait --for condition=ready --timeout=300s pipelines tfsimple -n seldon-mesh
```

    model.mlops.seldon.io/tfsimple1 condition met
    pipeline.mlops.seldon.io/tfsimple condition met


### HTTP Transport Protocol


```python
import tritonclient.http.aio as httpclient
import numpy as np

http_triton_client = httpclient.InferenceServerClient(
    url=f"{MESH_IP}:80",
    verbose=False,
)

print("model ready:", await http_triton_client.is_model_ready("tfsimple1"))
print("model metadata:", await http_triton_client.get_model_metadata("tfsimple1"))
```

    model ready: True
    model metadata: {'name': 'tfsimple1_1', 'versions': ['1'], 'platform': 'tensorflow_graphdef', 'inputs': [{'name': 'INPUT0', 'datatype': 'INT32', 'shape': [-1, 16]}, {'name': 'INPUT1', 'datatype': 'INT32', 'shape': [-1, 16]}], 'outputs': [{'name': 'OUTPUT0', 'datatype': 'INT32', 'shape': [-1, 16]}, {'name': 'OUTPUT1', 'datatype': 'INT32', 'shape': [-1, 16]}]}


#### Against Model


```python
binary_data = False

inputs = [
    httpclient.InferInput("INPUT0", (1, 16), "INT32"),
    httpclient.InferInput("INPUT1", (1, 16), "INT32"),
]
inputs[0].set_data_from_numpy(np.arange(1, 17).reshape(-1, 16).astype("int32"), binary_data=binary_data)
inputs[1].set_data_from_numpy(np.arange(1, 17).reshape(-1, 16).astype("int32"), binary_data=binary_data)

outputs = [httpclient.InferRequestedOutput("OUTPUT0", binary_data=binary_data)]


result = await http_triton_client.infer("tfsimple1", inputs, outputs=outputs)
result.as_numpy("OUTPUT0")
```




    array([[ 2,  4,  6,  8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32]],
          dtype=int32)




```python
binary_data = True

inputs = [
    httpclient.InferInput("INPUT0", (1, 16), "INT32"),
    httpclient.InferInput("INPUT1", (1, 16), "INT32"),
]
inputs[0].set_data_from_numpy(np.arange(1, 17).reshape(-1, 16).astype("int32"), binary_data=binary_data)
inputs[1].set_data_from_numpy(np.arange(1, 17).reshape(-1, 16).astype("int32"), binary_data=binary_data)

outputs = [httpclient.InferRequestedOutput("OUTPUT0", binary_data=binary_data)]


result = await http_triton_client.infer("tfsimple1", inputs, outputs=outputs)
result.as_numpy("OUTPUT0")
```




    array([[ 2,  4,  6,  8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32]],
          dtype=int32)



#### Against Pipeline


```python
binary_data = False

inputs = [
    httpclient.InferInput("INPUT0", (1, 16), "INT32"),
    httpclient.InferInput("INPUT1", (1, 16), "INT32"),
]
inputs[0].set_data_from_numpy(np.arange(1, 17).reshape(-1, 16).astype("int32"), binary_data=binary_data)
inputs[1].set_data_from_numpy(np.arange(1, 17).reshape(-1, 16).astype("int32"), binary_data=binary_data)

outputs = [httpclient.InferRequestedOutput("OUTPUT0", binary_data=binary_data)]


result = await http_triton_client.infer("tfsimple.pipeline", inputs, outputs=outputs)
result.as_numpy("OUTPUT0")
```




    array([[ 2,  4,  6,  8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32]],
          dtype=int32)




```python
# binary data does not work with http behind pipeline: no opened issue yet

# binary_data = True

# inputs = [
#     httpclient.InferInput("INPUT0", (1, 16), "INT32"),
#     httpclient.InferInput("INPUT1", (1, 16), "INT32"),
# ]
# inputs[0].set_data_from_numpy(np.arange(1, 17).reshape(-1, 16).astype("int32"), binary_data=binary_data)
# inputs[1].set_data_from_numpy(np.arange(1, 17).reshape(-1, 16).astype("int32"), binary_data=binary_data)

# outputs = [httpclient.InferRequestedOutput("OUTPUT0", binary_data=binary_data)]


# result = await http_triton_client.infer("tfsimple.pipeline", inputs, outputs=outputs)
# result.as_numpy("OUTPUT0")
```

### GRPC Transport Protocol


```python
import tritonclient.grpc.aio as grpcclient
import numpy as np


grpc_triton_client = grpcclient.InferenceServerClient(
    url=f"{MESH_IP}:80",
    verbose=False,
)
```


```python
model_name = "tfsimple1"
headers = {"seldon-model": model_name}

print("model ready:", await grpc_triton_client.is_model_ready(model_name, headers=headers))
print(await grpc_triton_client.get_model_metadata(model_name, headers=headers))
```

    model ready: True
    name: "tfsimple1_1"
    versions: "1"
    platform: "tensorflow_graphdef"
    inputs {
      name: "INPUT0"
      datatype: "INT32"
      shape: -1
      shape: 16
    }
    inputs {
      name: "INPUT1"
      datatype: "INT32"
      shape: -1
      shape: 16
    }
    outputs {
      name: "OUTPUT0"
      datatype: "INT32"
      shape: -1
      shape: 16
    }
    outputs {
      name: "OUTPUT1"
      datatype: "INT32"
      shape: -1
      shape: 16
    }
    


#### Against Model


```python
model_name = "tfsimple1"
headers = {"seldon-model": model_name}

inputs = [
    grpcclient.InferInput("INPUT0", (1, 16), "INT32"),
    grpcclient.InferInput("INPUT1", (1, 16), "INT32"),
]
inputs[0].set_data_from_numpy(np.arange(1, 17).reshape(-1, 16).astype("int32"))
inputs[1].set_data_from_numpy(np.arange(1, 17).reshape(-1, 16).astype("int32"))

outputs = [grpcclient.InferRequestedOutput("OUTPUT0")]


result = await grpc_triton_client.infer(model_name, inputs, outputs=outputs, headers=headers)
result.as_numpy("OUTPUT0")
```




    array([[ 2,  4,  6,  8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32]],
          dtype=int32)



#### Against Pipeline


```python
model_name = "tfsimple.pipeline"
headers = {"seldon-model": model_name}

inputs = [
    grpcclient.InferInput("INPUT0", (1, 16), "INT32"),
    grpcclient.InferInput("INPUT1", (1, 16), "INT32"),
]
inputs[0].set_data_from_numpy(np.arange(1, 17).reshape(-1, 16).astype("int32"))
inputs[1].set_data_from_numpy(np.arange(1, 17).reshape(-1, 16).astype("int32"))

outputs = [grpcclient.InferRequestedOutput("OUTPUT0")]


result = await grpc_triton_client.infer(model_name, inputs, outputs=outputs, headers=headers)
result.as_numpy("OUTPUT0")
```




    array([[ 2,  4,  6,  8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32]],
          dtype=int32)



## Cleanup


```python
await http_triton_client.close()
await grpc_triton_client.close()
```


```python
!kubectl delete -f models/tfsimple1.yaml -n seldon-mesh
!kubectl delete -f pipelines/tfsimple.yaml -n seldon-mesh
```

    model.mlops.seldon.io "tfsimple1" deleted
    pipeline.mlops.seldon.io "tfsimple" deleted

