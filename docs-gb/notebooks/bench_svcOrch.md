# Service Orchestrator Benchmark Tests

Using a pretrained model for [Tensorflow flowers dataset](https://www.tensorflow.org/datasets/catalog/tf_flowers)

 * Tests the extra latency added by the svcOrch for a medium size image (224x224) classification model.
 
 ## Setup
 
  * Create a 3 node cluster
  * Install Seldon Core


```python
!kubectl create namespace seldon
```


```python
!kubectl config set-context $(kubectl config current-context) --namespace=seldon
```


```python
import sys

sys.path.append("../")
from vegeta_utils import *
```

## Put Taints Nodes


```python
raw = !kubectl get nodes -o jsonpath='{.items[0].metadata.name}'
firstNode = raw[0]
raw = !kubectl get nodes -o jsonpath='{.items[1].metadata.name}'
secondNode = raw[0]
raw = !kubectl get nodes -o jsonpath='{.items[2].metadata.name}'
thirdNode = raw[0]
!kubectl taint nodes '{firstNode}' loadtester=active:NoSchedule
!kubectl taint nodes '{secondNode}' model=active:NoSchedule
!kubectl taint nodes '{thirdNode}' model=active:NoSchedule
```

    error: Node pool-triv8uq93-3oaz0 already has loadtester taint(s) with same effect(s) and --overwrite is false
    error: Node pool-triv8uq93-3oaz1 already has model taint(s) with same effect(s) and --overwrite is false
    error: Node pool-triv8uq93-3oazd already has model taint(s) with same effect(s) and --overwrite is false


## Tensorflow Flowers Model - Latency Test


```python
%%writefile tf_flowers.yaml
apiVersion: machinelearning.seldon.io/v1alpha2
kind: SeldonDeployment
metadata:
  name: tf-flowers
spec:
  protocol: tensorflow
  transport: rest
  predictors:
  - graph:
      implementation: TENSORFLOW_SERVER
      modelUri: gs://kfserving-samples/models/tensorflow/flowers
      name:  flowers
      parameters:
        - name: model_name
          type: STRING
          value: flowers
    componentSpecs:
    - spec:
        containers:
        - name: flowers
          resources:
            requests:
              cpu: '2'
        tolerations:
        - key: model
          operator: Exists
          effect: NoSchedule
    name: default
    replicas: 1
```

    Overwriting tf_flowers.yaml



```python
run_model("tf_flowers.yaml")
```

    Available with 1 pods



```python
results = run_vegeta_test("tf_vegeta_cfg.yaml", "vegeta_1worker.yaml", "60m")
print(json.dumps(results, indent=4))
mean_with_executor = results["latencies"]["mean"]
```

    {
        "latencies": {
            "total": 1200086051040,
            "mean": 82639171,
            "50th": 79832732,
            "90th": 95849466,
            "95th": 104009039,
            "99th": 128516774,
            "max": 964378237,
            "min": 58091922
        },
        "bytes_in": {
            "total": 3165796,
            "mean": 218
        },
        "bytes_out": {
            "total": 234893350,
            "mean": 16175
        },
        "earliest": "2020-07-12T15:59:55.298435559Z",
        "latest": "2020-07-12T16:19:55.34937906Z",
        "end": "2020-07-12T16:19:55.42413935Z",
        "duration": 1200050943501,
        "wait": 74760290,
        "requests": 14522,
        "rate": 12.10115293741936,
        "throughput": 12.100399111632546,
        "success": 1,
        "status_codes": {
            "200": 14522
        },
        "errors": []
    }


## Tensorflow Flowers Model - No executor - Latency Test



```python
%%writefile tf_flowers.yaml
apiVersion: machinelearning.seldon.io/v1alpha2
kind: SeldonDeployment
metadata:
  name: tf-flowers
spec:
  protocol: tensorflow
  transport: rest
  predictors:
  - graph:
      implementation: TENSORFLOW_SERVER
      modelUri: gs://kfserving-samples/models/tensorflow/flowers
      name:  flowers
      parameters:
        - name: model_name
          type: STRING
          value: flowers
    annotations:
        seldon.io/no-engine: "true"
    componentSpecs:
    - spec:
        containers:
        - name: flowers
          resources:
            requests:
              cpu: '2'
        tolerations:
        - key: model
          operator: Exists
          effect: NoSchedule
    name: default
    replicas: 1
```

    Overwriting tf_flowers.yaml



```python
run_model("tf_flowers.yaml")
```

    Available with 1 pods



```python
results = run_vegeta_test("tf_standalone_vegeta_cfg.yaml", "vegeta_1worker.yaml", "60m")
print(json.dumps(results, indent=4))
mean_no_executor = results["latencies"]["mean"]
```

    {
        "latencies": {
            "total": 1200089018347,
            "mean": 73670289,
            "50th": 73129037,
            "90th": 81823849,
            "95th": 84928884,
            "99th": 93248220,
            "max": 976431685,
            "min": 53958421
        },
        "bytes_in": {
            "total": 3551220,
            "mean": 218
        },
        "bytes_out": {
            "total": 263490750,
            "mean": 16175
        },
        "earliest": "2020-07-12T16:21:00.12358772Z",
        "latest": "2020-07-12T16:41:00.180620249Z",
        "end": "2020-07-12T16:41:00.255483814Z",
        "duration": 1200057032529,
        "wait": 74863565,
        "requests": 16290,
        "rate": 13.574354850177793,
        "throughput": 13.573508089417606,
        "success": 1,
        "status_codes": {
            "200": 16290
        },
        "errors": []
    }



```python
diff = (mean_with_executor - mean_no_executor) / 1e6
print("Diff in ms", diff)
```

    Diff in ms 8.968882


## GRPC Tensorflow Flowers Model - Latency Test

First create the binary proto for the flowers payload


```python
!python ../tf_proto_save.py --model flowers --input_path flowers.json --output_path flowers.bin
```

    /home/clive/anaconda3/envs/seldon-core/lib/python3.6/site-packages/tensorflow/python/framework/dtypes.py:516: FutureWarning: Passing (type, 1) or '1type' as a synonym of type is deprecated; in a future version of numpy, it will be understood as (type, (1,)) / '(1,)type'.
      _np_qint8 = np.dtype([("qint8", np.int8, 1)])
    /home/clive/anaconda3/envs/seldon-core/lib/python3.6/site-packages/tensorflow/python/framework/dtypes.py:517: FutureWarning: Passing (type, 1) or '1type' as a synonym of type is deprecated; in a future version of numpy, it will be understood as (type, (1,)) / '(1,)type'.
      _np_quint8 = np.dtype([("quint8", np.uint8, 1)])
    /home/clive/anaconda3/envs/seldon-core/lib/python3.6/site-packages/tensorflow/python/framework/dtypes.py:518: FutureWarning: Passing (type, 1) or '1type' as a synonym of type is deprecated; in a future version of numpy, it will be understood as (type, (1,)) / '(1,)type'.
      _np_qint16 = np.dtype([("qint16", np.int16, 1)])
    /home/clive/anaconda3/envs/seldon-core/lib/python3.6/site-packages/tensorflow/python/framework/dtypes.py:519: FutureWarning: Passing (type, 1) or '1type' as a synonym of type is deprecated; in a future version of numpy, it will be understood as (type, (1,)) / '(1,)type'.
      _np_quint16 = np.dtype([("quint16", np.uint16, 1)])
    /home/clive/anaconda3/envs/seldon-core/lib/python3.6/site-packages/tensorflow/python/framework/dtypes.py:520: FutureWarning: Passing (type, 1) or '1type' as a synonym of type is deprecated; in a future version of numpy, it will be understood as (type, (1,)) / '(1,)type'.
      _np_qint32 = np.dtype([("qint32", np.int32, 1)])
    /home/clive/anaconda3/envs/seldon-core/lib/python3.6/site-packages/tensorflow/python/framework/dtypes.py:525: FutureWarning: Passing (type, 1) or '1type' as a synonym of type is deprecated; in a future version of numpy, it will be understood as (type, (1,)) / '(1,)type'.
      np_resource = np.dtype([("resource", np.ubyte, 1)])
    /home/clive/anaconda3/envs/seldon-core/lib/python3.6/site-packages/tensorboard/compat/tensorflow_stub/dtypes.py:541: FutureWarning: Passing (type, 1) or '1type' as a synonym of type is deprecated; in a future version of numpy, it will be understood as (type, (1,)) / '(1,)type'.
      _np_qint8 = np.dtype([("qint8", np.int8, 1)])
    /home/clive/anaconda3/envs/seldon-core/lib/python3.6/site-packages/tensorboard/compat/tensorflow_stub/dtypes.py:542: FutureWarning: Passing (type, 1) or '1type' as a synonym of type is deprecated; in a future version of numpy, it will be understood as (type, (1,)) / '(1,)type'.
      _np_quint8 = np.dtype([("quint8", np.uint8, 1)])
    /home/clive/anaconda3/envs/seldon-core/lib/python3.6/site-packages/tensorboard/compat/tensorflow_stub/dtypes.py:543: FutureWarning: Passing (type, 1) or '1type' as a synonym of type is deprecated; in a future version of numpy, it will be understood as (type, (1,)) / '(1,)type'.
      _np_qint16 = np.dtype([("qint16", np.int16, 1)])
    /home/clive/anaconda3/envs/seldon-core/lib/python3.6/site-packages/tensorboard/compat/tensorflow_stub/dtypes.py:544: FutureWarning: Passing (type, 1) or '1type' as a synonym of type is deprecated; in a future version of numpy, it will be understood as (type, (1,)) / '(1,)type'.
      _np_quint16 = np.dtype([("quint16", np.uint16, 1)])
    /home/clive/anaconda3/envs/seldon-core/lib/python3.6/site-packages/tensorboard/compat/tensorflow_stub/dtypes.py:545: FutureWarning: Passing (type, 1) or '1type' as a synonym of type is deprecated; in a future version of numpy, it will be understood as (type, (1,)) / '(1,)type'.
      _np_qint32 = np.dtype([("qint32", np.int32, 1)])
    /home/clive/anaconda3/envs/seldon-core/lib/python3.6/site-packages/tensorboard/compat/tensorflow_stub/dtypes.py:550: FutureWarning: Passing (type, 1) or '1type' as a synonym of type is deprecated; in a future version of numpy, it will be understood as (type, (1,)) / '(1,)type'.
      np_resource = np.dtype([("resource", np.ubyte, 1)])



```python
%%writefile tf_flowers.yaml
apiVersion: machinelearning.seldon.io/v1alpha2
kind: SeldonDeployment
metadata:
  name: tf-flowers
spec:
  protocol: tensorflow
  transport: grpc
  predictors:
  - graph:
      implementation: TENSORFLOW_SERVER
      modelUri: gs://kfserving-samples/models/tensorflow/flowers
      name:  flowers
      parameters:
        - name: model_name
          type: STRING
          value: flowers
    componentSpecs:
    - spec:
        containers:
        - name: flowers
          resources:
            requests:
              cpu: '2'
        tolerations:
        - key: model
          operator: Exists
          effect: NoSchedule
    name: default
    replicas: 1
```

    Overwriting tf_flowers.yaml



```python
run_model("tf_flowers.yaml")
```

    Available with 1 pods



```python
results = run_ghz_test("flowers.bin", "ghz_1worker.yaml", "60m")
print(json.dumps(results, indent=4))
mean_with_executor = results["average"]
```

    {
        "date": "2020-07-12T17:12:04Z",
        "endReason": "timeout",
        "options": {
            "host": "tf-flowers-default.seldon.svc.cluster.local:8000",
            "proto": "/proto/prediction_service.proto",
            "import-paths": [
                "/proto",
                "."
            ],
            "call": "tensorflow.serving.PredictionService/Predict",
            "insecure": true,
            "total": 1000000,
            "concurrency": 1,
            "connections": 1,
            "duration": 1800000000000,
            "timeout": 20000000000,
            "dial-timeout": 10000000000,
            "keepalive": 1800000000000,
            "binary": true,
            "CPUs": 8
        },
        "count": 22978,
        "total": 1800000675146,
        "average": 78227435,
        "fastest": 54712167,
        "slowest": 938906233,
        "rps": 12.76555076743859,
        "errorDistribution": {
            "rpc error: code = Unavailable desc = transport is closing": 1
        },
        "statusCodeDistribution": {
            "OK": 22977,
            "Unavailable": 1
        },
        "latencyDistribution": [
            {
                "percentage": 10,
                "latency": 68291719
            },
            {
                "percentage": 25,
                "latency": 71762262
            },
            {
                "percentage": 50,
                "latency": 75875238
            },
            {
                "percentage": 75,
                "latency": 81163515
            },
            {
                "percentage": 90,
                "latency": 89225781
            },
            {
                "percentage": 95,
                "latency": 97536730
            },
            {
                "percentage": 99,
                "latency": 128647238
            }
        ]
    }


## GRPC Tensorflow Flowers Model - No executor - Latency Test



```python
%%writefile tf_flowers.yaml
apiVersion: machinelearning.seldon.io/v1alpha2
kind: SeldonDeployment
metadata:
  name: tf-flowers
spec:
  protocol: tensorflow
  transport: grpc
  predictors:
  - graph:
      implementation: TENSORFLOW_SERVER
      modelUri: gs://kfserving-samples/models/tensorflow/flowers
      name:  flowers
      parameters:
        - name: model_name
          type: STRING
          value: flowers
    annotations:
        seldon.io/no-engine: "true"
    componentSpecs:
    - spec:
        containers:
        - name: flowers
          resources:
            requests:
              cpu: '2'
        tolerations:
        - key: model
          operator: Exists
          effect: NoSchedule
    name: default
    replicas: 1
```

    Overwriting tf_flowers.yaml



```python
run_model("tf_flowers.yaml")
```

    Available with 1 pods



```python
results = run_ghz_test("flowers.bin", "ghz_standalone_1worker.yaml", "60m")
print(json.dumps(results, indent=4))
mean_no_executor = results["average"]
```

    {
        "date": "2020-07-12T18:04:44Z",
        "endReason": "timeout",
        "options": {
            "host": "tf-flowers-default.seldon.svc.cluster.local:9000",
            "proto": "/proto/prediction_service.proto",
            "import-paths": [
                "/proto",
                "."
            ],
            "call": "tensorflow.serving.PredictionService/Predict",
            "insecure": true,
            "total": 1000000,
            "concurrency": 1,
            "connections": 1,
            "duration": 1800000000000,
            "timeout": 20000000000,
            "dial-timeout": 10000000000,
            "keepalive": 1800000000000,
            "binary": true,
            "CPUs": 8
        },
        "count": 24132,
        "total": 1800013456837,
        "average": 74479232,
        "fastest": 53792435,
        "slowest": 1008191507,
        "rps": 13.406566438900391,
        "errorDistribution": {
            "rpc error: code = Unavailable desc = transport is closing": 1
        },
        "statusCodeDistribution": {
            "OK": 24131,
            "Unavailable": 1
        },
        "latencyDistribution": [
            {
                "percentage": 10,
                "latency": 67087978
            },
            {
                "percentage": 25,
                "latency": 70242403
            },
            {
                "percentage": 50,
                "latency": 73838624
            },
            {
                "percentage": 75,
                "latency": 77894265
            },
            {
                "percentage": 90,
                "latency": 82282422
            },
            {
                "percentage": 95,
                "latency": 85533746
            },
            {
                "percentage": 99,
                "latency": 93875540
            }
        ]
    }



```python
diff = (mean_with_executor - mean_no_executor) / 1e6
print("Diff in ms", diff)
```

    Diff in ms 3.748203



```python

```
