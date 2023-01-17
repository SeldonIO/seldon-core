### Seldon V2 Batch Examples

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

```
'172.19.255.1'

```

## Deploy Models and Pipelines

```bash
cat models/sklearn-iris-gs.yaml
```

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Model
metadata:
  name: iris
spec:
  storageUri: "gs://seldon-models/mlserver/iris"
  requirements:
  - sklearn
  memory: 100Ki

```

```bash
cat pipelines/iris.yaml
```

```yaml
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

```

```bash
cat models/tfsimple1.yaml
```

```yaml
apiVersion: mlops.seldon.io/v1alpha1
kind: Model
metadata:
  name: tfsimple1
spec:
  storageUri: "gs://seldon-models/triton/simple"
  requirements:
  - tensorflow
  memory: 100Ki

```

```bash
cat pipelines/tfsimple.yaml
```

```yaml
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

```

```bash
kubectl apply -f models/sklearn-iris-gs.yaml -n ${NAMESPACE}
kubectl apply -f pipelines/iris.yaml -n ${NAMESPACE}

kubectl apply -f models/tfsimple1.yaml -n ${NAMESPACE}
kubectl apply -f pipelines/tfsimple.yaml -n ${NAMESPACE}
```

```
model.mlops.seldon.io/iris created
pipeline.mlops.seldon.io/iris-pipeline created
model.mlops.seldon.io/tfsimple1 created
pipeline.mlops.seldon.io/tfsimple created

```

```bash
kubectl wait --for condition=ready --timeout=300s model --all -n ${NAMESPACE}
kubectl wait --for condition=ready --timeout=300s pipelines --all -n ${NAMESPACE}
```

```
model.mlops.seldon.io/iris condition met
model.mlops.seldon.io/tfsimple1 condition met
pipeline.mlops.seldon.io/iris-pipeline condition met
pipeline.mlops.seldon.io/tfsimple condition met

```

## Test Predictions

```bash
seldon model infer iris --inference-host ${MESH_IP}:80 \
  '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}' | jq -M .
```

```json
{
  "model_name": "iris_1",
  "model_version": "1",
  "id": "fce6921c-9828-40ce-99ff-a9ef76cff361",
  "parameters": {},
  "outputs": [
    {
      "name": "predict",
      "shape": [
        1,
        1
      ],
      "datatype": "INT64",
      "data": [
        2
      ]
    }
  ]
}

```

```bash
seldon pipeline infer iris-pipeline --inference-host ${MESH_IP}:80 \
  '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}' |  jq -M .
```

```json
{
  "model_name": "",
  "outputs": [
    {
      "data": [
        2
      ],
      "name": "predict",
      "shape": [
        1,
        1
      ],
      "datatype": "INT64"
    }
  ]
}

```

```bash
seldon model infer tfsimple1 --inference-host ${MESH_IP}:80 \
  '{"outputs":[{"name":"OUTPUT0"}], "inputs":[{"name":"INPUT0","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,16]},{"name":"INPUT1","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,16]}]}' | jq -M .
```

```json
{
  "model_name": "tfsimple1_1",
  "model_version": "1",
  "outputs": [
    {
      "name": "OUTPUT0",
      "datatype": "INT32",
      "shape": [
        1,
        16
      ],
      "data": [
        2,
        4,
        6,
        8,
        10,
        12,
        14,
        16,
        18,
        20,
        22,
        24,
        26,
        28,
        30,
        32
      ]
    }
  ]
}

```

```bash
seldon pipeline infer tfsimple --inference-host ${MESH_IP}:80 \
  '{"outputs":[{"name":"OUTPUT0"}], "inputs":[{"name":"INPUT0","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,16]},{"name":"INPUT1","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,16]}]}' | jq -M .
