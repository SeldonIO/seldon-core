### Seldon V2 Batch Examples


```python
import os
os.environ["NAMESPACE"] = "seldon"
```


```python
MESH_IP=!kubectl get svc seldon-mesh -n ${NAMESPACE} -o jsonpath='{.status.loadBalancer.ingress[0].ip}'
MESH_IP=MESH_IP[0]
import os
os.environ['MESH_IP'] = MESH_IP
MESH_IP
```




    '172.19.255.12'



## Deploy Models and Pipelines


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
!kubectl apply -f models/sklearn-iris-gs.yaml -n ${NAMESPACE}
!kubectl apply -f pipelines/iris.yaml -n ${NAMESPACE}

!kubectl apply -f models/tfsimple1.yaml -n ${NAMESPACE}
!kubectl apply -f pipelines/tfsimple.yaml -n ${NAMESPACE}
```

    model.mlops.seldon.io/iris created
    pipeline.mlops.seldon.io/iris-pipeline created
    model.mlops.seldon.io/tfsimple1 created
    pipeline.mlops.seldon.io/tfsimple created



```python
!kubectl wait --for condition=ready --timeout=300s model --all -n ${NAMESPACE}
!kubectl wait --for condition=ready --timeout=300s pipelines --all -n ${NAMESPACE}

!kubectl wait --for condition=ready --timeout=300s model tfsimple1 -n ${NAMESPACE}
!kubectl wait --for condition=ready --timeout=300s pipelines tfsimple -n ${NAMESPACE}
```

    model.mlops.seldon.io/iris condition met
    model.mlops.seldon.io/tfsimple condition met
    model.mlops.seldon.io/tfsimple1 condition met
    pipeline.mlops.seldon.io/iris-pipeline condition met
    pipeline.mlops.seldon.io/tfsimple condition met
    pipeline.mlops.seldon.io/tfsimple-pipeline condition met
    model.mlops.seldon.io/tfsimple1 condition met
    pipeline.mlops.seldon.io/tfsimple condition met


## Test Predictions


```python
!seldon model infer iris --inference-host ${MESH_IP}:80 \
  '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}' | jq -M .
```

    {
      "model_name": "iris_1",
      "model_version": "1",
      "id": "3e60ef91-4422-4989-85c6-05d54a0c7473",
      "parameters": {
        "content_type": null,
        "headers": null
      },
      "outputs": [
        {
          "name": "predict",
          "shape": [
            1
          ],
          "datatype": "INT64",
          "parameters": null,
          "data": [
            2
          ]
        }
      ]
    }



```python
!seldon pipeline infer iris-pipeline --inference-host ${MESH_IP}:80 \
  '{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}' |  jq -M .
```

    {
      "model_name": "",
      "outputs": [
        {
          "data": [
            2
          ],
          "name": "predict",
          "shape": [
            1
          ],
          "datatype": "INT64"
        }
      ]
    }



```python
!seldon model infer tfsimple --inference-host ${MESH_IP}:80 \
  '{"outputs":[{"name":"OUTPUT0"}], "inputs":[{"name":"INPUT0","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,16]},{"name":"INPUT1","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,16]}]}' | jq -M .
```

    {
      "model_name": "tfsimple_1",
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



```python
!seldon pipeline infer tfsimple-pipeline --inference-host ${MESH_IP}:80 \
  '{"outputs":[{"name":"OUTPUT0"}], "inputs":[{"name":"INPUT0","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,16]},{"name":"INPUT1","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,16]}]}' | jq -M .
