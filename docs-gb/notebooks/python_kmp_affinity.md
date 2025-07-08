# Python Wrapper KMP_AFFINITY Tests

This notebook illustrates testing your model with KMP_AFFINITY settings.

## Prequisites

 * An authenticated K8S cluster with istio and Seldon Core installed
   * You can use the ansible seldon-core playbook at https://github.com/SeldonIO/ansible-k8s-collection
 * vegeta and ghz benchmarking tools
 
 Port forward to istio
 
 ```
 kubectl port-forward $(kubectl get pods -l istio=ingressgateway -n istio-system -o jsonpath='{.items[0].metadata.name}') -n istio-system 8003:8080
 ```
 
  * Tested on GKE with 3 nodes of 32vCPU  e2-highcpu-32 



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


## CIFAR10 Model with KMP Settings

We run a custom python model built using [Intel's Tensorflow library](https://pypi.org/project/intel-tensorflow/).


```python
KMP_AFFINITY = "verbose,disabled"
OMP_NUM_THREADS = 1
```


```python
%%writetemplate model.yaml
apiVersion: machinelearning.seldon.io/v1alpha2
kind: SeldonDeployment
metadata:
  name: seldon-model
spec:
  predictors:
  - graph:
      name: classifier
    componentSpecs:    
    - spec:
        containers:
        - name: classifier
          image: seldonio/keras_cifar10:1.10.0-dev
          resources:
            requests:
              cpu: 10
            limits:
              cpu: 10
          env:
          - name: GUNICORN_WORKERS
            value: "10"
          - name: GUNICORN_THREADS
            value: "1"
          - name: KMP_AFFINITY
            value: "{KMP_AFFINITY}"
          - name: OMP_NUM_THREADS
            value: "{OMP_NUM_THREADS}"
          - name: KMP_SETTINGS
            value: "TRUE"
          - name: KMP_BLOCKTIME
            value: "1"
        tolerations:
        - key: model
          operator: Exists
          effect: NoSchedule
    name: default
    replicas: 1
```


```python
!kubectl apply -f model.yaml -n seldon
```

    seldondeployment.machinelearning.seldon.io/seldon-model created



```python
!kubectl wait --for condition=ready --timeout=600s pods --all -n seldon
```

    pod/seldon-model-default-0-classifier-7b6f7d5ddf-qnw48 condition met



```python
!curl -X POST -H 'Content-Type: application/json' \
   -d '@./cifar10.json' \
    http://localhost:8003/seldon/seldon/seldon-model/api/v1.0/predictions
```

    {"data":{"names":["t:0","t:1","t:2","t:3","t:4","t:5","t:6","t:7","t:8","t:9"],"ndarray":[[1.716785118333064e-05,1.575566102474113e-06,2.213756124547217e-05,0.047145213931798935,0.00011693393025780097,0.01819806545972824,0.9344443678855896,6.195103196660057e-06,4.7716683184262365e-05,6.440643574023852e-07]]},"meta":{"requestPath":{"classifier":"seldonio/keras_cifar10:1.10.0-dev"}}}


## Direct Tests

Inital warm-up tests via port-forward.


```bash
%%bash
vegeta attack -format=json -duration=20s -rate=0 -max-workers=1 -targets=vegeta_cifar10.json | 
  vegeta report -type=text
```

    Requests      [total, rate, throughput]         110, 5.49, 5.46
    Duration      [total, attack, wait]             20.128s, 20.028s, 100.14ms
    Latencies     [min, mean, 50, 90, 95, 99, max]  92.268ms, 182.983ms, 103.193ms, 349.642ms, 872.554ms, 1.013s, 1.036s
    Bytes In      [total, mean]                     42350, 385.00
    Bytes Out     [total, mean]                     7077730, 64343.00
    Success       [ratio]                           100.00%
    Status Codes  [code:count]                      200:110  
    Error Set:


## Run Vegeta Benchmark



```python
!kubectl create -f configmap_cifar10.yaml -n seldon
```

    configmap/vegeta-cfg created



```python
workers = 10
duration = "60s"
```


```python
%%writetemplate job-vegeta-cifar10.yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: cifar10-loadtest
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

    job.batch/cifar10-loadtest created



```python
!kubectl wait --for=condition=complete job/cifar10-loadtest -n seldon
```

    error: timed out waiting for the condition on jobs/cifar10-loadtest



```python
!kubectl logs $(kubectl get pod -l job-name=cifar10-loadtest -n seldon -o jsonpath='{.items[0].metadata.name}') -n seldon
```

    Requests      [total, rate, throughput]         8677, 144.61, 144.46
    Duration      [total, attack, wait]             1m0s, 1m0s, 63.126ms
    Latencies     [min, mean, 50, 90, 95, 99, max]  55.621ms, 69.182ms, 66.127ms, 79.525ms, 83.454ms, 126.962ms, 230.469ms
    Bytes In      [total, mean]                     3340645, 385.00
    Bytes Out     [total, mean]                     558304211, 64343.00
    Success       [ratio]                           100.00%
    Status Codes  [code:count]                      200:8677  
    Error Set:



```python
!kubectl delete -f job-vegeta-cifar10.yaml -n seldon
```

    job.batch "cifar10-loadtest" deleted



```python
!kubectl delete -f model.yaml
```

    Error from server (NotFound): error when deleting "model.yaml": seldondeployments.machinelearning.seldon.io "seldon-model" not found


## Notes

Always ensure your resource.cpu.linits >= GUNICORN_WORKERS otherwise there may be [CPU throttling](https://medium.com/omio-engineering/cpu-limits-and-aggressive-throttling-in-kubernetes-c5b20bd8a718)


```python

```
