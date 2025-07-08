# Tensorflow Load and Benchmark Tests

Using a pretrained model for [Tensorflow flowers dataset](https://www.tensorflow.org/datasets/catalog/tf_flowers)

 * Load test the model at fixed rate
 * Benchmark the model to find maximum throughput and saturation handling
 
 ## Setup
 
  * Create a 3 node GCP cluster with n1-standard-8 node
  * Install Seldon Core
  
 ## TODO
 
  * gRPC
  * Run vegeta on separate node to model servers using affinity/taints


```python
!kubectl create namespace seldon
```

    Error from server (AlreadyExists): namespaces "seldon" already exists



```python
!kubectl config set-context $(kubectl config current-context) --namespace=seldon
```

    Context "do-lon1-k8s-1-16-10-do-0-lon1-1594477430912" modified.



```python
import sys

sys.path.append("../")
from vegeta_utils import *
```

## Put Taint on Nodes


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

    node/pool-triv8uq93-3oaz0 tainted
    error: Node pool-triv8uq93-3oaz1 already has model taint(s) with same effect(s) and --overwrite is false
    error: Node pool-triv8uq93-3oazd already has model taint(s) with same effect(s) and --overwrite is false


## Benchmark with Saturation Test


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


Run test to gather the max throughput of the model


```python
results = run_vegeta_test("tf_vegeta_cfg.yaml", "vegeta_max.yaml", "11m")
print(json.dumps(results, indent=4))
saturation_throughput = int(results["throughput"])
```

    {
        "latencies": {
            "total": 18194676761380,
            "mean": 4069487085,
            "50th": 3865217401,
            "90th": 5285272466,
            "95th": 5768188708,
            "99th": 6667031940,
            "max": 7656080367,
            "min": 970003451
        },
        "bytes_in": {
            "total": 974678,
            "mean": 218
        },
        "bytes_out": {
            "total": 72318425,
            "mean": 16175
        },
        "earliest": "2020-07-13T09:38:48.517793327Z",
        "latest": "2020-07-13T09:41:48.535299333Z",
        "end": "2020-07-13T09:41:52.165570518Z",
        "duration": 180017506006,
        "wait": 3630271185,
        "requests": 4471,
        "rate": 24.836473403042152,
        "throughput": 24.34551655558568,
        "success": 1,
        "status_codes": {
            "200": 4471
        },
        "errors": []
    }



```python
print("Max Throughput=", saturation_throughput)
```

    Max Throughput= 24


## Load Tests with HPA

Run with an HPA at saturation rate to check:
  * Latencies affected by scaling



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
    - hpaSpec:
        minReplicas: 1
        maxReplicas: 5
        metrics:
        - resource:
            name: cpu
            targetAverageUtilization: 70
          type: Resource
      spec:
        containers:
        - name: flowers
          resources:
            requests:
              cpu: '1'
          livenessProbe:
            failureThreshold: 3
            initialDelaySeconds: 60
            periodSeconds: 5
            successThreshold: 1
            tcpSocket:
              port: http
            timeoutSeconds: 5
          readinessProbe:
            failureThreshold: 3
            initialDelaySeconds: 20
            periodSeconds: 5
            successThreshold: 1
            tcpSocket:
              port: http
            timeoutSeconds: 5
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
rate = saturation_throughput
duration = "10m"
%env DURATION=$duration
%env RATE=$rate/1s
!cat vegeta_cfg.tmpl.yaml | envsubst > vegeta.tmp.yaml
!cat vegeta.tmp.yaml
```

    env: DURATION=10m
    env: RATE=24/1s
    apiVersion: batch/v1
    kind: Job
    metadata:
      name: tf-load-test
    spec:
      backoffLimit: 6
      parallelism: 1
      template:
        metadata:
          annotations:
            sidecar.istio.io/inject: "false"
        spec:
          containers:
            - args:
                - vegeta -cpus=4 attack -keepalive=false -duration=10m -rate=24/1s -targets=/var/vegeta/cfg
                  | vegeta report -type=json
              command:
                - sh
                - -c
              image: peterevans/vegeta:latest
              imagePullPolicy: Always
              name: vegeta
              volumeMounts:
                - mountPath: /var/vegeta
                  name: tf-vegeta-cfg
          restartPolicy: Never
          volumes:
            - configMap:
                defaultMode: 420
                name: tf-vegeta-cfg
              name: tf-vegeta-cfg
          tolerations:
          - key: loadtester
            operator: Exists
            effect: NoSchedule



```python
results = run_vegeta_test("tf_vegeta_cfg.yaml", "vegeta.tmp.yaml", "11m")
print(json.dumps(results, indent=4))
```

    {
        "latencies": {
            "total": 3743859444532,
            "mean": 259990239,
            "50th": 131917169,
            "90th": 310053255,
            "95th": 916684759,
            "99th": 2775052710,
            "max": 7645706522,
            "min": 61953433
        },
        "bytes_in": {
            "total": 3139200,
            "mean": 218
        },
        "bytes_out": {
            "total": 232920000,
            "mean": 16175
        },
        "earliest": "2020-07-13T09:57:01.982849851Z",
        "latest": "2020-07-13T10:07:01.94120089Z",
        "end": "2020-07-13T10:07:02.043547541Z",
        "duration": 599958351039,
        "wait": 102346651,
        "requests": 14400,
        "rate": 24.001666074090423,
        "throughput": 23.997572337989126,
        "success": 1,
        "status_codes": {
            "200": 14400
        },
        "errors": []
    }



```python
print_vegeta_results(results)
```

    Latencies:
    	mean: 259.990239 ms
    	50th: 131.917169 ms
    	90th: 310.053255 ms
    	95th: 916.684759 ms
    	99th: 2775.05271 ms
    
    Throughput: 23.997572337989126/s
    Errors: False



```python

```
