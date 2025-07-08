# Python Wrapper Benchmarking

## Prequisites

 * An authenticated K8S cluster with istio and Seldon Core installed
   * You can use the ansible seldon-core playbook at https://github.com/SeldonIO/ansible-k8s-collection
 * vegeta and ghz benchmarking tools
 
 Port forward to istio
 
 ```
 kubectl port-forward $(kubectl get pods -l istio=ingressgateway -n istio-system -o jsonpath='{.items[0].metadata.name}') -n istio-system 8003:8080
 ```
 
 Tests
 
   * **Large Batch Size**
      * `predict` method with:
         * REST
            * ndarray
            * tensor
            * tftensor
         * gRPC
            * ndarray
            * tensor
            * tftensor
      * `predict_raw` method with:
          * REST
            * ndarray
            * tensor
            * tftensor
          * gRPC
            * ndarray
            * tensor
            * tftensor
   * **Small Batch Size**
       * `predict` method with:
         * REST
            * ndarray
            * tensor
            * tftensor
         * gRPC
            * ndarray
            * tensor
            * tftensor  
            
            
## TLDR

  * gRPC is faster than REST
  * tftensor is best for large batch size
  * ndarray with gRPC is bad for large batch size
  * simpler tensor/ndarray is better for small batch size


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




    '1.10.0-dev'




```python
!kubectl create namespace seldon
```

    Error from server (AlreadyExists): namespaces "seldon" already exists



```python
!helm upgrade --install seldon-core seldon-core-operator --repo https://storage.googleapis.com/seldon-charts --version 1.9.0 --namespace seldon-system --set istio.enabled="true" --set istio.gateway="seldon-gateway.istio-system.svc.cluster.local"
```

    Release "seldon-core" has been upgraded. Happy Helming!
    NAME: seldon-core
    LAST DEPLOYED: Thu Jul  1 14:03:55 2021
    NAMESPACE: seldon-system
    STATUS: deployed
    REVISION: 2
    TEST SUITE: None


## Test with Predict method on Large Batch Size

The `seldontest_predict` has simply a `predict` method that does a loop with a configurable number of iterations (default 1) to simulate work. The iterations can be set as a Seldon parameter but in this case we are looking to benchmark the serialization/deserialization cost so want a minimal amount of work.


```python
%%writetemplate model.yaml
apiVersion: machinelearning.seldon.io/v1
kind: SeldonDeployment
metadata:
  name: seldon-model
  namespace: seldon
spec:
  predictors:
  - annotations:
      seldon.io/no-engine: "true"
    componentSpecs:
    - spec:
        containers:
        - image: seldonio/seldontest_predict:{VERSION}
          imagePullPolicy: IfNotPresent
          name: classifier
          resources:
            requests:
              cpu: 1
            limits:
              cpu: 1
          env:
          - name: GUNICORN_WORKERS
            value: "1"
          - name: GUNICORN_THREADS
            value: "1"
        tolerations:
        - key: model
          operator: Exists
          effect: NoSchedule
    graph:
      children: []
      name: classifier
      type: MODEL
    name: default
    replicas: 1
```


```python
!kubectl apply -f model.yaml
```

    seldondeployment.machinelearning.seldon.io/seldon-model created



```python
!kubectl wait --for condition=ready --timeout=600s pods --all -n seldon
```

    pod/seldon-model-default-0-classifier-5445bd4ccf-c2vdr condition met


Create payloads and associated vegeta configurations for

  1. ndarray
  1. tensor
  1. tftensor
  
  We will create an array of 100,000 consecutive integers.


```python
import json

sz = 100000
vals = list(range(sz))
valStr = f"{vals}"
payload = '{"data": {"ndarray": [' + valStr + "]}}"
with open("data_ndarray.json", "w") as f:
    f.write(payload)
payload_tensor = (
    '{"data":{"tensor":{"shape":[1,' + str(sz) + '],"values":' + valStr + "}}}"
)
with open("data_tensor.json", "w") as f:
    f.write(payload_tensor)
```


```python
import numpy as np
import tensorflow as tf
from google.protobuf import json_format

array = np.array(vals)
tftensor = tf.make_tensor_proto(array)
jStrTensor = json_format.MessageToJson(tftensor)
jTensor = json.loads(jStrTensor)
payload_tftensor = (
    '{"data":{"tftensor":' + json.dumps(jTensor, separators=(",", ":")) + "}}"
)
with open("data_tftensor.json", "w") as f:
    f.write(payload_tftensor)
```