```

```json
{
  "model_name": "",
  "outputs": [
    {
      "data": [
        2,
        4,
        6,
        8,
        10,
        12,
        14,
        16,
        18,
        20,
        22,
        24,
        26,
        28,
        30,
        32
      ],
      "name": "OUTPUT0",
      "shape": [
        1,
        16
      ],
      "datatype": "INT32"
    },
    {
      "data": [
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0
      ],
      "name": "OUTPUT1",
      "shape": [
        1,
        16
      ],
      "datatype": "INT32"
    }
  ]
}

```

## MLServer Iris Batch Job

```bash
cat batch-inputs/iris-input.txt | head -n 1 | jq -M .
```

```json
{
  "inputs": [
    {
      "name": "predict",
      "data": [
        0.38606369295833043,
        0.006894049558299753,
        0.6104082981607108,
        0.3958954239450676
      ],
      "datatype": "FP64",
      "shape": [
        1,
        4
      ]
    }
  ]
}

```

```bash
%%bash
mlserver infer -u ${MESH_IP} -m iris -i batch-inputs/iris-input.txt -o /tmp/iris-output.txt --workers 5

```

```
2023-01-17 11:50:04,649 [mlserver] INFO - Using asyncio event-loop policy: uvloop
2023-01-17 11:50:04,650 [mlserver] INFO - server url: 172.19.255.1
2023-01-17 11:50:04,650 [mlserver] INFO - model name: iris
2023-01-17 11:50:04,650 [mlserver] INFO - request headers: {}
2023-01-17 11:50:04,650 [mlserver] INFO - input file path: batch-inputs/iris-input.txt
2023-01-17 11:50:04,650 [mlserver] INFO - output file path: /tmp/iris-output.txt
2023-01-17 11:50:04,650 [mlserver] INFO - workers: 5
2023-01-17 11:50:04,650 [mlserver] INFO - retries: 3
2023-01-17 11:50:04,650 [mlserver] INFO - batch interval: 0.0
2023-01-17 11:50:04,650 [mlserver] INFO - batch jitter: 0.0
2023-01-17 11:50:04,650 [mlserver] INFO - connection timeout: 60
2023-01-17 11:50:04,650 [mlserver] INFO - micro-batch size: 1
2023-01-17 11:50:04,771 [mlserver] INFO - Finalizer: processed instances: 100
2023-01-17 11:50:04,771 [mlserver] INFO - Total processed instances: 100
2023-01-17 11:50:04,771 [mlserver] INFO - Time taken: 0.12 seconds

```

```bash
%%bash
mlserver infer -u ${MESH_IP} -m iris-pipeline.pipeline -i batch-inputs/iris-input.txt -o /tmp/iris-pipeline-output.txt --workers 5

```

```
2023-01-17 11:50:09,010 [mlserver] INFO - Using asyncio event-loop policy: uvloop
2023-01-17 11:50:09,012 [mlserver] INFO - server url: 172.19.255.1
2023-01-17 11:50:09,012 [mlserver] INFO - model name: iris-pipeline.pipeline
2023-01-17 11:50:09,012 [mlserver] INFO - request headers: {}
2023-01-17 11:50:09,012 [mlserver] INFO - input file path: batch-inputs/iris-input.txt
2023-01-17 11:50:09,012 [mlserver] INFO - output file path: /tmp/iris-pipeline-output.txt
2023-01-17 11:50:09,012 [mlserver] INFO - workers: 5
2023-01-17 11:50:09,012 [mlserver] INFO - retries: 3
2023-01-17 11:50:09,012 [mlserver] INFO - batch interval: 0.0
2023-01-17 11:50:09,012 [mlserver] INFO - batch jitter: 0.0
2023-01-17 11:50:09,012 [mlserver] INFO - connection timeout: 60
2023-01-17 11:50:09,012 [mlserver] INFO - micro-batch size: 1
2023-01-17 11:50:09,281 [mlserver] INFO - Finalizer: processed instances: 100
2023-01-17 11:50:09,281 [mlserver] INFO - Total processed instances: 100
2023-01-17 11:50:09,281 [mlserver] INFO - Time taken: 0.27 seconds