```

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


## MLServer Iris Batch Job


```python
!cat batch-inputs/iris-input.txt | head -n 1 | jq -M .
```

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



```bash
%%bash
mlserver infer -u ${MESH_IP} -m iris -i batch-inputs/iris-input.txt -o /tmp/iris-output.txt --workers 5 
```

    2022-10-10 10:12:51,214 [mlserver] INFO - Using asyncio event-loop policy: uvloop
    2022-10-10 10:12:51,217 [mlserver] INFO - Server url: 172.19.255.12
    2022-10-10 10:12:51,217 [mlserver] INFO - input file path: batch-inputs/iris-input.txt
    2022-10-10 10:12:51,217 [mlserver] INFO - output file path: /tmp/iris-output.txt
    2022-10-10 10:12:51,408 [mlserver] INFO - Time taken: 0.19 seconds



```bash
%%bash
mlserver infer -u ${MESH_IP} -m iris-pipeline.pipeline -i batch-inputs/iris-input.txt -o /tmp/iris-pipeline-output.txt --workers 5
```

    2022-10-10 10:12:53,067 [mlserver] INFO - Using asyncio event-loop policy: uvloop
    2022-10-10 10:12:53,071 [mlserver] INFO - Server url: 172.19.255.12
    2022-10-10 10:12:53,071 [mlserver] INFO - input file path: batch-inputs/iris-input.txt
    2022-10-10 10:12:53,071 [mlserver] INFO - output file path: /tmp/iris-pipeline-output.txt
    2022-10-10 10:12:53,374 [mlserver] INFO - Time taken: 0.30 seconds



```python
!cat /tmp/iris-output.txt | head -n 1 | jq -M .
```

    {
      "model_name": "iris_1",
      "model_version": "1",
      "id": "75470ffc-e56f-4e41-9b3f-c0b5c236833f",
      "parameters": {
        "content_type": null,
        "headers": null,
        "batch_index": 0
      },
      "outputs": [
        {
          "name": "predict",
          "shape": [
            1
          ],
          "datatype": "INT64",
          "parameters": null,
          "data": [
            0
          ]
        }
      ]
    }



```python
!cat /tmp/iris-pipeline-output.txt | head -n 1 | jq .
```

    [1;39m{
      [0m[34;1m"model_name"[0m[1;39m: [0m[0;32m""[0m[1;39m,
      [0m[34;1m"outputs"[0m[1;39m: [0m[1;39m[
        [1;39m{
          [0m[34;1m"data"[0m[1;39m: [0m[1;39m[
            [0;39m0[0m[1;39m
          [1;39m][0m[1;39m,
          [0m[34;1m"name"[0m[1;39m: [0m[0;32m"predict"[0m[1;39m,
          [0m[34;1m"shape"[0m[1;39m: [0m[1;39m[
            [0;39m1[0m[1;39m
          [1;39m][0m[1;39m,
          [0m[34;1m"datatype"[0m[1;39m: [0m[0;32m"INT64"[0m[1;39m
        [1;39m}[0m[1;39m
      [1;39m][0m[1;39m,
      [0m[34;1m"parameters"[0m[1;39m: [0m[1;39m{
        [0m[34;1m"batch_index"[0m[1;39m: [0m[0;39m0[0m[1;39m
      [1;39m}[0m[1;39m
    [1;39m}[0m


## Triton TFSimple Batch Job


```python
!cat batch-inputs/tfsimple-input.txt | head -n 1 | jq -M .
```

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



```bash
%%bash
mlserver infer -u ${MESH_IP} -m tfsimple -i batch-inputs/tfsimple-input.txt -o /tmp/tfsimple-output.txt --workers 5 -b
```

    2022-10-10 10:12:55,454 [mlserver] INFO - Using asyncio event-loop policy: uvloop
    2022-10-10 10:12:55,457 [mlserver] INFO - Server url: 172.19.255.12
    2022-10-10 10:12:55,457 [mlserver] INFO - input file path: batch-inputs/tfsimple-input.txt
    2022-10-10 10:12:55,457 [mlserver] INFO - output file path: /tmp/tfsimple-output.txt
    2022-10-10 10:12:55,628 [mlserver] INFO - Time taken: 0.17 seconds



```bash
%%bash
mlserver infer -u ${MESH_IP} -m tfsimple-pipeline.pipeline -i batch-inputs/tfsimple-input.txt -o /tmp/tfsimple-pipeline-output.txt --workers 5
```

    2022-10-10 10:12:57,269 [mlserver] INFO - Using asyncio event-loop policy: uvloop
    2022-10-10 10:12:57,273 [mlserver] INFO - Server url: 172.19.255.12
    2022-10-10 10:12:57,273 [mlserver] INFO - input file path: batch-inputs/tfsimple-input.txt
    2022-10-10 10:12:57,273 [mlserver] INFO - output file path: /tmp/tfsimple-pipeline-output.txt
    2022-10-10 10:12:57,592 [mlserver] INFO - Time taken: 0.32 seconds



```python
!cat /tmp/tfsimple-output.txt | head -n 1 | jq -M .
```

    {
      "id": "ac7e8eb2-3e13-4b8d-9e9d-9633b12964c3",
      "model_name": "tfsimple_1",
      "model_version": "1",
      "outputs": [
        {
          "name": "OUTPUT0",
          "datatype": "INT32",
          "shape": [
            1,
            16
          ],
          "parameters": {},
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
          "datatype": "INT32",
          "shape": [
            1,
            16
          ],
          "parameters": {},
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
      ],
      "parameters": {
        "batch_index": 0
      }
    }



```python
!cat /tmp/tfsimple-pipeline-output.txt | head -n 1 | jq -M .
```

    {
      "model_name": "",
      "outputs": [
        {
          "data": [
            63,
            75,
            152,
            77,
            159,
            124,
            92,
            42,
            93,
            155,
            101,
            151,
            103,
            109,
            139,
            105
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
            13,
            -21,
            14,
            39,
            33,
            38,
            -42,
            -26,
            -33,
            -37,
            -31,
            -47,
            -41,
            75,
            57,
            37
          ],
          "name": "OUTPUT1",
          "shape": [
            1,
            16
          ],
          "datatype": "INT32"
        }
      ],
      "parameters": {
        "batch_index": 4
      }
    }


## Cleanup


```python
!kubectl delete -f models/sklearn-iris-gs.yaml -n ${NAMESPACE}
!kubectl delete -f pipelines/iris.yaml -n ${NAMESPACE}
```

    model.mlops.seldon.io "iris" deleted
    pipeline.mlops.seldon.io "iris-pipeline" deleted



```python
!kubectl delete -f models/tfsimple1.yaml -n ${NAMESPACE}
!kubectl delete -f pipelines/tfsimple.yaml -n ${NAMESPACE}
```

    model.mlops.seldon.io "tfsimple1" deleted
    pipeline.mlops.seldon.io "tfsimple" deleted