```python
import base64
import json

sample_string_bytes = payload_tensor.encode("ascii")
base64_bytes = base64.b64encode(sample_string_bytes)
base64_string = base64_bytes.decode("ascii")
jqPayload = {
    "method": "POST",
    "url": "http://localhost:8003/seldon/seldon/seldon-model/api/v1.0/predictions",
    "body": base64_string,
    "header": {"Content-Type": ["application/json"]},
}
with open("vegeta_tensor.json", "w") as f:
    f.write(json.dumps(jqPayload, separators=(",", ":")))
    f.write("\n")

sample_string_bytes = payload.encode("ascii")
base64_bytes = base64.b64encode(sample_string_bytes)
base64_string = base64_bytes.decode("ascii")
jqPayload = {
    "method": "POST",
    "url": "http://localhost:8003/seldon/seldon/seldon-model/api/v1.0/predictions",
    "body": base64_string,
    "header": {"Content-Type": ["application/json"]},
}
with open("vegeta_ndarray.json", "w") as f:
    f.write(json.dumps(jqPayload, separators=(",", ":")))
    f.write("\n")


sample_string_bytes = payload_tftensor.encode("ascii")
base64_bytes = base64.b64encode(sample_string_bytes)
base64_string = base64_bytes.decode("ascii")
jqPayload = {
    "method": "POST",
    "url": "http://localhost:8003/seldon/seldon/seldon-model/api/v1.0/predictions",
    "body": base64_string,
    "header": {"Content-Type": ["application/json"]},
}
with open("vegeta_tftensor.json", "w") as f:
    f.write(json.dumps(jqPayload, separators=(",", ":")))
    f.write("\n")
```

Smoke test port-forward to check everything is working


```python
!curl -X POST -H 'Content-Type: application/json' \
   -d '@./data_ndarray.json' \
    http://localhost:8003/seldon/seldon/seldon-model/api/v1.0/predictions
```

    {"data":{"names":[],"ndarray":[1]},"meta":{"requestPath":{"classifier":"seldonio/seldontest_predict:1.10.0-dev"}}}



```python
!curl -X POST -H 'Content-Type: application/json' \
   -d '@./data_tensor.json' \
    http://localhost:8003/seldon/seldon/seldon-model/api/v1.0/predictions
```

    {"data":{"names":[],"tensor":{"shape":[1],"values":[1]}},"meta":{"requestPath":{"classifier":"seldonio/seldontest_predict:1.10.0-dev"}}}



```python
!curl -X POST -H 'Content-Type: application/json' \
   -d '@./data_tftensor.json' \
    http://localhost:8003/seldon/seldon/seldon-model/api/v1.0/predictions
```

    {"data":{"names":[],"tftensor":{"dtype":"DT_INT64","int64Val":["1"],"tensorShape":{"dim":[{"size":"1"}]}}},"meta":{"requestPath":{"classifier":"seldonio/seldontest_predict:1.10.0-dev"}}}


Test REST

 1. ndarray
 1. tensor
 1. tftensor
 
 This can be done locally as the results should be indicative of the relative differences rather than very accurate timings.


```bash
%%bash
vegeta attack -format=json -duration=10s -rate=0 -max-workers=1 -targets=vegeta_ndarray.json | 
  vegeta report -type=text
```

    Requests      [total, rate, throughput]         518, 51.76, 51.66
    Duration      [total, attack, wait]             10.027s, 10.008s, 19.333ms
    Latencies     [min, mean, 50, 90, 95, 99, max]  17.337ms, 19.355ms, 19.136ms, 20.336ms, 21.214ms, 24.886ms, 27.831ms
    Bytes In      [total, mean]                     59570, 115.00
    Bytes Out     [total, mean]                     356857970, 688915.00
    Success       [ratio]                           100.00%
    Status Codes  [code:count]                      200:518  
    Error Set:



```bash
%%bash
vegeta attack -format=json -duration=10s -rate=0 -max-workers=1 -targets=vegeta_tensor.json | 
  vegeta report -type=text
```

    Requests      [total, rate, throughput]         504, 50.35, 50.25
    Duration      [total, attack, wait]             10.03s, 10.01s, 19.353ms
    Latencies     [min, mean, 50, 90, 95, 99, max]  17.885ms, 19.897ms, 19.616ms, 21.1ms, 22.205ms, 25.498ms, 34.99ms
    Bytes In      [total, mean]                     69048, 137.00
    Bytes Out     [total, mean]                     347225760, 688940.00
    Success       [ratio]                           100.00%
    Status Codes  [code:count]                      200:504  
    Error Set:



```bash
%%bash
vegeta attack -format=json -duration=10s -rate=0 -max-workers=1 -targets=vegeta_tftensor.json | 
  vegeta report -type=text
```

    Requests      [total, rate, throughput]         636, 63.55, 63.45
    Duration      [total, attack, wait]             10.023s, 10.008s, 14.782ms
    Latencies     [min, mean, 50, 90, 95, 99, max]  13.646ms, 15.756ms, 15.461ms, 17.41ms, 18.729ms, 20.628ms, 23.465ms
    Bytes In      [total, mean]                     118932, 187.00
    Bytes Out     [total, mean]                     678466356, 1066771.00
    Success       [ratio]                           100.00%
    Status Codes  [code:count]                      200:636  
    Error Set:


Example results

| ndarray | tensor | tftensor |
| ------- | ------ | -------- |
| 19.8ms | 19.7ms | 16.2ms |

 Test gRPC
 
  1. ndarray
  1. tensor
  1. tftensor


```bash
%%bash
ghz \
    --insecure \
    --proto ../../../proto/prediction.proto \
    --call seldon.protos.Seldon/Predict \
    --data-file=./data_ndarray.json \
    --qps=0 \
    --cpus=1 \
    --concurrency=1 \
    --duration="10s" \
    --format summary \
    --metadata='{"seldon": "seldon-model", "namespace": "seldon"}' \
    localhost:8003
```

    
    Summary:
      Count:	24
      Total:	10.13 s
      Slowest:	278.81 ms
      Fastest:	242.25 ms
      Average:	244.06 ms
      Requests/sec:	2.37
    
    Response time histogram:
      242.253 [1]	|∎∎∎∎∎∎∎
      245.909 [2]	|∎∎∎∎∎∎∎∎∎∎∎∎∎
      249.564 [4]	|∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎
      253.219 [6]	|∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎
      256.874 [4]	|∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎
      260.530 [1]	|∎∎∎∎∎∎∎
      264.185 [1]	|∎∎∎∎∎∎∎
      267.840 [2]	|∎∎∎∎∎∎∎∎∎∎∎∎∎
      271.496 [0]	|
      275.151 [1]	|∎∎∎∎∎∎∎
      278.806 [1]	|∎∎∎∎∎∎∎
    
    Latency distribution:
      10 % in 247.44 ms 
      25 % in 249.47 ms 
      50 % in 252.85 ms 
      75 % in 260.70 ms 
      90 % in 272.55 ms 
      95 % in 278.81 ms 
      0 % in 0 ns 
    
    Status code distribution:
      [OK]         23 responses   
      [Canceled]   1 responses    
    
    Error distribution:
      [1]   rpc error: code = Canceled desc = grpc: the client connection is closing   
    



```bash
%%bash
ghz \
    --insecure \
    --proto ../../../proto/prediction.proto \
    --call seldon.protos.Seldon/Predict \
    --data-file=./data_tensor.json \
    --qps=0 \
    --cpus=1 \
    --concurrency=1 \
    --duration="10s" \
    --format summary \
    --metadata='{"seldon": "seldon-model", "namespace": "seldon"}' \
    localhost:8003
```

    
    Summary:
      Count:	92
      Total:	10.10 s
      Slowest:	21.23 ms
      Fastest:	4.91 ms
      Average:	7.58 ms
      Requests/sec:	9.11
    
    Response time histogram:
      4.906 [1]	|∎
      6.539 [55]	|∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎
      8.171 [17]	|∎∎∎∎∎∎∎∎∎∎∎∎
      9.804 [4]	|∎∎∎
      11.436 [0]	|
      13.069 [4]	|∎∎∎
      14.701 [3]	|∎∎
      16.334 [3]	|∎∎
      17.966 [0]	|
      19.599 [2]	|∎
      21.232 [2]	|∎
    
    Latency distribution:
      10 % in 5.51 ms 
      25 % in 5.70 ms 
      50 % in 6.14 ms 
      75 % in 7.09 ms 
      90 % in 14.14 ms 
      95 % in 18.77 ms 
      0 % in 0 ns 
    
    Status code distribution:
      [OK]         91 responses   
      [Canceled]   1 responses    
    
    Error distribution:
      [1]   rpc error: code = Canceled desc = grpc: the client connection is closing   
    



