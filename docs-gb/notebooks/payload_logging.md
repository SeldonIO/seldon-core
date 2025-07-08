# Payload Logging 

An example of payload logging of Seldon Deployment requests and responses.

## Prerequisites

 * A kubernetes cluster with kubectl configured
 * curl
 * grpcurl
 * pygmentize
 

## Setup Seldon Core

Install Seldon Core as described in [docs](https://docs.seldon.io/projects/seldon-core/en/latest/workflow/install.html)

Then port-forward to that ingress on localhost:8003 in a separate terminal either with:

 * Ambassador: 
 
 ```bash 
 kubectl port-forward $(kubectl get pods -n seldon -l app.kubernetes.io/name=ambassador -o jsonpath='{.items[0].metadata.name}') -n seldon 8003:8080```
 
 * Istio: 
 
 ```bash 
 kubectl port-forward $(kubectl get pods -l istio=ingressgateway -n istio-system -o jsonpath='{.items[0].metadata.name}') -n istio-system 8003:80```
 


```python
!kubectl create namespace seldon
```

    Error from server (AlreadyExists): namespaces "seldon" already exists



```python
!kubectl config set-context $(kubectl config current-context) --namespace=seldon
```

    Context "kind-kind" modified.



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




    '1.5.0-dev'



## Deploy a Request Logger

This will echo CloudEvents it receives.



```python
!pygmentize message-dumper.yaml
```

    [94mapiVersion[39;49;00m: apps/v1
    [94mkind[39;49;00m: Deployment
    [94mmetadata[39;49;00m:
      [94mname[39;49;00m: logger
    [94mspec[39;49;00m:
      [94mselector[39;49;00m:
        [94mmatchLabels[39;49;00m:
          [94mrun[39;49;00m: logger
      [94mreplicas[39;49;00m: 1
      [94mtemplate[39;49;00m:
        [94mmetadata[39;49;00m:
          [94mlabels[39;49;00m:
            [94mrun[39;49;00m: logger
        [94mspec[39;49;00m:
          [94mcontainers[39;49;00m:
          - [94mname[39;49;00m: logger
            [94mimage[39;49;00m: mendhak/http-https-echo
            [94mports[39;49;00m:
            - [94mcontainerPort[39;49;00m: 80
    [04m[36m---[39;49;00m
    [94mapiVersion[39;49;00m: v1
    [94mkind[39;49;00m: Service
    [94mmetadata[39;49;00m:
      [94mname[39;49;00m: logger
      [94mlabels[39;49;00m:
        [94mrun[39;49;00m: logger
    [94mspec[39;49;00m:
      [94mports[39;49;00m:
      - [94mport[39;49;00m: 80
        [94mtargetPort[39;49;00m: 80
        [94mprotocol[39;49;00m: TCP
      [94mselector[39;49;00m:
        [94mrun[39;49;00m: logger
    
        



```python
!kubectl apply -f message-dumper.yaml -n seldon
```

    deployment.apps/logger created
    service/logger created



```python
!kubectl rollout status deploy/logger
```

    Waiting for deployment "logger" rollout to finish: 0 of 1 updated replicas are available...
    deployment "logger" successfully rolled out


## Create a Model with Logging


```python
%%writetemplate model_logger.yaml
apiVersion: machinelearning.seldon.io/v1
kind: SeldonDeployment
metadata:
  name: model-logs
spec:
  name: model-logs
  predictors:
  - componentSpecs:
    - spec:
        containers:
        - image: seldonio/mock_classifier:{VERSION}
          name: classifier
    graph:
      children: []
      endpoint:
        type: REST
      name: classifier
      type: MODEL
      logger:
        url: http://logger.seldon/
        mode: all
    name: logging
    replicas: 1

```


```python
!kubectl apply -f model_logger.yaml -n seldon
```

    seldondeployment.machinelearning.seldon.io/model-logs created



```python
!kubectl rollout status deploy/$(kubectl get deploy -l seldon-deployment-id=model-logs -o jsonpath='{.items[0].metadata.name}')
```

    Waiting for deployment "model-logs-logging-0-classifier" rollout to finish: 0 of 1 updated replicas are available...
    deployment "model-logs-logging-0-classifier" successfully rolled out


## Send a Prediction Request


```python
res=!curl -s -d '{"data": {"ndarray":[[1.0, 2.0, 5.0]]}}' \
   -X POST http://localhost:8003/seldon/seldon/model-logs/api/v1.0/predictions \
   -H "Content-Type: application/json";
print(res)
import json
j=json.loads(res[0])
assert(j["data"]["ndarray"][0][0]>0.2)
```

    ['{"data":{"names":["proba"],"ndarray":[[0.43782349911420193]]},"meta":{}}']


## Check Logger


```python
!kubectl logs $(kubectl get pods -l run=logger -n seldon -o jsonpath='{.items[0].metadata.name}') logger
```

    -----------------
    {
        "path": "/",
        "headers": {
            "host": "logger.seldon",
            "user-agent": "Go-http-client/1.1",
            "content-length": "39",
            "ce-endpoint": "logging",
            "ce-id": "e276a0c3-a522-4499-b509-5e98e06e96fe",
            "ce-inferenceservicename": "model-logs",
            "ce-modelid": "classifier",
            "ce-namespace": "seldon",
            "ce-requestid": "33fb419e-8d6b-44e0-a084-8bd4bd4dbf7b",
            "ce-source": "http://:8000/",
            "ce-specversion": "1.0",
            "ce-time": "2020-11-06T09:35:09.171644169Z",
            "ce-traceparent": "00-9c7cc210d352f7b68be6422c1b4b78f4-a7e8a2a38709bbc9-00",
            "ce-type": "io.seldon.serving.inference.request",
            "content-type": "application/json",
            "traceparent": "00-9c7cc210d352f7b68be6422c1b4b78f4-b49edd5ad25552ae-00",
            "accept-encoding": "gzip"
        },
        "method": "POST",
        "body": "{\"data\": {\"ndarray\":[[1.0, 2.0, 5.0]]}}",
        "fresh": false,
        "hostname": "logger.seldon",
        "ip": "::ffff:10.244.1.65",
        "ips": [],
        "protocol": "http",
        "query": {},
        "subdomains": [],
        "xhr": false,
        "os": {
            "hostname": "logger-766f99b9b7-mqtql"
        },
        "connection": {},
        "json": {
            "data": {
                "ndarray": [
                    [
                        1,
                        2,
                        5
                    ]
                ]
            }
        }
    }
    ::ffff:10.244.1.65 - - [06/Nov/2020:09:35:09 +0000] "POST / HTTP/1.1" 200 1220 "-" "Go-http-client/1.1"
    -----------------
    {
        "path": "/",
        "headers": {
            "host": "logger.seldon",
            "user-agent": "Go-http-client/1.1",
            "content-length": "73",
            "ce-endpoint": "logging",
            "ce-id": "68345c32-d144-4e21-a840-55e7e809c002",
            "ce-inferenceservicename": "model-logs",
            "ce-modelid": "classifier",
            "ce-namespace": "seldon",
            "ce-requestid": "33fb419e-8d6b-44e0-a084-8bd4bd4dbf7b",
            "ce-source": "http://:8000/",
            "ce-specversion": "1.0",
            "ce-time": "2020-11-06T09:35:09.180317759Z",
            "ce-traceparent": "00-cbb2fa5d83dbc42f2f8e9f8957b5c121-c15418121da2e992-00",
            "ce-type": "io.seldon.serving.inference.response",
            "content-type": "application/json",
            "traceparent": "00-cbb2fa5d83dbc42f2f8e9f8957b5c121-ce0a53c967ee8077-00",
            "accept-encoding": "gzip"
        },
        "method": "POST",
        "body": "{\"data\":{\"names\":[\"proba\"],\"ndarray\":[[0.43782349911420193]]},\"meta\":{}}\n",
        "fresh": false,
        "hostname": "logger.seldon",
        "ip": "::ffff:10.244.1.65",
        "ips": [],
        "protocol": "http",
        "query": {},
        "subdomains": [],
        "xhr": false,
        "os": {
            "hostname": "logger-766f99b9b7-mqtql"
        },
        "connection": {},
        "json": {
            "data": {
                "names": [
                    "proba"
                ],
                "ndarray": [
                    [
                        0.43782349911420193
                    ]
                ]
            },
            "meta": {}
        }
    }
    ::ffff:10.244.1.65 - - [06/Nov/2020:09:35:09 +0000] "POST / HTTP/1.1" 200 1312 "-" "Go-http-client/1.1"



```python
modelids = !kubectl logs $(kubectl get pods -l run=logger -n seldon -o jsonpath='{.items[0].metadata.name}') logger | grep "ce-modelid"
print(modelids)
assert modelids[0].strip() == '"ce-modelid": "classifier",'
```

    ['        "ce-modelid": "classifier",', '        "ce-modelid": "classifier",']


## Clean Up


```python
!kubectl delete -f model_logger.yaml -n seldon
```

    seldondeployment.machinelearning.seldon.io "model-logs" deleted



```python
!kubectl delete -f message-dumper.yaml -n seldon
```

    deployment.apps "logger" deleted
    service "logger" deleted



```python

```
