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
  storageUri: "gs://seldon-models/scv2/samples/mlserver_1.2.4/iris-sklearn"
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
  "id": "0fb4dde4-fced-4da5-b9fc-f939dca60fbd",
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
2023-03-10 10:53:40,831 [mlserver] INFO - Using asyncio event-loop policy: uvloop
2023-03-10 10:53:40,836 [mlserver] INFO - server url: 172.19.255.1
2023-03-10 10:53:40,836 [mlserver] INFO - model name: iris
2023-03-10 10:53:40,836 [mlserver] INFO - request headers: {}
2023-03-10 10:53:40,836 [mlserver] INFO - input file path: batch-inputs/iris-input.txt
2023-03-10 10:53:40,836 [mlserver] INFO - output file path: /tmp/iris-output.txt
2023-03-10 10:53:40,836 [mlserver] INFO - workers: 5
2023-03-10 10:53:40,836 [mlserver] INFO - retries: 3
2023-03-10 10:53:40,836 [mlserver] INFO - batch interval: 0.0
2023-03-10 10:53:40,836 [mlserver] INFO - batch jitter: 0.0
2023-03-10 10:53:40,836 [mlserver] INFO - connection timeout: 60
2023-03-10 10:53:40,836 [mlserver] INFO - micro-batch size: 1
2023-03-10 10:53:40,993 [mlserver] INFO - Finalizer: processed instances: 100
2023-03-10 10:53:40,994 [mlserver] INFO - Total processed instances: 100
2023-03-10 10:53:40,994 [mlserver] INFO - Time taken: 0.16 seconds

```

```bash
%%bash
mlserver infer -u ${MESH_IP} -m iris-pipeline.pipeline -i batch-inputs/iris-input.txt -o /tmp/iris-pipeline-output.txt --workers 5

```

```
2023-03-10 10:53:50,712 [mlserver] INFO - Using asyncio event-loop policy: uvloop
2023-03-10 10:53:50,713 [mlserver] INFO - server url: 172.19.255.1
2023-03-10 10:53:50,714 [mlserver] INFO - model name: iris-pipeline.pipeline
2023-03-10 10:53:50,714 [mlserver] INFO - request headers: {}
2023-03-10 10:53:50,714 [mlserver] INFO - input file path: batch-inputs/iris-input.txt
2023-03-10 10:53:50,714 [mlserver] INFO - output file path: /tmp/iris-pipeline-output.txt
2023-03-10 10:53:50,714 [mlserver] INFO - workers: 5
2023-03-10 10:53:50,714 [mlserver] INFO - retries: 3
2023-03-10 10:53:50,714 [mlserver] INFO - batch interval: 0.0
2023-03-10 10:53:50,714 [mlserver] INFO - batch jitter: 0.0
2023-03-10 10:53:50,714 [mlserver] INFO - connection timeout: 60
2023-03-10 10:53:50,714 [mlserver] INFO - micro-batch size: 1
2023-03-10 10:53:51,195 [mlserver] INFO - Finalizer: processed instances: 100
2023-03-10 10:53:51,195 [mlserver] INFO - Total processed instances: 100
2023-03-10 10:53:51,195 [mlserver] INFO - Time taken: 0.48 seconds

```

```bash
cat /tmp/iris-output.txt | head -n 1 | jq -M .
```

```json
{
  "model_name": "iris_1",
  "model_version": "1",
  "id": "d7707cc7-37be-426e-812c-3f79969ad82c",
  "parameters": {
    "batch_index": 0,
    "inference_id": "d7707cc7-37be-426e-812c-3f79969ad82c"
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
        1
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
  "id": "81e2294f-b8ae-4f6c-bcdb-144aae7cac5a",
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
        1
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
2023-03-10 10:54:05,427 [mlserver] INFO - Using asyncio event-loop policy: uvloop
2023-03-10 10:54:05,429 [mlserver] INFO - server url: 172.19.255.1
2023-03-10 10:54:05,429 [mlserver] INFO - model name: tfsimple1
2023-03-10 10:54:05,429 [mlserver] INFO - request headers: {}
2023-03-10 10:54:05,429 [mlserver] INFO - input file path: batch-inputs/tfsimple-input.txt
2023-03-10 10:54:05,429 [mlserver] INFO - output file path: /tmp/tfsimple-output.txt
2023-03-10 10:54:05,429 [mlserver] INFO - workers: 5
2023-03-10 10:54:05,429 [mlserver] INFO - retries: 3
2023-03-10 10:54:05,429 [mlserver] INFO - batch interval: 0.0
2023-03-10 10:54:05,429 [mlserver] INFO - batch jitter: 0.0
2023-03-10 10:54:05,429 [mlserver] INFO - connection timeout: 60
2023-03-10 10:54:05,429 [mlserver] INFO - micro-batch size: 1
2023-03-10 10:54:05,527 [mlserver] INFO - Finalizer: processed instances: 100
2023-03-10 10:54:05,527 [mlserver] INFO - Total processed instances: 100
2023-03-10 10:54:05,528 [mlserver] INFO - Time taken: 0.10 seconds

```

```bash
%%bash
mlserver infer -u ${MESH_IP} -m tfsimple.pipeline -i batch-inputs/tfsimple-input.txt -o /tmp/tfsimple-pipeline-output.txt --workers 5

```

```
2023-03-10 10:54:07,559 [mlserver] INFO - Using asyncio event-loop policy: uvloop
2023-03-10 10:54:07,560 [mlserver] INFO - server url: 172.19.255.1
2023-03-10 10:54:07,560 [mlserver] INFO - model name: tfsimple.pipeline
2023-03-10 10:54:07,560 [mlserver] INFO - request headers: {}
2023-03-10 10:54:07,560 [mlserver] INFO - input file path: batch-inputs/tfsimple-input.txt
2023-03-10 10:54:07,560 [mlserver] INFO - output file path: /tmp/tfsimple-pipeline-output.txt
2023-03-10 10:54:07,560 [mlserver] INFO - workers: 5
2023-03-10 10:54:07,560 [mlserver] INFO - retries: 3
2023-03-10 10:54:07,560 [mlserver] INFO - batch interval: 0.0
2023-03-10 10:54:07,560 [mlserver] INFO - batch jitter: 0.0
2023-03-10 10:54:07,560 [mlserver] INFO - connection timeout: 60
2023-03-10 10:54:07,560 [mlserver] INFO - micro-batch size: 1
2023-03-10 10:54:07,857 [mlserver] INFO - Finalizer: processed instances: 100
2023-03-10 10:54:07,858 [mlserver] INFO - Total processed instances: 100
2023-03-10 10:54:07,858 [mlserver] INFO - Time taken: 0.30 seconds

```

```bash
cat /tmp/tfsimple-output.txt | head -n 1 | jq -M .
```

```json
{
  "model_name": "tfsimple1_1",
  "model_version": "1",
  "id": "cbce88c5-508d-4c34-bb04-228c4ca4cfa3",
  "parameters": {
    "batch_index": 1,
    "inference_id": "cbce88c5-508d-4c34-bb04-228c4ca4cfa3"
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
  "id": "734b54c5-2253-4311-b097-5f63d821c29d",
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