```bash
%%bash
ghz \
    --insecure \
    --proto ../../../proto/prediction.proto \
    --call seldon.protos.Seldon/Predict \
    --data-file=./data_tftensor.json \
    --qps=0 \
    --cpus=1 \
    --concurrency=1 \
    --duration="10s" \
    --format summary \
    --metadata='{"seldon": "seldon-model", "namespace": "seldon"}' \
    localhost:8003
```

    
    Summary:
      Count:	425
      Total:	10.04 s
      Slowest:	16.38 ms
      Fastest:	3.97 ms
      Average:	5.33 ms
      Requests/sec:	42.31
    
    Response time histogram:
      3.970 [1]	|
      5.211 [281]	|∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎
      6.452 [91]	|∎∎∎∎∎∎∎∎∎∎∎∎∎
      7.692 [25]	|∎∎∎∎
      8.933 [8]	|∎
      10.174 [6]	|∎
      11.415 [7]	|∎
      12.656 [2]	|
      13.896 [1]	|
      15.137 [1]	|
      16.378 [1]	|
    
    Latency distribution:
      10 % in 4.34 ms 
      25 % in 4.54 ms 
      50 % in 4.89 ms 
      75 % in 5.52 ms 
      90 % in 6.79 ms 
      95 % in 8.30 ms 
      99 % in 11.71 ms 
    
    Status code distribution:
      [OK]         424 responses   
      [Canceled]   1 responses     
    
    Error distribution:
      [1]   rpc error: code = Canceled desc = grpc: the client connection is closing   
    


Example results

| ndarray | tensor | tftensor |
| ------- | ------ | -------- |
| 253ms | 8.4ms | 5.5ms |

## Conclusions

 * gRPC is generally faster than REST except for ndarray which is much worse and should not be used with gRPC
 * tftensor is fastest


```python
!kubectl delete -f model.yaml
```

    seldondeployment.machinelearning.seldon.io "seldon-model" deleted


## Test Predct Raw


```python
%%writetemplate model.yaml
apiVersion: machinelearning.seldon.io/v1
kind: SeldonDeployment
metadata:
  name: seldon-model
  namespace: seldon
spec:
  predictors:
  - annotations:
      seldon.io/no-engine: "true"
    componentSpecs:
    - spec:
        containers:
        - image: seldonio/seldontest_predict_raw:{VERSION}
          imagePullPolicy: IfNotPresent
          name: classifier
          resources:
            requests:
              cpu: 1
            limits:
              cpu: 1
          env:
          - name: GUNICORN_WORKERS
            value: "1"
          - name: GUNICORN_THREADS
            value: "1"
        tolerations:
        - key: model
          operator: Exists
          effect: NoSchedule
    graph:
      children: []
      name: classifier
      type: MODEL
    name: default
    replicas: 1
```


```python
!kubectl apply -f model.yaml
```

    seldondeployment.machinelearning.seldon.io/seldon-model created



```python
!kubectl wait --for condition=ready --timeout=600s pods --all -n seldon
```

    pod/seldon-model-default-0-classifier-5dc8fbd597-kk7td condition met


Smoke test port-forward to check everything is working


```python
!curl -X POST -H 'Content-Type: application/json' \
   -d '@./data_tftensor.json' \
    http://localhost:8003/seldon/seldon/seldon-model/api/v1.0/predictions
```

    [1]


Test REST

 1. ndarray
 1. tensor
 1. tftensor
 
 This can be done locally as the results should be indicative of the relative differences rather than very accurate timings.


```bash
%%bash
vegeta attack -format=json -duration=10s -rate=0 -max-workers=1 -targets=vegeta_ndarray.json | 
  vegeta report -type=text
```

    Requests      [total, rate, throughput]         724, 72.35, 72.25
    Duration      [total, attack, wait]             10.021s, 10.007s, 14.458ms
    Latencies     [min, mean, 50, 90, 95, 99, max]  12.228ms, 13.838ms, 13.683ms, 14.641ms, 15.489ms, 17.888ms, 22.263ms
    Bytes In      [total, mean]                     2896, 4.00
    Bytes Out     [total, mean]                     498774460, 688915.00
    Success       [ratio]                           100.00%
    Status Codes  [code:count]                      200:724  
    Error Set:



```bash
%%bash
vegeta attack -format=json -duration=10s -rate=0 -max-workers=1 -targets=vegeta_tensor.json | 
  vegeta report -type=text
```

    Requests      [total, rate, throughput]         724, 72.32, 72.22
    Duration      [total, attack, wait]             10.025s, 10.011s, 14.307ms
    Latencies     [min, mean, 50, 90, 95, 99, max]  12.362ms, 13.844ms, 13.701ms, 14.655ms, 15.493ms, 17.976ms, 18.802ms
    Bytes In      [total, mean]                     2896, 4.00
    Bytes Out     [total, mean]                     498792560, 688940.00
    Success       [ratio]                           100.00%
    Status Codes  [code:count]                      200:724  
    Error Set:



```bash
%%bash
vegeta attack -format=json -duration=10s -rate=0 -max-workers=1 -targets=vegeta_tftensor.json | 
  vegeta report -type=text
```

    Requests      [total, rate, throughput]         901, 90.04, 89.93
    Duration      [total, attack, wait]             10.018s, 10.007s, 11.64ms
    Latencies     [min, mean, 50, 90, 95, 99, max]  8.955ms, 11.116ms, 10.994ms, 12.099ms, 12.721ms, 15.208ms, 19.918ms
    Bytes In      [total, mean]                     3604, 4.00
    Bytes Out     [total, mean]                     961160671, 1066771.00
    Success       [ratio]                           100.00%
    Status Codes  [code:count]                      200:901  
    Error Set:


Example results

| ndarray | tensor | tftensor |
| ------- | ------ | -------- |
| 13.3ms | 13.3ms | 11.1ms |

 Test gRPC
 
  1. ndarray
  1. tensor
  1. tftensor


```bash
%%bash
ghz \
    --insecure \
    --proto ../../../proto/prediction.proto \
    --call seldon.protos.Seldon/Predict \
    --data-file=./data_ndarray.json \
    --qps=0 \
    --cpus=1 \
    --concurrency=1 \
    --duration="10s" \
    --format summary \
    --metadata='{"seldon": "seldon-model", "namespace": "seldon"}' \
    localhost:8003
```

    
    Summary:
      Count:	44
      Total:	10.04 s
      Slowest:	69.07 ms
      Fastest:	44.44 ms
      Average:	46.03 ms
      Requests/sec:	4.38
    
    Response time histogram:
      44.440 [1]	|∎
      46.904 [31]	|∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎
      49.367 [6]	|∎∎∎∎∎∎∎∎
      51.831 [2]	|∎∎∎
      54.294 [2]	|∎∎∎
      56.758 [0]	|
      59.221 [0]	|
      61.684 [0]	|
      64.148 [0]	|
      66.611 [0]	|
      69.075 [1]	|∎
    
    Latency distribution:
      10 % in 45.05 ms 
      25 % in 45.40 ms 
      50 % in 46.30 ms 
      75 % in 47.34 ms 
      90 % in 50.16 ms 
      95 % in 53.38 ms 
      0 % in 0 ns 
    
    Status code distribution:
      [OK]         43 responses   
      [Canceled]   1 responses    
    
    Error distribution:
      [1]   rpc error: code = Canceled desc = grpc: the client connection is closing   
    



```bash
%%bash
ghz \
    --insecure \
    --proto ../../../proto/prediction.proto \
    --call seldon.protos.Seldon/Predict \
    --data-file=./data_tensor.json \
    --qps=0 \
    --cpus=1 \
    --concurrency=1 \
    --duration="10s" \
    --format summary \
    --metadata='{"seldon": "seldon-model", "namespace": "seldon"}' \
    localhost:8003
```

    
    Summary:
      Count:	92
      Total:	10.10 s
      Slowest:	19.81 ms
      Fastest:	4.93 ms
      Average:	7.91 ms
      Requests/sec:	9.11
    
    Response time histogram:
      4.932 [1]	|∎
      6.419 [53]	|∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎
      7.907 [12]	|∎∎∎∎∎∎∎∎∎
      9.395 [5]	|∎∎∎∎
      10.882 [4]	|∎∎∎
      12.370 [1]	|∎
      13.858 [3]	|∎∎
      15.346 [3]	|∎∎
      16.833 [2]	|∎∎
      18.321 [3]	|∎∎
      19.809 [4]	|∎∎∎
    
    Latency distribution:
      10 % in 5.21 ms 
      25 % in 5.68 ms 
      50 % in 6.04 ms 
      75 % in 8.27 ms 
      90 % in 15.77 ms 
      95 % in 19.04 ms 
      0 % in 0 ns 
    
    Status code distribution:
      [OK]         91 responses   
      [Canceled]   1 responses    
    
    Error distribution:
      [1]   rpc error: code = Canceled desc = grpc: the client connection is closing   
    