```

```bash
cat /tmp/iris-output.txt | head -n 1 | jq -M .
```

```json
{
  "model_name": "iris_1",
  "model_version": "1",
  "id": "68d96fb3-5176-4802-9ca8-a1a918da9205",
  "parameters": {
    "batch_index": 0,
    "inference_id": "68d96fb3-5176-4802-9ca8-a1a918da9205"
  },
  "outputs": [
    {
      "name": "predict",
      "shape": [
        1,
        1
      ],
      "datatype": "INT64",
      "data": [
        0
      ]
    }
  ]
}

```

```bash
cat /tmp/iris-pipeline-output.txt | head -n 1 | jq .
```

```json
{
  "model_name": "",
  "id": "ded889d6-efd6-45b5-8df2-e88af5b166fc",
  "parameters": {
    "batch_index": 0
  },
  "outputs": [
    {
      "name": "predict",
      "shape": [
        1,
        1
      ],
      "datatype": "INT64",
      "data": [
        0
      ]
    }
  ]
}

```

## Triton TFSimple Batch Job

```bash
cat batch-inputs/tfsimple-input.txt | head -n 1 | jq -M .
```

```json
{
  "inputs": [
    {
      "name": "INPUT0",
      "data": [
        75,
        39,
        9,
        44,
        32,
        97,
        99,
        40,
        13,
        27,
        25,
        36,
        18,
        77,
        62,
        60
      ],
      "datatype": "INT32",
      "shape": [
        1,
        16
      ]
    },
    {
      "name": "INPUT1",
      "data": [
        39,
        7,
        14,
        58,
        13,
        88,
        98,
        66,
        97,
        57,
        49,
        3,
        49,
        63,
        37,
        12
      ],
      "datatype": "INT32",
      "shape": [
        1,
        16
      ]
    }
  ]
}

```

```bash
%%bash
mlserver infer -u ${MESH_IP} -m tfsimple1 -i batch-inputs/tfsimple-input.txt -o /tmp/tfsimple-output.txt --workers 5 -b

```

```
2023-01-17 11:50:17,224 [mlserver] INFO - Using asyncio event-loop policy: uvloop
2023-01-17 11:50:17,225 [mlserver] INFO - server url: 172.19.255.1
2023-01-17 11:50:17,226 [mlserver] INFO - model name: tfsimple1
2023-01-17 11:50:17,226 [mlserver] INFO - request headers: {}
2023-01-17 11:50:17,226 [mlserver] INFO - input file path: batch-inputs/tfsimple-input.txt
2023-01-17 11:50:17,226 [mlserver] INFO - output file path: /tmp/tfsimple-output.txt
2023-01-17 11:50:17,226 [mlserver] INFO - workers: 5
2023-01-17 11:50:17,226 [mlserver] INFO - retries: 3
2023-01-17 11:50:17,226 [mlserver] INFO - batch interval: 0.0
2023-01-17 11:50:17,226 [mlserver] INFO - batch jitter: 0.0
2023-01-17 11:50:17,226 [mlserver] INFO - connection timeout: 60
2023-01-17 11:50:17,226 [mlserver] INFO - micro-batch size: 1
2023-01-17 11:50:17,326 [mlserver] INFO - Finalizer: processed instances: 100
2023-01-17 11:50:17,327 [mlserver] INFO - Total processed instances: 100
2023-01-17 11:50:17,327 [mlserver] INFO - Time taken: 0.10 seconds

```

```bash
%%bash
mlserver infer -u ${MESH_IP} -m tfsimple.pipeline -i batch-inputs/tfsimple-input.txt -o /tmp/tfsimple-pipeline-output.txt --workers 5

