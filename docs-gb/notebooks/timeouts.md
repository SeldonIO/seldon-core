# Testing Custom Timeouts


## Prerequisites

  * Kubernetes cluster with kubectl authorized
  * curl and grpcurl installed

## Setup Seldon Core

Use the setup notebook to [Setup Cluster](../notebooks/seldon-core-setup.md) to setup Seldon Core with an ingress - either Ambassador or Istio.

Then port-forward to that ingress on localhost:8003 in a separate terminal either with:

 * Ambassador: `kubectl port-forward $(kubectl get pods -n seldon -l app.kubernetes.io/name=ambassador -o jsonpath='{.items[0].metadata.name}') -n seldon 8003:8080`
 * Istio: `kubectl port-forward $(kubectl get pods -l istio=ingressgateway -n istio-system -o jsonpath='{.items[0].metadata.name}') -n istio-system 8003:8080`


```python
!kubectl create namespace seldon
```


```python
!kubectl config set-context $(kubectl config current-context) --namespace=seldon
```


```python
from IPython.core.magic import register_line_cell_magic


@register_line_cell_magic
def writetemplate(line, cell):
    print("writing template")
    with open(line, "w") as f:
        f.write(cell.format(**globals()))
```


```python
VERSION = !cat ../version.txt
VERSION = VERSION[0]
VERSION
```

## Configure Istio

For this example we will create the default istio gateway for seldon which needs to be called `seldon-gateway`. You can supply your own gateway by adding to your SeldonDeployments resources the annotation `seldon.io/istio-gateway` with values the name of your istio gateway.


```python
%%writefile resources/seldon-gateway.yaml
apiVersion: networking.istio.io/v1alpha3
kind: Gateway
metadata:
  name: seldon-gateway
  namespace: istio-system
spec:
  selector:
    istio: ingressgateway # use istio default controller
  servers:
  - port:
      number: 80
      name: http
      protocol: HTTP
    hosts:
    - "*"
```


```python
!kubectl create -f resources/seldon-gateway.yaml -n istio-system
```

### Short timeouts deployment file
Below is the outputs of the Seldon config file required to set custom timeouts.

First we'll show how we can set a short lived example by setting the `"seldon.io/rest-timeout":"1000"` annotation with a fake model that delays for 30 secs.


```python
%%writetemplate resources/model_short_timeouts.json
apiVersion: machinelearning.seldon.io/v1
kind: SeldonDeployment
metadata:
  name: seldon-model
spec:
  annotations:
    seldon.io/rest-timeout: "1000"
    seldon.io/grpc-timeout: "1000"
  predictors:
  - componentSpecs:
    - spec:
        containers:
        - image: seldonio/mock_classifier:{VERSION}
          name: classifier
    graph:
      children: []
      name: classifier
      type: MODEL
      parameters:
      - name: delaySecs
        type: INT
        value: "30"       
    name: example
    replicas: 1
```

We can then apply this Seldon Deployment file we defined above in the `seldon` namespace.


```python
!kubectl apply -f resources/model_short_timeouts.json -n seldon
```

And wait until the resource runs correctly.


```python
!curl -i -d '{"data": {"ndarray":[[1.0, 2.0, 5.0]]}}' -X POST http://localhost:8003/seldon/seldon/seldon-model/api/v1.0/predictions -H "Content-Type: application/json"
```


```python
!cd ../executor/proto && grpcurl -d '{"data":{"ndarray":[[1.0,2.0,5.0]]}}' \
         -rpc-header seldon:seldon-model -rpc-header namespace:seldon \
         -plaintext \
         -proto ./prediction.proto  0.0.0.0:8003 seldon.protos.Seldon/Predict
```

Delete this graph and recreate one with a longer timeout


```python
!kubectl delete -f resources/model_short_timeouts.json
```

## Seldon Deployment with Long Timeout

Now we can have a look at how we can set a deployment with a longer timeout.


```python
%%writetemplate resources/model_long_timeouts.json
apiVersion: machinelearning.seldon.io/v1
kind: SeldonDeployment
metadata:
  name: seldon-model
spec:
  annotations:
    seldon.io/rest-timeout: "40000"
    seldon.io/grpc-timeout: "40000"
  predictors:
  - componentSpecs:
    - spec:
        containers:
        - image: seldonio/mock_classifier:{VERSION}
          name: classifier
    graph:
      children: []
      name: classifier
      type: MODEL
      parameters:
      - name: delaySecs
        type: INT
        value: "30"       
    name: example
    replicas: 1
```

And we can apply it to deploy the model with long timeouts.


```python
!kubectl apply -f resources/model_long_timeouts.json -n seldon
```


```python
!kubectl rollout status deploy/$(kubectl get deploy -l seldon-deployment-id=seldon-model -o jsonpath='{.items[0].metadata.name}')
```

This next request will work given that the timeout is much longer.


```python
!curl -i -d '{"data": {"ndarray":[[1.0, 2.0, 5.0]]}}'    -X POST http://localhost:8003/seldon/seldon/seldon-model/api/v1.0/predictions    -H "Content-Type: application/json"
```


```python
!cd ../executor/proto && grpcurl -d '{"data":{"ndarray":[[1.0,2.0,5.0]]}}' \
         -rpc-header seldon:seldon-model -rpc-header namespace:seldon \
         -plaintext \
         -proto ./prediction.proto  0.0.0.0:8003 seldon.protos.Seldon/Predict
```


```python
!kubectl delete -f resources/model_long_timeouts.json
```