```bash
%%bash
ghz \
    --insecure \
    --proto ../../../proto/prediction.proto \
    --call seldon.protos.Seldon/Predict \
    --data-file=./data_tftensor.json \
    --qps=0 \
    --cpus=1 \
    --concurrency=1 \
    --duration="10s" \
    --format summary \
    --metadata='{"seldon": "seldon-model", "namespace": "seldon"}' \
    localhost:8003
```

    
    Summary:
      Count:	426
      Total:	10.03 s
      Slowest:	11.74 ms
      Fastest:	3.67 ms
      Average:	5.02 ms
      Requests/sec:	42.48
    
    Response time histogram:
      3.668 [1]	|
      4.475 [174]	|∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎
      5.282 [141]	|∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎
      6.089 [43]	|∎∎∎∎∎∎∎∎∎∎
      6.897 [30]	|∎∎∎∎∎∎∎
      7.704 [16]	|∎∎∎∎
      8.511 [6]	|∎
      9.318 [8]	|∎∎
      10.126 [2]	|
      10.933 [1]	|
      11.740 [3]	|∎
    
    Latency distribution:
      10 % in 4.08 ms 
      25 % in 4.27 ms 
      50 % in 4.61 ms 
      75 % in 5.30 ms 
      90 % in 6.62 ms 
      95 % in 7.66 ms 
      99 % in 10.26 ms 
    
    Status code distribution:
      [OK]         425 responses   
      [Canceled]   1 responses     
    
    Error distribution:
      [1]   rpc error: code = Canceled desc = grpc: the client connection is closing   
    


Example results

| ndarray | tensor | tftensor |
| ------- | ------ | -------- |
| 46ms | 7.9ms | 5.0ms |

## Conclusions

 * `predict_raw` is faster than `predict` but you will need to handle the serialization/deserializtion yourself which maybe will make them equivalent unless specific techniques can be applied for your use case.

## Test with Predict method on Small Batch Size

The `seldontest_predict` has simply a `predict` method that does a loop with a configurable number of iterations (default 1) to simulate work. The iterations can be set as a Seldon parameter but in this case we are looking to benchmark the serialization/deserialization cost so want a minimal amount of work.


```python
%%writetemplate model.yaml
apiVersion: machinelearning.seldon.io/v1
kind: SeldonDeployment
metadata:
  name: seldon-model
  namespace: seldon
spec:
  predictors:
  - annotations:
      seldon.io/no-engine: "true"
    componentSpecs:
    - spec:
        containers:
        - image: seldonio/seldontest_predict:{VERSION}
          imagePullPolicy: IfNotPresent
          name: classifier
          resources:
            requests:
              cpu: 1
            limits:
              cpu: 1
          env:
          - name: GUNICORN_WORKERS
            value: "1"
          - name: GUNICORN_THREADS
            value: "1"
        tolerations:
        - key: model
          operator: Exists
          effect: NoSchedule
    graph:
      children: []
      name: classifier
      type: MODEL
    name: default
    replicas: 1
```


```python
!kubectl apply -f model.yaml
```

    seldondeployment.machinelearning.seldon.io/seldon-model configured



```python
!kubectl wait --for condition=ready --timeout=600s pods --all -n seldon
```

    pod/seldon-model-default-0-classifier-5445bd4ccf-bgkcm condition met


Create payloads and associated vegeta configurations for

  1. ndarray
  1. tensor
  1. tftensor
  
  We will create an array of 100,000 consecutive integers.


```python
import json

sz = 1
vals = list(range(sz))
valStr = f"{vals}"
payload = '{"data": {"ndarray": [' + valStr + "]}}"
with open("data_ndarray.json", "w") as f:
    f.write(payload)
payload_tensor = (
    '{"data":{"tensor":{"shape":[1,' + str(sz) + '],"values":' + valStr + "}}}"
)
with open("data_tensor.json", "w") as f:
    f.write(payload_tensor)
```


```python
import numpy as np
import tensorflow as tf
from google.protobuf import json_format

array = np.array(vals)
tftensor = tf.make_tensor_proto(array)
jStrTensor = json_format.MessageToJson(tftensor)
jTensor = json.loads(jStrTensor)
payload_tftensor = (
    '{"data":{"tftensor":' + json.dumps(jTensor, separators=(",", ":")) + "}}"
)
with open("data_tftensor.json", "w") as f:
    f.write(payload_tftensor)
```