```

```
2023-01-17 11:50:18,437 [mlserver] INFO - Using asyncio event-loop policy: uvloop
2023-01-17 11:50:18,439 [mlserver] INFO - server url: 172.19.255.1
2023-01-17 11:50:18,439 [mlserver] INFO - model name: tfsimple.pipeline
2023-01-17 11:50:18,439 [mlserver] INFO - request headers: {}
2023-01-17 11:50:18,439 [mlserver] INFO - input file path: batch-inputs/tfsimple-input.txt
2023-01-17 11:50:18,439 [mlserver] INFO - output file path: /tmp/tfsimple-pipeline-output.txt
2023-01-17 11:50:18,439 [mlserver] INFO - workers: 5
2023-01-17 11:50:18,439 [mlserver] INFO - retries: 3
2023-01-17 11:50:18,439 [mlserver] INFO - batch interval: 0.0
2023-01-17 11:50:18,439 [mlserver] INFO - batch jitter: 0.0
2023-01-17 11:50:18,439 [mlserver] INFO - connection timeout: 60
2023-01-17 11:50:18,439 [mlserver] INFO - micro-batch size: 1
2023-01-17 11:50:18,709 [mlserver] INFO - Finalizer: processed instances: 100
2023-01-17 11:50:18,709 [mlserver] INFO - Total processed instances: 100
2023-01-17 11:50:18,710 [mlserver] INFO - Time taken: 0.27 seconds

```

```bash
cat /tmp/tfsimple-output.txt | head -n 1 | jq -M .
```

```json
{
  "model_name": "tfsimple1_1",
  "model_version": "1",
  "id": "54517ac7-daed-4a42-a699-bab4e3271922",
  "parameters": {
    "batch_index": 1,
    "inference_id": "54517ac7-daed-4a42-a699-bab4e3271922"
  },
  "outputs": [
    {
      "name": "OUTPUT0",
      "shape": [
        1,
        16
      ],
      "datatype": "INT32",
      "parameters": {},
      "data": [
        115,
        69,
        97,
        112,
        73,
        106,
        58,
        182,
        114,
        66,
        64,
        110,
        100,
        24,
        22,
        77
      ]
    },
    {
      "name": "OUTPUT1",
      "shape": [
        1,
        16
      ],
      "datatype": "INT32",
      "parameters": {},
      "data": [
        -77,
        33,
        25,
        -52,
        -49,
        -88,
        -48,
        0,
        -50,
        26,
        -44,
        46,
        -2,
        18,
        -6,
        -47
      ]
    }
  ]
}

```

```bash
cat /tmp/tfsimple-pipeline-output.txt | head -n 1 | jq -M .
```

```json
{
  "model_name": "",
  "id": "98413742-caa3-43b9-a5f2-51d11a795399",
  "parameters": {
    "batch_index": 0
  },
  "outputs": [
    {
      "name": "OUTPUT0",
      "shape": [
        1,
        16
      ],
      "datatype": "INT32",
      "data": [
        114,
        46,
        23,
        102,
        45,
        185,
        197,
        106,
        110,
        84,
        74,
        39,
        67,
        140,
        99,
        72
      ]
    },
    {
      "name": "OUTPUT1",
      "shape": [
        1,
        16
      ],
      "datatype": "INT32",
      "data": [
        36,
        32,
        -5,
        -14,
        19,
        9,
        1,
        -26,
        -84,
        -30,
        -24,
        33,
        -31,
        14,
        25,
        48
      ]
    }
  ]
}

```

## Cleanup

```bash
kubectl delete -f models/sklearn-iris-gs.yaml -n ${NAMESPACE}
kubectl delete -f pipelines/iris.yaml -n ${NAMESPACE}
```

```
model.mlops.seldon.io "iris" deleted
pipeline.mlops.seldon.io "iris-pipeline" deleted

```

```bash
kubectl delete -f models/tfsimple1.yaml -n ${NAMESPACE}
kubectl delete -f pipelines/tfsimple.yaml -n ${NAMESPACE}
```

```
model.mlops.seldon.io "tfsimple1" deleted
pipeline.mlops.seldon.io "tfsimple" deleted

```

```python

```
