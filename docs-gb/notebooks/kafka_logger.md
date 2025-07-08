# Kafka Request Logging Tests

This notebook illustrates testing your model with Kafka payload logging.

## Prequisites

 * An authenticated K8S cluster with istio and Seldon Core installed
   * You can use the ansible seldon-core and kafka playbooks in the root ansible folder.
 * vegeta and ghz benchmarking tools
 
 Port forward to istio
 
 ```
 kubectl port-forward $(kubectl get pods -l istio=ingressgateway -n istio-system -o jsonpath='{.items[0].metadata.name}') -n istio-system 8003:8080
 ```
 
  * Tested on GKE with 6 nodes of 32vCPU  e2-standard-32 



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


```python
!kubectl create namespace seldon
```

## CIFAR10 Model running on Triton Inference Server

We run CIFAR10 image model on Triton inference server with settings to allow 5 CPUs to be used for model on Triton.


```python
%%writetemplate model.yaml
apiVersion: machinelearning.seldon.io/v1
kind: SeldonDeployment
metadata:
  name: cifar10
  namespace: seldon
spec:
  name: resnet32
  predictors:
  - componentSpecs:
    - spec:
        containers:
        - name: cifar10
          resources:
            requests:
              cpu: 5
            limits:
              cpu: 5
    graph:
      implementation: TRITON_SERVER
      logger:
        mode: all
      modelUri: gs://seldon-models/triton/tf_cifar10_5cpu
      name: cifar10
    name: default
    svcOrchSpec:
      env:
      - name: LOGGER_KAFKA_BROKER
        value: seldon-kafka-plain-0.kafka:9092
      - name: LOGGER_KAFKA_TOPIC
        value: seldon
      - name: GOMAXPROCS
        value: "2"
      resources:
        requests:
          memory: "3G"
          cpu: 2
        limits:
          memory: "3G"
          cpu: 2
    replicas: 15
  protocol: kfserving
```


```python
!kubectl apply -f model.yaml -n seldon
```


```python
!kubectl wait --for condition=ready --timeout=600s pods --all -n seldon
```


```python
!curl -X POST -H 'Content-Type: application/json' \
   -d '@./truck-v2.json' \
    http://localhost:8003/seldon/seldon/cifar10/v2/models/cifar10/infer
```

## Direct Tests to Validate Setup



```bash
%%bash
vegeta attack -format=json -duration=10s -rate=0 -max-workers=1 -targets=vegeta_cifar10.json | 
  vegeta report -type=text
```

## Run Vegeta Benchmark



```python
!kubectl create -f configmap_cifar10.yaml -n seldon
```


```python
workers = 10
duration = "300s"
```


```python
%%writetemplate job-vegeta-cifar10.yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: cifar10-loadtest
spec:
  backoffLimit: 6
  parallelism: 16
  template:
    metadata:
      annotations:
        sidecar.istio.io/inject: "false"
    spec:
      containers:
        - args:
            - vegeta -cpus=1 attack -format=json -keepalive=false -duration={duration} -rate=0 -max-workers={workers} -targets=/var/vegeta/cifar10.json
              | vegeta report -type=text
          command:
            - sh
            - -c
          image: peterevans/vegeta:latest
          imagePullPolicy: Always
          name: vegeta
          volumeMounts:
            - mountPath: /var/vegeta
              name: vegeta-cfg
      restartPolicy: Never
      volumes:
        - configMap:
            defaultMode: 420
            name: vegeta-cfg
          name: vegeta-cfg

```


```python
!kubectl create -f job-vegeta-cifar10.yaml -n seldon
```


```python
!kubectl wait --for=condition=complete job/cifar10-loadtest -n seldon
```


```python
!kubectl delete -f job-vegeta-cifar10.yaml -n seldon
```


```python
!kubectl delete -f model.yaml
```

## Summary

By looking at the Kafka Grafana monitoring on e can inspect the achieved message rate.

You can port-forward to it with:

```
kubectl port-forward svc/kafka-grafana -n kafka 3000:80
```

The default login and password is set to `admin`.

On the above deployment and test we see around 3K predictions per second resulting in 6K Kafka messages per second.


```python

```