```python
import base64
import json

sample_string_bytes = payload_tensor.encode("ascii")
base64_bytes = base64.b64encode(sample_string_bytes)
base64_string = base64_bytes.decode("ascii")
jqPayload = {
    "method": "POST",
    "url": "http://localhost:8003/seldon/seldon/seldon-model/api/v1.0/predictions",
    "body": base64_string,
    "header": {"Content-Type": ["application/json"]},
}
with open("vegeta_tensor.json", "w") as f:
    f.write(json.dumps(jqPayload, separators=(",", ":")))
    f.write("\n")

sample_string_bytes = payload.encode("ascii")
base64_bytes = base64.b64encode(sample_string_bytes)
base64_string = base64_bytes.decode("ascii")
jqPayload = {
    "method": "POST",
    "url": "http://localhost:8003/seldon/seldon/seldon-model/api/v1.0/predictions",
    "body": base64_string,
    "header": {"Content-Type": ["application/json"]},
}
with open("vegeta_ndarray.json", "w") as f:
    f.write(json.dumps(jqPayload, separators=(",", ":")))
    f.write("\n")


sample_string_bytes = payload_tftensor.encode("ascii")
base64_bytes = base64.b64encode(sample_string_bytes)
base64_string = base64_bytes.decode("ascii")
jqPayload = {
    "method": "POST",
    "url": "http://localhost:8003/seldon/seldon/seldon-model/api/v1.0/predictions",
    "body": base64_string,
    "header": {"Content-Type": ["application/json"]},
}
with open("vegeta_tftensor.json", "w") as f:
    f.write(json.dumps(jqPayload, separators=(",", ":")))
    f.write("\n")
```

Smoke test port-forward to check everything is working


```python
!curl -X POST -H 'Content-Type: application/json' \
   -d '@./data_tensor.json' \
    http://localhost:8003/seldon/seldon/seldon-model/api/v1.0/predictions
```

    {"data":{"names":[],"tensor":{"shape":[1],"values":[1]}},"meta":{"requestPath":{"classifier":"seldonio/seldontest_predict:1.10.0-dev"}}}


Test REST

 1. ndarray
 1. tensor
 1. tftensor
 
 This can be done locally as the results should be indicative of the relative differences rather than very accurate timings.


```bash
%%bash
vegeta attack -format=json -duration=10s -rate=0 -max-workers=1 -targets=vegeta_ndarray.json | 
  vegeta report -type=text
```

    Requests      [total, rate, throughput]         5538, 553.80, 553.67
    Duration      [total, attack, wait]             10.002s, 10s, 2.364ms
    Latencies     [min, mean, 50, 90, 95, 99, max]  1.569ms, 1.804ms, 1.739ms, 1.984ms, 2.198ms, 2.861ms, 6.62ms
    Bytes In      [total, mean]                     636870, 115.00
    Bytes Out     [total, mean]                     155064, 28.00
    Success       [ratio]                           100.00%
    Status Codes  [code:count]                      200:5538  
    Error Set:



```bash
%%bash
vegeta attack -format=json -duration=10s -rate=0 -max-workers=1 -targets=vegeta_tensor.json | 
  vegeta report -type=text
```

    Requests      [total, rate, throughput]         5557, 555.65, 555.55
    Duration      [total, attack, wait]             10.003s, 10.001s, 1.753ms
    Latencies     [min, mean, 50, 90, 95, 99, max]  1.578ms, 1.798ms, 1.74ms, 1.925ms, 2.119ms, 2.981ms, 5.968ms
    Bytes In      [total, mean]                     761309, 137.00
    Bytes Out     [total, mean]                     266736, 48.00
    Success       [ratio]                           100.00%
    Status Codes  [code:count]                      200:5557  
    Error Set:



```bash
%%bash
vegeta attack -format=json -duration=10s -rate=0 -max-workers=1 -targets=vegeta_tftensor.json | 
  vegeta report -type=text
```

    Requests      [total, rate, throughput]         4548, 454.75, 454.65
    Duration      [total, attack, wait]             10.003s, 10.001s, 2.141ms
    Latencies     [min, mean, 50, 90, 95, 99, max]  1.937ms, 2.197ms, 2.138ms, 2.351ms, 2.482ms, 3.215ms, 9.424ms
    Bytes In      [total, mean]                     850476, 187.00
    Bytes Out     [total, mean]                     436608, 96.00
    Success       [ratio]                           100.00%
    Status Codes  [code:count]                      200:4548  
    Error Set:


Example results

| ndarray | tensor | tftensor |
| ------- | ------ | -------- |
| 1.8ms | 1.8ms | 2.1ms |

 Test gRPC
 
  1. ndarray
  1. tensor
  1. tftensor


```bash
%%bash
ghz \
    --insecure \
    --proto ../../../proto/prediction.proto \
    --call seldon.protos.Seldon/Predict \
    --data-file=./data_ndarray.json \
    --qps=0 \
    --cpus=1 \
    --concurrency=1 \
    --duration="10s" \
    --format summary \
    --metadata='{"seldon": "seldon-model", "namespace": "seldon"}' \
    localhost:8003
```

    
    Summary:
      Count:	6506
      Total:	10.01 s
      Slowest:	18.58 ms
      Fastest:	1.26 ms
      Average:	1.46 ms
      Requests/sec:	650.23
    
    Response time histogram:
      1.260 [1]	|
      2.992 [6465]	|∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎
      4.724 [30]	|
      6.456 [5]	|
      8.187 [2]	|
      9.919 [1]	|
      11.651 [0]	|
      13.382 [0]	|
      15.114 [0]	|
      16.846 [0]	|
      18.578 [1]	|
    
    Latency distribution:
      10 % in 1.33 ms 
      25 % in 1.36 ms 
      50 % in 1.39 ms 
      75 % in 1.45 ms 
      90 % in 1.58 ms 
      95 % in 1.79 ms 
      99 % in 2.50 ms 
    
    Status code distribution:
      [OK]            6505 responses   
      [Unavailable]   1 responses      
    
    Error distribution:
      [1]   rpc error: code = Unavailable desc = transport is closing   
    



```bash
%%bash
ghz \
    --insecure \
    --proto ../../../proto/prediction.proto \
    --call seldon.protos.Seldon/Predict \
    --data-file=./data_tensor.json \
    --qps=0 \
    --cpus=1 \
    --concurrency=1 \
    --duration="10s" \
    --format summary \
    --metadata='{"seldon": "seldon-model", "namespace": "seldon"}' \
    localhost:8003
```

    
    Summary:
      Count:	6429
      Total:	10.01 s
      Slowest:	16.30 ms
      Fastest:	1.29 ms
      Average:	1.49 ms
      Requests/sec:	642.56
    
    Response time histogram:
      1.287 [1]	|
      2.789 [6375]	|∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎
      4.290 [36]	|
      5.792 [11]	|
      7.293 [2]	|
      8.795 [1]	|
      10.296 [0]	|
      11.798 [1]	|
      13.299 [0]	|
      14.801 [0]	|
      16.303 [1]	|
    
    Latency distribution:
      10 % in 1.36 ms 
      25 % in 1.38 ms 
      50 % in 1.42 ms 
      75 % in 1.48 ms 
      90 % in 1.60 ms 
      95 % in 1.80 ms 
      99 % in 2.67 ms 
    
    Status code distribution:
      [OK]            6428 responses   
      [Unavailable]   1 responses      
    
    Error distribution:
      [1]   rpc error: code = Unavailable desc = transport is closing   
    



```bash
%%bash
ghz \
    --insecure \
    --proto ../../../proto/prediction.proto \
    --call seldon.protos.Seldon/Predict \
    --data-file=./data_tftensor.json \
    --qps=0 \
    --cpus=1 \
    --concurrency=1 \
    --duration="10s" \
    --format summary \
    --metadata='{"seldon": "seldon-model", "namespace": "seldon"}' \
    localhost:8003
```

    
    Summary:
      Count:	6066
      Total:	10.01 s
      Slowest:	9.38 ms
      Fastest:	1.39 ms
      Average:	1.57 ms
      Requests/sec:	606.20
    
    Response time histogram:
      1.387 [1]	|
      2.187 [5945]	|∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎
      2.986 [84]	|∎
      3.785 [20]	|
      4.585 [7]	|
      5.384 [2]	|
      6.183 [4]	|
      6.983 [0]	|
      7.782 [0]	|
      8.582 [1]	|
      9.381 [1]	|
    
    Latency distribution:
      10 % in 1.46 ms 
      25 % in 1.48 ms 
      50 % in 1.52 ms 
      75 % in 1.57 ms 
      90 % in 1.66 ms 
      95 % in 1.81 ms 
      99 % in 2.61 ms 
    
    Status code distribution:
      [OK]            6065 responses   
      [Unavailable]   1 responses      
    
    Error distribution:
      [1]   rpc error: code = Unavailable desc = transport is closing   
    


Example results

| ndarray | tensor | tftensor |
| ------- | ------ | -------- |
| 1.46ms | 1.49ms | 1.57ms |

## Conclusions

 * gRPC is generally faster than REST
 * There is very little difference between payload types with simpler tensor/ndarray probably being slightly faster


```python
!kubectl delete -f model.yaml
```

    seldondeployment.machinelearning.seldon.io "seldon-model" deleted



```python

```
